# transcoder-cpu

CPU 软编转码服务。从 MinIO 读取源视频 → ffmpeg 转码为多档 HLS → 产物写回 MinIO。

## 架构

```
业务服务 (work) → POST /api/v1/transcode → transcoder-cpu → ffmpeg → HLS 产物
                        ↓                                        ↓
                   MinIO (读源/写产物)                      MinIO (写 playlist.m3u8)
```

**被调用方**：不关心业务逻辑，只接收 MinIO 对象引用，返回播放地址。

## 部署步骤

### 1. 安装 ffmpeg

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y ffmpeg

# CentOS/RHEL
sudo yum install -y epel-release && sudo yum install -y ffmpeg

# 验证
ffmpeg -version
ffprobe -version
```

其他平台见 https://ffmpeg.org/download.html

### 2. 配置环境变量

在 `dev/_env/common.env` 或 `.env` 中添加：

```env
HTTP_ADDR=:8092
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=mybilibili
```

### 3. 启动服务

**注意：** 转码服务**不做容器**，裸跑宿主机（需要系统 ffmpeg + 对应显卡驱动）。

```bash
cd external/transcoder-cpu
go run ./cmd/transcoder-cpu

# 验证
curl http://localhost:8092/health
# → {"status":"ok"}

# 查看能力
curl http://localhost:8092/api/v1/capabilities
# → {"encoder":"cpu","capabilities":["h264","h265"]}
```

### 4. 测试转码

```bash
curl -X POST http://localhost:8092/api/v1/transcode \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "mybilibili",
    "source_key": "manuscripts/10/videos/25/source/video.mp4",
    "manuscript_id": 10,
    "video_id": 25,
    "qualities": ["1080p", "720p", "480p"],
    "extract_audio": true
  }'
```

## API

### POST /api/v1/transcode

```json
// 请求
{
  "bucket": "mybilibili",
  "source_key": "manuscripts/10/videos/25/source/video.mp4",
  "manuscript_id": 10,
  "video_id": 25,
  "qualities": ["1080p", "720p", "480p"],
  "extract_audio": true
}

// 响应
{
  "play_urls": {
    "1080p": "/uploads/manuscripts/10/videos/25/transcoded/1080p/playlist.m3u8",
    "720p": "/uploads/manuscripts/10/videos/25/transcoded/720p/playlist.m3u8",
    "480p": "/uploads/manuscripts/10/videos/25/transcoded/480p/playlist.m3u8"
  },
  "audio_key": "manuscripts/10/videos/25/audio/audio.mp3",
  "is_vertical": 0
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| HTTP_ADDR | :8092 | 监听端口 |
| MINIO_ENDPOINT | localhost:9000 | MinIO 地址 |
| MINIO_ACCESS_KEY | minioadmin | MinIO Access Key |
| MINIO_SECRET_KEY | minioadmin | MinIO Secret Key |
| MINIO_BUCKET | mybilibili | MinIO Bucket |

## 三版转码服务对比

| 服务 | 编码器 | 适用场景 |
|------|--------|----------|
| transcoder-cpu | libx265 (CPU) | 通用，无 GPU 时 |
| transcoder-nvenc | NVIDIA NVENC | 有 NVIDIA GPU 时 |
| transcoder-vaapi | AMD/Intel VAAPI | 有 AMD/Intel GPU 时 |

三版 API 完全相同，可互换。优先用 GPU 版，失败自动 fallback 到 CPU。
