<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Refresh, VideoCamera } from '@element-plus/icons-vue'
import { useLiveRoomList } from '@/composables/useLiveRoomList.ts'
import { isLiveRoomStatus } from '@/utils/liveMeetingStatus.ts'

const router = useRouter()
const {
  myRoom,
  loading,
  keyword,
  categories,
  selectedCategory,
  filteredRooms,
  totalLive,
  loadData,
  refreshData,
  startPolling,
  cleanupLiveRoomList
} = useLiveRoomList()

const goToRoom = (id: number | string) => {
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

// 顶部大播放位固定取第一个房间（对齐 B 站首页首位主播直播位）
const heroRoom = computed(() => filteredRooms.value[0] || null)
// 推荐直播网格 = 其余房间
const restRooms = computed(() => filteredRooms.value.slice(1))

const fmtViewers = (n?: number) => {
  const v = n || 0
  if (v >= 10000) return `${(v / 10000).toFixed(1)}万`
  return String(v)
}

onMounted(() => {
  loadData()
  startPolling()
})

onUnmounted(() => {
  cleanupLiveRoomList()
})
</script>

<template>
  <div class="live-home">
    <!-- 分类导航条（对齐 B 站分类胶囊：高28 圆角8 灰底#f6f7f8 字14） -->
    <nav class="cat-nav">
      <div
        v-for="cat in categories"
        :key="cat"
        :class="['cat-pill', { active: selectedCategory === cat }]"
        @click="selectedCategory = cat"
      >{{ cat }}</div>
      <div class="cat-tools">
        <el-input
          v-model="keyword"
          placeholder="搜索直播间"
          clearable
          class="cat-search"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Refresh" circle @click="refreshData" :loading="loading" title="刷新" />
      </div>
    </nav>

    <div v-loading="loading" class="live-main">
      <!-- 头部操作（开始直播 / 进入我的直播间） -->
      <div class="home-bar">
        <h2 class="live-count"><span class="live-dot" />{{ totalLive }} 个直播间正在直播</h2>
        <div class="home-actions">
          <el-button
            v-if="myRoom && isLiveRoomStatus(myRoom.status)"
            type="primary"
            @click="enterMyRoom"
          >
            <el-icon><VideoCamera /></el-icon>
            进入我的直播间
          </el-button>
          <el-button v-else type="primary" @click="startLive">
            <el-icon><VideoCamera /></el-icon>
            开始直播
          </el-button>
        </div>
      </div>

      <!-- 主播大播放位（16:9，对齐 B 站 992×558 播放位） -->
      <div v-if="heroRoom" class="hero-row">
        <div class="hero-player" @click="goToRoom(heroRoom.id)">
          <img class="hero-cover" :src="heroRoom.coverUrl || '/live-placeholder.svg'" alt="cover" />
          <span class="hero-live-badge">直播中</span>
          <span class="hero-viewers">{{ fmtViewers(heroRoom.viewerCount) }} 人观看</span>
          <div class="hero-enter">点击进入直播间</div>
        </div>
        <aside class="hero-side">
          <div class="side-section">
            <header class="side-head">
              <h3>主播信息</h3>
            </header>
            <div class="side-body">
              <div class="side-room-name">{{ heroRoom.roomName || '直播间' }}</div>
              <div class="side-meta">
                <span>主播ID：{{ heroRoom.userId }}</span>
                <span v-if="heroRoom.category" class="side-cat">{{ heroRoom.category }}</span>
              </div>
            </div>
          </div>
          <div v-if="myRoom" class="side-section">
            <header class="side-head">
              <h3>我的直播间</h3>
            </header>
            <div class="side-body">
              <div class="side-room-name">{{ myRoom.roomName || '我的直播间' }}</div>
              <div class="side-status" :class="{ off: !isLiveRoomStatus(myRoom.status) }">
                {{ isLiveRoomStatus(myRoom.status) ? '直播中' : '未开播' }}
              </div>
              <el-button
                size="small"
                type="primary"
                plain
                class="side-btn"
                @click="isLiveRoomStatus(myRoom.status) ? enterMyRoom() : startLive()"
              >{{ isLiveRoomStatus(myRoom.status) ? '进入直播间' : '去开播' }}</el-button>
            </div>
          </div>
        </aside>
      </div>

      <!-- 推荐直播网格（对齐 B 站 5 列 / 363 宽 / 封面 16:9 圆角6 / 卡片圆角4） -->
      <div class="rec-section">
        <h2 class="rec-title">推荐直播</h2>
        <div v-if="restRooms.length === 0 && !loading" class="empty-state">
          <el-empty :description="keyword ? '没有匹配的直播间' : '暂无更多直播，来当第一个主播吧！'" />
        </div>
        <div v-else class="rec-grid">
          <div
            v-for="room in restRooms"
            :key="room.id"
            class="rec-card-wrap"
          >
            <a class="rec-card" @click.prevent="goToRoom(room.id)">
              <div class="rec-cover">
                <img class="rec-cover-img" :src="room.coverUrl || '/live-placeholder.svg'" alt="cover" />
                <span class="rec-viewers">{{ fmtViewers(room.viewerCount) }} 人观看</span>
              </div>
              <div class="rec-info">
                <span class="rec-name">{{ room.roomName || '直播间' }}</span>
                <span class="rec-up">{{ room.category || `主播ID：${room.userId}` }}</span>
              </div>
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 页面容器：对齐 B 站 live main —— 通栏 + 左右 20px margin；背景融入站点灰底（B 站 body/app 透明） */
.live-home {
  width: auto;
  margin: 0 20px;
  padding-top: 16px;
  background: transparent;
}

/* 分类导航条：高 28 胶囊 圆角8 灰底 #f6f7f8 字 14/20 */
.cat-nav {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}
.cat-pill {
  height: 28px;
  line-height: 20px;
  padding: 0 14px;
  font-size: 14px;
  color: #18191c;
  background: #f6f7f8;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background-color 0.2s, color 0.2s;
}
.cat-pill:hover {
  background: #e3e5e7;
}
.cat-pill.active {
  background: #00a1d6;
  color: #fff;
  font-weight: 500;
}
.cat-tools {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 10px;
}
.cat-search {
  width: 200px;
}

/* 头部条 */
.home-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.live-count {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
}
.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #f04040;
  animation: pulse 1.5s infinite;
}
.home-actions { display: flex; gap: 12px; }

/* 主播大播放位：左 992 + 右侧信息，16:9 */
.hero-row {
  display: flex;
  align-items: stretch;
  margin-bottom: 22px;
}
.hero-player {
  position: relative;
  flex: none;
  width: 992px;
  aspect-ratio: 16 / 9;
  background: #232527;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
}
.hero-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.hero-live-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 2px 8px;
  font-size: 12px;
  line-height: 18px;
  color: #fff;
  background: #f04040;
  border-radius: 4px;
}
.hero-viewers {
  position: absolute;
  bottom: 10px;
  right: 10px;
  padding: 2px 8px;
  font-size: 12px;
  line-height: 18px;
  color: #fff;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 4px;
}
.hero-enter {
  position: absolute;
  left: 50%;
  bottom: 24px;
  transform: translateX(-50%);
  padding: 8px 20px;
  font-size: 14px;
  color: #fff;
  background: rgba(0, 0, 0, 0.55);
  border-radius: 20px;
  opacity: 0;
  transition: opacity 0.2s;
}
.hero-player:hover .hero-enter { opacity: 1; }

/* 右侧信息栏：宽 865（对齐 B 站 idol 列），左外边距 13 */
.hero-side {
  flex: 1;
  min-width: 0;
  margin-left: 13px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.side-section {
  background: #f6f7f8;
  border-radius: 8px;
  padding: 10px;
}
.side-head {
  height: 20px;
  display: flex;
  align-items: center;
}
.side-head h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}
.side-body {
  margin-top: 10px;
}
.side-room-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.side-meta {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  line-height: 20px;
  color: #9499a0;
}
.side-cat {
  padding: 0 8px;
  background: #e3e5e7;
  border-radius: 8px;
  color: #61666d;
}
.side-status {
  margin-top: 6px;
  font-size: 12px;
  line-height: 20px;
  color: #f04040;
  font-weight: 500;
}
.side-status.off { color: #9499a0; }
.side-btn { margin-top: 10px; }

/* 推荐直播 */
.rec-section { margin-top: 22px; }
.rec-title {
  margin: 0 0 4px;
  font-size: 16px;
  font-weight: 500;
  line-height: 22px;
  color: #18191c;
}
/* 卡片宽 = (100% - 4*14px)/5，行/列间距 14/15，对齐 B 站 363 宽（靠末位 mr-14 溢出，无负 margin） */
.rec-grid {
  display: flex;
  flex-wrap: wrap;
}
.rec-card-wrap {
  width: calc((100% - 56px) / 5);
  margin: 15px 14px 0 0;
}
.rec-card {
  display: block;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  transition: background-color 0.2s, box-shadow 0.2s;
}
.rec-card:hover {
  box-shadow: 0 0 0 8px #f1f2f3;
}
/* 封面 16:9，圆角 6，灰底 #f1f2f3 */
.rec-cover {
  position: relative;
  width: 100%;
  height: 0;
  padding-top: 56.25%;
  background: #f1f2f3;
  border-radius: 6px;
  overflow: hidden;
}
.rec-cover-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.rec-viewers {
  position: absolute;
  bottom: 6px;
  right: 6px;
  padding: 1px 6px;
  font-size: 12px;
  line-height: 18px;
  color: #fff;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 4px;
}
/* 信息区：margin-top 5，标题 14/500/24 #18191c，副标题 12/20 #9499a0 */
.rec-info {
  margin-top: 5px;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.rec-name {
  height: 24px;
  line-height: 24px;
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rec-up {
  height: 20px;
  line-height: 20px;
  font-size: 12px;
  color: #9499a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-state { padding: 40px 0; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>
