# syntax=docker/dockerfile:1
# ============================================================================
# Cookbook 单镜像（multi-stage，三个构建阶段 + 精简 alpine 运行时）：
#   阶段 1  pnpm + nuxt build（构建产物 .output，依赖层不进运行时）
#   阶段 2  golang:alpine 编译 Go 二进制（CGO_ENABLED=0 静态链接，~15MB）
#   阶段 3  node:24-alpine 运行时：只拷 .output + 二进制 + 配置
# 一个容器内同时跑 Nuxt SSR（3000）与 GoFrame 后端（8000），前端 /api 反代
# 指向容器回环 127.0.0.1:8000（同一进程组，见 /entrypoint.sh）。
# 宿主机按需映射两端口（见 docker-compose.yml 的 COOKBOOK_WEB_PORT/COOKBOOK_API_PORT）。
# 构建命令（也可直接用 scripts/docker_build_push.py 一键构建）：
#   docker build -t cookbook:latest .
#   换依赖源：docker build --build-arg GOPROXY= --build-arg NPM_REGISTRY= -t cookbook:latest .
# ============================================================================

ARG NODE_IMAGE=node:24-alpine
ARG GO_IMAGE=golang:1.27-alpine

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
# 阶段 2：后端构建（pgsql 驱动基于 lib/pq，纯 Go，CGO 可关 → 静态二进制无运行时依赖）
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
# 阶段 3：运行时（node:24-alpine，~150MB；node_modules 构建依赖不进此层）
# 注意：数据库在外部 PostgreSQL（链接见 config.yaml）；附件本体存库
# （attachments.content BLOB）；manifest/init.sql 仅随镜像留档。
# ----------------------------------------------------------------------------
FROM ${NODE_IMAGE}
WORKDIR /app
ENV NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=3000 \
    NITRO_PORT=3000 \
    API_PORT=8000

# ca-certificates：连接启用 TLS 的 PostgreSQL 时需要（无则 PG 可能握手失败）
RUN apk add --no-cache ca-certificates

COPY --from=web-build /build/.output ./.output
COPY --from=api-build /out/cookbook ./cookbook
# 数据库配置不进镜像：运行时由 DB_TYPE/DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME
# 环境变量装配（internal/cmd/dbenv.go，见 docker-compose.yml），sqlite 模式用 DB_DATA_PATH。
COPY manifest/init.sql ./manifest/init.sql

RUN chmod +x ./cookbook \
    && printf '#!/bin/sh\nset -e\ncd /app\n./cookbook &\nAPI_PID=$!\nnode .output/server/index.mjs &\nWEB_PID=$!\ntrap "kill $API_PID $WEB_PID 2>/dev/null || true" INT TERM EXIT\nwait\n' > /entrypoint.sh \
    && chmod +x /entrypoint.sh

EXPOSE 3000 8000
CMD ["/entrypoint.sh"]