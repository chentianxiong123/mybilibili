<script setup>
import { safeStorage } from '@/utils/safeStorage'
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowDown } from '@element-plus/icons-vue'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { dynamicApi } from '@/api/dynamic.ts'
import { useUserStore } from '@/stores/user.ts'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { searchApi } from '@/api/search.ts'
import DynamicPublishPanel from '@/components/DynamicPublishPanel.vue'
import DynamicCard from '@/components/DynamicCard.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const currentUser = computed(() => userStore.userInfo || {})

const hydrated = ref(false)

const publishPanelRef = ref(null)

const followingUsers = ref([
  { id: null, name: '全部动态', avatar: '', isAll: true }
])
const selectedUserId = ref(null)

const dynamicList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const hasMore = ref(true)
const loading = ref(false)

// 虚拟滚动（客户端挂载后再创建，避免 SSR 引入 window）
const scrollContainer = ref(null)
const virtualizer = ref(null)
onMounted(() => {
  virtualizer.value = useVirtualizer({
    count: computed(() => dynamicList.value.length),
    getScrollElement: () => scrollContainer.value,
    estimateSize: () => 200,
    overscan: 3
  })
})

const hotSearchList = ref([])

const fetchHotSearch = async () => {
  try {
    const response = await searchApi.getHotSearch()
    if (Array.isArray(response)) {
      hotSearchList.value = response.map(item => ({
        rank: item.rank,
        title: item.keyword,
        hot: item.rank <= 3 ? '热' : (item.rank <= 5 ? '新' : ''),
        color: item.rank <= 3 ? '#ff2442' : '#ff6699'
      }))
    } else if (response && response.code === 200 && response.data) {
      hotSearchList.value = response.data.map(item => ({
        rank: item.rank,
        title: item.keyword,
        hot: item.rank <= 3 ? '热' : (item.rank <= 5 ? '新' : ''),
        color: item.rank <= 3 ? '#ff2442' : '#ff6699'
      }))
    }
  } catch (error) {
    console.error('获取热搜榜失败:', error)
    hotSearchList.value = []
  }
}

const handlePublish = async (payload) => {
  try {
    const formData = new FormData()
    formData.append('content', payload.content)
    if (payload.refVideoId) {
      formData.append('refVideoId', payload.refVideoId)
    }
    payload.images.forEach((file) => {
      formData.append('images', file)
    })
    const res = await dynamicApi.publishDynamic(formData)
    if (res.code === 200) {
      ElMessage.success('发布成功，经验值+5')
      publishPanelRef.value?.reset()
      await fetchDynamics()
    } else {
      ElMessage.error(res.message || '发布失败')
    }
  } catch (error) {
    ElMessage.error('发布失败：' + error.message)
  }
}

const fetchFollowingUsers = async () => {
  try {
    const currentUserId = userStore.userInfo?.id
    if (!currentUserId) {
      console.log('用户未登录，不获取关注列表')
      return
    }
    const res = await api.get(`/follow/user/${currentUserId}/following`)
    if (res.code === 200 && res.data) {
      followingUsers.value = [
        { id: null, name: '全部动态', avatar: '', isAll: true },
        ...res.data.map(user => ({
          id: user.id,
          name: user.username,
          avatar: user.avatar
        }))
      ]
    }
  } catch (error) {
    if (error.response?.status !== 404) {
      console.error('获取关注列表失败:', error)
    }
  }
}

const fetchDynamics = async () => {
  if (loading.value) return
  loading.value = true

  try {
    let res
    if (selectedUserId.value === null) {
      res = await dynamicApi.getDynamicList(currentPage.value, pageSize.value)
    } else {
      res = await dynamicApi.getFollowingDynamics(currentPage.value, pageSize.value, selectedUserId.value)
    }

    const data = Array.isArray(res) ? res : res?.data
    const list = data?.list || data || []
    const mappedList = list.map(item => ({
      ...item,
      stats: {
        shareCount: item.shareCount || 0,
        commentCount: item.commentCount || 0,
        likeCount: item.likeCount || 0,
        isLiked: item.isLiked || false
      }
    }))
    if (currentPage.value === 1) {
      dynamicList.value = mappedList
    } else {
      dynamicList.value.push(...mappedList)
    }
    hasMore.value = list.length === pageSize.value
  } catch (error) {
    console.error('获取动态失败:', error)
  } finally {
    loading.value = false
  }
}

const selectUser = (userId) => {
  selectedUserId.value = userId
  currentPage.value = 1
  dynamicList.value = []
  fetchDynamics()
}

const loadMore = () => {
  if (hasMore.value && !loading.value) {
    currentPage.value++
    fetchDynamics()
  }
}

const handleLike = async (item) => {
  try {
    if (!item.stats) {
      item.stats = {
        isLiked: false,
        likeCount: 0
      }
    }

    if (item.stats.isLiked) {
      const res = await dynamicApi.unlikeDynamic(item.id)
      if (res.code === 200) {
        if (res.data) {
          item.stats.isLiked = res.data.isLiked ?? false
          item.stats.likeCount = res.data.likeCount ?? Math.max(0, item.stats.likeCount - 1)
        } else {
          item.stats.isLiked = false
          item.stats.likeCount = Math.max(0, item.stats.likeCount - 1)
        }
        ElMessage.success('取消点赞成功')
      } else {
        ElMessage.error(res.message || '取消点赞失败')
      }
    } else {
      const res = await dynamicApi.likeDynamic(item.id)
      if (res.code === 200) {
        if (res.data) {
          item.stats.isLiked = res.data.isLiked ?? true
          item.stats.likeCount = res.data.likeCount ?? item.stats.likeCount + 1
        } else {
          item.stats.isLiked = true
          item.stats.likeCount = item.stats.likeCount + 1
        }
        ElMessage.success('点赞成功')
      } else {
        ElMessage.error(res.message || '点赞失败')
      }
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    ElMessage.error('操作失败')
  }
}

const handleForward = async (item) => {
  try {
    const res = await dynamicApi.shareDynamic(item.id)
    if (res.code === 200) {
      item.shareCount++
      ElMessage.success('分享成功')
    }
  } catch (error) {
    ElMessage.error('分享失败')
  }
}

const toggleComment = (item) => {
  item.showComments = !item.showComments
}

const goToUserProfile = (userId) => {
  router.push(`/profile/${userId}/home`)
}

const goToUserFollowing = (userId) => {
  router.push(`/profile/${userId}/following`)
}

const goToUserFollowers = (userId) => {
  router.push(`/profile/${userId}/followers`)
}

const goToUserDynamic = (userId) => {
  router.push(`/profile/${userId}/dynamic`)
}

const goToVideo = (videoId) => {
  router.push(`/manuscript/${videoId}`)
}

const goToManuscript = (manuscriptId) => {
  router.push(`/manuscript/${manuscriptId}`)
}

const goToSearch = (keyword) => {
  router.push(`/search?keyword=${encodeURIComponent(keyword)}`)
}

const initUserInfo = () => {
  const token = safeStorage.getItem("token")
  const userData = safeStorage.getItem('user')

  if (token && userData) {
    try {
      const user = JSON.parse(userData)
      userStore.setUserInfo(user)
      userStore.setLoginStatus(true)
      userStore.token = token
    } catch (error) {
      console.error('解析用户信息失败:', error)
    }
  }
}

onMounted(() => {
  hydrated.value = true
  initUserInfo()
  fetchFollowingUsers()
  fetchDynamics()
  fetchHotSearch()
})
</script>

<template>
  <div class="dynamic-page">
    <div class="dynamic-container">
      <aside class="left-sidebar">
        <div class="user-card" v-if="hydrated && userStore.isLoggedIn && currentUser.id">
          <div class="user-avatar-section" @click="goToUserProfile(currentUser.id)">
            <img loading="lazy" decoding="async" :src="currentUser.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default'" alt="头像" class="user-avatar">
            <div class="user-level" v-if="currentUser.level">LV{{ currentUser.level }}</div>
          </div>
          <div class="user-name">{{ currentUser.username }}</div>
          <div class="user-stats">
            <div class="stat-item" @click="goToUserFollowing(currentUser.id)">
              <div class="stat-value">{{ currentUser.followingCount || 0 }}</div>
              <div class="stat-label">关注</div>
            </div>
            <div class="stat-item" @click="goToUserFollowers(currentUser.id)">
              <div class="stat-value">{{ currentUser.followerCount || 0 }}</div>
              <div class="stat-label">粉丝</div>
            </div>
            <div class="stat-item" @click="goToUserDynamic(currentUser.id)">
              <div class="stat-value">{{ currentUser.dynamicCount || 0 }}</div>
              <div class="stat-label">动态</div>
            </div>
          </div>
        </div>
      </aside>

      <main class="main-content">
        <DynamicPublishPanel
          ref="publishPanelRef"
          @publish="handlePublish"
        />

        <div class="following-users-bar">
          <div
            v-for="user in followingUsers"
            :key="user.id || 'all'"
            :class="['following-user-item', { active: selectedUserId === user.id }]"
            @click="selectUser(user.id)"
          >
            <img loading="lazy" decoding="async" v-if="user.avatar" :src="user.avatar" :alt="user.name" class="following-avatar">
            <div v-else class="following-avatar default-avatar">全</div>
            <span class="following-name">{{ user.name }}</span>
          </div>
        </div>

        <div class="dynamic-list" ref="scrollContainer">
          <div v-if="virtualizer && dynamicList.length > 0" :style="{ height: `${virtualizer.value.getTotalSize()}px`, position: 'relative' }">
            <div
              v-for="vItem in virtualizer.value.getVirtualItems()"
              :key="vItem.key"
              :data-index="vItem.index"
              :ref="vItem.measureElement"
              :style="{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                transform: `translateY(${vItem.start}px)`
              }"
            >
              <DynamicCard
                :item="dynamicList[vItem.index]"
                @like="handleLike"
                @forward="handleForward"
                @toggle-comment="toggleComment"
                @go-to-user="goToUserProfile"
                @go-to-manuscript="goToManuscript"
              />
            </div>
          </div>

          <div v-else-if="dynamicList.length > 0">
            <DynamicCard
              v-for="item in dynamicList"
              :key="item.id"
              :item="item"
              @like="handleLike"
              @forward="handleForward"
              @toggle-comment="toggleComment"
              @go-to-user="goToUserProfile"
              @go-to-manuscript="goToManuscript"
            />
          </div>

          <div v-if="hasMore" class="load-more" @click="loadMore">
            <span v-if="!loading">加载更多</span>
            <span v-else>加载中...</span>
          </div>

          <div v-if="!loading && dynamicList.length === 0" class="empty-state">
            暂无动态
          </div>
        </div>
      </main>

      <aside class="right-sidebar">
        <div class="ad-banner">
          <img loading="lazy" decoding="async" src="https://picsum.photos/300/150?random=ad" alt="广告">
          <div class="ad-overlay">
            <div class="ad-title">社区中心</div>
          </div>
        </div>

        <div class="hot-search-card">
          <div class="hot-search-header">
            <span class="hot-search-title">热搜</span>
          </div>
          <div class="hot-search-list">
            <div v-for="item in hotSearchList" :key="item.rank" class="hot-search-item" @click="goToSearch(item.title)">
              <span :class="['hot-rank', { 'top-three': item.rank <= 3 }]">{{ item.rank }}</span>
              <span class="hot-title">{{ item.title }}</span>
              <span v-if="item.hot" class="hot-tag" :style="{ backgroundColor: item.color }">{{ item.hot }}</span>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.dynamic-page {
  min-height: 100vh;
  background-color: #f5f7fa;
}

.dynamic-container {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  gap: 20px;
  padding: 20px;
}

.left-sidebar {
  width: 240px;
  flex-shrink: 0;
}

.user-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.user-avatar-section {
  position: relative;
  display: inline-block;
  cursor: pointer;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #e3e5e7;
}

.user-level {
  position: absolute;
  bottom: 0;
  right: 0;
  background: linear-gradient(135deg, #ff6b9d, #feca57);
  color: #fff;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 600;
}

.user-name {
  margin-top: 12px;
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
}

.user-stats {
  display: flex;
  justify-content: space-around;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e3e5e7;
}

.stat-item {
  cursor: pointer;
  transition: opacity 0.3s;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #18191c;
}

.stat-label {
  font-size: 12px;
  color: #9499a0;
  margin-top: 4px;
}

.main-content {
  flex: 1;
  min-width: 0;
}

.following-users-bar {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  display: flex;
  gap: 16px;
  overflow-x: auto;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.following-user-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  flex-shrink: 0;
}

.following-user-item.active .following-avatar {
  border-color: #00aeec;
}

.following-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid transparent;
  transition: border-color 0.3s;
}

.default-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f2f3;
  color: #61666d;
  font-size: 14px;
  font-weight: 600;
}

.following-name {
  font-size: 12px;
  color: #61666d;
  max-width: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dynamic-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100vh - 260px);
  overflow-y: auto;
  padding-right: 4px;
}

.load-more {
  text-align: center;
  padding: 16px;
  color: #00aeec;
  cursor: pointer;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #9499a0;
}

.right-sidebar {
  width: 300px;
  flex-shrink: 0;
}

.ad-banner {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 16px;
}

.ad-banner img {
  width: 100%;
  height: 150px;
  object-fit: cover;
}

.ad-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(0, 174, 236, 0.8), rgba(0, 174, 236, 0.6));
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.ad-title {
  font-size: 24px;
  font-weight: 700;
}

.ad-subtitle {
  font-size: 14px;
  margin-top: 8px;
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
}

.hot-search-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
}

.hot-search-header {
  margin-bottom: 16px;
}

.hot-search-title {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
}

.hot-search-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hot-search-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.hot-rank {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: #9499a0;
  background: #f1f2f3;
  border-radius: 4px;
}

.hot-rank.top-three {
  color: #fff;
  background: #ff2442;
}

.hot-title {
  flex: 1;
  font-size: 14px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hot-tag {
  font-size: 10px;
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
}




</style>