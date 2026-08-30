# 网关路径放行和收藏夹问题修复计划

## 问题分析

### 1. 收藏夹数据获取失败的根本原因

**问题链路分析：**
1. 前端调用 `interactionApi.getFavoriteFolders()` 请求 `/api/manuscript/favorite/folders`
2. 网关 AuthFilter 检查白名单，发现 `/api/manuscript/favorite/folders` 不匹配任何白名单前缀
3. AuthFilter 要求 Token 认证，如果用户未登录则返回 401
4. 如果用户已登录，网关转发请求到 interaction 服务
5. interaction 服务的 `getFavoriteFolders` 方法从 `X-User-Id` header 获取用户ID
6. 如果 `userId` 为 null，返回空列表

**白名单匹配问题：**
- 当前白名单使用 `startsWith` 匹配
- `/api/favorite` 在白名单中，但前端调用的是 `/api/manuscript/favorite/folders`
- `/api/manuscript/favorite` 不在白名单中，所以被拦截

### 2. 网关路径放行问题汇总

需要放行但未放行的路径：
- `/api/manuscript/favorite/folders` - 获取收藏夹列表（需要登录后才能获取真实数据）
- `/api/manuscript/{id}/status` - 获取互动状态
- `/api/manuscript/{id}/favorite/folders` - 获取稿件的收藏夹
- 其他互动相关接口

### 3. 前端错误处理问题

前端在获取收藏夹失败时没有显示具体错误信息，导致看起来"没报错但没数据"。

## 修复方案

### 步骤 1: 修复 AuthFilter 白名单配置

**文件：** `mybilibili-gateway/src/main/java/com/mybilibili/gateway/filter/AuthFilter.java`

**修改内容：**
- 移除 `/api/favorite` 和 `/api/collection` 等模糊路径
- 添加明确的公开访问路径（不需要登录的接口）
- 保持需要登录的接口不在白名单中，让它们正常走认证流程

**需要公开访问的接口（白名单）：**
- 用户相关：登录、注册、用户信息查询
- 视频相关：推荐、热门、分类、视频详情、视频列表
- 搜索相关：搜索、热搜
- 分类相关：分类列表
- 轮播图：banner
- 静态资源：uploads

**需要登录的接口（不在白名单）：**
- 收藏夹操作：`/api/manuscript/favorite/folders`
- 点赞、投币、收藏：`/api/manuscript/{id}/like` 等
- 历史记录：`/api/watch-history`
- 消息：`/api/message`
- 关注：`/api/follow`

### 步骤 2: 检查并修复网关路由配置

**文件：** `mybilibili-gateway/src/main/resources/application.yml`

**确认路由规则：**
- `/api/manuscript/favorite/folders` → interaction 服务
- `/api/manuscript/{id}/status` → interaction 服务
- 其他互动相关路由

### 步骤 3: 修复前端收藏夹组件

**文件：** `mybilibili-web/src/layouts/components/dropdowns/FavoriteDropdown.vue`

**修改内容：**
- 改进错误处理，显示具体的错误信息
- 如果返回 401，显示"请先登录"提示
- 如果返回空数据，显示"暂无收藏夹"

### 步骤 4: 检查 interaction 服务

**文件：** `mybilibili-interaction/src/main/java/com/mybilibili/interaction/controller/InteractionController.java`

**确认：**
- `X-User-Id` header 正确获取
- 返回正确的响应格式

## 实施步骤

1. **修改 AuthFilter.java** - 更新白名单配置
2. **重启网关服务** - 应用配置更改
3. **测试收藏夹接口** - 验证登录后能正确获取数据
4. **测试其他接口** - 确保所有功能正常

## 测试验证

1. 未登录状态访问收藏夹 → 应返回 401
2. 登录后访问收藏夹 → 应返回收藏夹列表
3. 首页视频列表 → 应正常显示
4. 视频详情页 → 应正常显示
5. 历史记录页 → 登录后应正常显示
