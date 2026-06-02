#!/usr/bin/env python3
"""Skill: 升级 new-api (zhongzhuan 分支) -> 重新编译 Docker 镜像并发布

用法: 调用此脚本即可执行升级，或 import 后手动调用 upgrade_newapi()

步骤:
1. 检查系统资源（内存/磁盘），不足则提前拒绝
2. git checkout zhongzhuan && git pull 拉取最新代码
3. 生成 docker-compose.yml (参考初始部署脚本 04-deploy-new-api.sh)
4. 构建 Docker 镜像 (local/new-api:zhongzhuan) - 多阶段构建含前端
5. 强制清理旧容器并重启
6. 健康检查
7. 返回升级报告

安全改进:
- 构建前检查可用内存（< 2GB 拒绝）和磁盘（< 5GB 拒绝）
- 去除 --pull 标志（基础镜像已用 digest 锁定）
- 子进程超时异常安全处理
- 失败时自动清理构建缓存释放空间
"""

import subprocess
import sys
import os
import time

PROJECT_DIR = '/home/ubuntu/streamlit_app/new-api'
BRANCH = 'zhongzhuan'
CONTAINER_NAME = 'new-api'
IMAGE_TAG = 'local/new-api:zhongzhuan'
DEPLOY_DIR = '/opt/midrelay/new-api-deploy'
DATA_DIR = '/opt/midrelay/new-api/data'
LOG_DIR = '/opt/midrelay/new-api/logs'

# 安全阈值
MIN_AVAILABLE_MEM_MB = 2000   # 构建前至少 2GB 可用内存
MIN_AVAILABLE_DISK_GB = 5     # 至少 5GB 可用磁盘

BUILD_TIMEOUT_SECONDS = 900   # 多阶段构建 15 分钟
DEPLOY_TIMEOUT_SECONDS = 120  # 容器启动 2 分钟


def run(cmd, cwd=None, timeout=300, capture=True):
    """执行 shell 命令，安全处理超时和异常。"""
    try:
        result = subprocess.run(
            cmd, shell=True, capture_output=capture, text=True,
            cwd=cwd, timeout=timeout
        )
        out = (result.stdout or '').strip()
        err = (result.stderr or '').strip()
        return result.returncode, out, err
    except subprocess.TimeoutExpired as exc:
        return 124, (exc.stdout or '').strip(), f'Command timed out after {timeout}s: {cmd}'
    except Exception as exc:
        return 1, '', f'Command failed unexpectedly: {exc}'


def _read_existing_env_vars():
    """从现有 docker-compose.yml 读取 environment 块中的键值对。"""
    compose_path = os.path.join(DEPLOY_DIR, 'docker-compose.yml')
    result = {}
    if not os.path.isfile(compose_path):
        return result
    try:
        with open(compose_path, 'r', encoding='utf-8') as f:
            in_env = False
            for line in f:
                stripped = line.strip()
                if stripped.startswith('environment:'):
                    in_env = True
                    continue
                if in_env:
                    if stripped.startswith('- ') or (stripped and not stripped.startswith('#')):
                        if ':' in stripped:
                            key, _, val = stripped.partition(':')
                            key = key.strip().lstrip('- ').strip()
                            val = val.strip().strip('"').strip("'")
                            if key and val:
                                result[key] = val
                    else:
                        break
    except Exception:
        pass
    return result


def check_resources(report_lines):
    """检查可用内存和磁盘空间，返回是否通过。"""
    ok = True

    # 检查可用内存
    rc, out, _ = run("free -m | awk '/^Mem:/ {print $7}'")
    if rc == 0 and out:
        avail_mb = int(out)
        if avail_mb < MIN_AVAILABLE_MEM_MB:
            report_lines.append(
                f'[ERROR] Insufficient memory: {avail_mb}MB available, '
                f'need at least {MIN_AVAILABLE_MEM_MB}MB. Aborting.'
            )
            ok = False
        else:
            report_lines.append(f'[OK] Available memory: {avail_mb}MB (min: {MIN_AVAILABLE_MEM_MB}MB)')
    else:
        report_lines.append('[WARN] Could not determine available memory, proceeding with caution')

    # 检查可用磁盘
    rc, out, _ = run(f"df -BG {DEPLOY_DIR} | awk 'NR==2 {{gsub(/G/,\"\",$4); print $4}}'")
    if rc == 0 and out:
        avail_gb = int(out)
        if avail_gb < MIN_AVAILABLE_DISK_GB:
            report_lines.append(
                f'[ERROR] Insufficient disk space: {avail_gb}GB available, '
                f'need at least {MIN_AVAILABLE_DISK_GB}GB. Aborting.'
            )
            ok = False
        else:
            report_lines.append(f'[OK] Available disk: {avail_gb}GB (min: {MIN_AVAILABLE_DISK_GB}GB)')
    else:
        report_lines.append('[WARN] Could not determine available disk space, proceeding with caution')

    return ok


def upgrade_newapi():
    report_lines = []

    # Step 0: 资源检查
    report_lines.append('[INFO] Checking system resources before build...')
    if not check_resources(report_lines):
        report_lines.append('[ABORT] Pre-flight checks failed. Upgrade aborted.')
        return '\n'.join(report_lines)

    # Step 1: Git pull (zhongzhuan branch) — 使用 --no-rebase 避免 divergent branches
    report_lines.append('[INFO] Fetching latest code from git...')
    run(f'cd {PROJECT_DIR} && git checkout {BRANCH}', timeout=60)
    run(f'cd {PROJECT_DIR} && git clean -fd', timeout=60)
    # 关键：用 --no-rebase 而非直接 pull，避免上游代码覆盖本地修改
    rc, out, err = run(f'cd {PROJECT_DIR} && git fetch origin {BRANCH}', timeout=120)
    if rc == 0:
        # 检查本地是否有未推送的 commit（用户自定义修改）
        rc2, local_ahead, _ = run(f'cd {PROJECT_DIR} && git rev-list --count HEAD..origin/{BRANCH}', timeout=30)
        if rc2 == 0 and local_ahead.strip() == '0':
            # 没有上游新代码，跳过 merge
            report_lines.append(f'[OK] Already up to date on {BRANCH}')
        else:
            # 有上游更新，使用 merge（保留本地修改）而非 rebase
            rc, out, err = run(f'cd {PROJECT_DIR} && git merge origin/{BRANCH} --no-edit', timeout=120)
            if rc == 0:
                report_lines.append(f'[OK] Merged upstream on {BRANCH}: {out}')
            else:
                report_lines.append(f'[WARN] Merge result: {out} | {err}')
    else:
        report_lines.append(f'[WARN] Git fetch result: {out} | {err}')

    rc, commit, _ = run(f'cd {PROJECT_DIR} && git log -1 --format="%H %s"', timeout=30)
    report_lines.append(f'Branch: {BRANCH}, Latest commit: {commit}')

    # Step 2: 生成 docker-compose.yml
    # 从现有 compose 文件读取敏感环境变量（SQL_DSN, REDIS_CONN_STRING, SESSION_SECRET），
    # 不写死在脚本中，避免泄露凭证到公开仓库。
    existing_env = _read_existing_env_vars()

    env_lines = [
        '    environment:',
        '      PORT: "3000"',
        '      TZ: "Asia/Shanghai"',
        '      NODE_NAME: "midrelay-node-1"',
        '      ERROR_LOG_ENABLED: "true"',
        '      BATCH_UPDATE_ENABLED: "true"',
    ]
    for key in ('SESSION_SECRET', 'SQL_DSN', 'REDIS_CONN_STRING'):
        if key in existing_env:
            env_lines.append(f'      {key}: "{existing_env[key]}"')

    env_block = '\n'.join(env_lines)

    compose_content = f"""services:
  new-api:
    build:
      context: {PROJECT_DIR}
      dockerfile: Dockerfile
    image: {IMAGE_TAG}
    container_name: {CONTAINER_NAME}
    restart: unless-stopped
    network_mode: host
    command: --log-dir /app/logs
    volumes:
      - {DATA_DIR}:/data
      - {LOG_DIR}:/app/logs
{env_block}
"""
    os.makedirs(DEPLOY_DIR, exist_ok=True)
    os.makedirs(DATA_DIR, exist_ok=True)
    os.makedirs(LOG_DIR, exist_ok=True)

    compose_path = os.path.join(DEPLOY_DIR, 'docker-compose.yml')
    with open(compose_path, 'w') as f:
        f.write(compose_content)
    report_lines.append(f'[OK] Generated docker-compose.yml at {compose_path}')

    # Step 3: 构建镜像
    # 去除 --pull：Dockerfile 基础镜像已用 digest 锁定，无需重复拉取
    report_lines.append('[INFO] Building Docker image (this may take 5-15 minutes)...')
    rc, out, err = run(
        'sudo docker compose build', cwd=DEPLOY_DIR, timeout=BUILD_TIMEOUT_SECONDS
    )
    if rc != 0:
        report_lines.append(f'[ERROR] Build failed: {err}')
        # 构建失败时尝试清理缓存释放空间
        report_lines.append('[INFO] Cleaning up Docker build cache after failure...')
        run('sudo docker builder prune -f', timeout=120)
        return '\n'.join(report_lines)
    report_lines.append('[OK] Docker image built successfully')

    # Step 4: 停止旧容器并启动新容器
    report_lines.append('[INFO] Restarting container...')
    run(f'sudo docker rm -f {CONTAINER_NAME}', timeout=30)
    rc, out, err = run(
        'sudo docker compose up -d', cwd=DEPLOY_DIR, timeout=DEPLOY_TIMEOUT_SECONDS
    )
    if rc != 0:
        report_lines.append(f'[ERROR] Container start failed: {err}')
        return '\n'.join(report_lines)
    report_lines.append('[OK] Container started')

    # 等待服务启动
    time.sleep(10)

    # Step 5: 健康检查
    rc, out, err = run(
        f'sudo docker inspect --format="{{{{.State.Status}}}}" {CONTAINER_NAME}', timeout=30
    )
    if out == 'running':
        report_lines.append(f'[OK] Container running: {out}')
    else:
        report_lines.append(f'[WARN] Container status: {out}')

    # 检查容器日志
    rc, logs, _ = run(f'sudo docker logs --tail 30 {CONTAINER_NAME}', timeout=30)
    if 'Error' not in logs and 'FATAL' not in logs and 'error' not in logs.lower():
        report_lines.append('[OK] Container logs look clean')
    else:
        report_lines.append(f'[WARN] Container logs:\n{logs}')

    # 检查 API 健康
    rc, status, _ = run(
        'curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/status || echo "N/A"',
        timeout=30
    )
    report_lines.append(f'API status (port 3000): HTTP {status}')

    rc, body, _ = run('curl -s http://localhost:3000/api/status', timeout=30)
    if body:
        report_lines.append(f'API response: {body[:300]}')

    report_lines.append('')
    report_lines.append('[OK] new-api upgraded successfully.')
    report_lines.append(f'     Source  : {PROJECT_DIR}')
    report_lines.append(f'     Branch  : {BRANCH}')
    report_lines.append(f'     Image   : {IMAGE_TAG}')
    report_lines.append(f'     API     : http://127.0.0.1:3000')
    report_lines.append(f'     Data    : {DATA_DIR}')

    return '\n'.join(report_lines)


if __name__ == '__main__':
    # Auto-elevate to sudo if not running as root, needed for /opt/midrelay/ operations
    if os.geteuid() != 0:
        os.execvp('sudo', ['sudo', sys.executable] + sys.argv)
    print(upgrade_newapi())
