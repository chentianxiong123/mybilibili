# 播放页关注按钮不同步数据库修复 Spec

## Why
稿件播放页的关注和取消关注按钮没有真正同步数据库，只有前端有效果。用户点击关注/取消关注后，页面状态会变化，但数据库中的关注关系没有实际更新。

## What Changes
- 修复前端 API 调用逻辑（已修复）
- **修复后端：获取稿件详情时没有正确传递当前用户ID，导致无法判断是否已关注**

## Impact
- Affected specs: 用户关注功能
- Affected code: 
  - `mybilibili-web/src/views/VideoView.vue` (前端已修复)
  - `mybilibili-video/.../ManuscriptController.java`
  - `mybilibili-video/.../ManuscriptServiceImpl.java`
  - `mybilibili-user/.../UserClient.java`

## 问题分析

### 问题1：前端API调用错误（已修复）
前端 `handleFollow` 函数始终传递 `true`，导致取消关注时调用了关注接口。

### 问题2：后端没有传递当前用户ID（核心问题）
前端初始化关注状态时，后端返回的 `following` 字段是错误的。

**当前代码分析**：
1. 前端请求：`GET /api/manuscript/{id}` 获取稿件详情
2. 后端 `ManuscriptController.getManuscriptById` **没有接收 HttpServletRequest**
3. 后端 `ManuscriptServiceImpl.buildManuscriptVO` 调用 `userClient.getUserById(vo.getUserId())` 获取作者信息
4. 关键代码（第240行和282行）：
   ```java
   uploader.setFollowing(user.getFollowingCount() != null && user.getFollowingCount() > 0);
   ```
   这行代码**错误地将"作者的关注数>0"当作"当前用户是否关注了作者"**！

### 正确的逻辑应该是
- 传递当前登录用户ID给后端
- 后端查询 `user_interaction` 表，确认当前用户是否关注了作者
- 返回正确的 `following` 状态

## ADDED Requirements

### Requirement: 后端正确返回关注状态
获取稿件详情时，系统应接收当前用户ID，并正确查询该用户是否关注了作者。

#### Scenario: 已登录用户访问稿件详情
- **WHEN** 已登录用户访问稿件详情页
- **THEN** 后端接收当前用户ID
- **AND** 后端查询数据库确认是否关注
- **AND** 返回正确的 `following` 状态

#### Scenario: 未登录用户访问稿件详情
- **WHEN** 未登录用户访问稿件详情页
- **THEN** 后端返回 `following: false`

## MODIFIED Requirements

### Requirement: ManuscriptController 接收用户ID
修改 `ManuscriptController.getManuscriptById` 方法，接收 `HttpServletRequest` 并传递当前用户ID。

### Requirement: ManuscriptService 添加用户参数
修改 `ManuscriptServiceImpl.getManuscriptWithVideos` 方法，接收当前用户ID参数。

### Requirement: UserClient 添加关注状态查询
添加新方法：`isFollowing(currentUserId, targetUserId)` 查询关注状态。

## 修复计划

1. **修改 ManuscriptController**：添加 HttpServletRequest 参数
2. **修改 ManuscriptService**：添加 currentUserId 参数
3. **修改 UserClient**：添加 `isFollowing` 方法
4. **修改 FollowService**：添加 `isFollowing` 接口

## REMOVED Requirements
无
