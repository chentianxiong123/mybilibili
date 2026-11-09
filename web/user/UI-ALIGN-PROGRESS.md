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
| 首页-推荐卡片 | `/` | ✅ 已对齐 | `videoCard.css`：owner 行高锁 17px（修 info 68→65）、标题右留白 24→30px。卡片 312×251 / cover 312×176 / info 312×65 / bottom 312×17 / tit 312×44 全部对齐 B 站（commit 526c8e51） |
| 首页-顶部导航+全局 | `/` | 🟡 大致对齐可过 | body 背景 `#fafafa`→`#f1f2f3`、文字色 →`#18191c`；左导航容器 margin-right 20→0、导航项间距 15→20（commit 5c36ac77）。导航项字号 14px/宽度/间距已与 B 站一致；剩余宽度差(566 vs 472)是 B 站多出「赛事/S16」等**动态内容项**，非 CSS 偏差，不强行补 |
| 首页-搜索框 | `/` | 🟡 大致对齐可过 | 搜索条本身 500×40 两端一致；位置差是左导航宽度差的连带，已随导航修正改善（687→694，B 站 759） |
| 首页-全局字号 | `/` | ⏸️ 已知偏差 | B 站 body `16px/normal`，本地 `14px/1.25`（lineHeight 22.4 vs 17.5 根因）。**全局改 16px 回流风险大**，无人值守下不盲改，需逐页评估 |
| 搜索页-视频结果 | `/search/video` | ✅ 已对齐 | 复用已对齐的 `.video-card`。`search/video.vue`：1700–2199 断点网格 6→7 列（卡片 271→230px，B 站 233px）；`videoCard.css` 标题字重 400→500。卡片 230×204 vs B 站 233×205 |

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