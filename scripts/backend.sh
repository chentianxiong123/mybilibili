#!/usr/bin/env bash
# 一键拉起 Go 后端全家桶：基础设施(docker) → 编译 → 7 个服务后台运行
# 用法: scripts/backend.sh {start|stop|status|logs} [服务名] [--migrate]
#   start            拉起 infra + 全部服务（--migrate 首次建表用）
#   stop             停掉全部服务
#   status           各服务存活状态
#   logs [name]      跟踪日志，name ∈ core search msg-danmaku live ai studio work，缺省全跟
set -u
cd "$(dirname "$0")/.."

ROOT="$PWD"
GO_DIR="$ROOT/mybilibili-go"
LOG_DIR="/tmp/mybilibili-logs"
PID_FILE="$LOG_DIR/pids"
mkdir -p "$LOG_DIR"

# 基础设施与公共环境（可用环境变量覆盖）
export PG_DSN="${PG_DSN:-postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable}"
export JWT_SECRET="${JWT_SECRET:-dev-secret-change-in-production}"
export MQ_TYPE=nats
export NATS_URL="${NATS_URL:-nats://127.0.0.1:4222}"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"

SERVICES="core search msg-danmaku live ai studio work"

bin_of() {
    case "$1" in
        core)       echo /tmp/mybilibili-core ;;
        search)     echo /tmp/mybilibili-search ;;
        msg-danmaku) echo /tmp/mybilibili-msg-danmaku ;;
        live)       echo /tmp/mybilibili-live ;;
        ai)         echo /tmp/mybilibili-ai ;;
        studio)     echo /tmp/mybilibili-studio ;;
        work)       echo /tmp/mybilibili-work ;;
    esac
}

infra_up() {
    docker network inspect mylib >/dev/null 2>&1 || docker network create mylib
    docker compose -f "$ROOT/infra-compose.yml" up -d
    echo -n "等待 PostgreSQL 就绪 "
    until docker exec pg16 pg_isready -U postgres -d mybilibili >/dev/null 2>&1; do
        sleep 1; echo -n .
    done
    echo " OK"
}

migrate() {
    echo "== 执行 SQL 迁移（逐文件，出错继续）=="
    for f in "$GO_DIR"/sql/*.sql; do
        docker exec -i pg16 psql -v ON_ERROR_STOP=0 -q -U postgres -d mybilibili < "$f" \
            && echo "  ok  $(basename "$f")" || echo "  SKIP $(basename "$f")"
    done
}

start() {
    infra_up
    [ "${1:-}" = "--migrate" ] && migrate
    echo "== 编译 =="
    make -s -C "$GO_DIR" build || exit 1
    echo "== 启动服务 =="
    : > "$PID_FILE"
    for s in $SERVICES; do
        if kill -0 "$(cat "$PID_FILE" | awk -F= -v n="$s" '$1==n{print $2}')" 2>/dev/null; then
            echo "  $s 已在运行"; continue
        fi
        bin=$(bin_of "$s")
        nohup "$bin" >> "$LOG_DIR/$s.log" 2>&1 &
        echo "$s=$!" >> "$PID_FILE"
        echo "  $s pid=$! log=$LOG_DIR/$s.log"
    done
    sleep 1
    curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1 \
        && echo "核心链路 OK (:8080)" || echo "提示: :8080 尚未响应，看日志 scripts/backend.sh logs core"
}

stop() {
    [ -f "$PID_FILE" ] || { echo "无运行记录"; return; }
    while IFS== read -r s pid; do
        kill "$pid" 2>/dev/null && echo "  $s stopped" || echo "  $s 未在运行"
    done < "$PID_FILE"
    rm -f "$PID_FILE"
}

status() {
    [ -f "$PID_FILE" ] || { echo "未启动（无 $PID_FILE）"; return; }
    while IFS== read -r s pid; do
        kill -0 "$pid" 2>/dev/null && echo "  ● $s (pid $pid)" || echo "  ✗ $s (pid $pid 已退出, logs/$s.log)"
    done < "$PID_FILE"
}

logs() {
    target="${1:-}"
    [ -n "$target" ] && exec tail -f "$LOG_DIR/$target.log"
    exec tail -f "$LOG_DIR"/*.log
}

case "${1:-}" in
    start)  shift; start "$@" ;;
    stop)   stop ;;
    status) status ;;
    logs)   shift; logs "${1:-}" ;;
    *)      sed -n '2,8p' "$0" ;;
esac
