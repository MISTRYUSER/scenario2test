#!/bin/zsh
set -euo pipefail

ROOT="/Users/wentao.xue/Project"
INTEGRATIONS="$ROOT/integrations"

mkdir -p "$INTEGRATIONS/.venvs"

if [ ! -d "$INTEGRATIONS/AUTOTEST" ]; then
  git clone https://github.com/mindfiredigital/AUTOTEST.git "$INTEGRATIONS/AUTOTEST"
fi

python3 -m venv "$INTEGRATIONS/.venvs/autotest"
"$INTEGRATIONS/.venvs/autotest/bin/python" -m pip install --upgrade pip
"$INTEGRATIONS/.venvs/autotest/bin/python" -m pip install -r "$INTEGRATIONS/AUTOTEST/selenium-based-llm-model/requirements.txt"

cat <<'EOF'
Integration setup complete for:
- AUTOTEST
EOF
