<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getPendingSessions,
  getSessionMessages,
  sendReply,
  resolveSession,
  getPendingCount
} from '../api/customerSession'

// 状态
const loading = ref(false)
const sessions = ref([])
const pendingCount = ref(0)
const currentSession = ref(null)
const messages = ref([])
const messageInput = ref('')
const sending = ref(false)
const adminId = ref(parseInt(localStorage.getItem('admin_user')?.id || '1'))

// 轮询定时器
let pollTimer = null

// 格式化时间
const formatTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${month}-${day} ${hours}:${minutes}`
}

// 加载待处理会话列表
const loadSessions = async () => {
  try {
    const res = await getPendingSessions()
    if (res.code === 200 || res.success) {
      sessions.value = res.data || []
    }
  } catch (error) {
    console.error('加载会话列表失败', error)
  }
}

// 加载待处理数量
const loadPendingCount = async () => {
  try {
    const res = await getPendingCount()
    if (res.code === 200 || res.success) {
      pendingCount.value = res.data || 0
    }
  } catch (error) {
    console.error('加载待处理数量失败', error)
  }
}

// 加载会话消息
const loadMessages = async (sessionId) => {
  loading.value = true
  try {
    const res = await getSessionMessages(sessionId)
    if (res.code === 200 || res.success) {
      messages.value = res.data || []
      await nextTick()
      scrollToBottom()
    }
  } catch (error) {
    ElMessage.error('加载消息历史失败')
  } finally {
    loading.value = false
  }
}

// 选择会话
const selectSession = async (session) => {
  currentSession.value = session
  await loadMessages(session.sessionId)
}

// 滚动到底部
const scrollToBottom = () => {
  const container = document.querySelector('.messages-container')
  if (container) {
    container.scrollTop = container.scrollHeight
  }
}

// 发送消息
const handleSend = async () => {
  if (!messageInput.value.trim() || !currentSession.value || sending.value) return

  const content = messageInput.value.trim()
  messageInput.value = ''
  sending.value = true

  try {
    const res = await sendReply(currentSession.value.sessionId, adminId.value, content)
    if (res.code === 200 || res.success) {
      // 乐观更新：立即显示消息
      messages.value.push({
        id: Date.now(),
        sessionId: currentSession.value.sessionId,
        role: 'HUMAN',
        content: content,
        createdAt: new Date().toISOString()
      })
      await nextTick()
      scrollToBottom()
    } else {
      ElMessage.error(res.message || '发送失败')
    }
  } catch (error) {
    ElMessage.error('发送失败')
  } finally {
    sending.value = false
  }
}

// 标记已处理
const handleResolve = async () => {
  if (!currentSession.value) return

  try {
    const res = await resolveSession(currentSession.value.sessionId)
    if (res.code === 200 || res.success) {
      ElMessage.success('会话已标记为已处理')
      currentSession.value = null
      messages.value = []
      await loadSessions()
      await loadPendingCount()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 获取角色显示
const getRoleLabel = (role) => {
  const map = {
    'USER': '用户',
    'ASSISTANT': 'AI',
    'HUMAN': '客服',
    'SYSTEM': '系统'
  }
  return map[role] || role
}

// 获取角色样式
const getRoleClass = (role) => {
  return {
    'message-user': role === 'USER',
    'message-human': role === 'HUMAN',
    'message-system': role === 'SYSTEM'
  }
}

onMounted(async () => {
  await loadSessions()
  await loadPendingCount()

  // 每10秒轮询一次
  pollTimer = setInterval(async () => {
    await loadSessions()
    await loadPendingCount()
  }, 10000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})
</script>

<template>
  <div class="customer-chat-page">
    <h2 class="page-title">客服会话</h2>

    <div class="chat-container">
      <!-- 左侧会话列表 -->
      <div class="session-list">
        <div class="session-list-header">
          <span class="header-title">待处理会话</span>
          <el-badge :value="pendingCount" :hidden="pendingCount === 0" type="danger" />
        </div>

        <div class="session-items">
          <div
            v-for="session in sessions"
            :key="session.sessionId"
            class="session-item"
            :class="{ active: currentSession?.sessionId === session.sessionId }"
            @click="selectSession(session)"
          >
            <div class="session-user">用户 {{ session.userId }}</div>
            <div class="session-preview">{{ session.lastMessage || '暂无消息' }}</div>
            <div class="session-time">{{ formatTime(session.lastMessageTime) }}</div>
          </div>

          <div v-if="sessions.length === 0" class="empty-tip">
            暂无待处理会话
          </div>
        </div>
      </div>

      <!-- 右侧聊天窗口 -->
      <div class="chat-window">
        <!-- 顶部标题栏 -->
        <div class="chat-header">
          <template v-if="currentSession">
            <span class="chat-title">与用户 {{ currentSession.userId }} 的会话</span>
            <el-button type="success" size="small" @click="handleResolve">
              标记已处理
            </el-button>
          </template>
          <span v-else class="chat-title placeholder">请选择左侧会话</span>
        </div>

        <!-- 消息列表 -->
        <div class="messages-container" v-loading="loading">
          <template v-if="currentSession">
            <div
              v-for="msg in messages"
              :key="msg.id"
              class="message-item"
              :class="getRoleClass(msg.role)"
            >
              <div class="message-role">{{ getRoleLabel(msg.role) }}</div>
              <div class="message-content">{{ msg.content }}</div>
              <div class="message-time">{{ formatTime(msg.createdAt) }}</div>
            </div>

            <div v-if="messages.length === 0" class="empty-messages">
              暂无消息记录
            </div>
          </template>
          <div v-else class="empty-chat">
            <el-icon :size="60" color="#ccc"><ChatDotRound /></el-icon>
            <p>请从左侧选择会话</p>
          </div>
        </div>

        <!-- 输入区域 -->
        <div class="input-area">
          <el-input
            v-model="messageInput"
            type="textarea"
            :rows="3"
            placeholder="输入回复内容..."
            :disabled="!currentSession || sending"
            @keyup.ctrl.enter="handleSend"
          />
          <el-button
            type="primary"
            :disabled="!currentSession || !messageInput.trim() || sending"
            :loading="sending"
            @click="handleSend"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.customer-chat-page {
  padding: 20px;
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.chat-container {
  flex: 1;
  display: flex;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

/* 左侧会话列表 */
.session-list {
  width: 280px;
  border-right: 1px solid #eee;
  display: flex;
  flex-direction: column;
}

.session-list-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-title {
  font-weight: 600;
  font-size: 16px;
}

.session-items {
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: background-color 0.2s;
}

.session-item:hover {
  background-color: #f5f7fa;
}

.session-item.active {
  background-color: #ecf5ff;
  border-left: 3px solid #409eff;
}

.session-user {
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.session-preview {
  font-size: 13px;
  color: #999;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.session-time {
  font-size: 12px;
  color: #ccc;
}

.empty-tip {
  padding: 40px 16px;
  text-align: center;
  color: #999;
}

/* 右侧聊天窗口 */
.chat-window {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.chat-header {
  padding: 12px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chat-title {
  font-weight: 600;
  font-size: 16px;
}

.chat-title.placeholder {
  color: #999;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.empty-chat {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.empty-chat p {
  margin-top: 16px;
}

.empty-messages {
  text-align: center;
  color: #999;
  padding: 40px;
}

/* 消息气泡 */
.message-item {
  margin-bottom: 20px;
  max-width: 70%;
}

.message-user {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  margin-left: auto;
}

.message-user .message-content {
  background-color: #95ec69;
  color: #000;
  border-radius: 12px 12px 0 12px;
}

.message-human {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.message-human .message-content {
  background-color: #409eff;
  color: #fff;
  border-radius: 12px 12px 12px 0;
}

.message-system {
  display: flex;
  justify-content: center;
  max-width: 100%;
}

.message-system .message-content {
  background-color: #f5f5f5;
  color: #666;
  font-size: 13px;
  padding: 6px 12px;
  border-radius: 12px;
  text-align: center;
}

.message-role {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.message-content {
  padding: 10px 14px;
  line-height: 1.5;
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  color: #ccc;
  margin-top: 4px;
}

/* 输入区域 */
.input-area {
  padding: 12px 20px;
  border-top: 1px solid #eee;
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-area .el-textarea {
  flex: 1;
}

.input-area .el-button {
  height: 68px;
  padding: 0 24px;
}
</style>
