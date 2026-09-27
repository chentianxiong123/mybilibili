# docs/agents/web.md —— web agent 领域文档

> 领域：桌面 Web 用户端（Nuxt SSR），代码在 `web/user/**`
> 现状：2026-09-27

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

## 遗留问题（按优先级）

1. **`/message/whisper` 整页死**（`pages/message/whisper/index.vue` 362 行 + 依赖
   Vuex store 的 chatList/chatId/updateChatList）。多个页面链到它
   （`space/[uid].vue`、`message.vue`、`platform.vue`），不能直接删，需改写或
   用新 `/message/conversations` 系 API 替换。
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
- 不删 `pages-old/`、`web/next-old/`（历史留存）