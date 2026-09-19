#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
CI 自测：起隔离 compose 栈（项目名 cookbook-selftest，独立卷），对已构建镜像跑
HTTP 自测（/api/tags、/api/recipes、首页 SSR），测完 down -v 销毁。
生产数据在 PG，本脚本不接触；仅验证镜像自身可启动、接口可达。

用法：python3 scripts/ci_selftest.py <image> [port]
"""

import json
import os
import subprocess
import sys
import time
import urllib.request

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
PROJECT = "cookbook-selftest"


def log(msg: str) -> None:
    print(msg, flush=True)


def fail(msg: str) -> None:
    log(f"[FAIL] {msg}")
    sys.exit(1)


def sh(*args: str, check: bool = True, capture: bool = False, timeout: int = None):
    return subprocess.run(args, cwd=ROOT, check=check,
                          capture_output=capture, text=True, timeout=timeout)


def pick_free_port(preferred: int) -> int:
    import socket
    for port in range(preferred, preferred + 100):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            try:
                s.bind(("127.0.0.1", port))
                return port
            except OSError:
                continue
    fail(f"{preferred}+ 端口均被占用")


def http_get(url: str, timeout: int = 15):
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with opener.open(url, timeout=timeout) as resp:
        return resp.status, resp.read()


def main() -> None:
    if len(sys.argv) < 2:
        fail("用法: ci_selftest.py <image> [port]")
    image = sys.argv[1]
    port = pick_free_port(int(sys.argv[2]) if len(sys.argv) > 2 else 8080)

    def compose(*args: str, env: dict = None, timeout: int = None,
                check: bool = True) -> subprocess.CompletedProcess:
        return sh("docker", "compose", "-f", os.path.join(ROOT, "docker-compose.yml"),
                  "-p", PROJECT, *args, env=env, timeout=timeout, check=check)

    run_env = {**os.environ, "COOKBOOK_IMAGE": image,
               "COOKBOOK_WEB_PORT": str(port), "COOKBOOK_API_PORT": "18080"}
    log(f"起隔离栈 {PROJECT}（宿主机 {port} → 容器 3000）")

    try:
        compose("up", "-d", "--wait", "--wait-timeout", "300", env=run_env)
        base = f"http://127.0.0.1:{port}"
        deadline = time.time() + 120
        # 后端 API
        while True:
            try:
                status, body = http_get(f"{base}/api/tags")
                if status == 200:
                    break
            except Exception:
                if time.time() > deadline:
                    compose("logs", "--tail", "40", env=run_env)
                    fail(f"GET {base}/api/tags 超时")
                time.sleep(2)
        data = json.loads(body)
        if data.get("code") != 0:
            fail(f"/api/tags code={data.get('code')}")
        log(f"  [OK] /api/tags code=0（{len(data.get('data', {}).get('list', []))} 个标签）")

        status, body = http_get(f"{base}/api/recipes?page=1&pageSize=2")
        data = json.loads(body)
        if status != 200 or data.get("code") != 0:
            fail(f"/api/recipes code={data.get('code')}")
        log(f"  [OK] /api/recipes code=0（total={data['data']['total']}）")

        # 前端 SSR
        status, body = http_get(base)
        if status != 200 or b"__nuxt" not in body:
            fail(f"GET / status={status}，未检出 SSR 内容")
        log(f"  [OK] GET / SSR 200（{len(body)} 字节）")
        log("[OK] 自测全部通过")
    finally:
        log("销毁隔离栈…")
        compose("down", "-v", timeout=120, check=False)
        log("隔离栈已销毁（独立卷一并清除）")


if __name__ == "__main__":
    main()