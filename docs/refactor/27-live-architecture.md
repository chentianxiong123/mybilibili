# 直播架构：SRS + Live 模块

## 1. 职责划分

直播业务分两层：**媒体层**（SRS/OBS）和**业务层**（Live 模块）。

```
媒体层（SRS/OBS 干，不写代码）：
  OBS（主播电脑）──RTMP──▶ SRS（工作站）
                              │
                              ├── HLS 切片 → 观众 HLS.js 播放
                              ├── FLV 封装 → 低延迟播放
                              └── WebRTC SFU → 连麦媒体转发

业务层（Live 模块，写代码）：
  ├── 房间 CRUD + streamKey 生成
  ├── SRS 回调（on_publish → 标为直播中）
  ├── 连麦信令（WS 转发 offer/answer/ice）
  └── 语聊房麦位状态机
```

---

## 2. Live 模块职责

### 2.1 房间管理（纯 CRUD）

| 功能 | 说明 |
|------|------|
| 创建直播间 | 生成唯一 streamKey，OBS 用它推流 |
| 获取直播间 | 按用户/ID/streamKey 查询 |
| 更新直播间 | 封面/名称/分类 |
| 列表 | 直播中的房间列表 |
| 定时开播 | 预约开播时间 |
| 在线人数 | 更新 viewerCount |

### 2.2 SRS 回调（HTTP POST）

```
OBS 开始推流 → SRS → POST /live/room/srs/hook
  { action: on_publish, stream: "streamKey123" }

Live 收到 → 按 streamKey 匹配房间 → 更新 status = live
必须返回 {code:0}，否则 SRS 拒绝推流
```

### 2.3 连麦信令（WS）

```
观众WebSocket ──▶ Core ──▶ 主播WebSocket
  msg: { type: "offer", sdp: "..." }

Core 只做转发：收到 A 的消息，原样发给 B
```

### 2.4 语聊房麦位状态机

```
观众 ──申请上麦──▶ 排队 ──房主同意──▶ 麦上 ──下麦──▶ 观众
麦上 ──禁麦──▶ 麦上(静音) ──解禁──▶ 麦上
麦上 ──踢人──▶ 房间外
```

---

## 3. WS vs WebRTC 分工

| 维度 | WebSocket（信令） | WebRTC（媒体） |
|------|------------------|---------------|
| 职责 | 交换握手信息 | 传输音视频数据 |
| 握手 | 1 次，~10ms | 5-10 次，~2-3 秒 |
| 每连接内存 | ~20KB | ~100KB |
| 用途 | offer/answer/ice 转发 | 连麦/会议音视频 |
| 开销 | 低 | 高 |

```
连麦流程：
WS 信令：A 发 offer → 服务端转发 → B 收 → B 回 answer → A 收
         （完成后 WS 仍在，用于后续控制：静音/踢人/开关视频）
WebRTC：媒体数据直接 P2P 或经 SRS SFU 转发
```

**WS 和 WebRTC 是配合关系，不是替代关系。** WS 做轻量信令，WebRTC 做重媒体。

---

## 4. 合并进 Core 的理由

Live 模块是**轻业务**（CRUD + HTTP 回调 + WS 转发），资源消耗跟 manuscript/comment 一个级别：

| 维度 | 合并进 Core | 独立服务 |
|------|-----------|---------|
| 部署 | 1 个二进制 | 2 个二进制 |
| 内存 | +2MB | +15MB |
| WS 连接 | 200 连接 ~4MB，Core 扛得住 | 独立扩缩 |
| 代码复杂度 | 模块边界清晰 | 多一个 main.go |

**Live 合并进 Core 胖单体，模块放 `internal/live/`，边界按可独立编译标准隔离。**

---

## 5. 目录结构

```
internal/live/
├── room/        ← 房间 CRUD + streamKey
├── srs/         ← SRS 回调处理
├── linkmic/     ← 连麦信令（WS）
└── seat/        ← 语聊房麦位状态机（WS）

cmd/core/main.go
  ├── gRPC :9090  → user/manuscript/comment/interaction/live
  ├── HTTP :8080  → danmaku/SSE/SRS回调
  └── WS  :8080   → 连麦信令/语聊房麦位
```

---

## 6. 真正的独立服务：Media

跟 Live 不同，**Media 是真正该独立的服务**：

| 对比 | Live | Media |
|------|------|-------|
| 资源 | 内存/连接（轻） | **CPU/GPU（重）** |
| 任务 | 请求响应（同步） | **异步处理** |
| 部署 | 盒子/工作站 | **仅工作站** |
| 依赖 | 无 | FFmpeg/Whisper/MinIO |

**Live 和 Core 合一个二进制，Media 独立，Studio 后续独立。**

---

## 7. 组件清单

| 组件 | 大小 | 内存 | 职责 | 部署 |
|------|------|------|------|------|
| OBS | 200MB | - | 推流端（主播装） | 主播电脑 |
| OBS 插件 | 5MB | - | 扫码登录/推流管理 | 主播电脑 |
| SRS | 15MB | 10-30MB | 收流/转码/分发/WebRTC SFU | 工作站 |
| Core（含 Live） | 20MB | 15-30MB | 房间/回调/信令/CRUD | 盒子/工作站 |
| HLS.js | 100KB | - | 观众播放器 | 浏览器 |