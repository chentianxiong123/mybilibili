<template>
  <div class="comment-manager">
    <!-- 页面级 tab：用户可见评论 / 待精选评论（对齐真页 new-link-top-item） -->
    <div class="top-tabs">
      <span
        class="top-tab"
        :class="{ active: mainTab === 'visible' }"
        @click="switchTab('visible')"
      >
        用户可见评论
      </span>
      <span
        class="top-tab"
        :class="{ active: mainTab === 'pending' }"
        @click="switchTab('pending')"
      >
        待精选评论
      </span>
    </div>

    <div class="comment-content">
      <!-- 次级 tab + 搜索（对齐真页 commet_header） -->
      <div class="commet-header">
        <div class="title-wrap">
          <span class="sub-tab" :class="{ active: subTab === 'all' }" @click="switchSubTab('all')">全部评论</span>
          <span class="sub-tab" :class="{ active: subTab === 'comment' }" @click="switchSubTab('comment')">仅评论</span>
          <span class="sub-tab" :class="{ active: subTab === 'reply' }" @click="switchSubTab('reply')">仅回复</span>
        </div>
        <el-input
          v-model="keyword"
          placeholder="搜索评论内容"
          clearable
          class="keyword-input"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 评论列表：对齐真页 comment-list-item -->
      <div class="comment-list" v-loading="loading">
        <div v-if="list.length" class="comment-cards">
          <div v-for="item in list" :key="item.id" class="comment-list-item">
            <a class="user-avatar" :href="`/space/${item.userId}`" @click.prevent="goSpace(item.userId)">
              <img v-if="item.userAvatar" :src="item.userAvatar" alt="" />
              <span v-else class="avatar-fallback">{{ (item.userName || '?').slice(0, 1) }}</span>
            </a>

            <div class="article-wrap">
              <a class="pic" :href="`/video/${item.manuscriptId}`" @click.prevent="goManuscript(item.manuscriptId)">
                <img v-if="item.manuscriptCover" :src="item.manuscriptCover" alt="" />
                <span v-else class="cover-fallback">暂无封面</span>
              </a>
              <a class="title ellipsis" :href="`/video/${item.manuscriptId}`" @click.prevent="goManuscript(item.manuscriptId)">
                <span class="name">{{ item.manuscriptTitle || ('稿件 #' + item.manuscriptId) }}</span>
              </a>
            </div>

            <div class="ci-title">
              <a class="ci-user" :href="`/space/${item.userId}`" @click.prevent="goSpace(item.userId)">{{ item.userName }}</a>
              <template v-if="item.commentType === 'reply'">
                <span class="ci-title-split">回复</span>
                <span v-if="item.replyToUserName" class="ci-title-split">@{{ item.replyToUserName }}</span>
              </template>
              <span v-else class="ci-title-tag">评论</span>
            </div>

            <a class="ci-content-wrap" :href="`/video/${item.manuscriptId}`" @click.prevent="goManuscript(item.manuscriptId)">
              <div class="ci-content">
                <span v-if="item.commentType === 'reply'" class="reply-prefix">回复 @{{ item.replyToUserName }}：</span>
                {{ item.content }}
              </div>
            </a>

            <div class="ci-action">
              <span class="date">{{ formatTime(item.createTime) }}</span>
              <span class="action like">
                <el-icon><CaretRight /></el-icon>
                <span class="num">{{ item.likeCount }}</span>
              </span>
              <span class="action reply" @click="openReply(item)">
                <el-icon><ChatDotRound /></el-icon>回复
              </span>
              <span class="action delete" @click="deleteOne(item)">
                <el-icon><Delete /></el-icon>删除
              </span>
            </div>
          </div>
        </div>
        <el-empty v-else-if="!loading" description="暂无评论" />
      </div>

      <BccPagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, CaretRight, ChatDotRound, Delete } from '@element-plus/icons-vue'
import { creatorApi } from '@/api/creator'
import BccPagination from '@/components/createCenter/BccPagination.vue'

interface CommentItem {
  id: number
  manuscriptId: number
  manuscriptTitle?: string
  manuscriptCover?: string
  userId: number
  userName: string
  userAvatar?: string
  content: string
  likeCount: number
  replyCount: number
  status: number
  createTime: string
  commentType: 'comment' | 'reply'
  replyToUserName?: string
}

const mainTab = ref('visible')
const subTab = ref('all')
const list = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const loading = ref(false)

const switchTab = (tab: string) => {
  mainTab.value = tab
}

const switchSubTab = (tab: string) => {
  subTab.value = tab
  page.value = 1
  loadData()
}

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getComments({
      page: page.value,
      size: pageSize.value,
      commentType: subTab.value === 'all' ? 'all' : subTab.value,
      keyword: keyword.value || undefined,
    })
    if (res.code === 200) {
      list.value = res.data?.list || []
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

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const deleteOne = (item: CommentItem) => {
  ElMessageBox.confirm(`确定删除该条内容吗？\n「${item.content}」`, item.commentType === 'reply' ? '删除回复' : '删除评论', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const res: any = item.commentType === 'reply'
        ? await creatorApi.deleteReply(item.id)
        : await creatorApi.deleteComment(item.id)
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

const openReply = (item: CommentItem) => {
  goManuscript(item.manuscriptId)
}

const formatTime = (t: string) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const goSpace = (uid: number) => {
  if (!uid) return
  window.open(`/space/${uid}`, '_blank')
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
  padding: 0 40px;
}

/* 页面级 tab */
.top-tabs {
  display: flex;
  align-items: center;
  height: 64px;
}

.top-tab {
  position: relative;
  font-size: 16px;
  color: #505050;
  cursor: pointer;
  margin-right: 40px;
  line-height: 64px;
}

.top-tab.active {
  color: var(--brand_pink, #00a1d6);
  font-weight: 700;
}

.top-tab.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2.5px;
  background: var(--brand_pink, #00a1d6);
}

.comment-content {
  padding: 0 40px 20px;
}

/* 次级 header：sub-tab + 搜索（对齐真页 commet_header） */
.commett-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 0 12px;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 32px;
}

.sub-tab {
  font-size: 14px;
  color: #505050;
  cursor: pointer;
  padding: 6px 0;
}

.sub-tab.active {
  color: var(--brand_pink, #00a1d6);
  font-weight: 500;
}

.keyword-input {
  width: 240px;
}

.keyword-input :deep(.el-input__wrapper) {
  border-radius: 4px;
  box-shadow: 0 0 0 1px #e3e5e7 inset;
}

/* 评论列表：对齐真页 comment-list-item（头像+封面+标题+内容+底部操作） */
.comment-cards {
  display: flex;
  flex-direction: column;
}

.comment-list-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 20px 0;
  border-bottom: 1px solid #f0f1f4;
}

.comment-list-item:last-child {
  border-bottom: none;
}

.user-avatar {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
  display: block;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: #fff;
  font-size: 16px;
  background: #c3c9d1;
}

.article-wrap {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 120px;
}

.article-wrap .pic {
  position: relative;
  width: 120px;
  height: 69px;
  border-radius: 6px;
  overflow: hidden;
  display: block;
  background: #e3e5e8;
}

.article-wrap .pic img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 10px;
  color: #99a2aa;
}

.article-wrap .title {
  display: block;
  font-size: 12px;
  color: #99a2aa;
  cursor: pointer;
  max-width: 120px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.article-wrap .title:hover {
  color: var(--brand_pink, #00a1d6);
}

.main-col {
  flex: 1;
  min-width: 0;
}

.ci-title {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.ci-user {
  font-size: 14px;
  color: #212121;
  cursor: pointer;
  white-space: nowrap;
}

.ci-user:hover {
  color: var(--brand_pink, #00a1d6);
}

.ci-title-split {
  font-size: 12px;
  color: #9499a0;
}

.ci-title-tag {
  padding: 1px 8px;
  font-size: 12px;
  line-height: 16px;
  color: #00a1d6;
  background: rgba(0, 161, 214, .1);
  border-radius: 8px;
}

.ci-content-wrap {
  flex: 1;
  min-width: 0;
  cursor: pointer;
  display: block;
}

.ci-content {
  margin-top: 10px;
  font-size: 14px;
  color: #18191c;
  line-height: 1.6;
  word-break: break-word;
}

.reply-prefix {
  color: var(--brand_pink, #00a1d6);
}

.ci-action {
  display: inline-flex;
  align-items: center;
  gap: 20px;
  margin-top: 12px;
  flex: 1;
}

.ci-action .date {
  font-size: 12px;
  color: #999aa0;
}

.ci-action .action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #505050;
  cursor: pointer;
}

.ci-action .action .el-icon {
  font-size: 14px;
}

.ci-action .action:hover {
  color: var(--brand_pink, #00a1d6);
}

.ci-action .like .num {
  color: #999aa0;
}

.ci-action .delete:hover {
  color: #f56c6c;
}
</style>