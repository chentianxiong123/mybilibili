<script setup>
import { onMounted } from 'vue'
import {
  canForceEndMeeting,
  getMeetingStatusText,
  getMeetingStatusType,
  MEETING_STATUS
} from '../utils/liveMeetingStatus'
import { formatDateTime } from '../utils/adminPage'
import { useAdminMeetings } from '../composables/useAdminMeetings'

const {
  loading,
  rooms,
  pendingReservations,
  total,
  page,
  pageSize,
  activeTab,
  statusFilter,
  detailDialog,
  currentRoom,
  loadRooms,
  loadPending,
  refreshRooms,
  approve,
  reject,
  forceEnd,
  viewDetail,
  initialize
} = useAdminMeetings()

onMounted(() => {
  initialize()
})
</script>

<template>
  <div class="meeting-admin">
    <h2 style="margin:0 0 24px;font-size:24px;font-weight:600;color:#333">会议管理</h2>
    <el-tabs v-model="activeTab" @tab-change="activeTab==='list'?loadRooms():loadPending()">
      <el-tab-pane label="所有会议" name="list">
        <div style="margin-bottom:12px">
          <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width:160px" @change="refreshRooms">
            <el-option label="未开始" :value="MEETING_STATUS.NOT_STARTED" />
            <el-option label="进行中" :value="MEETING_STATUS.IN_PROGRESS" />
            <el-option label="已结束" :value="MEETING_STATUS.ENDED" />
            <el-option label="待审批" :value="MEETING_STATUS.PENDING_APPROVAL" />
            <el-option label="已拒绝" :value="MEETING_STATUS.REJECTED" />
          </el-select>
        </div>
        <el-table :data="rooms" v-loading="loading" border stripe>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="roomName" label="会议名称" min-width="150" />
          <el-table-column prop="creatorName" label="创建人" width="100" />
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="getMeetingStatusType(row.status)">{{ getMeetingStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="roomCode" label="邀请码" width="90" />
          <el-table-column prop="scheduledReason" label="预约事由" width="120" show-overflow-tooltip />
          <el-table-column label="预约/开始时间" width="180">
            <template #default="{ row }">
              <template v-if="row.scheduledStart">{{ formatDateTime(row.scheduledStart) }}</template>
              <template v-else>{{ formatDateTime(row.startTime) }}</template>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="viewDetail(row)">详情</el-button>
              <el-button link type="danger" size="small" @click="forceEnd(row)" v-if="canForceEndMeeting(row.status)">结束</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div style="margin-top:16px;display:flex;justify-content:center">
          <el-pagination
            v-model:current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="prev, pager, next"
            @current-change="loadRooms"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane :label="`待审批 (${pendingReservations.length})`" name="pending">
        <el-table :data="pendingReservations" border stripe>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="roomName" label="会议名称" min-width="150" />
          <el-table-column prop="creatorName" label="申请人" width="100" />
          <el-table-column prop="scheduledReason" label="事由" width="150" show-overflow-tooltip />
          <el-table-column label="预约时间段" width="200">
            <template #default="{ row }">
              {{ formatDateTime(row.scheduledStart) }} ~ {{ formatDateTime(row.scheduledEnd) }}
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
          <el-descriptions-item label="状态">{{ getMeetingStatusText(currentRoom.status) }}</el-descriptions-item>
          <el-descriptions-item label="最大人数">{{ currentRoom.maxParticipants }}</el-descriptions-item>
          <el-descriptions-item label="预约事由" :span="2">{{ currentRoom.scheduledReason || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预约时间" :span="2">{{ formatDateTime(currentRoom.scheduledStart) }} ~ {{ formatDateTime(currentRoom.scheduledEnd) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(currentRoom.createTime) }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>
