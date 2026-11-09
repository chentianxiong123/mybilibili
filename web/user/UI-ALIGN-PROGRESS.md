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
| 首页-推荐卡片 | `/` | ✅ 已对齐 | `videoCard.css`：owner 行高锁 17px、标题右留白 30px。卡片 312×251 / cover 312×176 / info 312×65 全部对齐（commit 526c8e51）。本轮 verify 无回归 |
| 首页-顶部导航+全局 | `/` | 🟡 大致对齐可过 | 导航项字号/宽度/间距一致；剩余宽度差是 B 站多出「赛事/S16」等**动态内容项**，非 CSS 偏差 |
| 首页-搜索框 | `/` | 🟡 大致对齐可过 | 搜索条 500×40 两端一致 |
| 首页-全局字号 | `/` | ⏸️ 已知偏差 | B 站 body `16px/normal`，本地 `14px/1.25`。**全局改 16px 回流风险大**，无人值守不盲改 |
| 搜索页-视频结果 | `/search/video` | ✅ 已对齐 | 1700–2199 断点网格 7 列（卡片 230px，B 站 233px）、标题字重 500 |
| 消息页 | `/message` | ✅ 已对齐 | 侧栏菜单 font-weight 700、背景层 z-index -1；各容器 rect 全部一致（commit 53df0164） |
| 动态页 | `/dynamic` | ✅ 已对齐 | 容器 1330px 居中、三栏 264/724/318(gap 12)、内容起点 y72、背景 #d3e9e8、卡片圆角 6px（commit f3842240） |
| **历史记录页** | `/history` | ✅ **本轮对齐** | `HistoryView.vue` 整页重构对齐 `bilibili.com/account/history`：白底通栏 1920 + **1432px 居中容器**（左内边距 72px）、头部标题 **28px/600/40px** + 右侧记录开关/视图切换、**粘性筛选条 74px**（综合/视频/直播/专栏 `radio-tabs` + 底部 28×4 游标 #00aede）、搜索框 **200×34 灰底 #f1f2f3 圆角 8**、清空/批量按钮 **100×34 圆角 8 border #e3e5e7**、时间轴**左侧 32px 锚点轨道**（#c9ccd0，margin-right 40）、卡片网格 **1360px / 5 列 / gap 48×16** / 封面 259×146 圆角 6（commit 116c22f1） |
| **专栏首页** | `/read/home` | ✅ **本轮对齐** | `read/home.vue` 整页重构对齐 `bilibili.com/read/home`：**680px 居中单列**（main 680 / padding-top 30）、头部 **40px**（标题「推荐文章」18px/25px + 搜索框 **260×40 灰底圆角 8**）、图文信息流条目 **680×128 padding 20/0**、标题 **20px/26px nowrap**、摘要 12px/22px #9499a0、meta 24px（头像 24 圆形 + 作者 max-width 120 + 标签 + 阅读/点赞/评论 间距 25）、封面 **117×88 圆角 4** 右侧 margin-left 20（commit a02b6c3b） |
| **视频播放页-左右栏宽度** | `/video/[vid]` | ✅ **本轮对齐** | `[vid].vue`：播放器高度分数 **0.7→0.75**（711→762px）、左栏 **1264→1355px**（B 站 1354）、右栏 **350→411px**（B 站 411）、左栏起点 **x107→x62**（B 站 63）。播放器 16:9 联动不变（commit 278ea49c） |
| **视频播放页-工具栏/简介/标签** | `/video/[vid]` | ✅ **本轮对齐** | 工具栏 `content-box` 高 **48 + padding-top 18**（总高 **67 一致**）、左侧项 **92→100px / 字号 13→14 / 间距 8→18px**、简介容器 margin **16→18px**、简介文字 **15/24→16/26** + max-height 84 折叠、标签容器 padding-bottom **6→10** / margin-top **16→18**、标签胶囊 **28→32px 高 / 字号 13→14 / 圆角 14→16 / 底色 #f1f2f3**（commit 79757d42） |
| **个人中心页** | `/personal-center/*` | ✅ **本轮对齐** | `PersonalCenterView.vue`：容器 **1200→980px 居中**（content-box + 1px #e1e2e5 边框 / 圆角 4 / 阴影，与 `account.bilibili.com` `.security_content` 一致 **982×@469,180**）、页面背景 **#f5f7fa→#fff**、侧栏 **200→150px** 去卡片阴影、标题 **50px 居中 #99a2aa**、导航项 **48px 行高**（图标 18×36 margin-left 16 + 文字 14px margin-left 12）、激活态 **#00a1d7 实底白字**去左边框、内容面板去圆角阴影（commit 00edbbef） |
| 空间页 | `/space/[uid]` | ⏸️ 结构性差异 | B 站当前 space 用**通栏** banner(1920×200) + 通栏 nav(1920×64) + 1640px 居中内容；本地用**居中盒装**。通栏 vs 盒装的布局差异，影响所有 `/space/[uid]/*` 子路由，无人值守不重建（风险大） |
| 直播首页 | `/live` | ✅ **本轮对齐（推荐直播区+hero）** | `LiveListView.vue` 整页重构对齐 `live.bilibili.com`：通栏 `margin 0 20px`（容器 1880 透明底融入站点灰）、分类胶囊条（高 28 / 圆角 8 / 灰底 #f6f7f8 / 字 14-20，激活 #00a1d6）、**主播大播放位 hero 992×558**（#232527 + 直播中/观看数角标 + 悬停进入按钮，右侧信息栏 margin-left 13 / 865 宽 / #f6f7f8 圆角 8）、**推荐直播 5 列网格**（卡宽 (100%-56px)/5=363·B站 362.8、列间距 14 / 行 15、卡片圆角 4、**封面 16:9 灰底 #f1f2f3 圆角 6**、信息 margin-top 5 高 44、标题 14/500/24 #18191c nowrap、副标题 12/20 #9499a0）（getComputedStyle 真值，949 测试全绿）。注：B 站顶部独立暗导航/分类 mega-menu 与 banner 轮播属另一套体系，本地沿用站点 HeaderBar + simple layout，此处只对齐内容区 |
| 合集和列表页 | `/collections` | ⬜ 未对齐 | 本地 `CollectionListView.vue` 为自定义盒装网格；B 站对应页在 `space.bilibili.com/{uid}/lists`，套在通栏 space 布局内，依赖空间页先迁移 |
| 搜索页-用户结果 | `/search/user` | 🟡 大致对齐可过 | 容器 **1728 居中 @x96**、搜索头 **122 高 / 输入框 640×54 / tabs 44** 全部与 B 站 `search.bilibili.com/upuser` 一致（本轮 computed-style 验证）；仅**用户卡片内部**（标题 20/600、描述 13 #61666d、关注按钮 100×32 圆角 8）未逐像素核。卡内对齐受阻：`searchUsers()` 在 `created()`（SSR 服务端）调用，axios baseURL 指向 `http://core:8080`，无后端时无法在浏览器渲染卡片测量 |
| **专栏详情页** | `/read/cv[id]` | ✅ **本轮对齐** | `[id].vue` 整页对齐 `bilibili.com/read/cv*`：容器 **720+padding24→708px 居中白卡**（圆角 6 + 顶部 1px #e3e5e5 内阴影 + padding-bottom 10，起点 y72）、页面背景 **#fff→#f1f2f3** + padding-top 72、标题 **padding 20/56 + 24/700/36**、作者区 **padding 12/56 + 头像 48 + 名字 17/600/24 + 时间 13/18 #9499a0**、正文 **padding 0/56/36/56 + 17/29 + 段落 #61666d + h1 19/500/32 margin-top 36**、右侧悬浮栏 **26→58 宽 @x1326 top180**（动作 46×42 间距 58 / 图标 24 / 文字 12-14）（commit d9cbc8f6） |

## 本轮已交付（2026-11-10 凌晨班，5 个页面对齐，均只 commit 未 push）
- `116c22f1` 历史记录页：整页重构（通栏白底 + 1432 容器 + 粘性筛选条 + 时间轴 + 5 列卡片网格）。
- `a02b6c3b` 专栏首页：整页重构（680 单列 + 推荐文章头 + 图文信息流）。
- `278ea49c` 视频播放页：左栏 1355 / 右栏 411 / 起点 x62（高度分数 0.7→0.75）。
- `79757d42` 视频播放页：工具栏 67 高 / 简介文案 16-26 / 标签胶囊 32px。
- `00edbbef` 个人中心页：980 容器 + 150 侧栏 + 48px 导航项 + #00a1d7 激活态。

## 本轮已交付（2026-11-10 晚班，1 个页面对齐 + 1 个页面核验，均只 commit 未 push）
- `d9cbc8f6` 专栏详情页：整页对齐（708 居中白卡 + 标题/作者/正文 56px 内边距体系 + 右侧悬浮栏 58 宽 @x1326）。
- 搜索页-用户结果 `/search/user`：核验容器 1728 + 搜索头 122/54/44 已一致（标 🟡 可过），卡内对齐受阻于 SSR 无后端。
- `d9cbc8f6` 专栏详情页（2026-11-10 晚班）：708 居中白卡 + 标题/作者/正文 56px 内边距体系 + 右侧悬浮栏 58 宽。

## 本轮已交付（2026-11-09 夜班，1 个页面对齐，只 commit 未 push）
- 直播首页 `/live`：`LiveListView.vue` 整页重构对齐 `live.bilibili.com` 内容区（分类胶囊条 + hero 992×558 播放位 + 推荐直播 5 列网格 363/封面 16:9 圆角 6/标题 14-500-24/副标题 12-20-#9499a0/卡片圆角 4）。关键排障：mock 必须排除 `/_nuxt/` 路径（否则把 Nuxt 源码模块当 API 拦截致白屏），且端点带 `/v1` 前缀为 `/api/v1/live/room/*`。网格负 margin 会撑大卡宽（去掉后回落 363）。949 测试全绿。

## 待对齐页面（优先级）
1. **合集和列表页** `/collections` — ⬜ 依赖空间页通栏布局，建议与空间页一起评估。
2. **空间页** `/space/[uid]` — ⏸️ 通栏 vs 盒装结构差异，需有人监督时做布局迁移。
3. 直播首页 `/live` — 已对齐内容区（hero + 推荐直播网格）；B 站顶部独立暗导航/banner 属另一套体系，未迁移。

## 环境备忘
- B 站 cookie：`/home/a1/bili-align/bili-cookies.json`（31 条，含 SESSDATA/bili_jct/DedeUserID，已确认登录态）。
- CDP Chrome：`google-chrome --headless=new --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-bili-profile`（profile 从默认复制，已登录 B 站）。
- 本地前端：`cd web/user && npx nuxt dev --port 3200`（本轮已在跑）。
- 对比脚本：`node /home/a1/bili-align/scripts/compare2.mjs <pairs.json> <out.json>`；辅助脚本 `bili-outline.mjs`（DOM 大纲）、`bili-cs.mjs`（批量 computed style）、`bili-html.mjs`（元素 HTML）、`local-cs.mjs` / `local-chain.mjs`（本地同款）、`shot2.mjs`（截图）、`verify-local.mjs`（回归巡检）。
- 各页 pairs 配置：`history-pairs.json` / `read-pairs.json` / `video-pairs.json` / `pc-pairs.json` / `search-pairs.json` / `home-pairs.json` / `message-pairs.json`。
- **后端未启动**（Go monorepo 需 DB/NATS/MinIO，风险大）；改用 Playwright route 拦截 mock `/api/*`。
- API 契约：`$get` → `request.ts`（baseURL `/api`），拦截器要求 `code===200`，列表接口经 `adaptCard`（raw: `{id,coverUrl,title,durationSeconds,createdAt,viewCount,...,uploader:{id,name,avatar}}`）。

## 铁律遵守
- 只 commit，绝不 push。✅（本轮 5 个 commit，无 push）
- 所有 CSS 值来自真实 `getComputedStyle()`。✅
- `npx vitest run`：949 passed（本轮每次改动后均跑，无新失败）。✅
- 只改 `web/user` 前端模板/样式（`HistoryView.vue` / `read/home.vue` / `video/[vid].vue` / `PersonalCenterView.vue`）。✅