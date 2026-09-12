<template>
  <div class="drafts-box">
    <div class="box-header">
      <div class="box-title">
        <h2>草稿箱</h2>
        <p class="box-subtitle">草稿保存在本地浏览器，最多保留 20 条</p>
      </div>
      <div class="header-actions">
        <el-button
          v-if="drafts.length > 0"
          type="danger"
          plain
          size="small"
          :loading="clearing"
          @click="onClearAll"
        >
          清空全部
        </el-button>
      </div>
    </div>

    <el-empty
      v-if="drafts.length === 0"
      description="草稿箱空空如也，投稿时点击「存草稿」可暂存稿件"
    >
      <el-button type="primary" @click="goUpload">去投稿</el-button>
    </el-empty>

    <el-table
      v-else
      :data="drafts"
      stripe
      style="width: 100%"
      v-loading="loading"
    >
      <el-table-column label="封面" width="120">
        <template #default="{ row }">
          <div class="draft-cover">
            <img v-if="row.coverPreview" :src="row.coverPreview" alt="封面" />
            <span v-else class="cover-fallback">无封面</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="200">
        <template #default="{ row }">
          <span class="draft-title">{{ row.title || '未命名稿件' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="分P" width="90">
        <template #default="{ row }">
          {{ row.videoParts?.length || 0 }} P
        </template>
      </el-table-column>
      <el-table-column label="视频文件" width="120">
        <template #default="{ row }">
          <el-tag :type="row.hasLocalVideoFiles ? 'success' : 'warning'" size="small" effect="light">
            {{ row.hasLocalVideoFiles ? '已含文件' : '待重选' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="170">
        <template #default="{ row }">
          {{ formatTime(row.updatedAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="continueEdit(row)">
            继续编辑
          </el-button>
          <el-button type="danger" link size="small" @click="removeDraft(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDrafts, deleteDraft, clearAllDrafts } from '@/utils/drafts'

const router = useRouter()
const drafts = ref([])
const loading = ref(false)
const clearing = ref(false)

const reload = () => {
  drafts.value = listDrafts()
}

const formatTime = (ts) => {
  const d = new Date(ts)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

const goUpload = () => {
  router.push('/create-center/upload')
}

const continueEdit = (row) => {
  router.push(`/create-center/upload?draftId=${row.id}`)
}

const removeDraft = (row) => {
  ElMessageBox.confirm(`确定要删除草稿「${row.title || '未命名稿件'}」吗？`, '提示', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(() => {
      if (deleteDraft(row.id)) {
        reload()
        ElMessage.success('草稿已删除')
      }
    })
    .catch(() => {})
}

const onClearAll = () => {
  ElMessageBox.confirm('确定要清空全部草稿吗？此操作不可恢复', '提示', {
    confirmButtonText: '清空',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(() => {
      clearing.value = true
      setTimeout(() => {
        clearAllDrafts()
        reload()
        clearing.value = false
        ElMessage.success('已清空全部草稿')
      }, 200)
    })
    .catch(() => {})
}

onMounted(reload)
</script>

<style scoped>
.drafts-box {
  padding: 8px;
}

.box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.box-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.box-subtitle {
  margin: 4px 0 0;
  font-size: 12px;
  color: #909399;
}

.draft-cover {
  width: 96px;
  height: 60px;
  border-radius: 4px;
  overflow: hidden;
  background: #f4f4f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.draft-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  font-size: 12px;
  color: #909399;
}

.draft-title {
  font-weight: 500;
}
</style>