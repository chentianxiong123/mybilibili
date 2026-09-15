# 前端架构与技术选型方案

## 1. 现状分析

### 1.1 现有前端项目

| 项目 | 框架 | TypeScript | 包管理 | 健康度 |
|------|------|-----------|--------|--------|
| mybilibili-web（PC主站） | Vue 3.4 + JS | ❌ 纯 JS | pnpm | 🟡 中 |
| mybilibili-wap（移动端） | Vue 3.4 + 部分 TS | ⚠️ 半 TS | pnpm | 🟡 中 |
| mybilibili-admin-web（后台） | Vue 3.3 + JS | ❌ 纯 JS | pnpm | 🟠 中下 |
| mybilibili-studio-web（创作者） | React + TS | ✅ 全 TS | pnpm | 🟢 良好 |
| mybilibili-live-desktop（直播） | Vue 2 + React 17 混用 | ✅ 有 TS | yarn | 🔴 已废弃 |

### 1.2 核心问题

1. **技术栈不统一**：Vue 2/3 + React 17 混用，维护成本高
2. **无 TypeScript**：3 个 Vue 项目全是 JS，157 个文件无类型检查
3. **版本混乱**：Vue 3.3/3.4，Vite 4.4/4.5，Element Plus 2.3/2.7
4. **前端体积过大**：Element Plus ~300KB，加载慢
5. **包管理器不统一**：pnpm + yarn 混用
6. **直播桌面端已废弃**：Vue 2 + React 17 混用，Vue 2 已 EOL

---

## 2. 技术选型

### 2.1 框架：Vue 3 + TypeScript

| 维度 | 选型 | 原因 |
|------|------|------|
| **框架** | Vue 3.5+ | 已有代码 Base，迁移成本最低 |
| **语言** | TypeScript strict | 类型安全，AI 时代代码质量保障 |
| **构建** | Vite 6 | 零配置，极快 HMR |
| **包管理** | pnpm 统一 | 节省磁盘，速度快 |
| **样式** | TailwindCSS v4 | 轻量（~10KB），零运行时 |
| **状态管理** | Pinia | Vue 官方，轻量 |
| **路由** | Vue Router 4 | Vue 官方 |

### 2.2 为什么不用 React

| 对比 | Vue 3 | React 19 |
|------|-------|---------|
| 已有代码 | ✅ 3 个项目，可直接迁移 | ❌ 全部重写 |
| 国内生态 | ✅ 中文资料丰富，Element Plus 成熟 | ⚠️ 英文为主 |
| 学习曲线 | ✅ 低，模板直观 | ⚠️ 中，JSX 灵活但复杂 |
| Bundle 体积 | ✅ Vapor Mode 后更小 | ⚠️ ~42KB 运行时 |
| 性能 | ✅ 270ms（创建 1 万行） | ⚠️ 829ms（3 倍慢） |

### 2.3 样式方案：TailwindCSS v4 替代 Element Plus

| 方案 | 体积 | 说明 |
|------|------|------|
| Element Plus（旧） | **~300KB** | 全量引入，只用 20% 功能 |
| TailwindCSS v4 | **~10KB** | 按需生成，只编译用到的样式 |
| 手写 CSS | ~5KB | 但开发效率低 |

**TailwindCSS 本质：** 不是"手写组件"，而是**工具类 CSS**——用预置的 class 拼样式，编译后只保留你用到的，最终体积 ~10KB。

```html
<!-- 旧：Element Plus -->
<el-button type="primary" @click="submit">提交</el-button>

<!-- 新：TailwindCSS -->
<button class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg" @click="submit">提交</button>
```

### 2.4 图标方案

| 方案 | 体积 | 说明 |
|------|------|------|
| Element Plus 图标库 | ~100KB | 随 UI 库引入 |
| **SVG 内联** | **~5KB（50个图标）** | 直接把 SVG 代码写进 HTML，零请求 |
| Iconify CDN | 按需加载 | 免费 CDN，15 万+ 图标 |

**推荐：SVG 内联为主，Iconify 兜底。**

```html
<!-- SVG 内联，零外部请求 -->
<svg viewBox="0 0 24 24" width="20" height="20">
  <path d="..." />
</svg>

<!-- 或 Iconify 公共 CDN -->
<svg width="20" height="20">
  <use href="https://api.iconify.design/mdi/home.svg"></use>
</svg>
```

---

## 3. 性能优化方案

### 3.1 首屏优化：减少 HTTP 请求

浏览器加载 HTML 和 CSS 是两次独立的 HTTP 请求，每次都有 **DNS 查询 + TCP 握手 + TLS 握手** 的固定开销（弱网下 1-3 秒）。

**方案：关键 CSS 内联（Critical CSS Inlining）**

| 方案 | 请求数 | 首屏速度 |
|------|--------|---------|
| 外部 CSS 文件 | 2 次请求（HTML + CSS） | 慢（多等一次 HTTP 往返） |
| **关键 CSS 内联** | **1 次请求（HTML 含 CSS）** | **快（零额外等待）** |

```html
<!-- 关键 CSS 内联进 HTML，零额外请求 -->
<style>
  /* 首屏需要的样式，~3KB */
  header { ... }
  .nav { ... }
</style>

<!-- 非关键 CSS 异步加载 -->
<link rel="preload" href="non-critical.css" as="style" onload="this.rel='stylesheet'">
```

**注意：** 内联不是"低级技术"，而是 Google Lighthouse 推荐的官方优化手段（Critical CSS Inlining），淘宝、Amazon、Next.js 都在用。

### 3.2 体积对比

| 方案 | 首屏 JS+CSS | 请求数 | 加载时间（4G） |
|------|------------|--------|--------------|
| Element Plus 全量 | ~400KB | 5-8 次 | 2-3 秒 |
| **TailwindCSS + 内联** | **~50KB** | **1 次** | **0.3-0.5 秒** |

### 3.3 缓存策略：Service Worker

```javascript
// 注册 Service Worker，资源永久缓存
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open('v1').then(cache => {
      return cache.addAll(['/icons.svg', '/app.css', '/app.js'])
    })
  )
})
self.addEventListener('fetch', event => {
  event.respondWith(caches.match(event.request))
})
```

**效果：** 第一次加载后，所有资源永久存在手机里，下次打开零请求，离线可用。

### 3.4 其他优化

| 优化项 | 做法 |
|--------|------|
| 图片 | WebP 格式，`loading="lazy"` 懒加载 |
| 非首屏组件 | 动态 import 异步加载 |
| 字体 | 用系统字体，不额外下载字体文件 |
| 预加载 | `<link rel="preload">` 关键资源 |
| 骨架屏 | 几行 CSS 占位，提升感知速度 |

---

## 4. 项目结构规划

### 4.1 合并方案

将 3 个 Vue 项目合并为 1 个 Monorepo：

```
mybilibili-frontend/
├── packages/
│   ├── web/          ← PC 主站（原 mybilibili-web）
│   ├── wap/          ← 移动端（原 mybilibili-wap）
│   └── admin/        ← 后台（原 mybilibili-admin-web）
├── packages/ui/      ← 共享组件库（TailwindCSS 组件）
├── packages/api/     ← 共享 API 层（对接后端 gRPC/REST）
├── packages/utils/   ← 共享工具函数
├── pnpm-workspace.yaml
├── vite.config.ts
└── package.json
```

### 4.2 组件库策略

不使用 Element Plus，改用 **TailwindCSS + 自建轻量组件库**：

```
packages/ui/
├── components/
│   ├── Button.vue      ← 按钮
│   ├── Input.vue       ← 输入框
│   ├── Dialog.vue      ← 弹窗
│   ├── Table.vue       ← 表格
│   ├── Pagination.vue  ← 分页
│   └── ...
├── icons/               ← SVG 图标（内联）
├── styles/              ← 全局样式（TailwindCSS）
└── index.ts
```

**只实现你页面里用到的组件，不改别人的，不加载不用的代码。**

### 4.3 studio-web 保留

**mybilibili-studio-web**（创作者工作室，React + TS Monorepo）保持现状，不影响主站。

---

## 5. 部署方案

### 5.1 CDN 策略

| 资源 | 部署位置 | 说明 |
|------|---------|------|
| **HTML** | 你的服务器 | 首屏，单次请求 |
| **CSS/JS/图标** | 内联进 HTML | 零额外请求 |
| **非关键 CSS/JS** | 七牛云 CDN 免费额度 | 每月 10GB 免费，100 人够用 |
| **图片/视频** | MinIO（主力机） | 内网自建，不费流量 |

### 5.2 缓存策略

```
静态资源（CSS/JS/图标）：
  Cache-Control: public, max-age=31536000, immutable
  → 浏览器永久缓存，下次不请求

HTML：
  Cache-Control: no-cache
  → 每次检查更新，但内容极小（~50KB）
```

---

## 6. 面试话术

> **"前端采用 Vue 3 + TypeScript + TailwindCSS 技术栈，通过关键 CSS 内联、SVG 图标内联、Service Worker 缓存等优化手段，将首屏体积从 400KB 压缩到 50KB，加载时间从 2-3 秒降到 0.3-0.5 秒，弱网环境下提升显著。不使用 Element Plus 等重型 UI 库，自建轻量组件库，只编译用到的样式，产物体积减少 90%。"**

---

## 7. 总结

| 决策 | 选型 | 原因 |
|------|------|------|
| 框架 | **Vue 3 + TypeScript** | 已有代码 Base，迁移成本最低 |
| 样式 | **TailwindCSS v4** | 替代 Element Plus，体积从 300KB→10KB |
| 图标 | **SVG 内联** | 零请求，零带宽 |
| 首屏 | **关键 CSS 内联** | 减少一次 HTTP 请求，快 0.5-1 秒 |
| 缓存 | **Service Worker** | 一次加载，永久离线可用 |
| 部署 | **七牛云 CDN（免费额度）** | 国内快，10GB/月够用 |
| 合并 | **Monorepo（3 合 1）** | 共享组件，统一技术栈 |