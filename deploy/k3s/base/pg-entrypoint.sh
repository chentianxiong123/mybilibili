#!/bin/bash
# ============================================================
# PG entrypoint: 按 PG_ROLE env 切换 primary / replica
# primary  = 启动 PG + 创建复制用户 + 创建 replication slot
# replica  = pg_basebackup 拉基线 + 写 standby.signal + 启动 PG
# ============================================================
set -e

PG_ROLE="${PG_ROLE:-primary}"
PG_REPL_USER="${PG_REPL_USER:-repluser}"
PG_REPL_PASSWORD="${PG_REPL_PASSWORD:-replpass}"
PG_PRIMARY_HOST="${PG_PRIMARY_HOST:-postgres-primary}"
PG_PRIMARY_PORT="${PG_PRIMARY_PORT:-5432}"
PG_REPL_SLOT="${PG_REPL_SLOT:-replica_slot_1}"

# 替换官方 entrypoint 的最后一步：实际启动 PG
PG_ENTRYPOINT="/usr/local/bin/docker-entrypoint.sh"

primary_init() {
  echo "[pg-entrypoint] PG_ROLE=primary, 启动 PostgreSQL"
  # 先起 PG 让 initdb/恢复能完成
  exec "$PG_ENTRYPOINT" "$@"
}

primary_post_start() {
  # 当 PG 起来后，创建复制账户和 slot
  # 这个函数由 init container 调用或者用 trap 在后台起
  for i in $(seq 1 30); do
    if pg_isready -U "$POSTGRES_USER" >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" <<EOSQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${PG_REPL_USER}') THEN
    CREATE ROLE ${PG_REPL_USER} WITH REPLICATION LOGIN PASSWORD '${PG_REPL_PASSWORD}';
  END IF;
END
\$\$;
SELECT pg_create_physical_replication_slot('${PG_REPL_SLOT}') WHERE NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = '${PG_REPL_SLOT}');
EOSQL

  # 追加 pg_hba 授权：允许集群网段（192.168.100/22 pod 段与 192.168.104/24 svc 段）做 replication
  PG_HBA="$PGDATA/pg_hba.conf"
  if [ -f "$PG_HBA" ] && ! grep -q "replication .*${PG_REPL_USER} 192.168.0.0/16" "$PG_HBA"; then
    cat >> "$PG_HBA" <<EOF
# k3s 集群 pg replication（entrypoint 自动追加）
host all ${PG_REPL_USER} 192.168.0.0/16 md5
host replication ${PG_REPL_USER} 192.168.0.0/16 md5
EOF
    echo "[pg-entrypoint] pg_hba.conf 追加 replication 授权"
  fi
  # 重载配置让授权立即生效
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -c "SELECT pg_reload_conf();" >/dev/null 2>&1 || true

  echo "[pg-entrypoint] replication user + slot ready"
}

replica_init() {
  echo "[pg-entrypoint] PG_ROLE=replica, 启动 standby"
  # 等 primary 可达
  for i in $(seq 1 60); do
    if pg_isready -h "$PG_PRIMARY_HOST" -p "$PG_PRIMARY_PORT" -U "$POSTGRES_USER" >/dev/null 2>&1; then
      echo "[pg-entrypoint] primary $PG_PRIMARY_HOST:$PG_PRIMARY_PORT ready"
      break
    fi
    echo "[pg-entrypoint] waiting for primary... ($i/60)"
    sleep 2
  done

  # 如果 data 目录是空的，做一次基线备份
  if [ ! -s "$PGDATA/PG_VERSION" ]; then
    echo "[pg-entrypoint] empty PGDATA, running pg_basebackup from primary"
    PGPASSWORD="$PG_REPL_PASSWORD" pg_basebackup \
      -h "$PG_PRIMARY_HOST" -p "$PG_PRIMARY_PORT" \
      -U "$PG_REPL_USER" -D "$PGDATA" \
      -Xs -P -R -S "$PG_REPL_SLOT"
    echo "[pg-entrypoint] pg_basebackup done"
  else
    echo "[pg-entrypoint] PGDATA 已存在, 跳过 pg_basebackup"
    # 确保 standby.signal 存在
    touch "$PGDATA/standby.signal"
  fi

  echo "[pg-entrypoint] starting replica PG"
  exec "$PG_ENTRYPOINT" "$@"
}

case "$PG_ROLE" in
  primary)
    # 后台跑 post-start 任务
    (sleep 5 && primary_post_start) &
    primary_init "$@"
    ;;
  replica)
    replica_init "$@"
    ;;
  *)
    echo "[pg-entrypoint] unknown PG_ROLE=$PG_ROLE, fallback to primary"
    primary_init "$@"
    ;;
esac
