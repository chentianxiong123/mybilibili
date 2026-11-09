<script setup>
import { ref, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, CircleCheckFilled, Loading, Service } from '@element-plus/icons-vue'
import { customerServiceApi } from '../api/customerService.js'

const messages = ref([])
const inputText = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const messageListRef = ref(null)
const eventSource = ref(null)
const aiOnline = ref(true)
const loadingHistory = ref(false)

// 合并显示的消息（包含流式正在输出的内容）
const displayMessages = computed(() => {
  const msgs = [...messages.value]
  if (isStreaming.value && streamingContent.value) {
    msgs.push({
      id: 'streaming',
      role: 'assistant',
      content: streamingContent.value,
      createdAt: new Date().toISOString()
    })
  }
  return msgs
})

// 自动滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

// 获取 userId
const getUserId = () => {
  return localStorage.getItem('user_id')
}

// 加载历史消息
const loadHistory = async () => {
  const userId = getUserId()
  if (!userId) return

  loadingHistory.value = true
  try {
    const history = await customerServiceApi.getHistory(userId)
    if (history && history.length > 0) {
      messages.value = history.map((msg, idx) => ({
        id: msg.id || `history-${idx}`,
        role: msg.role,
        content: msg.content,
        createdAt: msg.createdAt || new Date().toISOString()
      }))
      scrollToBottom()
    }
  } catch (e) {
    console.warn('加载历史消息失败', e)
  } finally {
    loadingHistory.value = false
  }
}

// 发送消息
const handleSend = () => {
  const content = inputText.value.trim()
  if (!content || isStreaming.value) return

  const userId = getUserId()
  if (!userId) {
    ElMessage.warning('请先登录')
    return
  }

  inputText.value = ''
  isStreaming.value = true
  streamingContent.value = ''

  // 先本地显示用户消息
  messages.value.push({
    id: 'user-' + Date.now(),
    role: 'user',
    content: content,
    createdAt: new Date().toISOString()
  })

  scrollToBottom()

  eventSource.value = customerServiceApi.sendMessage(userId, content, {
    onData: (chunk) => {
      streamingContent.value += chunk
      scrollToBottom()
    },
    onDone: () => {
      // 流式结束，把累积的内容作为一条 AI 消息保存
      if (streamingContent.value) {
        messages.value.push({
          id: 'ai-' + Date.now(),
          role: 'assistant',
          content: streamingContent.value,
          createdAt: new Date().toISOString()
        })
      }
      streamingContent.value = ''
      isStreaming.value = false
      eventSource.value = null
      scrollToBottom()
    },
    onError: (err) => {
      ElMessage.error(err || '回复失败，请重试')
      streamingContent.value = ''
      isStreaming.value = false
      eventSource.value = null
    }
  })
}

// 键盘发送
const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// 转人工
const handleTransfer = async () => {
  const userId = getUserId()
  if (!userId) {
    ElMessage.warning('请先登录')
    return
  }

  try {
    const confirmed = await ElMessageBox.confirm(
      '是否转人工客服？',
      '转人工',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    if (confirmed) {
      // 显示系统消息
      messages.value.push({
        id: 'system-' + Date.now(),
        role: 'system',
        content: '正在为您转接人工客服...',
        createdAt: new Date().toISOString()
      })
      scrollToBottom()

      await customerServiceApi.transferToHuman(userId, null, '用户主动转人工')

      // 显示转人工完成消息
      messages.value.push({
        id: 'system-' + Date.now() + 1,
        role: 'system',
        content: '[已转人工服务，请稍候]',
        createdAt: new Date().toISOString()
      })
      scrollToBottom()
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error('转人工失败', e)
    }
  }
}

onMounted(() => {
  loadHistory()
})
</script>

<template>
  <div class="customer-service-view">
    <!-- Header -->
    <div class="cs-header">
      <div class="header-left">
        <el-icon :size="22"><ChatDotRound /></el-icon>
        <span class="header-title">在线客服</span>
      </div>
      <div class="header-right">
        <span class="online-indicator">
          <el-icon :size="12" :class="aiOnline ? 'online-dot' : 'offline-dot'">
            <CircleCheckFilled />
          </el-icon>
          {{ aiOnline ? 'AI在线' : 'AI离线' }}
        </span>
      </div>
    </div>

    <!-- Message List -->
    <div ref="messageListRef" class="message-list">
      <!-- Loading History -->
      <div v-if="loadingHistory" class="loading-history">
        <el-icon :size="20" class="loading-icon"><Loading /></el-icon>
        <span>加载历史消息...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="displayMessages.length === 0" class="empty-state">
        <el-icon :size="48" class="empty-icon"><ChatDotRound /></el-icon>
        <p class="empty-title">你好！我是哔哩助手</p>
        <p class="empty-desc">有什么可以帮助你的吗？</p>
      </div>

      <!-- Messages -->
      <div
        v-for="msg in displayMessages"
        :key="msg.id"
        :class="['message-item', { 'message-self': msg.role === 'user' }, { 'message-system': msg.role === 'system' }]"
      >
        <div class="message-content" v-if="msg.role !== 'system'">
          <div :class="['message-bubble', msg.role === 'user' ? 'bubble-user' : 'bubble-ai']">
            <span class="message-role-tag">{{ msg.role === 'user' ? '我' : '哔哩助手' }}</span>
            <span class="message-text">{{ msg.content }}</span>
            <span v-if="msg.id === 'streaming'" class="streaming-cursor">|</span>
          </div>
        </div>
        <div v-else class="message-system-content">
          <span class="message-system-text">{{ msg.content }}</span>
        </div>
      </div>

      <!-- Streaming Typing Indicator -->
      <div v-if="isStreaming && !streamingContent" class="message-item">
        <div class="message-content">
          <div class="message-bubble bubble-ai">
            <span class="message-role-tag">哔哩助手</span>
            <span class="typing-indicator">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="input-area">
      <el-input
        v-model="inputText"
        type="textarea"
        :rows="3"
        placeholder="输入你的问题，Enter发送，Shift+Enter换行"
        resize="none"
        :disabled="isStreaming"
        @keydown="handleKeydown"
      />
      <div class="input-footer">
        <el-button
          text
          @click="handleTransfer"
          :disabled="isStreaming"
          style="color: #909399;"
        >
          <el-icon><Service /></el-icon>
          转人工
        </el-button>
        <el-button
          type="primary"
          :disabled="!inputText.trim() || isStreaming"
          :loading="isStreaming"
          @click="handleSend"
        >
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.customer-service-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #f5f7fa;
}

/* Header */
.cs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background-color: #fff;
  border-bottom: 1px solid #e3e5e7;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #00a1d6;
}

.header-title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
}

.online-indicator {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  color: #909399;
}

.online-dot {
  color: #67c23a;
}

.offline-dot {
  color: #909399;
}

/* Message List */
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Loading History */
.loading-history {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #909399;
  font-size: 14px;
  padding: 20px;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Empty State */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  text-align: center;
}

.empty-icon {
  color: #00a1d6;
  margin-bottom: 12px;
  opacity: 0.6;
}

.empty-title {
  font-size: 17px;
  font-weight: 500;
  color: #303133;
  margin: 0 0 8px 0;
}

.empty-desc {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* Message Items */
.message-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.message-self {
  flex-direction: row-reverse;
}

.message-self .message-content {
  align-items: flex-end;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 75%;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 12px;
  word-break: break-word;
  line-height: 1.6;
  font-size: 14px;
  position: relative;
}

.bubble-user {
  background-color: #00a1d6;
  color: #fff;
}

.bubble-ai {
  background-color: #fff;
  color: #303133;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.message-role-tag {
  font-size: 11px;
  opacity: 0.6;
  margin-bottom: 2px;
  display: block;
}

.message-text {
  white-space: pre-wrap;
}

/* System Message */
.message-system {
  justify-content: center;
}

.message-system-content {
  display: flex;
  justify-content: center;
}

.message-system-text {
  font-size: 13px;
  color: #909399;
  background-color: rgba(0, 0, 0, 0.03);
  padding: 6px 12px;
  border-radius: 6px;
  text-align: center;
}

.streaming-cursor {
  animation: blink 1s infinite;
  margin-left: 2px;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

/* Typing Indicator */
.typing-indicator {
  display: inline-flex;
  gap: 3px;
  align-items: center;
}

.typing-indicator .dot {
  width: 6px;
  height: 6px;
  background-color: #00a1d6;
  border-radius: 50%;
  animation: dotBounce 1.4s infinite;
}

.typing-indicator .dot:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator .dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes dotBounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-4px); opacity: 1; }
}

/* Input Area */
.input-area {
  padding: 14px 20px;
  background-color: #fff;
  border-top: 1px solid #e3e5e7;
  flex-shrink: 0;
}

.input-area :deep(.el-textarea__inner) {
  background-color: #f5f7fa;
  border: none;
  border-radius: 10px;
  padding: 10px 14px;
  font-size: 14px;
  line-height: 1.6;
}

.input-area :deep(.el-textarea__inner:focus) {
  background-color: #fff;
  box-shadow: 0 0 0 1px #00a1d6;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.input-footer :deep(.el-button) {
  background-color: #00a1d6;
  border-color: #00a1d6;
}

.input-footer :deep(.el-button:hover) {
  background-color: #00b4e0;
  border-color: #00b4e0;
}
</style>