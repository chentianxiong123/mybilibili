# docs/agents/web.md —— web agent 领域文档

> 领域：桌面 Web 用户端（Nuxt SSR），代码在 `web/user/**`
> 现状：2026-10-10

## 职责边界

- 只改 `web/user/**`（app/pages、app/components、app/api、server/api Nuxt proxy）
- 后端 `services/core/**`、`shared/**` 不是独占，但 web 涉及的接口改动优先自己处理
- 测试：`bash scripts/test.sh frontend`（web/user vitest，当前 932 条）

## 技术栈

Nuxt 4 + Vue 3 + Element Plus。**Nuxt SSR**：`web/user/app` = app 目录，
`web/user/server/api/[...path].ts` 是 teriteri 前端 URL → Go 后端路径的代理
（PATH_MAP，约 30 条映射）。dev 模式端口 3200，traefik `/` 兜底。

## 关键文件

| 文件 | 作用 |
|---|---|
| `web/user/server/api/[...path].ts` | 前端 URL → Go 后端 PATH_MAP 代理（CORE/SEARCH/MSG/LIVE 分端口）|
| `web/user/app/api/client.ts` | axios + 401 处理（连续 3 次 401 才清会话，见 MAX_401_BEFORE_DROP）|
| `web/user/app/api/subtitle.ts` | 字幕 API（对象双份：user 自用 + admin 管理站在同一文件）|
| `web/user/app/api/userPrivacy.ts` | 隐私 + 标签 API —— ⚠️ tags 路径疑似错位，见「遗留」|
| `web/user/app/api/message.ts` | 私信 API（getAtList 已补，whisper 相关看遗留）|
| `web/user/app/views/web/UserProfileView.vue` | 922 行大组件（视频列表/关注/搜索都在这里）|

## 已验证的事实（别再查）

- `/user/pinned-video` GET 是公开的（看别人主页 ProfileTabVideoList 也会拉），
  POST/DELETE 才是自身的。GET 不该锁。
- `getVideosByUserId` → `/manuscript/user/{id}?status=3` 返回 `{list,total}`，
  不是数组（已修过一处 `.map`：response.data.list.map）。
- web/user 的 `/message/whisper` → teriteri 旧私信 UI，调 `/msg/chat/*`（后端
  无此路由），整页死。见遗留。
- 隐私设置 camelCase 契约：`publicCollection`/`publicBirthdayTags`/`publicCoinVideos`/
  `publicLikeVideos`/`publicFollowingList`/`publicFollowersList`（表列 snake_case）。
- 全站导航已统一为 teriteri `HeaderBar`（`.header-bar` 64px）：`layouts/simple.vue`/
  `home.vue` 用布局，首页/空间/搜索/账号页内自带（banner 叠加需要），创作中心
  `platform.vue` 自有独立头。旧 Element 风格 `AppHeader.vue`（80px）已删，无残留引用。
- 消息中心已统一为 teriteri 侧栏框架（`pages/message.vue` + `[type].vue` 别名跳转，
  背景 `message-bg.png` + `msgUnread` 红点），内容沿用现有 List 实现；新增
  `message/ai.vue`（AI 客服）。旧 `message/index.vue`（自带 HeaderBar 导致双导航）已删。
- `app/pages-old/` 缓冲目录已删（Git 历史可查）；运行代码只剩 `app/pages/` 一份。
  `web/next-old/`（230 文件历史快照）、`web/teriteri-reference/`（原版对照）保留。
- 测试按领域分 5 组（`test:api/composables/utils/components/stores`，vitest projects），
  全量 `npm test` 932 条不变。

## 遗留问题（按优先级）

1. **`/message/whisper` 已修好**（2026-10-10，前端改调新接口，后端不动）。
   改法：`pages/message/whisper.vue` 新父组件（左栏会话+`<NuxtPage/>`，原 index.vue
   的 `<router-view>` 嵌套在 Nuxt 下不生效）；`whisper/index.vue` 瘦身为占位+自动
   跳首个会话；`[mid].vue` 对话框调 `messageApi.sendMessage`/`markConversationRead`，
   无会话时用 `/user/info/get-one` 补标题，发首条后占位转正；`MessageList.vue`
   调 `getMessages` 分页（DESC 需反转+按 id 去重），撤回下线（后端无接口+原走
   已不存在的 ws）；`adaptChatConversation` 不再预置伪消息（会与第一页重复），
   预览走新增 `preview` 字段；`updateChatList` 加 Array 守卫（401 降级曾致白屏）。
   已知限制：无撤回。实时推送已接上（见下）。
- 私信实时推送（2026-10-10 接上）：设计见
  `docs/archive/backend/refactor/20-websocket-sse-http-comparison.md`（单向推送用 SSE）。
  链路：`whisper.vue` 的 EventSource → `/sse/notification` → traefik 指到
  msg-danmaku:8086（compose/k3s ingress 里 `/sse/notification` 之前错指到 core，
  实测 404，已改）。事件：`unread_init`/`unread_counts` 校准红点，`message`
  增量进会话（取最新一条落本地，正看着的直接标已读）。注意两坑：
  （1）挂载时 pinia isLogin 可能还没就绪，必须 watch isLogin 补连；
  （2）后端 SendMessage 曾只写发送方会话行（接收方永远收不到），已改双写，
  见 Go 侧注释；migration 不需要（表结构不动）。
   教训：commit 进 store 前的原始对象直接改不触发视图更新，改完必须从
   `store.state` 里重新取 reactive 代理再改（见 `[mid].vue ensureChatItem` 注释）。
2. **`/user/privacy/tags` vs `/user/tags`**：前端 `userPrivacyApi.userTags` 三个
   函数调 `/user/privacy/tags`，但后端 `handlePrivacy` 只处理 GET(隐私JSON)/PUT，
   不做 tags；真正的 tags 在 `/api/v1/user/tags`（handleTags）。等于用户标签
   功能前端一直调错端点，从没生效。修法：前端三个函数改指 `/user/tags`。
3. `/api/v1/ai/summary/*`、`/creator/comments*`、`/studio/export-tasks` 等
   creator/内部端点未逐一核实该不该锁。
4. 遗留 Vue warn：ElInput `rows="2"`（字符串 vs Number）、ElMenu `router="true"`
   （字符串 vs Boolean）。功能无碍，纯类型告警。

## 不要做

- 不碰 `mobile/wap/**`、`web/admin/**`（其他 agent 领地）
- 不 stage `.air/work.toml`、`external/cf-whisper-worker/.wrangler/**`
- 不删 `web/next-old/`、`web/teriteri-reference/`（历史留存与原版对照）