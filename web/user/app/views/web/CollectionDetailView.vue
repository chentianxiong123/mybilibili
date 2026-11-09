<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { collectionApi } from '@/api/collection.ts'

const route = useRoute()
const router = useRouter()

// 合集ID
const collectionId = ref(route.params.id)

// 当前用户ID
const { user: currentUser } = useAuth()
const currentUserId = computed(() => currentUser.value?.id)

// 加载状态
const loading = ref(false)

// 合集信息
const collection = ref({
  id: null,
  name: '',
  description: '',
  coverUrl: '',
  isPublic: true,
  userId: null,
  userName: '',
  userAvatar: '',
  videoCount: 0,
  viewCount: 0,
  createTime: '',
  updateTime: ''
})

// 稿件列表
const manuscripts = ref([])

// 当前播放的稿件索引
const currentIndex = ref(0)

// 分页
const pagination = ref({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 排序方式
const sortBy = ref('default') // default-默认, newest-最新, oldest-最早

// 判断是否是自己的合集
const isOwnCollection = computed(() => {
  return currentUserId.value && collection.value.userId &&
         String(currentUserId.value) === String(collection.value.userId)
})

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

// 格式化播放量（万）
const formatView = (n) => {
  const v = Number(n) || 0
  if (v >= 10000) return (v / 10000).toFixed(1).replace(/\.0$/, '') + '万'
  return String(v)
}

// 获取默认封面
const getDefaultCover = () => {
  return 'https://picsum.photos/id/1025/400/225'
}

// 合集 API 返回 snake_case，归一化到 camelCase
const normalizeCollection = (d) => ({
  id: d.id,
  name: d.title || d.name || '',
  description: d.description || '',
  coverUrl: d.cover_url || d.coverUrl || '',
  isPublic: d.status === 1 || d.isPublic === true,
  userId: d.user_id || d.userId || null,
  userName: d.user_name || d.userName || '',
  userAvatar: d.user_avatar || d.userAvatar || '',
  videoCount: d.manuscript_count || d.videoCount || 0,
  viewCount: d.view_count || d.viewCount || 0,
  createTime: d.created_at || d.createTime || '',
  updateTime: d.updated_at || d.updateTime || ''
})

// 加载合集详情
const loadCollectionDetail = async () => {
  loading.value = true
  try {
    const response = await collectionApi.getCollectionById(collectionId.value)
    if (response.code === 200) {
      collection.value = normalizeCollection(response.data)
    }
  } catch (error) {
    console.error('获取合集详情失败:', error)
    ElMessage.error('获取合集详情失败')
  } finally {
    loading.value = false
  }
}

// 加载合集中的稿件
const loadManuscripts = async () => {
  loading.value = true
  try {
    const response = await collectionApi.getCollectionManuscripts(
      collectionId.value,
      pagination.value.currentPage,
      pagination.value.pageSize
    )
    if (response.code === 200) {
      manuscripts.value = response.data?.list || []
      pagination.value.total = response.data?.total || 0
    }
  } catch (error) {
    console.error('获取稿件列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 播放稿件
const playManuscript = (manuscript, index) => {
  currentIndex.value = index
  // 跳转到视频播放页面
  if (manuscript.id) {
    router.push(`/manuscript/${manuscript.id}`)
  }
}

// 播放全部
const playAll = () => {
  if (manuscripts.value.length > 0) {
    playManuscript(manuscripts.value[0], 0)
  }
}

// 返回上一页
const goBack = () => {
  router.back()
}

// 跳转到用户主页
const goToUserProfile = () => {
  if (collection.value.userId) {
    router.push(`/profile/${collection.value.userId}`)
  }
}

// 跳转到编辑页面
const goToEdit = () => {
  router.push(`/collection/${collectionId.value}/edit`)
}

// 处理排序变化
const handleSortChange = (value) => {
  sortBy.value = value
  // 根据排序方式重新排序稿件
  if (value === 'newest') {
    manuscripts.value.sort((a, b) => new Date(b.uploadTime || b.createTime) - new Date(a.uploadTime || a.createTime))
  } else if (value === 'oldest') {
    manuscripts.value.sort((a, b) => new Date(a.uploadTime || a.createTime) - new Date(b.uploadTime || b.createTime))
  } else {
    // 默认排序，重新加载
    loadManuscripts()
  }
}

onMounted(() => {
  loadCollectionDetail()
  loadManuscripts()
})
</script>

<template>
  <div class="space-list-details">
    <!-- 合集头部：返回 + 标题 + 统计 + 操作（播放全部 / 排序） -->
    <div class="list-header">
      <div class="title">
        <div class="back" @click="goBack">
          <i class="iconfont icon-zuojiantou"></i>
        </div>
        <div class="name">{{ collection.name }}</div>
      </div>
      <div class="subtitle">
        <div class="meta">
          <div class="stats">
            合集&nbsp;&nbsp; {{ collection.videoCount }}个视频 &nbsp; {{ formatDate(collection.updateTime) }}更新
          </div>
        </div>
      </div>
      <div class="actions">
        <div class="left">
          <div class="playall">
            <button class="vui_button action-btn playall-btn" @click="playAll">
              <i class="iconfont icon-bofang"></i>
              <span>播放全部</span>
            </button>
          </div>
        </div>
        <div class="right">
          <div class="sort-mode">
            <select class="sort-select" :value="sortBy" @change="handleSortChange($event.target.value)">
              <option value="default">默认排序</option>
              <option value="newest">最新发布</option>
              <option value="oldest">最早发布</option>
            </select>
            <i class="iconfont icon-xiajiantou sort-mode__icon"></i>
          </div>
        </div>
      </div>
    </div>

    <!-- 视频网格 -->
    <div class="list-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 空状态 -->
      <div v-else-if="manuscripts.length === 0" class="empty-state">
        <el-empty description="暂无视频">
          <button v-if="isOwnCollection" class="vui_button action-btn" @click="goToEdit">添加视频</button>
        </el-empty>
      </div>

      <!-- 视频卡片 -->
      <div
        v-for="(manuscript, index) in manuscripts"
        :key="manuscript.id"
        class="list-video-item"
        @click="playManuscript(manuscript, index)"
      >
        <div class="bili-video-card">
          <div class="bili-video-card__wrap">
            <div class="bili-video-card__cover">
              <a class="bili-cover-card">
                <div class="bili-cover-card__thumbnail">
                  <img :src="manuscript.coverUrl || getDefaultCover()" :alt="manuscript.title" loading="lazy" decoding="async" />
                </div>
                <div class="bili-cover-card__stats">
                  <div class="bili-cover-card__stat">
                    <i class="iconfont icon-bofangfangkuai"></i>
                    <span>{{ formatView(manuscript.viewCount) }}</span>
                  </div>
                  <div class="bili-cover-card__stat">
                    <span>{{ formatDuration(manuscript.duration) }}</span>
                  </div>
                </div>
              </a>
              <div class="bili-card-watch-later">
                <div class="bili-card-watch-later__btn">
                  <i class="iconfont icon-bofang"></i>
                </div>
                <span class="bili-card-watch-later__tip">稍后再看</span>
              </div>
            </div>
            <div class="bili-video-card__details">
              <div class="bili-video-card__title" :title="manuscript.title">
                <a>{{ manuscript.title }}</a>
              </div>
              <div class="bili-video-card__subtitle">
                <span>{{ formatDate(manuscript.uploadTime) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ===== 对齐 space.bilibili.com/{uid}/lists/{lid} 合集详情页 =====
   参考值来自真实 B 站 getComputedStyle 抓取（1920×1080 / dpr1）：
   detailWrapper 1640@x140 / listHeader 1640x128 / name 24-500-34 #18191c
   listContent grid gap 20px 16px / card 260x212 / cover 260x146 r6 bg #e3e5e7
   title 13-500-18 #18191c / subtitle 12-400-16 #9499a0 / stats 12-12 白
   playall-btn 100x34 r8 */
.space-list-details {
  width: 1640px;
  margin: 0 auto;
  padding: 20px 0 60px;
}

/* 头部 */
.list-header {
  display: block;
  margin-bottom: 22px;
}

.list-header .title {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
}

.list-header .back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  cursor: pointer;
  color: #18191c;
}

.list-header .back:hover {
  background-color: rgba(0, 0, 0, 0.06);
}

.list-header .back .iconfont {
  font-size: 20px;
  line-height: 1;
}

.list-header .name {
  font-size: 24px;
  font-weight: 500;
  line-height: 34px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-header .subtitle {
  margin-top: 20px;
  height: 20px;
}

.list-header .stats {
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
  color: #9499a0;
}

.list-header .actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 20px;
  height: 34px;
}

/* 播放全部按钮（B站 .playall-btn 100x34 r8） */
.vui_button.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  padding: 0 12px;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  background-color: #fff;
  color: #18191c;
  font-size: 14px;
  line-height: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.vui_button.action-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
}

.vui_button.action-btn .iconfont {
  font-size: 12px;
  margin-right: 4px;
}

/* 排序（B站 .sort-mode 80x16 #61666d） */
.sort-mode {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #61666d;
  font-size: 14px;
  line-height: 16px;
  cursor: pointer;
}

.sort-mode .sort-select {
  appearance: none;
  -webkit-appearance: none;
  border: none;
  background: transparent;
  color: #61666d;
  font-size: 14px;
  line-height: 16px;
  cursor: pointer;
  padding-right: 2px;
  outline: none;
}

.sort-mode .sort-mode__icon {
  font-size: 12px;
}

/* 视频网格（B站 .list-content grid gap 20px 16px，6 列，卡 260x212） */
.list-content {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 20px 16px;
  padding-bottom: 60px;
}

.list-video-item {
  cursor: pointer;
}

.bili-video-card {
  width: 100%;
}

.bili-video-card__wrap {
  width: 100%;
}

/* 封面（260x146 r6 底 #e3e5e7） */
.bili-video-card__cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 6px;
  overflow: hidden;
  background-color: #e3e5e7;
}

.bili-cover-card {
  display: block;
  width: 100%;
  height: 100%;
  text-decoration: none;
}

.bili-cover-card__thumbnail {
  width: 100%;
  height: 100%;
}

.bili-cover-card__thumbnail img {
  display: inline-block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
}

/* 封面底部统计（260x34 白字 12-12，圆角 0 0 6 6） */
.bili-cover-card__stats {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  height: 34px;
  padding: 14px 8px 4px 8px;
  border-radius: 0 0 6px 6px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.6) 100%);
  color: #fff;
}

.bili-cover-card__stat {
  display: flex;
  align-items: center;
  margin-right: 10px;
  font-size: 12px;
  font-weight: 400;
  line-height: 12px;
  color: #fff;
}

.bili-cover-card__stat .iconfont {
  font-size: 12px;
  margin-right: 2px;
}

.bili-cover-card__stat:last-child {
  margin-right: 0;
}

/* 稍后再看（28x28 圆角，hover 显现） */
.bili-card-watch-later {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  align-items: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background-color: rgba(0, 0, 0, 0.6);
  color: #fff;
  opacity: 0;
  transition: opacity 0.3s;
}

.list-video-item:hover .bili-card-watch-later {
  opacity: 1;
}

.bili-card-watch-later__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.bili-card-watch-later__btn .iconfont {
  font-size: 16px;
}

/* 卡片详情：标题 + 日期 */
.bili-video-card__details {
  margin-top: 6px;
}

.bili-video-card__title {
  display: flow-root;
  max-height: 36px;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bili-video-card__title a {
  color: #18191c;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  text-decoration: none;
}

.list-video-item:hover .bili-video-card__title a {
  color: #00aeec;
}

.bili-video-card__subtitle {
  display: flex;
  margin-top: 4px;
  font-size: 12px;
  font-weight: 400;
  line-height: 16px;
  color: #9499a0;
}

.bili-video-card__subtitle span {
  color: #9499a0;
}

/* 加载 / 空状态 */
.loading-state {
  grid-column: 1 / -1;
  padding: 20px;
}

.empty-state {
  grid-column: 1 / -1;
  padding: 60px 0;
}
</style>
