# MyBilibili Cloud

仿B站（哔哩哔哩）微服务架构视频平台，基于 Spring Cloud + Vue 3 构建，涵盖视频投稿、弹幕、评论、私信、动态、搜索、直播、视频会议、连麦、AI 字幕/摘要、AI 客服、AI 内容审核、API 管理等核心功能，并提供移动端（WAP）适配。

## 技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| Spring Boot | 3.2.4 | 基础框架 |
| Spring Cloud | 2023.0.1 | 微服务框架 |
| Spring Cloud Alibaba | 2023.0.1.0 | Nacos 注册/配置中心 |
| Spring Cloud Gateway | - | API 网关 |
| MyBatis Plus | 3.5.3 | ORM 框架 |
| MySQL | 8.0 | 关系型数据库 |
| RocketMQ | 4.9.4 | 消息队列 |
| Elasticsearch | - | 全文搜索 |
| MongoDB | - | 字幕文档存储 |
| SRS | 5.x | 直播推流服务器 (RTMP/HLS/HTTP-FLV) |
| WebRTC | - | 视频会议/连麦 |
| FFmpeg | - | 视频转码 (HLS) |
| Whisper | - | AI 语音识别/字幕生成 |
| DeepSeek | - | AI 视频摘要 / AI 客服 / AI 内容审核 |
| Java | 17 | 开发语言 |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.4 | 前端框架 |
| Vite | 4.5 | 构建工具 |
| Element Plus | 2.7 | UI 组件库 |
| Pinia | 2.1 | 状态管理 |
| Vue Router | 4.2 | 路由管理 |
| Artplayer | 5.1 | 视频播放器 |
| HLS.js | 1.5 | HLS 流媒体播放 |
| ECharts | 5.6 | 数据可视化 |

## 微服务架构

```
                        ┌─────────────┐
                        │    Nginx    │ :80
                        └──────┬──────┘
                               │ /api/*
                        ┌──────┴──────┐
                        │   Gateway   │ :8080
                        └──────┬──────┘
                               │
        ┌──────────┬──────────┼──────────┬──────────┬──────────┐
        │          │          │          │          │          │
   ┌────┴────┐ ┌───┴───┐ ┌───┴───┐ ┌───┴───┐ ┌───┴────┐ ┌───┴────┐
   │  User   │ │ Video │ │Comment│ │Danmaku│ │Search  │ │  Live  │
   │ Service │ │Service│ │Service│ │Service│ │Service │ │Service │
   └─────────┘ └───────┘ └───────┘ └───────┘ └────────┘ └────────┘
        │          │          │
   ┌────┴────┐ ┌───┴───┐ ┌───┴──────┐
   │Message  │ │Inter- │ │   AI     │
   │ Service │ │action │ │ Service  │
   └─────────┘ └───────┘ └──────────┘
```

### 服务清单

| 服务 | 端口 | 说明 |
|------|------|------|
| mybilibili-gateway | 8080 | API 网关，路由转发、JWT 鉴权 |
| mybilibili-account-social | 8081 | 用户注册/登录、个人信息、关注/粉丝、私信、系统消息、管理员/权限 |
| mybilibili-video-media | 8082 | 稿件管理、视频上传、分类、轮播图、字幕、直播推流、视频会议、观众连麦 |
| mybilibili-content-interaction | 8085 | 评论/回复、动态评论、弹幕、点赞/投币/收藏/分享、动态、收藏夹、合集、观看历史、内容审核 |
| mybilibili-search-recommend | 8084 | Elasticsearch 全文搜索、索引管理、推荐、用户画像读取、创作者数据统计 |
| mybilibili-ai | 8088 | 视频转码(HLS)、AI 字幕生成、AI 摘要、SSE 进度推送、AI 客服对话、AI 内容审核、API 配置管理 |
| mybilibili-common | - | 公共模块（JWT、VO、工具类） |
| mybilibili-mq | - | 消息队列公共模块 |

旧的 user/video/comment/danmaku/interaction/message/search/live/analytics 模块暂时保留为聚合服务的源码来源和回退参考，不作为默认运行时服务启动。

### 前端应用

| 应用 | 说明 |
|------|------|
| mybilibili-web | 用户端前端（视频播放、投稿、动态、私信等） |
| mybilibili-admin-web | 管理后台前端（稿件审核、用户管理、数据统计等） |
| mybilibili-wap | 移动端前端（Vue 3 移动适配，首页/视频/直播/搜索/空间等） |

## 数据库设计

数据库名：`mybilibili`，核心表结构如下：

| 表名 | 说明 |
|------|------|
| users | 用户表 |
| manuscripts | 稿件表 |
| videos | 视频表（分P、转码状态、字幕/摘要标记） |
| comments | 评论表 |
| replies | 回复表 |
| user_interactions | 用户交互记录（点赞/投币/收藏/关注） |
| user_dynamics | 用户动态表 |
| dynamic_comments | 动态评论表 |
| favorite_folders | 收藏夹表 |
| favorite_manuscripts | 收藏稿件关联表 |
| manuscript_collections | 稿件合集表 |
| manuscript_collection_relations | 合集-稿件关联表 |
| messages | 私信消息表 |
| conversations | 会话表 |
| message_settings | 消息通知设置表 |
| categories | 分区表 |
| tags | 标签表 |
| video_tags | 视频-标签关联表 |
| banner_images | 轮播图表 |
| admin_users | 管理员表 |
| roles | 角色表 |
| permissions | 权限表 |
| prohibited_words | 违禁词表 |
| creator_settings | 创作者设置表 |
| live_rooms | 直播间表（房间名、推流密钥、状态、封面、观看人数） |
| live_linkmic | 连麦记录表（主播、观众、状态、音视频开关） |
| meeting_room | 会议房间表（房间名、邀请码、状态） |
| meeting_participant | 会议参与者表（角色、音视频/屏幕共享状态） |
| ai_conversations | AI 客服对话表（用户、标题、状态） |
| ai_chat_messages | AI 客服消息表（角色、内容、Token 数） |
| reports | 举报/审核表（AI 审核状态、判定结果、风险等级） |
| prohibited_word | 违禁词详情表（匹配方式、分类、启用状态） |

完整建表 SQL 见 [scripts/mybilibili.sql](scripts/mybilibili.sql) 和 [init/init.sql](init/init.sql)。

## 快速开始

### 环境依赖

- JDK 17+
- MySQL 8.0+
- Nacos 2.x
- RocketMQ 4.9+
- Elasticsearch 7.x
- MongoDB（AI 字幕存储）
- SRS（直播推流服务器）
- Node.js 16+
- FFmpeg（视频转码）
- Nginx（可选，生产部署）

### 1. 启动基础设施

```bash
# 启动 Nacos
# 启动 MySQL，导入 scripts/mybilibili.sql

# Docker 基础设施组
docker compose -f scripts/docker-compose-infra.yml up -d
```

### 2. 启动后端微服务

按以下顺序启动各服务：

1. mybilibili-account-social
2. mybilibili-video-media
3. mybilibili-content-interaction
4. mybilibili-search-recommend
5. mybilibili-ai
6. mybilibili-gateway

或使用一键启动脚本：

```powershell
# 架构边界检查
.\scripts\check-architecture.ps1

# PowerShell
.\scripts\start-all.ps1

# CMD
.\scripts\start-all.bat
```

### 3. 启动前端

```bash
# 用户端
cd mybilibili-web
pnpm install
pnpm run dev

# 管理端
cd mybilibili-admin-web
pnpm install
pnpm run dev

# 移动端
cd mybilibili-wap
pnpm install
pnpm run dev
```

### 4. Nginx 配置（生产部署）

参考 [nginx/conf/nginx.conf](nginx/conf/nginx.conf)，配置前端静态资源和 API 反向代理。

## 核心功能

- **视频系统**：视频上传 → FFmpeg HLS 转码 → 多清晰度播放 → AI 字幕生成 → AI 视频摘要
- **弹幕系统**：实时弹幕发送与展示，基于 Artplayer 弹幕插件
- **评论系统**：多级评论/回复，违禁词自动过滤，内容审核
- **互动系统**：点赞、投币、收藏、分享、关注
- **动态系统**：发布图文动态，引用视频，动态评论
- **私信系统**：一对一私信，会话管理，消息通知
- **搜索系统**：Elasticsearch 全文搜索，索引管理
- **创作中心**：稿件管理、数据统计、收藏夹/合集管理
- **管理后台**：稿件审核、用户管理、轮播图管理、违禁词管理、数据看板
- **AI 功能**：Whisper 语音识别自动字幕、DeepSeek 视频内容摘要、SSE 实时进度推送
- **直播系统**：SRS 推流/拉流、RTMP 推流密钥鉴权、HLS/HTTP-FLV 播放、直播间管理、开播状态回调
- **视频会议**：WebRTC 多人会议、6 位邀请码入会、主持人管理（踢出/静音/聚焦/转让）、屏幕共享、聊天、举手
- **观众连麦**：直播中观众申请连麦、主播审核、WebRTC 音视频通话、音视频开关控制
- **AI 客服**：DeepSeek 驱动的"哔哩助手"、SSE 流式对话、多轮上下文、对话管理
- **AI 内容审核**：DeepSeek 自动审核举报内容（高/中/低风险分级）、RocketMQ 异步处理、违禁词匹配（精确/模糊）
- **API 管理**：运行时 AI 接口配置（Key/URL/模型/参数）、连接测试、Nacos 热更新
- **移动端（WAP）**：Vue 3 移动端适配、首页/频道/排行/视频/直播/搜索/空间/消息等功能

## 项目结构

```
mybilibili-cloud/
├── mybilibili-gateway/          # API 网关
├── mybilibili-account-social/   # 账户与社交聚合服务
├── mybilibili-video-media/      # 视频与媒体聚合服务
├── mybilibili-content-interaction/ # 内容互动聚合服务
├── mybilibili-search-recommend/ # 搜索与推荐聚合服务
├── mybilibili-ai/               # AI 服务
├── legacy-services/             # 旧服务源码池，不作为 Maven 运行模块
│   ├── mybilibili-user/
│   ├── mybilibili-video/
│   ├── mybilibili-comment/
│   ├── mybilibili-danmaku/
│   ├── mybilibili-interaction/
│   ├── mybilibili-message/
│   ├── mybilibili-search/
│   ├── mybilibili-live/
│   └── mybilibili-analytics/
├── mybilibili-common/           # 公共模块
├── mybilibili-mq/               # 消息队列模块
├── mybilibili-web/              # 用户端前端
├── mybilibili-admin-web/        # 管理端前端
├── mybilibili-wap/              # 移动端前端
├── nginx/                       # Nginx 配置
├── init/                        # 初始化 SQL
├── scripts/                     # 数据库脚本
├── pom.xml                      # Maven 父 POM
└── settings.xml                 # Maven 镜像配置
```

父 `pom.xml` 默认只构建 6 个后端运行服务及公共模块；`legacy-services/` 下旧源码由聚合服务通过 `build-helper-maven-plugin` 引入，不再单独打包成运行服务。

## License

MIT
