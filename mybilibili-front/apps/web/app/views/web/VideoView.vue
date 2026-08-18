<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import { interactionApi, userApi, videoApi } from '@/api/client'
import { recommendApi } from '@/api/recommend.ts'
import { Message } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AiAssistantPanel from '@/components/AiAssistantPanel.vue'
import LevelBadge from '@/components/LevelBadge.vue'
import UserFloatCard from '@/components/UserFloatCard.vue'
import VideoCommentSection from '@/components/VideoCommentSection.vue'
import VideoDescription from '@/components/VideoDescription.vue'
import VideoInteractionBar from '@/components/VideoInteractionBar.vue'
import VideoPlayer from '@/components/VideoPlayer.vue'
import VideoReportDialog from '@/components/VideoReportDialog.vue'
import VideoSidebar from '@/components/VideoSidebar.vue'

const route = useRoute()
const router = useRouter()

// 定义props - 从路由接收manuscriptId和p参数
const props = defineProps({
  manuscriptId: {
    type: String,
    required: true
  },
  p: {
    type: Number,
    default: 1
  }
})

// 当前稿件ID和分P - 从路由参数获取，确保响应式更新
const currentManuscriptId = ref(Number.parseInt(String(route.params.id)))
const currentP = ref(Number.parseInt(String(route.query.p)) || 1)
const resumeTime = ref(Number.parseInt(String(route.query.t)) || 0)

// 兼容旧代码 - videoId用于某些API调用
const videoId = ref(null)

// 稿件信息
const manuscriptInfo = ref({
  id: null,
  title: '',
  description: '',
  coverUrl: '',
  tags: [],
  videos: []
})

// 当前播放的视频分P索引
const currentVideoIndex = ref(0)

const videoInfo = ref({
  title: '',
  coverUrl: '',
  playUrl: '',
  playUrlHd: '',
  playUrlSd: '',
  playUrlLd: '',
  sourceVideoUrl: '',
  uploader: {
    id: null as any,
    name: '',
    avatar: '',
    bio: '',
    level: 0,
    followerCount: 0,
    followingCount: 0,
    likeCount: 0
  },
  viewCount: 0,
  likeCount: 0,
  dislikeCount: 0,
  coinCount: 0,
  collectCount: 0,
  shareCount: 0,
  commentCount: 0,
  duration: '00:00',
  uploadTime: '',
  description: '',
  watchingCount: 0,
  danmuLoadedCount: 0,
  tags: []
})

// 从路由读取续播时间
const getResumeTimeFromRoute = () => {
  const t = Number.parseInt(String(route.query.t))
  return Number.isFinite(t) && t > 0 ? t : 0
}

// 格式化日期
const formatDate = dateStr => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 弹幕列表
const danmuList = ref([])
// 弹幕加载状态
const loadingDanmus = ref(false)

// 弹幕列表折叠状态
const isDanmuListCollapsed = ref(false)

// 视频分P列表折叠状态
const isVideoPartsCollapsed = ref(false)

// 处理发消息
const handleSendMessage = () => {
  const token = safeStorage.getItem("token")
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }

  // 不能给自己发消息
  if (currentUser.value && currentUser.value.id === videoInfo.value.uploader.id) {
    ElMessage.warning('不能给自己发送消息')
    return
  }

  // 跳转到消息页面，带上对方用户ID
  router.push(`/message/private?userId=${videoInfo.value.uploader.id}`)
}

// 相关视频
const relatedVideos = ref([])
const loadingRelatedVideos = ref(false)

// 浏览历史记录相关
const watchProgress = ref(0)
const videoDuration = ref(0)

// 加载相关视频
const loadRelatedVideos = async () => {
  if (!videoId.value) return

  loadingRelatedVideos.value = true
  try {
    const response = await recommendApi.getRelatedVideos(videoId.value, 8)
    if (response.code === 200) {
      relatedVideos.value = (response.data || []).map(video => ({
        id: video.videoId,
        manuscriptId: video.manuscriptId,
        title: video.title,
        cover: video.coverUrl,
        author: video.userName,
        authorId: video.userId,
        viewCount: video.viewCount || 0,
        commentCount: video.commentCount || 0,
        duration: video.duration || '00:00'
      }))
    }
  } catch (error) {
    console.error('加载相关视频失败:', error)
  } finally {
    loadingRelatedVideos.value = false
  }
}

// 同步记录浏览历史（使用 fetch keepalive，用于页面离开时可靠发送）
const recordWatchHistorySync = () => {
  if (!videoId.value) return

  const progress = Math.floor(watchProgress.value || 0)
  let duration = Math.floor(videoDuration.value || 0)

  if (duration <= 0 && manuscriptInfo.value.videos && manuscriptInfo.value.videos.length > 0) {
    const currentVideo = manuscriptInfo.value.videos[currentVideoIndex.value]
    if (currentVideo && currentVideo.durationSeconds) {
      duration = currentVideo.durationSeconds
    }
  }

  const watchRatio = duration > 0 ? progress / duration : 0
  console.log('离开页面，记录最终播放进度', { progress, duration, watchRatio: (watchRatio * 100).toFixed(1) + '%' })

  try {
    const token = safeStorage.getItem("token")
    if (!token) return
    const params = new URLSearchParams({
      videoId: String(videoId.value),
      progressSeconds: String(progress),
      videoDuration: String(duration || progress || 0)
    })
    const url = `/api/watch-history?${params.toString()}`
    fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      keepalive: true
    }).catch(() => {})
  } catch (error) {
    console.error('记录浏览历史失败:', error)
  }
}

// 处理视频时间更新，仅用于更新进度变量（由 VideoPlayer 的 timeUpdate 事件驱动）
const handleVideoTimeUpdate = ({ currentTime, duration }) => {
  watchProgress.value = currentTime || 0
  videoDuration.value = duration || 0
}

// 视频播放器引用
const videoPlayerRef = ref(null)

// 跳转到弹幕时间（转发给 VideoPlayer）
const jumpToDanmuTime = time => {
  if (videoPlayerRef.value) {
    videoPlayerRef.value.seekTo(time)
  }
}

// 切换弹幕列表折叠状态
const toggleDanmuList = () => {
  isDanmuListCollapsed.value = !isDanmuListCollapsed.value
}

// 切换视频分P列表折叠状态
const toggleVideoParts = () => {
  isVideoPartsCollapsed.value = !isVideoPartsCollapsed.value
}

// 关注相关状态
const isFollowing = ref(false)
const followerCount = ref(0)
const loadingFollow = ref(false)

// 当前用户信息
const currentUser = ref(JSON.parse(safeStorage.getItem('user') || 'null'))

// 浮动用户卡片相关
const showUserFloatCard = ref(false)
const authorAvatarRef = ref(null)
const floatCardTimer = ref(null)
const authorBridgeRef = ref(null)

// 处理鼠标进入作者区域（头像+桥接区域）
const handleAuthorMouseEnter = () => {
  if (floatCardTimer.value) {
    clearTimeout(floatCardTimer.value)
    floatCardTimer.value = null
  }
  showUserFloatCard.value = true
}

// 处理鼠标离开作者区域
const handleAuthorMouseLeave = () => {
  floatCardTimer.value = setTimeout(() => {
    showUserFloatCard.value = false
  }, 300)
}

// 处理作者浮动卡片鼠标进入
const handleAuthorCardMouseEnter = () => {
  if (floatCardTimer.value) {
    clearTimeout(floatCardTimer.value)
    floatCardTimer.value = null
  }
}

// 处理作者浮动卡片鼠标离开
const handleAuthorCardMouseLeave = () => {
  floatCardTimer.value = setTimeout(() => {
    showUserFloatCard.value = false
  }, 300)
}

// 处理浮动卡片关注状态变化
const handleFollowChange = ({ userId, isFollowing: newFollowStatus }) => {
  if (videoInfo.value.uploader.id === userId) {
    isFollowing.value = newFollowStatus
    if (newFollowStatus) {
      followerCount.value++
    } else {
      followerCount.value = Math.max(0, followerCount.value - 1)
    }
  }
}

// 跳转到视频详情页
const goToVideo = video => {
  if (typeof video === 'object' && video.manuscriptId) {
    window.location.href = `/manuscript/${video.manuscriptId}`
  } else if (typeof video === 'object' && video.id) {
    window.location.href = `/manuscript/${video.id}`
  } else {
    window.location.href = `/manuscript/${video}`
  }
}

// 跳转到标签搜索页面
const goToTagSearch = (tagName: string) => {
  router.push({ path: '/search', query: { tag: tagName } })
}

// 切换分P视频
const switchVideoPart = index => {
  if (index === currentVideoIndex.value || !manuscriptInfo.value.videos[index]) return

  const query = new URLSearchParams({ p: String(index + 1) })
  const currentResumeTime = getResumeTimeFromRoute()
  if (currentResumeTime > 0) {
    query.set('t', String(currentResumeTime))
  }

  const newUrl = `/manuscript/${currentManuscriptId.value}?${query.toString()}`
  window.location.href = newUrl
}

// 加载互动状态
const loadInteractionStatus = async () => {
  console.log('=== 开始获取互动状态 ===')
  console.log('当前 manuscriptId:', currentManuscriptId.value)
  try {
    const token = safeStorage.getItem("token")
    const user = JSON.parse(safeStorage.getItem('user') || 'null')
    console.log('token:', token ? '存在' : '不存在')
    console.log('当前用户信息:', user)

    if (token && currentManuscriptId.value) {
      console.log('正在调用 getInteractionStatus API...')
      const statusResponse = await interactionApi.getInteractionStatus(currentManuscriptId.value)
      console.log('API 完整响应:', statusResponse)

      if (statusResponse.code === 200) {
        console.log('API 返回的数据:', statusResponse.data)

        interactionStatus.value = {
          liked: statusResponse.data.isLiked || statusResponse.data.liked || false,
          favorited: statusResponse.data.isCollected || statusResponse.data.collected || false,
          coined: statusResponse.data.coined || statusResponse.data.coinCount > 0 || false,
          shared: statusResponse.data.shared || false,
          coinCount: statusResponse.data.coinCount || 0
        }
        console.log('设置后的 interactionStatus:', interactionStatus.value)
      } else {
        console.error('API 返回错误代码:', statusResponse.code)
        console.error('错误消息:', statusResponse.message)
      }
    } else {
      console.log('未登录或manuscriptId为空，跳过获取互动状态')
    }
  } catch (error) {
    console.error('=== 获取互动状态异常 ===')
    console.error('错误对象:', error)
    console.error('错误消息:', error.message)
    if (error.response) {
      console.error('响应状态:', error.response.status)
      console.error('响应数据:', error.response.data)
    }
  }
  console.log('=== 互动状态获取结束 ===')
}

// 跳转到作者主页
const goToAuthor = authorId => {
  window.open(`/profile/${authorId}/home`, '_blank')
}

// 互动状态
const interactionStatus = ref({
  liked: false,
  favorited: false,
  shared: false,
  coined: false,
  coinCount: 0
})

// AI助手弹窗状态
const showAiAssistantDialog = ref(false)

const reportDialogVisible = ref(false)

// 处理关注/取消关注
const handleFollow = async () => {
  // 检查是否登录
  const token = safeStorage.getItem("token")
  if (!token) {
    // 未登录，显示登录弹窗
    ElMessage.warning('请先登录')
    return
  }

  // 检查是否是自己
  if (currentUser.value && currentUser.value.id === videoInfo.value.uploader.id) {
    // 不能关注自己
    ElMessage.warning('无法关注自己')
    return
  }

  // 防止重复点击
  if (loadingFollow.value) {
    console.log('正在处理中，请勿重复点击')
    return
  }

  // 如果是取消关注，先弹出确认框
  if (isFollowing.value) {
    try {
      await ElMessageBox.confirm(`确定不再关注 ${videoInfo.value.uploader.name} 吗？`, '取消关注', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      // 用户取消操作
      return
    }
  }

  try {
    loadingFollow.value = true
    console.log('【前端调试】当前关注状态:', isFollowing.value)
    console.log('【前端调试】准备调用关注API，目标用户ID:', videoInfo.value.uploader.id)

    // 根据当前关注状态决定调用关注还是取消关注接口
    const willFollow = !isFollowing.value
    const response = await userApi.follow(videoInfo.value.uploader.id, willFollow)
    console.log('【前端调试】关注API响应:', response)

    if (response.code === 200) {
      // 根据前端预期的操作结果更新状态
      isFollowing.value = willFollow
      if (willFollow) {
        followerCount.value++
        ElMessage.success('关注成功')
      } else {
        followerCount.value = Math.max(0, followerCount.value - 1)
        ElMessage.success('取消关注成功')
      }
    } else {
      ElMessage.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('关注操作失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  } finally {
    loadingFollow.value = false
  }
}

onMounted(async () => {
  // 检查token是否存在
  const token = safeStorage.getItem("token")
  console.log('【前端调试】localStorage中的token:', token ? '存在' : '不存在')
  console.log('【前端调试】localStorage中的user:', safeStorage.getItem('user'))
  console.log('【前端调试】当前路由参数:', route.params)
  console.log('【前端调试】当前路由查询:', route.query)

  // 从路由参数重新获取 manuscriptId（确保正确）
  const manuscriptIdFromRoute = Number.parseInt(String(route.params.id))
  const pFromRoute = Number.parseInt(String(route.query.p)) || 1
  console.log('【前端调试】从路由获取的 manuscriptId:', manuscriptIdFromRoute, 'p:', pFromRoute)

  if (!manuscriptIdFromRoute || isNaN(manuscriptIdFromRoute)) {
    console.error('【前端调试】manuscriptId 无效:', route.params.id)
    ElMessage.error('稿件ID无效')
    return
  }

  // 更新本地状态
  currentManuscriptId.value = manuscriptIdFromRoute
  currentP.value = pFromRoute
  resumeTime.value = getResumeTimeFromRoute()

  try {
    // 使用新的API获取视频数据
    console.log('【前端调试】开始获取视频详情，manuscriptId:', currentManuscriptId.value, 'p:', currentP.value)
    console.log('【前端调试】请求将携带Authorization header:', token ? '是' : '否')
    const videoResponse = await videoApi.getVideoByManuscriptId(currentManuscriptId.value, { p: currentP.value })
    console.log('【前端调试】视频详情响应:', videoResponse)

    if (videoResponse.code === 200) {
      const data = videoResponse.data

      // 从videos数组中获取当前分P的视频信息
      const videos = data.videos || []
      const videoIndex = currentP.value - 1 // 分P从1开始，数组索引从0开始
      const currentVideo = videos[videoIndex] || videos[0]

      // 从当前视频获取videoId - 这是关键！
      videoId.value = currentVideo?.id || null
      console.log('【前端调试】设置 videoId:', videoId.value, '当前分P:', currentP.value)

      // 更新当前视频索引
      currentVideoIndex.value = videoIndex >= 0 && videoIndex < videos.length ? videoIndex : 0

      // 更新稿件信息
      manuscriptInfo.value = {
        id: currentManuscriptId.value,
        title: data.title,
        description: data.description || '',
        coverUrl: data.coverUrl,
        tags: data.tags || [],
        videos: videos
      }

      // 更新视频信息 - 从当前分P视频获取播放地址
      videoInfo.value = {
        title: currentVideo?.title || data.title,
        coverUrl: data.coverUrl,
        playUrl: currentVideo?.playUrl || '',
        playUrlHd: currentVideo?.playUrlHd || '',
        playUrlSd: currentVideo?.playUrlSd || '',
        playUrlLd: currentVideo?.playUrlLd || '',
        sourceVideoUrl: currentVideo?.sourceVideoUrl || '',
        uploader: {
          name: data.uploader?.name || '',
          avatar: data.uploader?.avatar || '',
          id: data.uploader?.id || '',
          bio: data.uploader?.signature || data.uploader?.bio || '',
          level: data.uploader?.level || 0,
          followerCount: data.uploader?.followerCount || 0,
          followingCount: data.uploader?.followingCount || 0,
          likeCount: data.uploader?.likedCount || 0
        },
        viewCount: data.viewCount || 0,
        likeCount: data.likeCount || 0,
        dislikeCount: 0,
        coinCount: data.coinCount || 0,
        collectCount: data.collectCount || 0,
        shareCount: data.shareCount || 0,
        commentCount: data.commentCount || 0,
        duration: currentVideo?.durationSeconds
          ? `${Math.floor(currentVideo.durationSeconds / 60)
              .toString()
              .padStart(2, '0')}:${(currentVideo.durationSeconds % 60).toString().padStart(2, '0')}`
          : '00:00',
        uploadTime: data.uploadTime,
        description: data.description || '',
        watchingCount: 0,
        danmuLoadedCount: 0,
        tags: data.tags || []
      }

      // 初始化粉丝数 - 使用后端实时计算的粉丝数
      followerCount.value = data.uploader?.followerCount || 0
      console.log('【前端调试】后端返回的uploader:', data.uploader)
      console.log('【前端调试】后端返回的following:', data.uploader?.following)
      console.log('【前端调试】后端返回的followerCount:', data.uploader?.followerCount)
      console.log('【前端调试】后端返回的signature:', data.uploader?.signature)
      console.log('【前端调试】后端返回的bio:', data.uploader?.bio)

      // 使用后端返回的关注状态（如果用户已登录）
      if (data.uploader?.following != null) {
        isFollowing.value = data.uploader.following
        console.log('【前端调试】设置关注状态为:', isFollowing.value)
      } else {
        console.log('【前端调试】following为null或undefined，保持默认值false')
      }
      console.log('视频信息获取成功:', videoInfo.value)

      // 在设置 videoId 后，再调用依赖 videoId 的函数
      // 获取视频互动状态
      await loadInteractionStatus()

      // 加载相关视频推荐
      await loadRelatedVideos()

      // 不在打开时记录，只在退出时记录浏览历史
    } else {
      console.error('【前端调试】API 返回错误:', videoResponse)
      ElMessage.error(videoResponse.message || '获取视频信息失败')
    }
  } catch (error) {
    console.error('获取视频详情失败:', error)
    ElMessage.error('获取视频信息失败')
  }

  // 添加页面离开时记录浏览历史的监听
  window.addEventListener('beforeunload', recordWatchHistorySync)

  console.log('视频信息加载完成，播放器初始化由 VideoPlayer 组件负责')
})

onUnmounted(() => {
  // 记录最终播放进度
  recordWatchHistorySync()

  // 移除页面离开监听
  window.removeEventListener('beforeunload', recordWatchHistorySync)
})

// 监听稿件ID变化
watch(
  () => route.params.id,
  newId => {
    if (newId) {
      currentManuscriptId.value = Number.parseInt(String(newId))
    }
  }
)

// 监听路由参数变化，处理浏览器前进/后退
watch(
  () => [route.query.p, route.query.t],
  ([newP, newT]) => {
    const p = Number.parseInt(String(newP)) || 1
    const t = Number.parseInt(String(newT)) || 0
    resumeTime.value = t > 0 ? t : 0
    if (p !== currentP.value && manuscriptInfo.value.videos.length > 0) {
      // 切换到对应的分P（注意：p是1-based，index是0-based）
      const index = p - 1
      if (index >= 0 && index < manuscriptInfo.value.videos.length) {
        // 只更新数据和播放器，不修改URL（因为URL已经变了）
        currentP.value = p
        currentVideoIndex.value = index

        const video = manuscriptInfo.value.videos[index]

        // 更新videoId用于其他API调用
        videoId.value = video.id

        // 更新视频信息（保留稿件标题，只更新播放地址和时长）
        videoInfo.value.playUrl = video.playUrl || ''
        videoInfo.value.playUrlHd = video.playUrlHd || ''
        videoInfo.value.playUrlSd = video.playUrlSd || ''
        videoInfo.value.playUrlLd = video.playUrlLd || ''
        videoInfo.value.sourceVideoUrl = video.sourceVideoUrl || ''
        videoInfo.value.duration = video.duration || '00:00'

        // 重新获取互动状态（播放器切换、弹幕/字幕重载由 VideoPlayer 播放 props 自行处理）
        loadInteractionStatus()
      }
    }
  }
)
</script>

<template>
  <div class="video-container">
    <div class="main-content">
      <!-- 顶部区域：视频标题和作者信息 -->
      <div class="top-section">
        <!-- 左侧：视频标题和统计信息 -->
        <div class="video-header">
          <h1 class="video-title">{{ manuscriptInfo.title || videoInfo.title }}</h1>
          <div class="video-stats">
            <span>{{ (videoInfo.viewCount || 0).toLocaleString() }}次播放</span>
            <span>{{ (videoInfo.commentCount || 0) }}条评论</span>
            <span>{{ formatDate(videoInfo.uploadTime) }}</span>
          </div>
        </div>
        
        <!-- 右侧：作者信息 -->
        <div class="author-card"
          @mouseenter="handleAuthorMouseEnter"
          @mouseleave="handleAuthorMouseLeave"
        >
          <img loading="lazy" decoding="async" 
            ref="authorAvatarRef"
            :src="videoInfo.uploader.avatar || '/default-avatar.svg'" 
            alt="作者头像" 
            class="author-avatar" 
            @click="goToAuthor(videoInfo.uploader.id)"
          >
          <!-- 桥接区域：连接头像和浮动卡片 -->
          <div ref="authorBridgeRef" class="float-card-bridge"></div>
          <div class="author-meta">
            <div class="author-info-top">
              <span class="author-name" @click="goToAuthor(videoInfo.uploader.id)">{{ videoInfo.uploader.name }}</span>
              <LevelBadge :level="videoInfo.uploader.level" />
              <el-button 
                text 
                size="small" 
                class="message-btn"
                @click="handleSendMessage"
                :disabled="currentUser && currentUser.id === videoInfo.uploader.id"
              >
                <el-icon><Message /></el-icon>
                <span>发消息</span>
              </el-button>
            </div>
            <span class="author-bio" :title="videoInfo.uploader.bio">{{ videoInfo.uploader.bio || '该用户暂无简介' }}</span>
            <el-button 
              :type="isFollowing ? 'default' : 'primary'" 
              class="follow-btn"
              @click="handleFollow"
              :loading="loadingFollow"
              :disabled="loadingFollow || (currentUser && currentUser.id === videoInfo.uploader.id)"
            >
              {{ isFollowing ? '已关注' : '+ 关注' }} {{ (followerCount || 0).toLocaleString() }}
            </el-button>
          </div>
        </div>
      </div>
      
      <!-- 中间区域：视频播放器和弹幕列表 -->
      <div class="player-danmu-container">
        <!-- 左侧内容 -->
        <div class="left-section">
          <!-- 视频播放器 -->
          <VideoPlayer
            ref="videoPlayerRef"
            :current-manuscript-id="currentManuscriptId"
            :manuscript-info="manuscriptInfo"
            :video-info="videoInfo"
            :current-p="currentP"
            :current-video-index="currentVideoIndex"
            :resume-time="resumeTime"
            :danmu-list="danmuList"
            :loading-danmus="loadingDanmus"
            @update:video-info="videoInfo = $event"
            @update:danmu-list="danmuList = $event"
            @update:loading-danmus="loadingDanmus = $event"
            @time-update="handleVideoTimeUpdate"
          />
          
          <VideoInteractionBar
            :manuscript-id="currentManuscriptId"
            :video-info="videoInfo"
            :interaction-status="interactionStatus"
            @update:video-info="videoInfo = $event"
            @update:interaction-status="interactionStatus = $event"
            @ai-assistant="showAiAssistantDialog = true"
            @report="reportDialogVisible = true"
          />
          
          <VideoDescription
            :description="videoInfo.description"
            :tags="videoInfo.tags"
            @tag-search="goToTagSearch"
          />
          
          <VideoCommentSection
            :manuscript-id="currentManuscriptId"
            :comment-count="videoInfo.commentCount"
            :uploader-id="videoInfo.uploader.id"
            @update:comment-count="videoInfo.commentCount = $event"
          />
        </div>
        
        <VideoSidebar
        :danmu-list="danmuList"
        :loading-danmus="loadingDanmus"
        :manuscript-info="manuscriptInfo"
        :current-video-index="currentVideoIndex"
        :related-videos="relatedVideos"
        :loading-related-videos="loadingRelatedVideos"
        :is-danmu-list-collapsed="isDanmuListCollapsed"
        :is-video-parts-collapsed="isVideoPartsCollapsed"
        @toggle-danmu-list="toggleDanmuList"
        @toggle-video-parts="toggleVideoParts"
        @jump-to-danmu-time="jumpToDanmuTime"
        @switch-video-part="switchVideoPart"
        @go-to-video="goToVideo"
        @go-to-author="goToAuthor"
      />
      </div>
        

        

    </div>
  </div>
  
  <VideoReportDialog
    v-model:visible="reportDialogVisible"
    :manuscript-id="currentManuscriptId"
  />

  <!-- 浮动用户卡片 - 作者头像 -->
  <UserFloatCard
    v-model:visible="showUserFloatCard"
    :trigger-ref="authorAvatarRef"
    :bridge-ref="authorBridgeRef"
    :user-info="{
      id: videoInfo.uploader.id,
      name: videoInfo.uploader.name,
      avatar: videoInfo.uploader.avatar,
      bio: videoInfo.uploader.bio,
      signature: videoInfo.uploader.bio,
      following: isFollowing,
      followerCount: followerCount,
      followingCount: videoInfo.uploader.followingCount || 0,
      likeCount: videoInfo.uploader.likeCount || 0,
      level: videoInfo.uploader.level || 0
    }"
    placement="bottom"
    @follow-change="handleFollowChange"
    @mouseenter="handleAuthorCardMouseEnter"
    @mouseleave="handleAuthorCardMouseLeave"
  />
  
  <!-- AI助手侧边面板 -->
  <AiAssistantPanel
    v-model:visible="showAiAssistantDialog"
    :video-id="videoId"
    :video-title="videoInfo.title"
  />
</template>

<style scoped>
.video-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  padding-top: 0;
  background-color: #fff;
  min-height: 100vh;
}

/* 浮动卡片桥接区域 - 连接头像和浮动卡片，防止鼠标断开 */
.float-card-bridge {
  position: fixed;
  background: transparent;
  z-index: 2999;
  pointer-events: auto;
}

/* 收藏夹弹窗样式 */
.favorite-folders {
  max-height: 300px;
  overflow-y: auto;
  padding: 10px 0;
}

.folder-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.folder-count {
  color: #999;
  font-size: 12px;
}

.new-folder-section {
  margin-top: 20px;
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
}

.new-folder-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  cursor: pointer;
  color: #23ade5;
  font-size: 14px;
}

.new-folder-btn:hover {
  color: #1a91d0;
}

.new-folder-input {
  display: flex;
  align-items: center;
  margin-top: 10px;
}


/* 主内容区域 */
.main-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

/* 顶部区域：视频标题和作者信息 */
.top-section {
  display: flex;
  gap: 40px;
  align-items: flex-start;
  width: 100%;
}

/* 左侧：视频标题和统计信息 */
.video-header {
  flex: 1;
  background-color: #fff;
  padding: 10px 0;
  margin-top: 20px;
}

.video-header .video-title {
  font-size: 28px;
  font-weight: normal;
  color: #333;
  margin-bottom: 5px;
  line-height: 1.4;
}

.video-header .video-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #666;
  font-weight: normal;
}

/* 右侧：作者信息 */
.author-card {
  width: 350px;
  flex-shrink: 0;
  background-color: #fff;
  padding: 10px 20px;
  display: flex;
  align-items: flex-start;
  gap: 15px;
}

/* 中间区域：视频播放器和弹幕列表 */
.player-danmu-container {
  display: flex;
  gap: 40px;
  width: 100%;
  align-items: flex-start;
}

/* 左侧内容 */
.left-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* 字幕设置面板和播放器样式已移至 VideoPlayer.vue */

/* 右侧弹幕列表 */
.side-danmu-list {
  width: 350px;
  flex-shrink: 0;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.side-danmu-list .danmu-list-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  background-color: #f9f9f9;
}

.side-danmu-list .danmu-list-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.side-danmu-list .danmu-items {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.side-danmu-list .danmu-items.is-hidden {
  display: none;
}

/* 视频分P列表 */
.video-parts-section {
  width: 350px;
  flex-shrink: 0;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  margin-top: 16px;
}

.video-parts-section .video-parts-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  background-color: #f9f9f9;
}

.video-parts-section .video-parts-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.video-parts-section .video-parts-count {
  font-size: 12px;
  color: #999;
  margin-right: auto;
  margin-left: 8px;
}

.video-parts-section .video-parts-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 8px;
}

.video-parts-section .video-parts-list.is-hidden {
  display: none;
}

.video-parts-section .video-part-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  gap: 10px;
}

.video-parts-section .video-part-item:hover {
  background-color: #f5f5f5;
}

.video-parts-section .video-part-item.active {
  background-color: #e3f2fd;
  color: #00a1d6;
}

.video-parts-section .video-part-item .part-index {
  font-size: 12px;
  color: #999;
  min-width: 30px;
}

.video-parts-section .video-part-item.active .part-index {
  color: #00a1d6;
}

.video-parts-section .video-part-item .part-title {
  flex: 1;
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-parts-section .video-part-item.active .part-title {
  color: #00a1d6;
}

.video-parts-section .video-part-item .part-duration {
  font-size: 12px;
  color: #999;
}

/* 视频状态栏样式已移至 VideoPlayer.vue */

/* 互动按钮栏 */
.interaction-bar {
  background-color: #fff;
  padding: 8px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

/* 视频简介 */
.video-description {
  background-color: #fff;
  padding: 20px 0;
}

.video-description .description-content {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  margin-bottom: 10px;
  transition: all 0.3s ease;
}

.video-description .description-content.is-collapsed {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.video-description .description-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #00a1d6;
  transition: all 0.3s ease;
}

.video-description .description-toggle:hover {
  color: #0091c6;
}

.video-description .description-toggle .el-icon {
  transition: transform 0.3s ease;
  font-size: 14px;
}

.video-description .description-toggle .el-icon.is-rotated {
  transform: rotate(180deg);
}

/* 视频标签栏 */
.video-tags {
  background-color: #fff;
  padding: 10px 0 20px 0;
  border-bottom: 1px solid #f0f0f0;
}

.video-tags .tags-header {
  margin-bottom: 10px;
}

.video-tags .tags-header h4 {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.video-tags .tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.video-tags .tag-item {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 16px;
  color: #666;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  user-select: none;
}

.video-tags .tag-item:hover {
  background-color: #e6f7ff;
  border-color: #00a1d6;
  color: #00a1d6;
}

.interaction-bar .left-actions {
  display: flex;
  gap: 15px;
}

.interaction-bar .right-actions {
  display: flex;
  gap: 10px;
}

.interaction-bar .action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 6px 16px;
  border: none;
  background-color: #fff;
  color: #666;
  font-size: 14px;
  border-radius: 6px;
  transition: all 0.3s ease;
  min-width: 80px;
  min-height: 32px;
}

.interaction-bar .action-btn:hover {
  background-color: #f5f5f5;
  color: #00aeec;
}

.interaction-bar .action-btn.is-active {
  color: #00aeec;
  font-weight: 500;
}

.interaction-bar .action-btn.is-active:hover {
  background-color: #e6f7ff;
}

.interaction-bar .action-btn.is-animating {
  animation: likeAnimation 0.3s ease;
}

@keyframes likeAnimation {
  0% { transform: scale(1); }
  50% { transform: scale(1.3); }
  100% { transform: scale(1); }
}

.interaction-bar .action-btn .el-icon {
  font-size: 18px;
}

.interaction-bar .action-btn span {
  font-size: 14px;
}

.interaction-bar .ai-assistant-btn {
  gap: 8px;
}

.interaction-bar .more-btn {
  padding: 6px 12px;
  min-width: 40px;
}

/* 弹幕加载状态 */
.danmu-list .loading-danmus {
  padding: 20px 0;
}

/* 无弹幕状态 */
.danmu-list .no-danmus {
  text-align: center;
  padding: 40px 0;
  color: #999;
  font-size: 14px;
}

/* 右侧内容 */
.right-section {
  width: 350px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 作者卡片 */
.author-card {
  background-color: #fff;
  padding: 20px 0;
  display: flex;
  align-items: center;
  gap: 15px;
}

.author-card .author-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
}

.author-card .author-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 0;
  margin-top: 0;
  padding-top: 0;
}

.author-card .author-info-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.author-card .author-name {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  flex: 1;
}

.author-card .message-btn {
  color: #666;
  font-size: 13px;
  padding: 4px 8px;
  border-radius: 4px;
  border: none;
  background-color: transparent;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 4px;
}

.author-card .message-btn:hover {
  color: #00a1d6;
}

.author-card .message-btn .el-icon {
  font-size: 14px;
}

.author-card .author-bio {
  font-size: 13px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

/* 关注按钮 */
.follow-btn {
  background-color: #00a1d6;
  border-color: #00a1d6;
  color: #fff;
  border-radius: 4px;
  padding: 8px 20px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.follow-btn:hover {
  background-color: #0091c6;
  border-color: #0091c6;
}

/* 弹幕列表 */
.danmu-list {
  background-color: #fff;
  padding: 20px 0;
}

/* 弹幕列表头部 */
.danmu-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  padding: 5px 0;
  transition: background-color 0.3s ease;
}

.danmu-list-header:hover {
  background-color: #f5f5f5;
}

.danmu-list-header h3 {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

/* 折叠图标 */
.collapse-icon {
  transition: transform 0.3s ease;
  font-size: 16px;
  color: #666;
}

.collapse-icon.is-collapsed {
  transform: rotate(-90deg);
}

/* 弹幕列表内容 */
.danmu-items {
  display: flex;
  flex-direction: column;
  gap: 0;
  max-height: 400px;
  overflow-y: auto;
  transition: all 0.3s ease;
  margin: 0;
}

/* 弹幕发送区域 */
.danmu-send-section {
  padding: 12px;
  background-color: #f9f9f9;
  border-bottom: 1px solid #f0f0f0;
}

.danmu-send-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background-color: #00a1d6;
  color: #fff;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 14px;
}

.danmu-send-trigger:hover {
  background-color: #0091c6;
}

.danmu-input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.danmu-color-picker {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.color-option {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
}

.color-option:hover {
  transform: scale(1.2);
}

.danmu-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.danmu-input-row .el-input {
  flex: 1;
}

.danmu-list-content {
  flex: 1;
  overflow-y: auto;
}

/* 滚动条样式 */
.danmu-items::-webkit-scrollbar {
  width: 6px;
}

.danmu-items::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.danmu-items::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.danmu-items::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}

/* 表头样式 */
.danmu-header {
  display: flex;
  gap: 10px;
  padding: 8px 16px;
  background-color: #f5f5f5;
  border-radius: 0;
  font-size: 13px;
  font-weight: 600;
  color: #666;
  position: sticky;
  top: 0;
  z-index: 10;
}

.danmu-header .header-time {
  min-width: 50px;
  text-align: left;
}

.danmu-header .header-content {
  flex: 1;
  text-align: left;
}

.danmu-header .header-send-time {
  min-width: 140px;
  text-align: right;
}

.danmu-items.is-hidden {
  display: none;
}

.danmu-item {
  display: flex;
  gap: 10px;
  font-size: 14px;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.danmu-item:hover {
  background-color: #f0f0f0;
}

.danmu-item:active {
  background-color: #e0e0e0;
  transform: scale(0.98);
}

.danmu-time {
  color: #999;
  min-width: 50px;
  font-weight: 500;
  text-align: left;
}

.danmu-text {
  color: #333;
  flex: 1;
  text-align: left;
}

.danmu-send-time {
  color: #999;
  min-width: 140px;
  font-size: 12px;
  text-align: right;
}

/* 推荐视频 */
.related-videos {
  background-color: #fff;
  padding: 20px 0;
}

.related-videos h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 15px;
  color: #333;
}

.loading-related {
  padding: 20px 0;
}

.no-related {
  text-align: center;
  padding: 40px 0;
  color: #999;
  font-size: 14px;
}

.related-video-item {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.related-video-item:last-child {
  margin-bottom: 0;
}

.related-video-item .video-cover {
  position: relative;
  width: 160px;
  height: 90px;
  flex-shrink: 0;
  overflow: hidden;
  border-radius: 6px;
}

.related-video-item .video-cover-link {
  text-decoration: none;
  color: inherit;
  display: inline-block;
}

.related-video-item .video-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.related-video-item .video-cover:hover img {
  transform: scale(1.05);
  transition: transform 0.3s;
}

.related-video-item .video-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.related-video-item .video-info h4 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
  pointer-events: none;
}

.related-video-item .video-title-text {
  cursor: pointer;
  transition: color 0.3s;
  pointer-events: auto;
}

.related-video-item .video-title-text:hover {
  color: #00aeec;
}

.related-video-item .video-meta {
  font-size: 12px;
  color: #999;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.related-video-item .video-author {
  font-size: 12px;
  color: #999;
  cursor: pointer;
  transition: color 0.3s;
}

.related-video-item .video-author:hover {
  color: #00aeec;
}

.related-video-item .video-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.related-video-item .video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

/* 响应式设计 */






/* 字幕相关样式 */
.subtitle-control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  margin-top: 10px;
  border-top: 1px solid #e0e0e0;
}

.subtitle-controls {
  display: flex;
  align-items: center;
}

.no-subtitle-tip {
  color: #999;
  font-size: 14px;
}

/* 分P列表样式 */
.video-part-list {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  margin-top: 16px;
  overflow: hidden;
}

.part-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f9f9f9;
  border-bottom: 1px solid #e0e0e0;
}

.part-list-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.part-count {
  font-size: 12px;
  color: #999;
}

.part-items {
  max-height: 300px;
  overflow-y: auto;
}

.part-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-bottom: 1px solid #f0f0f0;
}

.part-item:last-child {
  border-bottom: none;
}

.part-item:hover {
  background-color: #f5f5f5;
}

.part-item.is-active {
  background-color: #e6f7ff;
  border-left: 3px solid #00a1d6;
}

.part-index {
  min-width: 40px;
  font-size: 13px;
  color: #999;
  font-weight: 500;
}

.part-item.is-active .part-index {
  color: #00a1d6;
}

.part-title {
  flex: 1;
  font-size: 14px;
  color: #333;
  margin: 0 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.part-item.is-active .part-title {
  color: #00a1d6;
  font-weight: 500;
}

.part-duration {
  font-size: 12px;
  color: #999;
  min-width: 50px;
  text-align: right;
}

/* 滚动条样式 */
.part-items::-webkit-scrollbar {
  width: 6px;
}

.part-items::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.part-items::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.part-items::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}
</style>
