# 视频互动功能修复计划

## 问题分析

### 现象
- 点赞和收藏按钮点击无反应
- 点赞/收藏状态没有显示（按钮没有变蓝）
- 控制台不报错，但功能不工作

### 根本原因

经过代码分析，发现**前端传参错误**：

#### 1. 互动状态获取 - 参数类型错误

**前端代码** ([VideoView.vue:829-831](file:///d:/files/mybilibili-next/mybilibili-web/src/views/VideoView.vue#L829-831)):
```javascript
if (token && videoId.value) {
  const statusResponse = await interactionApi.getInteractionStatus(videoId.value)
}
```

**API 定义** ([index.js:197](file:///d:/files/mybilibili-next/mybilibili-web/src/api/index.js#L197)):
```javascript
getInteractionStatus: (manuscriptId) => api.get(`/manuscript/${manuscriptId}/status`)
```

**后端期望** ([VideoInteractionServiceImpl.java:245](file:///d:/files/mybilibili-next/mybilibili-cloud/mybilibili-interaction/src/main/java/com/mybilibili/interaction/service/impl/VideoInteractionServiceImpl.java#L245)):
```java
public VideoInteractionStatus getInteractionStatus(Integer userId, Integer manuscriptId)
```

**问题**: 前端传的是 `videoId`（视频ID），但 API 和后端期望的是 `manuscriptId`（稿件ID）！

#### 2. 变量定义说明
- `currentManuscriptId` = 从路由参数获取的**稿件ID**（正确）
- `videoId` = 从视频数据中获取的**具体视频ID**（错误）

#### 3. 影响范围
- `loadInteractionStatus()` 函数使用错误的 `videoId.value`
- 导致互动状态获取失败，按钮状态不更新

---

## 修复计划

### 步骤1：修复 VideoView.vue 中 loadInteractionStatus 函数

**文件**: `mybilibili-web/src/views/VideoView.vue`

**修改内容**:
```javascript
// 修改前 (错误)
if (token && videoId.value) {
  const statusResponse = await interactionApi.getInteractionStatus(videoId.value)
}

// 修改后 (正确)
if (token && currentManuscriptId.value) {
  const statusResponse = await interactionApi.getInteractionStatus(currentManuscriptId.value)
}
```

### 步骤2：验证点赞功能

确认点赞功能使用正确的参数：
- 当前已使用 `currentManuscriptId.value` ✅（之前已修复）

### 步骤3：验证收藏功能

确认收藏功能使用正确的参数：
- 检查 `handleFavorite` 和 `confirmFavorite` 函数

### 步骤4：测试验证

1. 刷新视频页面
2. 检查控制台日志确认 API 调用正确
3. 测试点赞按钮 - 应该变蓝并显示成功消息
4. 测试收藏按钮 - 应该弹出收藏夹对话框
5. 验证状态持久化 - 刷新页面后状态保持

---

## 预期结果

- ✅ 点赞按钮点击后有反应，变蓝色
- ✅ 收藏按钮点击后弹出收藏夹对话框
- ✅ 已点赞/已收藏的视频刷新后状态保持
- ✅ 控制台无错误信息
