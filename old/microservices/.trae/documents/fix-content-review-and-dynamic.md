# 内容审核功能修复与增强计划

## 问题分析

### 问题1：已下架内容查询不到
**原因**：前端传递 `status='REMOVED'`，后端逻辑需要正确转换
- Comments: `status=1` 表示已下架（整数）
- Replies: `status='REMOVED'` 表示已下架（字符串）

**当前逻辑**：
- 前端：`activeTab === 'removed'` → 传递 `status='REMOVED'`
- 后端：判断 `"NORMAL"` 或 `"REMOVED"` 转换为对应的值

**可能的Bug**：需要确认后端转换逻辑是否正确处理了 Comment 和 Reply 的不同状态值

### 问题2：动态评论未纳入管理
**当前状态**：动态评论（dynamic_comments 表）独立于普通评论管理

**目标**：将动态评论和回复也纳入内容审核中心统一管理

---

## 修复步骤

### Step 1: 检查并修复后端状态转换逻辑

**文件**：`CommentServiceImpl.java` 的 `getAllContent` 方法

确认：
- Comment: `status=1` 为已下架
- Reply: `status='REMOVED'` 为已下架

### Step 2: 添加动态评论管理接口

**新增内容类型**：`DYNAMIC_COMMENT` 和 `DYNAMIC_REPLY`

**后端修改**：
1. `DynamicCommentMapper.java` - 添加查询方法
2. `CommentServiceImpl.java` - 在 `getAllContent` 中添加动态评论查询逻辑
3. `deleteContent` 和 `restoreContent` 方法中添加对动态评论的处理

**前端修改**：
1. `contentReview.js` - 添加动态评论相关 API
2. `ContentReviewView.vue` - 添加动态评论类型筛选

### Step 3: 验证数据库状态值
确认 `comments`、`replies`、`dynamic_comments` 表中的 status 值是否符合预期

---

## 文件修改清单

### 后端
1. `CommentServiceImpl.java` - 添加动态评论查询逻辑
2. `DynamicCommentMapper.java` - 添加 selectAllDynamicComments 等方法

### 前端
1. `contentReview.js` - 添加动态评论 API
2. `ContentReviewView.vue` - 添加动态评论类型选项
