#!/bin/zsh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP_PORT=8080
STATIC_PORT=8001
MOCK_PORT=11434

if command -v go >/dev/null 2>&1; then
  GO_BIN="$(command -v go)"
elif [[ -x "$HOME/sdk/go/bin/go" ]]; then
  GO_BIN="$HOME/sdk/go/bin/go"
else
  echo "go not found. Install Go or export GO_BIN." >&2
  exit 1
fi

if command -v npm >/dev/null 2>&1; then
  NPM_BIN="$(command -v npm)"
elif [[ -x "$HOME/.nvm/versions/node/v24.14.0/bin/npm" ]]; then
  NPM_BIN="$HOME/.nvm/versions/node/v24.14.0/bin/npm"
else
  echo "npm not found. Install Node.js or export NPM_BIN." >&2
  exit 1
fi

export PATH="$(dirname "$NPM_BIN"):$PATH"

if command -v python3 >/dev/null 2>&1; then
  PYTHON_BIN="$(command -v python3)"
else
  echo "python3 not found." >&2
  exit 1
fi

MOCK_LOG="$(mktemp -t scenario2test_mock)"
STATIC_LOG="$(mktemp -t scenario2test_static)"
SERVER_LOG="$(mktemp -t scenario2test_server)"
STARTED_PIDS=()

cleanup() {
  for pid in "${STARTED_PIDS[@]:-}"; do
    if [[ -n "${pid}" ]]; then
      kill "${pid}" >/dev/null 2>&1 || true
      wait "${pid}" >/dev/null 2>&1 || true
    fi
  done
}
trap cleanup EXIT INT TERM

echo "Building frontend bundle..."
"$NPM_BIN" --prefix "$ROOT/web" run build >/dev/null

if curl -sf "http://127.0.0.1:${MOCK_PORT}/healthz" >/dev/null; then
  echo "Reusing mock LLM on http://127.0.0.1:${MOCK_PORT} ..."
else
  echo "Starting mock LLM on http://127.0.0.1:${MOCK_PORT} ..."
  "$PYTHON_BIN" "$ROOT/scripts/mock_llm.py" >"$MOCK_LOG" 2>&1 &
  MOCK_PID=$!
  STARTED_PIDS+=("$MOCK_PID")
fi

if curl -sf "http://127.0.0.1:${STATIC_PORT}/login.html" >/dev/null; then
  echo "Reusing static example pages on http://127.0.0.1:${STATIC_PORT} ..."
else
  echo "Starting static example pages on http://127.0.0.1:${STATIC_PORT} ..."
  "$PYTHON_BIN" -m http.server "$STATIC_PORT" --directory "$ROOT/examples" >"$STATIC_LOG" 2>&1 &
  STATIC_PID=$!
  STARTED_PIDS+=("$STATIC_PID")
fi

if curl -sf "http://127.0.0.1:${APP_PORT}/healthz" >/dev/null; then
  echo "Reusing Scenario2Test server on http://127.0.0.1:${APP_PORT} ..."
else
  echo "Starting Scenario2Test server on http://127.0.0.1:${APP_PORT} ..."
  "$GO_BIN" run "$ROOT/cmd/server" --listen ":${APP_PORT}" --config "$ROOT/configs/config.integration.yaml" --web-dist "$ROOT/web/dist" >"$SERVER_LOG" 2>&1 &
  SERVER_PID=$!
  STARTED_PIDS+=("$SERVER_PID")
fi

for _ in {1..30}; do
  if curl -sf "http://127.0.0.1:${MOCK_PORT}/healthz" >/dev/null \
    && curl -sf "http://127.0.0.1:${STATIC_PORT}/login.html" >/dev/null \
    && curl -sf "http://127.0.0.1:${APP_PORT}/healthz" >/dev/null; then
    break
  fi
  sleep 1
done

curl -sf "http://127.0.0.1:${MOCK_PORT}/healthz" >/dev/null
curl -sf "http://127.0.0.1:${STATIC_PORT}/login.html" >/dev/null
curl -sf "http://127.0.0.1:${APP_PORT}/healthz" >/dev/null

cat <<EOF

Scenario2Test demo stack is running.

- Platform UI: http://127.0.0.1:${APP_PORT}
- Health check: http://127.0.0.1:${APP_PORT}/healthz
- Example pages: http://127.0.0.1:${STATIC_PORT}
- Mock LLM: http://127.0.0.1:${MOCK_PORT}/healthz

Demo scenarios:
- $ROOT/examples/login.yaml
- $ROOT/examples/signup.yaml
- $ROOT/examples/checkout.yaml
- $ROOT/examples/password-reset.yaml

Logs:
- Go server: $SERVER_LOG
- Static pages: $STATIC_LOG
- Mock LLM: $MOCK_LOG

Press Ctrl+C to stop all services.
EOF

if (( ${#STARTED_PIDS[@]} > 0 )); then
  wait "${STARTED_PIDS[@]}"
fi
