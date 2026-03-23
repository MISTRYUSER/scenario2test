#!/bin/zsh
set -euo pipefail

ROOT="/Users/wentao.xue/Project/scenario2test"
GO_BIN="/Users/wentao.xue/sdk/go/bin/go"
NODE_BIN_DIR="/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin"
SERVER_LOG="$(mktemp)"
PORT="${1:-8090}"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

cd "$ROOT"

PATH="/Users/wentao.xue/sdk/go/bin:$PATH" "$GO_BIN" test ./...
PATH="$NODE_BIN_DIR:$PATH" npm --prefix "$ROOT/web" test
PATH="$NODE_BIN_DIR:$PATH" npm --prefix "$ROOT/web" run build
PATH="/Users/wentao.xue/sdk/go/bin:$PATH" "$GO_BIN" run "$ROOT/cmd/server" --listen ":$PORT" --config "$ROOT/configs/config.example.yaml" --web-dist "$ROOT/web/dist" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

for _ in {1..30}; do
  if curl -sf "http://127.0.0.1:$PORT/healthz" >/dev/null; then
    break
  fi
  sleep 1
done

curl -sf "http://127.0.0.1:$PORT/healthz" | grep -q '^ok$'
curl -sf "http://127.0.0.1:$PORT/" | grep -q 'Scenario2Test Platform'
curl -sf -X POST "http://127.0.0.1:$PORT/generate" -H 'Content-Type: application/x-yaml' --data-binary @"$ROOT/examples/login.yaml" \
  | grep -q '"scenario":"login flow"'
cat <<'YAML' | curl -sf -X POST "http://127.0.0.1:$PORT/generate" -H 'Content-Type: application/x-yaml' --data-binary @- \
  | grep -q '"start_url":"https://mall.com"'
scenario:
  name: "E-Commerce Checkout Path Discovery"
  steps:
    - action: "open_url"
      name: "访问首页"
      params: { url: "https://mall.com" }
      next: "check_auth"
    - action: "conditional_branch"
      name: "检查登录状态"
      branches:
        - condition: "auth == 'unlogged'"
          next: "user_login"
        - condition: "auth == 'logged'"
          next: "search_item"
    - action: "user_login"
      name: "登录流程"
      params: { user: "demo", pwd: "secret" }
      next: "search_item"
    - action: "search_item"
      name: "搜索商品"
      params: { keyword: "MacBook" }
      next: "add_to_cart"
    - action: "add_to_cart"
      name: "加入购物车"
      next: "terminal_confirm"
    - action: "terminal_confirm"
      name: "结算并断言"
      type: "end"
YAML

echo "verification passed on port $PORT"
