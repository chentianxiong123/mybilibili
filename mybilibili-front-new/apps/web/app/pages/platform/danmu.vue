<template>
  <div class="danmu-manage">
    <div class="content-tabs">
      <el-tabs v-model="activeTab" type="card" @tab-change="loadData">
        <el-tab-pane label="全部弹幕" name="all"></el-tab-pane>
      </el-tabs>
    </div>

    <!-- 统计行 -->
    <div class="stat-bar" v-if="total > 0">
      共 <b>{{ total }}</b> 条弹幕
    </div>

    <!-- 弹幕列表 -->
    <div class="danmu-list" v-loading="loading">
      <el-empty v-if="!loading && list.length === 0" description="暂无弹幕" />

      <div v-for="item in list" :key="item.ID" class="danmu-item">
        <span class="danmu-content" :style="{ color: item.Color || '#fff' }">{{ item.Content }}</span>
        <span class="danmu-time">{{ formatTime(item.CreatedAt) }}</span>
        <span class="danmu-video" v-if="item.VideoID">
          <el-button link size="small" @click="goVideo(item.VideoID)">查看视频 #{{ item.VideoID }}</el-button>
        </span>
        <span class="danmu-actions">
          <el-button type="danger" link size="small" @click="handleDelete(item)">删除</el-button>
        </span>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrap" v-if="total > pageSize">
        <el-pagination
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="handlePageChange"
        />
      </div>
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

const activeTab = ref('all')
const list = ref<DanmuItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await creatorApi.getDanmakuList({
      video_id: '',
      page: page.value,
      size: pageSize,
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

const handlePageChange = (p: number) => {
  page.value = p
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
  return t.replace('T', ' ').slice(0, 19)
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
.danmu-manage {
  padding: 20px 24px;
}

.content-tabs {
  margin-bottom: 16px;
}

.stat-bar {
  font-size: 13px;
  color: var(--text3, #9499a0);
  margin-bottom: 12px;
}

.stat-bar b {
  color: var(--brand_pink, #fb7299);
  font-size: 16px;
}

.danmu-list {
  min-height: 300px;
}

.danmu-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--graph_bg_thin, #f0f0f0);
}

.danmu-item:last-child {
  border-bottom: none;
}

.danmu-content {
  flex: 1;
  font-size: 14px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.08);
  padding: 4px 12px;
  border-radius: 4px;
  color: var(--text1, #18191c);
}

.danmu-time {
  font-size: 12px;
  color: var(--text3, #9499a0);
  flex-shrink: 0;
}

.danmu-video {
  flex-shrink: 0;
}

.danmu-actions {
  flex-shrink: 0;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 0;
}
</style>
