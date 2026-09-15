<script setup>
import { ElMessage } from 'element-plus'
import { Edit, Check } from '@element-plus/icons-vue'
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
  isFollowing: {
    type: Boolean,
    default: false
  },
  followLoading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['follow', 'send-message', 'avatar-click'])

// 保存个人签名
const saveSignature = async () => {
  try {
    const response = await userApi.updateUser(props.userId, { signature: props.userInfo.signature })
    if (response.code === 200) {
      console.log('个人签名保存成功')
    }
  } catch (error) {
    console.error('保存个人签名失败:', error)
  }
}
</script>

<template>
  <div class="profile-header">
    <!-- 封面图片 -->
    <div class="cover-image" :style="{ backgroundImage: `url(${userInfo.cover})` }">
      <!-- 个人信息 -->
      <div class="profile-info">
        <div class="avatar-container" :class="{ 'clickable': isOwnSpace }" @click="emit('avatar-click')" :title="isOwnSpace ? '点击修改头像' : ''">
          <img loading="lazy" decoding="async" :src="userInfo.avatar" alt="头像" class="user-avatar">
          <div v-if="isOwnSpace" class="avatar-overlay">
            <el-icon :size="20"><Edit /></el-icon>
          </div>
        </div>
        <div class="user-details">
          <h1 class="username">
            {{ userInfo.username }}
            <span v-if="userInfo.gender === 1" class="gender-icon male" title="男">♂</span>
            <span v-else-if="userInfo.gender === 2" class="gender-icon female" title="女">♀</span>
          </h1>
          <textarea
            class="bio-input"
            v-model="userInfo.signature"
            :placeholder="isOwnSpace ? '点击编辑个人签名' : '暂无个人签名'"
            @blur="saveSignature"
            :disabled="!isOwnSpace"
          ></textarea>
        </div>
      </div>

      <!-- 右侧操作按钮 -->
      <button
        class="action-btn follow-btn profile-btn"
        v-if="!isOwnSpace"
        @click="emit('follow')"
        :class="{ 'following': isFollowing }"
        :disabled="followLoading"
      >
        <el-icon v-if="isFollowing"><Check /></el-icon>
        <span>{{ isFollowing ? '已关注' : '+ 关注' }}</span>
      </button>
      <button class="action-btn message-btn profile-btn" v-if="!isOwnSpace" @click="emit('send-message')">发消息</button>
    </div>
  </div>
</template>

<style scoped>
/* 背景框容器 */
.profile-header {
  position: relative;
  width: 100%;
  max-width: 1200px;
  height: 200px;
  background-color: #f0f0f0;
  overflow: hidden;
  margin: 0 auto;
  margin-bottom: 0;
}

/* 封面图片 */
.cover-image {
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  position: relative;
}

/* 个人信息 */
.profile-info {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 15px 20px;
  display: flex;
  align-items: center;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.6), transparent);
  color: #fff;
}

/* 顶部个人资料操作按钮 */
.profile-btn {
  position: absolute;
  bottom: 10px;
  background-color: transparent;
  color: #ffffff !important;
  border: 1px solid rgba(255, 255, 255, 0.5);
  z-index: 20;
  font-size: 14px;
  font-weight: 500;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

/* 关注按钮 */
.profile-btn.follow-btn {
  right: 120px;
  background-color: transparent;
  color: #ffffff !important;
}

.profile-btn.follow-btn:hover {
  background-color: rgba(0, 174, 236, 0.9);
  border-color: #00aeec;
  color: #ffffff !important;
}

/* 已关注状态 */
.profile-btn.follow-btn.following {
  background-color: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.5);
}

.profile-btn.follow-btn.following:hover {
  background-color: rgba(255, 100, 100, 0.8);
  border-color: rgba(255, 100, 100, 0.9);
}

.profile-btn.follow-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* 发消息按钮 */
.profile-btn.message-btn {
  right: 20px;
  background-color: transparent;
  color: #ffffff !important;
}

.profile-btn.message-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.8);
  color: #ffffff !important;
}

/* 头像容器 */
.avatar-container {
  margin-right: 20px;
  z-index: 10;
  border: 4px solid #fff;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  position: relative;
  width: 88px;
  height: 88px;
}

.avatar-container.clickable {
  cursor: pointer;
}

.avatar-container.clickable:hover .user-avatar {
  filter: brightness(0.8);
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: all 0.3s ease;
}

.avatar-container.clickable:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay .el-icon {
  color: #fff;
}

/* 用户头像 */
.user-avatar {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 50%;
  display: block;
  margin: 0 auto;
}

/* 用户详情 */
.user-details {
  flex: 1;
}

/* 用户名 */
.username {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 性别符号 */
.gender-icon {
  font-size: 18px;
  font-weight: normal;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.gender-icon.male {
  color: #4fc3f7;
}

.gender-icon.female {
  color: #ff8a80;
}

/* 个人签名 */
.bio-input {
  font-size: 14px;
  margin: 0;
  opacity: 0.9;
  width: 100%;
  min-height: 40px;
  max-height: 80px;
  border: 1px solid transparent;
  background-color: transparent;
  color: #fff;
  resize: none;
  padding: 4px 0;
  font-family: inherit;
}

.bio-input:focus {
  outline: none;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.bio-input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.bio-input:disabled {
  cursor: not-allowed;
  opacity: 0.7;
  background-color: transparent;
}

/* 个人资料样式（保持原页面行为一致） */
.profile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>