#!/bin/sh
# 开发模式 entrypoint — 用 air 热重载
# 根据 SERVICE 环境变量动态生成专属 .air-{service}.toml
# 多个容器共享 /app 卷时不会互相覆盖

set -eu

# 从 SERVICE 环境变量获取目录名（与 services/<name>/ 一致）
SVC="${SERVICE}"
CMD_DIR="${CMD_DIR:-$SVC}"
AIR_DIR="/app/.air"
AIR_FILE="${AIR_DIR}/${SVC}.toml"
TMP_DIR="tmp-${SVC}"

# 确保目录存在
mkdir -p "/app/${TMP_DIR}" "${AIR_DIR}"

# 动态生成专属 .air.toml
cat > "${AIR_FILE}" << EOF
root = "."
tmp_dir = "${TMP_DIR}"

[build]
  bin = "./${TMP_DIR}/main"
  cmd = "PLACEHOLDER_CMD"
  delay = 1000
  exclude_dir = ["tmp", "tmp-core", "tmp-search", "tmp-msg-danmaku", "tmp-live", "tmp-ai", "tmp-studio", "tmp-work", "tmp-bili-proxy", "vendor", "node_modules", "old", "web", "mobile", "scripts", "deploy", "docs", "sql", ".git", "proto"]
  exclude_regex = ["_test.go$"]
  include_ext = ["go", "toml", "yaml"]
  kill_delay = "0s"
  send_interrupt = false
  stop_on_error = true

[color]
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = false
EOF

# 用 sed 替换 cmd 中的占位符（支持 services/ 和 external/ 目录）
if [ -d "/app/external/${SERVICE}" ]; then
  sed -i "s|PLACEHOLDER_CMD|cd external/${SERVICE} \&\& go build -buildvcs=false -o ../../${TMP_DIR}/main ./cmd/${CMD_DIR}|" "${AIR_FILE}"
else
  sed -i "s|PLACEHOLDER_CMD|cd services/${SERVICE} \&\& go build -buildvcs=false -o ../../${TMP_DIR}/main ./cmd/${CMD_DIR}|" "${AIR_FILE}"
fi

echo "▶ Starting air for ${SERVICE} (cmd: ${CMD_DIR}, config: ${AIR_FILE})"
exec air -c "${AIR_FILE}"
