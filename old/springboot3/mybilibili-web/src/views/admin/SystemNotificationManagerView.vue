<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Bell, Promotion } from '@element-plus/icons-vue'
import { messageApi } from '../../api/message.js'

const form = ref({
  content: ''
})

const sending = ref(false)

const handleBroadcast = async () => {
  if (!form.value.content.trim()) {
    ElMessage.warning('请输入通知内容')
    return
  }

  try {
    await ElMessageBox.confirm(
      '确定要向全站用户发送系统通知吗？此操作不可撤回。',
      '发送全站系统通知',
      {
        confirmButtonText: '确定发送',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    sending.value = true
    const res = await messageApi.broadcastSystemNotification({ content: form.value.content.trim() })
    if (res.code === 200) {
      ElMessage.success('全站系统通知发送成功')
      form.value.content = ''
    } else {
      ElMessage.error(res.message || '发送失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('发送系统通知失败:', error)
      ElMessage.error('发送失败')
    }
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="notification-manager">
    <div class="page-header">
      <h1 class="page-title">
        <el-icon><Bell /></el-icon>
        全站系统通知
      </h1>
      <p class="page-desc">向所有开启了系统通知的用户发送公告</p>
    </div>

    <el-card class="broadcast-card">
      <template #header>
        <div class="card-header">
          <span>发送系统通知</span>
        </div>
      </template>

      <div class="broadcast-form">
        <div class="form-item">
          <label class="form-label">通知内容</label>
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="5"
            placeholder="输入要发送给全站用户的系统通知内容"
            maxlength="500"
            show-word-limit
          />
        </div>

        <div class="form-actions">
          <el-button
            type="primary"
            :icon="Promotion"
            :loading="sending"
            :disabled="!form.content.trim()"
            @click="handleBroadcast"
          >
            发送全站通知
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="info-card">
      <template #header>
        <span>使用说明</span>
      </template>
      <div class="info-content">
        <ul>
          <li>通知将发送给所有在消息设置中开启了"系统通知"的用户</li>
          <li>未开启系统通知的用户不会收到此通知</li>
          <li>通知内容最长500字</li>
          <li>用户可在"消息中心 - 系统通知"中查看收到的通知</li>
          <li>请勿频繁发送，以免影响用户体验</li>
        </ul>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.notification-manager {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #18191c;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 8px;
}

.page-desc {
  font-size: 14px;
  color: #9499a0;
  margin: 0;
}

.broadcast-card {
  margin-bottom: 20px;
}

.card-header {
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
}

.broadcast-form {
  padding: 8px 0;
}

.form-item {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  margin-bottom: 8px;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.form-actions :deep(.el-button--primary) {
  background-color: #00a1d6;
  border-color: #00a1d6;
}

.form-actions :deep(.el-button--primary:hover) {
  background-color: #00b5e5;
  border-color: #00b5e5;
}

.info-card :deep(.el-card__header) {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}

.info-content ul {
  margin: 0;
  padding-left: 20px;
  color: #61666d;
  font-size: 14px;
  line-height: 2;
}
</style>
