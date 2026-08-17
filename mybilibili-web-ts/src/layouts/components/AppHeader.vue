<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { House, VideoCamera } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { userApi } from '@/api/client'
import { messageApi } from '../../api/message.ts'
import {
  clearAuthSession,
  getCurrentUserId,
  getStoredUser,
  getToken,
  hasAuthSession,
  hasValidAccessToken,
  setAuthSession
} from '../../utils/auth.ts'
import { useNotificationWs } from '../../composables/useNotificationWs.ts'
import SearchBar from './SearchBar.vue'
import UserMenu from './UserMenu.vue'
import NotificationBell from './NotificationBell.vue'
import HeaderActions from './HeaderActions.vue'

const props = defineProps({
  mode: {
    type: String,
    default: 'transparent',
    validator: (value) => ['transparent', 'white'].includes(value)
  }
})

const emit = defineEmits(['showLogin', 'logout'])

const router = useRouter()
const route = useRoute()

const { unreadCounts: wsUnreadCounts, connect: wsConnect, disconnect: wsDisconnect } = useNotificationWs()

const isLogged = ref(false)
const userInfo = ref(null)

const isScrolled = ref(false)

const isZoomed = ref(false)
const ZOOM_THRESHOLD = 1.1

const unreadCounts = ref({
  private: 0,
  reply: 0,
  at: 0,
  like: 0,
  system: 0,
  dynamic: 0
})

const totalUnreadCount = computed(() => {
  return unreadCounts.value.private + unreadCounts.value.reply + 
         unreadCounts.value.at + unreadCounts.value.like + unreadCounts.value.system
})

const fetchUnreadCounts = async () => {
  try {
    const res = await messageApi.getUnreadCounts()
    if (res.code === 200) {
      unreadCounts.value = { ...unreadCounts.value, ...res.data }
    }
  } catch (error) {
    if (error.code !== 'ERR_ABORTED') {
      console.error('获取未读消息数失败:', error)
    }
  }
}

watch(wsUnreadCounts, (counts) => {
  if (counts) {
    unreadCounts.value = { ...unreadCounts.value, ...counts }
  }
}, { deep: true })

const shouldShowScrolled = computed(() => {
  if (props.mode === 'white') return true
  return isScrolled.value
})

const isSearchPage = computed(() => {
  return route.path === '/search'
})

const fetchUserInfo = async () => {
  const userId = getCurrentUserId()
  if (!userId) return false
  
  try {
    const response = await userApi.getUserById(userId)
    if (response.code === 200) {
      userInfo.value = response.data
      setAuthSession({ user: response.data })
      return true
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
  return false
}

const checkTokenExpiration = () => {
  if (!hasAuthSession()) {
    isLogged.value = false
    userInfo.value = null
    return false
  }

  if (!hasValidAccessToken() && !getStoredUser()) {
    isLogged.value = false
    userInfo.value = null
    return false
  }

  const userData = getStoredUser()
  if (userData) {
    userInfo.value = userData
    isLogged.value = true
  }
  fetchUserInfo()
  return true
}

const handleLogout = () => {
  clearAuthSession()
  isLogged.value = false
  userInfo.value = null
  wsDisconnect()
  ElMessage.success('退出登录成功')
  emit('logout')
}

const handleScroll = () => {
  isScrolled.value = window.scrollY > 10
}

let currentZoomRatio = 1
let initialDPR = 0
let zoomDebounceTimer = null
let isZoomInitialized = false

const checkZoom = () => {
  const currentDPR = window.devicePixelRatio || 1

  if (!isZoomInitialized) {
    initialDPR = currentDPR
    currentZoomRatio = 1
    isZoomed.value = false
    isZoomInitialized = true
    return
  }

  const newZoomRatio = currentDPR / initialDPR

  if (Math.abs(newZoomRatio - currentZoomRatio) > 0.001) {
    currentZoomRatio = newZoomRatio
    const shouldBeZoomed = currentZoomRatio >= ZOOM_THRESHOLD
    
    if (isZoomed.value !== shouldBeZoomed) {
      isZoomed.value = shouldBeZoomed
    }
  }
}

const debouncedCheckZoom = () => {
  if (zoomDebounceTimer) {
    clearTimeout(zoomDebounceTimer)
  }
  zoomDebounceTimer = setTimeout(() => {
    checkZoom()
    zoomDebounceTimer = null
  }, 100)
}

const resetZoomBase = () => {
  initialDPR = 0
  checkZoom()
}

const handleVisualViewportResize = () => {
}

const handleKeyZoom = (e) => {
  if (e.ctrlKey && (e.key === '+' || e.key === '-' || e.key === '0')) {
    debouncedCheckZoom()
  }
}

const handleWheelZoom = (e) => {
  if (e.ctrlKey) {
    debouncedCheckZoom()
  }
}

const forceCheckZoom = () => {
  resetZoomBase()
  setTimeout(checkZoom, 100)
}

onMounted(() => {
  checkTokenExpiration()

  if (props.mode === 'transparent') {
    window.addEventListener('scroll', handleScroll)
  }

  checkZoom()

  if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', handleVisualViewportResize)
  }

  window.addEventListener('resize', checkZoom)
  window.addEventListener('keydown', handleKeyZoom)
  window.addEventListener('wheel', handleWheelZoom, { passive: false })
  
  fetchUnreadCounts()
  const unreadInterval = setInterval(fetchUnreadCounts, 60000)

  if (getToken()) {
    wsConnect()
  }
  
  onUnmounted(() => {
    clearInterval(unreadInterval)
    wsDisconnect()
  })
})

onUnmounted(() => {
  if (props.mode === 'transparent') {
    window.removeEventListener('scroll', handleScroll)
  }
  window.removeEventListener('resize', checkZoom)
  window.removeEventListener('keydown', handleKeyZoom)
  window.removeEventListener('wheel', handleWheelZoom)
  if (window.visualViewport) {
    window.visualViewport.removeEventListener('resize', handleVisualViewportResize)
  }
  if (zoomDebounceTimer) {
    clearTimeout(zoomDebounceTimer)
    zoomDebounceTimer = null
  }
})
</script>

<template>
  <header :class="['app-header', { 'scrolled': shouldShowScrolled, 'white-mode': mode === 'white' }]">
    <div class="header-container">
      <div class="header-left">
        <el-button link @click="router.push('/')" class="home-icon">
          <el-icon><House /></el-icon>
          <span>首页</span>
        </el-button>
        <el-button link @click="router.push('/live')" class="home-icon">
          <el-icon><VideoCamera /></el-icon>
          <span>直播</span>
        </el-button>
      </div>
      
      <div class="header-center" v-if="!isSearchPage">
        <SearchBar />
      </div>
      <div class="header-center" v-else></div>
      
      <div class="header-right">
        <UserMenu
          :user-info="userInfo"
          :is-logged="isLogged"
          @show-login="emit('showLogin')"
          @logout="handleLogout"
        />
        <NotificationBell
          :total-unread-count="totalUnreadCount"
          :is-zoomed="isZoomed"
        />
        <HeaderActions
          :is-zoomed="isZoomed"
          :dynamic-unread-count="unreadCounts.dynamic"
        />
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 80px;
  background: transparent;
  display: flex;
  align-items: center;
  z-index: 100;
  transition: all 0.3s ease;
}

.app-header.scrolled,
.app-header.white-mode {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-container {
  max-width: 2560px;
  margin: 0 auto;
  padding: 10px 0;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 16px;
  height: 60px;
}

.home-icon {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 8px 12px;
  color: #fff;
  transition: all 0.3s;
  min-width: 60px;
  height: 60px;
}

.home-icon:hover {
  background: rgba(255, 255, 255, 0.1);
}

.app-header.scrolled .home-icon span,
.app-header.white-mode .home-icon span {
  color: #333;
}

.app-header.scrolled .home-icon,
.app-header.white-mode .home-icon {
  color: #333;
}

.home-icon .el-icon {
  font-size: 20px;
}

.home-icon span {
  font-size: 12px;
}

.header-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  padding: 0 20px;
  height: 60px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 2px;
  justify-content: flex-end;
}

.app-header.scrolled :deep(.action-btn),
.app-header.white-mode :deep(.action-btn) {
  color: #333;
}

@media (max-width: 1200px) {
  .header-right {
    gap: 2px;
  }
  
  :deep(.action-btn) {
    padding: 6px 12px;
    font-size: 13px;
  }
}

@media (max-width: 768px) {
  .header-container {
    padding: 0 16px;
  }
  
  .header-center {
    padding: 0 16px;
  }
  
  :deep(.search-box) {
    max-width: 300px;
  }
  
  :deep(.action-btn) {
    padding: 6px 4px;
    min-width: 40px;
  }
  
  :deep(.upload-btn) {
    padding: 6px 12px;
  }
}

@media (max-width: 480px) {
  .header-right {
    gap: 2px;
  }
  
  :deep(.search-box) {
    max-width: 200px;
  }
}
</style>