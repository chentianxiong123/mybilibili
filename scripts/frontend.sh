#!/usr/bin/env bash
# 一键拉起前端：web(Nuxt :3200) + admin(Vite :3100)，不含 flutter / nativescript
# 用法: scripts/frontend.sh {start|stop|status|logs} [web|admin]
set -u
cd "$(dirname "$0")/.."

FRONT="$PWD/mybilibili-front"
LOG_DIR="/tmp/mybilibili-logs"
PID_FILE="$LOG_DIR/front-pids"
mkdir -p "$LOG_DIR"

APPS="web admin"   # web=Nuxt:3200  admin=Vite:3100

port_of() { case "$1" in web) echo 3200 ;; admin) echo 3100 ;; esac; }

start_one() {
    app="$1"
    if kill -0 "$(awk -F= -v n="$app" '$1==n{print $2}' "$PID_FILE" 2>/dev/null)" 2>/dev/null; then
        echo "  $app 已在运行"; return
    fi
    (cd "$FRONT" && nohup pnpm --filter "@mybilibili/$app" dev \
        >> "$LOG_DIR/front-$app.log" 2>&1 & echo "$app=$!" >> "$PID_FILE")
    echo "  $app pid=$! port=$(port_of "$app") log=$LOG_DIR/front-$app.log"
}

start() {
    [ -d "$FRONT/node_modules" ] || { echo "== pnpm install =="; (cd "$FRONT" && pnpm install); }
    : > "$PID_FILE"
    for app in ${1:-$APPS}; do start_one "$app"; done
    sleep 3
    for app in ${1:-$APPS}; do
        curl -sf "http://localhost:$(port_of "$app")" >/dev/null 2>&1 \
            && echo "  $app OK (:$(port_of "$app"))" \
            || echo "  $app 未就绪（首次启动编译较慢），logs: scripts/frontend.sh logs $app"
    done
}

stop() {
    [ -f "$PID_FILE" ] || { echo "无运行记录"; return; }
    while IFS== read -r app pid; do
        pkill -P "$pid" 2>/dev/null; kill "$pid" 2>/dev/null && echo "  $app stopped" || echo "  $app 未在运行"
    done < "$PID_FILE"
    rm -f "$PID_FILE"
}

status() {
    [ -f "$PID_FILE" ] || { echo "未启动"; return; }
    while IFS== read -r app pid; do
        kill -0 "$pid" 2>/dev/null && echo "  ● $app (pid $pid, :$(port_of "$app"))" || echo "  ✗ $app 已退出"
    done < "$PID_FILE"
}

logs() {
    case "${1:-}" in web|admin) exec tail -f "$LOG_DIR/front-$1.log" ;; *) exec tail -f "$LOG_DIR"/front-*.log ;; esac
}

case "${1:-}" in
    start)  shift; start "${1:-}" ;;
    stop)   stop ;;
    status) status ;;
    logs)   shift; logs "${1:-}" ;;
    *)      sed -n '2,5p' "$0" ;;
esac
