<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getNotifications } from '../../api/message'
import { getToken } from '../../utils/session'
import noface from '../../assets/noface.gif'

const route = useRoute()
const router = useRouter()

const type = computed(() => (route.params.type as string) || 'reply')
const list = ref<any[]>([])
const loading = ref(true)

const titles = { reply: '回复与@', at: '@我的', like: '收到喜欢', system: '系统通知' }
const title = computed(() => titles[type.value] || '消息通知')

const load = async () => {
  loading.value = true
  try {
    const res = await getNotifications(type.value)
    if (res.code === '1') list.value = res.data || []
    else list.value = []
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const token = getToken()
  if (!token) {
    router.replace('/m/login')
    return
  }
  load()
})

const goBack = () => router.back()

const formatTime = (t: string) => {
  if (!t) return ''
  const d = new Date(t.replace(' ', 'T'))
  const now = new Date()
  if (isNaN(d.getTime())) return t
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const goManuscript = (item: any) => {
  const id = item.manuscriptId || item.targetManuscriptId
  if (id) router.push(`/m/video/${id}`)
}
</script>

<template>
  <div class="notify-page">
    <div class="top-nav">
      <div class="back-btn" @click="goBack">
        <span class="back-arrow">&lt;</span>
      </div>
      <div class="page-title">{{ title }}</div>
      <div class="right-spacer"></div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="list.length > 0" class="notify-list">
      <div
        v-for="item in list"
        :key="item.id"
        class="notify-item"
        :class="{ unread: !item.isRead }"
        @click="goManuscript(item)"
      >
        <div class="avatar-wrap">
          <img :src="item.userAvatar || noface" class="avatar" />
        </div>
        <div class="notify-content">
          <div class="notify-header">
            <span class="name">{{ item.username || '系统' }}</span>
            <span class="time">{{ formatTime(item.createTime || item.createdAt) }}</span>
          </div>
          <div class="action-text">{{ item.actionText || '' }}</div>
          <div v-if="item.content" class="content">{{ item.content }}</div>
          <div v-if="item.videoTitle" class="video-title">{{ item.videoTitle }}</div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <p>暂无{{ title }}</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
.notify-page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 60px;
}

.top-nav {
  display: flex;
  align-items: center;
  height: 52px;
  padding: 0 16px;
  border-bottom: 1px solid #f1f2f3;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;

  .back-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    .back-arrow {
      font-size: 20px;
      color: #18191c;
    }
  }
  .page-title {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 600;
    color: #18191c;
  }
  .right-spacer {
    width: 32px;
  }
}

.loading {
  text-align: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}

.notify-list {
  padding: 0 16px;
}

.notify-item {
  display: flex;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #f5f6f7;
  cursor: pointer;

  .avatar-wrap {
    flex-shrink: 0;
    .avatar {
      width: 44px;
      height: 44px;
      border-radius: 50%;
      object-fit: cover;
    }
  }
  .notify-content {
    flex: 1;
    min-width: 0;
  }
  .notify-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    .name {
      font-size: 15px;
      font-weight: 600;
      color: #18191c;
    }
    .time {
      font-size: 12px;
      color: #9499a0;
    }
  }
  .action-text {
    font-size: 12px;
    color: #9499a0;
    margin-top: 2px;
  }
  .content {
    font-size: 14px;
    color: #18191c;
    margin-top: 4px;
    line-height: 1.5;
  }
  .video-title {
    font-size: 13px;
    color: #fb7299;
    margin-top: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.empty-state {
  text-align: center;
  padding: 80px 0;
  color: #9499a0;
  font-size: 14px;
}
</style>
