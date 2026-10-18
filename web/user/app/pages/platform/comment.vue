<template>
  <div class="comment-manager">
    <!-- 顶部 tab + 搜索（对齐真页 new-link-top-wrap + bar-search_wrap） -->
    <div class="link-top-wrap">
      <div class="link-top-container">
        <a
          class="link-top-item"
          :class="{ active: mainTab === 'visible' }"
          @click="switchTab('visible')"
        >
          <span>用户可见评论</span>
        </a>
        <a
          class="link-top-item"
          :class="{ active: mainTab === 'pending' }"
          @click="switchTab('pending')"
        >
          <span>待精选评论</span>
        </a>
      </div>
      <div class="bar-search-wrap">
        <input
          v-model="keyword"
          class="bar-search-input"
          placeholder="搜索视频评论"
          @keyup.enter="handleSearch"
        />
        <i class="bar-search-icon" @click="handleSearch"></i>
      </div>
    </div>

    <div class="comment-content">
      <!-- 次级 header：sub-tab + 下拉（对齐真页 commet_header） -->
      <header class="commet-header">
        <div class="title-wrap">
          <a :class="{ active: contentType === 'video' }" @click="setContentType('video')">视频评论</a>
          <a :class="{ active: contentType === 'article' }" @click="setContentType('article')">专栏评论</a>
          <a :class="{ active: contentType === 'audio' }" @click="setContentType('audio')">音频评论</a>
        </div>
        <div class="select-wrap">
          <div class="filter-select" v-click-outside="closeStatusDropdown">
            <a class="filter-btn" @click.stop="statusDropdownOpen = !statusDropdownOpen">
              <span>{{ statusLabel }}</span>
              <span class="icon-arrow" :class="{ open: statusDropdownOpen }"></span>
            </a>
            <ul v-show="statusDropdownOpen" class="filter-menu">
              <li @click="setStatus('all')"><span :class="{ selected: statusFilter === 'all' }">全部评论</span></li>
              <li @click="setStatus('hidden')"><span :class="{ selected: statusFilter === 'hidden' }">隐藏评论</span></li>
            </ul>
          </div>
          <div class="filter-select" v-click-outside="closeVideoDropdown">
            <a class="filter-btn" @click.stop="videoDropdownOpen = !videoDropdownOpen">
              <span class="ellipsis">{{ videoLabel }}</span>
              <span class="icon-arrow" :class="{ open: videoDropdownOpen }"></span>
            </a>
            <ul v-show="videoDropdownOpen" class="filter-menu video-menu">
              <li @click="setVideo(0)"><span :class="{ selected: videoFilter === 0 }">全部视频</span></li>
              <li v-for="m in manuscripts" :key="m.id" @click="setVideo(m.id)">
                <span :class="{ selected: videoFilter === m.id }" class="ellipsis">{{ m.title }}</span>
              </li>
            </ul>
          </div>
        </div>
      </header>

      <!-- 操作栏（对齐真页 operate_wrap） -->
      <div class="operate-wrap">
        <div class="operate-left">
          <label class="bcc-checkbox" @click.prevent="toggleAll">
            <input type="checkbox" :checked="allSelected" @change="toggleAll" />
            <span class="checkbox-mask"></span>
            <span class="checkbox-label">全选</span>
          </label>
          <button class="bcc-button" :disabled="!selectedIds.length" @click="batchReport">举报</button>
          <button class="bcc-button del" :disabled="!selectedIds.length" @click="batchDelete">删除</button>
        </div>
        <div class="operate-right">
          <div class="operate-txt" :class="{ active: sortBy === 'time' }" @click="setSort('time')">最近发布</div>
          <div class="operate-txt" :class="{ active: sortBy === 'like' }" @click="setSort('like')">点赞最多</div>
          <div class="operate-txt" :class="{ active: sortBy === 'reply' }" @click="setSort('reply')">回复最多</div>
        </div>
      </div>

      <!-- 评论列表（对齐真页 comment-list-item 绝对定位布局） -->
      <section class="section-list-wrap" v-loading="loading">
        <template v-if="list.length">
          <div v-for="item in list" :key="item.id" class="comment-list-item">
            <div class="check-box" @click.stop="toggleSelect(item.id)">
              <input type="checkbox" :checked="selectedIds.includes(item.id)" />
              <span class="checkbox-mask"></span>
            </div>

            <a class="user-avatar" @click="goSpace(item.userId)">
              <img v-if="item.userAvatar" :src="item.userAvatar" alt="" />
            </a>

            <div class="ci-title">
              <a class="ci-user" @click="goSpace(item.userId)">{{ item.userName }}</a>
              <template v-if="item.commentType === 'reply'">
                <span class="ci-title-split">回复</span>
                <a class="parent-user" @click="goSpace(item.replyToUserId)">@{{ item.replyToUserName }}</a>
                <span>的评论</span>
                <span class="show-parent"><span>查看评论</span></span>
              </template>
            </div>

            <a class="ci-content" @click="goManuscript(item.manuscriptId)">
              <span v-if="item.commentType === 'reply'" class="reply-prefix">回复 @{{ item.replyToUserName }}：</span>
              {{ item.content }}
            </a>

            <div class="ci-action">
              <span class="date">{{ formatTime(item.createTime) }}</span>
              <span class="like action">
                <svg class="like-svg" viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linejoin="round" stroke-linecap="round">
                  <path d="M4 14V7l3-5c.6 0 1.2.5 1.2 1.2L7.6 6h3.7c.9 0 1.5.8 1.3 1.6l-1.2 5.5c-.1.7-.7 1.4-1.5 1.4H4z"/>
                  <rect x="1.5" y="7" width="2.5" height="7" rx=".5"/>
                </svg>
              </span>
              <span class="reply action" @click="openReply(item)">回复</span>
              <span class="report action" @click="reportOne(item)">举报</span>
              <span class="delete action" @click="deleteOne(item)">删除</span>
            </div>

            <div class="article-wrap">
              <a class="pic" @click="goManuscript(item.manuscriptId)">
                <img v-if="item.manuscriptCover" :src="item.manuscriptCover" alt="" />
              </a>
              <a class="title ellipsis" @click="goManuscript(item.manuscriptId)">
                <span class="name">{{ item.manuscriptTitle || ('稿件 #' + item.manuscriptId) }}</span>
                <span class="show-all">该视频全部评论</span>
              </a>
            </div>
          </div>
        </template>
        <el-empty v-else-if="!loading" description="暂无评论" />
      </section>

      <!-- 提示 + 分页 -->
      <div class="tips">仅展示最近的50000条评论</div>
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
import { ref, computed, onMounted, Directive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
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
  replyToUserId?: number
}

const mainTab = ref('visible')
const contentType = ref<'video' | 'article' | 'audio'>('video')
const statusFilter = ref<'all' | 'hidden'>('all')
const videoFilter = ref(0)
const sortBy = ref<'time' | 'like' | 'reply'>('time')
const keyword = ref('')
const list = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const manuscripts = ref<{ id: number; title: string }[]>([])
const selectedIds = ref<number[]>([])

const statusDropdownOpen = ref(false)
const videoDropdownOpen = ref(false)

const statusLabel = computed(() => statusFilter.value === 'hidden' ? '隐藏评论' : '全部评论')
const videoLabel = computed(() => {
  if (!videoFilter.value) return '全部视频'
  const m = manuscripts.value.find(x => x.id === videoFilter.value)
  return m ? m.title : '全部视频'
})
const allSelected = computed(() => list.value.length > 0 && selectedIds.value.length === list.value.length)

const vClickOutside = {
  beforeMount(el: any, binding: any) {
    el._clickOutside = (e: Event) => { if (!el.contains(e.target)) binding.value() }
    document.addEventListener('click', el._clickOutside)
  },
  unmounted(el: any) {
    document.removeEventListener('click', el._clickOutside)
  }
} as Directive

const closeStatusDropdown = () => { statusDropdownOpen.value = false }
const closeVideoDropdown = () => { videoDropdownOpen.value = false }

const switchTab = (tab: string) => {
  mainTab.value = tab
  page.value = 1
  loadData()
}

const setContentType = (t: 'video' | 'article' | 'audio') => {
  contentType.value = t
  page.value = 1
  loadData()
}

const setStatus = (s: 'all' | 'hidden') => {
  statusDropdownOpen.value = false
  statusFilter.value = s
  page.value = 1
  loadData()
}

const setVideo = (id: number) => {
  videoDropdownOpen.value = false
  videoFilter.value = id
  page.value = 1
  loadData()
}

const setSort = (s: 'time' | 'like' | 'reply') => {
  sortBy.value = s
  page.value = 1
  loadData()
}

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getComments({
      page: page.value,
      size: pageSize,
      commentType: contentType.value === 'video' ? 'all' : contentType.value,
      keyword: keyword.value || undefined,
    })
    if (res.code === 200) {
      list.value = res.data?.list || []
      total.value = res.data?.total || 0
      selectedIds.value = []
      extractManuscripts()
    } else {
      ElMessage.error(res.message || '获取评论失败')
    }
  } catch (e) {
    ElMessage.error('获取评论失败')
  } finally {
    loading.value = false
  }
}

const extractManuscripts = () => {
  const map = new Map<number, string>()
  list.value.forEach(item => {
    if (item.manuscriptId && !map.has(item.manuscriptId)) {
      map.set(item.manuscriptId, item.manuscriptTitle || '')
    }
  })
  manuscripts.value = Array.from(map.entries()).map(([id, title]) => ({ id, title }))
}

const handleSearch = () => { page.value = 1; loadData() }
const handlePageChange = () => { loadData() }

const toggleSelect = (id: number) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

const toggleAll = () => {
  if (allSelected.value) selectedIds.value = []
  else selectedIds.value = list.value.map(i => i.id)
}

const batchReport = () => {
  ElMessage.info('举报功能开发中')
}

const batchDelete = async () => {
  if (!selectedIds.value.length) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedIds.value.length} 条内容吗？`, '批量删除', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    })
    for (const id of selectedIds.value) {
      const item = list.value.find(i => i.id === id)
      if (item) {
        await (item.commentType === 'reply' ? creatorApi.deleteReply(id) : creatorApi.deleteComment(id))
      }
    }
    ElMessage.success('删除成功')
    loadData()
  } catch (e) { /* cancel */ }
}

const deleteOne = (item: CommentItem) => {
  ElMessageBox.confirm(`确定删除该条内容吗？\n「${item.content}」`, item.commentType === 'reply' ? '删除回复' : '删除评论', {
    confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning',
  }).then(async () => {
    try {
      const res: any = item.commentType === 'reply'
        ? await creatorApi.deleteReply(item.id)
        : await creatorApi.deleteComment(item.id)
      if (res.code === 200) { ElMessage.success('删除成功'); loadData() }
      else ElMessage.error(res.message || '删除失败')
    } catch (e) { ElMessage.error('删除失败') }
  }).catch(() => {})
}

const reportOne = (_item: CommentItem) => { ElMessage.info('举报功能开发中') }
const openReply = (item: CommentItem) => { goManuscript(item.manuscriptId) }

const formatTime = (t: string) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const goSpace = (uid?: number) => { if (uid) window.open(`/space/${uid}`, '_blank') }
const goManuscript = (id: number) => { if (id) window.open(`/video/${id}`, '_blank') }

onMounted(() => { loadData() })
</script>

<style scoped>
.comment-manager {
  padding: 0;
}

/* ===== 顶部 tab + 搜索 ===== */
.link-top-wrap {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0;
}

.link-top-container {
  display: flex;
  height: 64px;
  box-sizing: border-box;
  border-bottom: 1px solid #e7e7e7;
}

.link-top-item {
  font-size: 16px;
  color: #505050;
  font-weight: 400;
  padding: 26px 0 0;
  margin-left: 40px;
  box-sizing: border-box;
  cursor: pointer;
  border-bottom: 3px solid transparent;
  transition: color .2s;
  display: block;
}

.link-top-item:first-child {
  margin-left: 40px;
}


.link-top-item.active {
  color: #00a1d6;
  font-weight: 700;
  border-bottom-color: #00a1d6;
}

.link-top-item:hover {
  color: #00a1d6;
}

.bar-search-wrap {
  position: relative;
  width: 230px;
  margin-top: 20px;
  margin-right: 40px;
}

.bar-search-input {
  width: 100%;
  height: 32px;
  padding: 0 32px 0 12px;
  font-size: 13px;
  color: #222;
  border: 1px solid #e3e5e7;
  border-radius: 4px;
  outline: none;
  background: #fff;
  box-sizing: border-box;
  transition: border-color .2s;
}

.bar-search-input::placeholder {
  color: #99a2aa;
}

.bar-search-input:focus {
  border-color: #00a1d6;
}

.bar-search-icon {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.bar-search-icon::before {
  content: '';
  display: block;
  width: 12px;
  height: 12px;
  border: 1.5px solid #99a2aa;
  border-radius: 50%;
}

.bar-search-icon::after {
  content: '';
  display: block;
  position: absolute;
  right: -3px;
  bottom: -1px;
  width: 6px;
  height: 1.5px;
  background: #99a2aa;
  transform: rotate(45deg);
  transform-origin: left center;
}

/* ===== 主体内容 ===== */
.comment-content {
  padding: 0 40px;
}

/* ===== 次级 header ===== */
.commet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 11px 0 0;
  padding: 0;
  height: 33px;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 0;
}

.title-wrap a {
  font-size: 14px;
  color: #505050;
  font-weight: 500;
  line-height: 20px;
  padding: 6.5px 16px;
  cursor: pointer;
  border-radius: 4px;
  transition: color .2s;
}

.title-wrap a.active {
  color: #00a1d6;
  font-weight: 500;
  background: #f6f7f8;
}

.title-wrap a:hover {
  color: #00a1d6;
}

/* 下拉筛选 */
.select-wrap {
  display: flex;
  gap: 12px;
}

.filter-select {
  position: relative;
  width: 220px;
}

.filter-btn {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 32px;
  padding: 0 12px;
  font-size: 13px;
  color: #61666d;
  border: 1px solid #e3e5e7;
  border-radius: 4px;
  cursor: pointer;
  background: #fff;
  box-sizing: border-box;
  transition: border-color .2s;
  gap: 8px;
}

.filter-btn:hover {
  border-color: #c9ccd2;
}

.filter-btn .ellipsis {
  max-width: 110px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-arrow {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-right: 1.5px solid #99a2aa;
  border-bottom: 1.5px solid #99a2aa;
  transform: rotate(45deg) translateY(-2px);
  transition: transform .2s;
  flex-shrink: 0;
}

.icon-arrow.open {
  transform: rotate(225deg) translateY(-2px);
}

.filter-menu {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  min-width: 100%;
  background: #fff;
  border: 1px solid #ccd0d7;
  border-radius: 4px;
  box-shadow: 2px 2px 4px 0 rgba(0, 0, 0, .2);
  padding: 10px 0;
  z-index: 7;
  list-style: none;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
}

.filter-menu li {
  padding: 6px 16px;
  font-size: 13px;
  color: #222;
  cursor: pointer;
  transition: color .2s;
  line-height: 1.4;
}

.filter-menu li:hover {
  color: #00a1d6;
}

.filter-menu li span.selected {
  color: #00a1d6;
}

.video-menu {
  max-width: 260px;
}

.video-menu li .ellipsis {
  display: block;
  max-width: 228px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ===== 操作栏 ===== */
.operate-wrap {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 24px 0 0;
  padding: 0;
  height: 26px;
}

.operate-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bcc-checkbox {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  gap: 6px;
  margin-right: 24px;
  user-select: none;
}

.bcc-checkbox input {
  display: none;
}

.checkbox-mask {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 1px solid #e7e7e7;
  border-radius: 2px;
  background: #fff;
  position: relative;
  transition: all .15s;
  box-sizing: border-box;
}

.bcc-checkbox input:checked + .checkbox-mask {
  background: #00a1d6;
  border-color: #00a1d6;
}

.bcc-checkbox input:checked + .checkbox-mask::after {
  content: '';
  position: absolute;
  left: 4px;
  top: 1px;
  width: 5px;
  height: 9px;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(45deg);
}

.checkbox-label {
  font-size: 14px;
  color: #61666d;
}

.bcc-button {
  height: 26px;
  padding: 5px 16px;
  font-size: 14px;
  color: #61666d;
  border: 1px solid #e7e7e7;
  border-radius: 4px;
  background: #f4f4f4;
  cursor: pointer;
  transition: all .2s;
  line-height: 1;
}

.bcc-button:disabled {
  color: #c0c0c0;
  background: #f4f4f4;
  border-color: #e7e7e7;
  cursor: not-allowed;
}

.bcc-button:not(:disabled):hover {
  color: #00a1d6;
  border-color: #00a1d6;
  background: #fff;
}

.bcc-button.del:not(:disabled):hover {
  color: #f56c6c;
  border-color: #f56c6c;
}

.operate-right {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-right: -16px;
}

.operate-txt {
  position: relative;
  width: 75px;
  box-sizing: border-box;
  font-size: 14px;
  letter-spacing: .5px;
  color: #222;
  cursor: pointer;
  transition: color .2s;
}

.operate-txt:not(:last-child)::after {
  content: "";
  position: absolute;
  right: 0;
  top: 2px;
  width: 1px;
  height: 16px;
  background: #e7e7e7;
}

.operate-txt.active {
  color: #00a1d6;
}

.operate-txt:hover {
  color: #00a1d6;
}

/* ===== 评论列表 ===== */
.section-list-wrap {
  min-height: 200px;
}

.comment-list-item {
  position: relative;
  margin-top: 24px;
  padding: 0 175px 24px 96px;
  border-bottom: 1px solid #e7e7e7;
  box-sizing: border-box;
}

.comment-list-item:first-child {
  margin-top: 40px;
}

.comment-list-item:last-child {
  border-bottom: none;
}

.check-box {
  position: absolute;
  left: 0;
  top: 13px;
  cursor: pointer;
}

.check-box input {
  display: none;
}

.user-avatar {
  position: absolute;
  left: 40px;
  top: 2px;
  width: 40px;
  height: 40px;
  cursor: pointer;
  display: block;
}

.user-avatar img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

/* 右侧稿件缩略图 */
.article-wrap {
  position: absolute;
  right: 0;
  top: 0;
  width: 100px;
  flex-shrink: 0;
}

.article-wrap .pic {
  display: block;
  width: 100px;
  height: 60px;
  border-radius: 4px;
  overflow: hidden;
  background: #e3e5e8;
  cursor: pointer;
}

.article-wrap .pic img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.article-wrap .title {
  display: block;
  font-size: 12px;
  color: #00a1d6;
  margin-top: 4px;
  cursor: pointer;
  line-height: 1.4;
}

.article-wrap .title .name {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.article-wrap .title:hover {
  color: #00a1d6;
}

/* 标题行 */
.ci-title {
  font-size: 14px;
  color: #0f0f0f;
  margin-bottom: 12px;
  line-height: 18px;
}

.ci-user {
  color: #0f0f0f;
  cursor: pointer;
  text-decoration: none;
}

.ci-user:hover {
  color: #00a1d6;
}

.ci-title-split {
  color: #61666d;
  margin: 0 2px;
}

.parent-user {
  color: #00a1d6;
  cursor: pointer;
  text-decoration: none;
}

.parent-user:hover {
  color: #0077a3;
}

.show-parent {
  color: #00a1d6;
  cursor: pointer;
  margin-left: 8px;
}

/* 评论内容 */
.ci-content {
  display: block;
  font-size: 14px;
  color: #212121;
  line-height: 18px;
  cursor: pointer;
  text-decoration: none;
  margin-bottom: 16px;
  word-break: break-word;
}

.ci-content:hover {
  color: #00a1d6;
}

.reply-prefix {
  color: #00a1d6;
}

/* 操作行 */
.ci-action {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 19px;
}

.ci-action .date {
  font-size: 14px;
  color: #999;
}

.ci-action .action {
  font-size: 14px;
  color: #61666d;
  cursor: pointer;
  transition: color .2s;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.ci-action .action:hover {
  color: #00a1d6;
}

.ci-action .delete:hover {
  color: #f56c6c;
}

.like-svg {
  display: block;
}

/* ===== 提示 + 分页 ===== */
.tips {
  margin-top: 24px;
  font-size: 14px;
  color: #999;
}

.fans-footer,
:deep(.bcc-pagination-container) {
  margin-top: 24px;
}
</style>
