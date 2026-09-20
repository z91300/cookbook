# 容器自动更新与版本方案（保持 latest 场景）

> 适用：cookbook（单镜像双端口：Nuxt SSR 3000 + GoFrame 8000，外部 PostgreSQL）。
> 结论先行：镜像**可以**一直用 `latest` 并自动更新，但必须同时补齐三件事 ——
> **① sha 标签兜底回滚、② stable 通道做发布闸门、③ 不要开 cleanup 删旧镜像**。

---

## 1. 机制：容器怎么获得"操作 Docker"的能力

Docker 的 API 入口是 unix socket `/var/run/docker.sock`（不是 `sock.d`；Windows 宿主机是命名管道
`\\.\pipe\docker_engine`，Docker Desktop 跑 Linux 容器时仍走 sock）。Compose 里挂载它：

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
```

挂载之后，容器里的进程就能像宿主机 root 一样调 Docker API —— 可以创建特权容器、把宿主机 `/`
挂进去，**等价于宿主机 root**。所以：

- 这个 sock **只给专职更新器**，绝不要挂进 cookbook 这种业务容器（业务被攻破 = 宿主机沦陷）。
- `:ro` **不构成有效安全边界**（socket 是特殊文件，读写语义不同于普通文件），别把它当防线。
- 真正的加固是 **`tecnativa/docker-socket-proxy`**：HAProxy 按 HTTP 方法 + 路径做白名单，
  默认只放行 `EVENTS` / `PING` / `VERSION`。要更新容器至少需要 `CONTAINERS=1`、`IMAGES=1`、`POST=1`，
  还可以用 `ALLOW_START` / `ALLOW_STOP` 进一步细粒度控制；`EXEC=0`、`VOLUMES=0`、`BUILD=0`
  保持关闭（`EXEC` 是逃逸最常用的口子）。

不想挂 sock 的替代路径：更新器用 `DOCKER_HOST=tcp://<host>:2375` 连远端（**必须配 TLS**，
裸 2375 等于把 Docker 全权交给整个网络），或者干脆从外部 SSH 执行 `docker compose pull && up -d`。

---

## 2. 保持 `latest` 怎么自动更新：三种执行器

| 方案 | 镜像 | 自动更新 | 关键能力 | 适合 |
| --- | --- | --- | --- | --- |
| Watchtower | `nickfedor/watchtower` | ✅ | label 白名单、HTTP API 触发、通知 | 默认推荐 |
| WUD | `getwud/wud` | 可选 | Web UI、semver 过滤、Prometheus | 想看得见、要版本策略 |
| Diun | `crazymax/diun` | ❌ 只通知 | 17+ 通知渠道 | 必须人工确认 |
| compose + cron | `docker:cli` | ✅ | 完全自己写脚本 | 想 100% 可控 |

### 2.1 为什么 `latest` 也能被检测到

更新器比对的是**镜像 digest**，不是 tag 名字。所以：

- `latest`：CI 每次构建都重推 → digest 每次都变 → **能自动发现**。
- 固定 `v1.4.2` 且内容没重推 → digest 不变 → **永远不更新**（这是特性，不是 bug）。

一句话：**自动更新的粒度，由 compose 里钉的是哪个 tag 决定。**
用 `latest` 就是"每次 main 提交都自动上线"，用 `stable` 就是"每次确认发版才自动上线"（见 §3-C）。

### 2.2 Watchtower 完整片段（含 socket 代理加固）

`containrrr/watchtower` 已于 2025-12-17 归档（最后版 1.7.1，不再有补丁），改用活跃分支
`nickfedor/watchtower`，环境变量 / label / 行为完全兼容，换镜像名即可。

```yaml
services:
  docker-socket-proxy:
    image: tecnativa/docker-socket-proxy
    restart: unless-stopped
    environment:
      CONTAINERS: 1        # 必须：读容器、创建/重建容器
      IMAGES: 1            # 必须：拉取镜像
      POST: 1              # 必须：允许写操作（置 0 = 纯只读，更新会失败）
      NETWORKS: 1
      VOLUMES: 1
      EXEC: 0              # 禁止 exec，堵住最常见的逃逸口
      BUILD: 0
      SYSTEM: 0
      SWARM: 0
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks: [dockerctl]

  watchtower:
    image: nickfedor/watchtower
    restart: unless-stopped
    depends_on: [docker-socket-proxy]
    environment:
      DOCKER_HOST: tcp://docker-socket-proxy:2375   # 不挂 sock，只走代理
      WATCHTOWER_LABEL_ENABLE: "true"               # 只更新显式打 label 的容器
      WATCHTOWER_SCHEDULE: "0 30 4 * * *"           # 秒 分 时 日 月 周 → 每天 04:30
      WATCHTOWER_ROLLING_RESTART: "true"
      WATCHTOWER_NOTIFICATIONS: shoutrrr
      WATCHTOWER_NOTIFICATION_URL: "smtp://..."     # 或无 webhook：换 generic/tg/bark
    networks: [dockerctl]

networks:
  dockerctl:
    driver: bridge
    internal: true        # 代理网络不对外，宿主外无法直接碰到 2375
```

被管理的容器加 label（**推荐用 `stable` 而不是 `latest`，见 §3-C**）：

```yaml
  cookbook:
    image: ${COOKBOOK_IMAGE:-ghcr.io/z91300/cookbook:stable}
    labels:
      - com.centurylinklabs.watchtower.enable=true
```

要点与坑：
- `--cleanup`（`WATCHTOWER_CLEANUP=true`）会删掉旧镜像 —— **开了就没法回滚**，本方案建议**关闭**
  （见 §4.2）。
- 只给**无状态**服务打这个 label。数据库、有状态中间件一律不要纳入自动更新。
- `WATCHTOWER_ROLLING_RESTART` 只在**多副本**时才有意义；cookbook 是单容器，重启就是几秒不可用。
- 想只在收到触发时更新、平时不轮询：`WATCHTOWER_HTTP_API_UPDATE=true` +
  `WATCHTOWER_HTTP_API_TOKEN=<token>`，然后由 CI 调 `POST /v1/update`（见 §3-B）。

### 2.3 WUD（想要界面 + 版本策略）

WUD 支持语义化版本分析、正则 / include-exclude 标签过滤、Web UI、Prometheus `/metrics`，
默认**只监控不更新**，要更新得显式加触发器标签：

```yaml
  whatsupdocker:
    image: getwud/wud
    restart: unless-stopped
    ports: ["127.0.0.1:3000:3000"]
    environment:
      WUD_WATCHER_LOCAL_SOCKET: /var/run/docker.sock
      WUD_WATCHER_LOCAL_WATCHBYDEFAULT: "false"   # 默认不监控任何容器
      WUD_WATCHER_LOCAL_CRON: "0 */12 * * *"
      WUD_AUTH_ADMIN_USER: admin
      WUD_AUTH_ADMIN_PASSWORD: <改掉默认值>
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

```yaml
  cookbook:
    labels:
      - wud.watch=true                  # 只是"出现在 UI 里"
      - wud.trigger.include=docker.auto # 加这行才是"自动重建"
```

WUD 的独特价值：可以配置成"只在 patch 级变化时自动更新"，比 Watchtower 的纯 digest 比对更聪明。

### 2.4 Diun（只通知，人工确认）

挂 sock（读容器信息）扫描 registry，发现 tag/digest 变化就通知，**永不拉取、永不重启**。
适合"必须有变更审批"的场景，或作为 WUD/Watchtower 之外的"第二双眼睛"。

---

## 3. 谁来按下"发布"这个按钮

### A. 定时轮询（纯 `latest`，全自动）
上面 §2.2 的 `WATCHTOWER_SCHEDULE` 就是。实现最简单，代价是"凌晨 4 点自动上线一个没人看过的
main 提交"。

### B. CI 主动触发（推荐）
Watchtower 常驻但**不开轮询**，只在 CI 发布成功后触发一次：

```yaml
# .github/workflows/build.yml 末尾
- name: Trigger deploy
  if: startsWith(github.ref, 'refs/tags/v')
  run: |
    curl -fsS -X POST \
      -H "Authorization: Bearer ${{ secrets.WATCHTOWER_TOKEN }}" \
      https://deploy.example.com/v1/update
```

优点：发布时机在 CI 手里、不需要把服务器 SSH 私钥放进 GitHub、失败能在 Actions 里看到。
前提是把 Watchtower 的 HTTP API 暴露到一个受控入口（反代 + TLS + 长 token），别裸奔公网。

### C. 双通道（最推荐，两者兼得）

保持你要的 `latest`，同时**多推一个 `stable` 通道**：

| 通道 | 什么时候推 | 谁跟它 |
| --- | --- | --- |
| `latest` | 每次 push main | 开发 / 演示环境、想看最新的人 |
| `stable` | 只在打 `v*` tag 时推 | **生产 compose + 自动更新只跟这个** |
| `sha-3f9a1c2` | 每次构建 | 回滚、追溯 |
| `v1.4.2` | 打 tag 时 | 版本记录 |

CI 侧一行改动（`type=raw,value=stable,enable=${{ startsWith(github.ref,'refs/tags/v') }}`），
就能实现"**镜像保持浮动标签自动更新，但只有我打 tag 那一刻才会自动上线**"。
生产 compose 写 `ghcr.io/z91300/cookbook:stable`，既不会钉死版本号，也不会被 main 的每次提交推着走。

### D. SSH 直连（最简，已落地到 `.github/workflows/build.yml`）

本质就是 CI 成功后远程执行一条命令：`docker compose up -d --pull always --wait`。
仓库里已经加好 `deploy` job（用 `appleboy/ssh-action`，**不需要 `cd`** —— 用 `COMPOSE_FILE`
指定编排文件，见下），开箱即用的前提是先在服务器上做三件事：

```bash
# 1) 本地生成专用部署密钥（不要复用个人密钥；密码短语留空，否则 CI 无法非交互使用）
ssh-keygen -t ed25519 -C "github-actions-deploy" -f deploy_key -N ""

# 2) 公钥装到目标服务器（追加到待部署用户的 ~/.ssh/authorized_keys）
ssh-copy-id -i deploy_key.pub <user>@<your-server>

# 3) 该用户必须能操作 Docker，否则 pull/up 会 permission denied
sudo usermod -aG docker <user>     # 重新登录后生效
```

然后在 GitHub 仓库填四项 Secret：`DEPLOY_HOST`、`DEPLOY_USER`、`DEPLOY_SSH_KEY`（第 1 步生成的
私钥全文）、`DEPLOY_PORT`（可选，默认 22）。

⚠️ **必须建在 Repository secrets，不是 Environment secrets**（页面顶部的层级选择容易看漏）：
环境级 Secret 只在**声明了同名 environment 的 job** 里可见，而本项目的 `deploy` job 没有声明
`environment:`，于是 job 读不到值，drone-ssh 会抛一句很含糊的 `error: missing server host`。
两个修法：① 把 Secret 挪到 Repository secrets（推荐，最省事）；② 想保留环境级（可挂审批规则），
就在 `deploy` job 上加 `environment: self`（环境名大小写敏感；配了 required reviewers 的话
每次部署都会等人工确认）。

`deploy` job 里已经加了 `Check deploy secrets` 前置步骤，Secret 缺失时会直接指出缺哪个、该放哪一层，
不会再只给一句 `missing server host`；另外 `script_stop` 在 ssh-action v1 里已不是合法输入
（v0.x 时代参数），失败判定由脚本内的 `set -e` 与显式 `exit 1` 负责，不要加回去。

**最容易卡住的一步：私有镜像要先登录。** GHCR 的 package 默认是私有的，服务器上没有凭据时
`docker compose pull` 直接 403：

```bash
# 用一个具备 read:packages 的 PAT 登录（只需做一次，凭据落在 ~/.docker/config.json）
echo "$GHCR_PAT" | docker login ghcr.io -u <github-user> --password-stdin
```

（或者去 GitHub → Package 设置里把该包改成 public，就没这一步了。）

服务器上 `/opt/zspace/docker-compose.yml` 的 `image:` 必须指向 CI 推送的标签
（如 `ghcr.io/z91300/cookbook:latest`，或 §3-C 的 `:stable`）。首次需手动
`docker compose up -d` 起一次，之后交给 CI。两点提示：

- 服务器这份 compose 建议**删掉 `build:` 段和构建 args**：服务器上没有源码，只需要 `image:`，
  留着 `build:` 只会让 `compose pull/up` 的语义变模糊。
- 暂时还不想真部署的话，给 job 加个开关：`if: needs.build.result == 'success' && vars.DEPLOY_ENABLED == 'true'`，
  然后在仓库 Variables 里加 `DEPLOY_ENABLED=true` 才生效。

**不用 `cd`：用 `COMPOSE_FILE` 或 `-f` 指向编排文件即可**，效果与进目录执行等价：

```bash
# 环境变量（脚本里最省事：设一次，后面所有 compose 子命令都生效）
export COMPOSE_FILE=/opt/zspace/docker-compose.yml
docker compose up -d --pull always --wait --wait-timeout 300

# 或每次显式指定文件
docker compose -f /opt/zspace/docker-compose.yml up -d --pull always --wait
```

三个容易被担心的点，都有确定答案：

- **项目目录**默认取第一个 `-f` / `COMPOSE_FILE` 文件所在目录，所以相对路径挂载、`.env` 都从
  `/opt/zspace` 解析 —— 与 `cd` 进去执行一致。（想更保险可再加 `--project-directory /opt/zspace`。）
- **项目名**取值优先级是 `-p` → `COMPOSE_PROJECT_NAME` → 文件顶层 `name:` → 项目目录名。
  本项目 compose 有顶层 `name: cookbook`，所以在任何目录执行都是同一个项目、同一批容器，
  **不会平白多起一套栈**。
- `pull` / `ps` / `logs` / `up` 行为完全一致。

顺带一个简化：`up -d` 自带 `--pull`，可以把 `pull` 和 `up` 合成一条命令
（`--pull` 需 compose v2.15+）：

```bash
docker compose up -d --pull always --wait --wait-timeout 300
```

**只更新某一个服务**：在命令末尾加服务名，拉取与重建都只作用于它：

```bash
docker compose up -d --pull always --wait --wait-timeout 300 cookbook
# 不想连带处理 depends_on 的依赖，再加 --no-deps
```

（不带服务名时 compose 也只会重建"配置或镜像有变化"的服务，单服务栈两者等效。）

相比原始那一行命令，落地版本做了五处改进：

1. **加 `--wait`**：`up -d` 是异步的，命令返回 ≠ 服务起来了。`up -d --wait --wait-timeout 300`
   会等 compose 里的 healthcheck 通过，不通过就非零退出 → **CI 变红**，你能第一时间知道，
   而不是等用户反馈。（本项目的 healthcheck 同时探 3000 与 8000，正好用得上。）
2. **`--pull always` 合并成一条命令**：省掉单独的 `docker compose pull`，也不会踩
   "只 `up` 不 `pull` 等于没升级"这个经典坑。
3. **失败时自动打日志**：`if ! docker compose up ...; then docker compose logs --tail 50; exit 1; fi`
   —— 容器起不来时 CI 的报错里直接带原因，不用再登服务器。
4. **不加 `docker image prune`**：旧 `latest` 被 untag 成 `<none>` 后会被 prune 顺手删掉，
   一删就没法回滚了（§4.2）。想清磁盘就只删明确过期的 sha 标签。
5. **加 `concurrency: group: deploy-production`**：连续两次推送时不会并发部署，避免两次
   pull/recreate 互相踩。

不想引入第三方 action 的话，原生 ssh 等价写法：

```yaml
- name: Deploy over SSH
  env:
    SSH_KEY: ${{ secrets.DEPLOY_SSH_KEY }}
  run: |
    mkdir -p ~/.ssh && printf '%s\n' "$SSH_KEY" > ~/.ssh/id_ed25519
    chmod 600 ~/.ssh/id_ed25519
    ssh -o StrictHostKeyChecking=accept-new -p "${{ secrets.DEPLOY_PORT || 22 }}" \
      "${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }}" \
      'export COMPOSE_FILE=/opt/zspace/docker-compose.yml
       docker compose up -d --pull always --wait --wait-timeout 300'
```

什么时候该换掉 SSH 方案：私有镜像得在服务器上长期存一个 PAT（轮换麻烦）、部署目标变多要维护
多份密钥、或者想在 CI 里拿到比 ssh 输出更细的部署结果 —— 到那一步再上 §3-B（Watchtower HTTP API）。

---

## 4. 保持 `latest` / `stable` 的两大代价，以及补丁

### 4.1 可追溯性：让镜像自证版本

浮动标签最大的问题是"线上跑的是哪个 commit"只能靠猜。补三件事即可：

1. **推 sha 标签**（`build.yml` 已经在推 `${GITHUB_SHA::7}`）+ OCI label
   （`docker/metadata-action` 自动打 `org.opencontainers.image.version / revision / source`）。
2. 构建期注入版本：`--build-arg VERSION=$(git describe --tags) COMMIT=$(git rev-parse --short HEAD)`，
   Go 侧 `-ldflags "-X cookbook/internal/consts.Version=$VERSION -X cookbook/internal/consts.Commit=$COMMIT"`。
3. 对外可见：加 `/api/version`（返回 version / commit / buildTime），并在进程启动时打一行日志
   `cookbook v1.4.2 (3f9a1c2) starting`。这样 `docker logs` 一眼可查。

补充手段：`docker inspect <容器> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`
即可拿到当前容器的 commit。

### 4.2 回滚：不要 cleanup

- **关闭 `WATCHTOWER_CLEANUP`**（旧 `latest` 只被改标签成 `<none>`，镜像层还在）。
- 定期只清"绝不可能回滚到的"旧镜像，或干脆不自动清（几百 MB 换回滚能力，很值）。
- 回滚动作（30 秒内完成）：把 compose 里的镜像临时钉到 sha，再起来：

```bash
# 找出上一个可用版本
docker image ls ghcr.io/z91300/cookbook --format '{{.Tag}}\t{{.CreatedAt}}\t{{.ID}}'
# 改 compose: image: ghcr.io/z91300/cookbook:sha-3f9a1c2
docker compose pull && docker compose up -d --wait
```

- Watchtower 更新前后的镜像 ID 会写进日志 / 通知，**留着通知记录就是回滚线索**。

### 4.3 数据层：镜像能回滚，schema 回不去

迁移一律"**只加不删**"（expand → migrate → contract，删除动作滞后到确定不再回滚的那次发布）。
保证新 schema 上 **N-1 版镜像仍能正常运行**，回滚才是真的安全。本项目的 `manifest/init.sql`
是幂等的（`CREATE TABLE IF NOT EXISTS` + `ON CONFLICT DO NOTHING`），方向已正确；
新增**不可空且无默认值**的列、重命名列这两类操作要格外小心。

---

## 5. 与本项目现状的对接

| 现状 | 状态 | 说明 |
| --- | --- | --- |
| `build.yml` 推 `latest` + short-sha | ✅ 已具备 | 只缺 `stable` 通道与 OCI label |
| healthcheck（3000 + 8000 双探测） | ✅ 已具备 | 容器重建后可自动判定成败，`up -d --wait` 会等它 |
| 数据库外部 PostgreSQL | ✅ 有利 | 容器更新不动数据；附件本体也在库里，重装容器无状态丢失风险 |
| 单容器跑 SSR + API | ⚠️ 注意 | 更新时整体重启，Nuxt 冷启有几秒不可用；想零停机需把 web/api 拆成两个 service 再 scale |
| CI 不跑自测 | ❌ 缺口 | `build.yml` 已 `load: true` 把镜像拉到本地，push 之前插一步 `python scripts/ci_selftest.py` 就是天然发布门禁 |
| 镜像无版本自证 | ❌ 缺口 | 见 §4.1 |

建议给 cookbook 服务补 `stop_grace_period: 15s`，让重建时连接有机会优雅收尾。

---

## 6. 落地清单

**第一步（只改仓库，零风险）**
1. `build.yml` 的标签逻辑换成 `docker/metadata-action`，加 `stable` 通道与 OCI label。
2. 构建期注入 `VERSION` / `COMMIT`，加 `/api/version` 与启动日志。
3. push 前插入 `scripts/ci_selftest.py` 作为门禁。

**第二步（服务器侧，本机验证后执行）**
4. 加 `docker-socket-proxy` + `nickfedor/watchtower` 编排（独立 compose file 或 profile）。
5. cookbook 服务打 `watchtower.enable=true` label，`COOKBOOK_IMAGE` 默认值改为 `:stable`。
6. 配 `WATCHTOWER_NOTIFICATIONS`，**不要**开 `WATCHTOWER_CLEANUP`。

**第三步（可选进阶）**
7. ✅ **已落地**：`.github/workflows/build.yml` 新增 `deploy` job —— 构建推送成功后 SSH 到
   `DEPLOY_HOST` 执行
   `COMPOSE_FILE=/opt/zspace/docker-compose.yml docker compose up -d --pull always --wait --wait-timeout 300`
   （不 `cd`、失败自动打日志）。待办只剩两件：GitHub 里填
   `DEPLOY_HOST / DEPLOY_USER / DEPLOY_SSH_KEY / DEPLOY_PORT`（见 §3-D），
   以及服务器上 `docker login ghcr.io`（私有包必需）。
   想改成"只在发版时部署"：把该 job 的 `if` 换成 `startsWith(github.ref, 'refs/tags/v')`。
8. 如果以后要零停机：web / api 拆两 service + 多副本 + `--rolling-restart`。

---

## 7. 明确不建议的做法

- ❌ 把 `/var/run/docker.sock` 挂进 cookbook 业务容器（业务漏洞 = 宿主机 root）。
- ❌ 开 `WATCHTOWER_CLEANUP` 却指望能一键回滚（旧镜像已经被删了）。
- ❌ 静默自动更新（没有通知，出问题只能等用户反馈）。
- ❌ 把 socket-proxy 的 2375 端口裸暴露到公网（无 TLS 的 Docker API = 整个宿主机）。
- ❌ 对数据库 / 有状态服务启用自动更新。
- ❌ 依赖已归档的 `containrrr/watchtower`（2025-12-17 起无安全补丁）。

---

## 参考

- Watchtower 分支与迁移：<https://watchtowerdocker.com/blog/watchtower-docker-hub-images.html>
- Watchtower 进阶（socket 代理 / rolling / scope）：<https://watchtowerdocker.com/blog/watchtower-advanced-usage.html>
- WUD：<https://github.com/getwud/wud>
- Diun：<https://crazymax.dev/diun/>
- docker-socket-proxy 权限变量：<https://github.com/Tecnativa/docker-socket-proxy>
