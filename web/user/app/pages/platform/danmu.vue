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

    <!-- 弹幕列表 -->
    <div class="danmu-list" v-loading="loading">
      <el-table :data="list" stripe style="width: 100%">
        <el-table-column prop="ID" label="弹幕ID" width="100"></el-table-column>
        <el-table-column label="弹幕内容" min-width="280">
          <template #default="scope">
            <span class="danmu-content">{{ scope.row.Content }}</span>
          </template>
        </el-table-column>
        <el-table-column label="所属稿件" width="140">
          <template #default="scope">
            <el-button link type="primary" size="small" @click="goVideo(scope.row.VideoID)">
              稿件 #{{ scope.row.VideoID }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="发送时间" width="180">
          <template #default="scope">{{ formatTime(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button type="danger" size="small" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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
  padding: 8px 0;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.video-id-input {
  width: 220px;
}

.danmu-content {
  color: var(--text1);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>