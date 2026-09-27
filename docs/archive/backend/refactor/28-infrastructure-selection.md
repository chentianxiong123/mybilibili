# 基础设施选型：组件清单与资源消耗

## 1. 最终组件清单

| 组件 | 版本 | 大小 | 内存 | 语言 | 职责 | 部署位置 |
|------|------|------|------|------|------|---------|
| **Traefik** | v3 | 30MB | 30MB | Go | 网关/路由/JWT/限流/证书 | **k3s 内置** |
| **PostgreSQL** | 16 | 200MB | 200MB | C | 一库三用（关系/JSONB/全文检索） | 工作站 |
| **Redis** | 7 | 5MB | 50MB | C | 写缓冲/分布式锁/计数器 | 工作站 |
| **NATS** | 2 | 15MB | 10MB | Go | 消息队列（弹幕广播/异步写/任务队列） | 工作站 |
| **MinIO** | latest | 100MB | 100MB | Go | 对象存储（视频/字幕/头像/封面） | 工作站 |
| **SRS** | 6 | 15MB | 10-30MB | C++ | 直播收流/HLS分发/WebRTC SFU | 工作站 |
| **Core 胖单体** | - | 20MB | 15-30MB | Go | 全部业务逻辑（含 Live） | 盒子/工作站 |
| **Media 服务** | - | 15MB | 20-50MB | Go | 转码/音频/字幕/AI 摘要 | 仅工作站 |
| **etcd** | - | 50MB | 50MB | Go | ❌ **不需要**（k3s 内建 DNS） | - |

**工作站合计：~420MB（不含 Media 运行时）**
**盒子合计：~60MB（Core 胖单体 20MB + 系统开销）**

---

## 2. 为什么选这些组件

### 2.1 NATS（MQ）

| 对比 | NATS | Redis Stream | RocketMQ | RabbitMQ | Kafka |
|------|------|-------------|---------|----------|-------|
| 大小 | **15MB** | 5MB（已有Redis） | 500MB | 150MB | 300MB |
| 内存 | **10MB** | 5MB（已有Redis） | 600MB | 200MB | 500MB |
| 语言 | **Go** | C | Java | Erlang | Java/Scala |
| CNCF | ✅ 孵化 | ❌ | ❌ | ❌ | ❌ |
| JetStream | ✅ 持久化 | ❌ 无 | ✅ | ✅ | ✅ |

**结论：NATS 是最轻的专业 MQ，跟 k3s 同一生态。Redis Stream 能用但不够专业。**

### 2.2 SRS（媒体服务器）

| 对比 | SRS | NGINX-RTMP | MediaSoup | Janus | LiveGo |
|------|-----|-----------|-----------|-------|--------|
| 大小 | **15MB** | 5MB | 50MB | 100MB | 15MB |
| 内存 | **10MB** | 5MB | 80MB | 150MB | 15MB |
| RTMP 收流 | ✅ | ✅ | ❌ | ❌ | ✅ |
| HLS 分发 | ✅ | ✅ | ❌ | ❌ | ✅ |
| FLV 分发 | ✅ | ❌ | ❌ | ❌ | ✅ |
| WebRTC SFU | ✅ | ❌ | ✅ | ✅ | ✅ |
| HTTP 回调 | ✅ | ✅ | ❌ | ✅ | ✅ |

**结论：SRS 一个二进制覆盖 RTMP+HLS+FLV+WebRTC，15MB 全功能，无需组合多个组件。**

### 2.3 PostgreSQL（一库三用）

| 对比 | PG | MySQL + MongoDB + ES |
|------|-----|---------------------|
| 组件数 | **1 个** | 3 个 |
| 总大小 | **200MB** | 500MB+ |
| 关系表 | ✅ | ✅ MySQL |
| JSONB 文档 | ✅ | ✅ MongoDB |
| 全文检索 | ✅ tsvector | ✅ ES |
| 运维 | **1 套** | 3 套 |

**结论：PG 一库三用，1000 用户规模完全够用，省掉 MongoDB 和 ES。**

### 2.4 WebSocket vs WebRTC

| 对比 | WebSocket | WebRTC |
|------|-----------|--------|
| 握手 | 1 次，~10ms | 5-10 次，**~2-3 秒** |
| 每连接内存 | **~20KB** | ~100KB |
| 加密 | 可选 | 强制 DTLS/SRTP |
| 用途 | 文本/信令 | **音视频媒体** |
| 开销 | **低** | 高 |

**结论：弹幕/通知/信令用 WS，连麦/会议音视频用 WebRTC，各司其职。**

---

## 3. 依赖关系

```
Core 胖单体
  ├── PostgreSQL（必须）
  ├── Redis（必须）
  └── NATS（可选，无 MQ 时用内存队列降级）

Media 服务
  ├── NATS（必须）
  ├── MinIO（必须）
  ├── PostgreSQL（必须）
  └── FFmpeg（必须，系统安装）

SRS
  └── 独立运行，不需任何其他组件

盒子独立运行（无工作站）：
  Core 胖单体
    ├── SQLite（替代 PG）
    ├── 内存缓存（替代 Redis）
    └── 内存队列（替代 NATS）
```

---

## 4. 启动顺序

```
1. PostgreSQL（工作站）
2. Redis（工作站）
3. NATS（工作站）
4. MinIO（工作站）
5. SRS（工作站，如需直播）
6. Core 胖单体（盒子/工作站）
7. Media 服务（工作站，如有转码任务）
```

---

## 5. k3s 下的简化

在 k3s 集群中，以下组件不需要额外部署：

| 组件 | k3s 替代方案 |
|------|------------|
| Traefik | k3s 默认安装，无需单独部署 |
| etcd | k3s 用 SQLite3 替代 etcd，服务发现用 k8s DNS |
| 服务发现 | k8s Service + CoreDNS，无需 etcd/Consul |

**k3s 省掉了 etcd 和 Traefik 的独立部署，进一步降低资源消耗。**