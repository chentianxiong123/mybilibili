#!/usr/bin/env bash
# 从 Go 注释重新生成 services/core/docs/swagger.{json,yaml}。
# docs/ 已入库；改 handler swag 注释后重跑此脚本即可。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/services/core"

if ! command -v swag >/dev/null 2>&1; then
  if [ -x "$HOME/go/bin/swag" ]; then
    export PATH="$HOME/go/bin:$PATH"
  else
    echo "swag 未安装。装: go install github.com/swaggo/swag/cmd/swag@v1.16.3" >&2
    exit 1
  fi
fi

swag init -g cmd/core/main.go --output ./docs --outputTypes json,yaml --parseInternal --parseDepth 5
echo "✓ docs/swagger.{json,yaml} 已重新生成"
