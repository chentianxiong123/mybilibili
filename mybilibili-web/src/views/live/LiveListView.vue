<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoCamera, Star } from '@element-plus/icons-vue'
import { liveApi } from '../../api/live.js'

const router = useRouter()
const liveRooms = ref([])
const myRoom = ref(null)
const loading = ref(false)

const displayedRooms = computed(() => {
  if (!myRoom.value) return liveRooms.value
  return liveRooms.value.filter(room => room.userId !== myRoom.value.userId)
})

const loadData = async () => {
  loading.value = true
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
    }
  } catch (e) {
    console.error('获取直播数据失败:', e)
  } finally {
    loading.value = false
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

onMounted(loadData)
</script>

<template>
  <div class="live-list-page">
    <div class="page-header">
      <h2>正在直播</h2>
      <div class="header-actions">
        <el-button v-if="myRoom && myRoom.status === 'live'" type="warning" plain @click="enterMyRoom">
          <el-icon><VideoCamera /></el-icon>
          进入我的直播间
        </el-button>
        <el-button type="primary" @click="startLive">
          <el-icon><VideoCamera /></el-icon>
          开始直播
        </el-button>
      </div>
    </div>

    <div v-loading="loading" class="live-content">
      <!-- 我的直播间正在直播时显示 -->
      <div v-if="myRoom && myRoom.status === 'live'" class="my-room-section">
        <div class="section-title">
          <el-icon><VideoCamera /></el-icon>
          我的直播
        </div>
        <div class="live-card my-live-card" @click="enterMyRoom">
          <div class="card-cover">
            <img :src="myRoom.coverUrl || '/live-placeholder.svg'" alt="cover" />
            <span class="live-badge">LIVE</span>
            <span class="viewer-count">{{ myRoom.viewerCount || 0 }} 人观看</span>
          </div>
          <div class="card-info">
            <div class="room-name">{{ myRoom.roomName || '我的直播间' }}</div>
            <div class="live-status">
              <span class="status-dot" />
              直播中
            </div>
          </div>
        </div>
      </div>

      <!-- 全部直播 -->
      <div class="all-live-section">
        <div class="section-title" v-if="myRoom">
          <el-icon><Star /></el-icon>
          全部直播
        </div>
        <div v-if="displayedRooms.length === 0 && !loading" class="empty-state">
          <el-empty description="暂无直播，来当第一个主播吧！" />
        </div>
        <div class="live-grid">
          <div
            v-for="room in displayedRooms"
            :key="room.id"
            class="live-card"
            @click="goToRoom(room.id)"
          >
            <div class="card-cover">
              <img :src="room.coverUrl || '/live-placeholder.svg'" alt="cover" />
              <span v-if="room.status === 'live'" class="live-badge">LIVE</span>
              <span class="viewer-count">{{ room.viewerCount || 0 }} 人观看</span>
            </div>
            <div class="card-info">
              <div class="room-name">{{ room.roomName || '我的直播间' }}</div>
              <div class="user-id">主播ID: {{ room.userId }}</div>
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
  margin-bottom: 24px;
}
.page-header h2 { font-size: 22px; color: #18191c; margin: 0; }
.header-actions { display: flex; gap: 12px; }
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
  transition: box-shadow 0.2s;
}
.live-card:hover { box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
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
.live-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #f04040;
  font-weight: 500;
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