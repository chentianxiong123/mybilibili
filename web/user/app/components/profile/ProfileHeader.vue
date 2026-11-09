<script setup lang="ts">
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
    }
  } catch (error) {
    console.error('保存个人签名失败:', error)
  }
}
</script>

<template>
  <div class="profile-header">
    <!-- 封面图片（通栏 1920×200，与 space.bilibili.com 一致） -->
    <div class="cover-image" :style="{ backgroundImage: `url(${userInfo.cover})` }">
      <!-- 1640 居中内容容器（140px 左右边距） -->
      <div class="banner-content">
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
        <div class="banner-actions">
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
    </div>
  </div>
</template>

<style scoped>
/* 背景框容器（通栏 1920×200，与 space.bilibili.com .header.space-header 一致） */
.profile-header {
  position: relative;
  width: 100%;
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

/* 1640 居中内容容器（1920 视口下左右各 140px，与 B 站 140px 边距一致） */
.banner-content {
  position: relative;
  width: 100%;
  max-width: 1640px;
  height: 100%;
  margin: 0 auto;
}

/* 个人信息（左下角，与 B 站 .upinfo 一致：头像顶 y119、名字顶 y127） */
.profile-info {
  position: absolute;
  top: 119px;
  left: 0;
  display: flex;
  align-items: flex-start;
  background: none;
  color: #fff;
}

/* 右侧操作按钮组（右下角，与 B 站 .space-head-follow 一致） */
.banner-actions {
  position: absolute;
  right: 0;
  bottom: 22px;
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 顶部个人资料操作按钮 */
.profile-btn {
  background-color: transparent;
  color: #ffffff !important;
  border: 1px solid rgba(255, 255, 255, 0.5);
  z-index: 20;
  font-size: 14px;
  font-weight: 500;
  text-shadow: none;
  border-radius: 6px;
}

/* 关注按钮（B 站 .space-follow-btn：150×34 实底 #00aeec / 14-600 / 圆角 6） */
.profile-btn.follow-btn {
  min-width: 96px;
  height: 34px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #00aeec;
  border: none;
  color: #ffffff !important;
  font-size: 14px;
  font-weight: 600;
  border-radius: 6px;
}

.profile-btn.follow-btn:hover {
  background-color: #00a1d6;
  border: none;
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
  min-width: 96px;
  height: 34px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  color: #ffffff !important;
  border-radius: 6px;
}

.profile-btn.message-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.8);
  color: #ffffff !important;
}

/* 头像容器（B 站 .space-avatar 68×68 圆形 + 白描边） */
.avatar-container {
  margin-right: 22px;
  z-index: 10;
  border: 3px solid rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  position: relative;
  width: 68px;
  height: 68px;
  box-sizing: border-box;
  flex-shrink: 0;
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
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  display: block;
  margin: 0 auto;
}

/* 用户详情 */
.user-details {
  flex: 1;
  margin-top: 8px;
}

/* 用户名（B 站 .nickname 24px/700 白色） */
.username {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
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

/* banner 下方不留白，使 nav 紧贴 banner 底部（B 站 nav @y200） */
</style>