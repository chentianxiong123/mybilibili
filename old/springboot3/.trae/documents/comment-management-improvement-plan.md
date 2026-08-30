# 评论管理页面改进计划

## 一、问题分析

评论管理页面 (`/create-center/interaction-comment`) 存在以下问题：
1. 前端API文件缺少 `getComments` 方法定义
2. 前端数据映射可能与后端返回的字段不匹配
3. 删除评论和回复评论功能需要完善

## 二、现有资源

### 后端API
- `CreatorCommentController.java` 已实现：
  - `GET /creator/comments` - 获取评论列表
  - `DELETE /creator/comments/{commentId}` - 删除评论
  - `POST /creator/comments/{commentId}/reply` - 回复评论

### 后端返回数据结构 (CreatorCommentVO)
```java
- id: Integer
- manuscriptId: Integer
- manuscriptTitle: String
- manuscriptCover: String
- userId: Integer
- userName: String
- userAvatar: String
- content: String
- likeCount: Integer
- replyCount: Integer
- createTime: Date
- liked: boolean
```

### 前端期望数据结构
```javascript
- id
- username (需要映射为 userName)
- avatar (需要映射为 userAvatar)
- content
- time (需要映射为 createTime)
- videoThumbnail (需要映射为 manuscriptCover)
- videoTitle (需要映射为 manuscriptTitle)
- likeCount
- replyCount
- selected
```

## 三、实现步骤

### 步骤1: 添加前端API方法

在 `creator.js` 的 `creatorApi` 中添加：
```javascript
getComments: (params) => api.get('/creator/comments', { params }),
deleteComment: (commentId) => api.delete(`/creator/comments/${commentId}`),
replyComment: (commentId, data) => api.post(`/creator/comments/${commentId}/reply`, null, { params: data })
```

### 步骤2: 修复数据映射

在 `fetchComments` 函数中修复字段映射：
```javascript
comments.value = res.data.list.map(item => ({
  id: item.id,
  selected: false,
  username: item.userName,
  avatar: item.userAvatar,
  content: item.content,
  time: item.createTime,
  videoThumbnail: item.manuscriptCover,
  videoTitle: item.manuscriptTitle,
  likeCount: item.likeCount,
  replyCount: item.replyCount,
  liked: item.liked
}))
```

### 步骤3: 实现删除评论功能

修改 `handleBatchDelete` 函数：
```javascript
const handleBatchDelete = async () => {
  const selectedComments = comments.value.filter(c => c.selected)
  if (selectedComments.length === 0) {
    ElMessage.warning('请先选择要删除的评论')
    return
  }
  
  try {
    for (const comment of selectedComments) {
      await creatorApi.deleteComment(comment.id)
    }
    ElMessage.success('删除成功')
    fetchComments()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}
```

### 步骤4: 实现回复评论功能

添加回复评论对话框和函数：
```javascript
const replyDialogVisible = ref(false)
const replyCommentId = ref(null)
const replyContent = ref('')

const openReplyDialog = (comment) => {
  replyCommentId.value = comment.id
  replyContent.value = ''
  replyDialogVisible.value = true
}

const handleReply = async () => {
  if (!replyContent.value.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  
  try {
    await creatorApi.replyComment(replyCommentId.value, { content: replyContent.value })
    ElMessage.success('回复成功')
    replyDialogVisible.value = false
    fetchComments()
  } catch (error) {
    ElMessage.error('回复失败')
  }
}
```

### 步骤5: 添加回复对话框UI

在模板中添加回复对话框：
```html
<el-dialog v-model="replyDialogVisible" title="回复评论">
  <el-input v-model="replyContent" type="textarea" :rows="4" placeholder="请输入回复内容" />
  <template #footer>
    <el-button @click="replyDialogVisible = false">取消</el-button>
    <el-button type="primary" @click="handleReply">发送</el-button>
  </template>
</el-dialog>
```

### 步骤6: 绑定回复按钮点击事件

修改回复按钮：
```html
<el-button size="small" plain @click="openReplyDialog(comment)">
  <el-icon><ChatDotRound /></el-icon>回复
</el-button>
```

## 四、详细实现清单

### 前端任务
1. [ ] 在 `creator.js` 添加 `getComments`、`deleteComment`、`replyComment` API方法
2. [ ] 修复 `fetchComments` 中的字段映射
3. [ ] 实现 `handleBatchDelete` 批量删除功能
4. [ ] 添加回复评论对话框和相关状态
5. [ ] 实现 `openReplyDialog` 和 `handleReply` 函数
6. [ ] 绑定回复按钮点击事件

## 五、注意事项

1. 删除评论需要二次确认
2. 回复内容不能为空
3. 需要处理API错误情况
4. 批量操作需要显示进度
