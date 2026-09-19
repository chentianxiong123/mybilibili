<template>
  <div class="history-dropdown">
    <div class="dropdown-header">
      <el-icon class="header-icon"><Clock /></el-icon>
      <span class="header-title">历史</span>
    </div>
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="historyGroups.length > 0" class="history-content">
      <div v-for="group in historyGroups" :key="group.date" class="history-group">
        <div class="group-date">{{ group.date }}</div>
        <div class="group-videos">
          <div v-for="video in group.videos" :key="video.id" class="video-item" @click="handleVideoClick(video)">
            <div class="thumbnail-wrapper">
              <img loading="lazy" decoding="async" :src="video.thumbnail || '/assets/placeholder-cover.svg'" class="video-thumbnail" />
              <span v-if="video.duration" class="video-duration">{{ video.duration }}</span>
            </div>
            <div class="video-info">
              <div class="video-title">{{ video.title }}</div>
              <div class="video-meta">
                <span class="video-time">{{ video.time }}</span>
                <span class="video-uploader">{{ video.uploader }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="empty-state">
      <p>暂无历史记录</p>
    </div>
  </div>
</template>

<script lang="ts">
import { watchHistoryApi } from '@/api/watchHistory'
import { hasAuthSession } from '@/utils/auth'

export default {
  name: 'HistoryDropdown',
  data() {
    return {
      historyGroups: [] as any[],
      loading: false
    }
  },
  mounted() {
    if (!hasAuthSession()) return
    this.loadHistory()
  },
  methods: {
    formatDate(date: Date) {
      const today = new Date()
      const yesterday = new Date(today)
      yesterday.setDate(yesterday.getDate() - 1)

      if (this.isSameDay(date, today)) {
        return '今天'
      } else if (this.isSameDay(date, yesterday)) {
        return '昨天'
      } else {
        const month = date.getMonth() + 1
        const day = date.getDate()
        return `${month}月${day}日`
      }
    },
    isSameDay(date1: Date, date2: Date) {
      return date1.getFullYear() === date2.getFullYear() &&
        date1.getMonth() === date2.getMonth() &&
        date1.getDate() === date2.getDate()
    },
    formatTime(date: Date) {
      const now = new Date()
      const isToday = this.isSameDay(date, now)
      const hours = date.getHours().toString().padStart(2, '0')
      const minutes = date.getMinutes().toString().padStart(2, '0')
      const timeStr = `${hours}:${minutes}`

      if (isToday) {
        return timeStr
      } else {
        const month = (date.getMonth() + 1).toString().padStart(2, '0')
        const day = date.getDate().toString().padStart(2, '0')
        return `${month}-${day} ${timeStr}`
      }
    },
    formatDuration(seconds: number) {
      if (!seconds || seconds === 0) return ''
      const hours = Math.floor(seconds / 3600)
      const minutes = Math.floor((seconds % 3600) / 60)
      const secs = seconds % 60

      if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
      }
      return `${minutes}:${secs.toString().padStart(2, '0')}`
    },
    async loadHistory() {
      this.loading = true
      try {
        const response = await watchHistoryApi.getWatchHistory(1, 10)
        const raw = response?.data ?? response
        const list = (Array.isArray(raw) ? raw : (Array.isArray(raw?.list) ? raw.list : []))
          .map((item: any) => {
            const video = item.video || {}
            const uploader = video.uploader || {}
            const watchedAt = (item.watchedAt || item.watched_at) ? new Date(item.watchedAt || item.watched_at) : new Date()

            return {
              id: item.id,
              videoId: item.videoId || item.video_id,
              manuscriptId: video.manuscriptId || video.manuscript_id || video.id,
              title: video.title || '未知视频',
              thumbnail: video.coverUrl || video.cover_url || video.cover || '',
              duration: this.formatDuration(item.videoDuration || item.video_duration),
              progressSeconds: item.progressSeconds || item.progress_seconds || 0,
              watchedAt: watchedAt,
              dateKey: this.formatDate(watchedAt),
              time: this.formatTime(watchedAt),
              uploader: uploader.name || uploader.nickname || video.author || video.userName || '未知UP主'
            }
          })

        // 按日期分组
        const groups: Record<string, any[]> = {}
        list.forEach((item: any) => {
          if (!groups[item.dateKey]) {
            groups[item.dateKey] = []
          }
          groups[item.dateKey].push(item)
        })

        this.historyGroups = Object.keys(groups).map(date => ({
          date,
          videos: groups[date]
        }))
      } catch (error) {
        console.error('加载历史记录失败:', error)
      } finally {
        this.loading = false
      }
    },
    handleVideoClick(video: any) {
      const targetId = video.manuscriptId || video.videoId
      if (targetId) {
        const t = Number(video.progressSeconds) || 0
        const query = t > 0 ? `?t=${Math.floor(t)}` : ''
        window.location.href = `/manuscript/${targetId}${query}`
      }
    }
  }
}
</script>

<style scoped>
.history-dropdown {
  width: 400px;
  max-height: 500px;
  overflow-y: auto;
  background: #fff;
  border-radius: 8px;
}

.dropdown-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #e3e5e7;
}

.header-icon {
  font-size: 18px;
  color: #9499a0;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #18191c;
}

.loading-state {
  padding: 20px;
}

.history-content {
  padding: 8px 0;
}

.history-group {
  margin-bottom: 8px;
}

.group-date {
  font-size: 12px;
  color: #9499a0;
  padding: 8px 16px;
}

.group-videos {
  display: flex;
  flex-direction: column;
}

.video-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.video-item:hover {
  background-color: #f5f7fa;
}

.thumbnail-wrapper {
  position: relative;
  flex-shrink: 0;
}

.video-thumbnail {
  width: 160px;
  height: 90px;
  object-fit: cover;
  border-radius: 6px;
}

.video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 2px;
}

.video-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.video-title {
  font-size: 14px;
  color: #18191c;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.video-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.video-time {
  font-size: 12px;
  color: #9499a0;
}

.video-uploader {
  font-size: 12px;
  color: #9499a0;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #9499a0;
  font-size: 14px;
}
</style>
