#!/bin/zsh
set -euo pipefail

ROOT="/Users/wentao.xue/Project/scenario2test"
GO_BIN="/Users/wentao.xue/sdk/go/bin/go"
SERVER_LOG="$(mktemp)"
MOCK_LOG="$(mktemp)"
PORT="${1:-8093}"
TMP_SCENARIO="$(mktemp)"
TMP_CONFIG="$(mktemp)"
HTML_PORT="$((PORT + 1))"

cleanup() {
  for pid in "${SERVER_PID:-}" "${MOCK_PID:-}" "${HTML_PID:-}"; do
    if [[ -n "${pid}" ]]; then
      kill "${pid}" >/dev/null 2>&1 || true
      wait "${pid}" >/dev/null 2>&1 || true
    fi
  done
}
trap cleanup EXIT

python3 "$ROOT/scripts/mock_llm.py" >"$MOCK_LOG" 2>&1 &
MOCK_PID=$!

python3 -m http.server "$HTML_PORT" --directory "$ROOT/examples" >/dev/null 2>&1 &
HTML_PID=$!

PATH="/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH" npm --prefix "$ROOT/web" run build >/dev/null
sed "s|http://127.0.0.1:8080|http://127.0.0.1:${HTML_PORT}|g; s|target: /login|target: /login.html|g" "$ROOT/examples/login.yaml" > "$TMP_SCENARIO"
sed "s|port: 8080|port: ${PORT}|g; s|http://127.0.0.1:8080|http://127.0.0.1:${PORT}|g" "$ROOT/configs/config.integration.yaml" > "$TMP_CONFIG"

PATH="/Users/wentao.xue/sdk/go/bin:$PATH" "$GO_BIN" run "$ROOT/cmd/server" --listen ":${PORT}" --config "$TMP_CONFIG" --web-dist "$ROOT/web/dist" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

for _ in {1..30}; do
  if curl -sf http://127.0.0.1:11434/healthz >/dev/null && curl -sf http://127.0.0.1:${PORT}/healthz >/dev/null; then
    break
  fi
  sleep 1
done

curl -sf -X POST "http://127.0.0.1:${PORT}/generate" -H 'Content-Type: application/x-yaml' --data-binary @"$TMP_SCENARIO"
