#!/usr/bin/env bash
# 拉取并编译 llama.cpp (Vulkan 后端, AMD RX580/RX590)
# 用法: external/embedding-llamacpp-vulkan/setup.sh
set -euo pipefail

REPO_URL="https://github.com/chentianxiong123/llama.cpp-lora-embed.git"
BRANCH="qlora"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="$SCRIPT_DIR/llama.cpp-lora-embed"
BUILD_DIR="$SRC_DIR/build"
BIN_DIR="$SCRIPT_DIR/bin"

# 1. 拉取源码
if [ -d "$SRC_DIR/.git" ]; then
  echo "✓ 源码已存在, git pull…"
  git -C "$SRC_DIR" checkout "$BRANCH"
  git -C "$SRC_DIR" pull --ff-only
else
  echo "→ 克隆 $BRANCH …"
  git clone --branch "$BRANCH" --depth 1 "$REPO_URL" "$SRC_DIR"
fi

# 2. 编译 (Vulkan 后端)
echo "→ 编译 llama.cpp (Vulkan)…"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"
cmake .. -DGGML_VULKAN=ON -DCMAKE_BUILD_TYPE=Release
cmake --build . -j$(nproc)

# 3. 复制二进制到 bin/
mkdir -p "$BIN_DIR"
cp -f "$BUILD_DIR/bin/llama-server" "$BIN_DIR/" 2>/dev/null || \
cp -f "$BUILD_DIR/bin/llama-server-impl" "$BIN_DIR/" 2>/dev/null || true
cp -f "$BUILD_DIR/bin/llama-quantize" "$BIN_DIR/" 2>/dev/null || true

echo ""
echo "✓ 编译完成"
echo "  二进制: $BIN_DIR/llama-server"
echo "  用法:   $BIN_DIR/llama-server --model <model.gguf> --embedding --port 8081"
echo ""
echo "  模型文件放到 $SCRIPT_DIR/models/ 目录下"
echo "  推荐: qwen3-embedding-0.6b-q8_0.gguf (~610MB)"
