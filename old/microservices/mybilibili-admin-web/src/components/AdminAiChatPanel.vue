<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Close, Tools, Download, TrendCharts, DataLine, PieChart } from '@element-plus/icons-vue'
import { adminAiApi } from '../api/adminAi.js'

const visible = defineModel('visible', { type: Boolean, default: false })

const messages = ref([])
const inputText = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const toolCalling = ref(false)
const toolCallingName = ref('')
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

  messages.value.push({
    id: 'user-' + Date.now(),
    role: 'user',
    content: content,
    createdAt: new Date().toISOString()
  })

  scrollToBottom()

  eventSource.value = adminAiApi.sendMessage(content, {
    onData: (chunk) => {
      streamingContent.value += chunk
      scrollToBottom()
    },
    onToolCall: (name) => {
      toolCalling.value = true
      toolCallingName.value = name || '处理中'
      scrollToBottom()
    },
    onDone: (data) => {
      let reply = ''
      let render = null

      if (data && typeof data === 'object') {
        reply = data.content || data.reply || ''
        // 兼容 render 字段（可能已经是对象或 JSON 字符串）
        if (data.render) {
          try {
            render = typeof data.render === 'string' ? JSON.parse(data.render) : data.render
          } catch { render = data.render }
        }
      } else if (typeof data === 'string') {
        reply = data
      }

      messages.value.push({
        id: 'ai-' + Date.now(),
        role: 'assistant',
        content: reply,
        render: render,
        createdAt: new Date().toISOString()
      })

      toolCalling.value = false
      streamingContent.value = ''
      isStreaming.value = false
      eventSource.value = null
      scrollToBottom()
    },
    onError: (err) => {
      ElMessage.error(err || '回复失败，请重试')
      toolCalling.value = false
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

// 监听 visible 变化
watch(visible, (val) => {
  if (!val) {
    if (eventSource.value) {
      eventSource.value.abort()
      eventSource.value = null
    }
    isStreaming.value = false
    streamingContent.value = ''
    toolCalling.value = false
  }
})

// 解析 render data 里的行数据
const getTableRows = (data) => {
  if (!data || typeof data !== 'object') return []
  if (Array.isArray(data)) return data
  // 取第一个数组字段
  for (const key of Object.keys(data)) {
    if (Array.isArray(data[key])) return data[key]
  }
  return []
}

// 获取表格列（从第一条数据推断）
const getTableColumns = (rows) => {
  if (!rows || rows.length === 0) return []
  return Object.keys(rows[0]).map(k => ({ prop: k, label: k }))
}

// 判断 chartType
const getChartIcon = (chartType) => {
  switch (chartType) {
    case 'line': return TrendCharts
    case 'pie': return PieChart
    case 'table': return DataLine
    default: return DataLine
  }
}

// 数字摘要格式化
const formatNumberStat = (data) => {
  if (!data || typeof data !== 'object') return null
  return Object.entries(data).map(([key, val]) => ({
    label: key,
    value: typeof val === 'number' ? val.toLocaleString() : String(val)
  }))
}
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

            <!-- Tool Calling Indicator -->
            <div v-if="toolCalling && msg.id === 'streaming'" class="tool-calling">
              <el-icon :size="14" class="tool-icon-spin"><Tools /></el-icon>
              <span>正在查询：{{ toolCallingName }}</span>
            </div>

            <!-- Render Block: Number Cards -->
            <template v-if="msg.render?.chartType === 'number' && formatNumberStat(msg.render?.data)">
              <div class="stats-grid">
                <div v-for="stat in formatNumberStat(msg.render.data)" :key="stat.label" class="stat-card">
                  <span class="stat-label">{{ stat.label }}</span>
                  <span class="stat-value">{{ stat.value }}</span>
                </div>
              </div>
            </template>

            <!-- Render Block: Table -->
            <template v-if="msg.render?.chartType === 'table' && getTableRows(msg.render?.data).length">
              <div class="table-block">
                <el-tag :type="'primary'" size="small">
                  <el-icon><DataLine /></el-icon>
                  {{ msg.render.title || '数据表格' }}
                </el-tag>
                <el-table
                  :data="getTableRows(msg.render.data)"
                  size="small"
                  stripe
                  border
                  class="stat-table"
                  max-height="200"
                >
                  <el-table-column
                    v-for="col in getTableColumns(getTableRows(msg.render.data))"
                    :key="col.prop"
                    :prop="col.prop"
                    :label="col.label"
                    show-overflow-tooltip
                  />
                </el-table>
              </div>
            </template>

            <!-- Render Block: Pie -->
            <template v-if="msg.render?.chartType === 'pie' && getTableRows(msg.render?.data).length">
              <div class="chart-block">
                <el-tag type="primary" size="small">
                  <el-icon><PieChart /></el-icon>
                  {{ msg.render.title || '分布图' }}
                </el-tag>
                <div class="pie-list">
                  <div v-for="row in getTableRows(msg.render.data)" :key="row.label || row.name || row.key" class="pie-row">
                    <span class="pie-label">{{ row.label || row.name || row.key }}</span>
                    <span class="pie-value">{{ typeof row.value === 'number' ? row.value.toLocaleString() : row.value }}</span>
                  </div>
                </div>
              </div>
            </template>

            <!-- Render Block: Line -->
            <template v-if="msg.render?.chartType === 'line' && getTableRows(msg.render?.data).length">
              <div class="chart-block">
                <el-tag type="primary" size="small">
                  <el-icon><TrendCharts /></el-icon>
                  {{ msg.render.title || '趋势图' }}
                </el-tag>
                <div class="line-list">
                  <div v-for="row in getTableRows(msg.render.data)" :key="row.date || row.label || row.x" class="line-row">
                    <span class="line-label">{{ row.date || row.label || row.x }}</span>
                    <span class="line-bar" :style="{ width: Math.min(100, (row.value || row.y || 0) / (getTableRows(msg.render.data)[0]?.value || 1) * 100) + '%' }"></span>
                    <span class="line-value">{{ row.value || row.y }}</span>
                  </div>
                </div>
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

/* Tool Calling */
.tool-calling {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  padding: 4px 8px;
  background: #f5f7fa;
  border-radius: 6px;
  width: fit-content;
}

.tool-icon-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Stats Grid (number cards) */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 8px;
}

.stat-card {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  color: #909399;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

/* Table Block */
.table-block {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-table {
  font-size: 12px;
}

/* Chart Block */
.chart-block {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* Pie List */
.pie-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pie-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.pie-label {
  flex: 1;
  color: #606266;
}

.pie-value {
  font-weight: 500;
  color: #303133;
  min-width: 40px;
  text-align: right;
}

/* Line Chart */
.line-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.line-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.line-label {
  width: 60px;
  color: #909399;
  flex-shrink: 0;
}

.line-bar {
  height: 8px;
  background: linear-gradient(90deg, #00a1d6, #00d6b2);
  border-radius: 4px;
  min-width: 4px;
  max-width: 120px;
  transition: width 0.3s;
}

.line-value {
  font-weight: 500;
  color: #303133;
  min-width: 40px;
  text-align: right;
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