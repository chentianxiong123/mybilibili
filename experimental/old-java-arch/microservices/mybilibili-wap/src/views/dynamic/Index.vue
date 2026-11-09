<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../../components/Header.vue'
import dynamicApi from '../../api/dynamic'
import api from '../../api/client'
import { getLiveListData } from '../../api/live'

const router = useRouter()

// UI state
const activeHeaderTab = ref<'all' | 'video'>('all') // 全部 vs 视频
const loading = ref(false)
const dynamicList = ref<any[]>([])
const followingUsers = ref<any[]>([])
const liveRooms = ref<any[]>([])
const currentPage = ref(1)
const pageSize = ref(10)
const hasMore = ref(true)
const searchPlaceholder = ref('搜搜看...')

// Current logged in user info
const isLoggedIn = computed(() => !!localStorage.getItem('token'))
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  if (!userStr) return null
  try {
    return JSON.parse(userStr)
  } catch (e) {
    return null
  }
})

// Post Dynamic Modal state
const showPublishModal = ref(false)
const publishContent = ref('')
const selectedImages = ref<File[]>([])
const imagePreviewUrls = ref<string[]>([])
const selectedVideoId = ref<number | null>(null)
const selectedVideoInfo = ref<any>(null)
const showVideoSelect = ref(false)
const myVideos = ref<any[]>([])

// Comment states (per-dynamic map)
const showCommentsMap = ref<Record<number, boolean>>({})
const commentsMap = ref<Record<number, any[]>>({})
const commentInputs = ref<Record<number, string>>({})
const loadingCommentsMap = ref<Record<number, boolean>>({})

// Image lightbox preview state
const showLightbox = ref(false)
const lightboxImageUrl = ref('')

// Initialize data
onMounted(async () => {
  window.addEventListener('open-publish-modal', openPublishModal)
  if (isLoggedIn.value && currentUser.value) {
    await fetchFollowingUsers()
  }
  await fetchLiveRooms()
  await fetchDynamics()
})

onUnmounted(() => {
  window.removeEventListener('open-publish-modal', openPublishModal)
})

// Fetch followed users
const fetchFollowingUsers = async () => {
  try {
    const currentUserId = currentUser.value?.id
    if (!currentUserId) return
    const res = await api.get(`/follow/user/${currentUserId}/following`)
    const data = res?.data || res
    if (Array.isArray(data)) {
      followingUsers.value = data.map((u: any) => ({
        id: u.id,
        name: u.username || u.nickname || '用户',
        avatar: u.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'
      }))
    }
  } catch (e) {
    console.error('获取关注列表失败:', e)
  }
}

// Fetch active live rooms to show live indicators
const fetchLiveRooms = async () => {
  try {
    const res = await getLiveListData()
    if (res.code === '1' && res.data?.list) {
      liveRooms.value = res.data.list
    }
  } catch (e) {
    console.error('获取直播间失败:', e)
  }
}

// Compute which followed users are currently streaming
const followedUsersWithLive = computed(() => {
  return followingUsers.value.map(user => {
    // Check if user is an anchor in liveRooms
    const live = liveRooms.value.find(room => room.uname === user.name || room.uname === user.id.toString())
    return {
      ...user,
      isLive: !!live,
      roomId: live?.roomId || null
    }
  })
})

const liveCount = computed(() => {
  return followedUsersWithLive.value.filter(u => u.isLive).length
})

// Fetch dynamic feeds
const fetchDynamics = async (isRefresh = false) => {
  if (loading.value) return
  if (isRefresh) {
    currentPage.value = 1
    dynamicList.value = []
    hasMore.value = true
  }
  loading.value = true

  try {
    let res: any
    // For WAP, we fetch global list or filter by "video" if selected
    res = await dynamicApi.getDynamicList(currentPage.value, pageSize.value)

    if (res && res.code === 200) {
      const list = res.data?.list || res.data || []

      // If header tab is "video", only keep posts with video references
      let filtered = list
      if (activeHeaderTab.value === 'video') {
        filtered = list.filter((item: any) => !!item.refManuscriptId)
      }

      const mapped = filtered.map((item: any) => ({
        ...item,
        stats: {
          shareCount: item.shareCount || 0,
          commentCount: item.commentCount || 0,
          likeCount: item.likeCount || 0,
          isLiked: item.isLiked || false
        },
        showFullText: false
      }))

      if (currentPage.value === 1) {
        dynamicList.value = mapped
      } else {
        dynamicList.value.push(...mapped)
      }
      hasMore.value = list.length === pageSize.value
    } else {
      hasMore.value = false
    }
  } catch (e) {
    console.error('获取动态失败:', e)
    hasMore.value = false
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (hasMore.value && !loading.value) {
    currentPage.value++
    fetchDynamics()
  }
}

// Switch tabs: All vs Video
const switchHeaderTab = (tab: 'all' | 'video') => {
  activeHeaderTab.value = tab
  fetchDynamics(true)
}

// Interactions
const handleLike = async (item: any) => {
  if (!isLoggedIn.value) {
    router.push('/m/login')
    return
  }
  try {
    if (item.stats.isLiked) {
      const res = await dynamicApi.unlikeDynamic(item.id)
      if (res.code === 200) {
        item.stats.isLiked = false
        item.stats.likeCount = Math.max(0, item.stats.likeCount - 1)
      }
    } else {
      const res = await dynamicApi.likeDynamic(item.id)
      if (res.code === 200) {
        item.stats.isLiked = true
        item.stats.likeCount++
      }
    }
  } catch (e) {
    console.error('点赞失败:', e)
  }
}

const handleForward = async (item: any) => {
  try {
    const res = await dynamicApi.shareDynamic(item.id)
    if (res.code === 200) {
      item.stats.shareCount++
      showToast('分享成功')
    }
  } catch (e) {
    console.error('分享失败:', e)
  }
}

// Comments Toggle and Interaction
const toggleComments = async (item: any) => {
  const dynamicId = item.id
  showCommentsMap.value[dynamicId] = !showCommentsMap.value[dynamicId]

  if (showCommentsMap.value[dynamicId] && !commentsMap.value[dynamicId]) {
    await fetchComments(dynamicId)
  }
}

const fetchComments = async (dynamicId: number) => {
  loadingCommentsMap.value[dynamicId] = true
  try {
    const res = await dynamicApi.getComments(dynamicId, 1, 30)
    if (res.code === 200) {
      commentsMap.value[dynamicId] = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('获取评论失败:', e)
  } finally {
    loadingCommentsMap.value[dynamicId] = false
  }
}

const submitComment = async (item: any) => {
  if (!isLoggedIn.value) {
    router.push('/m/login')
    return
  }
  const dynamicId = item.id
  const text = commentInputs.value[dynamicId] || ''
  if (!text.trim()) return

  try {
    const res = await dynamicApi.addComment(dynamicId, text)
    if (res.code === 200) {
      commentInputs.value[dynamicId] = ''
      showToast('评论成功')
      item.stats.commentCount++
      await fetchComments(dynamicId)
    } else {
      showToast(res.message || '评论失败')
    }
  } catch (e) {
    showToast('评论失败')
  }
}

// Lightbox for image preview
const openLightbox = (url: string) => {
  lightboxImageUrl.value = url
  showLightbox.value = true
}

// Format numbers
const formatNum = (num: number) => {
  if (!num) return '0'
  return num >= 10000 ? (num / 10000).toFixed(1) + '万' : num.toString()
}

// Format date time
const formatTime = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`
  return date.toLocaleDateString()
}

// Navigation helpers
const goToLiveRoom = (roomId: number) => {
  router.push(`/m/live/${roomId}`)
}

const goToVideo = (manuscriptId: number) => {
  router.push(`/m/video/${manuscriptId}`)
}

const showToast = (msg: string) => {
  const el = document.createElement('div')
  el.className = 'wap-toast'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

// Publishing dynamic functions
const openPublishModal = () => {
  if (!isLoggedIn.value) {
    router.push('/m/login')
    return
  }
  publishContent.value = ''
  selectedImages.value = []
  imagePreviewUrls.value = []
  selectedVideoId.value = null
  selectedVideoInfo.value = null
  showPublishModal.value = true
}

const handleImageSelect = (event: any) => {
  const files = Array.from(event.target.files) as File[]
  if (files.length + selectedImages.value.length > 9) {
    showToast('最多只能上传9张图片')
    return
  }
  files.forEach(file => {
    if (!file.type.startsWith('image/')) {
      showToast('只能上传图片文件')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      showToast('图片大小不能超过5MB')
      return
    }
    selectedImages.value.push(file)
    imagePreviewUrls.value.push(URL.createObjectURL(file))
  })
}

const removeImage = (index: number) => {
  URL.revokeObjectURL(imagePreviewUrls.value[index])
  selectedImages.value.splice(index, 1)
  imagePreviewUrls.value.splice(index, 1)
}

const openVideoSelect = async () => {
  showVideoSelect.value = true
  try {
    const res = await api.get('/manuscript/list?page=1&size=50')
    const list = res?.data || res || []
    myVideos.value = list
  } catch (e) {
    showToast('获取视频列表失败')
  }
}

const handleSelectVideo = (video: any) => {
  selectedVideoId.value = video.manuscriptId || video.id
  selectedVideoInfo.value = {
    title: video.title,
    cover: video.coverUrl || video.pic
  }
  showVideoSelect.value = false
}

const handlePublish = async () => {
  if (!publishContent.value.trim() && selectedImages.value.length === 0 && !selectedVideoId.value) {
    showToast('请输入内容或添加图片/视频')
    return
  }

  try {
    const formData = new FormData()
    formData.append('content', publishContent.value)

    if (selectedVideoId.value) {
      formData.append('refVideoId', selectedVideoId.value.toString())
    }

    selectedImages.value.forEach((file) => {
      formData.append('images', file)
    })

    const res = await dynamicApi.publishDynamic(formData)
    if (res.code === 200) {
      showToast('发布成功，经验值+5')
      showPublishModal.value = false
      await fetchDynamics(true)
    } else {
      showToast(res.message || '发布失败')
    }
  } catch (error: any) {
    showToast('发布失败')
  }
}

// Expose openPublishModal for trigger from outside
defineExpose({
  openPublishModal
})
</script>

<template>
  <div class="dynamic-page">
    <!-- Header matching the screenshot -->
    <header class="dynamic-header-nav">
      <div class="header-tab-container">
        <span
          :class="['header-tab', { active: activeHeaderTab === 'all' }]"
          @click="switchHeaderTab('all')"
        >
          全部
        </span>
        <span
          :class="['header-tab', { active: activeHeaderTab === 'video' }]"
          @click="switchHeaderTab('video')"
        >
          视频
        </span>
      </div>
      <!-- Edit/Pen post icon on the right -->
      <button class="post-edit-btn" @click="openPublishModal">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
          <path d="M18.5 2.5a2.121 2.121 0 1 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
        </svg>
      </button>
    </header>

    <!-- Most visited horizontal scroll list -->
    <section class="most-visited-section" v-if="isLoggedIn && followedUsersWithLive.length">
      <div class="visited-title-row">
        <span class="title">最常访问</span>
        <span class="live-info" v-if="liveCount > 0">{{ liveCount }}人直播中 · 更多 &gt;</span>
        <span class="live-info" v-else>更多 &gt;</span>
      </div>
      <div class="horizontal-scroll-container">
        <div
          v-for="user in followedUsersWithLive"
          :key="user.id"
          :class="['user-avatar-card', { 'is-live': user.isLive }]"
          @click="user.isLive && user.roomId ? goToLiveRoom(user.roomId) : null"
        >
          <div class="avatar-wrapper">
            <img :src="user.avatar" class="user-avatar" alt="" />
            <div class="live-pill" v-if="user.isLive">
              <span class="live-dot"></span>直播中
            </div>
          </div>
          <span class="user-name">{{ user.name }}</span>
        </div>
      </div>
    </section>

    <!-- Dynamic feed list -->
    <main class="dynamic-feed-container">
      <div v-for="item in dynamicList" :key="item.id" class="feed-card">
        <!-- Feed Card Header -->
        <div class="feed-card-header">
          <div class="user-avatar-wrap">
            <img
              :src="item.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'"
              class="user-avatar"
              alt=""
            />
            <!-- Optional Live Tag if the user matches liveRooms -->
            <span
              class="live-status-badge"
              v-if="liveRooms.some(r => r.uname === item.user?.username)"
              @click="goToLiveRoom(liveRooms.find(r => r.uname === item.user?.username)?.roomId)"
            >
              直播中
            </span>
          </div>
          <div class="user-info-text">
            <span class="user-nickname">{{ item.user?.username || '哔哩哔哩用户' }}</span>
            <span class="post-time">{{ formatTime(item.createdAt) }}</span>
          </div>
          <button class="menu-dots-btn">
            <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
              <circle cx="12" cy="5" r="2"></circle>
              <circle cx="12" cy="12" r="2"></circle>
              <circle cx="12" cy="19" r="2"></circle>
            </svg>
          </button>
        </div>

        <!-- Feed Card Content Body -->
        <div class="feed-card-body">
          <p :class="['post-text', { 'line-clamp': !item.showFullText }]">
            {{ item.content }}
          </p>
          <span
            class="expand-text-toggle"
            v-if="item.content && item.content.length > 90"
            @click="item.showFullText = !item.showFullText"
          >
            {{ item.showFullText ? '收起' : '展开全文' }}
          </span>

          <!-- Images Grid -->
          <div
            v-if="item.imageUrls && item.imageUrls.length > 0"
            :class="['images-grid', `grid-${Math.min(9, item.imageUrls.length)}`]"
          >
            <div
              v-for="(url, index) in item.imageUrls"
              :key="index"
              class="grid-image-item"
              @click="openLightbox(url)"
            >
              <img :src="url" class="feed-img" alt="" />
            </div>
          </div>

          <!-- Video Reference Card -->
          <div
            v-if="item.refManuscriptId"
            class="ref-video-card"
            @click="goToVideo(item.refManuscriptId)"
          >
            <div class="video-cover-wrap">
              <img
                :src="item.refVideo?.cover || 'https://picsum.photos/320/180?random=' + item.refManuscriptId"
                class="video-cover"
                alt=""
              />
              <span class="video-duration" v-if="item.refVideo?.duration">{{ item.refVideo.duration }}</span>
            </div>
            <div class="video-details">
              <h4 class="video-title">{{ item.refVideo?.title || '引用视频 #' + item.refManuscriptId }}</h4>
              <div class="video-stats">
                <span class="view-count">
                  <svg viewBox="0 0 24 24" fill="currentColor" width="12" height="12" style="display: inline-block; vertical-align: middle; margin-right: 2px;">
                    <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"></path>
                  </svg>
                  {{ formatNum(item.refVideo?.viewCount || 0) }}次观看
                </span>
                <span class="video-anchor" v-if="item.refVideo?.username">UP: {{ item.refVideo.username }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Feed Card Action Footer -->
        <div class="feed-card-footer">
          <button class="footer-action-btn" @click="handleForward(item)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"></path>
              <polyline points="16 6 12 2 8 6"></polyline>
              <line x1="12" y1="2" x2="12" y2="15"></line>
            </svg>
            <span>{{ formatNum(item.stats?.shareCount) || '转发' }}</span>
          </button>

          <button
            :class="['footer-action-btn', { active: showCommentsMap[item.id] }]"
            @click="toggleComments(item)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
            <span>{{ formatNum(item.stats?.commentCount) || '评论' }}</span>
          </button>

          <button
            :class="['footer-action-btn', { liked: item.stats?.isLiked }]"
            @click="handleLike(item)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
              <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"></path>
            </svg>
            <span>{{ formatNum(item.stats?.likeCount) || '点赞' }}</span>
          </button>
        </div>

        <!-- Comments list inside card -->
        <div v-show="showCommentsMap[item.id]" class="card-comments-section">
          <div v-if="loadingCommentsMap[item.id]" class="comments-loading">
            加载中...
          </div>
          <div v-else class="comments-list">
            <div v-for="c in commentsMap[item.id]" :key="c.id" class="comment-item">
              <img :src="c.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" class="comment-avatar" alt="" />
              <div class="comment-right">
                <div class="comment-meta">
                  <span class="commenter-name">{{ c.user?.username || '用户' }}</span>
                  <span class="comment-time">{{ formatTime(c.createdAt) }}</span>
                </div>
                <p class="comment-text">{{ c.content }}</p>
              </div>
            </div>
            <div v-if="!commentsMap[item.id] || commentsMap[item.id].length === 0" class="empty-comments">
              暂无评论，发一条吧~
            </div>
          </div>
          <!-- Comment input -->
          <div class="comment-write-row">
            <input
              v-model="commentInputs[item.id]"
              placeholder="发一条友善的评论吧..."
              class="comment-write-input"
              @keyup.enter="submitComment(item)"
            />
            <button class="comment-send-btn" @click="submitComment(item)">发送</button>
          </div>
        </div>
      </div>

      <!-- Load more / empty state -->
      <div v-if="hasMore" class="feed-load-more" @click="loadMore">
        <span v-if="!loading">点击加载更多</span>
        <span v-else>加载中...</span>
      </div>
      <div v-else class="feed-end-label">
        — 已经到底啦 —
      </div>
      <div v-if="!loading && dynamicList.length === 0" class="feed-empty-state">
        暂无关注或动态推荐
      </div>
    </main>

    <!-- Post Dynamic Modal Drawer -->
    <div :class="['publish-modal-overlay', { show: showPublishModal }]" @click.self="showPublishModal = false">
      <div class="publish-drawer">
        <div class="drawer-header">
          <button class="close-drawer-btn" @click="showPublishModal = false">取消</button>
          <span class="drawer-title">发布动态</span>
          <button class="publish-submit-btn" :disabled="!publishContent.trim() && selectedImages.length === 0" @click="handlePublish">
            发布
          </button>
        </div>
        <div class="drawer-body">
          <textarea
            v-model="publishContent"
            placeholder="分享点有意思的事吧~"
            class="publish-textarea"
            rows="5"
          ></textarea>

          <!-- Uploaded images previews -->
          <div class="publish-images-wrap" v-if="imagePreviewUrls.length > 0">
            <div v-for="(url, idx) in imagePreviewUrls" :key="idx" class="img-preview-card">
              <img :src="url" alt="" />
              <button class="del-img-btn" @click="removeImage(idx)">×</button>
            </div>
          </div>

          <!-- Video ref preview -->
          <div class="publish-video-ref-wrap" v-if="selectedVideoInfo">
            <div class="selected-video-card">
              <img :src="selectedVideoInfo.cover" alt="" class="video-mini-cover" />
              <span class="video-mini-title">{{ selectedVideoInfo.title }}</span>
              <button class="del-video-btn" @click="selectedVideoId = null; selectedVideoInfo = null">×</button>
            </div>
          </div>

          <!-- Options bar -->
          <div class="drawer-options-bar">
            <label class="option-tool-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                <circle cx="8.5" cy="8.5" r="1.5"></circle>
                <polyline points="21 15 16 10 5 21"></polyline>
              </svg>
              <input type="file" accept="image/*" multiple hidden @change="handleImageSelect" />
            </label>
            <button class="option-tool-btn" @click="openVideoSelect">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                <polygon points="23 7 16 12 23 17 23 7"></polygon>
                <rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Video Selection sub-modal -->
    <div :class="['video-select-overlay', { show: showVideoSelect }]" @click.self="showVideoSelect = false">
      <div class="video-select-modal">
        <div class="modal-header">
          <span>选择要引用的稿件</span>
          <button @click="showVideoSelect = false">关闭</button>
        </div>
        <div class="modal-list-body">
          <div
            v-for="v in myVideos"
            :key="v.id"
            class="my-video-select-item"
            @click="handleSelectVideo(v)"
          >
            <img :src="v.coverUrl || v.pic" class="mini-cov" alt="" />
            <span class="mini-tit">{{ v.title }}</span>
          </div>
          <div v-if="myVideos.length === 0" class="no-videos-placeholder">
            暂无你发布的视频稿件
          </div>
        </div>
      </div>
    </div>

    <!-- Fullscreen image Lightbox -->
    <div :class="['lightbox-overlay', { show: showLightbox }]" @click="showLightbox = false">
      <img :src="lightboxImageUrl" class="lightbox-img" alt="" />
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.dynamic-page {
  background-color: #f4f4f4;
  min-height: 100vh;
  padding-bottom: 60px;
}

/* WAP Toast custom styling */
:global(.wap-toast) {
  position: fixed;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  background-color: rgba(0, 0, 0, 0.82);
  color: #fff;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 13px;
  z-index: 10000;
  box-shadow: 0 4px 10px rgba(0,0,0,0.15);
  pointer-events: none;
}

/* Header style matching screenshot */
.dynamic-header-nav {
  position: sticky;
  top: 0;
  height: 48px;
  background-color: #fff;
  border-bottom: 1px solid #e3e5e7;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99;

  .header-tab-container {
    display: flex;
    gap: 20px;

    .header-tab {
      font-size: 15px;
      color: #61666d;
      padding: 10px 4px;
      position: relative;
      cursor: pointer;

      &.active {
        color: $theme-pink;
        font-weight: bold;

        &::after {
          content: '';
          position: absolute;
          bottom: 2px;
          left: 0;
          right: 0;
          height: 3px;
          background-color: $theme-pink;
          border-radius: 2px;
        }
      }
    }
  }

  .post-edit-btn {
    position: absolute;
    right: 12px;
    background: none;
    border: none;
    color: #61666d;
    cursor: pointer;
    display: flex;
    align-items: center;
    padding: 6px;

    svg {
      width: 20px;
      height: 20px;
    }
  }
}

/* Most Visited section scroll bar */
.most-visited-section {
  background-color: #fff;
  margin-bottom: 8px;
  padding: 12px;

  .visited-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;

    .title {
      font-size: 14px;
      font-weight: bold;
      color: #18191c;
    }

    .live-info {
      font-size: 12px;
      color: #9499a0;
    }
  }

  .horizontal-scroll-container {
    display: flex;
    overflow-x: auto;
    gap: 14px;
    padding-bottom: 4px;

    &::-webkit-scrollbar {
      display: none;
    }

    .user-avatar-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      flex-shrink: 0;
      width: 60px;
      cursor: pointer;

      .avatar-wrapper {
        position: relative;
        width: 48px;
        height: 48px;

        .user-avatar {
          width: 100%;
          height: 100%;
          border-radius: 50%;
          object-fit: cover;
          border: 1.5px solid #e3e5e7;
        }

        .live-pill {
          position: absolute;
          bottom: -4px;
          left: 50%;
          transform: translateX(-50%);
          background-color: $theme-pink;
          color: #fff;
          font-size: 8px;
          padding: 1px 4px;
          border-radius: 6px;
          white-space: nowrap;
          display: flex;
          align-items: center;
          gap: 2px;
          border: 1px solid #fff;

          .live-dot {
            width: 4px;
            height: 4px;
            background-color: #fff;
            border-radius: 50%;
            display: inline-block;
          }
        }
      }

      &.is-live {
        .avatar-wrapper .user-avatar {
          border-color: $theme-pink;
        }
      }

      .user-name {
        margin-top: 6px;
        font-size: 10px;
        color: #61666d;
        text-align: center;
        max-width: 100%;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
}

/* Feed Cards styling */
.dynamic-feed-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feed-card {
  background-color: #fff;
  padding: 14px;

  .feed-card-header {
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .user-avatar-wrap {
      position: relative;
      width: 42px;
      height: 42px;
      flex-shrink: 0;

      .user-avatar {
        width: 100%;
        height: 100%;
        border-radius: 50%;
        object-fit: cover;
      }

      .live-status-badge {
        position: absolute;
        bottom: -2px;
        right: -2px;
        background-color: $theme-pink;
        color: #fff;
        font-size: 7px;
        padding: 0 3px;
        border-radius: 4px;
        white-space: nowrap;
        border: 1px solid #fff;
        transform: scale(0.9);
      }
    }

    .user-info-text {
      flex: 1;
      margin-left: 10px;
      display: flex;
      flex-direction: column;

      .user-nickname {
        font-size: 13.5px;
        font-weight: 500;
        color: $theme-pink;
      }

      .post-time {
        font-size: 11px;
        color: #9499a0;
        margin-top: 2px;
      }
    }

    .menu-dots-btn {
      background: none;
      border: none;
      color: #9499a0;
      cursor: pointer;
      padding: 4px;
    }
  }

  .feed-card-body {
    padding-left: 0;
    margin-bottom: 12px;

    .post-text {
      font-size: 14.5px;
      color: #18191c;
      line-height: 1.5;
      white-space: pre-wrap;
      word-break: break-word;

      &.line-clamp {
        display: -webkit-box;
        -webkit-line-clamp: 4;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }
    }

    .expand-text-toggle {
      display: inline-block;
      margin-top: 6px;
      color: #178aeb;
      font-size: 13px;
      cursor: pointer;
    }

    /* Images grid matching different counts */
    .images-grid {
      display: grid;
      gap: 5px;
      margin-top: 10px;

      &.grid-1 {
        grid-template-columns: 1fr;
        max-width: 70%;

        .grid-image-item {
          padding-top: 75%; /* 4:3 */
        }
      }

      &.grid-2 {
        grid-template-columns: repeat(2, 1fr);

        .grid-image-item {
          padding-top: 100%;
        }
      }

      &.grid-3, &.grid-5, &.grid-6, &.grid-7, &.grid-8, &.grid-9 {
        grid-template-columns: repeat(3, 1fr);

        .grid-image-item {
          padding-top: 100%;
        }
      }

      &.grid-4 {
        grid-template-columns: repeat(2, 1fr);
        max-width: 80%;

        .grid-image-item {
          padding-top: 100%;
        }
      }

      .grid-image-item {
        position: relative;
        overflow: hidden;
        border-radius: 6px;
        background-color: #f1f2f3;
        cursor: pointer;

        .feed-img {
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          object-fit: cover;
        }
      }
    }

    /* Embedded Video Ref Card */
    .ref-video-card {
      display: flex;
      margin-top: 10px;
      background-color: #f6f7f8;
      border-radius: 8px;
      overflow: hidden;
      border: 1px solid #e3e5e7;
      cursor: pointer;

      .video-cover-wrap {
        position: relative;
        width: 120px;
        height: 72px;
        flex-shrink: 0;

        .video-cover {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }

        .video-duration {
          position: absolute;
          bottom: 4px;
          right: 4px;
          background-color: rgba(0, 0, 0, 0.65);
          color: #fff;
          font-size: 9px;
          padding: 1px 3px;
          border-radius: 2px;
        }
      }

      .video-details {
        flex: 1;
        padding: 8px 10px;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        min-width: 0;

        .video-title {
          font-size: 12.5px;
          color: #18191c;
          font-weight: 500;
          line-height: 1.4;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
          word-break: break-all;
        }

        .video-stats {
          display: flex;
          justify-content: space-between;
          font-size: 10px;
          color: #9499a0;
          margin-top: 4px;

          .video-anchor {
            max-width: 100px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }
        }
      }
    }
  }

  /* Card Actions footer */
  .feed-card-footer {
    display: flex;
    border-top: 1px solid #e3e5e7;
    padding-top: 10px;

    .footer-action-btn {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      background: none;
      border: none;
      color: #61666d;
      font-size: 12px;
      cursor: pointer;
      padding: 6px 0;
      transition: color 0.2s;

      svg {
        stroke: #61666d;
      }

      &.liked {
        color: $theme-pink;
        svg {
          stroke: $theme-pink;
          fill: $theme-pink;
        }
      }

      &.active {
        color: #178aeb;
        svg {
          stroke: #178aeb;
        }
      }
    }
  }
}

/* Inline Comments section inside card */
.card-comments-section {
  margin-top: 12px;
  background-color: #f6f7f8;
  border-radius: 6px;
  padding: 10px;

  .comments-loading, .empty-comments {
    text-align: center;
    font-size: 12px;
    color: #9499a0;
    padding: 12px 0;
  }

  .comments-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 250px;
    overflow-y: auto;
    margin-bottom: 10px;
    padding-right: 4px;

    .comment-item {
      display: flex;
      align-items: flex-start;
      gap: 8px;

      .comment-avatar {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        object-fit: cover;
      }

      .comment-right {
        flex: 1;
        border-bottom: 1px solid #e3e5e7;
        padding-bottom: 6px;

        .comment-meta {
          display: flex;
          justify-content: space-between;
          font-size: 11px;

          .commenter-name {
            color: #61666d;
            font-weight: 500;
          }

          .comment-time {
            color: #9499a0;
          }
        }

        .comment-text {
          font-size: 12.5px;
          color: #18191c;
          margin-top: 3px;
          line-height: 1.4;
          word-break: break-all;
        }
      }

      &:last-child .comment-right {
        border-bottom: none;
      }
    }
  }

  .comment-write-row {
    display: flex;
    gap: 8px;
    align-items: center;

    .comment-write-input {
      flex: 1;
      height: 32px;
      border: 1px solid #e3e5e7;
      border-radius: 16px;
      padding: 0 12px;
      font-size: 13px;
      outline: none;
      background-color: #fff;

      &:focus {
        border-color: $theme-pink;
      }
    }

    .comment-send-btn {
      height: 32px;
      background-color: $theme-pink;
      color: #fff;
      border: none;
      padding: 0 14px;
      border-radius: 16px;
      font-size: 13px;
      cursor: pointer;
      font-weight: 500;
    }
  }
}

/* Feed end labels */
.feed-load-more {
  text-align: center;
  padding: 16px;
  color: $theme-pink;
  font-size: 13px;
  cursor: pointer;
}

.feed-end-label {
  text-align: center;
  padding: 24px;
  color: #9499a0;
  font-size: 12px;
}

.feed-empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #9499a0;
  font-size: 13.5px;
}

/* Lightbox full screen preview */
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.25s ease;

  &.show {
    opacity: 1;
    pointer-events: auto;
  }

  .lightbox-img {
    max-width: 95%;
    max-height: 90%;
    object-fit: contain;
  }
}

/* Publishing Drawer / Modal */
.publish-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 2000;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;

  &.show {
    opacity: 1;
    pointer-events: auto;

    .publish-drawer {
      transform: translateY(0);
    }
  }

  .publish-drawer {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    background-color: #fff;
    border-radius: 12px 12px 0 0;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    transform: translateY(100%);
    transition: transform 0.3s cubic-bezier(0.1, 0.76, 0.55, 0.94);

    .drawer-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;
      border-bottom: 1px solid #e3e5e7;

      .drawer-title {
        font-size: 16px;
        font-weight: 500;
        color: #18191c;
      }

      .close-drawer-btn {
        background: none;
        border: none;
        color: #61666d;
        font-size: 14px;
        cursor: pointer;
      }

      .publish-submit-btn {
        background-color: $theme-pink;
        color: #fff;
        border: none;
        padding: 5px 14px;
        border-radius: 15px;
        font-size: 13.5px;
        cursor: pointer;
        font-weight: 500;

        &:disabled {
          background-color: #e3e5e7;
          color: #9499a0;
          cursor: not-allowed;
        }
      }
    }

    .drawer-body {
      padding: 16px;
      overflow-y: auto;

      .publish-textarea {
        width: 100%;
        border: none;
        outline: none;
        resize: none;
        font-size: 15px;
        color: #18191c;
        line-height: 1.5;
        font-family: inherit;

        &::placeholder {
          color: #9499a0;
        }
      }

      /* publish images preview list */
      .publish-images-wrap {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        margin-top: 16px;

        .img-preview-card {
          position: relative;
          width: 80px;
          height: 80px;

          img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            border-radius: 6px;
            border: 1px solid #e3e5e7;
          }

          .del-img-btn {
            position: absolute;
            top: -5px;
            right: -5px;
            width: 18px;
            height: 18px;
            background-color: rgba(0,0,0,0.7);
            color: #fff;
            border-radius: 50%;
            border: none;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            cursor: pointer;
          }
        }
      }

      /* publish ref video mini card */
      .publish-video-ref-wrap {
        margin-top: 16px;

        .selected-video-card {
          display: flex;
          align-items: center;
          background-color: #f6f7f8;
          padding: 8px 12px;
          border-radius: 8px;
          border: 1px solid #e3e5e7;
          position: relative;

          .video-mini-cover {
            width: 50px;
            height: 32px;
            object-fit: cover;
            border-radius: 4px;
          }

          .video-mini-title {
            margin-left: 10px;
            font-size: 12px;
            color: #18191c;
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            padding-right: 20px;
          }

          .del-video-btn {
            position: absolute;
            right: 8px;
            background: none;
            border: none;
            color: #9499a0;
            font-size: 16px;
            cursor: pointer;
          }
        }
      }

      .drawer-options-bar {
        display: flex;
        gap: 16px;
        margin-top: 24px;
        border-top: 1px solid #f4f4f4;
        padding-top: 12px;

        .option-tool-btn {
          background: none;
          border: none;
          color: #61666d;
          cursor: pointer;
          padding: 8px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;

          &:active {
            background-color: #f1f2f3;
          }
        }
      }
    }
  }
}

/* Video Select Modal Overlay */
.video-select-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;

  &.show {
    opacity: 1;
    pointer-events: auto;
  }

  .video-select-modal {
    background-color: #fff;
    width: 85%;
    max-width: 350px;
    max-height: 70vh;
    border-radius: 10px;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .modal-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px;
      border-bottom: 1px solid #e3e5e7;
      font-size: 14px;
      font-weight: 500;

      button {
        background: none;
        border: none;
        color: $theme-pink;
        cursor: pointer;
      }
    }

    .modal-list-body {
      flex: 1;
      overflow-y: auto;
      padding: 8px;

      .my-video-select-item {
        display: flex;
        align-items: center;
        padding: 8px;
        border-bottom: 1px solid #f4f4f4;
        cursor: pointer;

        &:active {
          background-color: #f6f7f8;
        }

        .mini-cov {
          width: 60px;
          height: 38px;
          object-fit: cover;
          border-radius: 4px;
          flex-shrink: 0;
        }

        .mini-tit {
          margin-left: 10px;
          font-size: 12.5px;
          color: #18191c;
          overflow: hidden;
          text-overflow: ellipsis;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
        }
      }

      .no-videos-placeholder {
        text-align: center;
        font-size: 13px;
        color: #9499a0;
        padding: 40px 10px;
      }
    }
  }
}
</style>
