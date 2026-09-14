#!/bin/sh
# 标准 Docker entrypoint — 与官方 postgres/redis/nginx 镜像相同的初始化模式:
#   1. 以 root 身份处理数据目录权限（解决 external volume 任意所有权问题）
#   2. 用 su-exec drop 到非 root 用户 (app) 跑业务二进制
#
# 用法:
#   在 docker-compose 的 environment 里设置 DATA_DIR=/data/<service>
#   容器启动时会自动 mkdir -p 并 chown -R app:app，兼容任何来源的 volume
#   未设置 DATA_DIR 则跳过初始化，二进制直接以 app 身份启动
#
# 为什么不用 gosu: alpine 原生仓库自带 su-exec，更轻量

set -eu

if [ -n "${DATA_DIR:-}" ]; then
    mkdir -p "$DATA_DIR"
    # 兼容 root-owned / 混合所有权 / 已存在的子文件，全部递归 chown
    chown -R app:app "$DATA_DIR"
fi

exec su-exec app "$@"
