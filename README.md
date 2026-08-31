# mybilibili · 仿哔哩哔哩全栈项目

一套完整仿 B 站的视频社区平台，覆盖用户端 Web（SSR）、管理后台、移动 H5、Flutter / NativeScript / Android WebView 多端。**当前主体为 Go 微服务架构**，容器化部署在三机 k3s 集群（部署机 + 开发机 + 边缘 Worker）。

## 三个版本

本项目完整实现过三代架构，代码均保留在仓库中：

| 版本 | 后端技术 | 状态 | 代码位置 |
|---|---|---|---|
| **Go 微服务**（当前主线） | Go + gRPC + NATS JetStream | ✅ 运行中 | `mybilibili-go/` |
| **Java 微服务**（历史） | Spring Cloud + Nacos + RocketMQ | 存档 | `old/springboot3/`、`old/microservices/` |
| **Java 单体**（历史） | Spring Boot + MyBatis + Redis + ES | 存档 | `old/java/` |

> 三代版本功能域一致（投稿/弹幕/直播/AI 等），演进方向是**技术栈现代化 + 资源效率 + 高可用**。详见 [Java 版本演进历史](./docs/Java版本-SpringCloud微服务历史.md)。

### Go 版 vs Java 微服务版

同为微服务形态，Go 版针对同域功能做了**极致的资源效率**改造，尤其适合边缘/轻量部署（本项目的 k3s 集群：

| 维度 | **Go 版（当前）** | Java 微服务版（历史） |
|---|---|---|
| **内存占用** | **极低**：单服务常驻内存 ≈ **几十 MB 级**，静态编译无 JVM 开销 | **高**：每个服务挂 JVM（堆 + 元空间 + GC 预留）通常数百 MB 起步 |
| **部署资源** | 10 个服务在 3 机 k3s 轻松多副本冗余，2 副本 × 10 服务容器总量可控 | 同数量服务多副本内存开销翻几倍，轻量机难以承受 |
| **启动速度** | **毫秒级**启动（原生编译），滚动更新秒级完成 | JVM 冷启动秒级~十秒级，回滚/扩展慢 |
| **镜像体积** | 单二进制静态编译，镜像 **几十 MB 量级**（多阶段构建） | JRE + 依赖 + 类库，镜像 GB 级 |
| **运行时依赖** | 无 GC 停顿、可预测延迟，天然适合长连接/高并发 IO | JVM GC 停顿 + 内存压力调优复杂 |
| **基础设施** | PostgreSQL + NATS JetStream（轻量）+ Redis，全主从/集群化 | MySQL + RocketMQ + Nacos，各自常驻内存开销大 |
| **脚手架** | `go vet` + `go build` 门禁，编译期类型/接口安全 | 运行时较多（反射/代理，MyBatis/Spring 容器） |
| **演进价值** | 保留 Java 版全部业务模型，去除重运行时 | 全业务域 + AI + 直播全栈的**功能原型**沉淀 |

> **一句话**：功能完整性继承自 Java 版，但要跑在 3 台轻量机上并做多副本高可用，需要 **Go 这种低内存、快启动的静态二进制**——这也是本项目迁移到 Go 的核心动力。

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
└── old/                      # Java Spring Cloud 历史版本（单体/微服务/直播桌面临播/OBS插件）
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
| **core-service** | 核心服务 | 用户/鉴权/分类/稿件/视频/评论/弹幕管理/动态/收藏/关注/历史/互动/分享/审核/举报/违禁词/客服工单/运营统计/定时与运营任务/SSE，HTTP :8080 + gRPC :9090 |
| **search-service** | 搜索/推荐/画像 | 视频/用户搜索、搜索建议、热搜榜、推荐(for-you/热门/相关)、创作者统计、用户画像、搜索引擎索引管理，:8084/:9084 |
| **msg-danmaku-service** | 弹幕/消息 | 弹幕(发送/获取/趋势/管理)、私信会话、@/点赞/回复/系统通知、未读数、SSE 弹幕与通知推送，:8086/:9086 |
| **live-service** | 直播 | 直播间 CRUD/开播停播状态/定时开播、SRS hook、WebSocket 房间会话、连麦(麦位/音视频)，:8087 |
| **ai-service** | 字幕/AI 中心 | 字幕(生成/上传/SRT导入/扫描/默认)、AI 摘要、AI 审核(评论/回复/举报/内容)、AI 助手对话、AI 客服技能与路由、用量统计，:8088/:9088 |
| **studio-service** | 创作中心 | 素材上传、导出任务，:8089 |
| **bili-proxy-service** | B 站代理 | 链接解析、流代理，对接真实 B 站数据源，:8091 |
| **work-service** | 编排/任务 | NATS 订阅视频处理流水线：转码/抽音/生成字幕/AI 摘要，自动链式编排 + 进度上报 |
| **transcoder-service** | 转码服务 | 视频方向检测/转码/抽音 |
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

> 以下功能点逐一列出（后端按 `mybilibili-go` 各服务注册的全部 HTTP/gRPC 端点 + 89 张数据表 + 前端各端页面组件整理，详见服务路由章节）。

### 1. 账号与用户体系（core-service：user / auth）
- 注册、登录、登出，JWT 签发与刷新（`/user/token/refresh`）、`/auth/verify` 鉴权验证
- 邮箱验证码（`/user/email/code`、`/user/email/verify`）
- 忘记密码 / 重置（`/user/password/forgot`）
- 图形/行为验证码（`/captcha/`）
- 用户资料：昵称、头像（`/user/me/avatar`）、性别、生日、简介、个性签名等
- 默认头像获取（`/user/default-avatar`）、批量用户信息（`/user/batch`）
- 看空间隐私设置（`/user/privacy/`）、消息偏好设置（`/user/settings/message`）、创作者设置（`/user/settings/creator`）
- 兴趣标签（`/user/tags`）、个人标签管理
- **经验等级系统**（`user/experience.go`）：按公式 `floor(100*level^1.8)` 升级，完成任务/行为自动 AwardExperience，跨阈值自动升级并结转经验
- 我的信息（`/user/me`）、登录日志（`/user/login-logs` 及计数）、钉钉/置顶视频（`/user/pinned-video`）

### 2. 首页与内容分发（core + search）
- 顶部导航（首页 / 直播 / 分区 / 搜索 / 投稿 / 消息）
- 轮播 Banner 管理（`banner-images/home` 多张 + `banner-images/background` 背景大图），运营可配 `sortOrder`/`status`/`type`
- 普通 Banner 位（`/banner/`）
- 推荐流：热门（`/recommend/hot`）+ **个性推荐**（`/recommend/for-you`）+
- **相关推荐**（`/recommend/related/`，稿件详情页"相关推荐"）
- 分类体系（`/category/`）：多级分类、分类下内容

### 3. 视频、稿件与发布流水线（core + work + transcoder + studio）
- 稿件（manuscript）齐全的 CRUD（`/manuscript/`），含 `video_tags`、`videos` 多视频、封面、分 P
- **分块断点上传**（前端 `useChunkedManuscriptUpload`：切片上传、进度、续传）
- **视频处理流水线**（NATS JetStream 驱动，`work-service/pipeline.go`）：
  转码（方向检测/转码）→ 抽音频 → 生成字幕（ASR）→ **AI 智能摘要**，`AutoChain` 自动链式处理，进度走 **SSE**（`/video/process/sse/`）
- 转码服务（`/transcode`）：方向检测、转码、抽音（transcoder-service）
- 稿件审核状态机（`manuscript_status_events`）、编辑版本（`manuscript_edit_versions`）、每日指标（`manuscript_daily_metrics`）
- 视频信息获取（`/video/`）、观看历史（`watch_history`，记录播放进度 `progress_seconds`、续播）
- 上传会话（`upload_sessions`）、素材上传（`/studio/assets/upload`）、导出任务（`/studio/export-tasks`）
- 视频处理管理端点（`/video/process/admin/*`：当前任务、队列、统计、SSE 流）
- 分享（`/share/`）：不同渠道分享计数（channel + ip 追踪）

### 4. 互动体系（core：interaction / favorite / follow / comment）
- 点赞/取消、批量点赞计数（`/interaction/`、`/comment/batch-like-counts`）
- 收藏：收藏夹（`favorite_folders`）创建/编辑/删除，收藏稿件（`favorite_folder_videos`），`/favorites/check` 判断是否已收藏，`/favorites/list` 收藏列表
- 关注/粉丝：关注、取关、关注检查、我的粉丝/关注列表（`/follow/me/followers`、`/follow/me/following`）、用户粉丝/关注（`/follow/user/`）
- 评论系统（comment）：发表评论（`/comment/add`）、回复（`/comment/reply`）、评论列表、回复列表、评论/回复点赞（`/comment/{id}/like`）
  - 评论限流（`comment_rate_limit.go`）、评论审核状态、创作者视角评论管理（`/creator/comments`）、公开 API（`public_api_handler.go`）

### 5. 动态与社区（core：social / dynamic）
- 动态发布/删除、关注动态流（`/dynamic/all`）
- 动态点赞（`/dynamic/like/`）、分享（`/dynamic/share/`）
- 动态评论（发布/删除/列表/回复/点赞/计数，`/dynamic/comment/*`）
- 交互记录（`user_interactions`）、点赞关系

### 6. 搜索与发现（search-service）
- **多路搜索**：视频搜索（`/search/videos`）、用户搜索（`/search/users`）、搜索建议（`/search/suggest`）
- 热搜：榜单（`/search/hot`、`/search/hot/rank`）、关键词获取/删除/自增热度/分数、过期清理（`/hot/clean-expired`）
- 推荐：热门 `for-you`、相关、创作者推荐
- **推荐位配置**（`/search/admin/recommend-config` 管理 + 重置）
- **搜索引擎索引管理**：全量重建、增量、刷新、状态、校验、批量索引（`/search/admin/index/*`）
- 创作统计（search-service）：
  - 创作者总览（`/creator/stats/overview`）、粉丝排行（`fans-ranking`）、粉丝趋势（`fans-trend`）、稿件趋势（`manuscript-trend`）、最新评论（`latest-comments`）、总趋势（`trend`）、榜单（`ranking`）
- **用户画像**（`/profile/`、`/profile/record/`，`clients/profile_recorder.go`）

### 7. 消息与通知（msg-danmaku-service）
- **私信会话**：会话列表（`/message/conversations`）、会话内消息（`/message/conversations/`）、指定用户会话（`/message/user/`）、发送（`/message/send`）
- 未读：未读数（`/message/unread`）、未读统计（`/message/unread/counts`）、按会话未读（`/message/conversation/unread/`）、批量已读（`/message/batch/read`）
- **通知类型**：@提及（`/message/at`）、点赞（`/message/likes`）、回复（`/message/replies`）、系统通知（`/message/system`）、通知落库（`notifications` 表）
- 消息设置（`/message/settings`，开关各类型通知）
- **SSE 实时推送**：弹幕流（`/sse/danmaku`）、通知流（`/sse/notification`）、视频处理进度（core `/video/process/sse/`）
- 弹幕：发送（`/danmaku/send`）、按视频获取（`/danmaku/video/`）、批量计数（`/danmaku/batch-count`）、趋势（`/danmaku/trend`）、创作者弹幕管理（`/creator/danmaku/`）
- 管理端点：弹幕审核（msg-danmaku 的 `/message/admin/`）

### 8. 直播（live-service + 前端全量 composables）
- 直播间管理：创建/详情/按主播查询/分页列表/更新/开播-停播状态（`/live/room*`）
- **定时开播**（`ScheduleRoom`，设置 scheduledAt）、直播状态机（`RoomStatusLive` 等）、SRS 事件回调（`/live/srs/hook`：推流通知自动置为开播/停播）
- 直播管理后台（`/live/admin/`）
- 直播间 WebSocket 会话（`live_ws.go`）：房间内实时消息
- **连麦系统**（`linkmic.go` + `linkmic_handler.go` + `live_linkmic`/`live_seats` 表）：
  - **麦位管理**：seat_index 位置、加入/离开、静音（muted）、状态
  - **WebRTC 基础 P2P 网格**（前端 `usePeerConnectionMesh`：offer/answer/ice-candidate、ICE 重启重协商）
  - **WHIP 协议推流**（`useWhipPublisher`，连接 SRS 的 `/rtc/v1/whip/stream/`，`getDisplayMedia` 支持**屏幕共享**）
- 直播推流房间（`/live/push`）：开播设置（`useLivePushRoomSetup`）、房间控制（`useLivePushRoomControls`：麦克风/摄像头/Mute/开播）、观众监控（`useLivePushAudienceMonitor`：实时在线人数）
- 观众端直播间（LiveRoomView）：播放器（`useLiveRoomPlayer`：清晰度、直播/回放 Tab、回放列表与播放）、关注（`useLiveRoomFollow`）、实时数据（`useLiveRoomRealtime`：在线人数、房间消息）、互动（`useLiveRoomInteractions`）、会话（分享房间、离线房间状态判断 `isOfflineRoomStatus`）
- 直播列表/分区（`/live/rooms`、`/live/room/list`，wap live/Area、live/List）

### 9. 消息中心 + AI 客服（web 端 message + ai-service）
- 消息中心分类型页签（私信/提及/回复/点赞/系统/通知）
- **AI 聊天助手**（AiChatWindow）、**AI 客服技能系统**（ai-service skills）：
  - 技能 CRUD（`/ai/skills`、`/ai/skills/`）、客服技能默认技能（`/ai/skills/customer-service/defaults`）
  - 客服路由（`/ai/skills/customer-service/route` 及 route-test 测试）
  - AI 客服对话（`/ai/customer/chat`）、会话历史（`/ai/customer/history/`）、转人工（`/ai/customer/transfer`）、专家座席（`/ai/customer/`）
- AI 助手对话（`/ai/assistant/send`）、AI 会话/消息落库（`ai_sessions`、`ai_chat_messages`）
- AI 技能绑定（`/ai/bindings/`）、配置管理（`/ai/configs/`、配置测试）、用量统计（`/ai/usage/`，`ai_usage_logs` 按月分区）

### 10. AI 能力中心（ai-service）
- **字幕全链路**：
  - 生成字幕（`/subtitle/generate`，ASR）
  - 字幕上传（`/subtitle/upload`、`/subtitle/upload-srt`、Srt 导入）
  - 系统字幕导入（`/subtitle/import-system`）、待处理队列（`/subtitle/pending`）、扫描（`/subtitle/scan/`）、设为默认（`/subtitle/set-default`）、按视频查（`/subtitle/video/`）、全部视频（`/subtitle/videos`）
- **AI 内容审核**：评论审核（`/ai/review/comment`）、回复审核（`/ai/review/reply`）、举报审核（`/ai/review/report`）、内容审核（`/ai/review/content`）
- **AI 摘要**（`/ai/summary/`，视频摘要）
- AI 配置与用量管理（前端 adminAi.ts）

### 11. AI 数据处理与画像（search-service + ai-service）
- 用户画像记录（`/profile/record/`）
- B 站代理（bili-proxy-service）：链接解析（`/api/v1/bili/resolve/`）、流代理（`/api/v1/bili/stream/`），对接真实 B 站内容源

### 12. 审核与内容安全（core：moderation）
- 内容审核后台（`/moderation/admin/comments|danmaku|prohibited-words|report`）
- 违禁词库：CRUD + **批量导入**（`/moderation/admin/prohibited-words/batch-import`）
- 审核状态机（`content_reviews`）、举报（`/report/submit`）、评论/弹幕审核（`moderation_admin`）

### 13. 平台统计与运营（core）
- 平台统计（`/statistics/`）
- **定时任务系统**（`scheduled_tasks` 表 + 后台 `/admin/scheduled-tasks`：列表、启用 toggle、手动 trigger）
- **运营任务**（`operation_tasks` + `/admin/operation-tasks`）
- 安全设置（`/admin/security-settings`）、**存储迁移**（`/admin/storage/migrate`：本地/MinIO 迁移）
- 转码配置管理（`/admin/transcode-config`，`transcode_config` 表）

### 14. 客服工单系统（support）
- 用户工单（`support_tickets`）
- 客服会话（`/operation/message/tickets/customer-session`、`/operation/tickets/session/`）
- 工单后台管理（`/support/admin/tickets`、`/operation/tickets`）

### 15. 管理后台（`/admin/`，Vue3 + Element Plus，`apps/admin/app/views/admin/`）
- **登录与权限**：管理员登录（LoginView）、角色管理（RolesView、`roles`/`permissions`/`role_permissions`/`admin_user_roles` 表）、管理员列表（AdminsView）、无权限页；权限用 `/admin/permissions`
- **核心业务管理**：
  - 视频稿件管理（ManuscriptsView）＋稿件后台端点（`/manuscript/admin/all|pending|processing|statistics`）
  - 分类管理（CategoriesView）
  - 用户管理（UsersView、`/user/admin/list`、用户后台端点）
  - 评论面板（CommentPanel）＋评论管理（`/admin/comments`）
  - 弹幕面板（DanmakuPanel）＋弹幕审核
  - 视频管理（VideoProcessView，视频处理队列）
- **运营配置**：
  - 轮播 Banner 管理（BannerImagesView）
  - 首页管理（IndexManagerView）
  - 推荐位配置（RecommendConfigView）
  - 搜索索引/热搜管理（search admin）
- **内容审核**：内容审核（ContentReviewView）、违禁词管理（ProhibitedWordsView）、举报中心
- **直播管理**：直播间管理（LiveRoomsView）
- **AI 能力**：AI 技能配置（AiSkillsView）、AI 用量统计（AiUsageView）、AI 对话面板（AdminAiChatPanel / 悬浮按钮）
- **字幕管理**：字幕管理（SubtitleManagementView）
- **系统**：系统通知（SystemNotificationManagerView）、操作审计日志（AuditLogsView，`audit_logs` 按月分区）、API 管理（ApiManagementView，`api_configs`/`ai_api_configs`）、登录日志（LoginLogsView）、安全设置
- **任务/Pipeline**：视频处理（VideoProcessView）、转码配置（TranscodeConfigView）、定时任务（ScheduledTasksView）、运营任务（OperationTasksView）
- **客服**：在线客服会话（CustomerChatView）、工单（SupportTicketsView）
- **数据看板**：DashboardView

### 16. 数据中心（web datacenter 组件）
- 数据概览卡（DataCards）
- 数据总览面板（DataOverviewPanel）
- **账号诊断**（AccountDiagnosisPanel）
- **粉丝分析**（FanAnalysisPanel）
- **稿件分析**（ManuscriptAnalysisPanel）
- 趋势图（TrendChart）

### 17. 创作中心（`/create-center/`，`CreateCenterView`）
- 侧边导航（CenterSidebar）＋头部"成为创作者第 X 天"徽章
- 首页仪表盘（CenterDashboard）
- 稿件管理（ManuscriptManager）＋**稿件编辑弹窗**（ManuscriptEditDialog）
- 投稿（UploadView，分块上传）
- 草稿箱（DraftsBox）
- 数据中心（DataCenterView）
- 粉丝管理（FansManager）
- 收藏夹管理（CollectionManager / CreateDialog / AddVideoDialog / Detail / List）

### 18. 用户空间（个人主页，`profile/personal-center`）
- 主页 Tab（ProfileHeader + ProfileSidebar + Tab* 全家）
- 投稿列表（Submission）、视频列表（VideoList）
- 收藏（Favorite）/ 收藏夹（CollectionGrid / Detail / EditDialog / AddVideoDialog / FavoriteFolder*）
- 关注/粉丝（FollowList / Fans）、动态（DynamicList）
- 空间搜索（TabSearch）、设置（TabSettings）、兴趣（InterestsPanel）
- 创作中心入口、个人中心（home / info / avatar / login-logs）

### 19. 搜索/社交前端组件
- 搜索页（SearchView + SearchResults：视频/用户/分区筛选）、热搜
- 动态页（DynamicView + 详情 DynamicDetailView）

### 20. 多端（移动）
- **H5 (wap，`/m/*` 全部路由)**：首页、频道（channel）、视频播放（横屏）+ **竖屏视频流**（vertical）、直播（areas/list/room）、搜索、热搜、排行、动态、空间（投稿/收藏/粉丝/历史/草稿/资料编辑）、消息（私信 chat/通知 notify）、创作中心、登录
- **Flutter App**：auth / home / search / live / video / dynamic / follow / message / creator / profile / upload 模块
- **NativeScript**：跨端备选
- **Android WebView**：H5 原生壳

### 21. 平台能力
- 微服务 + gRPC（`proto/` 定义，服务间 gRPC 客户端：`clients/ai_client`、`msg_danmaku_client`、`search_client`、`profile_recorder`）
- NATS JetStream 流水线：转码/抽音/字幕/AI 摘要/进度主题，事件发布（`events/event_publisher.go`）
- SSE 实时推送（弹幕/通知/处理进度）、WebSocket（直播间）
- Traefik 网关：限流、压缩、安全头、静态资源永久缓存、按 path 路由（含 live 反向代理、bili 反向代理、`/uploads/` 本地/MinIO 双后端）
- 多副本高可用：业务、PostgreSQL、Redis、NATS 全部主从/集群冗余
- 数据备份 / 恢复（`deploy/db/` 全量备份 + `README` 说明）
- 数据库 89 张表：用户/稿件/视频/评论/弹幕/动态/直播(rooms+seats+linkmic)/收藏/关注/私信/通知/审核/举报/违禁词/热搜/上传会话/订阅/分享/客服工单/操作任务/定时任务/审计日志/登录日志/AI(会话/消息/技能/配置/用量/Api)/画像/交互/短视频 tags 等，多表按月分区

## 快速访问（生产 k3s）

```
用户端：    http://192.168.31.182/
管理后台：  http://192.168.31.182/admin/
SWAGGER/探活：  GET /api/v1/health
```

## CI/CD

### 流水线总览（GitHub Actions，`.github/workflows/ci.yml`）

`push` 到 `main` 自动触发（也可 `workflow_dispatch` 手动指定服务），产物全部推送到 **GHCR**（`ghcr.io/chentianxiong123/mybilibili-*`）。

```
push main ──► test(go vet + go build 门禁) ──► build-backend(8 服务镜像) ──► GHCR
                          │
                          └───────────────► build-frontend(web/admin 镜像) ──► GHCR
```

| Job | 内容 | 说明 |
|---|---|---|
| **test** | `go vet` + `go build`（逐模块） | 质量门禁，失败则不构建镜像 |
| **build-backend** | 8 个后端服务镜像 | core / search / msg-danmaku / live / ai / studio / work / bili |
| **build-frontend** | 2 个前端应用镜像 | web（:3200）/ admin（:3100） |

### 镜像与版本策略

- **镜像仓库**：`ghcr.io/chentianxiong123/mybilibili-{core,search,msg-danmaku,live,ai,studio,work,bili,web,admin}`
- **tag 策略（分场景）**：
  - `:latest` —— 仅开发机 compose / 追溯用，不作为生产引用
  - `:$git-sha` —— CI 随构建推送，仅作版本追溯（GHCR 上偶发不完整，已不用作生产引用）
  - **生产部署：`@sha256:` 不可变 digest** —— 在 `deploy/k3s/base/kustomization.yaml` 的 `images:` 段**单点锁定**，升级/回滚只改一个 digest
- **发布流程**：push main → CI 构建并推 `latest`+`git-sha` → 取出新镜像 digest → 更新 `kustomization.yaml` 一处 → `kubectl apply -k` 滚动更新
- **构建缓存**：`type=gha` 层缓存（`cache-from/cache-to`），增量构建提速
- **多阶段构建**：`mybilibili-go/Dockerfile` 按 `SERVICE` build-arg + `target` 选服务，复用公共 base
- **节点绑定**：k3s 两节点拓扑约束——数据层主从按 nodeSelector 钉死节点，业务双副本用 topologySpreadConstraints 强制分散（fnos+laptop 各一）

### 边界说明

- **transcoder 不打镜像**：裸跑宿主机使用系统 FFmpeg（无容器），故不在矩阵中
- **wap 不打镜像**：将打包为移动 App
- 构建仅覆盖 `mybilibili-go/`、`mybilibili-front/`、CI 自身变更（`paths` 过滤，减少无效构建）
- 部署侧：CI 推镜像后更新 kustomization digest → `kubectl apply -k`（fnos），业务多副本滚动更新

## 文档

`docs/` 记录了关键架构决策与实战踩坑：CSR→SSR 演进、首屏体积优化、SSR 取舍、构建部署优化、移动端三技术栈对比、企业级性能优化方法论等，详见各 `0X-*.md`。

### 历史版本

- **[Java 版本演进历史](./docs/Java版本-SpringCloud微服务历史.md)**：单体版（Spring Boot 2.7）→ 全量微服务（Spring Cloud + Nacos + RocketMQ）→ 聚合微服务（Spring Boot 3.2）。含全业务域、AI（Whisper 字幕 + DeepSeek 摘要/客服/审核）、直播全栈（SRS + WebRTC 视频会议 + 直播桌面 App + OBS 插件）、前端三端。代码保留在 `old/`。