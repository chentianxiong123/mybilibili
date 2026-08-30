<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../api/request'

const loading = ref(false)
const rooms = ref([])
const pendingReservations = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const activeTab = ref('list')
const statusFilter = ref(null)
const detailDialog = ref(false)
const currentRoom = ref(null)

const statusMap = { 0: '未开始/已预约', 1: '进行中', 2: '已结束/已拒绝' }
const statusTypeMap = { 0: 'warning', 1: 'success', 2: 'info' }

async function loadRooms() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    params.set('page', page.value)
    params.set('size', pageSize.value)
    if (statusFilter.value !== null && statusFilter.value !== '') params.set('status', statusFilter.value)
    const res = await request.get(`/admin/meeting/rooms?${params.toString()}`)
    if (res.code === 200) {
      rooms.value = res.data || []
      total.value = rooms.value.length
    }
  } catch (e) {
    console.error(e)
  } finally { loading.value = false }
}

async function loadPending() {
  try {
    const res = await request.get('/admin/meeting/pending')
    pendingReservations.value = res.data || []
  } catch (e) { console.error(e) }
}

async function approve(row) {
  try {
    await ElMessageBox.confirm(`确认通过「${row.roomName}」的预约？`, '审批')
    await request.post(`/admin/meeting/approve/${row.id}`)
    ElMessage.success('已通过')
    loadPending()
    loadRooms()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

async function reject(row) {
  try {
    await ElMessageBox.confirm(`确认拒绝「${row.roomName}」的预约？`, '审批')
    await request.post(`/admin/meeting/reject/${row.id}`)
    ElMessage.success('已拒绝')
    loadPending()
    loadRooms()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

async function forceEnd(row) {
  try {
    await ElMessageBox.confirm(`确认强制结束「${row.roomName}」？`, '警告')
    await request.post(`/admin/meeting/end/${row.id}`)
    ElMessage.success('已强制结束')
    loadRooms()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

function viewDetail(row) {
  currentRoom.value = row
  detailDialog.value = true
}

onMounted(() => { loadRooms(); loadPending() })
</script>

<template>
  <div class="meeting-admin">
    <h2 style="margin:0 0 24px;font-size:24px;font-weight:600;color:#333">会议管理</h2>
    <el-tabs v-model="activeTab" @tab-change="activeTab==='list'?loadRooms():loadPending()">
      <el-tab-pane label="所有会议" name="list">
        <div style="margin-bottom:12px">
          <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width:160px" @change="page=1;loadRooms()">
            <el-option label="未开始/已预约" :value="0" />
            <el-option label="进行中" :value="1" />
            <el-option label="已结束/已拒绝" :value="2" />
          </el-select>
        </div>
        <el-table :data="rooms" v-loading="loading" border stripe>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="roomName" label="会议名称" min-width="150" />
          <el-table-column prop="creatorName" label="创建人" width="100" />
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="statusTypeMap[row.status] || 'info'">{{ statusMap[row.status] || '未知' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="roomCode" label="邀请码" width="90" />
          <el-table-column prop="scheduledReason" label="预约事由" width="120" show-overflow-tooltip />
          <el-table-column label="预约/开始时间" width="180">
            <template #default="{ row }">
              <template v-if="row.scheduledStart">{{ row.scheduledStart?.replace('T',' ') }}</template>
              <template v-else>{{ row.startTime?.replace('T',' ') || '-' }}</template>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="viewDetail(row)">详情</el-button>
              <el-button link type="danger" size="small" @click="forceEnd(row)" v-if="row.status === 1 || row.status === 0">结束</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane :label="`待审批 (${pendingReservations.length})`" name="pending">
        <el-table :data="pendingReservations" border stripe>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="roomName" label="会议名称" min-width="150" />
          <el-table-column prop="creatorName" label="申请人" width="100" />
          <el-table-column prop="scheduledReason" label="事由" width="150" show-overflow-tooltip />
          <el-table-column label="预约时间段" width="200">
            <template #default="{ row }">
              {{ row.scheduledStart?.replace('T',' ') }} ~ {{ row.scheduledEnd?.replace('T',' ') }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button type="success" size="small" @click="approve(row)">通过</el-button>
              <el-button type="danger" size="small" @click="reject(row)">拒绝</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="detailDialog" title="会议详情" width="480px">
      <template v-if="currentRoom">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="会议名称">{{ currentRoom.roomName }}</el-descriptions-item>
          <el-descriptions-item label="邀请码">{{ currentRoom.roomCode }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ currentRoom.creatorName }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ statusMap[currentRoom.status] }}</el-descriptions-item>
          <el-descriptions-item label="最大人数">{{ currentRoom.maxParticipants }}</el-descriptions-item>
          <el-descriptions-item label="预约事由" :span="2">{{ currentRoom.scheduledReason || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预约时间" :span="2">{{ currentRoom.scheduledStart?.replace('T',' ') || '-' }} ~ {{ currentRoom.scheduledEnd?.replace('T',' ') || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ currentRoom.createTime?.replace('T',' ') }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>