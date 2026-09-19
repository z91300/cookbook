# syntax=docker/dockerfile:1
# ============================================================================
# Cookbook 单镜像（multi-stage）：一个容器内同时运行
#   Nuxt 前端 SSR（node，容器内 3000 端口）与 GoFrame 后端（容器内 8000 端口），
# 前端 /api 反代指向容器回环 127.0.0.1:8000（同一进程组，见 /entrypoint.sh）。
# 宿主机按需映射两个端口（见 docker-compose.yml 的 COOKBOOK_WEB_PORT /
# COOKBOOK_API_PORT）。
# 构建命令（也可直接用 scripts/docker_build_push.py 一键构建+自测+推送）：
#   docker build -t cookbook:latest .
#   换依赖源：docker build --build-arg GOPROXY= --build-arg NPM_REGISTRY= -t cookbook:latest .
# ============================================================================

ARG NODE_IMAGE=node:22-alpine
ARG GO_IMAGE=golang:1.26-alpine
ARG RUNTIME_IMAGE=node:22-alpine

# ----------------------------------------------------------------------------
# 阶段 1：前端构建（pnpm + nuxt build）
# ----------------------------------------------------------------------------
FROM ${NODE_IMAGE} AS web-build
WORKDIR /build

# 默认走国内镜像源，可构建时覆盖：--build-arg GOPROXY= / NPM_REGISTRY=
ARG NPM_REGISTRY=https://registry.npmmirror.com
ENV npm_config_registry=${NPM_REGISTRY}

# corepack 启用 pnpm（lockfile v9，pnpm 10 兼容）
RUN corepack enable && corepack prepare pnpm@10 --activate

# 先拷依赖清单再安装，提升层缓存命中
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts

# 再拷源码构建。合并镜像内前端与后端同容器，
# /api 反代目标固定为容器回环 http://127.0.0.1:8000（nuxt.config.ts 默认值即此），
# 无需外部注入 NITRO_API_UPSTREAM。
COPY web/ ./
RUN pnpm build

# ----------------------------------------------------------------------------
# 阶段 2：后端构建（pgsql 驱动基于 lib/pq，纯 Go，CGO 可关）
# ----------------------------------------------------------------------------
FROM ${GO_IMAGE} AS api-build
WORKDIR /src
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY api/ ./api/
COPY internal/ ./internal/
COPY utility/ ./utility/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cookbook .

# ----------------------------------------------------------------------------
# 阶段 3：合并运行镜像（node 跑前端 SSR，Go 二进制跑后端）
# 注意：数据库在外部 PostgreSQL（链接见 config.yaml）；附件本体存库
# （attachments.content BLOB）；manifest/init.sql 仅随镜像留档。
# ----------------------------------------------------------------------------
FROM ${RUNTIME_IMAGE}
WORKDIR /app
ENV NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=3000 \
    NITRO_PORT=3000

# ca-certificates：连接启用 TLS 的 PostgreSQL 时需要（无则 PG 可能握手失败）
RUN apk add --no-cache ca-certificates

COPY --from=web-build /build/.output ./.output
COPY --from=api-build /out/cookbook ./cookbook
# 真实生产配置含数据库凭据不入 git；构建前把 config.prod.example.yaml 复制为
# config.prod.yaml 并填入实际链接（本地构建），或由 CI 从 GitHub Secret 写入该文件。
COPY config.prod.yaml ./config.yaml
COPY manifest/init.sql ./manifest/init.sql

RUN chmod +x ./cookbook \
    && printf '#!/bin/sh\nset -e\ncd /app\n./cookbook &\nAPI_PID=$!\nnode .output/server/index.mjs &\nWEB_PID=$!\ntrap "kill $API_PID $WEB_PID 2>/dev/null || true" INT TERM EXIT\nwait\n' > /entrypoint.sh \
    && chmod +x /entrypoint.sh

EXPOSE 3000 8000
CMD ["/entrypoint.sh"]
