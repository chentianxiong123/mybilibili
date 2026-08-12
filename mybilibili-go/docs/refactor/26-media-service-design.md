# 媒体服务设计：MQ 驱动的转码流水线

## 1. 定位

Media 是 Core 胖单体之外的**独立服务**，处理所有 CPU/GPU 密集型任务。它不处理用户请求，而是消费 MQ 消息执行异步任务。

| 服务 | 定位 | 资源需求 | 部署位置 |
|------|------|---------|---------|
| Core | 用户请求处理 | 内存/连接 | 盒子/工作站 |
| **Media** | **异步任务处理** | **CPU/GPU** | **仅工作站** |
| Studio | 剪辑导出（后续） | CPU/GPU | 仅工作站 |

---

## 2. 流水线架构

```
Core（审核通过）
  │
  │ 发 MQ 消息（NATS）
  ▼
NATS（工作站）
  │
  │ Media 消费消息
  ▼
Media 流水线（按顺序执行）
  ├── 1. 转码：源视频 → FFmpeg → HLS(1080p/720p/480p) → 上传 MinIO
  ├── 2. 提音频：源视频 → FFmpeg → AAC → 上传 MinIO
  ├── 3. 字幕：音频 → Whisper → SRT/VTT → 上传 MinIO
  ├── 4. AI 摘要：内容 → AI API → 存 PG
  └── 5. 完成：检查稿件下所有视频 → 全部完成 → 自动上架
        │
        ▼
  Core（收到回调 → 更新 video 状态 → 自动上架 manuscript）
```

---

## 3. 各步骤详解

### 3.1 转码（Transcode）

```
输入：源视频文件（mp4/mov/avi...）
输出：HLS 切片（m3u8 + ts），三个清晰度
  - 1080p: scale=1920:-2, bitrate=4000k
  - 720p:  scale=1280:-2, bitrate=2000k
  - 480p:  scale=854:-2,  bitrate=1000k
工具：FFmpeg（exec 调用）
进度：解析 FFmpeg 的 progress 输出 → SSE 推给前端
存储：上传到 MinIO：/manuscripts/{id}/videos/{id}/transcoded/{res}/playlist.m3u8
```

### 3.2 音频提取（Audio Extract）

```
输入：源视频文件
输出：AAC 音频文件
工具：FFmpeg（ffmpeg -i input -vn -c:a aac output.aac）
用途：后续字幕生成和 AI 摘要的输入
存储：上传到 MinIO：/manuscripts/{id}/videos/{id}/audio/audio.aac
```

### 3.3 字幕生成（Subtitle Generate）

```
输入：音频文件
输出：SRT/VTT 字幕文件
工具：Whisper（本地模型或 API）
存储：上传到 MinIO：/manuscripts/{id}/videos/{id}/subtitle/subtitle.vtt
```

### 3.4 AI 摘要（AI Summary）

```
输入：视频标题 + 描述 + 字幕文本
输出：AI 生成的摘要
工具：AI API（DeepSeek/OpenAI 兼容接口）
存储：写入 PG videos 表的 summary 字段
```

---

## 4. 状态机

### 4.1 Video 处理状态

| 值 | 状态 | 说明 |
|----|------|------|
| 0 | PENDING | 等待处理 |
| 1 | TRANSCODING | 转码中 |
| 10 | TRANSCODE_FAILED | 转码失败 |
| 11 | TRANSCODE_SUCCESS | 转码成功 |
| 2 | AUDIO_EXTRACTING | 提取音频中 |
| 20 | AUDIO_FAILED | 音频提取失败 |
| 21 | AUDIO_SUCCESS | 音频提取成功 |
| 3 | SUBTITLE_GENERATING | 生成字幕中 |
| 30 | SUBTITLE_FAILED | 字幕生成失败 |
| 31 | SUBTITLE_SUCCESS | 字幕生成成功 |
| 4 | AI_SUMMARIZING | AI 摘要中 |
| 40 | AI_FAILED | AI 摘要失败 |
| 41 | AI_SUCCESS | AI 摘要成功 |
| 5 | COMPLETED | 全部完成 |

### 4.2 流水线执行逻辑

```
每条消息包含 processType + processMode

processType:
  - transcode: 仅转码
  - extract_audio: 仅提取音频
  - generate_subtitle: 仅生成字幕
  - ai_summary: 仅 AI 摘要
  - auto_process: 全流程自动执行

processMode:
  - auto: 自动进入下一步
  - manual_single: 只执行当前步骤
```

---

## 5. 进度回传

每步处理通过 HTTP 回调 Core 的 SSE 端点，推给前端：

```
Media → POST /api/v1/video/progress
  { video_id, process_status, process_progress, process_stage }

Core → SSE → 前端
  前端收到进度更新，显示在创作中心
```

---

## 6. 目录结构

```
cmd/media/main.go
internal/media/
├── pipeline/          ← 流水线编排
│   ├── orchestrator.go   ← 状态机调度
│   └── processor.go      ← 任务分发
├── transcoder/        ← 转码（FFmpeg）
│   └── ffmpeg.go
├── audio/             ← 音频提取
│   └── extractor.go
├── subtitle/          ← 字幕生成
│   └── generator.go
├── ai/                ← AI 摘要
│   └── summarizer.go
└── storage/           ← MinIO 文件管理
    └── manager.go
```

---

## 7. 与 Core 的交互

```
Media 不需要 gRPC，只通过两种方式与 Core 通信：
  1. MQ（NATS）：接收任务
  2. HTTP：回传进度、更新状态

Core 暴露给 Media 的接口：
  POST /api/v1/internal/video/status   ← 更新 video 处理状态
  POST /api/v1/internal/video/progress ← 推 SSE 进度
  POST /api/v1/internal/manuscript/auto-publish ← 检查并自动上架
```

---

## 8. 部署

```
Media 只跑在 E5 工作站，不在盒子上。
需要：
  - FFmpeg（系统安装）
  - Whisper（可选，本地或 API）
  - MinIO 访问权限
  - NATS 访问权限
  - PG 访问权限（更新状态）

启动：
  FFMPEG_PATH=/usr/bin/ffmpeg \
  NATS_URL=nats://localhost:4222 \
  MINIO_ENDPOINT=localhost:9000 \
  PG_DSN=postgres://postgres@localhost:5432/mybilibili \
  ./media
```