<template>
  <div class="create-center-container">
    <!-- 顶部导航栏 -->
    <header class="create-center-header">
      <div class="header-left">
        <el-button class="center-title-btn" @click="goToCreateCenterHome">
          创作中心
        </el-button>
        <el-button class="main-site-btn" @click="goToMainSite">
          <el-icon><House /></el-icon>
          <span>主站</span>
        </el-button>
      </div>
      <div class="header-right">
        <el-avatar
          class="user-avatar"
          :size="40"
          :src="currentUser?.avatar || 'https://i0.hdslb.com/bfs/face/3378829f555891d2d5a4537e10264593a1d076b1.jpg@50w_50h_1c_1s_!web-avatar-nav.avif'"
          @click="goToUserProfile"
          style="cursor: pointer;"
        ></el-avatar>
        <div class="up-day-box">
          {{ creatorDaysText }}
        </div>
      </div>
    </header>

    <!-- 主体内容区域 -->
    <div class="create-center-main">
      <!-- 侧边导航栏 -->
      <CenterSidebar :active-key="currentActive" />

      <!-- 主内容区域 -->
      <main class="content-area">
        <!-- 首页内容 -->
        <div v-show="currentActive === 'home'" class="content-section">
          <div class="content-body">
            <CenterDashboard :active="currentActive === 'home'" />
          </div>
        </div>

        <!-- 稿件管理内容 -->
        <div v-show="currentActive === 'content-articles'" class="content-section">
          <div class="content-body">
            <ManuscriptManager :active="currentActive === 'content-articles'" />
          </div>
        </div>

        <!-- 数据中心内容 -->
        <div v-if="currentActive === 'data'" class="content-section">
          <DataCenterView />
        </div>

        <!-- 评论管理内容 -->
        <div v-show="currentActive === 'interaction-comment'" class="content-section">
          <div class="content-body">
            <CommentManager :active="currentActive === 'interaction-comment'" />
          </div>
        </div>

        <!-- 弹幕管理内容 -->
        <div v-show="currentActive === 'interaction-danmu'" class="content-section">
          <div class="content-body">
            <DanmuManager />
          </div>
        </div>

        <!-- 投稿内容 -->
        <div v-if="currentActive === 'upload'" class="content-section">
          <UploadView />
        </div>

        <!-- 粉丝管理 -->
        <div v-show="currentActive === 'fans'" class="content-section">
          <FansManager :active="currentActive === 'fans'" />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { safeStorage } from '@/utils/safeStorage'
import { useRouter, useRoute } from 'vue-router'
import { ref, computed, watch, onMounted } from 'vue'
import { manuscriptApi } from '@/api/creator'
import { useUserStore } from '@/stores/user'
import { House } from '@element-plus/icons-vue'
import UploadView from './UploadView.vue'
import DataCenterView from './DataCenterView.vue'
import CenterSidebar from '@/components/createCenter/CenterSidebar.vue'
import CenterDashboard from '@/components/createCenter/CenterDashboard.vue'
import ManuscriptManager from '@/components/createCenter/ManuscriptManager.vue'
import CommentManager from '@/components/createCenter/CommentManager.vue'
import DanmuManager from '@/components/createCenter/DanmuManager.vue'
import FansManager from '@/components/createCenter/FansManager.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// 当前登录用户信息 - 从localStorage获取以确保有值
const currentUser = computed(() => {
  const userStr = safeStorage.getItem('user')
  if (userStr) {
    try {
      return JSON.parse(userStr)
    } catch (e) {
      return userStore.userInfo
    }
  }
  return userStore.userInfo
})

// 创作者天数计算
const creatorDaysText = ref('还没成为创作者')

const calcCreatorDays = async () => {
  try {
    const response = await manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    if (response.code === 200) {
      const list = response.data.list || response.data.records || []
      if (list.length > 0) {
        const firstTime = list.reduce((earliest, item) => {
          const t = item.uploadTime || item.createdAt || item.createTime
          return t && (!earliest || new Date(t) < new Date(earliest)) ? t : earliest
        }, null)
        if (firstTime) {
          const startDate = new Date(firstTime)
          const now = new Date()
          const diffMs = now.getTime() - startDate.getTime()
          const days = Math.max(1, Math.floor(diffMs / (1000 * 60 * 60 * 24)))
          creatorDaysText.value = `成为创作者的第${days}天`
          return
        }
      }
    }
  } catch (error) {
    console.error('获取首个稿件失败:', error)
  }
  creatorDaysText.value = '还没成为创作者'
}

// 返回主站
const goToMainSite = () => {
  router.push('/')
}

// 跳转到个人主页
const goToUserProfile = () => {
  if (currentUser.value && currentUser.value.id) {
    router.push(`/profile/${currentUser.value.id}/home`)
  }
}

// 回到创作中心首页
const goToCreateCenterHome = () => {
  router.push('/create-center/home')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 当前激活的菜单索引
const currentActive = ref('home')

// 监听路由变化，同步当前激活的菜单
watch(
  () => route.path,
  (newPath) => {
    // 根据当前路径设置activeIndex
    const pathMap = {
      '/create-center/home': 'home',
      '/create-center/upload': 'upload',
      '/create-center/content': 'content',
      '/create-center/content-articles': 'content-articles',

      '/create-center/data': 'data',
      '/create-center/fans': 'fans',
      '/create-center/interaction': 'interaction',
      '/create-center/interaction-comment': 'interaction-comment',
      '/create-center/interaction-danmu': 'interaction-danmu',


    }

    if (pathMap[newPath]) {
      currentActive.value = pathMap[newPath]
    }
  },
  { immediate: true }
)

onMounted(() => {
  calcCreatorDays()
})
</script>

<style scoped>
.create-center-container {
  width: 100%;
  height: 100vh;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
}

/* 顶部导航栏样式 */
.create-center-header {
  height: 60px;
  background-color: #fff;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.center-title-btn {
  font-size: 20px;
  font-weight: 600;
  color: #1890ff;
  background-color: transparent;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  line-height: 40px;
}

.center-title-btn:hover {
  background-color: rgba(24, 144, 255, 0.1);
  color: #409eff;
}

.main-site-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  background-color: #fff;
  border: none;
  color: #606266;
}

.main-site-btn:hover {
  background-color: #f0f2f5;
  color: #1890ff;
  border: none;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-avatar {
  cursor: pointer;
}

.up-day-box {
  background-color: #f0f9ff;
  border: 1px solid #91d5ff;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 14px;
  color: #1890ff;
  display: flex;
  align-items: center;
}

/* 主体内容区域样式 */
.create-center-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 主内容区域样式 */
.content-area {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background-color: #f5f7fa;
}

.content-header {
  margin-bottom: 20px;
}

.content-header h2 {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.content-body {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

/* 内容区域通用样式 */
.content-section {
  height: 100%;
}
</style>