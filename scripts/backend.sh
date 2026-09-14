#!/usr/bin/env bash
# 一键拉起 Go 后端全家桶：基础设施(docker) → 编译 → 8 个服务后台运行
# 用法: scripts/backend.sh {start|stop|status|logs|docker} [服务名] [--migrate]
#   start            拉起 infra + 编译 + 裸跑全部服务（开发用）
#   stop             停掉全部裸跑服务
#   status           各服务存活状态
#   logs [name]      跟踪日志，name ∈ core search msg-danmaku live ai studio work bili，缺省全跟
#   docker up        容器化: 构建镜像 + compose 起全部(infra+后端8服务)
#   docker down      停掉全部容器
#   docker build     只构建镜像，不启动
#   docker logs [n]   跟踪容器日志
#   docker status    容器存活状态
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
# bili-proxy 服务：BILI_SESSDATA 可选（提升清晰度档位），BILI_PROXY_ADDR 供 core 反代（默认 http://127.0.0.1:8091）
export BILI_SESSDATA="${BILI_SESSDATA:-}"
export BILI_PROXY_ADDR="${BILI_PROXY_ADDR:-http://127.0.0.1:8091}"

SERVICES="core search msg-danmaku live ai studio work bili transcoder"

bin_of() {
    case "$1" in
        core)       echo /tmp/mybilibili-core ;;
        search)     echo /tmp/mybilibili-search ;;
        msg-danmaku) echo /tmp/mybilibili-msg-danmaku ;;
        live)       echo /tmp/mybilibili-live ;;
        ai)         echo /tmp/mybilibili-ai ;;
        studio)     echo /tmp/mybilibili-studio ;;
        work)       echo /tmp/mybilibili-work ;;
        # transcoder: 实际二进制按 host 硬件自动选（缓存一次）
        transcoder)
            if [ -z "$TRANSCODER_BIN" ]; then
                TRANSCODER_VARIANT=$(pick_transcoder_bin)
                case "$TRANSCODER_VARIANT" in
                    nvenc) TRANSCODER_BIN=/tmp/mybilibili-transcoder-nvenc ;;
                    vaapi) TRANSCODER_BIN=/tmp/mybilibili-transcoder-vaapi ;;
                    *)     TRANSCODER_BIN=/tmp/mybilibili-transcoder ;;
                esac
            fi
            echo "$TRANSCODER_BIN"
            ;;
        bili)       echo /tmp/mybilibili-bili ;;
    esac
}

# 自动检测 transcoder 硬件加速变体
#   优先级: NVIDIA(NVENC) > AMD/Intel(VAAPI) > 软编 fallback
#   由机器实际硬件 + 已编译的二进制决定，单机开发一条命令拉起
pick_transcoder_bin() {
    # 1) NVIDIA: nvidia-smi 在 + lspci 看到 NVIDIA 设备 + NVENC 二进制已编
    if command -v nvidia-smi >/dev/null 2>&1 \
       && lspci 2>/dev/null | grep -qi 'nvidia' \
       && [ -x /tmp/mybilibili-transcoder-nvenc ]; then
        echo "nvenc"
        return
    fi
    # 2) VAAPI: /dev/dri/renderD128 存在 (Intel/AMD) + VAAPI 二进制已编
    if [ -e /dev/dri/renderD128 ] \
       && [ -x /tmp/mybilibili-transcoder-vaapi ]; then
        echo "vaapi"
        return
    fi
    # 3) 软编 fallback
    echo "soft"
}

infra_up() {
    docker network inspect mylib >/dev/null 2>&1 || docker network create mylib
    docker compose -f "$ROOT/dev/docker-compose.yml" up -d
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
        case "$s" in
            transcoder)
                # transcoder 裸跑宿主机: 用系统 ffmpeg + 硬件加速变体(由 pick_transcoder_bin 决定)
                if [ ! -x "$bin" ]; then
                    echo "  ! transcoder 二进制不存在: $bin"
                    echo "    编译: make -C mybilibili-go build-transcoder   (或 -vaapi / -nvenc)"
                    echo "    当前 host 硬件能力: $(pick_transcoder_bin 2>/dev/null || echo unknown)"
                    continue
                fi
                echo "  transcoder 硬件加速: $TRANSCODER_VARIANT ($bin)"
                HTTP_ADDR=:8092 \
                MINIO_ENDPOINT="${MINIO_ENDPOINT:-127.0.0.1:9000}" \
                nohup "$bin" >> "$LOG_DIR/$s.log" 2>&1 &
                ;;
            *)
                nohup "$bin" >> "$LOG_DIR/$s.log" 2>&1 &
                ;;
        esac
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

# ====== 容器化模式 ======
COMPOSE_FILE="$ROOT/dev/docker-compose.yml"

docker_up() {
    echo "== 容器化: 构建镜像 + 启动全部 =="
    docker compose -f "$COMPOSE_FILE" up -d --build
    echo "== 等待 PostgreSQL 就绪 =="
    until docker exec pg16 pg_isready -U postgres -d mybilibili >/dev/null 2>&1; do
        sleep 1; echo -n .
    done
    echo " OK"
    sleep 2
    curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1 \
        && echo "核心链路 OK (:8080)" || echo "提示: :8080 尚未响应，看日志 docker logs"
}

docker_down() {
    echo "== 停掉全部容器 =="
    docker compose -f "$COMPOSE_FILE" down
}

docker_build() {
    echo "== 构建全部后端镜像 =="
    docker compose -f "$COMPOSE_FILE" build
}

docker_status() {
    echo "== 容器状态 =="
    docker compose -f "$COMPOSE_FILE" ps
}

docker_logs() {
    target="${1:-}"
    if [ -n "$target" ]; then
        docker logs -f "mybilibili-$target"
    else
        docker compose -f "$COMPOSE_FILE" logs -f --tail=50
    fi
}

case "${1:-}" in
    start)  shift; start "$@" ;;
    stop)   stop ;;
    status) status ;;
    logs)   shift; logs "${1:-}" ;;
    docker)
        case "${2:-}" in
            up)      docker_up ;;
            down)    docker_down ;;
            build)   docker_build ;;
            status)  docker_status ;;
            logs)    shift 2; docker_logs "${1:-}" ;;
            *)       echo "用法: scripts/backend.sh docker {up|down|build|status|logs [服务名]}" ;;
        esac
        ;;
    *)      sed -n '2,12p' "$0" ;;
esac
