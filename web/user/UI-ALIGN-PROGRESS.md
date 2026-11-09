# UI 对齐进度表（web/user ↔ 真实 bilibili.com）

> 方法：无头 Chrome（`/usr/bin/google-chrome`）+ Playwright，B 站登录态来自复制 profile 的 CDP 实例注入 cookie。
> 所有 CSS 值来自 `getComputedStyle()` 真实抓取，**本地与 B 站在同一个 1920×1080 / dpr1 视口**下对比（`/home/a1/bili-align/scripts/compare2.mjs`）。
> 动态内容通过 Playwright route 拦截 `/api/*` 注入样本数据（不改后端），使真实前端渲染出真实 DOM 再测量。

## 状态图例
- ✅ 已对齐（本轮用 computed-style 验证）
- 🟡 大致对齐可过
- ⬜ 待对齐
- ⏸️ 已知偏差，需谨慎后续处理

## 页面进度

| 页面 | 路由 | 状态 | 本轮改动 / 备注 |
|---|---|---|---|
| 首页-推荐卡片 | `/` | ✅ 卡片已对齐 | `videoCard.css`：owner 行高锁 17px（修 info 68→65）、标题右留白 24→30px。卡片 312×251 / cover 312×176 / info 312×65 / bottom 312×17 全部对齐 B 站 |
| 首页-顶部导航 | `/` | ⬜ | body 背景 `#fafafa`→B 站 `#f1f2f3`、body color →`#18191c`（纯色低风险）。左导航项间距 15→20、容器 margin；搜索框圆角 0→6、容器左右 margin 30px —— 待做 |
| 首页-全局字号 | `/` | ⏸️ | B 站 body `16px/normal`，本地 `14px/1.25`。是 lineHeight 22.4 vs 17.5 的根因。**全局改 16px 回流风险大**，无人值守下不盲改，需逐页评估 |

## 待对齐页面（优先级）
1. 视频播放页 `/video/[vid]`
2. 空间页 `/space/[uid]`
3. 搜索页 `/search`
4. 消息页 `/message`
5. 动态页 `/dynamic`

## 环境备忘
- B 站 cookie：`/home/a1/bili-align/bili-cookies.json`（31 条，含 SESSDATA/bili_jct/DedeUserID，已确认登录态）。
- CDP Chrome：`google-chrome --headless=new --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-bili-profile`（profile 从默认复制，已登录 B 站）。
- 本地前端：`cd web/user && npx nuxt dev --port 3200`（已清理 root 残留的 `.nuxt`、`node_modules/.cache`）。
- 对比脚本：`node /home/a1/bili-align/scripts/compare2.mjs <pairs.json> <out.json>`。
- **后端未启动**（Go monorepo 需 DB/NATS/MinIO，风险大）；改用 Playwright route 拦截 mock `/api/*`。
- API 契约：`$get` → `request.ts`（baseURL `/api`），拦截器要求 `code===200`，列表接口经 `adaptCard`（raw: `{id,coverUrl,title,durationSeconds,createdAt,viewCount,...,uploader:{id,name,avatar}}`）。

## 铁律遵守
- 只 commit，绝不 push。✅
- 所有 CSS 值来自真实 `getComputedStyle()`。✅
- `npx vitest run`：949 passed（本轮无新失败）。✅
- 只改 `web/user` 前端模板/样式。✅