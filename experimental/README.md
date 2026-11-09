# experimental/ — 实验性 / 存档区

> ⚠️ **此目录不参与构建、CI、部署，不再维护。** 仅供历史留存与原版对照，勿在此新增功能。

本目录集中存放已停止主线开发、避免干扰主工程的客户端/前端代码。它们**不**在 `go.work`、`Makefile`、`scripts/`、`.github/workflows/ci.yml`、`deploy/` 的任何构建或部署路径中。

| 目录 | 来源 | 说明 | 状态 |
|---|---|---|---|
| `mobile-flutter/` | `mobile/flutter` | Flutter 客户端（Bilibili 风格视频平台） | **已废弃** — 不再维护 |
| `mobile-ns/` | `mobile/ns` | NativeScript + Vue 跨端客户端 | **已废弃** — 不再维护 |
| `web-next-old/` | `web/next-old` | 旧版用户端（Nuxt）历史快照 | **存档** — 仅历史留存 |
| `web-teriteri-reference/` | `web/teriteri-reference` | 原版 teriteri 前端对照 | **存档** — 仅原版对照 |

## 当前活跃端（不在本目录）

- 用户端 Web：`web/user`（Nuxt 4，活跃）
- 管理后台：`web/admin`（Vite + Vue，活跃）
- 移动 H5：`mobile/wap`（活跃）
- Android WebView 壳：`mobile/webview`

## 说明

- 此移动为纯目录重组，不改变任何构建 / CI / 部署行为。
- 如需恢复某端为活跃开发，将其移回原位并接入 `scripts/frontend.sh` / `Makefile` / CI。
