<template>
  <div class="comment-manager">
    <!-- 主要选择栏 -->
    <div class="content-tabs">
      <el-tabs v-model="mainTab" type="card" @tab-change="handleTabChange">
        <el-tab-pane label="全部评论" name="all"></el-tab-pane>
        <el-tab-pane label="收到的评论" name="comment"></el-tab-pane>
        <el-tab-pane label="收到的回复" name="reply"></el-tab-pane>
      </el-tabs>
    </div>

    <!-- 次级选择栏 -->
    <div class="comment-filters">
      <div class="filter-row">
        <el-input
          v-model="keyword"
          placeholder="搜索评论内容"
          size="small"
          clearable
          class="keyword-input"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button type="primary" size="small" @click="handleSearch">搜索</el-button>
      </div>
    </div>

    <!-- 评论列表：卡片式列表（对齐B站互动管理列表层级） -->
    <div class="comment-list" v-loading="loading">
      <div v-if="list.length" class="comment-cards">
        <div v-for="item in list" :key="item.id" class="comment-card">
          <div class="cc-head">
            <el-avatar :size="32" :src="item.userAvatar || defaultAvatar"></el-avatar>
            <div class="cc-user">
              <span class="username">{{ item.userName }}</span>
              <span class="cc-type-pill" :class="item.commentType === 'reply' ? 'pill-reply' : 'pill-comment'">
                {{ item.commentType === 'reply' ? '回复' : '评论' }}
              </span>
            </div>
            <span class="cc-date">{{ formatTime(item.createTime) }}</span>
          </div>
          <div class="cc-body">
            <span v-if="item.commentType === 'reply' && item.replyToUserName" class="reply-to">
              回复 @{{ item.replyToUserName }}：
            </span>
            <span class="cc-content">{{ item.content }}</span>
          </div>
          <div class="cc-foot">
            <span class="cc-like">
              <el-icon><CaretRight /></el-icon>
              <span>{{ item.likeCount }}</span>
              <em>获赞</em>
            </span>
            <span class="cc-manuscript" @click="goManuscript(item.manuscriptId)">
              {{ item.manuscriptTitle || ('稿件#' + item.manuscriptId) }}
            </span>
            <span class="cc-actions">
              <button class="delete-btn" @click="item.commentType === 'reply' ? handleDeleteReply(item) : handleDeleteComment(item)">
                {{ item.commentType === 'reply' ? '删除回复' : '删除' }}
              </button>
            </span>
          </div>
        </div>
      </div>
      <el-empty v-else-if="!loading" description="暂无评论" />
    </div>

    <!-- 分页导航栏 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        @size-change="handlePageChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CaretRight } from '@element-plus/icons-vue'
import { creatorApi } from '@/api/creator'

const defaultAvatar = 'https://i0.hdslb.com/bfs/face/3378829f555891d2d5a4537e10264593a1d076b1.jpg@50w_50h_1c_1s_!web-avatar-nav.avif'

interface CommentItem {
  id: number
  manuscriptId: number
  manuscriptTitle?: string
  manuscriptCover?: string
  userName: string
  userAvatar?: string
  content: string
  likeCount: number
  replyCount: number
  status: number
  createTime: string
  liked: boolean
  commentType: 'comment' | 'reply'
  parentCommentId?: number
  replyToUserName?: string
}

const mainTab = ref('all')
const list = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const loading = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getComments({
      page: page.value,
      size: pageSize.value,
      commentType: mainTab.value === 'all' ? 'all' : mainTab.value,
      keyword: keyword.value || undefined,
    })
    if (res.code === 200) {
      list.value = (res.data?.list || []).map((item: any) => ({
        ...item,
        userAvatar: item.userAvatar || defaultAvatar,
      }))
      total.value = res.data?.total || 0
    } else {
      ElMessage.error(res.message || '获取评论失败')
    }
  } catch (e) {
    ElMessage.error('获取评论失败')
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  page.value = 1
  loadData()
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const handleDeleteComment = (item: CommentItem) => {
  ElMessageBox.confirm(`确定删除该评论吗？\n「${item.content}」`, '删除评论', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const res: any = await creatorApi.deleteComment(item.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const handleDeleteReply = (item: CommentItem) => {
  ElMessageBox.confirm(`确定删除该回复吗？\n「${item.content}」`, '删除回复', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const res: any = await creatorApi.deleteReply(item.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const formatTime = (t: string) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const goManuscript = (id: number) => {
  if (!id) return
  window.open(`/video/${id}`, '_blank')
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.comment-manager {
  padding: 12px 0;
}

.content-tabs {
  margin-bottom: 8px;
}

.comment-filters {
  padding: 8px 0 16px;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.keyword-input {
  width: 280px;
}

/* 评论卡片列表：独立白卡 radius4px 灰底隔开（对齐B站互动管理） */
.comment-cards {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.comment-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 16px 20px;
  transition: box-shadow .2s, border-color .2s;
}

.comment-card:hover {
  border-color: #d0d3d8;
  box-shadow: 0 2px 12px rgba(0, 0, 0, .04);
}

.cc-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cc-user {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  font-size: 14px;
  color: #212121;
  font-weight: 500;
}

.cc-type-pill {
  padding: 0 8px;
  font-size: 12px;
  line-height: 18px;
  border-radius: 10px;
}

.pill-comment {
  color: #00a1d6;
  background: rgba(0, 161, 214, .1);
}

.pill-reply {
  color: #61666d;
  background: rgba(97, 102, 109, .1);
}

.cc-date {
  margin-left: auto;
  font-size: 12px;
  color: #99a2aa;
}

.cc-body {
  margin-top: 12px;
  font-size: 14px;
  color: #18191c;
  line-height: 1.6;
}

.reply-to {
  color: #ff6699;
}

.cc-foot {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #f0f1f4;
}

.cc-like {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #99a2aa;
}

.cc-like .el-icon {
  font-size: 12px;
}

.cc-manuscript {
  font-size: 12px;
  color: #99a2aa;
  cursor: pointer;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cc-manuscript:hover {
  color: #ff6699;
}

.cc-actions {
  margin-left: auto;
}

.delete-btn {
  height: 28px;
  padding: 0 12px;
  font-size: 12px;
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 2px;
  cursor: pointer;
  transition: color .2s, border-color .2s;
}

.delete-btn:hover {
  color: #f56c6c;
  border-color: #f56c6c;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>