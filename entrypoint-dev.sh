#!/bin/sh
# 开发模式 entrypoint — 用 air 热重载
# 根据 SERVICE 环境变量动态生成 .air.toml

set -eu

# 从 SERVICE 环境变量获取目录名（与 services/<name>/ 一致）
SVC="${SERVICE}"
CMD_DIR="${CMD_DIR:-$SVC}"

# 动态生成 .air.toml（用 'EOF' 防止 shell 展开）
cat > /app/.air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "PLACEHOLDER_CMD"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "node_modules"]
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

# 用 sed 替换 cmd 中的占位符
sed -i "s|PLACEHOLDER_CMD|cd ${SERVICE} \&\& go build -o ../tmp/main ./cmd/${CMD_DIR}|" /app/.air.toml

echo "▶ Starting air for ${SERVICE} (cmd: ${CMD_DIR})"
exec air -c /app/.air.toml
