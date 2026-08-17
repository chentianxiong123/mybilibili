<script setup>
import { useRouter } from 'vue-router'
import { User, Lock, Coin, Upload } from '@element-plus/icons-vue'
import { usePrefetch as _usePrefetch } from '@/composables/usePrefetch'

const prefetch = _usePrefetch().prefetch

const props = defineProps({
  userInfo: {
    type: Object,
    default: null
  },
  isLogged: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['showLogin', 'logout'])

const router = useRouter()

const handleAvatarClick = () => {
  if (props.isLogged && props.userInfo && props.userInfo.id) {
    router.push(`/profile/${props.userInfo.id}/home`)
  } else {
    emit('showLogin')
  }
}

const handleLogout = () => {
  emit('logout')
}
</script>

<template>
  <div class="avatar-container">
    <el-button link @click="handleAvatarClick" class="action-btn avatar-btn">
      <el-avatar 
        :size="60" 
        :src="userInfo?.avatar || '/default-avatar.svg'" 
        class="header-avatar" 
      />
    </el-button>
    
    <div class="user-profile-popup" v-if="isLogged && userInfo">
      <div class="popup-content">
        <div class="popup-username">{{ userInfo.nickname || userInfo.username || '用户名' }}</div>
        <div class="popup-level-coins">
          <span class="level">Lv.{{ userInfo.level || 1 }}</span>
          <span class="coins"><el-icon><Coin /></el-icon> {{ userInfo.coinCount || 0 }}</span>
        </div>
        
        <div class="stats-buttons">
          <div class="stat-button" @click="router.push(`/profile/${userInfo.id}/following`)">
            <div class="stat-number">{{ userInfo.followingCount || 0 }}</div>
            <div class="stat-label">关注</div>
          </div>
          <div class="stat-button" @click="router.push(`/profile/${userInfo.id}/followers`)">
            <div class="stat-number">{{ userInfo.followerCount || 0 }}</div>
            <div class="stat-label">粉丝</div>
          </div>
          <div class="stat-button" @click="router.push(`/profile/${userInfo.id}/dynamic`)">
            <div class="stat-number">{{ userInfo.dynamicCount || 0 }}</div>
            <div class="stat-label">动态</div>
          </div>
        </div>
        
        <div class="profile-options">
          <div class="option-item" @click="router.push('/personal-center/home')" @mouseenter="prefetch('/personal-center/home')">
            <el-icon><User /></el-icon>
            <span>个人中心</span>
            <span class="option-arrow">></span>
          </div>
          <div class="option-item" @click="router.push('/create-center')" @mouseenter="prefetch('/create-center')">
            <el-icon><Upload /></el-icon>
            <span>投稿管理</span>
            <span class="option-arrow">></span>
          </div>
          <div class="option-item logout-item" @click="handleLogout">
            <el-icon><Lock /></el-icon>
            <span>退出登录</span>
            <span class="option-arrow">></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.avatar-container {
  position: relative;
  display: inline-block;
  width: 60px;
  height: 60px;
  z-index: 1001;
  cursor: pointer;
}

.header-avatar {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 1003;
}

.avatar-container:hover .header-avatar {
  transform: translate(-20px, 30px) scale(1.7);
  z-index: 1004;
}

.user-profile-popup {
  position: absolute;
  top: 60px;
  left: calc(10px - 140px);
  width: 280px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 1002;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-10px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.avatar-container:hover .user-profile-popup {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.popup-content {
  padding: 50px 0 16px;
}

.popup-username {
  text-align: center;
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin: 0 0 8px;
}

.popup-level-coins {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 14px;
}

.popup-level-coins .level {
  color: #ff6699;
  font-weight: 500;
}

.popup-level-coins .coins {
  color: #ff9d00;
  display: flex;
  align-items: center;
  gap: 4px;
}

.stats-buttons {
  display: flex;
  justify-content: space-around;
  padding: 0 20px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.stat-button {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: background-color 0.3s;
}

.stat-button:hover {
  background-color: #f5f5f5;
}

.stat-number {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

.profile-options {
  padding: 8px 0;
}

.option-item {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  height: 40px;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  transition: all 0.3s;
  padding: 0 20px;
  gap: 12px;
  position: relative;
}

.option-arrow {
  color: #ccc;
  font-size: 12px;
  position: absolute;
  right: 20px;
}

.option-item .el-icon {
  font-size: 16px;
  color: #999;
  transition: color 0.3s;
  width: 20px;
  text-align: left;
}

.option-item span {
  flex: 1;
  text-align: left;
}

.option-item:hover {
  background-color: #f5f5f5;
  color: #00aeec;
}

.option-item:hover .el-icon {
  color: #00aeec;
}

.option-item:hover .option-arrow {
  color: #999;
}

.logout-item {
  color: #ff4d4f;
}

.logout-item .el-icon {
  color: #ff4d4f;
}

.logout-item:hover {
  background-color: rgba(255, 77, 79, 0.1);
  color: #ff4d4f;
}

.logout-item:hover .el-icon {
  color: #ff4d4f;
}

.action-btn.avatar-btn {
  margin-right: 10px;
}

.action-btn.avatar-btn:hover {
  background: transparent !important;
}

.action-btn.avatar-btn .el-icon {
  animation: none !important;
}

.action-btn.avatar-btn:hover .el-icon {
  animation: none !important;
  transform: none !important;
}

.avatar-btn {
  display: flex !important;
  flex-direction: row !important;
  align-items: center !important;
  justify-content: center !important;
  padding: 0 !important;
  min-width: 60px !important;
  width: 60px !important;
  height: 60px !important;
}

.avatar-btn .el-avatar {
  margin-bottom: 0;
  width: 50px !important;
  height: 50px !important;
}
</style>