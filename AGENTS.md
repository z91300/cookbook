# AGENTS.md

本文件是 AI 协作（agent）必须遵循的项目工程约定。改动相关代码前先读本文件。人机共用，人工开发同样遵守。

## 项目形态

- Go 1.26 + GoFrame v2 后端（module 名 `cookbook`）+ Nuxt 4 前端（`web/`），单仓全栈。
- 模板基线只保留 `hello` 示例模块与最小首页；`manifest/init.sql` 不含业务表。新业务模块在此基础上按「新增模块流程」生长，不要回填示例业务。

## 通用设施（模板自带，禁止删除或改坏）

| 设施 | 位置 | 作用 |
| --- | --- | --- |
| 统一响应信封 | `internal/handler.MiddlewareResponse` + `internal/model.StandardRes` | 所有 JSON 接口出口 `{code, message, data}` |
| 多值查询参数归一 | `internal/handler.MiddlewareQueryMultiValue` | `?k=a&k=b` 等价 `k[]=a&k[]=b`，须在绑定前注册 |
| 分页模型 | `internal/model.PaginationInput` / `PageRes[T]` / `NewPageRes` | 列表接口入参出参复用 |
| 信封拆包 | `web/app/api/request.ts` | code≠0 抛 `ApiError`，前端只拿 data |
| 无限加载 | `web/app/composables/useInfiniteList.ts` | 上拉分页，配合 PageRes 形状 |
| operationId 拆分 | `web/worma.config.ts` splitOperationId 插件 | `资源_动作` → 前端命名空间 |

删除 hello 示例时同步清掉：`api/hello/`、`internal/controller/hello/`、`internal/cmd/cmd.go` 的 Bind、`web/app/api/hello.ts` 与 `web/app/api/index.ts` 中对应导入、首页自检调用。

## 分层职责（后端）

```
api/<资源>/v1        接口定义：Req/Res（组合 model 结构，只补 g.Meta）
internal/model       字段唯一定义点：请求/响应字段、校验、dc 描述、分页与信封
internal/controller  控制器：组合 api 接口并转发 logic，不写业务
internal/logic       业务实现：实现 service 接口
internal/service     服务接口：gf gen service 从 logic 生成，勿手改
internal/dao         数据访问：gf gen dao 生成，勿手改
internal/handler     全局中间件（统一响应信封、多值查询参数归一）
internal/consts      全局常量
```

- 依赖方向：controller → (api + service 接口)，logic → dao/model；禁止反向、禁止 controller 直连 dao。
- **生成文件不可手改**：`api/**/api 根接口文件`、`internal/model/entity|do`、`internal/dao`、`internal/service` 由 CLI 维护；controller 与 logic 可自由填充。

## 新增模块流程

以资源 `foo` 为例：

1. `internal/model/foo.go` 定义 `FooListInput`（嵌入 `model.PaginationInput`）、`FooItem` 等结构，字段带校验 `v:"..."` 与描述 `dc:"..."`
2. `api/foo/v1/foo.go` 定义 Req/Res，`g.Meta` 声明 path/tags/method/summary/`operationId`
3. `api/foo/foo.go`：`gf gen ctrl -m` 生成（**必须带 `-m`**，否则回退为每方法一个文件）
4. `internal/logic/foo/foo.go` 实现业务；`internal/controller/foo/` 组合转发
5. `internal/logic/logic.go` 追加 `_ "cookbook/internal/logic/foo"` 导入；`internal/cmd/cmd.go` 的 `group.Bind` 挂载
6. `gf gen service` 更新 service 接口；后端启动后 `cd web && pnpm api:gen` 生成前端客户端

## API 层（api/ 目录）

### operationId（强制）

每个接口的 `g.Meta` **必须**声明 `operationId`，格式：`资源_动作`。

- **资源**：小写单数，对应领域名/表名主体，如 `pet`、`user`
- **动作**：驼峰动词，常用集合：
  - `getList` —— 分页列表查询（`GET /foos`）
  - `getOne` —— 单条详情（`GET /foos/{id}`）
  - `create` —— 创建（`POST`）
  - `update` —— 全量修改（`PUT`）
  - `delete` —— 删除（`DELETE`）
  - `reorder` —— 整体排序（`PUT /foos/sort`，提交完整 id 顺序；静态路径优先于 `{id}` 模糊规则，已在 `/tags/sort`、`/recipes/sort` 上实测可用）
  - `getManageList` —— 后台管理态全量列表（不分页，如 `GET /recipes/manage`；同样受静态路径优先保护）
- 全库 operationId 不得重复；它是 OpenAPI 的唯一操作标识（前端 worma 代码生成、mock 等依赖它）
- 示例：

```go
type GetReq struct {
    g.Meta `path:"/foos/{id}" tags:"Foo管理" method:"get" summary:"查询Foo详情" operationId:"foo_getOne"`
    model.FooIdInput
}
```

### 其他接口约定

- **多选查询参数（同名重复键）**：列表接口的多选条件统一用 `?elements=火&elements=水` 写法（ofetch/axios 等客户端对数组的默认序列化）。注意 **GoFrame 原生只识别 `k[]=a&k[]=b`**，对同名重复键只保留最后一个值（`gstr.Parse` 语义），会造成多选静默退化为单选；已由 `internal/handler.MiddlewareQueryMultiValue` 在参数绑定前归一（在 `internal/cmd/cmd.go` 全局注册）。新增多选条件时无需额外处理，字段按 `[]string`/`[]uint` 定义即可，两种写法都可用。
- **字段唯一定义点**：请求/响应字段与校验规则只写在 `internal/model/<领域>.go`；api 层组合嵌入 model 结构（只补 `g.Meta`），controller 只做组合转发，service 输入输出直接使用 model 结构。禁止在 api 层和 service 层各定义一遍字段。
- **编辑接口的「可选字段」用指针语义**：标量字段一律定义成 `*string` / `*int` / `*int64`（GoFrame `gconv` 对缺省键留 nil、对传 `""`/`0` 给出非 nil 指针，前端正常发全量字段不受影响）。写入侧按 `nil=本次不修改、非 nil=显式写入（含清空）` 处理，否则 `summary`/封面/难度/热量这类字段一旦设过就再也清不掉。新建（Create）相反：所有列显式写入，零值即零值。
- **多表写入必须同事务**：一次业务动作里若有「更新主表 + 重建关联表」，用 `dao.X.Transaction(ctx, func(ctx, tx) error {...})`（回调内继续用 `dao.X.Ctx(ctx)` 即自动落到该事务），中途失败要整体回滚，不留半更新。关联表批量插入用 `Data([]g.Map{...}).Insert()`（一次多值 INSERT），不要逐条循环插。
- **删除主记录要清理引用**：`recipe.Delete` 为逻辑删除，同事务内硬删 `recipe_tags` / `favorite_items`，`schedulings` 行保留但置 `is_deleted=1`（历史可追溯、不再指向已删食谱）。
- **列表查询用 Fields 白名单**：`recipe.List` / `ListByIds` 只取卡片列，不要 `SELECT *` 把 `ingredients`/`tools`/`steps` 三个 JSON 大列拉出来。用 `ScanAndCount(ptr, &total, true)` 时 Fields 走多列也不影响 total（内部回落 `COUNT(1)`）。
- **GET 不写库**：读接口不得带写副作用（前端弹窗/页面会反复调用）。默认收藏夹「我的收藏」由 `favorite.EnsureDefaultFolder` 在**注册（初始化时机）**用一条 `INSERT ... SELECT ... WHERE NOT EXISTS` 原子创建（幂等、无并发重复插入窗口、归属用户正确），`favorite.List` 保持纯读。
- **食谱手动排序（`recipes.sort`）**：`0 = 未参与排序`（等价「新菜谱置顶」），拖拽排序后由 `PUT /recipes/sort`（`recipe_reorder`）把全部菜谱的 `sort` 重写为连续的 `1..N`。展示顺序统一为 `ORDER BY sort ASC, id DESC`（首页 `recipe.List` 与设置页 `recipe.ManageList` 都按它，`ListByIds` 保持调用方给的收藏顺序）。与 `tag.Reorder` 同规矩：**必须提交全部未删除菜谱的完整顺序（不重不漏），否则整体拒绝**。新建菜谱不写 sort（默认 0）→ 仍保持「刚建的排最前」的既有观感。
- **拖拽交互一律用 `vue-draggable-plus`**（`web/package.json` 依赖，SortableJS 的 Vue 封装），不要再手写 `draggable` + drag 事件——手写版在触屏上不可用、也没有自动滚动。统一配置 `:force-fallback="true"`（用指针事件驱动，表格行/移动端才可靠）+ `handle=".xxx__handle"` + `ghost-class`/`drag-class` 复用现有样式类；松手回调里 `@end` 提交「完整顺序」，失败回滚为服务端顺序。
- **字段描述**：字段的对外描述用 `dc:"…"` 标签写在 model 字段上（与 Go `//` 注释文案保持一致）；GoFrame OpenAPI 只读 `dc`/`des`/`description` 标签，不读 Go 注释，注释不写 `dc` 则 swagger 无字段说明。
- **分页**：入参嵌入 `model.PaginationInput`；响应用别名 `type FooListOutput = model.PageRes[FooItem]`。
- **统一响应**：HTTP 出口信封 `model.StandardRes`（`{code, message, data}`），由 `internal/handler.MiddlewareResponse` 输出；错误码通过 `gerror`/`gcode` 携带。

## 用户与权限（auth）

角色极简：只有 `users.is_admin` 一个角色位，不做角色表/权限表。

| 能力 | 管理员 | 普通用户 | 未登录 |
| --- | --- | --- | --- |
| 浏览/搜索/查看菜谱、收藏夹、成员、编排、通用设置（读） | ✅ | ✅ | ✅（只能看） |
| 新建/编辑菜谱、收藏夹与收藏项、成员、编排、上传附件、保存通用设置（写） | ✅ | ✅ | ❌（61 请先登录） |
| 标签管理（`tag_create` / `tag_update` / `tag_delete`） | ✅ | ❌（61 无权限） | ❌ |
| 删除菜谱（`recipe_delete`） | ✅ | ❌（61 无权限） | ❌ |
| 用户管理（`user_getList` / `user_update` / `user_resetPassword`） | ✅ | ❌（61 无权限） | ❌ |

- **首个注册用户即管理员**：`internal/logic/auth.Register` 的判据是「尚不存在可登录的管理员」
  （`is_admin=1 AND status=1 AND password_hash<>''`），兼容库中残留的无密码占位行。
- **双令牌登录态**（表 `user_sessions`，明文都不落库、只存 sha256）：
  - `access`：**2 小时**，前端存 `localStorage['cookbook_token']`，请求头 `Authorization: Bearer <token>`
  - `refresh`：**30 天**（每次刷新重置，滑动续期），存 `localStorage['cookbook_refresh_token']`，
    仅用于 `POST /auth/refresh`，**每次刷新都轮换 refresh**（旧的立即作废）
  - 登出删除会话；禁用账号 / 重置密码 / 取消管理员会清空该用户全部会话
- **错误码区分「未登录」与「登录态过期」**——这是前端能否自动续期的关键：
  - `61`（`gcode.CodeNotAuthorized`）：未带令牌，或已登录但无权限 → 前端按未登录处理
  - `4401`（`internal/consts.CodeTokenExpired`）：带了 access 但已过期/会话被删 →
    前端用 refresh 换新令牌对后**自动重试原请求**；刷新也失败才清空本地令牌
- **鉴权中间件**：`internal/handler.MiddlewareAuth` 只识别身份、不拦截（无令牌→匿名；令牌失效→
  在上下文标记 `TokenInvalid`），注册顺序必须在 `MiddlewareResponse` 之前（见 `internal/cmd/cmd.go`）。
  权限判断全在 **logic 层**：`auth.MustLogin(ctx)`（写操作）/ `auth.MustAdmin(ctx)`（管理操作）。
- **前端**：`app/composables/useAuth.ts`（`useState('auth-user')` 全局共享 + 进站 `fetchProfile`；
  暴露 `isLoggedIn` / `isAdmin` / `canEdit`）；`canEdit=false`（未登录）时页面上的新建、编辑、收藏、
  编排、保存等入口一律隐藏（首页/收藏/成员/编排/设置 + `QuickNav` 悬浮球整体隐藏），后端仍独立校验。
  令牌读写与「单飞刷新 + 重试」都在 `app/api/request.ts`: **务必从 `~/api/request` 导入**这些工具——
  `app/api/index.ts` 是 worma 生成文件，只重复导出 `ApiError/request/ApiRequestConfig`，
  往聚合入口加东西会在下次 `api:gen` 时丢失（`MISSING_EXPORT` 会让整站失去交互）。
- **用户保护性约束**（`auth.Update`）：不能禁用/降权自己；系统至少保留一名「正常状态」的管理员。

## 数据库

- **本地开发**：外部 PostgreSQL（链接写在根 `config.yaml` 的 `database.default.link`；
  `hack/config.yaml` 的 gen dao link 同步指向它；真实配置含凭据不入 git，入库
  `config.example.yaml` / `hack/config.example.yaml` 模板）。换库方式：改 `link`；
  **注意**：`GF_GCFG_DATABASE_DEFAULT_LINK` 这类变量对 GoFrame 无效（实测被忽略，仅 `GF_GCFG_FILE` /
  `GF_GCFG_PATH` 有效），不要指望用环境变量直接覆盖配置项；配置内 `${ENV|默认}` 占位符在本版本同样不生效。
- **生产部署**：数据库运行时环境变量注入（`internal/cmd/dbenv.go`：`DB_TYPE` postgres|sqlite +
  `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME`，sqlite 用 `DB_DATA_PATH`），镜像内无配置文件
  无凭据；本地开发仍走根 `config.yaml`。换库方式：改 compose 环境变量。
  **注意**：`GF_GCFG_DATABASE_DEFAULT_LINK` 这类变量对 GoFrame 无效（实测被忽略），必须走
  dbenv 装配或配置文件，见 AGENTS.md「Docker 部署 / CI」节。
- 写库前先更新 `manifest/init.sql` 再整体执行（脚本幂等：`CREATE TABLE IF NOT EXISTS` +
  `ON CONFLICT DO NOTHING`；生产 PG 与本地均连 PostgreSQL，同一脚本维护）。
  一次性迁移命令存档：`cmd/migrate_sqlite_to_pg`（历史 SQLite → PG，已执行过）。
- 字段含义写在行尾 `--` 注释（gen dao 读取为字段描述）
- 生成命令 `gf gen dao`（读取 `hack/config.yaml` 的 link，指向外部 PostgreSQL）
- 时间戳字段 `*_at` 统一 Unix 秒（INTEGER）；逻辑删除固定 `is_deleted`；主键 `id`
  `GENERATED BY DEFAULT AS IDENTITY`，显式插入 id 后需 `setval` 校准序列
- **附件 blob 存储**：文件本体以二进制存 `attachments.content`（PG `BYTEA`），
  访问 URL 统一为 `/attachments/{id}/content`（由 `internal/logic/attachment.ContentUrl` 拼装，
  `GET` 直出二进制并带 `Content-Type` / `Content-Disposition`）。`storage_path` 仅为唯一标识
  （历史迁移字段），不再指向磁盘文件；新增写附件的代码一律写 `content` 列，不要落盘。
  前端开发期 `/attachments` 由 nuxt.config.ts 的 devProxy/routeRules 与 `/api` 同款反代到后端。
  内容路由支持 `?w=360` 缩略图（等比缩小不放大，仅 360 档位；结果内存缓存 + ETag/长缓存
  不可变头，ETag 含内容指纹）；列表卡片等小尺寸封面一律用 `web/app/utils/thumbUrl.ts`
  拼 `?w=360`，详情大图保留原 URL。查询参数绑定用 `p:"w"` 标签（`json:"-"` 字段靠 p 标签
  从 query 绑定）。
  **图片落库统一 webp**：`Upload` 时栅格图（jpeg/png/webp/gif）经
  `attachment.ToWebpBytes` 重编码为 webp q85；svg/动图跳过；重编码无体积收益（已高度压缩
  的小 JPEG）则保留原格式。存量迁移命令 `go run ./cmd/migrate_webp`（dry-run，`--apply` 写库），已执行过。
  webp 编码用 `github.com/gen2brain/webp`（libwebp 转译纯 Go，无 CGo），解码用
  `golang.org/x/image/webp`。
  **上传安全（禁止放松，见 `internal/logic/attachment/mime.go`）**：
  - 类型只认**文件魔数**（`detectMime`：`net/http` 嗅探 + 自补 mp4 的 `ftyp`），
    客户端 `Content-Type` 与扩展名都可能伪造，不作为判据；文件名扩展名按真实类型纠正
    （`alignFileName`），存储名经 `sanitizeFileName` 剔除引号/控制字符（防响应头注入）。
  - `inlineSafeMimes` 只放栅格图与视频。**`image/svg+xml`、`text/html`、`text/xml` 等一律不在白名单**
    （可执行脚本），上传时即降级为 `kind=file` + `application/octet-stream`。
  - 内容路由：白名单类型才 `Content-Disposition: inline`，其余一律 `attachment` + `application/octet-stream`，
    且所有响应都带 `X-Content-Type-Options: nosniff`。判定在**读路径**（`GetContent` → `Inline`）复核，
    历史库里客户端声明的 `image/svg+xml` 等老行同样按下载处置。
    副作用：**SVG 附件不再能在 `<img>` 里显示**（防存储型 XSS 的代价），需要展示就转成栅格图。
  - 单文件上限 `consts.MaxUploadBytes`(20MB)，读文件必须 `io.ReadFull`（单次 `Read` 不保证读满）；
    同时 `internal/cmd/cmd.go` 必须设 `SetClientMaxBodySize(consts.MaxRequestBodyBytes)`——
    GoFrame 默认请求体上限仅 8MB，不放开则 8~20MB 的文件在 multipart 解析阶段直接 500。
  - `attachments.sha256` 记内容哈希，`Upload` 命中同哈希（未删除）即**秒传**复用已有附件，不重复落库。

## 前端（web/）
- **API 客户端**：`web/app/api/` 由 `pnpm api:gen`（worma，配置见 `web/worma.config.ts`）从后端 `/api.json` 生成；`request.ts` 已做 StandardRes 信封拆包（code≠0 抛 `ApiError`）。生成文件除 `request.ts` 外勿手改。
- **调用方式**：`apis.<资源>.<动作>()`，如 `apis.hello.get()`；命名空间来自 operationId 的 `资源_动作` 拆分（worma.config.ts 内置 splitOperationId 插件）。
- **代理**：开发期 `/api` 由 nuxt.config.ts 的 devProxy 转发到 `http://127.0.0.1:8000`，后端地址变更只改这里与 runtimeConfig。
- **devServer**：显式绑定 `127.0.0.1`（本机 DNS 常把 localhost 解析为 ::1 导致 127.0.0.1 打不开），不要移除该配置。
- **分页列表**：直接复用 `app/composables/useInfiniteList.ts`（上拉无限加载，配合 `model.PageRes` 形状）。
- **页面**：放 `web/app/pages/`，按 Nuxt 路由约定命名；通用展示组件放 `web/app/components/`。
- **设置页（`pages/settings.vue`）**：页签按角色分叉——管理员 `菜谱管理 / 标签管理 / 用户管理 / 通用设置 / 关于`，普通用户与匿名只有 `通用设置 / 关于`（默认落地页 = 管理员落「菜谱管理」、其余落「通用设置」）。菜谱管理是 10 列宽表格，该页签下 `<main>` 用 `page`（max-w-6xl），其余页签用 `page--narrow`（max-w-3xl）。用法注意：`useBodyScrollLock` / `useModalBackClose` 会**立即求值**传入的谓词，写在目标 ref 声明之前会命中 TDZ 直接白屏（500），新增弹窗时把这两行放在 ref 声明之后。

## 启动与验证

- 后端：仓库根目录 `go run .`（端口 8000）；改动后先 `go build ./...` 再启动验证。
- 前端：`cd web && pnpm dev`（端口 3000，绑定 127.0.0.1）；改前端必须 `pnpm dev` 实际访问验证，重要改动跑 `pnpm build`。
- Windows 提示：本机 DNS 可能将 localhost 解析为 ::1，验证统一用 `http://127.0.0.1:<port>`。

## Docker 部署 / CI

- GitHub Actions（`.github/workflows/build.yml`）：push main / 打 v* 标签时自动构建镜像
  → 推送 GHCR。CI 不需要 Secret（数据库凭据运行时经 DB_* 环境变量注入，镜像内无配置）。
  推送前人工本地验证：起隔离栈跑 `python scripts/ci_selftest.py cookbook:latest`（CI 不跑自测）。
- 一键构建/自测/推送（本地）：`uv run python scripts/docker_build_push.py -v <版本> [-r <registry前缀>]`
  （本机无 python 时用 uv；需 Docker 引擎健康；自测需先设好 DB_* 环境变量）。
  默认流程 = 构建单个镜像 `cookbook`（容器内同时跑 Nuxt SSR 与 GoFrame 后端，暴露 3000/8000
  两端口）→ 起隔离测试栈跑 HTTP 自测（`/api/tags`、`/api/recipes`、首页 SSR）→ 按版本
  tag + latest 推送。推送需 `-r` 且先 `docker login`。
- 编排文件：根目录 `docker-compose.yml`，**不使用 .env**；变量（`COOKBOOK_IMAGE`、
  `COOKBOOK_WEB_PORT`、`COOKBOOK_API_PORT`、`GOPROXY`、`NPM_REGISTRY`）均来自环境变量且带默认值。
  单服务 `cookbook` 对外暴露两个端口：`COOKBOOK_WEB_PORT`（默认 8080 → 容器 3000，前端 SSR）、
  `COOKBOOK_API_PORT`（默认 8000 → 容器 8000，后端 API）。
- 镜像：根目录 `Dockerfile` 多阶段单 target `cookbook`（web-build + api-build 合并进 `node:24-alpine`
  运行时，`/entrypoint.sh` 同进程拉起 `./cookbook` 与 node SSR），构建上下文受 `.dockerignore` 约束
  （后端连接外部 PostgreSQL，镜像内不再生成库文件，`manifest/init.sql` 仅随镜像留档）。
- 生产 `/api` 代理：容器内前端与后端同机，`web/nuxt.config.ts` 的 `nitro.routeRules` 固定反代
  `http://127.0.0.1:8000`（无需构建期注入）；开发期仍走 devProxy，两者行为一致（去 `/api` 前缀
  转发后端根路径）。
- 数据持久化：数据库在外部 PG，附件本体存库（`attachments.content` BLOB），无附件目录卷需求。
- 脚本自检：加 `--dry-run` 可只打印将要执行的命令（不连 Docker 引擎也能校验脚本/编排改动）。
- 自测隔离：脚本用独立 compose 项目名 `cookbook-selftest` + 独立卷，测完 `down -v` 销毁，不影响生产栈。

## 命名约定

- Go 文件/包名小写下划线，按资源目录组织；TS/Vue 用官方风格（PascalCase 组件、camelCase 变量）。
- 表名 snake_case 复数或单数均可，但须与资源名一致（如资源 `note` → 表 `notes`）；时间戳字段 `*_at`（Unix 秒），逻辑删除固定 `is_deleted`。
