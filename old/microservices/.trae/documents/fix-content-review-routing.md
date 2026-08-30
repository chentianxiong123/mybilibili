# 修复内容审核路由问题计划

## 问题分析

**症状**：
- 前端请求 `GET /api/admin/content-review/pending` 返回 404
- 内容审核页面无法正常工作

**根本原因**：网关路由顺序错误

在 Spring Cloud Gateway 中，路由按配置顺序匹配。更具体的路径（如 `/api/admin/content-review/**`）应该放在更通用的路径（如 `/api/admin/**`）**之前**，否则通用路径会先匹配到，导致请求被转发到错误的服务。

**当前路由顺序**（错误）：
1. `/api/admin/**` → user-service（会匹配所有 `/api/admin/...` 请求）
2. `/api/admin/content-review/**` → comment-service（永远不会被匹配到）

**正确的路由顺序**：
1. `/api/admin/content-review/**` → comment-service（更具体的路径先匹配）
2. `/api/admin/**` → user-service（通用的路径后匹配）

## 修复步骤

### 1. 编辑网关配置文件
文件路径：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-gateway\src\main\resources\application.yml`

### 2. 调整路由顺序
将 `# 内容审核接口` 路由块移动到 `# 管理端接口（复用user-service）` 路由块**之前**。

### 3. 验证修复
修改后，确保：
- `/api/admin/content-review/**` → 转发到 comment-service
- `/api/admin/users/**` → 转发到 user-service
- `/api/admin/manuscripts/**` → 转发到 video-service

## 注意事项
- 只调整路由顺序，不删除或修改任何路由配置
- 确保所有现有路由保持不变
- 调整顺序后不影响其他功能的正常路由
