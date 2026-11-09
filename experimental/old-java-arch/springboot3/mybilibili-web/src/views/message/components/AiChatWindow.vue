<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { aiChatApi } from '../../../api/aiChat.js'

const conversations = ref([])
const activeConversationId = ref(null)
const messages = ref([])
const loading = ref(false)
const isStreaming = ref(false)
const streamingContent = ref('')
const messageText = ref('')
const messageListRef = ref(null)
const maxLength = 1000

const currentUserId = computed(() => {
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    return user.id || null
  } catch (e) { return null }
})

const currentUserAvatar = computed(() => {
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    return user.avatar || '/default-avatar.svg'
  } catch (e) { return '/default-avatar.svg' }
})

const activeConversation = computed(() => {
  return conversations.value.find(c => c.id === activeConversationId.value) || null
})

const canSend = computed(() => {
  return messageText.value.trim().length > 0 && messageText.value.length <= maxLength && !isStreaming.value
})

// 流式消息（虚拟消息对象用于渲染流式内容）
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

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

// 加载会话列表
const loadConversations = async () => {
  try {
    const res = await aiChatApi.getConversations()
    if (res.code === 200) {
      conversations.value = res.data || []
    }
  } catch (e) {
    console.error('加载会话失败:', e)
  }
}

// 加载历史消息
const loadMessages = async (conversationId) => {
  if (!conversationId) return
  loading.value = true
  try {
    const res = await aiChatApi.getMessages(conversationId)
    if (res.code === 200) {
      messages.value = res.data || []
    }
  } catch (e) {
    console.error('加载消息失败:', e)
  } finally {
    loading.value = false
  }
}

// 选择会话
const switchConversation = (id) => {
  if (isStreaming.value) return
  activeConversationId.value = id
  messages.value = []
  streamingContent.value = ''
  if (id) {
    loadMessages(id)
  }
}

// 新建会话
const handleNewChat = async () => {
  if (isStreaming.value) {
    ElMessage.warning('请等待当前回复完成')
    return
  }
  try {
    const res = await aiChatApi.createConversation()
    if (res.code === 200) {
      const conv = res.data
      conversations.value.unshift(conv)
      activeConversationId.value = conv.id
      messages.value = []
      streamingContent.value = ''
    }
  } catch (e) {
    ElMessage.error('创建会话失败')
  }
}

// 删除会话
const handleDeleteConversation = async (id, event) => {
  event.stopPropagation()
  try {
    const res = await aiChatApi.deleteConversation(id)
    if (res.code === 200) {
      conversations.value = conversations.value.filter(c => c.id !== id)
      if (activeConversationId.value === id) {
        activeConversationId.value = null
        messages.value = []
      }
    }
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

// 发送消息
const handleSend = () => {
  if (!canSend.value) return
  if (!activeConversationId.value) {
    ElMessage.warning('请先创建或选择一个会话')
    return
  }

  const content = messageText.value.trim()
  messageText.value = ''
  isStreaming.value = true
  streamingContent.value = ''

  // 立即添加用户消息到列表
  messages.value.push({
    id: 'temp-' + Date.now(),
    role: 'user',
    content: content,
    createdAt: new Date().toISOString()
  })

  // SSE 流式调用
  aiChatApi.sendMessage(activeConversationId.value, content, {
    onStart: (data) => {
      streamingContent.value = ''
    },
    onData: (chunk) => {
      streamingContent.value += chunk
      scrollToBottom()
    },
    onDone: (data) => {
      // 流完成，保存 assistant 消息到列表
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

      // 如果返回了新的 conversationId，更新它
      if (data && data.conversationId) {
        activeConversationId.value = data.conversationId
      }
      // 刷新会话列表获取最新标题
      loadConversations()
      scrollToBottom()
    },
    onError: (err) => {
      ElMessage.error(err || '回复失败，请重试')
      streamingContent.value = ''
      isStreaming.value = false
    }
  })
}

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

watch(isStreaming, scrollToBottom)

onMounted(() => {
  loadConversations()
})
</script>

<template>
  <div class="ai-chat-window">
    <!-- 左侧会话列表 -->
    <div class="conversation-panel">
      <div class="panel-header">
        <span class="panel-title">对话列表</span>
        <el-button type="primary" size="small" :icon="Plus" @click="handleNewChat" :disabled="isStreaming">
          新建对话
        </el-button>
      </div>
      <div class="conversation-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          :class="['conv-item', { active: conv.id === activeConversationId }]"
          @click="switchConversation(conv.id)"
        >
          <div class="conv-info">
            <span class="conv-title">{{ conv.title || 'AI客服对话' }}</span>
          </div>
          <el-button
            link
            type="danger"
            size="small"
            :icon="Delete"
            class="conv-delete"
            @click="handleDeleteConversation(conv.id, $event)"
          />
        </div>
        <div v-if="conversations.length === 0" class="empty-conv">
          <span>暂无对话</span>
        </div>
      </div>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-main">
      <!-- 消息列表 -->
      <div ref="messageListRef" class="message-list" v-loading="loading">
        <div v-if="displayMessages.length === 0 && !loading" class="empty-messages">
          <el-empty description="有什么可以帮你的？开始对话吧" />
        </div>
        <div
          v-for="msg in displayMessages"
          :key="msg.id"
          :class="['message-item', { 'message-self': msg.role === 'user' }]"
        >
          <el-avatar
            :size="36"
            :src="msg.role === 'user' ? currentUserAvatar : '/ai-avatar.svg'"
            class="message-avatar"
          />
          <div class="message-content">
            <div :class="['message-bubble', msg.role === 'user' ? 'bubble-user' : 'bubble-ai']">
              <span class="message-text">{{ msg.content }}</span>
              <span v-if="msg.id === 'streaming'" class="streaming-cursor">|</span>
            </div>
            <span v-if="msg.id !== 'streaming'" class="message-time">{{ formatTime(msg.createdAt) }}</span>
          </div>
        </div>
        <!-- AI输入指示器 -->
        <div v-if="isStreaming && !streamingContent" class="message-item">
          <el-avatar :size="36" src="/ai-avatar.svg" class="message-avatar" />
          <div class="message-content">
            <div class="message-bubble bubble-ai">
              <span class="typing-indicator">
                <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="chat-input-area">
        <div class="input-wrapper">
          <el-input
            v-model="messageText"
            type="textarea"
            :rows="3"
            :maxlength="maxLength"
            :placeholder="isStreaming ? 'AI正在回复...' : '输入消息，Enter发送'"
            resize="none"
            :disabled="isStreaming"
            @keydown="handleKeydown"
          />
        </div>
        <div class="input-footer">
          <span class="char-count">{{ messageText.length }}/{{ maxLength }}</span>
          <el-button
            type="primary"
            :disabled="!canSend"
            :loading="isStreaming"
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
.ai-chat-window {
  flex: 1;
  display: flex;
  min-width: 0;
  height: 100%;
  background-color: #f5f7fa;
}

.conversation-panel {
  width: 240px;
  background-color: #fff;
  border-right: 1px solid #e3e5e7;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #e3e5e7;
}

.panel-title {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.conv-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
  transition: background-color 0.2s;
}

.conv-item:hover {
  background-color: #f5f7fa;
}

.conv-item.active {
  background-color: rgba(0, 161, 214, 0.1);
}

.conv-info {
  flex: 1;
  min-width: 0;
}

.conv-title {
  font-size: 13px;
  color: #18191c;
  white-space: nowrap;
  display: block;
}

.conv-delete {
  opacity: 0;
  transition: opacity 0.2s;
}

.conv-item:hover .conv-delete {
  opacity: 1;
}

.empty-conv {
  text-align: center;
  color: #c0c4cc;
  padding: 20px;
  font-size: 13px;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.empty-messages {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.message-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.message-self {
  flex-direction: row-reverse;
}

.message-self .message-content {
  align-items: flex-end;
}

.message-avatar {
  flex-shrink: 0;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 65%;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  word-break: break-word;
  line-height: 1.6;
}

.bubble-user {
  background-color: #00a1d6;
  color: #fff;
}

.bubble-ai {
  background-color: #fff;
  color: #18191c;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.message-text {
  font-size: 14px;
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

.typing-indicator {
  display: inline-flex;
  gap: 2px;
}

.typing-indicator .dot {
  animation: dotBounce 1.4s infinite;
  font-size: 20px;
  line-height: 1;
}

.typing-indicator .dot:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator .dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes dotBounce {
  0%, 60%, 100% { opacity: 0.3; }
  30% { opacity: 1; }
}

.message-time {
  font-size: 12px;
  color: #9499a0;
}

.chat-input-area {
  background-color: #fff;
  border-top: 1px solid #e3e5e7;
  padding: 16px 20px;
  flex-shrink: 0;
}

.input-wrapper {
  margin-bottom: 8px;
}

.input-wrapper :deep(.el-textarea__inner) {
  background-color: #f5f7fa;
  border: none;
  border-radius: 8px;
}

.input-wrapper :deep(.el-textarea__inner:focus) {
  background-color: #fff;
  box-shadow: 0 0 0 1px #00a1d6;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.char-count {
  font-size: 12px;
  color: #9499a0;
}
</style>