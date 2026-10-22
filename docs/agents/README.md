# docs/agents —— 当前协作文档（索引）

> 快照日期：2026-10-22
>
> 本目录只保留**当前有效**的文档。文档过时后整份移到
> `docs/archive/agents/<快照日期>/`，再按新情况写新的，**不原地追改历史**。
> 归档目录只读，仅作追溯。

## 当前文档

| 文件 | 领域 | 代码位置 | 快照 |
|---|---|---|---|
| `00-协作方案.md` | 三 agent 分工、边界、验收门禁 | — | 2026-10-22 |
| `web.md` | 桌面 Web 用户端（Nuxt 4，`ssr:false`） | `web/user/**` | 2026-10-22 |
| `wap.md` | 移动 H5（Vue3 + Vite） | `mobile/wap/**` | 2026-10-22 |
| `admin.md` | 管理后台（Vue3 + Element Plus） | `web/admin/**` | 2026-10-22 |

## 维护规则

1. **新事实写新文档**：领域大改时新建带日期的文档（如 `2026-11-01-search.md`），
   旧文档整份归档；只有小勘误才在当前文档上改，并更新顶部「快照日期」。
2. **归档方式**：
   ```bash
   mkdir -p docs/archive/agents/2026-10-22
   git mv docs/agents/web.md docs/archive/agents/2026-10-22/web.md
   # 然后写新的 docs/agents/web.md
   ```
3. **一次性记录**（排查、方案对比、评测、实验）放 `docs/archive/<domain>/`，
   不进 `docs/agents/`。
4. 本 README 只做索引，不写细节；细节放对应领域文档。

## 快速入口

- 启动 / 端口 / 账号：根 `README.md` 的「本地开发（Docker Compose）」
- 测试：`bash scripts/test.sh frontend` / `bash scripts/test.sh backend`
- 前端类型检查：
  `docker exec -w /app mybilibili-web bun node_modules/vue-tsc/bin/vue-tsc.js --noEmit`
- 历史文档：`docs/archive/README.md`
