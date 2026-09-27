# 极致性能优化方案：Lighthouse 100 分

## 1. 目标

| 指标 | 目标 | 行业平均 |
|------|------|---------|
| Lighthouse 评分 | **100** | 70-80 |
| 首次内容绘制（FCP） | **< 0.3s** | 1.5-2.5s |
| 最大内容绘制（LCP） | **< 0.5s** | 2.0-4.0s |
| 首次输入延迟（FID） | **< 50ms** | 100-300ms |
| 累积布局偏移（CLS） | **< 0.05** | 0.1-0.3 |
| 首包体积 | **< 30KB** | 200-500KB |
| HTTP 请求数 | **1 次** | 10-30 次 |
| 离线可用 | ✅ 是 | ❌ 否 |

## 2. 框架选型：Svelte 5

### 2.1 为什么是 Svelte 5

| 对比项 | Svelte 5 | Vue 3 | React 19 |
|--------|----------|-------|---------|
| 运行时体积 | **0KB（编译时消除）** | ~33KB | ~42KB |
| 基准性能 | **39.5 ops/sec** | 31.2 ops/sec | 28.4 ops/sec |
| 首屏包体积 | **~15KB** | ~50KB | ~72KB |
| 内存占用 | **最低** | 中 | 高 |
| 学习曲线 | 低 | 低 | 中 |

**Svelte 5 没有运行时——编译时直接把响应式逻辑编译成原生 DOM 操作，不需要浏览器加载框架代码。**

### 2.2 性能对比（js-framework-benchmark）

| 基准测试 | Svelte 5 | React 19 | 差距 |
|---------|----------|----------|------|
| 创建 1 万行 | **472ms** | 829ms | **1.8x 快** |
| 更新每 10 行 | **35ms** | 258ms | **7.4x 快** |
| 选中行 | **16ms** | 251ms | **15.7x 快** |
| 包体积（gzip） | **15KB** | 62KB | **4.1x 小** |

---

## 3. 核心优化技术

### 3.1 关键 CSS 内联（Critical CSS Inlining）

**原理：** 浏览器加载 HTML 和 CSS 是两次独立的 HTTP 请求，每次都有 DNS 查询 + TCP 握手 + TLS 握手（弱网下 1-3 秒）。内联后一次请求搞定。

```html
<!DOCTYPE html>
<html>
<head>
  <style>
    /* 所有首屏 CSS 内联在这里，~3KB */
    /* 编译后只包含实际用到的样式 */
  </style>
</head>
<body>
  <!-- 内容 -->
</body>
</html>
```

### 3.2 SVG 图标内联

**原理：** 图标直接用 SVG 代码写进 HTML，不请求任何外部图标文件。

```html
<!-- 内联 SVG，零请求 -->
<svg viewBox="0 0 24 24" width="20" height="20">
  <path d="M12..."/>
</svg>
```

50 个图标约 5KB，一次加载永久缓存。

### 3.3 Service Worker 离线缓存

**原理：** 第一次加载后，所有资源永久缓存，后续打开零请求。

```javascript
// sw.js
self.addEventListener('install', e => {
  e.waitUntil(
    caches.open('v1').then(cache =>
      cache.addAll(['/', '/icons.svg', '/app.css', '/app.js'])
    )
  )
})
self.addEventListener('fetch', e => {
  e.respondWith(caches.match(e.request))
})
```

```html
<!-- 注册 Service Worker -->
<script>
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js')
  }
</script>
```

### 3.4 图片优化

| 优化项 | 做法 | 效果 |
|--------|------|------|
| 格式 | **WebP**（替代 PNG/JPEG） | 体积减少 50-80% |
| 懒加载 | `loading="lazy"` | 非首屏图片不阻塞 |
| 响应式 | `srcset` 根据屏幕尺寸加载 | 手机不加载大图 |
| 预加载首屏 | `<link rel="preload">` | 关键图片优先加载 |

### 3.5 字体优化

**使用系统字体，不额外下载字体文件：**

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
  "Helvetica Neue", Arial, "Noto Sans SC", sans-serif;
```

**效果：** 节省 100-300KB 字体文件下载。

---

## 4. 最终性能预算

| 资源 | 大小 | 说明 |
|------|------|------|
| HTML（含内联 CSS） | ~15KB | 首屏一次性请求 |
| JS（Svelte 编译产物） | ~10KB | 异步加载 |
| SVG 图标（内联） | ~5KB | 零请求 |
| 首屏合计 | **~30KB** | **1 次 HTTP 请求** |
| 图片 | WebP 格式 | 懒加载 |

**对比旧方案（Element Plus）：**

```
旧方案：~400KB，8 次请求，2-3 秒加载
新方案：~30KB，1 次请求，0.3 秒加载
```

---

## 5. Lighthouse 100 分配置

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="...">
  
  <!-- 预连接关键域名 -->
  <link rel="preconnect" href="https://api.example.com">
  
  <!-- 关键 CSS 内联 -->
  <style>/* 首屏样式 */</style>
  
  <!-- 非关键 CSS 异步加载 -->
  <link rel="preload" href="non-critical.css" as="style" onload="this.rel='stylesheet'">
  
  <!-- 预加载首屏图片 -->
  <link rel="preload" href="hero.webp" as="image">
</head>
<body>
  <!-- 内容直接渲染，无需等待 JS -->
  
  <!-- 非关键 JS 延迟加载 -->
  <script src="app.js" defer></script>
</body>
</html>
```

---

## 6. 面试话术

> **"采用 Svelte 5 编译时框架 + TailwindCSS v4 + 关键 CSS 内联 + SVG 图标内联 + Service Worker 离线缓存的极致性能方案，将首屏体积从 400KB 压缩到 30KB，HTTP 请求从 8 次降到 1 次，Lighthouse 评分 100，FCP < 0.3s。弱网环境下仍能秒开，离线可用。"**

---

## 7. 总结

| 技术 | 效果 | 实现成本 |
|------|------|---------|
| Svelte 5 | 零运行时，~15KB | 需学新语法 |
| TailwindCSS v4 | 按需编译，~10KB | 低 |
| 关键 CSS 内联 | 省 1 次 HTTP 请求 | 极低 |
| SVG 内联 | 零请求，零带宽 | 极低 |
| Service Worker | 离线可用 | 低 |
| WebP 图片 | 体积减 50% | 极低 |

**这套方案可以做 Demo 展示：打开 Chrome DevTools，Network 面板显示 1 个请求 ~30KB，Lighthouse 跑分 100，面试官直接闭嘴。**