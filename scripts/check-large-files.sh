#!/usr/bin/env bash
# =============================================================================
# 大文件门禁：拒绝提交任何 >5MB 的 staged 文件
# 用法:
#   bash scripts/check-large-files.sh                 # 检查已 staged
#   bash scripts/check-large-files.sh --install       # 装成 pre-commit hook
#   bash scripts/check-large-files.sh --check-remote  # 检查远程已存在的大文件
#
# 背景：2026-09-20 commit 685b284d 一次性提交了 ~280MB Go 编译产物
#       + BadApple.mp4，把仓库撑到 170M。此 hook 是补漏。
# =============================================================================
set -euo pipefail

THRESHOLD_MB="${LARGE_FILE_THRESHOLD_MB:-5}"

if [[ "${1:-}" == "--install" ]]; then
  HOOK=".git/hooks/pre-commit"
  cat > "$HOOK" <<'EOF'
#!/usr/bin/env bash
# Auto-installed by scripts/check-large-files.sh --install
exec bash "$(git rev-parse --show-toplevel)/scripts/check-large-files.sh" "$@"
EOF
  chmod +x "$HOOK"
  echo "✓ pre-commit hook installed: $HOOK"
  exit 0
fi

if [[ "${1:-}" == "--check-remote" ]]; then
  echo "→ 检查远程 main 上 >${THRESHOLD_MB}MB 的文件..."
  git ls-remote origin main >/dev/null 2>&1 || {
    echo "✗ 无法访问 origin"; exit 1
  }
  # 从 HEAD 取文件列表
  MAX_BYTES=$((THRESHOLD_MB * 1024 * 1024))
  FOUND=0
  while IFS=$'\t' read -r sha size path; do
    [[ "$path" == -* ]] && path="${path:1}"
    if [[ "$size" -gt "$MAX_BYTES" ]]; then
      echo "✗ $path: $(numfmt --to=iec "$size")"
      FOUND=1
    fi
  done < <(git ls-tree -r -l HEAD | awk '{print $3"\t"$4"\t"$5}')
  if [[ "$FOUND" -eq 1 ]]; then
    echo ""
    echo "远程仓库存在大文件，建议 git filter-repo 清理。"
    exit 1
  fi
  echo "✓ 远程仓库无大文件"
  exit 0
fi

# 默认行为：检查 staged 文件
MAX_BYTES=$((THRESHOLD_MB * 1024 * 1024))
FOUND=0

# 用 git diff 找出新增/修改的 staged 文件名
while IFS= read -r -d '' path; do
  if [[ -f "$path" ]]; then
    size=$(stat -c '%s' "$path" 2>/dev/null || echo 0)
    if [[ "$size" -gt "$MAX_BYTES" ]]; then
      human=$(numfmt --to=iec "$size" 2>/dev/null || echo "${size}B")
      echo "✗ $path: $human (阈值 ${THRESHOLD_MB}MB)"
      FOUND=1
    fi
  fi
done < <(git diff --cached --name-only -z)

if [[ "$FOUND" -eq 1 ]]; then
  echo ""
  echo "拒绝提交：大文件会让仓库膨胀。"
  echo "如需排除，请用 git filter-repo 从历史中清理，或确认是必要的资源文件后修改阈值："
  echo "  LARGE_FILE_THRESHOLD_MB=20 git commit ..."
  exit 1
fi

exit 0