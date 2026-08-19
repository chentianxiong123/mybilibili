#!/usr/bin/env bash
set -u
ROOT="$(cd "$(dirname "$0")" && pwd)"
FRONT="$ROOT/mybilibili-front"
GO_DIR="$ROOT/mybilibili-go"
PG_DSN="postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"

log() { echo -e "\033[36m[$(date +%H:%M:%S)]\033[0m $*"; }
ok()  { echo -e "  \033[32m✓\033[0m $*"; }
skip(){ echo -e "  \033[33m-\033[0m $*"; }

# ---------- 1. Docker 基础设施 ----------
log "启动 Docker 容器..."
DOCKER_CONTAINERS=(pg16 mybilibili-redis mybilibili-minio nats napcat)
for c in "${DOCKER_CONTAINERS[@]}"; do
  if docker inspect "$c" >/dev/null 2>&1; then
    docker start "$c" >/dev/null 2>&1 && ok "$c" || skip "$c 已运行或启动失败"
  else
    skip "$c 不存在(跳过)"
  fi
done
sleep 3

# ---------- 2. Go 后端 ----------
log "构建/启动 Go 后端..."
cd "$GO_DIR" || exit 1
[[ -x /tmp/mybilibili-core ]]  || go build -ldflags="-s -w" -o /tmp/mybilibili-core   ./core-service/cmd/core
[[ -x /tmp/mybilibili-search ]] || go build -ldflags="-s -w" -o /tmp/mybilibili-search  ./search-service/cmd/search

pkill -f mybilibili-core  2>/dev/null; pkill -f mybilibili-search 2>/dev/null; sleep 1

PG_DSN="$PG_DSN" HTTP_ADDR=":8080" GRPC_ADDR=":9090" setsid nohup /tmp/mybilibili-core  > /tmp/core-prod.log   2>&1 < /dev/null & disown
PG_DSN="$PG_DSN" HTTP_ADDR=":8084" GRPC_ADDR=":9094" setsid nohup /tmp/mybilibili-search > /tmp/search-prod.log 2>&1 < /dev/null & disown
sleep 3

# ---------- 3. 前端 web / admin ----------
log "启动 web/admin..."
fuser -k 3200/tcp 2>/dev/null; fuser -k 3100/tcp 2>/dev/null; sleep 1

(cd "$FRONT/apps/web"   && NITRO_PORT=3200 setsid nohup node .output/server/index.mjs > /tmp/web-prod.log   2>&1 < /dev/null & disown)
(cd "$FRONT/apps/admin" && NITRO_PORT=3100 setsid nohup node .output/server/index.mjs > /tmp/admin-prod.log 2>&1 < /dev/null & disown)
sleep 4

# ---------- 4. 健康检查 ----------
log "健康检查..."
check(){ local name=$1 url=$2; code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 6 "$url" 2>/dev/null); if [[ "$code" == "200" || "$code" == "302" ]]; then ok "$name → $code"; else echo -e "  \033[31m✗\033[0m $name → ${code:-连接失败}"; fi }
check "docker.pg16              " "http://localhost:5432/nonexist" 2>/dev/null || docker exec pg16 pg_isready -U postgres >/dev/null 2>&1 && ok "docker.pg16 (pg_isready)"
docker exec mybilibili-redis redis-cli ping >/dev/null 2>&1 && ok "docker.redis PONG"
check "core  :8080              " "http://localhost:8080/api/v1/health"
check "search:8084              " "http://localhost:8084/api/v1/search/hot"
check "web   :3200              " "http://localhost:3200/"
check "admin :3100              " "http://localhost:3100/"
echo -e "\033[32m\n全部启动完成。\033[0m"