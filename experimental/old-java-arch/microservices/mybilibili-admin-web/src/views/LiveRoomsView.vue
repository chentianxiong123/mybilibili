<script setup>
import { onMounted } from 'vue'
import { getLiveStatusText, getLiveStatusType, isLiveRoomStatus } from '../utils/liveMeetingStatus'
import { formatDateTime } from '../utils/adminPage'
import { useAdminLiveRooms } from '../composables/useAdminLiveRooms'

const {
  loading,
  rooms,
  total,
  page,
  pageSize,
  statusFilter,
  stats,
  statusOptions,
  handlePageChange,
  handleStatusFilterChange,
  toggleRoomStatus,
  initialize
} = useAdminLiveRooms()

onMounted(() => {
  initialize()
})
</script>

<template>
  <div class="live-rooms">
    <h2 class="page-title">直播管理</h2>

    <div class="stats-row">
      <el-card shadow="never" class="stat-card">
        <div class="stat-num">{{ stats.totalRooms }}</div>
        <div class="stat-label">总直播间</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-num" style="color:#67c23a">{{ stats.onlineRooms }}</div>
        <div class="stat-label">直播中</div>
      </el-card>
      <el-card shadow="never" class="stat-card">
        <div class="stat-num" style="color:#409eff">{{ stats.totalViewers }}</div>
        <div class="stat-label">在线观众</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <div class="toolbar">
        <el-select v-model="statusFilter" placeholder="状态筛选" @change="handleStatusFilterChange" style="width:140px">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </div>

      <el-table :data="rooms" v-loading="loading" border stripe style="width:100%">
        <el-table-column label="房间名" prop="roomName" min-width="160" />
        <el-table-column label="主播ID" prop="userId" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getLiveStatusType(row.status)">{{ getLiveStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="观看人数" prop="viewerCount" width="100" />
        <el-table-column label="分类" prop="category" width="120" />
        <el-table-column label="开播时间" prop="createdAt" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130">
          <template #default="{ row }">
            <el-button
              :type="isLiveRoomStatus(row.status) ? 'danger' : 'success'"
              size="small"
              @click="toggleRoomStatus(row)"
            >
              {{ isLiveRoomStatus(row.status) ? '下播' : '开播' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.live-rooms { padding: 20px; }
.page-title { margin: 0 0 24px; font-size: 24px; font-weight: 600; color: #333; }
.stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 20px; }
.stat-card :deep(.el-card__body) { display: flex; flex-direction: column; align-items: center; padding: 24px; }
.stat-num { font-size: 28px; font-weight: 700; color: #333; }
.stat-label { font-size: 14px; color: #999; margin-top: 6px; }
.toolbar { margin-bottom: 16px; display: flex; gap: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: center; }
</style>
