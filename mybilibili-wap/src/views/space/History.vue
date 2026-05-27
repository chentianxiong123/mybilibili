<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { historyApi } from '../../api/index'

const router = useRouter()
const histories = ref<Array<{ date: string; items: any[] }>>([])
const loading = ref(true)
const page = ref(1)

// 顶部分类 Tabs
const activeTab = ref('all')
const tabs = [
  { id: 'all', label: '全部' },
  { id: 'video', label: '视频' },
  { id: 'live', label: '直播' },
  { id: 'column', label: '专栏' },
  { id: 'goods', label: '商品' },
  { id: 'show', label: '展演' }
]

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.replace('/m/login')
    return
  }
  await loadHistory()
})

const loadHistory = async () => {
  loading.value = true
  try {
    const res = await historyApi.getHistoryList(page.value, 20)
    const list = res?.data?.records || res?.data || []

    if (list.length > 0) {
      const groupMap: Record<string, any[]> = {}
      const now = new Date()

      list.forEach((item: any) => {
        const viewTime = item.watchedAt || item.viewTime || item.createTime || new Date().toISOString()
        const d = new Date(viewTime)

        // 计算天数差
        const nowZero = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
        const dZero = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
        const diffDays = Math.round((nowZero - dZero) / 86400000)

        let label: string
        if (diffDays === 0) label = '今天'
        else if (diffDays === 1) label = '昨天'
        else if (diffDays === 2) label = '前天'
        else label = '更早'

        if (!groupMap[label]) groupMap[label] = []

        const video = item.video || {}
        const uploader = video.uploader || {}
        const progress = item.progressSeconds || 0
        const totalDuration = item.videoDuration || 0

        groupMap[label].push({
          aId: video.manuscriptId || item.manuscriptId || item.id,
          title: video.title || item.manuscriptTitle || item.title || '',
          pic: video.coverUrl || item.coverUrl || item.cover || '',
          author: uploader.name || uploader.nickname || item.authorName || item.author || '',
          timeStr: viewTime ? formatDate(viewTime) : '',
          progressStr: formatProgress(progress, totalDuration),
          durationStr: formatDuration(totalDuration),
          progressPercent: totalDuration > 0 ? Math.min(100, Math.floor((progress / totalDuration) * 100)) : 0
        })
      })

      histories.value = Object.entries(groupMap).map(([date, items]) => ({ date, items }))
    }
  } catch (e) {
    console.error('加载历史记录失败:', e)
  } finally {
    loading.value = false
  }
}

// 格式化观看进度
const formatProgress = (prog: number, dur: number) => {
  if (prog >= dur) return '已看完'
  return `${formatTimeDigit(Math.floor(prog / 60))}:${formatTimeDigit(prog % 60)}`
}

// 格式化总时长
const formatDuration = (dur: number) => {
  return `${formatTimeDigit(Math.floor(dur / 60))}:${formatTimeDigit(dur % 60)}`
}

const formatTimeDigit = (num: number) => {
  return num < 10 ? '0' + num : num
}

// 格式化完整时间（截图格式：2026年5月27日 14:11）
const formatDate = (dateStr: string) => {
  const d = new Date(dateStr)
  const y = d.getFullYear()
  const m = d.getMonth() + 1
  const day = d.getDate()
  const h = formatTimeDigit(d.getHours())
  const min = formatTimeDigit(d.getMinutes())
  return `${y}年${m}月${day}日 ${h}:${min}`
}

const clearAll = async () => {
  try {
    await historyApi.clearHistory()
    histories.value = []
  } catch (e) {
    console.error(e)
    histories.value = []
  }
}

const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="history-page">
    <!-- 顶部状态栏 -->
    <div class="top-nav">
      <div class="back-btn" @click="goBack">
        <span class="back-arrow">&lt;</span>
      </div>
      <div class="page-title">历史记录</div>
      <div class="right-icons">
        <div class="icon-btn" title="搜索">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </div>
        <div class="icon-btn" title="更多">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="5" r="1" />
            <circle cx="12" cy="12" r="1" />
            <circle cx="12" cy="19" r="1" />
          </svg>
        </div>
      </div>
    </div>

    <!-- 分类 Tabs (全部、视频、直播、专栏...) -->
    <div class="tabs-scroll-container">
      <div class="tabs-list">
        <div
          v-for="t in tabs"
          :key="t.id"
          :class="['tab-tab', { active: activeTab === t.id }]"
          @click="activeTab = t.id"
        >
          {{ t.label }}
        </div>
      </div>
      <div class="filter-icon" title="筛选列表">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="#61666d" stroke-width="2">
          <line x1="4" y1="21" x2="4" y2="14" />
          <line x1="4" y1="10" x2="4" y2="3" />
          <line x1="12" y1="21" x2="12" y2="12" />
          <line x1="12" y1="8" x2="12" y2="3" />
          <line x1="20" y1="21" x2="20" y2="16" />
          <line x1="20" y1="12" x2="20" y2="3" />
          <line x1="1" y1="14" x2="7" y2="14" />
          <line x1="9" y1="8" x2="15" y2="8" />
          <line x1="17" y1="16" x2="23" y2="16" />
        </svg>
      </div>
    </div>

    <!-- 历史数据呈现区域 -->
    <div class="history-content">
      <div v-if="loading" class="loading-wrap">加载中...</div>

      <div v-else-if="histories.length > 0" class="history-list">
        <div v-for="group in histories" :key="group.date" class="history-group">
          <!-- 日期标题（今天、昨天、更早） -->
          <div class="group-day-title">{{ group.date }}</div>

          <div class="history-items">
            <!-- 历史记录项 -->
            <div v-for="item in group.items" :key="item.aId" class="history-card-item">
              <router-link :to="`/m/video/${item.aId}`" class="card-link">
                <!-- 左侧：封面与进度叠加 -->
                <div class="cover-wrapper">
                  <img :src="item.pic" class="cover-img" />
                  <!-- 进度与时长条 -->
                  <div class="duration-badge">
                    {{ item.progressStr }} / {{ item.durationStr }}
                  </div>
                  <!-- 观看进度底条 -->
                  <div class="progress-bar-container">
                    <div class="progress-fill" :style="{ width: item.progressPercent + '%' }"></div>
                  </div>
                </div>

                <!-- 右侧：视频详情 -->
                <div class="video-info-wrap">
                  <h3 class="video-title">{{ item.title }}</h3>

                  <!-- UP主昵称 -->
                  <div class="author-row">
                    <span class="up-tag">UP</span>
                    <span class="author-name">{{ item.author }}</span>
                  </div>

                  <!-- 观看历史时间与更多按钮 -->
                  <div class="time-meta">
                    <div class="time-left">
                      <span class="device-icon">📱</span>
                      <span class="view-time">{{ item.timeStr }}</span>
                    </div>
                    <div class="more-btn">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="#9499a0" stroke-width="2">
                        <circle cx="12" cy="5" r="1" />
                        <circle cx="12" cy="12" r="1" />
                        <circle cx="12" cy="19" r="1" />
                      </svg>
                    </div>
                  </div>
                </div>
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- 空记录状态 -->
      <div v-else class="empty-state">
        <p>你还没有历史记录哦～</p>
      </div>

      <!-- 清空按钮 -->
      <div v-if="histories.length > 0" class="clear-btn-wrap">
        <button class="clear-btn" @click="clearAll">清空历史</button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.history-page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 60px;
  box-sizing: border-box;
}

/* 顶部状态栏 */
.top-nav {
  display: flex;
  align-items: center;
  height: 52px;
  padding: 0 16px;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 100;

  .back-btn {
    cursor: pointer;
    font-size: 20px;
    color: #61666d;
    width: 24px;
    display: flex;
    align-items: center;
  }

  .page-title {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 600;
    color: #18191c;
    margin-right: 12px;
  }

  .right-icons {
    display: flex;
    gap: 16px;
    color: #61666d;
    .icon-btn {
      cursor: pointer;
      display: flex;
      align-items: center;
    }
  }
}

/* 分类 Tabs */
.tabs-scroll-container {
  display: flex;
  align-items: center;
  border-bottom: 1px solid #f1f2f3;
  padding: 0 16px;
  height: 40px;
  background: #fff;
  position: sticky;
  top: 52px;
  z-index: 99;

  .tabs-list {
    flex: 1;
    display: flex;
    gap: 20px;
    overflow-x: auto;
    scrollbar-width: none; /* Firefox */
    &::-webkit-scrollbar {
      display: none; /* Chrome/Safari */
    }

    .tab-tab {
      font-size: 14px;
      color: #61666d;
      white-space: nowrap;
      height: 38px;
      line-height: 38px;
      position: relative;
      cursor: pointer;

      &.active {
        color: #fb7299;
        font-weight: bold;
        &::after {
          content: '';
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          height: 2px;
          background: #fb7299;
          border-radius: 1px;
        }
      }
    }
  }

  .filter-icon {
    padding-left: 12px;
    background: #fff;
    box-shadow: -6px 0 6px -4px rgba(0,0,0,0.1);
    display: flex;
    align-items: center;
    cursor: pointer;
  }
}

/* 历史内容区域 */
.history-content {
  padding: 0 16px;

  .loading-wrap {
    text-align: center;
    padding: 40px;
    color: #9499a0;
  }
}

.group-day-title {
  font-size: 14px;
  font-weight: bold;
  color: #18191c;
  margin: 16px 0 12px;
}

.history-card-item {
  margin-bottom: 16px;

  .card-link {
    display: flex;
    text-decoration: none;
    color: inherit;
    gap: 12px;
  }

  .cover-wrapper {
    position: relative;
    width: 140px;
    height: 87px;
    border-radius: 6px;
    overflow: hidden;
    background: #f1f2f3;
    flex-shrink: 0;

    .cover-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .duration-badge {
      position: absolute;
      bottom: 4px;
      right: 4px;
      background: rgba(0,0,0,0.6);
      color: #fff;
      font-size: 10px;
      padding: 1px 4px;
      border-radius: 2px;
    }

    .progress-bar-container {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: rgba(255,255,255,0.4);

      .progress-fill {
        height: 100%;
        background: #fb7299;
      }
    }
  }

  .video-info-wrap {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;

    .video-title {
      font-size: 14px;
      font-weight: 500;
      color: #18191c;
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      margin: 0;
    }

    .author-row {
      display: flex;
      align-items: center;
      gap: 4px;
      margin-top: 4px;

      .up-tag {
        font-size: 9px;
        color: #9499a0;
        border: 1px solid #e3e5e7;
        padding: 0 2px;
        border-radius: 2px;
        line-height: 1.1;
      }

      .author-name {
        font-size: 11px;
        color: #9499a0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }

    .time-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-top: 4px;

      .time-left {
        display: flex;
        align-items: center;
        gap: 4px;
        color: #9499a0;
        font-size: 11px;

        .device-icon {
          font-size: 10px;
          opacity: 0.8;
        }
      }

      .more-btn {
        cursor: pointer;
        display: flex;
        align-items: center;
      }
    }
  }
}

.empty-state {
  text-align: center;
  padding: 80px 0;
  color: #9499a0;
}

.clear-btn-wrap {
  text-align: center;
  padding: 16px 0 24px;
  .clear-btn {
    width: 140px;
    height: 34px;
    background: #fff;
    border: 1px solid #fb7299;
    color: #fb7299;
    font-size: 13px;
    border-radius: 17px;
    cursor: pointer;
    font-weight: bold;
    &:active {
      background: #fff0f3;
    }
  }
}
</style>
