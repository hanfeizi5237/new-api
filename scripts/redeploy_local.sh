#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

SERVICE_NAME="new-api"
BINARY_PATH="runtime/bin/new-api"
BACKUP_DIR="backups"
LOG_DIR="logs"
VERSION_FILE="VERSION"

mkdir -p "$BACKUP_DIR" "$LOG_DIR" "$(dirname "$BINARY_PATH")"

if [[ ! -f "$VERSION_FILE" || ! -s "$VERSION_FILE" ]]; then
  echo "[info] VERSION is empty, generating from git commit"
  git rev-parse --short HEAD > "$VERSION_FILE"
fi

if [[ -f "$BINARY_PATH" ]]; then
  cp -f "$BINARY_PATH" "$BACKUP_DIR/new-api.bin.pre-redeploy-$(date +%F-%H%M%S).bak"
fi

echo "[1/3] build unified web with bun (rsbuild)"
pushd web >/dev/null
bun install --frozen-lockfile
VITE_REACT_APP_VERSION="$(cat ../VERSION)" bun run build
popd >/dev/null

echo "[2/3] build go binary"
go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o "$BINARY_PATH"

echo "[3/3] restart systemd service: ${SERVICE_NAME}"
systemctl restart "$SERVICE_NAME"

sleep 5

echo ""
echo "=== verification ==="
systemctl is-active "$SERVICE_NAME"
systemctl show "$SERVICE_NAME" -p ExecMainPID -p ActiveEnterTimestamp
curl -fsS http://127.0.0.1:3000/api/status || echo 'status check failed'

echo ""
echo "[done] redeploy completed"
