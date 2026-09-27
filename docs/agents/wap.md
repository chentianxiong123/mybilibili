# docs/agents/wap.md —— wap agent 领域文档

> 领域：移动端 H5（Vite + Vue3 + SCSS），代码在 `mobile/wap/**`
> 现状：2026-09-27

## 职责边界

- 只改 `mobile/wap/**`（src/api、src/views、src/utils、index.html、vite.config）
- 测试：`bash scripts/test.sh e2e`（`tests/e2e/wap/**`，当前 47 条含 wap 40 条）
- WAP 是 Vite dev server：容器 `mybilibili-wap`，volume `../mobile/wap:/app`，
  port 5174，traefik `PathPrefix('/wap/')`。改代码 HMR 自动生效，无需重建。

## 技术栈

Vue 3 `<script setup>` SPA，Vite `base:'/wap/'`（**所有静态资源引用路径都会被
重写一次 /wap/ 前缀**——HTML/相对路径里的 /wap/ 会重复，这是历史 bug 源的
根源）。移动端 375×667 布局，localStorage 前缀 `wap:`。

## 关键文件

| 文件 | 作用 |
|---|---|
| `src/utils/storage_layer.ts` | localStorage 封装，PREFIX=`wap:`，键 K.searchHistory=`search:history` |
| `src/utils/session.ts` | `isLogin`/`getLocalUserId`/`saveSession`——登录态唯一信号是 localStorage `wap:user`（HttpOnly cookie 读不到）|
| `src/api/client.ts` | axios，401 尝试 tryRefresh 一次 |
| `src/views/message/Chat.vue` | 私信聊天（isMine 用 getLocalUserId）|
| `src/views/live/Room.vue` | 直播间（之前 ref<any> 白屏 bug）|

## 已验证的事实（别再查）

- e2e fixture：`conftest.py` 有 `wap_url`、`mobile_page`（未登录）、`wap_session`
  （API 登录 + 写 `wap:user`）、`clean_history`（须先 goto 同源页否则 SecurityError）。
- WAP 登录凭证同样 HttpOnly cookie；前端只靠 `wap:user` 判断登录。
- `KEY:searchHistory='search:history'` 写 `wap:search:history`；别用裸键。
- 24 条路由已扫过（匿名+登录）；直播房/消息/搜索/排行都覆盖了 e2e。

## 遗留问题

- 无大项待办（13 个静默 bug 已修 + 47 e2e 全绿）。
- 可继续探索的方向：投稿上传流（/platform 系在 web 不在 wap，WAP 的创作入口？
  实际 WAP 只有浏览侧）、弹幕发送、边下边看、黑暗模式切换后的持久化。

## 不要做

- 不碰 `web/user/**`、`web/admin/**`（其他 agent 领地）
- 不 stage `.air/work.toml`、`external/cf-whisper-worker/.wrangler/**`