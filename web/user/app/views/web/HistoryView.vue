<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Delete, List, Clock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { watchHistoryApi } from '@/api/watchHistory.ts'
import { videoApi } from '@/api/client'

const router = useRouter()

// 加载状态
const loading = ref(false)

// 视图模式：grid 或 list
const viewMode = ref('grid')

// 筛选标签页（对齐 B 站：综合/视频/直播/专栏）
const tabs = ['综合', '视频', '直播', '专栏']
const activeTab = ref('综合')

// 是否开启记录浏览历史
const recordHistory = ref(true)

// 搜索关键词
const searchKeyword = ref('')

// 批量管理模式
const isBatchMode = ref(false)

// 选中的历史记录ID
const selectedIds = ref([])

// 分页信息
const pagination = ref({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 历史记录数据
const historyList = ref([])

// 按日期分组的历史记录
const groupedHistory = computed(() => {
  let filtered = historyList.value

  // 搜索过滤
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    filtered = filtered.filter(item =>
      item.title.toLowerCase().includes(keyword) ||
      item.author.toLowerCase().includes(keyword)
    )
  }

  const groups = {}

  filtered.forEach(item => {
    const date = new Date(item.watchedAt)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)

    let dateKey
    if (isSameDay(date, today)) {
      dateKey = '今天'
    } else if (isSameDay(date, yesterday)) {
      dateKey = '昨天'
    } else {
      dateKey = formatDate(date)
    }

    if (!groups[dateKey]) {
      groups[dateKey] = []
    }
    groups[dateKey].push(item)
  })

  return groups
})

// 判断是否同一天
const isSameDay = (date1, date2) => {
  return date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
}

// 格式化日期
const formatDate = (date) => {
  const month = date.getMonth() + 1
  const day = date.getDate()
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const weekDay = weekDays[date.getDay()]
  return `${month}月${day}日 ${weekDay}`
}

// 格式化时间（显示完整日期时间）
const formatTime = (date) => {
  const now = new Date()
  const isToday = isSameDay(date, now)

  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  const timeStr = `${hours}:${minutes}`

  if (isToday) {
    // 今天只显示时间
    return timeStr
  } else {
    // 非今天显示月-日 时:分
    const month = (date.getMonth() + 1).toString().padStart(2, '0')
    const day = date.getDate().toString().padStart(2, '0')
    return `${month}-${day} ${timeStr}`
  }
}

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds || seconds === 0) return '00:00'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`
}

// 格式化进度时间
const formatProgress = (progress, total, percentage) => {
  if (!progress || progress === 0) return ''
  if (percentage && percentage > 0) {
    return `看到 ${formatDuration(progress)} · ${percentage}%`
  }
  return `看到 ${formatDuration(progress)}`
}

// 加载历史记录
const loadHistory = async () => {
  loading.value = true
  try {
    const response = await watchHistoryApi.getWatchHistory(
      pagination.value.currentPage,
      pagination.value.pageSize
    )
    if (response && response.code === 200) {
      // 转换后端数据为前端格式
      historyList.value = (response.data || []).map(item => {
const video = item.video || {}
        const uploader = video.uploader || {}
        return {
          id: item.id,
          videoId: item.videoId || item.video_id,
          title: video.title || '未知视频',
          author: uploader.name || uploader.nickname || video.author || video.userName || video.user_name || '未知UP主',
          cover: video.coverUrl || video.cover_url || video.cover || '',
          duration: item.progressSeconds || item.progress_seconds || 0,
          watchedAt: (item.watchedAt || item.watched_at) ? new Date(item.watchedAt || item.watched_at) : new Date(),
          progress: item.progressSeconds || item.progress_seconds || 0,
          totalDuration: item.videoDuration || item.video_duration || 0,
          watchPercentage: item.watchPercentage || item.watch_percentage || 0,
          isFinished: (item.watchPercentage || item.watch_percentage || 0) >= 90,
          authorAvatar: uploader.avatar || '/default-avatar.svg',
          manuscriptId: video.manuscriptId || video.manuscript_id || video.id
        }
      })
      pagination.value.total = historyList.value.length
    } else {
      ElMessage.error(response?.message || '加载历史记录失败')
    }
  } catch (error) {
    console.error('加载历史记录失败:', error)
    ElMessage.error('加载历史记录失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  // 搜索逻辑已在computed中实现
}

// 清空历史
const clearHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有历史记录吗？此操作不可恢复。',
      '清空历史',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await watchHistoryApi.clearWatchHistory()
    historyList.value = []
    pagination.value.total = 0
    ElMessage.success('历史记录已清空')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清空历史记录失败:', error)
      ElMessage.error('清空历史记录失败')
    }
  }
}

// 切换批量管理模式
const toggleBatchMode = () => {
  isBatchMode.value = !isBatchMode.value
  if (!isBatchMode.value) {
    selectedIds.value = []
  }
}

// 选择/取消选择
const toggleSelect = (id) => {
  const index = selectedIds.value.indexOf(id)
  if (index > -1) {
    selectedIds.value.splice(index, 1)
  } else {
    selectedIds.value.push(id)
  }
}

// 全选当前组
const selectAllInGroup = (groupItems, selected) => {
  if (selected) {
    groupItems.forEach(item => {
      if (!selectedIds.value.includes(item.id)) {
        selectedIds.value.push(item.id)
      }
    })
  } else {
    groupItems.forEach(item => {
      const index = selectedIds.value.indexOf(item.id)
      if (index > -1) {
        selectedIds.value.splice(index, 1)
      }
    })
  }
}

// 批量删除
const batchDelete = async () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要删除的记录')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 条历史记录吗？`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await Promise.all(selectedIds.value.map(id => watchHistoryApi.deleteWatchHistory(id)))
    historyList.value = historyList.value.filter(item => !selectedIds.value.includes(item.id))
    selectedIds.value = []
    isBatchMode.value = false
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除历史记录失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 删除单条记录
const deleteItem = async (item) => {
  try {
    await watchHistoryApi.deleteWatchHistory(item.id)
    const index = historyList.value.findIndex(h => h.id === item.id)
    if (index > -1) {
      historyList.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    console.error('删除历史记录失败:', error)
    ElMessage.error('删除失败')
  }
}

// 点击视频
const handleVideoClick = async (item) => {
  if (isBatchMode.value) {
    toggleSelect(item.id)
  } else {
    const manuscriptId = item.manuscriptId || item.videoId
    if (!manuscriptId) return

    let targetP = 1
    const targetT = Math.max(0, Math.floor(item.progress || 0))

    if (item.manuscriptId && item.videoId) {
      try {
        const response = await videoApi.getVideoByManuscriptId(item.manuscriptId)
        if (response.code === 200 && response.data?.videos?.length) {
          const partIndex = response.data.videos.findIndex(video => video.id === item.videoId)
          if (partIndex >= 0) {
            targetP = partIndex + 1
          }
        }
      } catch (error) {
        console.error('获取稿件分P信息失败:', error)
      }
    }

    router.push({
      path: `/manuscript/${manuscriptId}`,
      query: {
        p: String(targetP),
        ...(targetT > 0 ? { t: String(targetT) } : {})
      }
    })
  }
}

// 切换记录浏览历史
const handleRecordChange = (val) => {
  if (!val) {
    ElMessage.info('已关闭浏览历史记录')
  } else {
    ElMessage.success('已开启浏览历史记录')
  }
}

onMounted(() => {
  loadHistory()
})
</script>

<template>
  <div class="history-page">
    <main class="history-record">
      <div class="history-container">
        <!-- 页面头部：标题 + 记录开关 + 视图切换 -->
        <div class="page-header">
          <h1 class="page-title">
            <el-icon class="page-title__icon"><Clock /></el-icon>历史记录
          </h1>
          <div class="header-actions">
            <div class="record-toggle">
              <el-switch
                v-model="recordHistory"
                class="record-switch"
                active-color="#00aede"
                @change="handleRecordChange"
              />
              <span class="record-toggle__label">记录浏览历史</span>
            </div>
            <button
              :class="['view-toggle', { 'view-toggle--list': viewMode === 'list' }]"
              :title="viewMode === 'grid' ? '切换为列表视图' : '切换为网格视图'"
              @click="viewMode = viewMode === 'grid' ? 'list' : 'grid'"
            >
              <el-icon><List /></el-icon>
            </button>
          </div>
        </div>
      </div>

      <!-- 筛选/搜索粘性条 -->
      <div class="page-breadcrumbs">
        <div class="history-container breadcrumbs-inner">
          <div class="breadcrumbs-top">
            <div class="breadcrumbs-left">
              <div class="radio-tabs">
                <div
                  v-for="tab in tabs"
                  :key="tab"
                  :class="['radio-tabs__item', { 'radio-tabs__item--active': activeTab === tab }]"
                  @click="activeTab = tab"
                >
                  <div class="radio-tabs__content">{{ tab }}</div>
                </div>
                <div class="radio-tabs__cursor"></div>
              </div>
            </div>
            <div class="breadcrumbs-right">
              <div class="search-bar">
                <div class="search-input-wrap">
                  <input
                    v-model="searchKeyword"
                    class="search-input"
                    type="text"
                    placeholder="搜索标题/up主昵称"
                    @keyup.enter="handleSearch"
                  />
                </div>
                <div class="search-btn" @click="handleSearch">
                  <el-icon><Search /></el-icon>
                </div>
              </div>
              <button class="history-btn clear-btn" @click="clearHistory">
                <el-icon class="history-btn__icon"><Delete /></el-icon>
                <span class="history-btn__text">清空历史</span>
              </button>
              <button
                :class="['history-btn', 'batch-btn', { 'batch-btn--active': isBatchMode }]"
                @click="toggleBatchMode"
              >
                <el-icon class="history-btn__icon"><List /></el-icon>
                <span class="history-btn__text">{{ isBatchMode ? '完成' : '批量管理' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 批量操作栏 -->
      <div v-if="isBatchMode" class="batch-bar">
        <div class="history-container batch-bar__inner">
          <span class="batch-info">已选择 {{ selectedIds.length }} 项</span>
          <el-button type="danger" size="small" @click="batchDelete">
            删除选中
          </el-button>
        </div>
      </div>

      <!-- 历史记录内容 -->
      <div v-loading="loading" class="main-content">
          <template v-if="Object.keys(groupedHistory).length > 0">
            <div
              v-for="(groupItems, dateKey) in groupedHistory"
              :key="dateKey"
              class="history-group"
            >
              <div class="group-anchor">
                <div class="group-anchor__track"></div>
              </div>
              <div class="group-content">
                <div class="group-header">
                  <h3 class="group-title">{{ dateKey }}</h3>
                  <el-checkbox
                    v-if="isBatchMode"
                    :model-value="groupItems.every(item => selectedIds.includes(item.id))"
                    @change="(val) => selectAllInGroup(groupItems, val)"
                  >
                    全选
                  </el-checkbox>
                </div>

                <div :class="['video-list', viewMode]">
                  <div
                    v-for="item in groupItems"
                    :key="item.id"
                    :class="['video-item', { selected: selectedIds.includes(item.id) }]"
                    @click="handleVideoClick(item)"
                  >
                    <div v-if="isBatchMode" class="select-box" @click.stop>
                      <el-checkbox
                        :model-value="selectedIds.includes(item.id)"
                        @change="toggleSelect(item.id)"
                      />
                    </div>

                    <div class="video-cover">
                      <img loading="lazy" decoding="async" :src="item.cover" :alt="item.title" />

                      <div v-if="!item.isFinished && item.progress > 0" class="progress-bar">
                        <div
                          class="progress-fill"
                          :style="{ width: `${(item.progress / item.totalDuration) * 100}%` }"
                        ></div>
                      </div>

                      <div class="video-duration">
                        {{ formatDuration(item.totalDuration) }}
                      </div>

                      <div v-if="item.isCollected" class="collected-badge">已收藏</div>

                      <div
                        v-if="!isBatchMode"
                        class="delete-btn"
                        @click.stop="deleteItem(item)"
                      >
                        <el-icon><Delete /></el-icon>
                      </div>
                    </div>

                    <div class="video-info">
                      <h4 class="video-title" :title="item.title">{{ item.title }}</h4>
                      <div class="video-meta">
                        <span class="author-name">{{ item.author }}</span>
                        <span class="watch-time">{{ formatTime(new Date(item.watchedAt)) }}</span>
                      </div>
                      <div v-if="!item.isFinished" class="watch-progress">
                        {{ formatProgress(item.progress, item.totalDuration, item.watchPercentage) }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>

          <!-- 空状态 -->
          <div v-else-if="!loading" class="empty-state">
            <el-icon :size="64"><Clock /></el-icon>
            <p>暂无历史记录</p>
            <p class="empty-subtitle">开启记录浏览历史，不错过任何精彩视频</p>
          </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* ===== 页面骨架（对齐 bilibili.com/account/history：白底通栏 + 1432px 居中容器） ===== */
.history-page {
  min-height: 100vh;
  background: #fff;
  color: #18191c;
}

.history-record {
  display: block;
  min-height: 100vh;
  padding-top: 30px;
  background: #fff;
}

.history-container {
  width: 1432px;
  min-width: 912px;
  margin: 0 auto;
  padding-left: 72px;
  box-sizing: border-box;
}

/* ===== 头部：标题 28px/600/40px 行高 + 右侧开关与视图切换 ===== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  min-width: 912px;
  margin-bottom: 10px;
}

.page-title {
  display: flex;
  align-items: center;
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  line-height: 40px;
  color: #18191c;
}

.page-title__icon {
  margin-right: 16px;
  font-size: 32px;
  color: #18191c;
}

.header-actions {
  display: flex;
  align-items: center;
  height: 29px;
}

.header-actions > * {
  margin-right: 0;
}

/* 记录浏览历史开关 */
.record-toggle {
  display: flex;
  align-items: center;
  height: 20px;
  margin-right: 16px;
}

.record-toggle__label {
  margin-left: 8px;
  font-size: 14px;
  line-height: 14px;
  color: #61666d;
}

.record-switch :deep(.el-switch__core) {
  width: 32px;
  height: 20px;
  border-radius: 18px;
}

.record-switch :deep(.el-switch__core .el-switch__action) {
  width: 14px;
  height: 14px;
}

/* 视图切换（单图标，点击在网格/列表间切换） */
.view-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 16px;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #9499a0;
  font-size: 28px;
  line-height: 28px;
  cursor: pointer;
  transition: color 0.2s;
}

.view-toggle:hover {
  color: #00aede;
}

/* ===== 筛选/搜索粘性条 ===== */
.page-breadcrumbs {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: #fff;
}

.breadcrumbs-inner {
  box-sizing: border-box;
  padding-top: 14px;
  padding-bottom: 14px;
}

.breadcrumbs-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 46px;
}

.breadcrumbs-left {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  margin-right: 12px;
}

.breadcrumbs-right {
  display: flex;
  align-items: center;
  flex: 0 1 auto;
}

/* 标签页（对齐 B 站 radio-tabs：无背景高亮、底部游标） */
.radio-tabs {
  position: relative;
  display: flex;
  column-gap: 28px;
}

.radio-tabs__item {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  box-sizing: border-box;
  padding: 12px 10px;
  color: #61666d;
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  white-space: normal;
  cursor: pointer;
  transition: color 0.3s;
}

.radio-tabs__item:hover {
  color: #00aede;
}

.radio-tabs__item--active {
  color: #00aede;
}

.radio-tabs__content {
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.radio-tabs__cursor {
  position: absolute;
  bottom: 0;
  left: 10px;
  width: 28px;
  height: 4px;
  border-radius: 2px;
  background: #00aede;
  transition: transform 0.3s;
}

/* 筛选按钮 */
.history-btn {
  display: flex;
  align-items: center;
  justify-content: normal;
  flex: 0 0 auto;
  height: 34px;
  min-width: 100px;
  box-sizing: border-box;
  padding: 0;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  background: #fff;
  color: #18191c;
  font-size: 14px;
  font-weight: 400;
  line-height: 14px;
  white-space: nowrap;
  cursor: pointer;
  transition: color 0.2s, background-color 0.2s, border 0.2s;
}

.history-btn:hover {
  color: #00aede;
  border-color: #00aede;
}

.history-btn__icon {
  margin-right: 4px;
  font-size: 16px;
}

.history-btn__text {
  font-size: 14px;
  line-height: 14px;
}

.clear-btn {
  margin-left: 24px;
}

.clear-btn .history-btn__icon {
  color: #c9ccd0;
}

.batch-btn {
  margin-left: 8px;
}

.batch-btn .history-btn__icon {
  color: #c9ccd0;
}

.batch-btn--active {
  color: #00aede;
  border-color: #00aede;
}

/* 搜索框（对齐 B 站 folded search-bar：200x34 灰底圆角 8px） */
.search-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  box-sizing: border-box;
  width: 200px;
  height: 34px;
  padding: 2px 0 2px 2px;
  margin-right: 16px;
  border: 1px solid #f1f2f3;
  border-radius: 8px;
  background: #f1f2f3;
  transition: 0.3s;
}

.search-input-wrap {
  flex: 1;
  overflow: hidden;
}

.search-input {
  display: inline-block;
  box-sizing: border-box;
  width: 164px;
  height: 28px;
  padding: 0 30px 0 10px;
  border: 0;
  outline: none;
  background: transparent;
  color: #18191c;
  font-size: 14px;
  line-height: normal;
  border-radius: 8px;
  transition: background-color 0.3s;
}

.search-input::placeholder {
  color: #9499a0;
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 32px;
  height: 32px;
  color: #18191c;
  font-size: 16px;
  cursor: pointer;
}

/* ===== 批量操作栏 ===== */
.batch-bar {
  margin-top: 12px;
}

.batch-bar__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-radius: 8px;
  background: #f1f2f3;
}

.batch-info {
  font-size: 14px;
  color: #61666d;
}

/* ===== 内容区：时间轴（左侧 32px 锚点轨道 + 右侧 1360px 内容） ===== */
.main-content {
  width: 1432px;
  min-width: 912px;
  box-sizing: border-box;
  margin: 20px auto 0;
}

.history-group {
  position: relative;
  display: flex;
  align-items: normal;
  margin-bottom: 0;
}

.group-anchor {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 0 1 auto;
  width: 32px;
  margin-right: 40px;
}

.group-anchor__track {
  flex: 1 1 0%;
  width: 2px;
  background: #c9ccd0;
}

.group-content {
  flex: 1;
  min-width: 0;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.group-title {
  margin: 0;
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  line-height: 22px;
  color: #18191c;
}

/* 视频列表 - 网格模式（对齐 B 站 grid-mode：1360px / 5 列 / gap 16px 48px） */
.video-list.grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 48px 16px;
  padding-bottom: 100px;
}

/* 视频列表 - 列表模式 */
.video-list.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 40px;
}

.video-list.list .video-item {
  display: flex;
  gap: 16px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.video-list.list .video-cover {
  width: 200px;
  padding-top: 112px;
  flex-shrink: 0;
}

.video-list.list .video-info {
  flex: 1;
  padding: 0;
}

/* 视频卡片 */
.video-item {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: 0;
  overflow: visible;
}

.video-item:hover .video-cover img {
  transform: scale(1.05);
}

.video-item.selected {
  box-shadow: 0 0 0 2px #00aede;
}

/* 选择框 */
.select-box {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 10;
  padding: 4px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 4px;
}

/* 视频封面 */
.video-cover {
  position: relative;
  box-sizing: border-box;
  z-index: 1;
  width: 100%;
  padding-top: 146px;
  background: #c9ccd0;
  border-radius: 6px;
  overflow: hidden;
}

.video-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

/* 进度条 */
.progress-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: rgba(255, 255, 255, 0.3);
}

.progress-fill {
  height: 100%;
  background: #00aede;
  transition: width 0.3s;
}

/* 视频时长 */
.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  padding: 2px 6px;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 2px;
  color: #fff;
  font-size: 12px;
}

/* 收藏标记 */
.collected-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 6px;
  background: #00aede;
  border-radius: 2px;
  color: #fff;
  font-size: 11px;
}

/* 删除按钮 */
.delete-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 4px;
  color: #fff;
  opacity: 0;
  cursor: pointer;
  transition: all 0.3s;
}

.video-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background: rgba(255, 77, 79, 0.8);
}

/* 视频信息 */
.video-info {
  padding: 8px 0 0;
}

.video-title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  height: 2.6em;
  margin: 0 0 6px;
  overflow: hidden;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.4;
  color: #18191c;
  text-overflow: ellipsis;
}

.video-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: #9499a0;
}

.author-name {
  max-width: 100px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.watch-time {
  color: #c9ccd0;
}

.watch-progress {
  margin-top: 4px;
  font-size: 12px;
  color: #00aede;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
  color: #c9ccd0;
}

.empty-state p {
  margin: 16px 0 0;
  font-size: 16px;
  color: #61666d;
}

.empty-subtitle {
  margin-top: 8px !important;
  font-size: 14px !important;
  color: #9499a0 !important;
}

/* 响应式：窄视口下退化为自适应列数 */
@media (max-width: 1431px) {
  .history-container {
    width: auto;
    margin: 0 24px;
    padding-left: 72px;
  }
}

@media (max-width: 1200px) {
  .video-list.grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 992px) {
  .video-list.grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .search-bar {
    width: 160px;
  }
}

@media (max-width: 768px) {
  .video-list.grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .page-title {
    font-size: 22px;
    line-height: 32px;
  }

  .page-header {
    height: 32px;
  }
}
</style>
