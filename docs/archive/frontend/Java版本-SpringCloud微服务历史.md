# Java Spring Cloud 版本（历史遗产）

> 当前 Go 版本为主体（见根 README）。在此之前项目经历了两轮 Java 架构实现，沉淀了大量架构演进价值。所有代码均保留在仓库 `old/` 目录，本文档提炼其价值与演进脉络。

## 演进脉络（三阶段）

```
单体版(old/java)
   └─→ Spring Boot 2.7 全量微服务(old/springboot3)
          └─→ Spring Boot 3.2 聚合微服务(old/microservices)  ←  Java 最终形态
                  └─→ Go 微服务(当前, mybilibili-go)
```

| 阶段 | 代码位置 | 后端技术 | Java | 形态 |
|---|---|---|---|---|
| 1. 单体版 | `old/java` | Spring Boot 2.7.18 + MyBatis + Redis + ES 7.17 | 8 | web / admin / common 三模块 |
| 2. 微服务版 | `old/springboot3` | Spring Cloud 2021.0.8 + Alibaba(Nacos) + RocketMQ + ES + MongoDB + SRS + WebRTC | 8 | 11 个微服务（user/video/comment/danmaku/interaction/message/search/live/ai…） |
| 3. 聚合微服务版 | `old/microservices` | Spring Cloud 2023.0.1 + Alibaba(Nacos) + RocketMQ + ES + MongoDB + SRS + WebRTC | 17 | 聚合为 6 核心服务 + 桌面临播端 + OBS 插件 |

## 为什么有价值（能力沉淀）

这套 Java 实现**不是玩具 demo**，而是在功能完整度、架构广度和工程深度上都达到了可演示的 B 站全站水平，大量设计被 Go 版本继承演进：

### 1. 全业务域覆盖
仿 B 站完整业务闭环，Go 版中的绝大多数业务在此版本已具备原始实现：

- 用户注册/登录、JWT 鉴权、关注/粉丝、管理员/RBAC
- 稿件管理、视频上传、分 P、**视频转码（HLS）**、**音频提取**、处理进度推送
- 弹幕、评论/回复（含动态评论）、点赞/投币/收藏/分享、收藏夹、合集、观看历史、动态
- 私信/会话/消息通知设置（`message_settings`）、系统消息
- 搜索（Elasticsearch 全文）+ 索引管理 + 推荐 + 创作者数据统计
- 轮播图、分类、标签、违禁词、内容审核、举报
- **字幕生成、观看历史、创作者设置**

### 2. AI 能力（Whisper + DeepSeek）
- **Whisper** 语音识别 → AI 字幕生成/导入
- **DeepSeek** 视频摘要、AI 客服对话（`ai_conversations`/`ai_chat_messages`，带 Token 统计）、AI 内容审核（`reports` 风险等级/判定结果）
- API/模型配置集中管理（`ai_api_configs`）

### 3. 直播全栈（SRS + WebRTC + 桌面 + 插件）
- 直播间（房间名/推流密钥/状态/封面/观看人数，`live_rooms`）+ SRS 推流（RTMP/HLS/HTTP-FLV）
- **视频会议（WebRTC）**：会议房间/参与者表（`meeting_room`/`meeting_participant`，角色/音视频/**屏幕共享**状态）、观众连麦（`live_linkmic`）、WebSocket 信令
- **mybilibili Live Desktop**（`old/microservices/mybilibili-live-desktop`）：fork 自 Streamlabs Desktop 的开源直播桌面 App（Electron + OBS 核心），OSX/Windows
- **OBS 直播插件**（`old/microservices/obs-plugin`）：C++ 编写的 OBS Studio 插件，扫码登录 B 站、更新直播间信息、获取 RTMP 推流地址与推流码，CI 打包 Windows 预编译 DLL

### 4. 微服务工程化
- Spring Cloud Gateway 统一入口（路由 + JWT 鉴权）
- Nacos 注册/配置中心、RocketMQ 异步解耦、MyBatis Plus ORM
- 聚合层设计：account-social/video-media/content-interaction/search-recommend 合服精简（第三阶段），展示服务治理演进思路
- ES 与 MySQL 双写、字幕存 MongoDB 的异构存储选型

### 5. 前端三端（Vue 3）
- `mybilibili-web` 用户端（ArtPlayer 播放器 + HLS.js + ECharts + Pinia）
- `mybilibili-admin-web` 管理后台（稿件审核/用户/数据统计）
- `mybilibili-wap` 移动端（首页/视频/直播/搜索/空间）

### 6. 数据库设计沉淀
`scripts/mybilibili.sql` / `init/init.sql` 包含完整表结构，Go 版的 89 张表与之高度对应：用户、稿件、`videos`(分 P/转码状态/字幕摘要标记)、`user_interactions`(点赞/投币/收藏/关注)、收藏夹、合集、私信、轮播、RBAC、违禁词、直播间/连麦/会议、AI 对话表……是业务模型的活文档。

## 与 Go 版的关系

| 维度 | Java 版（old） | Go 版（当前） |
|---|---|---|
| 语言/框架 | Java + Spring Cloud | Go + gRPC |
| 数据库 | MySQL 8.0 | PostgreSQL 16（主从） |
| MQ | RocketMQ | NATS JetStream（集群） |
| 搜索 | Elasticsearch | Elasticsearch（search-service 索引管理/重建） |
| 直播 | SRS + WebRTC + 桌面 + OBS 插件 | SRS + WebRTC(WHIP/P2P) 连麦，桌面/插件经验沿用 |
| AI | Whisper 字幕 + DeepSeek 摘要/客服/审核 | AI 字幕/摘要/审核 + AI 客服技能路由 + 用量统计（升级） |
| 部署 | 单体/微服务 jar | k3s 多副本 + 主从冗余（高可用） |
| 前端 | Vue3（web/admin/wap） | Nuxt4 SSR + Vue3 admin + H5 + Flutter/NS/WebView |

> 结论：Java 版证明了业务完整性与功能广度，Go 版在此基础上完成架构现代化（微服务拆分、消息流式化、高可用部署、AI 深化），两代技术沉淀共同构成该项目的完整工程履历。