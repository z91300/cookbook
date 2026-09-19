#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Cookbook 一键 Docker 构建 / 自测 / 推送
========================================

把前后端打成同一个镜像（cookbook：容器内同时跑 Nuxt SSR 与 GoFrame 后端，
暴露 3000 / 8000 两端口），起隔离 compose 栈做 HTTP 端到端自测，
通过后按指定版本推送。不依赖 .env 文件。

用法（仓库根目录执行）：
    python scripts/docker_build_push.py -v 1.0.0
    python scripts/docker_build_push.py -v 1.0.0 -r registry.example.com/myns
    python scripts/docker_build_push.py -v 1.0.0 -r registry.example.com/myns --no-test
    python scripts/docker_build_push.py --no-build -v 1.0.2 -r registry.example.com/myns

参数：
    -v, --version     镜像版本 tag（默认：v + 时间戳，如 v20260915-1830）
    -r, --registry    镜像仓库前缀（如 registry.example.com/namespace；
                      推送必填，仅本地构建可不填）
    --no-build        跳过构建（镜像已存在，只做自测/推送）
    --no-test         跳过自测
    --no-push         跳过推送（只构建+自测）
    --no-latest       不额外打 latest 标签 / 不推 latest
    --port N          自测占用的宿主机端口（默认 8080，被占用时自动顺延到 8099）
    --keep            自测后不销毁测试栈（便于人工查看，容器名前缀 cookbook-selftest）
    --platform P      构建平台（默认本机原生；跨架构部署时指定，如 linux/amd64）
    --no-cache        构建时不使用缓存
    --go-proxy URL    Go 模块代理（默认 https://goproxy.cn,direct）
    --npm-registry URL npm 源（默认 https://registry.npmmirror.com）

自测内容（全部通过才算成功）：
    1. compose 栈健康（单容器 healthcheck 通过：3000 前端 + 8000 后端都通）
    2. GET /api/tags        → 信封 code=0，预置标签 ≥ 16 个（经前端 /api 反代打到同容器后端）
    3. GET /api/recipes     → 信封 code=0，分页结构完整
    4. GET /                → 200 且为 Nuxt SSR 渲染的 HTML

自测使用独立 compose 项目名 cookbook-selftest 与独立数据卷，测完自动销毁，
不影响 `docker compose up` 起的生产栈与数据。
"""

import argparse
import json
import os
import shutil
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
COMPOSE_FILE = os.path.join(ROOT, "docker-compose.yml")
TEST_PROJECT = "cookbook-selftest"

# --dry-run 时置位：只打印将要执行的命令，不真正执行（无 Docker 引擎也能自检脚本逻辑）
DRY_RUN = False

# Windows 下 docker 不在 PATH 时的常见安装位置
_DOCKER_FALLBACKS = [
    r"C:\Program Files\Docker\Docker\resources\bin\docker.exe",
    os.path.expandvars(
        r"%LOCALAPPDATA%\Programs\DockerDesktop\resources\bin\docker.exe"
    ),
]


def log(msg: str) -> None:
    print(msg, flush=True)


def step(title: str) -> None:
    log("")
    print("=" * 62, flush=True)
    log(f"== {title}")
    print("=" * 62, flush=True)


def fail(msg: str) -> "None":
    log(f"[FAIL] {msg}")
    sys.exit(1)


# ---------------------------------------------------------------------------
# docker 可执行文件定位
# ---------------------------------------------------------------------------

def find_docker() -> str:
    docker = shutil.which("docker")
    if docker:
        return docker
    for path in _DOCKER_FALLBACKS:
        if os.path.isfile(path):
            return path
    fail(
        "找不到 docker 可执行文件：既不在 PATH，常见安装位置也不存在。\n"
        "      请确认已安装 Docker Desktop，或把 docker 加入 PATH 后重试。"
    )


def sh(
    cmd: list, *, capture: bool = False, env: dict = None, check: bool = True,
    timeout: int = None, quiet: bool = False,
) -> subprocess.CompletedProcess:
    """执行外部命令；capture 时收集输出，否则直接透传到终端（构建日志可实时看）。"""
    prefix = "[dry-run] " if DRY_RUN else ""
    if not quiet:
        log(f"$ {prefix}{' '.join(str(c) for c in cmd)}")
    if DRY_RUN:
        return subprocess.CompletedProcess(cmd, 0, stdout="0.0.0-dry-run",
                                           stderr="")
    kwargs = dict(env=env)
    if capture:
        kwargs.update(stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                      encoding="utf-8", errors="replace")
    if timeout:
        kwargs["timeout"] = timeout
    proc = subprocess.run(cmd, **kwargs)
    if check and proc.returncode != 0:
        err = ""
        if capture:
            err = (proc.stderr or "").strip()[:1000]
        fail(f"命令执行失败（退出码 {proc.returncode}）：\n      {err}")
    return proc


def compose(docker: str, *args, env: dict = None, capture: bool = False,
            timeout: int = None, check: bool = True,
            quiet: bool = False) -> subprocess.CompletedProcess:
    return sh([docker, "compose", "-f", COMPOSE_FILE, *args], env=env,
              capture=capture, timeout=timeout, check=check, quiet=quiet)


# ---------------------------------------------------------------------------
# 自测
# ---------------------------------------------------------------------------

def pick_free_port(preferred: int) -> int:
    """优先用 preferred，被占用则在 8080..8099 里找空闲端口。"""
    for port in [preferred] + [p for p in range(8080, 8100) if p != preferred]:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            try:
                s.bind(("127.0.0.1", port))
                return port
            except OSError:
                continue
    fail("8080-8099 端口均被占用，请用 --port 指定其他端口")


def http_get_json(url: str, timeout: int = 15) -> tuple:
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with opener.open(url, timeout=timeout) as resp:
        return resp.status, json.loads(resp.read().decode("utf-8"))


def http_get_text(url: str, timeout: int = 15) -> tuple:
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with opener.open(url, timeout=timeout) as resp:
        return resp.status, resp.read()


def run_selftest(docker: str, image: str, port: int, keep: bool) -> None:
    step(f"自测：起隔离栈 cookbook-selftest（宿主机端口 {port} → 容器 3000）")
    env = dict(os.environ)
    env.update({
        "COOKBOOK_IMAGE": image,
        "COOKBOOK_WEB_PORT": str(port),
    })

    # 残留的同名测试栈先清掉（上次 --keep 或中断留下的）
    compose(docker, "-p", TEST_PROJECT, "down", "-v", "--remove-orphans",
            env=env, check=False, quiet=True)

    log("启动容器（等待健康检查通过，检查 3000 前端 + 8000 后端）……")
    compose(docker, "-p", TEST_PROJECT, "up", "-d", "--wait",
            "--wait-timeout", "300", env=env)

    checks = []

    def record(name: str, ok: bool, detail: str) -> None:
        checks.append(ok)
        log(f"  [{'OK' if ok else 'FAIL'}] {name}  {detail}")

    if DRY_RUN:
        for name in ("/api/tags", "/api/recipes", "GET /"):
            record(name, True, "dry-run 跳过实际请求")
        compose(docker, "-p", TEST_PROJECT, "down", "-v", "--remove-orphans",
                env=env, quiet=True)
        log("[OK] dry-run 流程走通（未真正起栈）")
        return

    # 自测请求走前端 SSR 端口（经容器内 /api 反代打到同容器的后端 8000）
    base = f"http://127.0.0.1:{port}"

    # 1) /api/tags：经反代打到后端，验证预置标签
    try:
        st, body = http_get_json(f"{base}/api/tags")
        tag_list = (body.get("data") or {}).get("list") or []
        record("/api/tags", st == 200 and body.get("code") == 0 and len(tag_list) >= 16,
               f"code={body.get('code')} 标签数={len(tag_list)}")
    except Exception as exc:
        record("/api/tags", False, f"请求异常: {exc}")

    # 2) /api/recipes：分页结构完整
    try:
        st, body = http_get_json(f"{base}/api/recipes?page=1&pageSize=2")
        data = body.get("data") or {}
        record("/api/recipes", st == 200 and body.get("code") == 0
               and "list" in data and "total" in data,
               f"code={body.get('code')} total={data.get('total')}")
    except Exception as exc:
        record("/api/recipes", False, f"请求异常: {exc}")

    # 3) 首页：Nuxt SSR HTML
    try:
        st, raw = http_get_text(f"{base}/")
        record("GET /", st == 200 and b"<" in raw and b"__nuxt" in raw,
               f"HTTP {st}，长度 {len(raw)} 字节")
    except Exception as exc:
        record("GET /", False, f"请求异常: {exc}")

    if not all(checks):
        log("")
        log("自测未通过，最近容器日志：")
        compose(docker, "-p", TEST_PROJECT, "logs", "--tail", "40", env=env,
                check=False)
        if not keep:
            compose(docker, "-p", TEST_PROJECT, "down", "-v",
                    "--remove-orphans", env=env, check=False, quiet=True)
        fail("自测失败，已保留上方日志用于排查")

    log("")
    log("[OK] 自测全部通过")

    if keep:
        log(f"[KEEP] 测试栈保留运行中，访问 {base} 查看；")
        log(f"       手动清理：docker compose -p {TEST_PROJECT} down -v")
    else:
        compose(docker, "-p", TEST_PROJECT, "down", "-v", "--remove-orphans",
                env=env, quiet=True)
        log("测试栈已销毁（独立卷一并清除，不影响生产数据）")


# ---------------------------------------------------------------------------
# 主流程
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Cookbook 一键 Docker 构建 / 自测 / 推送",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="示例：python scripts/docker_build_push.py -v 1.0.0 -r registry.example.com/myns",
    )
    parser.add_argument("-v", "--version",
                        default="v" + datetime.now().strftime("%Y%m%d-%H%M"),
                        help="镜像版本 tag（默认：v + 时间戳）")
    parser.add_argument("-r", "--registry", default="",
                        help="镜像仓库前缀，如 registry.example.com/namespace（推送必填）")
    parser.add_argument("--no-build", action="store_true", help="跳过构建")
    parser.add_argument("--no-test", action="store_true", help="跳过自测")
    parser.add_argument("--no-push", action="store_true", help="跳过推送")
    parser.add_argument("--no-latest", action="store_true",
                        help="不额外打 / 推送 latest 标签")
    parser.add_argument("--port", type=int, default=8080,
                        help="自测宿主机端口（默认 8080，占用则自动顺延）")
    parser.add_argument("--keep", action="store_true",
                        help="自测后保留测试栈便于人工查看")
    parser.add_argument("--platform", default="",
                        help="构建平台（如 linux/amd64，默认本机原生）")
    parser.add_argument("--no-cache", action="store_true", help="构建不用缓存")
    parser.add_argument("--go-proxy", default="https://goproxy.cn,direct",
                        help="Go 模块代理")
    parser.add_argument("--npm-registry", default="https://registry.npmmirror.com",
                        help="npm 源")
    parser.add_argument("--dry-run", action="store_true",
                        help="只打印将要执行的命令，不实际构建/启动/推送")
    args = parser.parse_args()
    global DRY_RUN
    DRY_RUN = args.dry_run

    tag = args.version.strip()
    if tag.lower() in ("latest", "") or any(c in tag for c in " /:@"):
        fail(f"非法版本 tag：{tag!r}（不可为 latest/空，且不能含空格 / : @）")

    prefix = args.registry.strip().rstrip("/") if args.registry else ""
    if not prefix and not args.no_push:
        log("[提示] 未指定 --registry：本次只构建+自测，推送阶段将跳过")

    base = f"{prefix}/cookbook" if prefix else "cookbook"
    image = f"{base}:{tag}"

    log("Cookbook Docker 一键构建（单容器：前端 SSR 3000 + 后端 API 8000）")
    log(f"  镜像：{image}")
    if not args.no_latest:
        log(f"  附带标签：{base}:latest")

    docker = find_docker()

    # 引擎连通性
    step("检查 Docker 引擎")
    ver = sh([docker, "version", "--format", "{{.Server.Version}}"], capture=True,
              check=False)
    if ver.returncode != 0:
        fail("Docker 引擎未运行，请先启动 Docker Desktop 再重试")
    log(f"[OK] Docker 引擎 {ver.stdout.strip()}")

    # 构建
    if not args.no_build:
        step("构建镜像")
        build_env = dict(os.environ)
        build_env.update({
            "COOKBOOK_IMAGE": image,
            "GOPROXY": args.go_proxy,
            "NPM_REGISTRY": args.npm_registry,
        })
        build_cmd = ["build"]
        if args.platform:
            build_cmd += ["--platform", args.platform]
        if args.no_cache:
            build_cmd += ["--no-cache"]
        t0 = time.time()
        compose(docker, *build_cmd, env=build_env)
        log(f"[OK] 构建完成，耗时 {time.time() - t0:.0f}s")

        if not args.no_latest:
            sh([docker, "tag", image, f"{base}:latest"])
            log("[OK] 已打 latest 标签")

    # 自测
    if not args.no_test:
        port = pick_free_port(args.port)
        run_selftest(docker, image, port, args.keep)

    # 推送
    if not args.no_push:
        step("推送镜像")
        if not prefix:
            log("[SKIP] 未指定 --registry，跳过推送（本地镜像已就绪）")
        else:
            tags = [tag] + ([] if args.no_latest else ["latest"])
            for t in tags:
                sh([docker, "push", f"{base}:{t}"])
            log(f"[OK] 已推送 {len(tags)} 个标签到 {prefix}")

    # 收尾
    step("完成")
    log(f"  镜像：{image}")
    if not args.no_latest:
        log("        以及 :latest")
    if prefix and not args.no_push:
        log("")
        log("服务器部署：")
        log(f"  docker pull {image}")
        log("  COOKBOOK_IMAGE=%s \\" % image)
        log("  docker compose up -d --wait")
        log("  （数据在外部 PostgreSQL；卷 api-manifest 只存附件，升级不丢）")


if __name__ == "__main__":
    main()
