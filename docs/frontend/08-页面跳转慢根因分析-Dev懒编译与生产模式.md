# 页面跳转慢根因分析：Dev 模式懒编译 vs 生产模式

> 背景：演示时每个新页面跳转都很慢（某页面首次访问高达 20 秒），切换到生产模式后所有页面 3~30ms 秒开，"快得离谱"。本文记录排查过程与结论。

## 1. 现象

- **Dev 模式**（`pnpm dev`）下：每个**没去过的**新页面首次跳转都慢（几秒 ~ 20 秒）；进过一次的页面再回去就正常。
- **生产模式**（`nuxt build` + `node .output/server/index.mjs`）下：所有页面首次访问即秒开。

## 2. 究极原因

### 根因一（主因）：Vite Dev 懒编译

Nuxt dev 模式下，每个页面路由的 JS 代码块是**首次访问时才由 Vite 实时编译**的：

- 点按钮跳转到一个没去过的路由，浏览器要先等 Vite 编译该页面**整棵依赖树**；
- 页面组件越重、import 越深，首次编译越慢；
  - 实测：`UserProfileView.vue` 有 19 个 import，900+ 行，依赖 ElPlus 等大量组件；
  - 实测证据：`/profile/8` 首次 SSR 请求 **20.2s** → 再次请求仅 **51ms**。
- 生产 `nuxt build` 提前把所有页面预编译进 `.output`，**没有按需编译等待**，跳转秒开。

### 根因二（体感卡顿叠加）：数据全在 `onMounted` 拉取

整个项目所有视图没有使用 `useAsyncData` / `useFetch`，数据全部在 `onMounted` 里 `await`：

- SSR 先渲染**空白页** → 客户端水合 → 再发 API → 再渲染内容；
- 每次跳转都伴随"白屏闪烁 + 先空后填"，放大慢的体感。

### 根因三：每页并发 API Storm

header（`AppHeader.vue`）挂载时并发发起 6+ 个请求（用户信息、未读数、热搜、收藏、历史、动态），与页面自身请求挤在同一瞬间，抢占主线程。

### 根因四：Hydration Mismatch 触发整树重建

服务端与客户端渲染结果不一致时，Vue 会销毁并重建整棵组件树。已修复的案例：

- `CategoryTabs.vue`：`Math.random()` 在 setup 中执行 → SSR 与客户端随机结果不同 → 修复为 `useState` 同步；
- `VideoView.vue` / `VideoCommentSection.vue` / `CommentSystem.vue`：`localStorage` 在 setup 中读取，SSR 恒为 `null`、客户端有值 → 修复为 `onMounted` 中初始化 ref。

## 3. 实测数据

### 后端 API 延迟（排除后端嫌疑）

```
core-service  GET /api/v1/user/info/8         0.5 ~ 0.7 ms
msg-danmaku   GET /api/v1/message/unread      1.5 ~ 2.0 ms
```

### SSR 页面加载时间对比

| 路由 | Dev 第一次 | Dev 第二次 | 生产模式 |
|------|-----------|-----------|---------|
| `/` | 2.7s（冷启动） | 0.38s | **14 ms** |
| `/profile/8/home` | **20.2s（懒编译）** | 51 ms | **18 ms** |
| `/manuscript/1` | 首次数秒级 | — | **174 ms** |
| `/category/1` | — | — | **10 ms** |
| `/history` | — | 31 ms | **31 ms** |
| `/message` | — | 0.52s | **41 ms** |
| `/dynamic` | — | 0.46s | **49 ms** |
| `/search` | — | — | **22 ms** |

## 4. 结论

1. **生产模式没有懒编译问题**——所有页面预编译，跳转秒开。dev 模式的慢是开发工具链特性，**不代表最终成品性能**。
2. 演示 / 验收 / 交付时一律使用生产模式运行。
3. 需要用 dev 模式日常开发时，将一个页面编译后的延迟误判为自研代码问题，是本误区的核心教训。

## 5. 启动方式

```bash
# 生产模式（演示 / 验收推荐）
cd apps/web
npx nuxt build
setsid env PORT=3200 NITRO_PORT=3200 node .output/server/index.mjs & disown

# Dev 模式（日常开发，有热更新但首次访问有编译延迟）
pnpm dev
```

## 6. 后续可优化项（可选）

- 所有视图改为 `useAsyncData`/`useFetch`，SSR 直接渲染真实内容，消除白屏闪烁；
- header 下拉组件（收藏/历史/动态/search 热搜）改为 hover 时懒加载，减少并发请求；
- 登录/注册 dialog 从 `app.vue` 全局挂载改为按需懒挂载。