<template>
  <div class="comment-manage">
    <div class="content-tabs">
      <el-tabs v-model="activeTab" type="card" @tab-change="handleTabChange">
        <el-tab-pane label="全部" name="all"></el-tab-pane>
        <el-tab-pane label="评论" name="comment"></el-tab-pane>
        <el-tab-pane label="回复" name="reply"></el-tab-pane>
      </el-tabs>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-item">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索评论内容"
          size="small"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>
      <el-button type="primary" size="small" @click="handleSearch">搜索</el-button>
    </div>

    <!-- 评论列表 -->
    <div class="comment-list" v-loading="loading">
      <el-empty v-if="!loading && list.length === 0" description="暂无评论" />

      <div v-for="item in list" :key="item.id" class="comment-item">
        <div class="comment-head">
          <el-avatar :size="32" :src="item.userAvatar || defaultAvatar"></el-avatar>
          <div class="comment-meta">
            <div class="username">{{ item.userName }}</div>
            <div class="time">{{ formatTime(item.createTime) }}</div>
          </div>
          <el-tag v-if="item.commentType === 'reply'" size="small" type="info">回复</el-tag>
          <el-tag v-else size="small" type="primary">评论</el-tag>
        </div>
        <div class="comment-content">
          <template v-if="item.commentType === 'reply' && item.replyToUserName">
            <span class="reply-to">回复 @{{ item.replyToUserName }}：</span>
          </template>
          {{ item.content }}
        </div>
        <div class="comment-foot">
          <span class="manuscript-title" v-if="item.manuscriptTitle" @click="goManuscript(item.manuscriptId)">
            稿件：{{ item.manuscriptTitle }}
          </span>
          <span class="stats">
            <span class="stat"><el-icon><Star /></el-icon>{{ item.likeCount || 0 }}</span>
            <span class="stat" v-if="item.replyCount"><el-icon><ChatDotRound /></el-icon>{{ item.replyCount }}</span>
          </span>
          <span class="actions">
            <el-button
              v-if="item.commentType === 'reply'"
              type="danger"
              link
              size="small"
              @click="handleDeleteReply(item)"
            >删除回复</el-button>
            <el-button
              v-else
              type="danger"
              link
              size="small"
              @click="handleDeleteComment(item)"
            >删除评论</el-button>
          </span>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrap" v-if="total > pageSize">
        <el-pagination
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Star, ChatDotRound } from '@element-plus/icons-vue'
import { creatorApi } from '@/api/creator'

const defaultAvatar = 'https://i0.hdslb.com/bfs/face/3378829f555891d2d5a4537e10264593a1d076b1.jpg@50w_50h_1c_1s_!web-avatar-nav.avif'

interface CommentItem {
  id: number
  manuscriptId: number
  manuscriptTitle?: string
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

const activeTab = ref('all')
const list = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const filters = reactive<{ keyword: string }>({ keyword: '' })

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getComments({
      page: page.value,
      size: pageSize,
      commentType: activeTab.value === 'all' ? 'all' : activeTab.value,
      keyword: filters.keyword || undefined,
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

const handlePageChange = (p: number) => {
  page.value = p
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
  return t.replace('T', ' ').slice(0, 19)
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
.comment-manage {
  padding: 20px 24px;
}

.content-tabs {
  margin-bottom: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.filter-item {
  width: 280px;
}

.comment-list {
  min-height: 300px;
}

.comment-item {
  padding: 14px 0;
  border-bottom: 1px solid var(--graph_bg_thin, #f0f0f0);
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.comment-meta {
  flex: 1;
}

.username {
  font-size: 14px;
  color: var(--text1, #18191c);
  font-weight: 500;
}

.time {
  font-size: 12px;
  color: var(--text3, #9499a0);
  margin-top: 2px;
}

.comment-content {
  padding: 8px 0 4px 42px;
  font-size: 14px;
  color: var(--text1, #18191c);
  line-height: 1.6;
}

.reply-to {
  color: var(--brand_pink, #fb7299);
}

.comment-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-left: 42px;
  margin-top: 6px;
}

.manuscript-title {
  font-size: 12px;
  color: var(--text3, #9499a0);
  cursor: pointer;
}

.manuscript-title:hover {
  color: var(--brand_pink, #fb7299);
}

.stats {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: var(--text3, #9499a0);
}

.stat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 0;
}
</style>
