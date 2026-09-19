# Cookbook · 全栈项目模板

GoFrame v2（Go）后端 + Nuxt 4（@nuxt/ui）前端的全栈单仓初始化模板：仅保留 Hello 示例接口与最小页面，数据库 `init.sql` 不含任何业务表。新项目请从本模板起步，业务模块按 [AGENTS.md](AGENTS.md) 的「新增模块流程」生长。

## 技术栈

| 层 | 技术 | 说明 |
| --- | --- | --- |
| 后端 | Go 1.26 / GoFrame v2.10 | HTTP 服务、OpenAPI（`/api.json`）与 Swagger（`/swagger`）自动生成 |
| 数据库 | PostgreSQL 18 | 外部实例（链接在 `config.yaml`，见 `config.example.yaml` 模板），结构由 `manifest/init.sql` 维护 |
| 前端 | Nuxt 4 / @nuxt/ui v4 / TailwindCSS v4 | SSR + 组件库 |
| API 客户端 | wormajs | 读取后端 OpenAPI 自动生成前端 TS 请求代码（`web/app/api/`） |

## 目录结构

```
cookbook/
├── main.go                  # 后端入口
├── config.example.yaml      # 后端本地配置模板（复制为 config.yaml 填入真实链接；真实配置不入 git）
├── api/                     # 接口定义层（每资源一个目录，v1 版本子目录）
│   └── hello/
├── internal/
│   ├── cmd/                 # 启动命令与路由注册
│   ├── consts/              # 全局常量
│   ├── controller/          # 控制器：组合 api 层接口并转发 logic
│   ├── handler/             # 全局中间件（统一响应信封 / 多值查询参数归一）
│   ├── logic/               # 业务逻辑实现（service 接口的实现方）
│   ├── model/               # 请求/响应/实体模型（字段唯一定义点）
│   ├── dao/                 # 数据访问（gf gen dao 生成，勿手改）
│   └── service/             # 服务接口（gf gen service 生成，勿手改）
├── manifest/
│   ├── init.sql             # 数据库结构初始化脚本
│   └── config/config.yaml   # 部署环境配置
├── hack/                    # GoFrame CLI 构建配置（gen dao 链接等）
├── resource/                # 静态资源/模板
├── utility/                 # 通用工具函数
└── web/                     # 前端（Nuxt 4）
    ├── app/api/             # worma 生成的 API 客户端（request.ts 信封拆包）
    ├── app/pages/           # 路由页面
    ├── app/composables/     # 组合式函数（如 useInfiniteList 分页加载）
    ├── nuxt.config.ts       # devServer 绑定 127.0.0.1，/api 代理到后端 8000
    └── worma.config.ts      # API 代码生成配置（operationId 拆分插件）
```

## 快速开始

### 后端（端口 8000）

前置要求：Go 1.26+；PostgreSQL 可达（复制 `config.example.yaml` 为 `config.yaml` 并填入数据库链接，首次启动前用 `manifest/init.sql` 初始化库结构，也可让驱动按需建表）。

```bash
# 仓库根目录
go run .
```

启动后：

- 健康检查：`GET http://127.0.0.1:8000/hello` → `Hello World!`
- OpenAPI：`http://127.0.0.1:8000/api.json`
- Swagger UI：`http://127.0.0.1:8000/swagger`

### 前端（端口 3000）

```bash
cd web
pnpm install
pnpm dev
```

访问 `http://127.0.0.1:3000`，首页会调用 `GET /hello` 做前后端连通性自检（后端未启动时显示错误提示）。开发期 `/api` 由 Nuxt devProxy 代理到 `http://127.0.0.1:8000`。

### 生产构建（可选）

```bash
cd web && pnpm build   # 前端产物输出到 web/.output/
```

## API 开发流程（后端 → 前端全链路）

1. **定义**：在 `internal/model/<领域>.go` 定义请求/响应结构（字段 + 校验 + `dc` 描述）
2. **接口**：在 `api/<资源>/v1/` 定义 Req/Res（组合 model 结构，`g.Meta` 必须带 `operationId`，规范见 [AGENTS.md](AGENTS.md)）
3. **实现**：`internal/logic/<资源>/` 写业务逻辑，`internal/controller/<资源>/` 组合转发
4. **注册**：`internal/cmd/cmd.go` 的 `group.Bind(...)` 挂载控制器
5. **生成**：后端运行中执行前端生成
   ```bash
   gf gen ctrl -m     # controller 骨架（必须带 -m）
   gf gen service     # service 接口
   cd web && pnpm api:gen   # 生成前端 TS 客户端（需后端已启动）
   ```

## 数据库流程

1. 修改 `manifest/init.sql`（字段含义写在行尾 `--` 注释，gen dao 会读取）
2. 初始化/重建库：在 PostgreSQL 上整体执行 `init.sql`（脚本幂等，`CREATE TABLE IF NOT EXISTS`；可用 psql：`psql "<你的链接>" -f manifest/init.sql`）
3. 生成数据层：`gf gen dao`（读取 `hack/config.yaml` 的数据库链接）

## 配置与安全

本地开发用配置文件（凭据不入 git）：`config.yaml` / `hack/config.yaml`（复制 `*.example.yaml` 填入真实链接）。

**Docker 部署不需要配置文件**：运行时环境变量注入数据库（见下方「Docker 部署」），镜像内无任何凭据。

## CI/CD（GitHub Actions）

`.github/workflows/build.yml`：push 到 `main` / 打 `v*` 标签 / 手动触发时自动构建单镜像（前端 SSR + 后端 API）并推送 GHCR（`ghcr.io/<owner>/cookbook`，tag：`latest` + short-sha，标签推送另有版本 tag）。CI 不需要任何 Secret（数据库凭据运行时注入）。

**推送前人工本地验证**（CI 不跑自测，约定）：

```bash
docker build -t cookbook:latest .
uv run python scripts/ci_selftest.py cookbook:latest   # 隔离栈自测：/api/tags、/api/recipes、SSR 首页
```

## Docker 部署

数据库通过环境变量注入（装配逻辑 `internal/cmd/dbenv.go`），无需挂载配置文件：

| 变量 | 说明 | 默认 |
| --- | --- | --- |
| `DB_TYPE` | `postgres`（别名 `pg`）或 `sqlite` | `postgres` |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | postgres 连接参数 | 5432 / root / cookbook |
| `DB_DATA_PATH` | sqlite 库文件路径（`DB_TYPE=sqlite` 时） | `/data/cookbook.db`（compose 已挂卷持久化） |

```bash
# PostgreSQL
DB_TYPE=postgres DB_HOST=<host> DB_PORT=25432 DB_USER=root DB_PASSWORD=<pw> DB_NAME=cookbook \
  docker compose up -d --wait

# SQLite（本地单机）
DB_TYPE=sqlite docker compose up -d --wait
```

- 一键构建（本地）：`uv run python scripts/docker_build_push.py -v <版本> [-r <registry前缀>]`（需本机 Docker 引擎健康）
- CI 产物直接可用：`docker compose pull` 后按上面设 DB_* 环境变量 `docker compose up -d --wait`。
- 编排：根目录 `docker-compose.yml`，默认镜像 `ghcr.io/z91300/cookbook:latest`，单容器暴露 `COOKBOOK_WEB_PORT`（默认 8080 → 3000，前端 SSR）与 `COOKBOOK_API_PORT`（默认 8000，后端 API），变量全部带默认值、不用 `.env`。

## 模板内容说明

模板内置的最小示例（新项目可保留作连通性自检，或删除）：

- **后端**：`api/hello/` + `internal/controller/hello/` —— `GET /hello` 返回纯文本 `Hello World!`（演示「不包统一信封」的直出响应）
- **前端**：`web/app/pages/index.vue` —— 首页连通性自检卡片；`web/app/api/hello.ts` —— 生成的客户端调用示例
- **通用设施**（必须保留）：统一响应信封、多值查询参数归一（`internal/handler/`）、分页模型（`internal/model/pagination.go`）、无限加载组合式（`web/app/composables/useInfiniteList.ts`）、API 客户端信封拆包（`web/app/api/request.ts`）

## 工程规范

所有 AI 协作与人工开发共用的强制约定（operationId、分层职责、字段唯一定义点、统一响应信封、生成文件不可手改等）见 [AGENTS.md](AGENTS.md)。
