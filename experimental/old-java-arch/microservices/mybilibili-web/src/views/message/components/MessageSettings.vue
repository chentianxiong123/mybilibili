<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { messageApi } from '../../../api/message.js'

const settings = ref({
  privateMessageNotify: true,
  replyNotify: true,
  atNotify: true,
  likeNotify: true,
  systemNotify: true
})

const loading = ref(false)

const fetchSettings = async () => {
  try {
    const res = await messageApi.getMessageSettings()
    if (res.code === 200 && res.data) {
      settings.value = {
        privateMessageNotify: res.data.privateMessageNotify ?? true,
        replyNotify: res.data.replyNotify ?? true,
        atNotify: res.data.atNotify ?? true,
        likeNotify: res.data.likeNotify ?? true,
        systemNotify: res.data.systemNotify ?? true
      }
    }
  } catch (error) {
    console.error('获取消息设置失败:', error)
  }
}

const saveSettings = async () => {
  loading.value = true
  try {
    const res = await messageApi.updateMessageSettings(settings.value)
    if (res.code === 200) {
      ElMessage.success('设置已保存')
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    console.error('保存设置失败:', error)
    ElMessage.error('保存失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<template>
  <div class="message-settings">
    <div class="settings-header">
      <span class="header-title">消息设置</span>
    </div>

    <div class="settings-content">
      <!-- 私信消息通知 -->
      <div class="setting-section">
        <div class="section-title">私信消息通知</div>
        <div class="section-content">
          <el-switch
            v-model="settings.privateMessageNotify"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
      </div>

      <!-- 回复我的通知 -->
      <div class="setting-section">
        <div class="section-title">回复我的通知</div>
        <div class="section-content">
          <el-switch
            v-model="settings.replyNotify"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
      </div>

      <!-- @我的通知 -->
      <div class="setting-section">
        <div class="section-title">@我的通知</div>
        <div class="section-content">
          <el-switch
            v-model="settings.atNotify"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
      </div>

      <!-- 收到的赞通知 -->
      <div class="setting-section">
        <div class="section-title">收到的赞通知</div>
        <div class="section-content">
          <el-switch
            v-model="settings.likeNotify"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
      </div>

      <!-- 系统通知 -->
      <div class="setting-section">
        <div class="section-title">系统通知</div>
        <div class="section-content">
          <el-switch
            v-model="settings.systemNotify"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
      </div>

      <!-- 保存按钮 -->
      <div class="settings-footer">
        <el-button
          type="primary"
          :loading="loading"
          @click="saveSettings"
        >
          保存设置
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-settings {
  width: 680px;
  background-color: #fff;
  border-right: 1px solid #e3e5e7;
  height: 100%;
  overflow-y: auto;
}

.settings-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e3e5e7;
  position: sticky;
  top: 0;
  background-color: #fff;
  z-index: 10;
}

.header-title {
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
}

.settings-content {
  padding: 20px;
}

.setting-section {
  padding: 16px 0;
  border-bottom: 1px solid #f1f2f3;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.setting-section:last-of-type {
  border-bottom: none;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}

.section-desc {
  font-size: 13px;
  color: #9499a0;
  font-weight: normal;
}

.section-content {
  padding-left: 8px;
}

.section-content :deep(.el-switch) {
  --el-switch-on-color: #00a1d6;
}

.settings-footer {
  padding: 20px 0;
  display: flex;
  justify-content: flex-start;
}

.settings-footer :deep(.el-button--primary) {
  background-color: #00a1d6;
  border-color: #00a1d6;
}

.settings-footer :deep(.el-button--primary:hover) {
  background-color: #00b5e5;
  border-color: #00b5e5;
}
</style>
