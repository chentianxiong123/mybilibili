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
| 消息页 | `/message` | ✅ 已对齐 | `message-center.css`：侧栏菜单项 `.message-sidebar__item-name` font-weight 400→700（B 站全项 700，激活 #00aede/未激活 #61666d 颜色本已一致）；背景层 `.message-bgc`/`.message-bg` z-index 0→-1（B 站真值）。DOM 类名与 B 站 message-pc 完全同名，各容器 rect 全部一致（commit 53df0164） |
| 动态页 | `/dynamic` | ✅ 已对齐 | `DynamicView.vue`：容器 1200→1330px 居中（B 站 x=295）、三栏 240/580/300→264/724/318（gap 20→12，对齐 B 站 t.bilibili.com）、内容起点 y84→y72、背景 `#f5f7fa`→`#d3e9e8`（B 站 `.bgc` 真值）、卡片圆角 8→6px/去 box-shadow、区块间距 16→8px（commit f3842240） |

## 待对齐页面（优先级）
1. 视频播放页 `/video/[vid]` — 🟢 **已解除受阻（重大更正）**：之前的 500 是**对比脚本的 mock 数据形状错误**（`/video/getone` 返回 `data:[]` 数组，页面期望对象），并非页面本身受阻。用正确形状 mock（含可播放 mp4 的 `videos[].playUrl`）后页面 **200 正常渲染**：artplayer 初始化（`.art-video-player` 1264×711）、标题/清晰度/倍速/弹幕框全出。**剩余唯一偏差**：左栏宽 1264px vs B 站 1354px、起点 x107 vs 63（右栏 411px/gap 30px 已一致）。左栏宽由 `changeWindowSize()` 的 `0.7*(vh-64)` 高度分数公式决定，改它需谨慎（影响 16:9 高度联动），建议有人监督时调整分数（0.7→0.75 可得 1354）。渲染验证脚本：`/home/a1/bili-align/scripts/video-mock.json`。
2. 空间页 `/space/[uid]` — 🟡 大致对齐可过，但有**结构性差异**：B 站当前 space（space.bilibili.com）用**通栏** banner（1920×200 `space-header`）+ 通栏 nav（1920×64 `space-navbar`）+ 1640px 居中内容（`space-home` x=140）；本地用**居中盒装** 1283px wrapper（banner 1283×192、nav 1283×66）。这是通栏 vs 盒装的布局差异，非 CSS 微调，且影响所有 `/space/[uid]/*` 子路由，无人值守下不做重建（风险大）。头像 60/64 vs B 站 ~68。建议有人监督时评估是否迁移到通栏布局。
3. ~~消息页~~ — ✅ 已对齐（见上表）。
4. ~~动态页~~ — ✅ 已对齐（见上表）。
5. 视频页之外其余（直播/专栏/收藏/历史）— ⬜ 未开始。

## 本轮已交付（2026-11-09 夜班，2 个页面对齐 + 1 个受阻解除，均只 commit 未 push）
- `53df0164` 消息页：侧栏菜单 font-weight 400→700、背景层 z-index 0→-1（computed-style 对齐 message.bilibili.com）。
- `f3842240` 动态页：容器 1200→1330 居中、三栏 264/724/318（gap 12）、内容起点 y72、背景 #d3e9e8、卡片圆角 6px/去阴影、区块间距 8px（对齐 t.bilibili.com）。
- 视频播放页：**解除受阻**（更正进度表——之前 500 是 mock 形状错误，非页面问题；现可正常渲染），左栏宽偏差已记录待监督调整。
- 高杠杆：`videoCard.css` 被首页/搜索/空间/动态共用，前轮对齐惠及所有用卡片的页面。

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