<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Close } from '@element-plus/icons-vue'
import { adminAiApi } from '../api/adminAi.js'

const visible = defineModel('visible', { type: Boolean, default: false })

const messages = ref([])
const inputText = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const messageListRef = ref(null)
const eventSource = ref(null)

// 流式消息合并显示
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

// 发送消息
const handleSend = () => {
  const content = inputText.value.trim()
  if (!content || isStreaming.value) return

  inputText.value = ''
  isStreaming.value = true
  streamingContent.value = ''

  // 添加用户消息
  messages.value.push({
    id: 'user-' + Date.now(),
    role: 'user',
    content: content,
    createdAt: new Date().toISOString()
  })

  scrollToBottom()

  // 调用 SSE 接口
  eventSource.value = adminAiApi.sendMessage(content, {
    onData: (chunk) => {
      streamingContent.value += chunk
      scrollToBottom()
    },
    onDone: (data) => {
      // 提取 render 数据
      let render = null
      if (data && typeof data === 'object') {
        render = data.render || null
      }

      // 添加 AI 消息
      messages.value.push({
        id: 'ai-' + Date.now(),
        role: 'assistant',
        content: streamingContent.value,
        render: render,
        createdAt: new Date().toISOString()
      })

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

// 关闭面板
const handleClose = () => {
  if (eventSource.value) {
    eventSource.value.abort()
    eventSource.value = null
  }
  visible.value = false
}

// 监听 visible 变化，关闭时中止请求
watch(visible, (val) => {
  if (!val) {
    if (eventSource.value) {
      eventSource.value.abort()
      eventSource.value = null
    }
    isStreaming.value = false
    streamingContent.value = ''
  }
})
</script>

<template>
  <transition name="slide-up">
    <div v-if="visible" class="admin-ai-chat-panel">
      <!-- Header -->
      <div class="panel-header">
        <div class="header-left">
          <el-icon :size="20"><ChatDotRound /></el-icon>
          <span class="header-title">管理助手</span>
        </div>
        <el-button link type="default" @click="handleClose">
          <el-icon :size="18"><Close /></el-icon>
        </el-button>
      </div>

      <!-- Message List -->
      <div ref="messageListRef" class="message-list">
        <!-- Empty State -->
        <div v-if="displayMessages.length === 0" class="empty-state">
          <el-icon :size="48" class="empty-icon"><ChatDotRound /></el-icon>
          <p class="empty-title">你好！我是管理助手</p>
          <p class="empty-desc">可以问我平台数据、统计报表等问题</p>
        </div>

        <!-- Messages -->
        <div
          v-for="msg in displayMessages"
          :key="msg.id"
          :class="['message-item', { 'message-self': msg.role === 'user' }]"
        >
          <div class="message-content">
            <div :class="['message-bubble', msg.role === 'user' ? 'bubble-user' : 'bubble-ai']">
              <span class="message-text">{{ msg.content }}</span>
              <span v-if="msg.id === 'streaming'" class="streaming-cursor">|</span>
            </div>

            <!-- Render Block -->
            <template v-if="msg.render">
              <el-divider />
              <div class="render-block">
                <el-tag size="small" type="primary">{{ msg.render.title || '数据分析' }}</el-tag>
                <span v-if="msg.render.chartType" class="chart-type">{{ msg.render.chartType }}</span>
              </div>
            </template>
          </div>
        </div>

        <!-- Streaming Indicator -->
        <div v-if="isStreaming && !streamingContent" class="message-item">
          <div class="message-content">
            <div class="message-bubble bubble-ai">
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
          :rows="2"
          placeholder="输入你的问题，Enter发送"
          resize="none"
          :disabled="isStreaming"
          @keydown="handleKeydown"
        />
        <div class="input-footer">
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
  </transition>
</template>

<style scoped>
.admin-ai-chat-panel {
  position: fixed;
  bottom: 140px;
  right: 24px;
  width: 400px;
  height: 520px;
  z-index: 2001;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e3e5e7;
}

/* Header */
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #00a1d6, #00d6b2);
  color: #fff;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 16px;
  font-weight: 500;
}

.panel-header :deep(.el-button) {
  color: #fff;
}

.panel-header :deep(.el-button:hover) {
  background-color: rgba(255, 255, 255, 0.2);
}

/* Message List */
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
  gap: 12px;
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
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin: 0 0 8px 0;
}

.empty-desc {
  font-size: 13px;
  color: #909399;
  margin: 0;
}

/* Message Items */
.message-item {
  display: flex;
  gap: 8px;
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
  max-width: 80%;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 12px;
  word-break: break-word;
  line-height: 1.5;
  font-size: 14px;
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

.message-text {
  white-space: pre-wrap;
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

/* Render Block */
.render-block {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.chart-type {
  font-size: 12px;
  color: #909399;
}

/* Input Area */
.input-area {
  padding: 12px 16px;
  background-color: #fff;
  border-top: 1px solid #e3e5e7;
  flex-shrink: 0;
}

.input-area :deep(.el-textarea__inner) {
  background-color: #f5f7fa;
  border: none;
  border-radius: 8px;
  padding: 8px 12px;
}

.input-area :deep(.el-textarea__inner:focus) {
  background-color: #fff;
  box-shadow: 0 0 0 1px #00a1d6;
}

.input-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

/* Slide Up Transition */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>