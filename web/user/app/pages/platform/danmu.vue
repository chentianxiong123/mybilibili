<template>
  <div class="danmu-manager">
    <!-- 主要选择栏 -->
    <div class="content-tabs">
      <el-tabs v-model="mainTab" type="card" @tab-change="loadData">
        <el-tab-pane label="全部弹幕" name="all"></el-tab-pane>
        <el-tab-pane label="按视频筛选" name="byVideo"></el-tab-pane>
      </el-tabs>
    </div>

    <!-- 次级筛选栏 -->
    <div class="danmu-filters" v-show="mainTab === 'byVideo'">
      <div class="filter-row">
        <el-input
          v-model="videoId"
          placeholder="输入视频稿ID筛选弹幕"
          size="small"
          class="video-id-input"
        />
        <el-button type="primary" size="small" @click="handleFilterByVideo">筛选</el-button>
      </div>
    </div>

    <!-- 弹幕列表：卡片式列表 -->
    <div class="danmu-list" v-loading="loading">
      <div v-if="list.length" class="danmu-cards">
        <div v-for="item in list" :key="item.ID" class="danmu-card">
          <div class="dm-head">
            <span class="dm-id">弹幕 #{{ item.ID }}</span>
            <span class="dm-date">{{ formatTime(item.CreatedAt) }}</span>
          </div>
          <div class="dm-body">
            <span class="dm-content">{{ item.Content }}</span>
          </div>
          <div class="dm-foot">
            <span class="dm-manuscript" @click="goVideo(item.VideoID)">稿件 #{{ item.VideoID }}</span>
            <span class="dm-actions">
              <button class="delete-btn" @click="handleDelete(item)">删除</button>
            </span>
          </div>
        </div>
      </div>
      <el-empty v-else-if="!loading" description="暂无弹幕" />
    </div>

    <!-- 分页导航栏 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        @size-change="handlePageChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi } from '@/api/creator'

interface DanmuItem {
  ID: number
  VideoID: number
  UserID: number
  Content: string
  Time: number
  Color: string
  Mode: number
  CreatedAt: string
}

const mainTab = ref('all')
const videoId = ref('')
const list = ref<DanmuItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getDanmakuList({
      video_id: videoId.value,
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

const handleFilterByVideo = () => {
  page.value = 1
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const handleDelete = (item: DanmuItem) => {
  ElMessageBox.confirm(`确定删除这条弹幕吗？\n「${item.Content}」`, '删除弹幕', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const res: any = await creatorApi.deleteDanmaku(item.ID)
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
  padding: 12px 0;
}

.content-tabs {
  margin-bottom: 8px;
}

.danmu-filters {
  padding: 8px 0 16px;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.video-id-input {
  width: 220px;
}

/* 弹幕卡片列表：独立白卡 radius4px（对齐B站互动管理） */
.danmu-cards {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.danmu-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 16px 20px;
  transition: box-shadow .2s, border-color .2s;
}

.danmu-card:hover {
  border-color: #d0d3d8;
  box-shadow: 0 2px 12px rgba(0, 0, 0, .04);
}

.dm-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dm-id {
  font-size: 12px;
  color: #99a2aa;
}

.dm-date {
  margin-left: auto;
  font-size: 12px;
  color: #99a2aa;
}

.dm-body {
  margin-top: 12px;
  font-size: 14px;
  color: #18191c;
  line-height: 1.6;
}

.dm-foot {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #f0f1f4;
}

.dm-manuscript {
  font-size: 12px;
  color: #99a2aa;
  cursor: pointer;
}

.dm-manuscript:hover {
  color: #ff6699;
}

.dm-actions {
  margin-left: auto;
}

.delete-btn {
  height: 28px;
  padding: 0 12px;
  font-size: 12px;
  color: #505050;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 2px;
  cursor: pointer;
  transition: color .2s, border-color .2s;
}

.delete-btn:hover {
  color: #f56c6c;
  border-color: #f56c6c;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>