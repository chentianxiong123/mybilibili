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

    <!-- 评论列表 -->
    <div class="comment-list" v-loading="loading">
      <el-table :data="list" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column label="用户" width="180">
          <template #default="scope">
            <div class="user-cell">
              <el-avatar :size="28" :src="scope.row.userAvatar || defaultAvatar"></el-avatar>
              <span class="username">{{ scope.row.userName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="内容" min-width="300">
          <template #default="scope">
            <span v-if="scope.row.commentType === 'reply' && scope.row.replyToUserName" class="reply-to">
              回复 @{{ scope.row.replyToUserName }}：
            </span>
            <span>{{ scope.row.content }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.commentType === 'reply' ? 'info' : 'primary'" size="small">
              {{ scope.row.commentType === 'reply' ? '回复' : '评论' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="稿件" min-width="180">
          <template #default="scope">
            <span class="manuscript-title" @click="goManuscript(scope.row.manuscriptId)">
              {{ scope.row.manuscriptTitle || ('稿件#' + scope.row.manuscriptId) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="likeCount" label="获赞" width="80"></el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="scope">{{ formatTime(scope.row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button
              :type="scope.row.commentType === 'reply' ? 'danger' : 'danger'"
              size="small"
              @click="scope.row.commentType === 'reply' ? handleDeleteReply(scope.row) : handleDeleteComment(scope.row)"
            >
              {{ scope.row.commentType === 'reply' ? '删除回复' : '删除' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
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
.comment-manager {
  padding: 12px 0;
}

.content-tabs {
  margin-bottom: 8px;
}

.comment-filters {
  padding: 8px 0;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.keyword-input {
  width: 280px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.username {
  font-size: 13px;
  color: var(--text1);
}

.reply-to {
  color: var(--brand_pink);
}

.manuscript-title {
  color: var(--text2);
  cursor: pointer;
}

.manuscript-title:hover {
  color: var(--brand_pink);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>