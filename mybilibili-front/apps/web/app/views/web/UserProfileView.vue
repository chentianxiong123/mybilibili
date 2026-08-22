<script setup>
import { useAuth } from '@/composables/useAuth'
import { safeStorage } from '@/utils/safeStorage'
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { userApi, videoApi } from '@/api/client'
import { getUserProfileBackground } from '@/api/banner.ts'
import ProfileHeader from '@/components/profile/ProfileHeader.vue'
import ProfileSidebar from '@/components/profile/ProfileSidebar.vue'
import ProfileTabVideoList from '@/components/profile/ProfileTabVideoList.vue'
import ProfileTabDynamicList from '@/components/profile/ProfileTabDynamicList.vue'
import ProfileTabCollection from '@/components/profile/ProfileTabCollection.vue'
import ProfileTabSubmission from '@/components/profile/ProfileTabSubmission.vue'
import ProfileTabFavorite from '@/components/profile/ProfileTabFavorite.vue'
import ProfileTabSearch from '@/components/profile/ProfileTabSearch.vue'
import ProfileTabFollowList from '@/components/profile/ProfileTabFollowList.vue'
import ProfileTabFans from '@/components/profile/ProfileTabFans.vue'
import ProfileTabSettings from '@/components/profile/ProfileTabSettings.vue'
import InterestsPanel from '@/components/profile/InterestsPanel.vue'

const route = useRoute()
const router = useRouter()

// 个人信息数据
const userInfo = ref({
  username: '加载中...',
  avatar: 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
  signature: '加载中...',
  announcement: '',
  cover: 'https://picsum.photos/id/1025/1920/200',
  uid: '',
  birthday: '',
  gender: 0,
  stats: {
    following: 0,
    followers: 0,
    likes: 0,
    views: 0
  },
  tags: []
})

// 当前用户
const { user: currentUser } = useAuth()

// 当前用户ID，从路由参数或本地存储获取
const userId = ref(route.params.id || currentUser.value?.id)

// 当前登录用户ID
const currentUserId = computed(() => currentUser.value?.id)

// 判断是否是自己的空间
const isOwnSpace = computed(() => {
  return currentUserId.value && userId.value && String(currentUserId.value) === String(userId.value)
})

const profileTabs = computed(() => {
  const tabs = ['主页', '动态', '投稿', '合集和列表', '收藏', '兴趣画像']
  if (isOwnSpace.value) {
    tabs.push('设置')
  }
  return tabs
})

// 跳转到我的头像页面
const goToAvatar = () => {
  if (isOwnSpace.value) {
    router.push('/personal-center/avatar')
  }
}

// 加载用户主页背景图
const loadUserProfileBackground = async () => {
  try {
    const res = await getUserProfileBackground()
    if (res.code === 200 && res.data) {
      userInfo.value.cover = res.data.imageUrl
    }
  } catch (error) {
    console.error('获取用户主页背景图失败:', error)
  }
}

// 加载用户信息
const loadUserInfo = async () => {
  try {
    const response = await userApi.getUserById(userId.value)
    if (response.code === 200) {
      const data = response.data
      userInfo.value = {
        username: data.nickname || data.username,
        avatar: data.avatar || 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
        signature: data.signature || '该用户暂无简介',
        announcement: data.announcement || '',
        cover: 'https://picsum.photos/id/1025/1920/200',
        uid: data.id || '',
        birthday: data.birthdate || '',
        gender: data.gender || 0,
        stats: {
          following: data.followingCount || 0,
          followers: data.followerCount || 0,
          likes: data.totalLikeCount || 0,
          views: data.totalViewCount || 0
        },
        tags: data.tags || []
      }
      // 加载背景图
      await loadUserProfileBackground()
      // 检查关注状态
      await checkFollowStatus()
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

// 根据路由路径获取当前活跃标签
const activeTab = computed(() => {
  const path = route.path
  if (path.endsWith('/home')) return '主页'
  if (path.endsWith('/dynamic')) return '动态'
  if (path.endsWith('/submissions')) return '投稿'
  if (path.endsWith('/collections')) return '合集和列表'
  if (path.endsWith('/favorites')) return '收藏'
  if (path.endsWith('/settings')) return '设置'
  if (path.endsWith('/interests')) return '兴趣画像'
  if (path.endsWith('/following')) return '关注'
  if (path.endsWith('/followers')) return '粉丝'
  if (path.endsWith('/search')) return '搜索'
  return '主页'
})

// 处理标签点击
const handleTabClick = (tab) => {
  let path = ''
  switch (tab) {
    case '主页':
      path = `/profile/${userId.value}/home`
      break
    case '动态':
      path = `/profile/${userId.value}/dynamic`
      break
    case '投稿':
      path = `/profile/${userId.value}/submissions`
      break
    case '合集和列表':
      path = `/profile/${userId.value}/collections`
      break
    case '收藏':
      path = `/profile/${userId.value}/favorites`
      break
    case '设置':
      path = `/profile/${userId.value}/settings`
      break
    case '兴趣画像':
      path = `/profile/${userId.value}/interests`
      break
    case '关注':
      path = `/profile/${userId.value}/following`
      break
    case '粉丝':
      path = `/profile/${userId.value}/followers`
      break
    case '搜索':
      path = `/profile/${userId.value}/search`
      break
    default:
      path = `/profile/${userId.value}/home`
  }
  router.push(path)
}

// 关注状态
const isFollowing = ref(false)
const followLoading = ref(false)

// 检查是否已关注
const checkFollowStatus = async () => {
  if (!currentUserId.value || !userId.value || isOwnSpace.value) return

  try {
    const response = await userApi.checkFollow(userId.value)
    if (response.code === 200) {
      isFollowing.value = response.data.following === true
    }
  } catch (error) {
    console.error('检查关注状态失败:', error)
  }
}

// 处理关注/取消关注
const handleFollow = async () => {
  const token = safeStorage.getItem("token")
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }

  // 不能关注自己
  if (isOwnSpace.value) {
    ElMessage.warning('不能关注自己')
    return
  }

  if (followLoading.value) return

  try {
    followLoading.value = true
    const response = await userApi.follow(userId.value, !isFollowing.value)
    if (response.code === 200) {
      const isNowFollowing = !isFollowing.value
      isFollowing.value = isNowFollowing
      // 更新粉丝数
      if (isNowFollowing) {
        userInfo.value.stats.followers++
      } else {
        userInfo.value.stats.followers = Math.max(0, userInfo.value.stats.followers - 1)
      }
      ElMessage.success(isNowFollowing ? '关注成功' : '取消关注成功')
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('关注操作失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  } finally {
    followLoading.value = false
  }
}

// 处理发消息
const handleSendMessage = () => {
  const token = safeStorage.getItem("token")
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }

  // 不能给自己发消息
  if (isOwnSpace.value) {
    ElMessage.warning('不能给自己发送消息')
    return
  }

  // 跳转到消息页面，带上对方用户ID
  router.push(`/message/private?userId=${userId.value}`)
}

// 加载状态
const loading = ref({
  userInfo: false,
  videos: false,
  dynamics: false,
  collections: false,
  favorites: false
})

// 视频排序选项
const videoSortOption = ref('最新发布')
const sortOptionMap = {
  '最新发布': 'latest',
  '最多播放': 'views',
  '最多收藏': 'collects'
}

// 视频数据
const representativeVideos = ref([])
const allVideos = ref([])

// 投稿数据
const submissions = ref({
  categories: [
    { name: '视频', count: 0 }
  ],
  videos: [],
  activeCategory: '视频',
  activeSort: '最新发布',
  sortOptions: ['最新发布', '最多播放', '最多收藏'],
  viewType: 'grid',
  pagination: {
    currentPage: 1,
    totalPages: 1,
    totalItems: 0
  }
})

// 视频搜索相关数据
const videoSearch = ref({
  keyword: '',
  activeCategory: '视频',
  activeSort: '最新发布',
  sortOptions: ['最新发布', '最多播放', '最多收藏'],
  searchResults: [],
  totalCount: 0,
  loading: false,
  viewType: 'grid'
})

// 搜索分类
const searchCategories = ref([
  { name: '视频', count: 0 },
  { name: '动态', count: 0 }
])

// 关注/粉丝列表数据
const followList = ref({
  activeSidebar: 'following', // 'following' | 'followers'
  filterType: 'all', // 'all' | 'recent' | 'frequent'
  searchKeyword: '',
  followingList: [],
  followersList: [],
  loading: false
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

// 加载用户视频列表
const loadUserVideos = async () => {
  if (!userId.value) return

  loading.value.videos = true
  try {
    // 获取当前排序参数
    const sortParam = sortOptionMap[videoSortOption.value] || 'latest'
    console.log('【调试】当前排序选项:', videoSortOption.value)
    console.log('【调试】映射后的排序参数:', sortParam)
    console.log('【调试】用户ID:', userId.value)

    // 只获取已上架的稿件（status=3）
    const response = await videoApi.getVideosByUserId(userId.value, sortParam, 3)
    console.log('【调试】API响应:', response)

    if (response.code === 200) {
      // 处理视频数据，添加date字段
      let videos = response.data.map(video => {
        return {
          ...video,
          date: formatDate(video.uploadTime)
        }
      })

      // 前端排序
      if (videoSortOption.value === '最多播放') {
        videos.sort((a, b) => (b.viewCount || 0) - (a.viewCount || 0))
      } else if (videoSortOption.value === '最多收藏') {
        videos.sort((a, b) => (b.collectCount || 0) - (a.collectCount || 0))
      } else {
        // 最新发布，按uploadTime降序
        videos.sort((a, b) => new Date(b.uploadTime) - new Date(a.uploadTime))
      }

      console.log('【调试】获取到的视频数量:', videos.length)
      console.log('【调试】视频列表（前3个）:', videos.slice(0, 3).map(v => ({
        id: v.id,
        title: v.title,
        viewCount: v.viewCount,
        collectCount: v.collectCount,
        uploadTime: v.uploadTime
      })))

      allVideos.value = videos
      // 更新投稿视频计数
      submissions.value.categories[0].count = videos.length
      submissions.value.videos = videos
      submissions.value.pagination.totalItems = videos.length
    }
  } catch (error) {
    console.error('【调试】获取用户视频失败:', error)
  } finally {
    loading.value.videos = false
  }
}

// 关注/取消关注用户
const handleFollowUser = async (targetUserId, isFollowing) => {
  try {
    const response = await userApi.follow(targetUserId, !isFollowing)
    if (response.code === 200) {
      // 更新本地状态
      const targetUser = followList.value.followersList.find(u => u.id === targetUserId)
      if (targetUser) {
        targetUser.isFollowing = !isFollowing
      }
      // 同时更新关注列表中的状态
      const followingUser = followList.value.followingList.find(u => u.id === targetUserId)
      if (followingUser) {
        followingUser.isFollowing = !isFollowing
      }
      ElMessage.success(isFollowing ? '已取消关注' : '关注成功')
      // 刷新用户统计信息
      await loadUserInfo()
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  }
}

// 加载关注列表
const loadFollowingList = async () => {
  followList.value.loading = true
  try {
    const response = await userApi.getFollowingList(userId.value)
    if (response.code === 200) {
      followList.value.followingList = response.data.map(user => ({
        id: user.id,
        nickname: user.nickname || user.username,
        avatar: user.avatar || 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
        signature: user.signature || '暂无简介',
        isFollowing: true
      }))
    } else {
      followList.value.followingList = []
      ElMessage.error(response.message || '获取关注列表失败')
    }
  } catch (error) {
    console.error('获取关注列表失败:', error)
    followList.value.followingList = []
    ElMessage.error('获取关注列表失败')
  } finally {
    followList.value.loading = false
  }
}

// 加载粉丝列表
const loadFollowersList = async () => {
  followList.value.loading = true
  try {
    const response = await userApi.getFollowerList(userId.value)
    if (response.code === 200) {
      // 获取当前登录用户的关注列表，用于判断粉丝是否被关注
      let currentUserFollowingIds = []
      if (currentUserId.value) {
        try {
          const followingResponse = await userApi.getFollowingList(currentUserId.value)
          if (followingResponse.code === 200) {
            currentUserFollowingIds = followingResponse.data.map(u => u.id)
          }
        } catch (e) {
          console.error('获取当前用户关注列表失败:', e)
        }
      }

      followList.value.followersList = response.data.map(user => ({
        id: user.id,
        nickname: user.nickname || user.username,
        avatar: user.avatar || 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
        signature: user.signature || '暂无简介',
        isFollowing: currentUserFollowingIds.includes(user.id)
      }))
    } else {
      followList.value.followersList = []
      ElMessage.error(response.message || '获取粉丝列表失败')
    }
  } catch (error) {
    console.error('获取粉丝列表失败:', error)
    followList.value.followersList = []
    ElMessage.error('获取粉丝列表失败')
  } finally {
    followList.value.loading = false
  }
}

// 切换关注/粉丝侧边栏
const handleSidebarClick = (type) => {
  followList.value.activeSidebar = type
  if (type === 'following') {
    router.push(`/profile/${userId.value}/following`)
  } else {
    router.push(`/profile/${userId.value}/followers`)
  }
}

// 处理主页排序变化
const handleSortChange = (option) => {
  videoSortOption.value = option
  videoSearch.value.activeSort = option
  // 重新加载视频列表
  loadUserVideos()
}

// 处理投稿页面排序变化
const handleSubmissionsSortChange = (option) => {
  submissions.value.activeSort = option
  // 同步更新主页和搜索页的排序选项
  videoSortOption.value = option
  videoSearch.value.activeSort = option
  // 重新加载视频列表
  loadUserVideos()
}

// 处理视频搜索
const handleVideoSearch = async () => {
  const keyword = videoSearch.value.keyword.trim()
  if (!keyword) {
    ElMessage.warning('请输入搜索关键词')
    return
  }

  videoSearch.value.loading = true
  try {
    if (activeTab.value !== '搜索') {
      router.push(`/profile/${userId.value}/search`)
    }
  } finally {
    videoSearch.value.loading = false
  }
}

// 监听路由参数变化
watch(() => route.params.id, (newId) => {
  console.log('路由参数变化，新的用户ID:', newId)
  userId.value = newId || JSON.parse(safeStorage.getItem('user'))?.id
  if (userId.value) {
    loadUserInfo()
    loadUserVideos()
    loadFollowingList()
    loadFollowersList()
  }
}, { immediate: true })

// 在组件挂载时加载数据
onMounted(() => {
  // 如果watch已经触发，这里不再重复加载
  if (!userId.value) {
    loadUserInfo()
    loadUserVideos()
    loadFollowingList()
    loadFollowersList()
  }

  // 根据当前路由设置关注/粉丝侧边栏状态
  const path = route.path
  if (path.endsWith('/following')) {
    followList.value.activeSidebar = 'following'
  } else if (path.endsWith('/followers')) {
    followList.value.activeSidebar = 'followers'
  }
})
</script>

<template>
  <div class="user-profile-page">
    <!-- 顶部个人信息 -->
    <ProfileHeader
      :user-info="userInfo"
      :user-id="userId"
      :is-own-space="isOwnSpace"
      :is-following="isFollowing"
      :follow-loading="followLoading"
      @follow="handleFollow"
      @send-message="handleSendMessage"
      @avatar-click="goToAvatar"
    />

    <!-- 选择栏和统计数据 -->
    <div class="profile-tabs-container">
      <div class="profile-tabs-wrapper">
        <!-- 左侧选择栏 -->
        <div class="profile-tabs">
          <div
            v-for="tab in profileTabs"
            :key="tab"
            :class="['tab-item', { active: activeTab === tab }]"
            @click="handleTabClick(tab)"
          >
            {{ tab }}
          </div>

          <!-- 视频搜索栏 -->
          <div class="search-bar-wrapper">
            <div class="search-input-wrapper">
              <el-icon class="search-icon"><Search /></el-icon>
              <input type="text"
                placeholder="搜索视频"
                class="search-input"
                v-model="videoSearch.keyword"
                @keyup.enter="handleVideoSearch"
              >
              </div>
          </div>
        </div>

        <!-- 右侧统计数据 -->
        <div class="stats-container">
          <div class="stat-item stat-entry-button" @click="router.push(`/profile/${userId}/following`)" v-if="userId">
            <div class="stat-label">关注</div>
            <div class="stat-value">{{ userInfo.stats.following }}</div>
          </div>
          <div class="stat-item stat-entry-button" @click="router.push(`/profile/${userId}/followers`)" v-if="userId">
            <div class="stat-label">粉丝</div>
            <div class="stat-value">{{ userInfo.stats.followers }}</div>
          </div>
          <div class="stat-item non-interactive">
            <div class="stat-label">获赞数</div>
            <div class="stat-value">{{ userInfo.stats.likes }}</div>
          </div>
          <div class="stat-item non-interactive">
            <div class="stat-label">播放数</div>
            <div class="stat-value">{{ userInfo.stats.views }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="profile-content">
      <!-- 左侧内容区域 -->
      <div class="main-content">
        <!-- 主页 -->
        <ProfileTabVideoList
          v-if="activeTab === '主页'"
          :user-id="userId"
          :all-videos="allVideos"
          :video-sort-option="videoSortOption"
          :loading="loading"
          :is-own-space="isOwnSpace"
          @sort-change="handleSortChange"
        />

        <!-- 动态列表 -->
        <ProfileTabDynamicList
          v-else-if="activeTab === '动态'"
          :user-id="userId"
          :user-info="userInfo"
          :loading="loading"
        />

        <!-- 合集和列表 -->
        <ProfileTabCollection
          v-else-if="activeTab === '合集和列表'"
          :user-id="userId"
          :is-own-space="isOwnSpace"
          :loading="loading"
        />

        <!-- 投稿 -->
        <ProfileTabSubmission
          v-else-if="activeTab === '投稿'"
          :submissions="submissions"
          :loading="loading"
          :is-own-space="isOwnSpace"
          @sort-change="handleSubmissionsSortChange"
        />

        <!-- 收藏 -->
        <ProfileTabFavorite
          v-else-if="activeTab === '收藏'"
          :user-id="userId"
          :is-own-space="isOwnSpace"
          :user-info="userInfo"
          :loading="loading"
        />

        <!-- 搜索 -->
        <ProfileTabSearch
          v-else-if="activeTab === '搜索'"
          :video-search="videoSearch"
          :all-videos="allVideos"
          :search-categories="searchCategories"
        />

        <!-- 关注列表 -->
        <ProfileTabFollowList
          v-else-if="activeTab === '关注'"
          :follow-list="followList"
          @sidebar-click="handleSidebarClick"
          @follow-user="handleFollowUser"
        />

        <!-- 粉丝列表 -->
        <ProfileTabFans
          v-else-if="activeTab === '粉丝'"
          :follow-list="followList"
          @sidebar-click="handleSidebarClick"
          @follow-user="handleFollowUser"
        />

        <!-- 兴趣画像 -->
        <div v-else-if="activeTab === '兴趣画像'" class="interests-section">
          <InterestsPanel :user-id="userId" :is-own-space="isOwnSpace" />
        </div>

        <!-- 设置 -->
        <ProfileTabSettings
          v-else-if="activeTab === '设置'"
          :is-own-space="isOwnSpace"
        />

        <!-- 其他标签内容占位符 -->
        <div v-else class="content-placeholder">
          <h2>{{ activeTab }}内容区域</h2>
          <p>此区域将显示{{ activeTab }}相关内容</p>
        </div>
      </div>

      <!-- 右侧内容区域 - 投稿、合集和列表、收藏、关注、粉丝页面不显示 -->
      <div v-if="activeTab !== '投稿' && activeTab !== '合集和列表' && activeTab !== '收藏' && activeTab !== '关注' && activeTab !== '粉丝' && activeTab !== '兴趣画像'" class="side-content">
        <ProfileSidebar
          :user-info="userInfo"
          :user-id="userId"
          :is-own-space="isOwnSpace"
          :active-tab="activeTab"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 个人主页整体样式 */
.user-profile-page {
  width: 100%;
  min-height: 100vh;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* 选择栏容器 */
.profile-tabs-container {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto 20px;
}

/* 选择栏包装器 */
.profile-tabs-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
  background-color: #fff;
}

/* 选择栏 */
.profile-tabs {
  display: flex;
  align-items: center;
  gap: 20px;
  background-color: #fff;
}

/* 标签项 */
.tab-item {
  padding: 8px 16px;
  font-size: 16px;
  color: #666;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: #fff;
}

.tab-item:hover {
  color: #00aeec;
  background-color: rgba(0, 174, 236, 0.1);
}

.tab-item.active {
  color: #00aeec;
  background-color: rgba(0, 174, 236, 0.1);
  font-weight: 500;
  border-bottom: 2px solid #00aeec;
}

/* 搜索栏包装器 */
.search-bar-wrapper {
  margin-left: 30px;
}

/* 搜索输入包装器 */
.search-input-wrapper {
  display: flex;
  align-items: center;
  background-color: #f0f2f5;
  border-radius: 20px;
  padding: 0 15px;
  height: 36px;
  width: 200px;
  transition: all 0.3s ease;
}

.search-input-wrapper:hover {
  background-color: #e0e2e5;
}

/* 搜索图标 */
.search-icon {
  color: #9499a0;
  font-size: 16px;
  margin-right: 8px;
}

/* 搜索输入框 */
.search-input {
  flex: 1;
  border: none;
  outline: none;
  background-color: transparent;
  font-size: 14px;
  color: #333;
}

/* 统计数据容器 */
.stats-container {
  display: flex;
  gap: 0;
}

/* 统计项 */
.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 56px;
  padding: 6px 10px;
  border-radius: 8px;
}

.stat-item:hover .stat-label {
  color: #00aeec;
}

.stat-item:hover .stat-value {
  color: #00aeec;
}

.stat-entry-button:hover {
  background-color: #f1f9ff;
}

/* 非交互式统计项样式 */
.stat-item.non-interactive {
  cursor: default;
}

.stat-item.non-interactive:hover .stat-label {
  color: #9499a0;
}

.stat-item.non-interactive:hover .stat-value {
  color: #333;
}

/* 统计标签 */
.stat-label {
  font-size: 12px;
  color: #9499a0;
  transition: color 0.3s ease;
}

/* 统计数值 */
.stat-value {
  font-size: 16px;
  font-weight: 400;
  color: #333;
  transition: color 0.3s ease;
}

/* 主内容区域 */
.profile-content {
  flex: 1;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 0 40px;
  box-sizing: border-box;
  display: flex;
  gap: 20px;
  width: 100%;
}

/* 左侧主内容 */
.main-content {
  flex: 1;
  min-width: 0;
}

/* 右侧边栏 */
.side-content {
  width: 300px;
  min-width: 300px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 内容占位符 */
.content-placeholder {
  background-color: #fff;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  text-align: center;
  color: #666;
}

.content-placeholder h2 {
  margin-bottom: 10px;
  color: #333;
}

/* 响应式设计 */





</style>