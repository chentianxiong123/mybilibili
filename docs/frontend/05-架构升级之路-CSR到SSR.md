# 架构升级之路：从 CSR 到 SSR

## 为什么需要 SSR

### 当前瓶颈

CSR（客户端渲染）的典型问题：

```
用户访问页面：
1. 下载空壳 HTML（~2KB）
2. 下载 JS（170K gzip）
3. 解析 + 编译 JS（~200ms）
4. 执行 JS，渲染页面（~300ms）
5. 用户看到内容（总计 ~2-3s）
```

### SSR 的解决方式

```
用户访问页面：
1. 服务器调用 Go API 获取数据
2. 服务器渲染完整 HTML（含标题、描述、内容）
3. 返回给浏览器（~200ms）
4. 用户直接看到内容
5. 后台下载 JS → 激活交互（Hydration）
```

## 技术选型

### 为什么选 Nuxt 4

| 特性 | Nuxt 4 | 自研 SSR | 其他框架 |
|------|--------|---------|---------|
| 社区生态 | 最大 | 无 | 中等 |
| 文件路由 | ✅ 内置 | ❌ 手写 | ✅ |
| 自动布局 | ✅ 内置 | ❌ 手写 | ✅ |
| SSR 数据水合 | ✅ useFetch | ❌ 手动 | ✅ |
| 模块系统 | 丰富 | 无 | 有限 |
| 大厂案例 | 知乎、京东 | B 站 | 较少 |

### 为什么用 Bun 二进制部署

Bun 可以将 Nuxt 应用编译为单文件二进制，运行时无需 Node.js 环境。

```bash
# 安装 nuxt-bun-compile 模块后，一行命令生成二进制
bun run -b build
# 输出 nuxtbin 文件，直接 ./nuxtbin 运行
```

```dockerfile
# Dockerfile：多阶段构建
FROM oven/bun:alpine AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN NODE_OPTIONS="--max-old-space-size=8192" bun run -b build

FROM alpine:latest AS release
RUN apk add --no-cache libstdc++ libgcc
WORKDIR /app
COPY --from=build /app/nuxtbin /app/nuxtbin
EXPOSE 3000
CMD ["./nuxtbin"]
```

| 对比 | Node.js 容器 | Bun 二进制 |
|------|-------------|-----------|
| 镜像大小 | 100MB | 64MB |
| 启动时间 | 2-3s | 0.5-1s |
| 部署方式 | 多文件 | 单文件 |

## 架构变化

### 当前架构

```
┌─────────────────────────────────────────┐
│ 浏览器                                   │
│  ├─ 下载空壳 HTML（Nginx 静态文件）       │
│  ├─ 下载 JS 170K gzip                    │
│  ├─ 客户端渲染                           │
│  └─ 调用 Go API 获取数据                  │
│                                          │
│ Go 后端（一分不改）                        │
└─────────────────────────────────────────┘
```

### 目标架构

```
┌─────────────────────────────────────────┐
│ 浏览器                                   │
│  ├─ 请求 Nuxt SSR 服务器                  │
│  │   ├─ Nuxt 调用 Go API 获取数据         │
│  │   └─ 返回完整 HTML（~30K gzip JS）     │
│  ├─ 直接显示内容                           │
│  └─ 后续交互直连 Go API                    │
│                                          │
│ Go 后端（一分不改）                        │
│ Nginx → Nuxt → Go 的路由链不变             │
└─────────────────────────────────────────┘
```

## 部署架构

### docker-compose 配置

```yaml
services:
  traefik:
    image: traefik:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./traefik:/etc/traefik

  nuxt:
    build: ./mybilibili-nuxt
    expose:
      - "3000"
    environment:
      - NUXT_API_BASE=http://go-api:8080/api/v1

  go-api:
    build: ./mybilibili-go
    expose:
      - "8080"
```

### Traefik 路由规则

```
/api/v1/*       → Go 后端（不变）
/                → Nuxt SSR 服务器（原来指向静态文件，现在指向 Node.js 进程）
/_nuxt/*        → 静态文件（Nuxt 构建产物，走 CDN）
```

## 折中方案：SSG 混合渲染

如果不想跑 Node.js 服务，Nuxt 支持 SSG（静态生成）模式，构建时预渲染页面为 HTML，部署时 Nginx 直接服务，不需要 Node.js 进程。

```typescript
// nuxt.config.ts
routeRules: {
  '/'          → prerender（首页，构建时生成）
  '/category/*'→ prerender（分类页，构建时生成）
  '/manuscript/*' → ssr: false（视频页，CSR）
  '/admin/**'  → ssr: false（管理后台，CSR）
  '/live/**'   → ssr: false（直播，CSR）
}
```

**代价：** 非预渲染页面仍是 CSR，首屏 JS 比完整 SSR 多，但比当前 Vite 项目小（Nuxt 内建优化更强）。

## 分阶段迁移计划

### Phase 1：基础搭建（1-2 天）

1. 创建 Nuxt 4 项目
2. 安装 `nuxt-bun-compile` 模块
3. 配置 API 代理到 Go 后端
4. 迁移 3 个布局文件到 `layouts/` 目录
5. 迁移首页，API 改用 `useFetch`

### Phase 2：核心页面（3-5 天）

1. 视频播放页（SEO 核心页面）
2. 搜索页、分类页
3. 用户页、个人主页
4. 路由守卫 → `middleware/`
5. 认证 → `useCookie` 替代 `localStorage`

### Phase 3：复杂页面（3-5 天）

1. 直播页面 → `<ClientOnly>` 包裹
2. WebSocket composable → `onMounted` 保护
3. 创作中心、消息中心
4. 管理后台（保持独立或迁入）

### Phase 4：UI 库优化（可选，2-3 天）

1. 评估 `@element-plus/nuxt` 或 `@nuxt/ui`
2. 替换 Element Plus（如选 `@nuxt/ui`）
3. 全局主题统一

## 迁移后效果预估

| 指标 | 当前（CSR） | 迁移后（SSR） | 改善 |
|------|-----------|-------------|------|
| 首屏 JS | 170K gzip | ~30K gzip | -82% |
| LCP | ~2-3s | ~0.5-1s | -60% |
| 部署镜像 | 20MB（Nginx） | 84MB（Nginx + Nuxt） | 多 64MB |
| Go 后端 | 不改 | 不改 | 0 |
| 前端代码 | 不动 | 逐步迁移 | 有序 |

## 风险与应对

| 风险 | 概率 | 影响 | 应对 |
|------|------|------|------|
| Element Plus SSR 不兼容 | 中 | 高 | `<ClientOnly>` 包裹，或换库 |
| 浏览器 API 在服务端报错 | 高 | 中 | `onMounted` 保护 |
| 认证迁移问题 | 中 | 高 | `useCookie` + 兼容旧 localStorage |
| 构建时间增加 | 低 | 低 | 利用 Nuxt 构建缓存 |
| 迁移期间双项目维护 | 中 | 中 | 分阶段，每页独立验证 |

## 总结

SSR 迁移是前端性能优化的终极手段，但也是架构级变更。**是否执行必须以实测数据为前提，而不是默认选择。**

## 其他讨论过的方案

### 方案 A：Go 模板渲染 SEO（已否决）

```
Nginx → Go 后端（Go 模板渲染标题/描述/封面）
       → Vue CSR（交互部分）
```

- 不需要 Node.js，不改部署
- 只解决 SEO，不解决首屏速度
- 用户仍需等 JS 加载完才能交互

### 方案 B：继续压 CSR（采用，见最终决定）

```
Nginx → 静态文件
```

- 不需要额外进程
- 首屏优化空间有限（170K → 目标 80-100K）
- 无法达到 SSR 的 30K 级别，但对玩具项目已够用

### 方案 C：Nuxt 4 SSR + Bun 二进制（备选，等数据触发）

```
Nginx → Nuxt（Bun 二进制，64MB）→ Go 后端
```

- 首屏 170K → 30K，SSR/SSG/ISR/CSR 混合渲染
- 前提条件：实测 LCP 不达标，且性能影响业务

## 最终决定（2026-08-17 更新）

**暂不采用 SSR。** 依据企业级方法论（详见 06 文档）：

1. 玩具项目 + 无 SEO 需求 + 无 CDN = SSR 收益递减区
2. 权威指南：月访客 < 1 万、不依赖搜索排名 → 达到 CWV 阈值即停止
3. 正确路径：先接 RUM 看真实 LCP/INP/CLS 数据 → 达标则继续 CSR 优化路线（图片、缓存头、关键 CSS、虚拟滚动）→ 全做完了还慢且影响业务才触发 SSR

**已完成的 CSR 优化**（170K gzip 首屏）已是比多数生产项目更强的基线。SSR 作为"最后手段"，由实测数据触发，不主动执行。

## 参考

- [Nuxt 4 文档](https://nuxt.com/docs)
- [Bun 文档 - 单文件可执行文件](https://bun.com/docs/bundler/executables)
- [nuxt-bun-compile](https://github.com/jprando/nuxt-bun-compile)
- [Element Plus 升级变化](https://github.com/element-plus/element-plus/issues/15834)
- [Vue 官方性能指南](https://cn.vuejs.org/guide/best-practices/performance)
- [web.dev 性能预算](https://web.dev/articles/performance-budgets-101?hl=zh-cn)
- [企业级性能优化方法论](06-企业级性能优化方法论.md)