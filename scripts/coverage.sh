#!/bin/bash
# 汇总所有 Go 模块的测试覆盖率
# 用法: ./scripts/coverage.sh [--html]
#   --html  同时生成 HTML 报告

set -e

COVERAGE_DIR="/tmp/mybilibili-coverage"
HTML_MODE=false

if [ "$1" = "--html" ]; then
  HTML_MODE=true
  mkdir -p "$COVERAGE_DIR/html"
fi

echo "========== Go 模块覆盖率 =========="

TOTAL_LINES=0
TOTAL_COVERED=0

for d in shared/pkg services/core services/search services/ai services/msg-danmaku services/live services/studio services/work services/bili-proxy; do
  if [ -f "$d/go.mod" ]; then
    name=$(echo "$d" | tr '/' '_')
    out="$COVERAGE_DIR/${name}.out"

    result=$(cd "$d" && go test -count=1 -coverprofile="$out" ./... 2>&1 | grep "^ok" || true)

    if [ -n "$result" ]; then
      pct=$(echo "$result" | awk -F'coverage: ' '{print $2}' | awk -F'%' '{s+=$1; n++}END{if(n>0) printf "%.1f", s/n; else print "N/A"}')
      printf "  %-30s %s%%\n" "$d" "$pct"

      if [ "$HTML_MODE" = true ]; then
        go tool cover -html="$out" -o "$COVERAGE_DIR/html/${name}.html" 2>/dev/null
        echo "    → $COVERAGE_DIR/html/${name}.html"
      fi
    else
      printf "  %-30s %s\n" "$d" "(no tests)"
    fi
  fi
done

echo "===================================="
echo "完成。报告目录: $COVERAGE_DIR/html/"
