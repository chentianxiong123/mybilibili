<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, View, Upload, Document, Warning, Download } from '@element-plus/icons-vue'
import api from '@/api/client'
import {
  getVideosWithSubtitleInfo,
  getVideoSubtitles,
  uploadSubtitle,
  importSrtToMongo,
  setDefaultSubtitle,
  deleteSubtitle,
  previewSubtitle,
  scanSystemSubtitles,
  approveSubtitle
} from '@/api/subtitle'

const tableData = ref([])
const loading = ref(false)

const keyword = ref('')

// 字幕详情弹窗
const subtitleDialogVisible = ref(false)
const currentVideo = ref({})
const videoSubtitles = ref([])

// 上传字幕弹窗
const uploadDialogVisible = ref(false)
const uploadForm = ref({
  videoId: null,
  language: 'zh-CN',
  isDefault: false
})
const uploadFile = ref(null)
const uploadLoading = ref(false)

// 导入SRT弹窗
const importDialogVisible = ref(false)
const importForm = ref({
  videoId: null,
  srtContent: '',
  language: 'zh-CN',
  isDefault: false
})
const importLoading = ref(false)

// 预览弹窗
const previewDialogVisible = ref(false)
const previewData = ref({
  subtitle: {},
  content: []
})

// video titles cache
const videoTitles = ref({})

const pendingImportSubtitles = ref([])

const normalizeSubtitleRecord = (d) => ({
  id: d.id,
  videoId: d.video_id ?? d.videoId ?? null,
  language: d.language || '',
  languageName: d.language_name || d.languageName || '',
  status: d.status,
  source: d.source || '',
  createdAt: d.created_at || d.createdAt || ''
})

const loadVideos = async () => {
  loading.value = true
  try {
    const res = await getVideosWithSubtitleInfo()
    if (res.code === 200 || res.success) {
      let list = res.data?.list || res.data || []
      if (keyword.value) {
        list = list.filter(item =>
          (item.language || '').includes(keyword.value) ||
          (item.language_name || '').includes(keyword.value) ||
          (item.video_id || item.videoId || '').toString().includes(keyword.value) ||
          (item.id || '').toString().includes(keyword.value)
        )
      }
      let records = (Array.isArray(list) ? list : []).map(normalizeSubtitleRecord)
      // 去重：主列表按视频展示
      const seen = new Set()
      records = records.filter(r => {
        const key = r.videoId || r.id
        if (seen.has(key)) return false
        seen.add(key)
        return true
      })
      tableData.value = records
      // 后台 enrich 标题
      enrichVideoTitles(records.map(r => r.videoId))
    } else {
      ElMessage.error(res.message || '获取字幕列表失败')
    }
  } catch (error) {
    ElMessage.error('获取字幕列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadVideos()
}

const handleReset = () => {
  keyword.value = ''
  loadVideos()
}

const enrichVideoTitles = async (videoIds) => {
  const unique = [...new Set(videoIds)].filter(Boolean).slice(0, 30)
  for (const vid of unique) {
    if (videoTitles.value[vid]) continue
    try {
      const res = await api.get(`/manuscript/admin/video-source/${vid}`)
      const d = (res.data && (res.data.data || res.data)) || {}
      const t = d.title || d.videoTitle || `视频 #${vid}`
      videoTitles.value[vid] = t
    } catch (e) {}
  }
  // trigger reactivity
  tableData.value = tableData.value.map(x => ({...x}))
}

// 查看字幕详情
const handleViewSubtitles = async (videoId) => {
  currentVideo.value = { id: videoId, title: `视频 #${videoId}` }
  subtitleDialogVisible.value = true
  // 尝试获取真实标题
  try {
    const res = await api.get(`/manuscript/admin/video-source/${videoId}`)
    if (res.code === 200 && res.data) {
      const t = res.data.title || res.data.videoTitle
      if (t) currentVideo.value.title = t
    }
  } catch (e) {
    // ignore, keep default
  }
  await loadVideoSubtitles(videoId)
  await loadPendingImportSubtitles(videoId)
}

// 加载待入库字幕
const loadPendingImportSubtitles = async (videoId) => {
  try {
    const res = await scanSystemSubtitles(videoId)
    if (res.code === 200 || res.success) {
      pendingImportSubtitles.value = res.data || []
    } else {
      pendingImportSubtitles.value = []
    }
  } catch (error) {
    console.error('扫描系统字幕失败:', error)
    pendingImportSubtitles.value = []
  }
}

// 处理字幕入库
const handleImportSystemSubtitle = async (sub) => {
  try {
    await ElMessageBox.confirm(
      `确定将 ${sub.language_name || sub.language} 字幕入库到数据库吗？`,
      '字幕入库确认',
      { type: 'warning' }
    )
    
    importLoading.value = true
    const res = await approveSubtitle(sub.id)
    
    if (res.code === 200 || res.success) {
      ElMessage.success('字幕入库成功')
      // 刷新列表
      await loadVideoSubtitles(currentVideo.value.id)
      await loadPendingImportSubtitles(currentVideo.value.id)
      loadVideos()
    } else {
      ElMessage.error(res.message || '入库失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('字幕入库异常:', error)
      ElMessage.error('入库失败: ' + (error.message || '未知错误'))
    }
  } finally {
    importLoading.value = false
  }
}

// 获取语言显示名称
const getLanguageDisplayName = (language) => {
  const languageMap = {
    'zh-CN': '简体中文',
    'zh-TW': '繁体中文',
    'en': 'English',
    'ja': '日本語',
    'ko': '한국어'
  }
  return languageMap[language] || language
}

// 字幕 API 返回 snake_case (upload_time/uploaded_by)，归一化
const normalizeSubtitle = (d) => ({
  id: d.id,
  videoId: d.video_id ?? d.videoId,
  language: d.language,
  languageName: d.language_name || d.languageName || '',
  format: d.format || '',
  isDefault: d.is_default ?? d.isDefault ?? false,
  uploadId: d.uploaded_by ?? d.uploadId ?? d.uploadedBy ?? null,
  status: d.status,
  source: d.source || '',
  createTime: d.upload_time || d.createTime || d.created_at || d.uploadTime || ''
})

const loadVideoSubtitles = async (videoId) => {
  try {
    const res = await getVideoSubtitles(videoId)
    if (res.code === 200 || res.success) {
      videoSubtitles.value = (res.data || []).map(normalizeSubtitle)
    } else {
      ElMessage.error(res.message || '获取字幕列表失败')
    }
  } catch (error) {
    ElMessage.error('获取字幕列表失败: ' + (error.message || '未知错误'))
  }
}

// 打开上传弹窗
const handleOpenUpload = (videoId) => {
  uploadForm.value = {
    videoId: videoId,
    language: 'zh-CN',
    isDefault: false
  }
  uploadFile.value = null
  uploadDialogVisible.value = true
}

// 文件选择
const handleFileChange = (file) => {
  const isSrt = file.name.endsWith('.srt')
  if (!isSrt) {
    ElMessage.error('请上传 SRT 格式的字幕文件')
    return false
  }
  uploadFile.value = file.raw
  return false
}

// 确认上传
const handleUploadSubmit = async () => {
  if (!uploadFile.value) {
    ElMessage.error('请选择字幕文件')
    return
  }

  uploadLoading.value = true
  try {
    const res = await uploadSubtitle(
      uploadForm.value.videoId,
      uploadFile.value,
      uploadForm.value.language,
      uploadForm.value.isDefault
    )
    if (res.code === 200 || res.success) {
      ElMessage.success('字幕上传成功')
      uploadDialogVisible.value = false
      loadVideos()
      if (subtitleDialogVisible.value) {
        loadVideoSubtitles(uploadForm.value.videoId)
      }
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch (error) {
    ElMessage.error('上传失败: ' + (error.message || '未知错误'))
  } finally {
    uploadLoading.value = false
  }
}

// 打开导入SRT弹窗
const handleOpenImport = (videoId) => {
  importForm.value = {
    videoId: videoId,
    srtContent: '',
    language: 'zh-CN',
    isDefault: false
  }
  importDialogVisible.value = true
}

// 确认导入SRT
const handleImportSubmit = async () => {
  if (!importForm.value.srtContent || !importForm.value.srtContent.trim()) {
    ElMessage.error('请输入SRT内容')
    return
  }

  importLoading.value = true
  try {
    const res = await importSrtToMongo(
      importForm.value.videoId,
      importForm.value.srtContent,
      importForm.value.language,
      importForm.value.isDefault
    )
    if (res.code === 200 || res.success) {
      ElMessage.success('SRT导入成功')
      importDialogVisible.value = false
      loadVideos()
      if (subtitleDialogVisible.value) {
        loadVideoSubtitles(importForm.value.videoId)
      }
    } else {
      ElMessage.error(res.message || '导入失败')
    }
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.message || '未知错误'))
  } finally {
    importLoading.value = false
  }
}

// 设置默认字幕
const handleSetDefault = async (subtitle) => {
  try {
    await ElMessageBox.confirm('确定将该字幕设为默认吗？', '设置默认字幕', { type: 'warning' })
    const res = await setDefaultSubtitle(subtitle.id, subtitle.videoId)
    if (res.code === 200 || res.success) {
      ElMessage.success('设置成功')
      loadVideoSubtitles(subtitle.videoId)
      loadVideos()
    } else {
      ElMessage.error(res.message || '设置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('设置默认字幕异常:', error)
    }
  }
}

// 删除字幕
const handleDelete = async (subtitle) => {
  try {
    await ElMessageBox.confirm('确定删除该字幕吗？', '删除字幕', { type: 'warning' })
    const res = await deleteSubtitle(subtitle.id)
    if (res.code === 200 || res.success) {
      ElMessage.success('删除成功')
      loadVideoSubtitles(subtitle.videoId)
      loadVideos()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除字幕异常:', error)
    }
  }
}

// 预览字幕
const handlePreview = async (subtitle) => {
  try {
    const res = await previewSubtitle(subtitle.id)
    if (res.code === 200 || res.success) {
      const cues = Array.isArray(res.data) ? res.data : (res.data?.cues || [])
      previewData.value = { subtitle, content: cues }
      previewDialogVisible.value = true
    } else {
      ElMessage.error(res.message || '获取字幕内容失败')
    }
  } catch (error) {
    ElMessage.error('获取字幕内容失败: ' + (error.message || '未知错误'))
  }
}

// 状态显示
const getStatusText = (status) => {
  const statusMap = {
    0: '待审核',
    1: '审核通过',
    2: '审核拒绝',
    3: '系统生成'
  }
  return statusMap[status] || '未知'
}

const getStatusType = (status) => {
  const typeMap = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    3: 'info'
  }
  return typeMap[status] || ''
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 格式化字幕时长（秒 -> mm:ss 或 hh:mm:ss）
const formatDuration = (seconds) => {
  if (seconds == null || isNaN(seconds)) return '00:00'
  const secs = Math.floor(seconds)
  const hours = Math.floor(secs / 3600)
  const minutes = Math.floor((secs % 3600) / 60)
  const s = secs % 60
  if (hours > 0) {
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }
  return `${minutes.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

onMounted(() => {
  loadVideos()
})
</script>

<template>
  <div class="subtitle-management-page">
    <h2 class="page-title">字幕管理</h2>

    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索 视频ID / 语言 / 来源"
        clearable
        style="width: 250px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
      <el-button type="primary" @click="loadVideos">刷新</el-button>
    </div>

    <!-- 字幕列表 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="字幕ID(示例)" width="100" />
      <el-table-column prop="videoId" label="视频ID" width="80" />
      <el-table-column label="视频标题" min-width="180">
        <template #default="{ row }">
          {{ videoTitles[row.videoId] || `视频 #${row.videoId}` }}
        </template>
      </el-table-column>
      <el-table-column prop="language" label="语言代码" width="100" />
      <el-table-column prop="languageName" label="语言名称" width="120" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="来源" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.source === 'system'" type="info" size="small">系统</el-tag>
          <el-tag v-else type="primary" size="small">用户</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="170">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="280">
        <template #default="{ row }">
          <el-button
            type="primary"
            size="small"
            @click="handleViewSubtitles(row.videoId)"
          >
            <el-icon><View /></el-icon>
            视频字幕
          </el-button>
          <el-button
            type="success"
            size="small"
            @click="handleOpenUpload(row.videoId)"
          >
            <el-icon><Upload /></el-icon>
            上传
          </el-button>
          <el-button
            type="warning"
            size="small"
            @click="handleOpenImport(row.videoId)"
          >
            <el-icon><Document /></el-icon>
            导入SRT
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 字幕详情弹窗 -->
    <el-dialog
      v-model="subtitleDialogVisible"
      :title="`字幕列表 - ${currentVideo.title}`"
      width="900px"
    >
      <!-- 待入库字幕区域 -->
      <div v-if="pendingImportSubtitles.filter(s => s.status === 0 || s.status === '0' || ['system','whisper'].includes(s.source)).length > 0" class="pending-import-section">
        <div class="section-title">
          <el-icon><Warning /></el-icon>
          <span>待入库字幕（系统生成）</span>
        </div>
        <div class="pending-import-list">
          <div 
            v-for="sub in pendingImportSubtitles.filter(s => s.status === 0 || s.status === '0' || ['system','whisper'].includes(s.source))" 
            :key="sub.id || sub.language"
            class="pending-import-item"
          >
            <div class="pending-info">
              <span class="language">{{ getLanguageDisplayName(sub.language) }}</span>
              <span class="file-name">{{ sub.file_name || sub.fileName || sub.language_name || '' }}</span>
              <span class="file-size" v-if="sub.file_size || sub.fileSize">({{ ((sub.file_size || sub.fileSize) / 1024).toFixed(1) }} KB)</span>
            </div>
            <el-button 
              type="primary" 
              size="small"
              :loading="importLoading"
              @click="handleImportSystemSubtitle(sub)"
            >
              <el-icon><Download /></el-icon>
              入库
            </el-button>
          </div>
        </div>
      </div>

      <div class="section-title">已入库字幕</div>
      <el-table :data="videoSubtitles" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault" type="success" size="small">是</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="uploadId" label="上传者" width="100">
          <template #default="{ row }">
            <span>{{ row.uploadId ? '用户' + row.uploadId : '系统' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160">
          <template #default="{ row }">
            <span>{{ formatTime(row.createTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handlePreview(row)"
            >
              预览
            </el-button>
            <el-button
              v-if="!row.isDefault && row.status === 1"
              type="success"
              size="small"
              @click="handleSetDefault(row)"
            >
              设为默认
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 上传字幕弹窗 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传字幕"
      width="500px"
    >
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="视频ID">
          <el-input v-model="uploadForm.videoId" disabled />
        </el-form-item>
        <el-form-item label="语言">
          <el-select v-model="uploadForm.language" style="width: 100%">
            <el-option label="简体中文" value="zh-CN" />
            <el-option label="繁体中文" value="zh-TW" />
            <el-option label="English" value="en" />
            <el-option label="日本語" value="ja" />
            <el-option label="한국어" value="ko" />
          </el-select>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="uploadForm.isDefault" />
        </el-form-item>
        <el-form-item label="字幕文件">
          <el-upload
            accept=".srt"
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">请上传 SRT 格式的字幕文件</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="handleUploadSubmit">
          上传
        </el-button>
      </template>
    </el-dialog>

    <!-- 导入SRT弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入SRT到数据库"
      width="500px"
    >
      <el-form :model="importForm" label-width="100px">
        <el-form-item label="视频ID">
          <el-input v-model="importForm.videoId" disabled />
        </el-form-item>
        <el-form-item label="SRT内容">
          <el-input
            v-model="importForm.srtContent"
            type="textarea"
            :rows="8"
            placeholder="粘贴完整的 SRT 字幕内容..."
          />
        </el-form-item>
        <el-form-item label="语言">
          <el-select v-model="importForm.language" style="width: 100%">
            <el-option label="简体中文" value="zh-CN" />
            <el-option label="繁体中文" value="zh-TW" />
            <el-option label="English" value="en" />
            <el-option label="日本語" value="ja" />
            <el-option label="한국어" value="ko" />
          </el-select>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="importForm.isDefault" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImportSubmit">
          导入
        </el-button>
      </template>
    </el-dialog>

    <!-- 字幕预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="字幕预览"
      width="700px"
    >
      <div v-if="previewData.subtitle" class="subtitle-preview">
        <div class="preview-header">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="语言">{{ previewData.subtitle.language }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getStatusType(previewData.subtitle.status)" size="small">
                {{ getStatusText(previewData.subtitle.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="默认">
              {{ previewData.subtitle.isDefault ? '是' : '否' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <div class="preview-content">
          <div
            v-for="(item, index) in previewData.content"
            :key="index"
            class="subtitle-item"
          >
            <div class="subtitle-time">{{ formatDuration(item.startTime) }} - {{ formatDuration(item.endTime) }}</div>
            <div class="subtitle-text">{{ item.text }}</div>
          </div>
          <el-empty v-if="!previewData.content || previewData.content.length === 0" description="暂无字幕内容" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.subtitle-management-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
}

.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.subtitle-stats {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.subtitle-preview {
  max-height: 500px;
  overflow-y: auto;
}

.preview-header {
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.preview-content {
  padding: 10px;
}

.subtitle-item {
  padding: 10px;
  margin-bottom: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.subtitle-time {
  font-size: 12px;
  color: #909399;
  margin-bottom: 5px;
}

.subtitle-text {
  font-size: 14px;
  color: #303133;
  line-height: 1.5;
}

.pending-import-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.pending-tag {
  margin: 2px 0;
}

.no-pending {
  color: #909399;
}

.pending-import-section {
  background: #fdf6ec;
  border: 1px solid #f5dab1;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 20px;
}

.pending-import-section .section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e6a23c;
  font-weight: 500;
  margin-bottom: 12px;
}

.pending-import-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pending-import-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 12px;
  border-radius: 4px;
  border: 1px solid #ebeef5;
}

.pending-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pending-info .language {
  font-weight: 500;
  color: #303133;
}

.pending-info .file-name {
  color: #606266;
  font-size: 13px;
}

.pending-info .file-size {
  color: #909399;
  font-size: 12px;
}

.section-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}
</style>
