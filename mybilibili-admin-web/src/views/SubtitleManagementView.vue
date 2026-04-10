<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getVideosWithSubtitleInfo,
  getVideoSubtitles,
  uploadSubtitle,
  importSrtToMongo,
  setDefaultSubtitle,
  deleteSubtitle,
  previewSubtitle,
  scanSystemSubtitles,
  importSystemSubtitle
} from '../api/subtitle'

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
  srtFilePath: '',
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

// 待入库字幕
const pendingImportSubtitles = ref([])

const loadVideos = async () => {
  loading.value = true
  try {
    const res = await getVideosWithSubtitleInfo()
    if (res.code === 200 || res.success) {
      // 处理分页响应格式 { list: [], total: number, page: number, size: number }
      let list = res.data?.list || res.data || []
      if (keyword.value) {
        list = list.filter(item =>
          item.title?.includes(keyword.value) ||
          item.id?.toString().includes(keyword.value)
        )
      }
      tableData.value = list
    } else {
      ElMessage.error(res.message || '获取视频列表失败')
    }
  } catch (error) {
    ElMessage.error('获取视频列表失败: ' + (error.message || '未知错误'))
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

// 查看字幕详情
const handleViewSubtitles = async (row) => {
  currentVideo.value = row
  subtitleDialogVisible.value = true
  await loadVideoSubtitles(row.id)
  await loadPendingImportSubtitles(row.id)
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
const handleImportSystemSubtitle = async (language) => {
  try {
    await ElMessageBox.confirm(
      `确定将 ${language} 字幕入库到数据库吗？`,
      '字幕入库确认',
      { type: 'warning' }
    )
    
    importLoading.value = true
    const res = await importSystemSubtitle(currentVideo.value.id, language)
    
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

const loadVideoSubtitles = async (videoId) => {
  try {
    const res = await getVideoSubtitles(videoId)
    if (res.code === 200 || res.success) {
      videoSubtitles.value = res.data || []
    } else {
      ElMessage.error(res.message || '获取字幕列表失败')
    }
  } catch (error) {
    ElMessage.error('获取字幕列表失败: ' + (error.message || '未知错误'))
  }
}

// 打开上传弹窗
const handleOpenUpload = (row) => {
  uploadForm.value = {
    videoId: row.id,
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
const handleOpenImport = (row) => {
  importForm.value = {
    videoId: row.id,
    srtFilePath: '',
    language: 'zh-CN',
    isDefault: false
  }
  importDialogVisible.value = true
}

// 确认导入SRT
const handleImportSubmit = async () => {
  if (!importForm.value.srtFilePath) {
    ElMessage.error('请输入SRT文件路径')
    return
  }

  importLoading.value = true
  try {
    const res = await importSrtToMongo(
      importForm.value.videoId,
      importForm.value.srtFilePath,
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
    const res = await setDefaultSubtitle(subtitle.id)
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
      previewData.value = res.data || { subtitle: {}, content: [] }
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

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
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
        placeholder="搜索视频标题、ID"
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

    <!-- 视频列表 -->
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="视频标题" min-width="200" show-overflow-tooltip />
      <el-table-column prop="durationSeconds" label="时长" width="100">
        <template #default="{ row }">
          <span>{{ formatDuration(row.durationSeconds) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="字幕统计" width="200">
        <template #default="{ row }">
          <div class="subtitle-stats">
            <el-tag size="small" type="success">通过: {{ row.approvedCount || 0 }}</el-tag>
            <el-tag size="small" type="warning">待审: {{ row.pendingCount || 0 }}</el-tag>
            <el-tag size="small" type="info">系统: {{ row.systemCount || 0 }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="默认字幕" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.hasDefaultSubtitle" type="success" size="small">有</el-tag>
          <el-tag v-else type="info" size="small">无</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="待入库字幕" width="120">
        <template #default="{ row }">
          <div v-if="row.pendingImportCount > 0" class="pending-import-tags">
            <el-tag 
              v-for="sub in row.pendingImportSubtitles?.filter(s => s.status === 'pending')" 
              :key="sub.language"
              type="warning" 
              size="small"
              class="pending-tag"
            >
              {{ getLanguageDisplayName(sub.language) }}
            </el-tag>
          </div>
          <span v-else class="no-pending">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="280">
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            size="small"
            @click="handleViewSubtitles(row)"
          >
            <el-icon><View /></el-icon>
            查看字幕
          </el-button>
          <el-button
            link
            type="success"
            size="small"
            @click="handleOpenUpload(row)"
          >
            <el-icon><Upload /></el-icon>
            上传字幕
          </el-button>
          <el-button
            link
            type="warning"
            size="small"
            @click="handleOpenImport(row)"
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
      <div v-if="pendingImportSubtitles.filter(s => s.status === 'pending').length > 0" class="pending-import-section">
        <div class="section-title">
          <el-icon><Warning /></el-icon>
          <span>待入库字幕（系统生成）</span>
        </div>
        <div class="pending-import-list">
          <div 
            v-for="sub in pendingImportSubtitles.filter(s => s.status === 'pending')" 
            :key="sub.language"
            class="pending-import-item"
          >
            <div class="pending-info">
              <span class="language">{{ getLanguageDisplayName(sub.language) }}</span>
              <span class="file-name">{{ sub.fileName }}</span>
              <span class="file-size">({{ (sub.fileSize / 1024).toFixed(1) }} KB)</span>
            </div>
            <el-button 
              type="primary" 
              size="small"
              :loading="importLoading"
              @click="handleImportSystemSubtitle(sub.language)"
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
              link
              type="primary"
              size="small"
              @click="handlePreview(row)"
            >
              预览
            </el-button>
            <el-button
              v-if="!row.isDefault && row.status === 1"
              link
              type="success"
              size="small"
              @click="handleSetDefault(row)"
            >
              设为默认
            </el-button>
            <el-button
              link
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
        <el-form-item label="SRT文件路径">
          <el-input
            v-model="importForm.srtFilePath"
            placeholder="例如: /uploads/subtitles/video_123.srt"
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
