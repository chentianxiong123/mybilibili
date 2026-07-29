<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getConversations, getUnreadCounts } from '../../api/message'
import noface from '../../assets/noface.gif'

const router = useRouter()
const conversations = ref<any[]>([])
const unreadCounts = ref<any>({
  reply: 0,
  at: 0,
  like: 0,
  system: 0
})
const loading = ref(true)

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.replace('/m/login')
    return
  }

  try {
    const [convRes, unreadRes] = await Promise.all([
      getConversations(),
      getUnreadCounts()
    ])

    if (convRes.code === '1') {
      conversations.value = convRes.data || []
    }
    if (unreadRes.code === '1') {
      unreadCounts.value = unreadRes.data || { reply: 0, at: 0, like: 0, system: 0 }
    }
  } catch (e) {
    console.error('加载消息失败:', e)
  } finally {
    loading.value = false
  }
})

// 时间格式化
const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()

  if (date.toDateString() === now.toDateString()) {
    // 今天的消息显示具体时间
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天'
  }

  // 否则显示月份和天
  return `${date.getMonth() + 1}月${date.getDate()}日`
}

const goBack = () => {
  router.back()
}

const openConversation = (item: any) => {
  const id = item.targetUserId || item.userId || item.id
  router.push({
    path: `/m/message/chat/${id}`,
    query: {
      name: item.targetUserNickname || item.targetUsername || item.nickname || '私信',
      avatar: item.targetUserAvatar || item.avatar || '',
      time: item.lastMessageTime || '',
      last: item.lastMessageContent || ''
    }
  })
}
</script>

<template>
  <div class="message-page">
    <!-- 顶部状态栏 -->
    <div class="top-nav">
      <div class="back-btn" @click="goBack">
        <span class="back-arrow">&lt;</span>
      </div>
      <div class="page-title">消息</div>
      <div class="right-icons">
        <div class="icon-btn" title="消息设置">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z" />
            <path d="M12 6v6l4 2" />
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

    <!-- 顶部三个分类功能入口（回复与@、收到喜欢、新增粉丝） -->
    <div class="category-tabs">
      <div class="tab-item">
        <div class="icon-wrap green">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="#47ccaa">
            <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
          </svg>
        </div>
        <span class="tab-label">回复与@</span>
        <div v-if="unreadCounts.reply + unreadCounts.at > 0" class="badge">
          {{ unreadCounts.reply + unreadCounts.at }}
        </div>
      </div>
      <div class="tab-item">
        <div class="icon-wrap pink">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="#ff5c7c">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
          </svg>
        </div>
        <span class="tab-label">收到喜欢</span>
        <div v-if="unreadCounts.like > 0" class="badge">
          {{ unreadCounts.like }}
        </div>
      </div>
      <div class="tab-item">
        <div class="icon-wrap blue">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="#23ade5">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
        </div>
        <span class="tab-label">新增粉丝</span>
        <div v-if="unreadCounts.system > 0" class="badge">
          {{ unreadCounts.system }}
        </div>
      </div>
    </div>

    <!-- 消息列表 -->
    <div class="chat-list-section">
      <div v-if="loading" class="loading">加载中...</div>

      <div v-else-if="conversations.length > 0" class="chat-list">
        <!-- 单个聊天项 -->
        <div
          v-for="item in conversations"
          :key="item.id"
          class="chat-item"
          @click="openConversation(item)"
        >
          <div class="avatar-wrap">
            <img :src="item.targetUserAvatar || noface" class="avatar" />
            <span v-if="item.unreadCount > 0" class="dot"></span>
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">{{ item.targetUserNickname || item.targetUsername }}</span>
              <span class="time">{{ formatTime(item.lastMessageTime) }}</span>
            </div>
            <div class="last-msg">
              {{ item.lastMessageContent || '暂无新消息' }}
            </div>
          </div>
        </div>
      </div>

      <!-- 骨架屏 / 兜底初始聊天记录（一比一复刻 B站 真实聊天列表风格） -->
      <div v-else class="chat-list">
        <!-- UP主小助手 -->
        <div class="chat-item">
          <div class="avatar-wrap system-avatar blue-light">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="#fff">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
            </svg>
            <span class="dot"></span>
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name system-title">UP主小助手</span>
              <span class="time">2天前</span>
            </div>
            <div class="last-msg">UP主荣誉周报</div>
          </div>
        </div>

        <!-- PCD沈和璧 -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1025/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">PCD沈和璧</span>
              <span class="time">5月15日</span>
            </div>
            <div class="last-msg">[tv_冷漠][热词系列_谢谢老师]还在偷摸的看.....</div>
          </div>
        </div>

        <!-- 陌生人消息 -->
        <div class="chat-item">
          <div class="avatar-wrap system-avatar pink-light">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="#fff">
              <rect x="3" y="4" width="18" height="16" rx="2" />
              <path d="M2 10h20" />
            </svg>
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">陌生人消息</span>
              <span class="time">5月12日</span>
            </div>
            <div class="last-msg">暂无新的陌生人消息</div>
          </div>
        </div>

        <!-- 大模型论文解读 -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1012/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">大模型论文解读</span>
              <span class="time">5月8日</span>
            </div>
            <div class="last-msg">谢谢你的喜欢～</div>
          </div>
        </div>

        <!-- 玩转Code -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1015/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">玩转Code</span>
              <span class="time">5月5日</span>
            </div>
            <div class="last-msg">[自动回复] 感谢关注 程序员陈师兄 qq30414...</div>
          </div>
        </div>

        <!-- AI_小波 -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1018/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">AI_小波</span>
              <span class="time">5月4日</span>
            </div>
            <div class="last-msg">[自动回复] 谢谢支持和关注❤️ 教程+入口：...</div>
          </div>
        </div>

        <!-- 山河江川美 -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1019/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">山河江川美</span>
              <span class="time">5月3日</span>
            </div>
            <div class="last-msg">[自动回复] thu.chatopens.vip/ 这是我自己...</div>
          </div>
        </div>

        <!-- QLHazycoder -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1020/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">QLHazycoder</span>
              <span class="time">5月3日</span>
            </div>
            <div class="last-msg">https://apikey.qzz.io/, 这里看哦</div>
          </div>
        </div>

        <!-- 大海资源 -->
        <div class="chat-item">
          <div class="avatar-wrap">
            <img src="https://picsum.photos/id/1021/48/48" class="avatar" />
          </div>
          <div class="chat-content">
            <div class="chat-header">
              <span class="name">大海资源</span>
              <span class="time">4月16日</span>
            </div>
            <div class="last-msg">[自动回复] 我是大海，谢谢你的支持和关注～ ...</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.message-page {
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
  border-bottom: 1px solid #f1f2f3;
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
    margin-right: 12px; /* 补偿返回键宽度，使文字居中 */
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

/* 三个分类功能按钮 */
.category-tabs {
  display: flex;
  justify-content: space-around;
  padding: 16px 0;
  border-bottom: 1px solid #f1f2f3;

  .tab-item {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    cursor: pointer;
    width: 80px;

    .icon-wrap {
      width: 48px;
      height: 48px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 8px;

      &.green {
        background: #f0fbf8;
      }
      &.pink {
        background: #fff3f5;
      }
      &.blue {
        background: #f0f9ff;
      }
    }

    .tab-label {
      font-size: 12px;
      color: #18191c;
    }

    .badge {
      position: absolute;
      top: 0;
      right: 6px;
      background: #fa5151;
      color: #fff;
      font-size: 10px;
      min-width: 16px;
      height: 16px;
      padding: 0 4px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: bold;
      border: 1px solid #fff;
    }
  }
}

/* 消息列表 */
.chat-list-section {
  padding: 8px 0;

  .loading {
    text-align: center;
    padding: 40px;
    color: #9499a0;
  }
}

.chat-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f6f7f8;

  &:last-child {
    border-bottom: none;
  }

  &:active {
    background: #f9f9fa;
  }

  .avatar-wrap {
    position: relative;
    width: 44px;
    height: 44px;
    flex-shrink: 0;

    .avatar {
      width: 100%;
      height: 100%;
      border-radius: 50%;
      object-fit: cover;
      background: #f1f2f3;
    }

    &.system-avatar {
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;

      &.blue-light {
        background: #23ade5;
      }

      &.pink-light {
        background: #ff5c7c;
      }
    }

    .dot {
      position: absolute;
      top: 0;
      right: 0;
      width: 8px;
      height: 8px;
      background: #fa5151;
      border-radius: 50%;
      border: 1px solid #fff;
    }
  }

  .chat-content {
    flex: 1;
    margin-left: 12px;
    overflow: hidden;

    .chat-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 4px;

      .name {
        font-size: 14px;
        font-weight: 500;
        color: #18191c;

        &.system-title {
          color: #fb7299; /* “玩转Code” or system titles pink theme */
        }
      }

      .time {
        font-size: 11px;
        color: #9499a0;
      }
    }

    .last-msg {
      font-size: 12px;
      color: #9499a0;
      white-space: nowrap;
      overflow: hidden;
      text-transform: none;
      text-overflow: ellipsis;
    }
  }
}
</style>
