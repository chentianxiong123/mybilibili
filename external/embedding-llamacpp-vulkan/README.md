# embedding-llamacpp-vulkan

本地 embedding 向量生成服务。基于 llama.cpp (Vulkan 后端) 运行 GGUF embedding 模型，提供 OpenAI 兼容 API。

## 架构

```
业务服务 → POST /v1/embeddings → embedding-vulkan → llama-server → 向量结果
```

**被调用方**：不关心业务逻辑，只接收文本，返回 embedding 向量。

## 部署步骤

### 1. 下载 llama-server 二进制

从 llama.cpp Releases 下载预编译包（含 Vulkan 支持）：

```bash
cd /tmp
wget https://github.com/ggerganov/llama.cpp/releases/latest/download/llama-ubuntu-x64-vulkan.zip
unzip llama-ubuntu-x64-vulkan.zip
sudo mv bin/llama-server /usr/local/bin/
sudo chmod +x /usr/local/bin/llama-server

# 验证
llama-server --version
```

其他平台见 https://github.com/ggerganov/llama.cpp/releases

> Vulkan 支持需要显卡驱动已安装 Vulkan SDK。
> AMD: `sudo apt install mesa-vulkan-drivers`
> NVIDIA: 驱动自带 Vulkan
> Intel: `sudo apt install mesa-vulkan-drivers`

### 2. 下载 Embedding 模型

```bash
mkdir -p /models/embedding

# qwen3-embedding-0.6b（推荐，~610MB）
wget -P /models/embedding/ \
  https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF/resolve/main/qwen3-embedding-0.6b-q8_0.gguf

# 验证
ls -lh /models/embedding/
```

其他可用模型：
- `bge-small-zh-v1.5`（~130MB，中文轻量）
- `bge-large-zh-v1.5`（~1.3GB，中文高精度）
- `nomic-embed-text`（~260MB，英文通用）

### 3. 配置环境变量

在 `dev/_env/common.env` 或 `.env` 中添加：

```env
HTTP_ADDR=:8081
LLAMA_SERVER_URL=
LLAMA_SERVER_BIN=/usr/local/bin/llama-server
LLAMA_MODEL=/models/embedding/qwen3-embedding-0.6b-q8_0.gguf
```

- `LLAMA_SERVER_URL`：如果 llama-server 已在别处运行，填地址（如 `http://127.0.0.1:8081`），留空则自动启动
- `LLAMA_SERVER_BIN`：llama-server 二进制路径，留空则只做代理模式

### 4. 启动服务

```bash
# Docker Compose
docker compose -f dev/docker-compose.yml up -d embedding-vulkan

# 验证
curl http://localhost:8081/health
# → {"status":"ok"}
```

### 5. 测试 Embedding

```bash
curl -X POST http://localhost:8081/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen3-embedding-0.6b",
    "input": "hello world"
  }'
```

## API（OpenAI 兼容）

### POST /v1/embeddings

```json
// 请求
{
  "model": "qwen3-embedding-0.6b",
  "input": "要向量化的文本"
}

// 响应
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0123, -0.0456, ...],
      "index": 0
    }
  ],
  "model": "qwen3-embedding-0.6b",
  "usage": {
    "prompt_tokens": 10,
    "total_tokens": 10
  }
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| HTTP_ADDR | :8081 | 监听端口 |
| LLAMA_SERVER_URL | (空) | 已运行的 llama-server 地址，留空则自动启动 |
| LLAMA_SERVER_BIN | (空) | llama-server 二进制路径 |
| LLAMA_MODEL | (空) | GGUF 模型文件完整路径 |

## 模型选择

| 模型 | 大小 | 维度 | 用途 |
|------|------|------|------|
| qwen3-embedding-0.6b | ~610MB | 1024 | 中文+英文，推荐 |
| bge-small-zh-v1.5 | ~130MB | 512 | 中文轻量 |
| bge-large-zh-v1.5 | ~1.3GB | 1024 | 中文高精度 |
| nomic-embed-text | ~260MB | 768 | 英文通用 |
