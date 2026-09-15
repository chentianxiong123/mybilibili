# whisper-local

本地 whisper.cpp 语音转写服务。从 MinIO 读音频 → whisper-cli 转写 → 返回文本/时间戳。

## 架构

```
业务服务 (AI) → POST /api/v1/transcribe → whisper-local → whisper-cli → 结果
                      ↓
                 MinIO (读音频)
```

**被调用方**：不关心业务逻辑，只接收 MinIO 对象引用，返回转写结果。

## 部署步骤

### 1. 下载 whisper-cli 二进制

```bash
# Linux x86_64
cd /tmp
wget https://github.com/ggerganov/whisper.cpp/releases/latest/download/whisper-linux-x64.tar.gz
tar xzf whisper-linux-x64.tar.gz
sudo mv whisper-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/whisper-cli

# 验证
whisper-cli --version
```

其他平台见 https://github.com/ggerganov/whisper.cpp/releases

### 2. 下载模型

```bash
mkdir -p /models/whisper

# base 模型（~150MB，推荐起步用）
wget -P /models/whisper/ \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin

# small 模型（~500MB，精度更高）
wget -P /models/whisper/ \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin

# 验证
ls -lh /models/whisper/
```

### 3. 配置环境变量

在 `dev/_env/common.env` 或 `.env` 中添加：

```env
WHISPER_CLI_PATH=/usr/local/bin/whisper-cli
WHISPER_MODEL_DIR=/models/whisper
WHISPER_DEFAULT_MODEL=base
WHISPER_LANGUAGE=zh
WHISPER_THREADS=4
```

### 4. 启动服务

```bash
# Docker Compose
docker compose -f dev/docker-compose.yml up -d whisper-local

# 验证
curl http://localhost:8095/health
# → {"status":"ok"}

# 查看能力
curl http://localhost:8095/api/v1/capabilities
# → {"engine":"whisper-cpp","model_dir":"/models/whisper"}
```

### 5. 测试转写

```bash
# 准备一个测试音频文件放到 MinIO（manuscripts/10/videos/25/audio/audio.mp3）

curl -X POST http://localhost:8095/api/v1/transcribe \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "mybilibili",
    "source_key": "manuscripts/10/videos/25/audio/audio.mp3",
    "language": "zh",
    "model": "base"
  }'
```

## API

### POST /api/v1/transcribe

纯文本转写（快速，无精确时间戳）。

```json
// 请求
{
  "bucket": "mybilibili",
  "source_key": "manuscripts/10/videos/25/audio/audio.mp3",
  "language": "zh",
  "model": "base"
}

// 响应
{
  "text": "完整转写文本",
  "segments": [
    {"index": 1, "start": 0, "end": 0, "text": "第一行"},
    {"index": 2, "start": 0, "end": 0, "text": "第二行"}
  ]
}
```

### POST /api/v1/transcribe/srt

带时间戳转写（SRT 格式解析，精确到毫秒）。

```json
// 响应
{
  "text": "完整转写文本",
  "segments": [
    {"index": 1, "start": 0.0, "end": 2.5, "text": "第一句"},
    {"index": 2, "start": 2.5, "end": 5.1, "text": "第二句"}
  ]
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| HTTP_ADDR | :8095 | 监听端口 |
| WHISPER_CLI_PATH | whisper-cli | whisper-cli 二进制路径 |
| WHISPER_MODEL_DIR | /models/whisper | 模型目录 |
| WHISPER_DEFAULT_MODEL | base | 默认模型名（对应 ggml-base.bin） |
| WHISPER_LANGUAGE | zh | 默认语言 |
| WHISPER_THREADS | 4 | 推理线程数 |
| MINIO_ENDPOINT | localhost:9000 | MinIO 地址 |
| MINIO_ACCESS_KEY | minioadmin | MinIO Access Key |
| MINIO_SECRET_KEY | minioadmin | MinIO Secret Key |
| MINIO_BUCKET | mybilibili | MinIO Bucket |

## 模型选择

| 模型 | 大小 | 速度 | 精度 |
|------|------|------|------|
| tiny | ~75MB | 最快 | 低 |
| base | ~150MB | 快 | 中 |
| small | ~500MB | 中 | 较高 |
| medium | ~1.5GB | 慢 | 高 |
| large | ~3GB | 最慢 | 最高 |

## Docker

docker-compose.yml 中已配置挂载宿主机的 whisper-cli 和模型：

```yaml
volumes:
  - /usr/local/bin/whisper-cli:/usr/local/bin/whisper-cli:ro
  - /home/a1/models/whisper:/models/whisper:ro
```
