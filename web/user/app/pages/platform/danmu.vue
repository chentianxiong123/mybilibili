<template>
  <div class="danmu-manager">
    <!-- 页面级 tab：稿件弹幕 / 弹幕设置 / 弹幕反馈（对齐真页 new-link-top） -->
    <div class="top-tabs">
      <span
        class="top-tab"
        :class="{ active: mainTab === 'list' }"
        @click="switchTab('list')"
      >
        稿件弹幕
      </span>
      <span
        class="top-tab"
        :class="{ active: mainTab === 'setting' }"
        @click="switchTab('setting')"
      >
        弹幕设置
      </span>
      <span
        class="top-tab"
        :class="{ active: mainTab === 'feedback' }"
        @click="switchTab('feedback')"
      >
        弹幕反馈
      </span>
    </div>

    <!-- 稿件弹幕 -->
    <div v-show="mainTab === 'list'" class="danmu-content">
      <!-- 筛选表单 -->
      <div class="danmu-form">
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">弹幕内容</span>
            <el-input v-model="filters.content" placeholder="请输入弹幕内容" clearable class="form-input" />
          </div>
          <div class="form-item">
            <span class="form-label">播放时间</span>
            <el-input v-model.number="filters.timeStart" placeholder="例如: 22" class="form-input time" />
            <span class="zhi">至</span>
            <el-input v-model.number="filters.timeEnd" placeholder="例如: 33" class="form-input time" />
          </div>
        </div>

        <div class="form-row">
          <div class="form-item">
            <span class="form-label">发送者/UID</span>
            <el-input v-model="filters.sender" placeholder="请输入昵称或UID，多个以,号隔开" clearable class="form-input sender" />
          </div>
          <div class="form-item">
            <span class="form-label">发送日期</span>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              class="form-date"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-item">
            <span class="form-label">弹幕模式</span>
            <el-checkbox-group v-model="filters.modes" class="mode-group">
              <el-checkbox label="滚动" value="0" />
              <el-checkbox label="底端" value="4" />
              <el-checkbox label="顶端" value="5" />
              <el-checkbox label="逆向" value="6" />
              <el-checkbox label="高级" value="7" />
            </el-checkbox-group>
          </div>
        </div>

        <div class="form-row">
          <div class="form-item">
            <span class="form-label">子弹幕池</span>
            <el-checkbox-group v-model="filters.pools" class="mode-group">
              <el-checkbox label="普通" value="0" />
              <el-checkbox label="字幕" value="1" />
              <el-checkbox label="特殊" value="2" />
            </el-checkbox-group>
          </div>
          <div class="form-btns">
            <button class="reset-btn" @click="handleReset">重置</button>
            <button class="search-btn" @click="handleSearch">搜索</button>
          </div>
        </div>
      </div>

      <!-- 弹幕列表：表格式 -->
      <div class="danmu-table-wrap" v-loading="loading">
        <table class="danmu-table">
          <thead>
            <tr>
              <th class="col-sender">发送者</th>
              <th class="col-time">播放时间</th>
              <th class="col-content">弹幕内容</th>
              <th class="col-date">发送时间</th>
              <th class="col-op">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in list" :key="item.ID || item.id" class="danmu-row">
              <td class="col-sender">
                <span class="sender-name">{{ item.user_name || item.username || '匿名用户' }}</span>
              </td>
              <td class="col-time">{{ formatPlayTime(item.Time ?? item.time) }}</td>
              <td class="col-content">
                <span class="danmu-content">{{ item.Content || item.content }}</span>
                <div class="video-title" @click="goVideo(item.VideoID ?? item.video_id)">
                  <span class="video-title-text">{{ item.manuscript_title || ('视频 #' + (item.VideoID ?? item.video_id)) }}</span>
                </div>
              </td>
              <td class="col-date">{{ formatTime(item.CreatedAt || item.created_at) }}</td>
              <td class="col-op">
                <button class="op-btn" @click="handleDelete(item)">删除</button>
              </td>
            </tr>
            <tr v-if="!loading && !list.length">
              <td colspan="5" class="empty-cell">
                <el-empty description="暂无弹幕" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <BccPagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 弹幕设置 -->
    <div v-show="mainTab === 'setting'" class="placeholder-tab">
      <el-empty description="暂无弹幕设置" />
    </div>

    <!-- 弹幕反馈 -->
    <div v-show="mainTab === 'feedback'" class="placeholder-tab">
      <el-empty description="暂无弹幕反馈" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi } from '@/api/creator'
import BccPagination from '@/components/createCenter/BccPagination.vue'

interface DanmuItem {
  ID?: number
  id?: number
  VideoID?: number
  video_id?: number
  UserID?: number
  user_id?: number
  Content?: string
  content?: string
  Time?: number
  time?: number
  CreatedAt?: string
  created_at?: string
  user_name?: string
  manuscript_title?: string
  manuscript_cover?: string
}

const mainTab = ref('list')
const list = ref<DanmuItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const filters = ref({
  content: '',
  timeStart: undefined as number | undefined,
  timeEnd: undefined as number | undefined,
  sender: '',
  modes: [] as string[],
  pools: [] as string[]
})
const dateRange = ref<[string, string] | null>(null)

const switchTab = (tab: string) => {
  mainTab.value = tab
}

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getDanmakuList({
      page: page.value,
      size: pageSize.value,
    })
    if (res.code === 200) {
      list.value = res.data?.list || []
      total.value = res.data?.total || 0
    } else {
      ElMessage.error(res.message || '获取弹幕失败')
    }
  } catch (e) {
    ElMessage.error('获取弹幕失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handleReset = () => {
  filters.value = { content: '', timeStart: undefined, timeEnd: undefined, sender: '', modes: [], pools: [] }
  dateRange.value = null
  page.value = 1
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const handleDelete = (item: DanmuItem) => {
  const id = item.ID ?? item.id
  const content = item.Content || item.content
  ElMessageBox.confirm(`确定删除这条弹幕吗？\n「${content}」`, '删除弹幕', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const res: any = await creatorApi.deleteDanmaku(id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const formatPlayTime = (sec: number) => {
  if (!sec && sec !== 0) return '00:00'
  const s = Math.floor(sec)
  const mm = String(Math.floor(s / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  const hh = Math.floor(s / 3600)
  if (hh > 0) {
    return `${String(hh).padStart(2, '0')}:${mm}:${ss}`
  }
  return `${mm}:${ss}`
}

const formatTime = (t: string) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const goVideo = (vid: number) => {
  if (!vid) return
  window.open(`/video/${vid}`, '_blank')
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.danmu-manager {
  padding: 0 40px;
}

/* 页面级 tab：对齐真页 new-link-top-item（16px/选中700粉色+底边线） */
.top-tabs {
  display: flex;
  align-items: center;
  height: 64px;
}

.top-tab {
  position: relative;
  font-size: 16px;
  color: #505050;
  cursor: pointer;
  margin-right: 40px;
  line-height: 64px;
}

.top-tab.active {
  color: var(--brand_pink, #ff6699);
  font-weight: 700;
}

.top-tab.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2.5px;
  background: var(--brand_pink, #ff6699);
}

/* 筛选表单 */
.danmu-form {
  padding: 20px 0;
  font-size: 14px;
}

.form-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 24px 40px;
  margin-bottom: 16px;
}

.form-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-label {
  flex-shrink: 0;
  min-width: 70px;
  color: #505050;
}

.form-input {
  width: 220px;
}

.form-input.time {
  width: 90px;
}

.form-input.sender {
  width: 260px;
}

.form-date {
  width: 280px;
}

.zhi {
  color: #9499a0;
}

.mode-group :deep(.el-checkbox) {
  margin-right: 20px;
}

.form-btns {
  margin-left: auto;
  display: flex;
  gap: 12px;
}

.reset-btn,
.search-btn {
  height: 32px;
  padding: 0 16px;
  font-size: 14px;
  border-radius: 4px;
  cursor: pointer;
}

.reset-btn {
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
}

.search-btn {
  color: #fff;
  background: var(--brand_pink, #ff6699);
  border: 1px solid var(--brand_pink, #ff6699);
}

/* 弹幕表格 */
.danmu-table-wrap {
  margin-top: 8px;
}

.danmu-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
}

.danmu-table thead th {
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 400;
  color: #757575;
  text-align: left;
  border-bottom: 1px solid #e7e7e7;
  background: #fff;
}

.danmu-table tbody td {
  padding: 14px 16px;
  font-size: 14px;
  color: #505050;
  border-bottom: 1px solid #f0f1f4;
}

.danmu-row:hover {
  background: #fafbfc;
}

.danmu-row td:last-child {
  border-bottom: 1px solid #f0f1f4;
}

.col-sender { width: 16%; }
.col-time { width: 12%; text-align: center; }
.col-content { width: 40%; }
.col-date { width: 18%; }
.col-op { width: 14%; }

.sender-name {
  color: #212121;
}

.col-time {
  color: #757575;
}

.danmu-content {
  display: block;
  color: #212121;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-title {
  margin-top: 4px;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-title-text {
  font-size: 12px;
  color: #99a2aa;
}

.video-title:hover .video-title-text {
  color: var(--brand_pink, #ff6699);
}

.op-btn {
  height: 28px;
  padding: 0 12px;
  font-size: 12px;
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 2px;
  cursor: pointer;
}

.op-btn:hover {
  color: #f56c6c;
  border-color: #f56c6c;
}

.empty-cell {
  text-align: center;
  padding: 40px 0 !important;
}

.placeholder-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
}
</style>