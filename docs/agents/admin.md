# docs/agents/admin.md —— admin agent 领域文档

> 领域：管理后台（Vite SPA，Vue 3 + Element Plus），代码在 `web/admin/**`
> 现状：2026-09-27

## 职责边界

- 只改 `web/admin/**`（app/router、app/api、app/views/admin）、`shared`（如
  需要会触碰 admin 鉴权的部分）
- 测试：`bash scripts/test.sh frontend`（web/admin vitest，当前 481 条）
- dev 端口 3100，traefik `PathPrefix('/admin/')`，Vite history base `/admin/`

## 技术栈

Vite SPA + Vue Router（`createWebHistory('/admin/')`）。路由在
`web/admin/app/router/index.ts`，**每个 route 带 `meta.permission`**
（如 `statistics:manage`），`firstAllowedPathByPermissions` 决定默认着陆页。
权限来自后端 `/api/v1/admin/login` 返回的 `permissions` 数组。

## 关键文件

| 文件 | 作用 |
|---|---|
| `app/router/index.ts` | 27 条路由 + permission meta |
| `app/api/client.ts` | axios + admin session（注意它有段 401 里 `window.location.href='/admin/login'`）|
| `app/utils/auth.ts` | `hasAdminSession` |
| `app/views/admin/*.vue` | 27 个视图（UsersView/ManuscriptsView/DashboardView 等）|
| `app/views/admin/LoginView.vue` | 登录页 |

## 已验证的事实（别再查）

- **后台账号**：`super_admin` / `admin123`（admin_level=2，21 权限全覆盖）。
  数据库里还有 review_admin/operation_admin/support_admin/data_analyst 等低权限账号。
- 27 条路由已用 Playwright + super_admin 登录态扫过：**无真实 bug**。唯一的
  「/subtitle-management 无路由」是我探测脚本误把 embedded tab 当路由——它实际
  是 ContentReviewView 里的 `v-else-if="activeTab==='subtitle'"` 子组件，不是路由。
- 鉴权挂 `auth.AdminPathGuard`，管理口路径含 `/admin/` 或命中 `adminOnlyPrefixes`
  （`shared/pkg/auth/require_admin.go`）。新增后台端点必须在后者登记，否则匿名可达。
- 后台 API 走 `/admin/**` 前缀（见 `app/utils/adminPage.ts`），再映射到后端。

## 遗留问题

- 无真 bug 待修。
- 可探索：带权限的前端路由守卫（低权限账号访问未授权路由是否跳到 no-permission）、
  大数据量表格的分页/排序（UsersView/ManuscriptsView 用 el-table 的远程分页）。
- 后端侧（shared 不是独占，动前打招呼）：admin_api.go（work 服务）
  `/api/v1/admin/transcoders/*` 已是 /admin/ 前缀被门禁覆盖，无需再改。

## 不要做

- 不碰 `web/user/**`、`mobile/wap/**`（其他 agent 领地）
- 不 stage `.air/work.toml`、`external/cf-whisper-worker/.wrangler/**`
- 不删历史 `web/next-old/`、`web/teriteri-reference/`