<script setup>
import { userApi } from '@/api/client'

const props = defineProps({
  userInfo: {
    type: Object,
    required: true
  },
  userId: {
    type: [String, Number],
    default: null
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  },
  activeTab: {
    type: String,
    default: ''
  }
})

// 保存公告
const saveAnnouncement = async () => {
  try {
    const response = await userApi.updateUser(props.userId, { announcement: props.userInfo.announcement })
    if (response.code === 200) {
      console.log('公告保存成功')
    }
  } catch (error) {
    console.error('保存公告失败:', error)
  }
}
</script>

<template>
  <div class="profile-sidebar">
    <!-- 公告 -->
    <div class="content-section" v-if="activeTab === '主页'">
      <h3 class="section-title">公告</h3>
      <textarea
        class="announcement-input"
        v-model="userInfo.announcement"
        :placeholder="isOwnSpace ? '点击编辑公告' : '暂无公告'"
        @blur="saveAnnouncement"
        :disabled="!isOwnSpace"
      ></textarea>
    </div>

    <!-- 个人资料 - 动态页面显示 -->
    <div class="content-section" v-if="activeTab === '动态' || activeTab === '主页'">
      <h3 class="section-title">个人资料</h3>
      <div class="profile-details">
        <div class="detail-row">
          <div class="detail-item">
            <span class="detail-label">UID</span>
            <span class="detail-value">{{ userInfo.uid }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">生日</span>
            <span class="detail-value">{{ userInfo.birthday }}</span>
          </div>
        </div>
      </div>
      <!-- 标签区域 -->
      <div class="profile-tags" v-if="userInfo.tags && userInfo.tags.length > 0">
        <span class="profile-tag" v-for="tag in userInfo.tags" :key="tag">{{ tag }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 内容区域通用样式 */
.content-section {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0 0 16px 0;
}

/* 公告输入框 */
.announcement-input {
  width: 100%;
  min-height: 100px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 10px;
  font-size: 14px;
  color: #333;
  resize: vertical;
  font-family: inherit;
  line-height: 1.5;
}

.announcement-input:focus {
  outline: none;
  border-color: #00aeec;
}

.announcement-input::placeholder {
  color: #999;
}

.announcement-input:disabled {
  cursor: not-allowed;
  background-color: #f5f5f5;
  color: #999;
}

/* 公告样式 */
.announcement-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.announcement-item {
  font-size: 14px;
  color: #333;
  margin: 0;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.announcement-item:last-child {
  border-bottom: none;
}

/* 个人资料样式 */
.profile-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  gap: 40px;
}

.detail-row .detail-item {
  flex: 1;
}

.detail-row .detail-label {
  color: #999;
  font-weight: normal;
  min-width: auto;
  margin-right: 8px;
}

.detail-row .detail-value {
  color: #666;
}

.detail-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 14px;
}

.detail-label {
  font-weight: 600;
  color: #666;
  min-width: 60px;
}

.detail-value {
  color: #333;
  flex: 1;
  word-break: break-word;
}

/* 标签样式 */
.profile-tags {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.profile-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  background-color: #e3f2fd;
  color: #2196f3;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.profile-tag::before {
  content: '';
  display: inline-block;
  width: 14px;
  height: 14px;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%232196f3'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z'/%3E%3C/svg%3E");
  background-size: contain;
  background-repeat: no-repeat;
}

.profile-tag:hover {
  background-color: #bbdefb;
}
</style>