# mybilibili · 仿哔哩哔哩全栈项目

一套完整仿 B 站的视频社区平台，覆盖用户端 Web（SSR）、管理后台、移动 H5、Flutter / NativeScript / Android WebView 多端，后端为 Go 微服务架构，容器化部署在三机 k3s 集群（部署机 + 开发机 + 边缘 Worker）。

## 目录结构

```
mybilibili/
├── mybilibili-go/            # Go 后端（10 个微服务）
├── mybilibili-front/         # Web 前端（pnpm monorepo：web SSR + admin 后台 + ui 组件库）
├── mybilibili-wap/           # 移动端 H5（Vue3 + Vite）
├── mybilibili-app-flutter/   # Flutter 跨端 App
├── mybilibili-ns/            # NativeScript 移动端
├── mybilibili-webview-app/   # Android WebView 壳工程
├── deploy/                   # 部署：k3s manifests + docker compose + 数据库备份 + SRS
├── docs/                     # 架构演进与实战踩坑文档
├── scripts/                  # 构建/启动脚本
└── old/                      # 历史遗留
```

## 架构

### 前端 / 客户端

| 端 | 技术栈 | 说明 |
|---|---|---|
| **Web 用户端** (`mybilibili-front/apps/web`) | Nuxt 4 (SSR) | 主站，SSR 渲染，traefik 按路径分流 |
| **管理后台** (`mybilibili-front/apps/admin`) | Vue3 + Vite + Element Plus | 运营/内容管理后台，路径 `/admin/`，端口 3100 |
| **UI 组件库** (`mybilibili-front/packages/ui`) | 自研组件库 | 多端共享 |
| **移动 H5** (`mybilibili-wap`) | Vue3 + Vite | 手机浏览器版，独立 build |
| **Flutter App** (`mybilibili-app-flutter`) | Flutter | iOS/Android/桌面 |
| **NativeScript** (`mybilibili-ns`) | NativeScript + Vue | 备选移动端方案 |
| **Android WebView** (`mybilibili-webview-app`) | Kotlin WebView 壳 | 包 H5 的原生壳 |

### 后端（Go 微服务，`mybilibili-go/`）

| 服务 | 职责 | 关键能力 |
|---|---|---|
| **core-service** | 核心服务 | 用户/登录注册/权限/分类/稿件/评论/动态/历史/上传与静态资源，HTTP :8080 + gRPC :9090 |
| **ai-service** | 字幕/AI 服务 | AI 字幕、AI 辅助，:8088/:9088 |
| **bili-proxy-service** | B 站代理 | 代理 B 站数据源，健康检查 `/api/v1/bili/health`，:8091 |
| **live-service** | 直播服务 | 直播间、语音房（WebSocket `/ws/`），:8087 |
| **msg-danmaku-service** | 弹幕/消息 | 弹幕、私信/消息中心（SSE `/sse/`），:8086/:9086 |
| **search-service** | 搜索/推荐 | 搜索、热门、推荐、创作者统计、画像，:8084/:9084 |
| **studio-service** | 创作中心 | 创作中心、投稿管理、素材，:8089 |
| **work-service** | 编排/任务 | 视频发布流水线编排（NATS 订阅） |
| **transcoder-service** | 转码服务 | 视频转码处理 |
| **minio** | 对象存储 | 封面/头像/稿件/视频文件 |

服务间通过 **gRPC** 通信（`mybilibili-go/proto/` 定义），异步任务走 **NATS JetStream**。HTTP 路由按 path 由 **Traefik IngressRoute** 分发，JWT 由各后端服务自行验证。

### 基础设施

| 组件 | 方案 | 拓扑 |
|---|---|---|
| **容器编排** | k3s | 3 节点：部署机(fnos 控制面) + 开发机(laptop) + 边缘 Worker，pod CIDR 192.168.100.0/22，service CIDR 192.168.104.0/24 |
| **数据库** | PostgreSQL 16 | 主从流复制：primary=fnos，replica=laptop |
| **缓存** | Redis | 主从：master=fnos，replica=laptop |
| **消息队列** | NATS JetStream | 3 副本 StatefulSet 集群 |
| **对象存储** | MinIO | 存储 images/manuscripts/avatars |
| **入口** | Traefik | 统一 IngressRoute，按 path 路由 + 限流/压缩/安全头/静态缓存 |
| **流媒体** | SRS | 直播/视频流服务 |
| **数据库备份** | `deploy/db/*.sql.gz` | 全量备份 |

### 部署布局（`deploy/k3s/base/`）

- `infra.yaml`：PostgreSQL 主从、Redis 主从、NATS StatefulSet 集群、MinIO
- `backend.yaml`：9 个业务 Deployment（core/ai/bili/live/msg-danmaku/search/studio/work/transcoder），全部 `replicas: 2` 跨节点分布
- `frontend.yaml`:web/admin 双副本
- `ingress.yaml`：Traefik IngressRoute + Middleware（限流、压缩、安全头、静态缓存）
- `pg-entrypoint.sh` / `pg-entrypoint.yaml`：PG 主从初始化与 pg_hba 授权
- `kustomization.yaml` + `config/*.env`：配置集中管理（common + prod 差异）

开发环境可用 `docker compose`（`deploy/` 下）一键起 17 个服务。

## 功能清单

### 账号与用户体系
- 注册 / 登录 / 登出，JWT 鉴权、登录态保持与刷新
- 忘记密码，支持找回重置
- 用户信息：头像、昵称、简介、性别、生日、签名等资料编辑
- 角色 / 权限体系，管理员角色（RBAC）
- 个人中心：主页、编辑资料、头像管理、登录日志
- 创作者身份识别（"成为创作者的 X 天"）

### 首页与内容分发
- 顶部导航（首页 / 直播 / 分区 / 搜索 / 投稿等）
- 轮播 Banner（`banner-images/home`，运营可配置多张轮播）
- 背景 Banner 大图
- 推荐视频流（`/recommend/hot`，瀑布流卡片）
- 热门视频（`/recommend`）
- 分区（category）浏览：分类列表 + 分类下视频
- 频道 / 分区页（wap 端 Channel）

### 视频与稿件
- 稿件列表、详情页展示（封面、标题、作者、播放/点赞/评论数、发布时间）
- 视频播放（清晰度选择、横屏/竖屏适配、播放器）
- 视频信息与元数据（/videos/、/covers/、/uploads/ 静态资源）
- 稿件上传（`UploadView`，创作中心内）
- 视频处理流水线：上传 → work-service 编排 → transcoder 转码 → 发布（NATS JetStream 驱动）
- 视频审核 / 内容审核队列
- 转码配置、视频过程管理（后台）
- 视频竖屏（VerticalFeed / VideoPlayer）

### 互动体系
- 点赞 / 取消点赞
- 收藏（favorites）、收藏夹管理（collections 列表 / 创建 / 编辑）
- 关注与粉丝：关注 / 取关，互关、粉丝列表、关注列表
- 观看历史（history）
- 评论（稿件评论区）
- 弹幕：视频弹幕（msg-danmaku-service），直播间弹幕，弹幕面板
- @ 提及、回复（ReplyList）、点赞通知（LikeList）

### 动态与社区
- 动态列表（关注的人动态流，dynamic）
- 动态详情、动态发布
- 用户空间（space）：主页、投稿列表、收藏、关注/粉丝、草稿、编辑资料

### 搜索与发现
- 全站搜索（search：`/search`）
- 搜索联想 / 热门搜索（`/search/hot` 热搜榜）
- 搜索结果筛选（SearchResults：分区/类型筛选）
- 推荐（`/recommend/`）
- 排行榜（ranking）
- 热度排行 / 分区排行（wap）

### 消息与通知中心
消息中心（message）分类型页签：
- 私信会话（ConversationList + ChatWindow，实时聊天）
- @ 提及（AtList）
- 回复（ReplyList）
- 点赞/喜欢（LikeList）
- 系统通知（SystemList）
- 站内信/订阅（NotificationList）
- AI 聊天助手（AiChatWindow）
- 消息设置（MessageSettings：通知开关）、消息侧边栏
- SSE 实时推送（`/sse/` → core）、私信/通知实时到达

### 直播（live）
- 直播列表（分区浏览：游戏/颜值/户外/音乐/聊天等）
- 直播间：
  - 直播播放器（`useLiveRoomPlayer`，清晰度 / 码率选择）
  - 主播关注（`useLiveRoomFollow`，是否关注 + 粉丝数、一键关注）
  - 房间会话（`useLiveRoomSession`：房间信息、当前用户、主播识别、分享房间）
  - 实时在线人数（`useLiveRoomRealtime` 实时观众数）
  - 房间聊天 / 弹幕（sendRoomMessage，WebSocket `/ws/`）
  - 连麦（`useLiveLinkmic`：语音连麦、麦位、音视频通话）
  - 麦克风 / 摄像头控制（Mute / VideoPlace）
- 直播推流（`/live/push`，主播推流端，SRS 承接）
- 直播回放（ReplayTab：回放列表、回放播放）
- 直播区域 / 分类页（wap live/Area）

### 创作中心（`/create-center/`）
- 首页仪表盘（CenterDashboard：作品概览、近期数据）
- 稿件管理（ManuscriptManager：自己投稿的增删改查、审核状态）
- 投稿（UploadView：上传稿件）
- 草稿箱（DraftsBox）
- 数据中心（DataCenterView：播放/互动数据报表）
- 粉丝管理（FansManager：粉丝列表/统计）
- 侧边导航（CenterSidebar）+ 创作者天数徽章

### 用户空间（个人主页 Profile）
- 主页（home）：个人资料、投稿、动态
- 投稿列表（submissions）
- 收藏（favorites）、收藏夹（collections）
- 关注 / 粉丝列表（following / followers）
- 动态（dynamic）
- 兴趣标签（interests）
- 设置（settings）、站内搜索

### 管理后台（`/admin/`，Vue3 + Element Plus）
- **登录与权限**：管理员登录（LoginView）、角色管理（RolesView）、管理员列表（AdminsView）、无权限页
- **核心业务管理**：
  - 视频稿件管理（ManuscriptsView）
  - 分类管理（CategoriesView）
  - 用户管理（UsersView）
  - 评论面板（CommentPanel）+ 评论审核
  - 弹幕面板（DanmakuPanel）+ 管理
- **运营配置**：
  - 轮播 Banner 管理（BannerImagesView）
  - 首页管理（IndexManagerView）
  - 推荐位配置（RecommendConfigView）
- **内容审核**：内容审核（ContentReviewView）、违禁词管理（ProhibitedWordsView）、举报中心
- **直播管理**：直播间管理（LiveRoomsView）
- **AI 能力**：AI 技能配置（AiSkillsView）、AI 用量（AiUsageView）、AI 对话面板（AdminAiChatPanel / FloatingButton）
- **字幕管理**：字幕管理（SubtitleManagementView）
- **系统**：系统通知（SystemNotificationManagerView）、操作审计日志（AuditLogsView）、API 管理（ApiManagementView）、登录日志（LoginLogsView）
- **任务/Pipeline**：视频处理任务（VideoProcessView）、转码配置（TranscodeConfigView）、定时任务（ScheduledTasksView）、运营任务（OperationTasksView）
- **客服**：用户在线客服聊天（CustomerChatView + SupportTicketsView 工单）
- **数据看板**：DashboardView

### 多端（移动）
- **H5 (wap)**：首页、视频播放（弹幕）、直播（分区/房间/回放）、搜索、排行、动态、空间（投稿/收藏/粉丝/历史/草稿）、消息（私信/通知）、创作中心、登录
- **Flutter App**：auth / home / search / live / video / dynamic / follow / message / creator / profile / upload 模块
- **NativeScript**：跨端备选
- **Android WebView**：H5 原生壳

### 平台能力
- 微服务 + gRPC 服务间通信（`proto/` 定义）
- NATS JetStream 流水线（视频发布/处理/进度主题）
- SSE 实时推送、WebSocket 通信
- Traefik 网关：限流（rate-limit）、压缩、安全头、静态资源永久缓存、按 path 路由
- 多副本高可用：业务、PostgreSQL、Redis、NATS 全部主从/集群冗余
- 数据备份 / 恢复（`deploy/db/` 全量备份）
- B 站数据代理：对接真实 B 站内容源（bili-proxy-service）

## 快速访问（生产 k3s）

```
用户端：    http://192.168.31.182/
管理后台：  http://192.168.31.182/admin/
SWAGGER/探活：  GET /api/v1/health
```

## 生成产物 / 镜像

业务镜像统一推送 GHCR（`ghcr.io/...`），由部署机执行 `docker pull` 并注入 k3s。CI 位于 `.github/workflows/`。

## 文档

`docs/` 记录了关键架构决策与实战踩坑：CSR→SSR 演进、首屏体积优化、SSR 取舍、构建部署优化、移动端三技术栈对比、企业级性能优化方法论等，详见各 `0X-*.md`。