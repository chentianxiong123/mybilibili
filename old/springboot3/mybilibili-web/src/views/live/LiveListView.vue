<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoCamera, Star, Search, Refresh } from '@element-plus/icons-vue'
import { liveApi } from '../../api/live.js'

const router = useRouter()
const liveRooms = ref([])
const myRoom = ref(null)
const loading = ref(false)
const keyword = ref('')
const sortBy = ref('hot') // 'hot' | 'new'
const categories = ['全部', '娱乐', '游戏', '学习', '购物', '赛事', '其他']
const selectedCategory = ref('全部')
let pollTimer = null

const filteredRooms = computed(() => {
  let list = liveRooms.value
  // 排除自己（已有"我的直播间"区块）
  if (myRoom.value) {
    list = list.filter(r => r.userId !== myRoom.value.userId)
  }
  // 分类筛选
  if (selectedCategory.value && selectedCategory.value !== '全部') {
    list = list.filter(r => r.category === selectedCategory.value)
  }
  // 关键字搜索
  const kw = keyword.value.trim().toLowerCase()
  if (kw) {
    list = list.filter(r =>
      (r.roomName || '').toLowerCase().includes(kw) ||
      String(r.userId).includes(kw)
    )
  }
  // 排序
  list = [...list]
  if (sortBy.value === 'hot') {
    list.sort((a, b) => (b.viewerCount || 0) - (a.viewerCount || 0))
  } else {
    list.sort((a, b) => (b.id || 0) - (a.id || 0))
  }
  return list
})

const totalLive = computed(() => liveRooms.value.length)

const loadData = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    const [listRes, myRoomRes] = await Promise.allSettled([
      liveApi.getLiveList(),
      liveApi.getMyRoom()
    ])

    if (listRes.status === 'fulfilled' && listRes.value.code === 200) {
      liveRooms.value = listRes.value.data || []
    }

    if (myRoomRes.status === 'fulfilled' && myRoomRes.value.code === 200 && myRoomRes.value.data) {
      myRoom.value = myRoomRes.value.data
    } else {
      myRoom.value = null
    }
  } catch (e) {
    console.error('获取直播数据失败:', e)
  } finally {
    if (!silent) loading.value = false
  }
}

const goToRoom = (id) => {
  router.push(`/live/${id}`)
}

const startLive = () => {
  router.push('/live/push')
}

const enterMyRoom = () => {
  if (myRoom.value) {
    router.push(`/live/${myRoom.value.id}`)
  }
}

const handleRefresh = () => {
  loadData()
  ElMessage.success('已刷新')
}

onMounted(() => {
  loadData()
  // 每 10s 静默刷新一次列表
  pollTimer = setInterval(() => loadData(true), 10000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div class="live-list-page">
    <div class="page-header">
      <div class="title-row">
        <h2>正在直播</h2>
        <el-tag round effect="plain" type="danger">
          <span class="dot" />{{ totalLive }} 个直播间
        </el-tag>
      </div>
      <div class="header-actions">
        <el-button v-if="myRoom && myRoom.status === 'live'" type="warning" plain @click="enterMyRoom">
          <el-icon><VideoCamera /></el-icon>
          进入我的直播间
        </el-button>
        <el-button v-else-if="myRoom" plain @click="startLive">
          <el-icon><VideoCamera /></el-icon>
          开始直播
        </el-button>
        <el-button v-else type="primary" @click="startLive">
          <el-icon><VideoCamera /></el-icon>
          创建直播间
        </el-button>
      </div>
    </div>

    <!-- 搜索 + 排序 -->
    <div class="filter-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索直播间名称或主播ID"
        clearable
        class="search-box"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-radio-group v-model="sortBy" size="default">
        <el-radio-button value="hot">最热</el-radio-button>
        <el-radio-button value="new">最新</el-radio-button>
      </el-radio-group>
      <el-button :icon="Refresh" circle @click="handleRefresh" :loading="loading" title="刷新" />
    </div>

    <!-- 分类标签 -->
    <div class="category-chips">
      <span
        v-for="cat in categories"
        :key="cat"
        :class="['chip', { active: selectedCategory === cat }]"
        @click="selectedCategory = cat"
      >{{ cat }}</span>
    </div>

    <div v-loading="loading" class="live-content">
      <!-- 我的直播间 -->
      <div v-if="myRoom" class="my-room-section">
        <div class="section-title">
          <el-icon><VideoCamera /></el-icon>
          我的直播间
        </div>
        <div class="live-card my-live-card" @click="myRoom.status === 'live' ? enterMyRoom() : startLive()">
          <div class="card-cover">
            <img :src="myRoom.coverUrl || '/live-placeholder.svg'" alt="cover" />
            <span v-if="myRoom.status === 'live'" class="live-badge">LIVE</span>
            <span v-else class="offline-badge">未开播</span>
            <span v-if="myRoom.status === 'live'" class="viewer-count">
              {{ myRoom.viewerCount || 0 }} 人观看
            </span>
          </div>
          <div class="card-info">
            <div class="room-name">{{ myRoom.roomName || '我的直播间' }}</div>
            <div v-if="myRoom.status === 'live'" class="live-status">
              <span class="status-dot" />直播中（点击进入）
            </div>
            <div v-else class="offline-hint">
              点击进入控制台开播
            </div>
          </div>
        </div>
      </div>

      <!-- 全部直播 -->
      <div class="all-live-section">
        <div class="section-title">
          <el-icon><Star /></el-icon>
          全部直播
          <span v-if="keyword" class="filter-hint">"{{ keyword }}" 共 {{ filteredRooms.length }} 个</span>
        </div>
        <div v-if="filteredRooms.length === 0 && !loading" class="empty-state">
          <el-empty :description="keyword ? '没有匹配的直播间' : '暂无直播，来当第一个主播吧！'" />
        </div>
        <div v-else class="live-grid">
          <div
            v-for="room in filteredRooms"
            :key="room.id"
            class="live-card"
            @click="goToRoom(room.id)"
          >
            <div class="card-cover">
              <img :src="room.coverUrl || '/live-placeholder.svg'" alt="cover" />
              <span class="live-badge">LIVE</span>
              <span class="viewer-count">{{ room.viewerCount || 0 }} 人观看</span>
            </div>
            <div class="card-info">
              <div class="room-name">{{ room.roomName || '直播间' }}</div>
              <div class="user-id">主播ID: {{ room.userId }}</div>
              <div v-if="room.category" class="room-category">{{ room.category }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-list-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.page-header h2 { font-size: 22px; color: #18191c; margin: 0; }
.dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #f04040;
  border-radius: 50%;
  margin-right: 4px;
  animation: pulse 1.5s infinite;
}
.header-actions { display: flex; gap: 12px; }

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.search-box {
  max-width: 320px;
  flex: 1;
}
.category-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.chip {
  padding: 5px 16px;
  border-radius: 20px;
  font-size: 13px;
  background: #e3e5e7;
  color: #61666d;
  cursor: pointer;
  transition: all 0.2s;
}
.chip:hover { background: #e8e9eb; }
.chip.active {
  background: #00a1d6;
  color: #fff;
  font-weight: 500;
}

.live-content { display: flex; flex-direction: column; gap: 32px; }
.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 16px;
}
.section-title .el-icon { font-size: 18px; color: #fb7299; }
.filter-hint {
  font-size: 13px;
  color: #9499a0;
  font-weight: 400;
  margin-left: 8px;
}
.live-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}
.live-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s;
}
.live-card:hover {
  box-shadow: 0 4px 20px rgba(0,0,0,0.1);
  transform: translateY(-2px);
}
.my-live-card { max-width: 400px; }
.card-cover {
  position: relative;
  aspect-ratio: 16/9;
  background: #e3e5e7;
}
.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.live-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  background: #f04040;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}
.offline-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  background: #9499a0;
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}
.viewer-count {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0,0,0,0.6);
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}
.card-info { padding: 12px; }
.room-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.user-id { font-size: 12px; color: #9499a0; }
.room-category {
  display: inline-block;
  margin-top: 4px;
  padding: 1px 8px;
  background: #e3e5e7;
  border-radius: 10px;
  font-size: 11px;
  color: #61666d;
}
.live-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #f04040;
  font-weight: 500;
}
.offline-hint {
  font-size: 12px;
  color: #9499a0;
}
.status-dot {
  width: 6px;
  height: 6px;
  background: #f04040;
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.empty-state { padding: 40px 0; }
</style>
