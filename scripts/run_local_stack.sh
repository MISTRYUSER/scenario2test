#!/bin/zsh
set -euo pipefail

ROOT="/Users/wentao.xue/Project/scenario2test"
GO_BIN="/Users/wentao.xue/sdk/go/bin/go"
NODE_BIN_DIR="/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin"

PATH="$NODE_BIN_DIR:$PATH" npm run build --prefix "$ROOT/web"
PATH="/Users/wentao.xue/sdk/go/bin:$PATH" "$GO_BIN" run "$ROOT/cmd/server" --listen :8080 --config "$ROOT/configs/config.local.yaml" --web-dist "$ROOT/web/dist"
