#!/usr/bin/env bash
# =============================================================================
# 项目统一测试入口 —— 本地与 CI 跑同一份脚本，禁止两套逻辑。
#
# 用法:
#   ./scripts/test.sh                  # 跑全部 + 覆盖率门禁
#   ./scripts/test.sh backend          # 只跑后端 Go 测试
#   ./scripts/test.sh backend-cov      # 只跑后端 + 覆盖率门禁
#   ./scripts/test.sh frontend         # 只跑前端 vitest
#   ./scripts/test.sh integration      # 只跑集成测试
#   ./scripts/test.sh contract         # 只跑契约测试 (OpenAPI 3.1 + 打真后端)
#   ./scripts/test.sh e2e              # 只跑 Python E2E（需本地有后端）
#
# 退出码: 0=全过, 1=有失败, 2=覆盖率低于阈值
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# 默认阈值（CI 与本地一致）
# 当前加权覆盖率 ~67.1%（修完 manuscript/favorite 历史 fail 后）
# TODO 后续加测试覆盖到 70%+
COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-65}"

# 颜色
if [ -t 1 ]; then
  RED=$'\033[0;31m'; GREEN=$'\033[0;32m'; YEL=$'\033[0;33m'; NC=$'\033[0m'
else
  RED=''; GREEN=''; YEL=''; NC=''
fi

step() { printf "\n${YEL}==> %s${NC}\n" "$1"; }
pass() { printf "${GREEN}✓ %s${NC}\n" "$1"; }
fail() { printf "${RED}✗ %s${NC}\n" "$1"; exit 1; }

# -----------------------------------------------------------------------------
# 1. 后端 Go 单元测试 + 覆盖率门禁
# -----------------------------------------------------------------------------
backend() {
  step "后端 Go 单元测试"
  local modules=(
    shared/pkg
    services/core
    services/search
    services/msg-danmaku
    services/live
    services/ai
    services/studio
    services/work
    services/bili-proxy
  )
  for d in "${modules[@]}"; do
    if [ -f "$d/go.mod" ]; then
      (cd "$d" && go test -count=1 ./...) || fail "$d 单测失败"
      pass "$d"
    fi
  done
}

backend_cov() {
  step "后端 Go 覆盖率 (加权, 阈值 ${COVERAGE_THRESHOLD}%)"
  local modules=(
    shared/pkg
    services/core
    services/search
    services/msg-danmaku
    services/live
    services/ai
    services/studio
    services/work
    services/bili-proxy
  )

  # 合并每个模块的 coverprofile（加权覆盖率 = 真实覆盖率）
  local merged="/tmp/merged-cov.out"
  echo "mode: count" > "$merged"
  for d in "${modules[@]}"; do
    if [ -f "$d/go.mod" ]; then
      (cd "$d" && go test -count=1 -coverprofile=/tmp/single-cov.out -covermode=count ./... >/dev/null 2>&1) || true
      if [ -s /tmp/single-cov.out ]; then
        # 跳过第一行 "mode: count"
        tail -n +2 /tmp/single-cov.out >> "$merged"
      fi
    fi
  done

  if [ ! -s "$merged" ]; then
    fail "没有任何覆盖率数据生成"
  fi

  # 用 go tool cover 算真实加权覆盖率
  local total
  total=$(go tool cover -func="$merged" 2>/dev/null | tail -1 | awk '{print $3}' | sed 's/%//')
  if [ -z "$total" ]; then
    fail "无法解析覆盖率"
  fi
  local pct_int=${total%.*}
  printf "  总加权覆盖率: %s%% (阈值 %d%%)\n" "$total" "$COVERAGE_THRESHOLD"
  if [ "$pct_int" -lt "$COVERAGE_THRESHOLD" ]; then
    fail "覆盖率 ${pct_int}% < 阈值 ${COVERAGE_THRESHOLD}%"
  fi
  pass "覆盖率门禁通过"
}

# -----------------------------------------------------------------------------
# 2. 前端 vitest
# -----------------------------------------------------------------------------
frontend() {
  step "前端 web/user vitest"
  if [ ! -d web/user ]; then
    echo "  (跳过: web/user 不存在)"
    return
  fi
  (cd web/user && bun run test) || fail "web/user vitest 失败"
  pass "web/user vitest"

  if [ -d web/admin ]; then
    step "前端 web/admin vitest"
    (cd web/admin && bun run test) || fail "web/admin vitest 失败"
    pass "web/admin vitest"
  fi
}

# -----------------------------------------------------------------------------
# 3. 集成测试 (tests/integration)
# -----------------------------------------------------------------------------
integration() {
  step "集成测试 (tests/integration)"
  if [ ! -d tests/integration ]; then
    echo "  (跳过: tests/integration 不存在)"
    return
  fi
  (cd tests/integration && go test -count=1 ./...) || fail "集成测试失败"
  pass "集成测试"
}

# -----------------------------------------------------------------------------
# 4. 契约测试 (OpenAPI 3.1 契约 + 打真后端 + jsonschema 校验)
# -----------------------------------------------------------------------------
contract() {
  step "契约测试 (contracts/openapi/v1.yaml + 打真后端 + jsonschema)"
  if [ ! -f contracts/openapi/v1.yaml ]; then
    echo "  (跳过: contracts/openapi/v1.yaml 不存在)"
    return
  fi
  if ! command -v pytest >/dev/null 2>&1; then
    echo "  (跳过: pytest 未安装)"
    return
  fi
  if ! python3 -c "import openapi_spec_validator, jsonschema, yaml" >/dev/null 2>&1; then
    echo "  (跳过: openapi-spec-validator / jsonschema / pyyaml 未安装)"
    echo "         pip install --break-system-packages openapi-spec-validator jsonschema pyyaml"
    return
  fi
  # 检测后端是否在跑（默认 localhost:8080）
  local api_base="${API_BASE_URL:-http://localhost:8080/api/v1}"
  if ! curl -s -m 2 -o /dev/null -w "%{http_code}" "${api_base%/api/v1}/api/v1/manuscript/hot" 2>/dev/null | grep -q '^[1-5]'; then
    echo "  (跳过: 后端未运行于 ${api_base})"
    return
  fi
  pytest tests/contract/ -v || fail "契约测试失败"
  pass "契约测试"
}

# -----------------------------------------------------------------------------
# 5. Python E2E (playwright + pytest)
# -----------------------------------------------------------------------------
e2e() {
  step "Python E2E (playwright + pytest)"
  if ! command -v pytest >/dev/null 2>&1; then
    echo "  (跳过: pytest 未安装。装: pip install pytest playwright pytest-playwright)"
    return
  fi
  if ! python3 -c "import playwright" >/dev/null 2>&1; then
    echo "  (跳过: playwright Python 包未安装)"
    return
  fi
  pytest tests/e2e/ -v || fail "Python E2E 失败"
  pass "Python E2E"
}

# -----------------------------------------------------------------------------
# 主流程
# -----------------------------------------------------------------------------
case "${1:-all}" in
  backend)         backend ;;
  backend-cov)     backend; backend_cov ;;
  frontend)        frontend ;;
  integration)     integration ;;
  contract)        contract ;;
  e2e)             e2e ;;
  all)
    backend
    frontend
    integration
    contract
    e2e
    backend_cov
    step "全部测试通过"
    ;;
  *) echo "用法: $0 {all|backend|backend-cov|frontend|integration|contract|e2e}"; exit 2 ;;
esac