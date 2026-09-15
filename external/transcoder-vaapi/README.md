# transcoder-vaapi

AMD/Intel GPU 硬编转码服务。从 MinIO 读取源视频 → ffmpeg VAAPI 转码为多档 HLS → 产物写回 MinIO。

## 架构

```
业务服务 (work) → POST /api/v1/transcode → transcoder-vaapi → ffmpeg (VAAPI) → HLS 产物
                        ↓                                               ↓
                   MinIO (读源/写产物)                             MinIO (写 playlist.m3u8)
```

**被调用方**：不关心业务逻辑，只接收 MinIO 对象引用，返回播放地址。

## 部署步骤

### 1. 安装 GPU 驱动 + ffmpeg

**AMD GPU:**
```bash
# 安装 Mesa VAAPI 驱动
sudo apt update && sudo apt install -y mesa-va-drivers vainfo

# 验证
vainfo
```

**Intel GPU:**
```bash
# 安装 Intel VAAPI 驱动
sudo apt update && sudo apt install -y intel-media-va-driver vainfo

# 验证
vainfo
```

**安装 ffmpeg:**
```bash
sudo apt install -y ffmpeg

# 验证 vaapi 支持
ffmpeg -encoders 2>/dev/null | grep vaapi
# → V..... h264_vaapi     H.264 VAAPI encoder
# V..... hevc_vaapi      H.265 VAAPI encoder
```

### 2. 配置环境变量

在 `dev/_env/common.env` 或 `.env` 中添加：

```env
HTTP_ADDR=:8094
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=mybilibili
VAAPI_DEVICE=/dev/dri/renderD128
```

### 3. 启动服务

**注意：** 转码服务**不做容器**，裸跑宿主机（需要系统 ffmpeg + GPU 驱动）。

```bash
cd external/transcoder-vaapi
go run -tags vaapi ./cmd/transcoder-vaapi

# 验证
curl http://localhost:8094/health
# → {"status":"ok"}

# 查看能力
curl http://localhost:8094/api/v1/capabilities
# → {"encoder":"vaapi","capabilities":["h264_vaapi","hevc_vaapi"]}
```

### 4. 测试转码

```bash
curl -X POST http://localhost:8094/api/v1/transcode \
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
| HTTP_ADDR | :8094 | 监听端口 |
| MINIO_ENDPOINT | localhost:9000 | MinIO 地址 |
| MINIO_ACCESS_KEY | minioadmin | MinIO Access Key |
| MINIO_SECRET_KEY | minioadmin | MinIO Secret Key |
| MINIO_BUCKET | mybilibili | MinIO Bucket |
| VAAPI_DEVICE | /dev/dri/renderD128 | VAAPI 设备路径 |

## 三版转码服务对比

| 服务 | 编码器 | 适用场景 |
|------|--------|----------|
| transcoder-cpu | libx265 (CPU) | 通用，无 GPU 时 |
| transcoder-nvenc | NVIDIA NVENC | 有 NVIDIA GPU 时 |
| transcoder-vaapi | AMD/Intel VAAPI | 有 AMD/Intel GPU 时 |

三版 API 完全相同，可互换。优先用 GPU 版，失败自动 fallback 到 CPU。
