# mybilibili-app-flutter 设计文档

## 1. 项目概述

基于 Flutter 的 B站仿制客户端，对接自定义 Go 后端（`mybilibili-go`，HTTP `/api/v1/` 120+ 端点），**完全像素级别复刻** WAP 端 24 个页面功能。复用 PiliPlus 的 `pl_player` 播放器组件（media_kit + canvas_danmaku），实现 B站风格播放体验。

**目标平台**：Android、iOS（首发）→ Windows、macOS、Linux（后续桌面端）

## 2. 技术选型

| 层 | 方案 | 理由 |
|----|------|------|
| 状态管理 | Riverpod | 编译安全、无 context 依赖、测试友好 |
| 路由 | GoRouter + ShellRoute | 底部导航嵌套、深链接 |
| HTTP | Dio + 拦截器 | 自动 token 刷新、SSE 支持 |
| 视频播放 | **media_kit** (libmpv) | 全平台 + HLS + 硬解 |
| 播放器 UI | **pl_player**（从 PiliPlus 提取） | 完整仿 B站播放器（弹幕/手势/画质/三连） |
| 弹幕 | canvas_danmaku | CustomPainter 高性能渲染 |
| 投屏 | dlna_dart | DLNA 投屏 |
| 存储 | flutter_secure_storage + hive | Token 安全存储 + 本地缓存 |
| 状态恢复 | shared_preferences | 播放进度/设置 |

## 3. 目录结构

```
mybilibili-app-flutter/
├── lib/
│   ├── main.dart
│   ├── app.dart
│   ├── core/
│   │   ├── api/          # Dio 客户端 + 各模块 API 封装
│   │   ├── router/       # GoRouter 配置
│   │   ├── theme/        # 主题（B站粉色系）
│   │   └── utils/        # token存储/SSE客户端/格式化
│   ├── features/
│   │   ├── home/         # 首页推荐
│   │   ├── video/        # 视频详情 + 播放
│   │   ├── search/       # 搜索/热搜/结果
│   │   ├── auth/         # 登录/注册
│   │   ├── user/         # 用户主页
│   │   ├── profile/      # 个人中心（历史/收藏/稿件/编辑）
│   │   ├── creator/      # 创作中心
│   │   ├── dynamic/      # 关注动态
│   │   ├── message/      # 消息/私信
│   │   ├── live/         # 直播
│   │   └── ranking/      # 排行榜
│   ├── plugin/
│   │   └── pl_player/    # ★ 从 PiliPlus 提取的播放器组件
│   ├── shared/
│   │   ├── models/       # 数据模型（对齐 Go API）
│   │   └── widgets/      # 通用组件
│   └── providers/        # Riverpod providers
├── pubspec.yaml
└── ...
```

## 4. 播放器集成方案（核心）

**来源**：PiliPlus `lib/plugin/pl_player/`（GPL-3.0）

**提取内容**：
- `controller.dart`（PlPlayerController）
- `view/view.dart`（PLVideoPlayer 视图）
- `widgets/`、`models/`、`utils/`

**改造点**：
1. 移除 B站 API 依赖（`bilibili-API-collect` 相关）
2. 数据源改为 Go 后端：视频 URL（多清晰度 HLS/MP4）、弹幕列表、字幕
3. 保留：弹幕渲染/手势/画质切换/倍速/全屏/三连动画
4. 弹幕发送 → `POST /danmaku/send`，实时弹幕 → SSE

**License 注意**：pl_player 为 GPL-3.0，复用需保持项目同样 GPL-3.0 许可。

## 5. API 对接清单

对齐 Go 后端 120+ 端点，按 WAP 功能划分：

- **认证**：登录/注册/token刷新 → `POST /user/login` 等
- **内容**：推荐/热门/分类/排行 → `GET /manuscript/recommended` 等
- **播放**：稿件详情/分P视频 → `GET /manuscript/{id}`、`GET /video/{id}`
- **互动**：点赞/投币/收藏/三连 → `/manuscript/{id}/like` 等
- **弹幕**：`GET /danmaku/video/{id}`、`POST /danmaku/send`、`SSE /sse/danmaku`
- **社交**：动态/关注/私信 → `/dynamic/*`、`/follow/*`、`/message/*`
- **用户**：主页/历史/收藏/稿件 → `/user/*`、`/watch-history/*`、`/manuscript/user/*`
- **直播**：房间列表/直播间 → `/live/room/*`、`/live/linkmic/*`
- **搜索**：`/search/videos`、`/search/suggest`、`/search/hot`

## 6. 开发计划（6 个阶段）

| 阶段 | 内容 | 交付物 |
|------|------|--------|
| **S1 骨架** | flutter 脚手架 + pl_player 提取 + Dio/Riverpod/路由 + 主题 | 可运行空壳 + 播放器组件编译通过 |
| **S2 首页+播放** | 首页推荐、视频详情+弹幕播放器、分类 | 核心播放链路打通 |
| **S3 搜索+用户** | 搜索/热搜/结果、用户主页、个人中心（历史/收藏/稿件/编辑） | 浏览+个人功能 |
| **S4 社交** | 动态、消息/私信、关注/粉丝 | 社交链路 |
| **S5 直播** | 直播列表、直播间（含弹幕）、直播分类 | 直播链路 |
| **S6 创作中心** | 投稿管理、数据面板、评论管理 | WAP 全功能对等 |

## 7. 风险与对策

| 风险 | 对策 |
|------|------|
| pl_player 与 PiliPlus 深度耦合 | 提取时先跑通编译，逐步剥离 B站 API |
| GPL-3.0 许可 | 项目声明 GPL-3.0，或仅参考不复制 |
| 后端 SSE 弹幕格式 | 先读 Go 源码确认 SSE 数据结构 |
| media_kit 桌面端 libmpv 依赖 | 桌面端后续阶段，Linux 需系统安装 libmpv |