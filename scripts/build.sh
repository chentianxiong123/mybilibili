#!/usr/bin/env bash
# 纯编译入口: 编译全部 Go 后端服务到 $BIN_DIR (默认 ./bin, Go 生态约定).
#
# 不依赖 docker, 不起基础设施, 单纯编译.
# 被 scripts/backend.sh 调用, 也可独立使用:
#
#   scripts/build.sh                      # 编译全部到 ./bin (相对 repo root)
#   BIN_DIR=/usr/local/bin scripts/build.sh
#   BIN_DIR=$HOME/bin scripts/build.sh
#   scripts/build.sh transcoder           # 只编译 transcoder (3 个变体)
#   scripts/build.sh clean                # 清掉 $BIN_DIR 下所有 mybilibili-* 二进制
#   scripts/build.sh clean-source         # 兜底清理源码目录里的 in-source build 残留
#
# 设计:
#   - Makefile 的 BIN 变量支持命令行覆盖 (BIN=/path make build), 这里透传
#   - 默认输出 ./bin 是因为: 1) Go 约定 2) 跨平台 (不需要 /tmp) 3) 一行 .gitignore
#   - 真正的裸跑 + 起基础设施请用 scripts/backend.sh (会调本脚本)

set -u
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GO_DIR="$ROOT"
BIN_DIR="${BIN_DIR:-./bin}"

cmd="${1:-build}"
shift || true

case "$cmd" in
    build)
        echo "== 编译全部服务 → $BIN_DIR (相对 $ROOT) =="
        mkdir -p "$BIN_DIR"
        BIN="$BIN_DIR" make -s -C "$GO_DIR" build
        echo "✓ 产物:"
        ls -lh "$BIN_DIR"/mybilibili-* 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
        ;;
    transcoder)
        echo "== 编译 transcoder 全部变体 → $BIN_DIR =="
        mkdir -p "$BIN_DIR"
        BIN="$BIN_DIR" make -s -C "$GO_DIR" build-transcoder
        echo "✓ 产物:"
        ls -lh "$BIN_DIR"/mybilibili-transcoder* 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
        ;;
    clean)
        echo "== 清理 $BIN_DIR/mybilibili-* =="
        rm -f "$BIN_DIR"/mybilibili-*
        [ -d "$BIN_DIR" ] && [ -z "$(ls -A "$BIN_DIR" 2>/dev/null)" ] && rmdir "$BIN_DIR" 2>/dev/null
        echo "✓ done"
        ;;
    clean-source)
        make -s -C "$GO_DIR" clean-source
        ;;
    *)
        sed -n '2,20p' "$0"
        exit 1
        ;;
esac
