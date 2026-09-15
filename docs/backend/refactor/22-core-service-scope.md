# Core 胖单体边界与内嵌前端决策

## 1. Core 胖单体边界

Core 胖单体 = 轻量 CRUD + SSE 推送，**不包含任何长连接 WS 和重组件**。

| 模块 | 包含功能 | 来源 | 说明 |
|------|---------|------|------|
| **user** | 注册/登录/资料、关注/粉丝、验证码、隐私设置、角色权限、审计日志 | account-social | 纯 CRUD |
| **video** | 稿件 CRUD、分类、横幅、字幕映射、播放统计、播放地址 | video-media | 纯 CRUD，不含转码/剪辑 |
| **comment** | 评论 CRUD、举报、敏感词、内容审核、安全设置 | content-interaction | 纯 CRUD |
| **danmaku** | 弹幕收发（发→HTTP POST，收→SSE） | content-interaction | 发=写操作，收=SSE 单向推送 |
| **interaction** | 点赞、收藏、动态、用户画像、观看历史 | content-interaction | 纯 CRUD |
| **message** | 消息通知（SSE） | content-interaction | 纯 SSE 单向推送，无 WS |

**不包含**（独立服务）：

| 服务 | 原因 | 状态 |
|------|------|------|
| **live/** | 语聊房+连麦，WS 长连接 + 麦位状态机 | 独立，本期做 |
| **media/** | 转码/字幕/剪辑，FFmpeg 重组件 | 独立，粒度后定 |
| **payment/** | 支付/订单/B币，强一致 PG 事务 | 后续再加 |

## 2. 二进制内嵌前端决策

### 方案：内嵌，但留开关

```
static:
  mode: embed          # embed | external
  external_url: ""     # external 模式时前端 CDN 地址
```

- **默认 embed**：`go:embed web/dist/*` 将前端构建产物（Vue 3 + TailwindCSS v4）内嵌进二进制
- **支持 external**：配置切到 `mode: external` 时，Go 不提供静态文件服务，由 Traefik/CDN 分发

### 为什么内嵌

| 对比 | 内嵌 | 不内嵌 |
|------|------|--------|
| 部署 | 一个二进制搞定 | 二进制 + 静态文件/Nginx |
| 前端体积 | ~10KB（优化后） | 同上 |
| 更新前端 | 需重新编译 | 独立部署不重启 |
| 盒子适用 | ✅ 极简 | ⚠️ 多一个组件 |

前端目标体积 10KB（Vue 3 + TailwindCSS v4 关键 CSS 内联 + SVG 内联 + Service Worker 缓存），内嵌成本几乎为零，盒子部署单文件优势巨大。

### 内嵌不影响前端开发

- 开发时：`vite dev` 独立运行，热更新正常
- 构建时：`vite build` 输出到 `web/dist/`，Go 内嵌
- CI 时：`go generate` 或 Makefile 自动构建前端再编译 Go

## 3. Core 目录结构

```
cmd/core/main.go
internal/
├── user/        handler + service + repo
├── video/       handler + service + repo
├── comment/     handler + service + repo
├── danmaku/     handler + service + repo（SSE handler）
├── interaction/ handler + service + repo
├── message/     handler + service + repo（SSE handler）
├── common/      公共：错误码、中间件、工具
└── abstraction/ 抽象层接口（配置/日志/缓存/存储/MQ/服务发现）
web/             前端源码（Vue 3 + TS + TailwindCSS v4）
├── dist/        go:embed 目标
├── src/
├── package.json
└── vite.config.ts
sql/             PG 初始化 SQL
proto/           gRPC proto 定义
```