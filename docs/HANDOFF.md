# mybilibili-cloud 项目交接文档

> 最后更新: 2026-06-01 | JDK: Corretto 17

---

## 一、项目概览

Spring Cloud 微服务仿 B 站平台，14 个后端模块 + 2 个主要前端（用户端 + 管理端，仓库内另有 WAP 目录）。

| 模块 | 端口 | 职责 |
|------|------|------|
| gateway | 8080 | WebFlux 网关，JWT 鉴权 + 路由 + Sentinel 限流 |
| user | 8081 | 用户/管理员/角色/权限/JWT 双 token |
| video | 8082 | 稿件/视频/分类/轮播图/播放量防刷 |
| danmaku | 8083 | 弹幕 + WebSocket 实时广播 |
| search | 8084 | ES 搜索/推荐/热词/索引管理/推荐配置 |
| comment | 8085 | 评论/举报/违禁词/内容审核 |
| interaction | 8086 | 动态/收藏/合集/观看历史/用户画像 |
| message | 8087 | 私信/系统通知 + WebSocket 实时推送 |
| ai | 8088 | AI 字幕/总结/客服/技能/渠道 |
| live | 8089 | 直播间/连麦/会议/WebSocket |
| analytics | 8090 | 创作者数据中心/趋势统计/排行 |
| common | - | 共享实体/工具/Security/Storage/全局异常处理 |
| mq | - | RocketMQ 消息定义 |

前端: `mybilibili-web`(用户端 Vite+Vue3) / `mybilibili-admin-web`(管理端)

服务边界: 业务服务之间不再通过 Maven 直接依赖对方模块；同步调用走 Feign，异步协作走 RocketMQ，共享类型只放在 `common` / `mq`。

---

## 二、基础设施层（已实现）

### 2.1 Spring Security RBAC 权限体系

**网关层 (WebFlux)**
- `GatewaySecurityConfig` / `JwtReactiveAuthenticationManager` / `BearerTokenServerAuthenticationConverter` / `GatewayAuthorizationManager`
- `GatewayRequestPolicy`: fail-closed 路径分类（公开 / 需登录 / 需管理员 / 需超管）
- 权限码校验: 管理员 token 携带 `permissions` claim，网关解析为 `PERMISSION_xxx` authority

**服务层 (Servlet)**
- `SecurityConfig` + `JwtFilter`，authority 包括 `ROLE_USER` / `ROLE_ADMIN` / `ROLE_SUPER_ADMIN` / `PERMISSION_xxx`
- `JwtRequestPolicy`: 方法+路径级权限规则

**17 个权限码**: user/video/comment/category/tag/review/statistics/role/admin/security/live/meeting/storage/banner/search/ai/message:manage

### 2.2 JWT 双 Token 无感刷新

- **Access Token**: 2 小时有效期，`type=access`
- **Refresh Token**: 30 天有效期，`type=refresh`
- **端点**: `POST /api/user/token/refresh`（公开路径，无需鉴权）
- **前端拦截器**: 401 时自动用 refreshToken 刷新，并发请求队列化等待，刷新失败才清除登录态
- **文件**: `JwtUtils.java`(generateToken/generateRefreshToken/parseToken)

### 2.3 Sentinel 网关限流

- 依赖: `spring-cloud-starter-alibaba-sentinel` + `spring-cloud-alibaba-sentinel-gateway`
- `SentinelRuleConfig`: 按 API 分组限流（video 200/s, user 100/s, search 100/s, interaction 150/s, auth 10/s）
- 自定义 429 响应体: `{"code":429,"message":"请求过于频繁，请稍后再试"}`
- 支持 Sentinel Dashboard (`127.0.0.1:8858`) 动态调整

### 2.4 MinIO 对象存储

- `StorageService` 接口 + `MinioStorageService` 实现
- `StorageKeys`: 集中定义所有存储路径 key
- `StorageMigrationTool`: 本地文件到 MinIO 迁移

### 2.5 分片上传

- `ManuscriptUploadSessionService`: Redis 存储会话，串行分片
- 前端: `useChunkedManuscriptUpload.js` composable
- 视频格式白名单: mp4/avi/mkv/mov/flv/wmv/webm/m4v/ts/mpeg/mpg/3gp

### 2.6 全局异常处理 + 输入校验

- `GlobalExceptionHandler`: 统一处理 MethodArgumentNotValid / ConstraintViolation / MissingParam / TypeMismatch / NotReadable / BusinessException
- 校验注解覆盖: LoginDTO / UserDTO / SendMessageDTO / AdminLoginDTO / VideoUploadDTO / UserUpdateDTO / ManuscriptUploadDTO / NotificationDTO
- 控制器 @Valid: register / login / sendMessage / updateUser / adminLogin

### 2.7 服务健康探针 + Docker 编排

- 所有运行时服务已加入 `spring-boot-starter-actuator`
- 暴露端点: `/actuator/health`, `/actuator/health/liveness`, `/actuator/health/readiness`, `/actuator/info`
- Gateway 与服务层安全策略只放行 health/info，未开放 Prometheus 等更深观测端点
- `scripts/docker-compose-infra.yml`: 基础设施组，统一使用 `mybilibili-net`，持久化目录均在 `D:\DockerFiles\mybilibili`
- `scripts/docker-compose-infra.yml` 直接挂载仓库内的 `scripts/rocketmq/broker.conf` 和 `srs.conf`
- AI 处理目录默认改为 `D:/DockerFiles/mybilibili/processing`
- 本地 Whisper 模式使用 `D:\DockerFiles\mybilibili\whisper`
- MinIO 支持 `minio.public-endpoint`，容器内服务访问 `http://minio:9000`，返回给浏览器的 URL 使用 `MINIO_PUBLIC_ENDPOINT`

---

## 三、本次新增功能（2403da7 ~ 41ff715，共 12 次提交）

### 3.1 WebSocket 实时通知推送

**后端 (mybilibili-message)**
- `NotificationSessionRegistry`: userId → WebSocketSession 映射（单会话覆盖）
- `NotificationWebSocketHandler`: WS 端点 `/ws/notification`，JWT query param 认证
- `NotificationPushService`: 推送通知 + 未读数给在线用户
- `WebSocketConfig`: 注册 WS 端点
- `MessageServiceImpl`: 所有通知方法（私信/点赞/回复/系统/广播）发送后自动 WS 推送

**前端**
- `composables/useNotificationWs.js`: 全局 WS 连接管理，ElNotification toast
- `AppHeader.vue`: WS 连接 + 未读数实时同步，保留 60s 辅助轮询

### 3.2 弹幕实时推送

**后端 (mybilibili-danmaku)**
- `DanmakuSessionRegistry`: 按 videoId 分房间管理 WS sessions
- `DanmakuWebSocketHandler`: 处理连接/断开/ping/弹幕发送
- `DanmakuBroadcastService`: 广播弹幕给同房间所有观看者
- `DanmakuServiceImpl`: REST 发弹幕时也触发广播（双通道）

**前端**
- `composables/useDanmakuWs.js`: 连接/断开/接收弹幕
- `VideoView.vue`: 视频加载后连接 WS，收到实时弹幕时 emit 到 artplayer，切 P 时重连

### 3.3 用户画像 + 个性化推荐

**用户画像 (MongoDB)**
- `UserProfile` entity: `user_profiles` collection，分类权重 + 标签权重 + 偏好时长 + 活跃时段 + 行为计数
- `UserProfileRepository`: Spring Data MongoDB
- `UserProfileServiceImpl`: 行为驱动更新（观看 1.0 / 点赞 3.0 / 收藏 5.0），权重衰减 0.95，标签上限 50
- 行为采集集成: `WatchHistoryServiceImpl.recordWatch` / `VideoInteractioncollectVideo` 自动更新画像
- `UserProfileController`: REST 接口供内部 Feign 调用

**推荐算法 (mybilibili-search)**
- `UserProfileClient` Feign: search 服务获取 interaction 服务的画像
- `VideoRecommendServiceImpl.getRecommendedVideosForUser`: 取 top N 分类 + 标签构建 ES 加权 should 查询，不足时热门补充，附带个性化推荐理由
- `VideoRecommendServiceImpl.getRelatedVideos`: 基于当前视频标题/描述/标签/分类的 ES 相似度查询

**推荐配置管理**
- `RecommendConfig` entity (MongoDB): 画像衰减/行为权重/boost/限流 QPS 等 16 项可调参数
- `RecommendConfigService`: 带缓存的配置读取，支持动态更新
- `RecommendConfigController`: 管理端 GET/PUT/RESET 接口
- 管理后台 `RecommendConfigView.vue`: 用户画像参数 / 推荐算法参数 / 防刷限流三大分区

**推荐冷启动**
- `UserDTO.interestTags`: 注册时可选兴趣标签
- `UserProfileController.initProfileWithTags`: 初始化画像，选中标签权重 5.0
- `RegisterView.vue`: 注册成功后弹出兴趣标签选择弹窗（32 个标签，最多选 10 个）

**首页个性化**
- `HomeView.vue`: 登录用户优先调用 `/api/recommend/for-you`，未取到个性化数据时展示默认列表
- `recommend.js`: `getRecommendedVideos` 指向 `/recommend/for-you`

**个人画像展示页**
- 路由: `/profile/:id/interests`
- `InterestsPanel.vue`: 标签云、分类偏好条形图、观看时段分布图、行为统计卡片
- 网关路由: `/api/profile/**` → interaction 服务，需登录

### 3.4 播放量防刷

- Redis SETNX 去重: `view:dedup:{manuscriptId}:{viewerKey}`，30 分钟 TTL
- `viewerKey`: 登录用户用 `u:{userId}`，未登录用 `ip:{clientIp}`
- `ManuscriptServiceImpl.incrementViewCount(manuscriptId, viewerKey)`
- `ManuscriptController.getClientIp`: X-Forwarded-For → X-Real-IP → RemoteAddr

### 3.5 创作者统计服务拆分

- 新增 `mybilibili-analytics` 服务，端口 `8090`，承接 `/creator/stats/**`
- 网关 `/api/creator/stats/**` 已从 `mybilibili-video` 改路由到 `mybilibili-analytics`
- `GatewayRequestPolicy` 将创作者统计归为登录态接口，非公开、非管理员接口
- `SentinelRuleConfig` 新增 `creator-api` 限流分组
- `video` 模块已移除 `CreatorStatsController` / `CreatorStatsService` / `CreatorStatsMapper`，不再承载创作者数据中心
- `video` 不再直接写创作者统计表，稿件状态/播放增量通过 `manuscript-analytics-topic` 发往 `analytics`
- `video-media` / `ai` 视频处理状态通过 `video-process-analytics-topic` 发往 `analytics`，由 `video_process_events` 保存转码/音频/字幕/AI 摘要处理流水
- `interaction` / `comment` / `danmaku` 的点赞、投币、收藏、分享、评论、弹幕动作也会汇入 `manuscript_daily_metrics`
- 视频处理控制面和进度面已统一回到 RocketMQ；`video-process-topic` 负责任务调度，`video-process-progress-topic` 负责进度/SSE 广播，已废弃 `ai` 内部 `video_pipeline_tasks` DB pipeline，避免出现第二套任务调度语义
- 新增趋势统计表: `manuscript_status_events` / `manuscript_daily_metrics` / `video_process_events`，迁移文件 `init/V13__creator_analytics_events.sql`

### 3.6 搜索联想词

- 后端 `VideoSearchServiceImpl.suggest`: `match_phrase_prefix`(title, boost 3.0) + `prefix`(tags, boost 2.0) + `match_phrase_prefix`(description, boost 1.0)
- 前端 `AppHeader.vue`: 输入时 300ms 防抖调用 `searchApi.getSearchSuggestions`，显示联想词下拉

### 3.7 网关路由新增

- `/ws/notification` → mybilibili-message
- `/ws/danmaku` → mybilibili-danmaku
- `/api/profile/**` → mybilibili-interaction
- `/api/creator/stats/**` → mybilibili-analytics
- 以上路径均加入公开路径或需登录路径

---

## 四、测试状态

```
JDK: Corretto 17.0.16
mvn compile → BUILD SUCCESS (全量)
mvn test → BUILD SUCCESS
Tests run: 56, Failures: 0, Errors: 0
mybilibili-web vite build → ✓ built in ~25s
mybilibili-admin-web vite build → ✓ built in ~19s
```

本次服务拆分回归（2026-06-01）：

```powershell
mvn -pl mybilibili-analytics,mybilibili-gateway,mybilibili-video -am test
# BUILD SUCCESS, Tests run: 28, Failures: 0, Errors: 0

mvn -pl mybilibili-ai,mybilibili-analytics,mybilibili-gateway,mybilibili-video -am test
# BUILD SUCCESS, Tests run: 28, Failures: 0, Errors: 0

mvn -pl mybilibili-gateway,mybilibili-user,mybilibili-video,mybilibili-danmaku,mybilibili-search,mybilibili-comment,mybilibili-interaction,mybilibili-message,mybilibili-live,mybilibili-ai,mybilibili-analytics -am test
# BUILD SUCCESS

docker compose -f scripts/docker-compose-infra.yml config --quiet
# 配置渲染通过

pnpm run build  # mybilibili-web / mybilibili-admin-web / mybilibili-wap
# 三套前端均构建通过；存在 chunk size 与 Sass deprecation warning，非阻断

mvn -DskipTests package
# BUILD SUCCESS，14 个 Maven 模块 jar 产出成功
```

---

## 五、构建命令

```powershell
$env:JAVA_HOME='D:\Program Files (x86)\corretto-17'
$env:Path="$env:JAVA_HOME\bin;$env:Path"

# 全量编译
mvn compile

# 全量测试
mvn test

# 构建后端 jar
mvn -DskipTests package

# 基础设施容器
docker compose -f scripts/docker-compose-infra.yml up -d

# 单模块
mvn -pl mybilibili-gateway -am test

# 前端
cd mybilibili-web && npx vite build
cd mybilibili-admin-web && npx vite build
```

---

## 六、已知待办

### P0 — 生产必备
- [ ] Docker Compose 端到端实跑验收（infra/services 编排文件已落地，仍需按真实 `.env` 和数据迁移结果完整启动一次）
- [ ] 视频处理流水线增强（当前 FFmpeg 转码/音频提取归属 `video-media`，字幕/摘要归属 `ai`，任务和进度已走 RocketMQ；后续仍缺更细的分布式调度/优先级编排）

### P1 — 体验提升
- [ ] 单元测试覆盖（目前仅 56 个测试，大部分模块无测试）
- [ ] 分布式链路追踪（SkyWalking / Zipkin）
- [ ] CI/CD（GitHub Actions / Jenkins）
- [ ] 集中日志（ELK / Loki）

### P2 — 锦上添花
- [ ] 虚拟币充值/打赏体系
- [ ] 创作者激励/收益系统
- [ ] 移动端适配 / PWA
- [ ] 国际化 i18n
- [ ] 内容审核 AI 流水线

---

## 七、关键文件索引

### 实时通信
- `mybilibili-message/src/main/java/.../websocket/NotificationSessionRegistry.java`
- `mybilibili-message/src/main/java/.../websocket/NotificationWebSocketHandler.java`
- `mybilibili-message/src/main/java/.../websocket/NotificationPushService.java`
- `mybilibili-danmaku/src/main/java/.../websocket/DanmakuSessionRegistry.java`
- `mybilibili-danmaku/src/main/java/.../websocket/DanmakuWebSocketHandler.java`
- `mybilibili-danmaku/src/main/java/.../websocket/DanmakuBroadcastService.java`

### 用户画像 + 推荐
- `mybilibili-interaction/src/main/java/.../entity/UserProfile.java`
- `mybilibili-interaction/src/main/java/.../service/impl/UserProfileServiceImpl.java`
- `mybilibili-interaction/src/main/java/.../controller/UserProfileController.java`
- `mybilibili-search/src/main/java/.../service/impl/VideoRecommendServiceImpl.java`
- `mybilibili-search/src/main/java/.../entity/RecommendConfig.java`
- `mybilibili-search/src/main/java/.../service/RecommendConfigService.java`
- `mybilibili-search/src/main/java/.../controller/RecommendConfigController.java`
- `mybilibili-search/src/main/java/.../feign/UserProfileClient.java`

### 安全体系
- `mybilibili-common/src/main/java/.../utils/JwtUtils.java`
- `mybilibili-common/src/main/java/.../config/JwtRequestPolicy.java`
- `mybilibili-common/src/main/java/.../exception/GlobalExceptionHandler.java`
- `mybilibili-gateway/src/main/java/.../security/GatewaySecurityConfig.java`
- `mybilibili-gateway/src/main/java/.../security/GatewayAuthorizationManager.java`
- `mybilibili-gateway/src/main/java/.../filter/GatewayRequestPolicy.java`
- `mybilibili-gateway/src/main/java/.../config/SentinelRuleConfig.java`

### 前端 Composable
- `mybilibili-web/src/composables/useNotificationWs.js`
- `mybilibili-web/src/composables/useDanmakuWs.js`
- `mybilibili-web/src/utils/reconnectingWs.js`
- `mybilibili-web/src/composables/useChunkedManuscriptUpload.js`

### 前端页面
- `mybilibili-web/src/layouts/components/AppHeader.vue` (通知 WS + 搜索联想词)
- `mybilibili-web/src/views/HomeView.vue` (个性化推荐)
- `mybilibili-web/src/views/VideoView.vue` (弹幕实时推送)
- `mybilibili-web/src/views/RegisterView.vue` (兴趣标签冷启动)
- `mybilibili-web/src/components/profile/InterestsPanel展示)
- `mybilibili-web/src/api/index.js` (JWT 无感刷新拦截器)
- `mybilibili-admin-web/src/views/RecommendConfigView.vue` (推荐配置管理)

### 数据库
- `init/mybilibili-mysql.sql` (种子数据)
- `init/V6__meeting_reserve.sql` ~ `init/V9__admin_permission_hardening.sql` (迁移)
- `init/V13__creator_analytics_events.sql` (创作者趋势/稿件状态流水/视频处理流水)
- `init/V17__drop_video_pipeline_tasks.sql` (清理废弃的视频 DB pipeline 表)
- MongoDB: `user_profiles` / `recommend_config` / `danmakus` collections

### Docker
- `scripts/docker-compose-infra.yml` (基础设施容器组)
- `docs/INFRA_DOCKER.md` (Docker 使用说明)

---

## 八、数据库存储分布

| 数据库 | 存储内容 |
|--------|----------|
| MySQL | 用户/稿件/视频/评论/消息/角色/权限/分类/直播/会议 |
| Redis | 上传会话/播放量去重/观看历史/验证码/热搜榜 |
| MongoDB | 弹幕/用户画像/推荐配置 |
| Elasticsearch | 稿件文档(全文搜索+推荐) |
| MinIO | 视频文件/HLS切片/封面/头像/轮播图 |
