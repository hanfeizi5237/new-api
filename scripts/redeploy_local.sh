#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

SERVICE_NAME="new-api"
BINARY_PATH="runtime/bin/new-api"
BACKUP_DIR="backups"
LOG_DIR="logs"
VERSION_FILE="VERSION"
NODE_HEAP_MB="${NODE_HEAP_MB:-8192}"

mkdir -p "$BACKUP_DIR" "$LOG_DIR" "$(dirname "$BINARY_PATH")"

if [[ ! -f "$VERSION_FILE" || ! -s "$VERSION_FILE" ]]; then
  echo "[info] VERSION is empty, generating from git commit"
  git rev-parse --short HEAD > "$VERSION_FILE"
fi

if [[ -f "$BINARY_PATH" ]]; then
  cp -f "$BINARY_PATH" "$BACKUP_DIR/new-api.bin.pre-redeploy-$(date +%F-%H%M%S).bak"
fi

echo "[1/6] build web/default with npm"
pushd web/default >/dev/null
npm install
VITE_REACT_APP_VERSION="$(cat ../../VERSION)" npm run build
popd >/dev/null

echo "[2/6] build web/classic with npm legacy peer deps"
pushd web/classic >/dev/null
npm install --legacy-peer-deps
NODE_OPTIONS="--max-old-space-size=${NODE_HEAP_MB}" \
VITE_REACT_APP_VERSION="$(cat ../../VERSION)" \
npm run build
popd >/dev/null

echo "[3/6] build go binary"
go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o "$BINARY_PATH"

echo "[4/6] restart systemd service: ${SERVICE_NAME}"
systemctl restart "$SERVICE_NAME"

echo "[5/6] wait for service"
sleep 5

STATUS_JSON="$(curl -fsS http://127.0.0.1:3000/api/status)"

echo "[6/6] verification"
systemctl is-active "$SERVICE_NAME"
systemctl show "$SERVICE_NAME" -p ExecMainPID -p ActiveEnterTimestamp
printf '%s\n' "$STATUS_JSON"

echo "[done] redeploy completed"
