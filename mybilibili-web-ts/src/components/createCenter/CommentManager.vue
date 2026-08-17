<template>
  <div class="comment-manager">
    <div class="comment-management">
      <!-- 主标签页和搜索框 -->
      <div class="main-tabs-with-search">
        <!-- 视频评论蓝色字样 -->
        <div class="video-comment-label">视频评论</div>

        <!-- 搜索框 -->
        <div class="main-search">
          <el-input
            v-model="commentSearchText"
            placeholder="搜索视频评论"
            size="small"
            style="width: 200px;"
            clearable
          >
            <template #append>
              <el-button size="small" @click="searchComments"><el-icon><Search /></el-icon></el-button>
            </template>
          </el-input>
        </div>
      </div>
      
      <!-- 搜索和过滤区域 -->
      <div class="comment-filter-bar">
        <div class="left-section">
          <!-- 子标签页：视频评论、专栏评论、音频评论 -->
          <div class="sub-tabs">
            <el-radio-group v-model="activeCommentSubTab" size="small">
              <el-radio-button value="video">视频评论</el-radio-button>

              <el-radio-button value="audio">音频评论</el-radio-button>
            </el-radio-group>
          </div>
        </div>
        
        <div class="right-section">
          <!-- 评论类型和视频筛选 -->
          <div class="filter-dropdowns">
            <el-select v-model="commentTypeFilter" placeholder="全部评论" size="small" style="min-width: 120px; margin-right: 10px;">
              <el-option label="评论+回复" value="all"></el-option>
              <el-option label="仅评论" value="comment"></el-option>
              <el-option label="仅回复" value="reply"></el-option>
            </el-select>
            
            <el-select v-model="videoFilter" placeholder="全部视频" size="small" style="min-width: 120px;">
              <el-option label="全部视频" value="all"></el-option>
              <el-option
                v-for="video in videoList"
                :key="video.id"
                :label="video.title"
                :value="video.id"
              ></el-option>
            </el-select>
          </div>
        </div>
      </div>
      
      <!-- 操作栏 -->
      <div class="comment-actions">
        <div class="action-buttons">
          <el-button size="small" plain @click="handleSelectAll(true)">全选</el-button>
          <el-button size="small" plain @click="handleBatchDelete">删除</el-button>
        </div>
        
        <!-- 排序选项 -->
        <div class="sort-options">
          <el-radio-group v-model="commentSortBy" size="small">
            <el-radio-button value="latest">最近发布</el-radio-button>
            <el-radio-button value="likes">点赞最多</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      
      <!-- 评论列表 -->
      <div class="comment-list">
        <div
          v-for="comment in pagedComments"
          :key="comment.id"
          class="comment-item"
        >
          <!-- 复选框 -->
          <div class="comment-checkbox">
            <el-checkbox v-model="comment.selected"></el-checkbox>
          </div>
          
          <!-- 评论主体：头像、用户名、内容、操作 -->
          <div class="comment-main">
            <!-- 头像和用户名 -->
            <div class="comment-header">
              <el-avatar :size="40" :src="comment.avatar"></el-avatar>
              <span class="username">{{ comment.username }}</span>
              <el-tag v-if="comment.commentType === 'reply'" size="small" type="warning" style="margin-left: 8px;">回复</el-tag>
              <span v-if="comment.commentType === 'reply' && comment.replyToUserName" class="reply-to-hint">
                回复给 <b>{{ comment.replyToUserName }}</b>
              </span>
            </div>

            <!-- 评论内容 -->
            <div class="comment-content">
              {{ comment.content }}
            </div>

            <!-- 评论时间和操作 -->
            <div class="comment-meta">
              <span class="comment-time">{{ formatDate(comment.time) }}</span>
              <el-button size="small" plain :type="comment.liked ? 'primary' : 'default'" @click="handleLikeComment(comment)">
                <el-icon><CircleCheck /></el-icon>{{ comment.likeCount || 0 }}
              </el-button>
              <el-button size="small" plain @click="openReplyDialog(comment)">
                <el-icon><ChatDotRound /></el-icon>回复
              </el-button>

              <!-- 删除按钮（鼠标悬停显示） -->
              <div class="comment-actions-hover">
                <el-button size="small" plain type="danger" @click="handleDeleteComment(comment)">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </div>
            </div>
          </div>
          
          <!-- 视频缩略图 -->
          <div class="comment-right">
            <div class="video-thumbnail" v-if="comment.videoThumbnail">
              <img loading="lazy" decoding="async" :src="comment.videoThumbnail" alt="视频缩略图">
              <div class="video-title">{{ comment.videoTitle }}</div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 分页 -->
      <div class="comment-pagination">
        <div class="custom-pagination">
          <el-button 
            v-for="(page, index) in visiblePages" 
            :key="index"
            :type="page === commentCurrentPage ? 'primary' : 'default'"
            :plain="page !== commentCurrentPage"
            :disabled="page === '...'"
            @click="typeof page === 'number' && (commentCurrentPage = page)"
            size="small"
          >
            {{ page }}
          </el-button>
          
          <el-button 
            v-if="commentCurrentPage < totalPages" 
            @click="commentCurrentPage++"
            size="small"
          >
            下一页
          </el-button>
          
          <div class="pagination-info">
            共{{ totalPages }}页 / {{ totalComments }}个
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="replyDialogVisible" title="回复评论" width="500px">
      <el-input
        v-model="replyContent"
        type="textarea"
        :rows="4"
        placeholder="请输入回复内容"
        maxlength="500"
        show-word-limit
      />
      <template #footer>
        <el-button @click="replyDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleReplyComment">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi, manuscriptApi } from '@/api/creator'
import { commentApi } from '@/api/client'
import { Search, CircleCheck, ChatDotRound, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

// 评论管理相关数据
const activeCommentMainTab = ref('visible')
const activeCommentSubTab = ref('video')
const commentTypeFilter = ref('all')
const videoFilter = ref('all')
const commentSearchText = ref('')
const commentSortBy = ref('latest')
const commentCurrentPage = ref(1)
const commentPageSize = ref(10)

const comments = ref([])
const commentLoading = ref(false)
const commentRawList = ref([])
const totalComments = computed(() => filteredComments.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(totalComments.value / commentPageSize.value)))
const videoList = ref([])
const replyDialogVisible = ref(false)
const replyCommentId = ref(null)
const replyContent = ref('')
const replyToUserId = ref(null)

const normalizeSearchText = (value) => (value || '').toString().trim().toLowerCase()

const filteredComments = computed(() => {
  const keyword = normalizeSearchText(commentSearchText.value)
  if (!keyword) {
    return commentRawList.value
  }
  return commentRawList.value.filter(comment => {
    const fields = [
      comment.username,
      comment.content,
      comment.videoTitle,
      comment.replyToUserName,
      comment.commentType === 'reply' ? '回复' : '评论'
    ]
    return fields.some(field => normalizeSearchText(field).includes(keyword))
  })
})

const pagedComments = computed(() => {
  const start = (commentCurrentPage.value - 1) * commentPageSize.value
  return filteredComments.value.slice(start, start + commentPageSize.value)
})

const handleDeleteComment = async (comment) => {
  const isReply = comment.commentType === 'reply'
  const label = isReply ? '回复' : '评论'
  try {
    await ElMessageBox.confirm(`确定要删除这条${label}吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = isReply
      ? await creatorApi.deleteReply(comment.id)
      : await creatorApi.deleteComment(comment.id)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchComments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error(`删除${label}失败:`, error)
      ElMessage.error('删除失败')
    }
  }
}

// 点赞/取消点赞评论
const handleLikeComment = async (comment) => {
  const isReply = comment.commentType === 'reply'
  try {
    if (comment.liked) {
      const res = isReply
        ? await commentApi.unlikeReply(comment.id)
        : await commentApi.unlikeComment(comment.id)
      if (res.code === 200) {
        comment.likeCount = Math.max(0, (comment.likeCount || 0) - 1)
        comment.liked = false
      } else {
        ElMessage.error(res.message || '取消点赞失败')
      }
    } else {
      const res = isReply
        ? await commentApi.likeReply(comment.id)
        : await commentApi.likeComment(comment.id)
      if (res.code === 200) {
        comment.likeCount = (comment.likeCount || 0) + 1
        comment.liked = true
      } else {
        ElMessage.error(res.message || '点赞失败')
      }
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    ElMessage.error('操作失败')
  }
}

const fetchComments = async () => {
  commentLoading.value = true
  try {
    const params = {
      page: 1,
      size: 1000,
      manuscriptId: videoFilter.value === 'all' ? undefined : videoFilter.value,
      sort: commentSortBy.value,
      commentType: commentTypeFilter.value
    }
    const res = await creatorApi.getComments(params)
    if (res.code === 200 && res.data) {
      commentRawList.value = res.data.list.map(item => ({
        id: item.id,
        selected: false,
        username: item.userName || '未知用户',
        avatar: item.userAvatar || '',
        content: item.content,
        time: item.createTime,
        videoThumbnail: item.manuscriptCover || '',
        videoTitle: item.manuscriptTitle || '',
        likeCount: item.likeCount || 0,
        replyCount: item.replyCount || 0,
        liked: item.liked || false,
        userId: item.userId,
        manuscriptId: item.manuscriptId,
        commentType: item.commentType || 'comment',
        parentCommentId: item.parentCommentId,
        replyToUserName: item.replyToUserName
      }))
      comments.value = pagedComments.value
      if (commentCurrentPage.value > totalPages.value) {
        commentCurrentPage.value = totalPages.value
      }
    }
  } catch (error) {
    console.error('获取评论列表失败:', error)
  } finally {
    commentLoading.value = false
  }
}

const fetchVideoList = async () => {
  try {
    const res = await manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    if (res.code === 200 && res.data) {
      videoList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取视频列表失败:', error)
  }
}

const openReplyDialog = (comment) => {
  // If clicking reply on a reply, use the parent comment ID
  const targetCommentId = comment.commentType === 'reply' ? comment.parentCommentId : comment.id
  if (!targetCommentId) {
    ElMessage.warning('无法回复该内容')
    return
  }
  replyCommentId.value = targetCommentId
  replyToUserId.value = comment.userId
  replyContent.value = ''
  replyDialogVisible.value = true
}

const handleReplyComment = async () => {
  if (!replyContent.value.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  try {
    const res = await creatorApi.replyComment(replyCommentId.value, replyContent.value, replyToUserId.value)
    if (res.code === 200) {
      ElMessage.success('回复成功')
      replyDialogVisible.value = false
      fetchComments()
    }
  } catch (error) {
    console.error('回复评论失败:', error)
  }
}

const handleSelectAll = (select) => {
  pagedComments.value.forEach(comment => {
    comment.selected = select
  })
}

const handleBatchDelete = async () => {
  const selectedComments = commentRawList.value.filter(c => c.selected)
  if (selectedComments.length === 0) {
    ElMessage.warning('请选择要删除的评论')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedComments.length} 条内容吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await Promise.all(selectedComments.map(c =>
      c.commentType === 'reply'
        ? creatorApi.deleteReply(c.id)
        : creatorApi.deleteComment(c.id)
    ))
    ElMessage.success('批量删除成功')
    fetchComments()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
    }
  }
}

watch([commentTypeFilter, videoFilter, commentSortBy], () => {
  if (commentCurrentPage.value !== 1) {
    commentCurrentPage.value = 1
  }
  fetchComments()
})

watch(commentCurrentPage, () => {
  comments.value = pagedComments.value
})

watch(commentSearchText, () => {
  if (commentCurrentPage.value !== 1) {
    commentCurrentPage.value = 1
  }
  comments.value = pagedComments.value
})

// 搜索评论
const searchComments = () => {
  if (commentCurrentPage.value !== 1) {
    commentCurrentPage.value = 1
  }
  comments.value = pagedComments.value
}

watch(totalPages, (newTotal) => {
  if (commentCurrentPage.value > newTotal) {
    commentCurrentPage.value = newTotal
  }
})

// 可见页码列表
const visiblePages = computed(() => {
  const pages = []
  const current = commentCurrentPage.value
  const total = totalPages.value
  
  // 总是显示第一页
  pages.push(1)
  
  // 如果当前页码大于3，显示省略号
  if (current > 3) {
    pages.push('...')
  }
  
  // 显示当前页码附近的页码
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  // 如果当前页码小于总页码-2，显示省略号
  if (current < total - 2) {
    pages.push('...')
  }
  
  // 如果总页码大于1，显示最后一页
  if (total > 1) {
    pages.push(total)
  }
  
  return pages
})

// 监听当前激活菜单变化，加载评论数据
watch(
  () => props.active,
  (newVal) => {
    if (newVal) {
      fetchComments()
      fetchVideoList()
    }
  },
  { immediate: true }
)

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
</script>

<style scoped>
.comment-manager {
  width: 100%;
}

.comment-management {
  width: 100%;
}

/* 主标签页样式 */
.main-tabs {
  margin-bottom: 20px;
}

.main-tabs .el-radio-group {
  font-size: 16px;
}

/* 主标签页和搜索框组合样式 */
.main-tabs-with-search {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

/* 主搜索框样式 */
.main-search {
  display: flex;
  align-items: center;
}

/* 评论过滤栏样式 */
.comment-filter-bar {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

.left-section {
  display: flex;
  align-items: center;
  margin-left: 20px;
}

.right-section {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

/* 子标签页样式 */
.sub-tabs {
  margin-right: 20px;
}

/* 筛选下拉框样式 */
.filter-dropdowns {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 视频评论蓝色字样样式 */
.video-comment-label {
  color: #1890ff;
  font-weight: 500;
  font-size: 14px;
}

/* 评论操作栏样式 */
.comment-actions {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

/* 操作按钮组样式 */
.action-buttons {
  display: flex;
  align-items: center;
  gap: 5px; /* 紧凑排布 */
}

.action-buttons .el-button {
  margin-right: 0; /* 移除默认右边距，使用gap控制间距 */
}

/* 排序选项样式 */
.sort-options {
  display: flex;
  align-items: center;
  margin-left: auto; /* 靠右对齐 */
}

/* 评论列表样式 */
.comment-list {
  margin-bottom: 20px;
}

/* 评论项样式 */
.comment-item {
  display: flex;
  align-items: flex-start;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
  position: relative;
  gap: 15px;
}

.comment-item:hover .comment-actions-hover {
  display: flex;
}

/* 复选框样式 */
.comment-checkbox {
  margin-top: 5px;
  flex-shrink: 0;
}

/* 评论主体样式 */
.comment-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-self: flex-start;
}

/* 头像和用户名样式 */
.comment-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 5px;
}

.comment-header .el-avatar {
  margin-right: 10px;
  flex-shrink: 0;
}

/* 用户名样式 */
.username {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
  line-height: 1;
}

.reply-to-hint {
  font-size: 12px;
  color: #909399;
  margin-left: 4px;
}

.reply-to-hint b {
  color: #409eff;
}

/* 评论内容样式 */
.comment-content {
  line-height: 1.5;
  color: #303133;
  font-size: 14px;
  margin-left: 50px;
  margin-top: 0;
  margin-bottom: 0;
}

/* 评论元信息样式 */
.comment-meta {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-top: 5px;
  margin-left: 50px;
}

.comment-time {
  font-size: 12px;
  color: #909399;
}

.comment-meta .el-button {
  min-width: auto;
  padding: 0 8px;
  height: 24px;
  line-height: 24px;
  font-size: 12px;
}

/* 举报和删除按钮样式（鼠标悬停显示） */
.comment-actions-hover {
  display: none;
  align-items: center;
  gap: 5px;
  margin-left: 15px;
}

.comment-actions-hover .el-button {
  min-width: auto;
  padding: 0 8px;
  height: 24px;
  line-height: 24px;
  font-size: 12px;
}

.comment-right {
  margin-left: 20px;
  align-self: flex-start;
}

/* 视频缩略图样式 */
.video-thumbnail {
  width: 120px;
  text-align: center;
  margin-top: 0;
}

.video-thumbnail img {
  width: 100%;
  height: 68px;
  object-fit: cover;
  border-radius: 4px;
  display: block;
}

.video-title {
  font-size: 12px;
  color: #606266;
  margin-top: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 评论分页样式 */
.comment-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #e0e0e0;
}

.comment-total {
  font-size: 12px;
  color: #909399;
}

.pagination-control {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

/* 自定义分页样式 */
.custom-pagination {
  display: flex;
  align-items: center;
  gap: 5px;
}

.custom-pagination .el-button {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 32px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.custom-pagination .el-button--primary {
  background-color: #1890ff;
  border-color: #1890ff;
  color: #fff;
}

.custom-pagination .el-button--primary.is-plain {
  background-color: #fff;
  border-color: #d9d9d9;
  color: #1890ff;
}

/* 省略号样式 */
.custom-pagination .ellipsis {
  min-width: 32px;
  height: 32px;
  line-height: 32px;
  text-align: center;
  font-size: 14px;
  color: #606266;
}

/* 分页信息样式 */
.pagination-info {
  margin-left: 15px;
  font-size: 14px;
  color: #606266;
}

/* 隐藏次级选择栏和状态选择栏 */
.sub-tabs,
.status-tabs {
  display: none;
}
</style>