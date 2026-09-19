<template>
  <div class="dynamic-dropdown">
    <div class="dropdown-header">
      <span class="header-title">最新动态</span>
      <span class="view-all" @click="handleViewAll">查看全部</span>
    </div>
    <div class="dynamic-list" v-loading="loading">
      <div
        v-for="item in dynamicList"
        :key="item.id"
        class="dynamic-item"
        @click="handleItemClick(item)"
      >
        <el-avatar
          :size="40"
          :src="item.userAvatar"
          class="user-avatar"
        />
        <div class="dynamic-content">
          <div class="user-name">{{ item.userName }}</div>
          <div class="content-text">{{ item.content }}</div>
          <div class="item-time">{{ item.time }}</div>
        </div>
        <img loading="lazy" decoding="async"
          v-if="item.thumbnail"
          :src="item.thumbnail"
          class="item-thumbnail"
        />
      </div>
      <div v-if="dynamicList.length === 0 && !loading" class="empty-state">
        暂无动态
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import api from '@/api/client'
import { hasAuthSession } from '@/utils/auth'

export default {
  name: 'DynamicDropdown',
  data() {
    return {
      dynamicList: [] as any[],
      loading: false
    }
  },
  mounted() {
    if (!hasAuthSession()) return
    this.fetchDynamics()
  },
  methods: {
    formatTime(dateStr: string) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const minutes = Math.floor(diff / 60000)
      const hours = Math.floor(diff / 3600000)
      const days = Math.floor(diff / 86400000)

      if (minutes < 1) return '刚刚'
      if (minutes < 60) return `${minutes}分钟前`
      if (hours < 24) return `${hours}小时前`
      if (days < 30) return `${days}天前`
      return date.toLocaleDateString()
    },
    getThumbnail(item: any) {
      if (item.refVideo?.cover) return item.refVideo.cover
      if (item.imageUrls && item.imageUrls.length > 0) return item.imageUrls[0]
      return ''
    },
    async fetchDynamics() {
      this.loading = true
      try {
        const res = await api.get('/dynamic/all', { params: { page: 1, size: 10 } })
        const raw = res?.data ?? res
        const list = raw?.list || raw || []
        this.dynamicList = list.map((item: any) => ({
          id: item.id,
          userAvatar: item.user?.avatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=default',
          userName: item.user?.username || '用户',
          content: item.content || '',
          time: this.formatTime(item.createdAt),
          thumbnail: this.getThumbnail(item),
          userId: item.userId,
          refManuscriptId: item.refManuscriptId
        }))
      } catch (error) {
        console.error('获取动态失败:', error)
      } finally {
        this.loading = false
      }
    },
    handleItemClick(item: any) {
      if (item.refManuscriptId) {
        window.location.href = `/manuscript/${item.refManuscriptId}`
      } else {
        window.location.href = '/dynamic'
      }
    },
    handleViewAll() {
      window.location.href = '/dynamic'
    }
  }
}
</script>

<style scoped>
.dynamic-dropdown {
  width: 400px;
  max-height: 500px;
  overflow-y: auto;
  background: #fff;
  border-radius: 8px;
}

.dropdown-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e3e5e7;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #18191c;
}

.view-all {
  font-size: 12px;
  color: #00a1d6;
  cursor: pointer;
  transition: color 0.3s;
}

.view-all:hover {
  color: #0091c6;
}

.dynamic-list {
  padding: 8px 0;
  min-height: 100px;
}

.dynamic-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.dynamic-item:hover {
  background-color: #f5f7fa;
}

.user-avatar {
  flex-shrink: 0;
}

.dynamic-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}

.content-text {
  font-size: 13px;
  color: #61666d;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.item-time {
  font-size: 12px;
  color: #9499a0;
}

.item-thumbnail {
  width: 120px;
  height: 80px;
  object-fit: cover;
  border-radius: 6px;
  flex-shrink: 0;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #9499a0;
  font-size: 13px;
}
</style>
