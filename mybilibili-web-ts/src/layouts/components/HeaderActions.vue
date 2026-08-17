<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Star, Clock, ChatDotRound, Upload } from '@element-plus/icons-vue'
import FavoriteDropdown from './dropdowns/FavoriteDropdown.vue'
import HistoryDropdown from './dropdowns/HistoryDropdown.vue'
import DynamicDropdown from './dropdowns/DynamicDropdown.vue'
import { usePrefetch } from '@/composables/usePrefetch'

defineProps({
  isZoomed: {
    type: Boolean,
    default: false
  },
  dynamicUnreadCount: {
    type: Number,
    default: 0
  }
})

const router = useRouter()
const { prefetch } = usePrefetch()

const showFavoriteDropdown = ref(false)
const showHistoryDropdown = ref(false)
const showDynamicDropdown = ref(false)
</script>

<template>
  <div
    class="dropdown-container"
    @mouseenter="showFavoriteDropdown = true"
    @mouseleave="showFavoriteDropdown = false"
  >
    <div :class="['action-btn', { 'hide-text-on-small': isZoomed }]" @click="router.push('/profile/favorites')" @mouseenter="prefetch('/profile/favorites')">
      <el-icon><Star /></el-icon>
      <span>收藏</span>
    </div>
    <div class="dropdown-bridge"></div>
    <FavoriteDropdown v-show="showFavoriteDropdown" />
  </div>

  <div
    class="dropdown-container"
    @mouseenter="showHistoryDropdown = true"
    @mouseleave="showHistoryDropdown = false"
  >
    <div :class="['action-btn', { 'hide-text-on-small': isZoomed }]" @click="router.push('/history')" @mouseenter="prefetch('/history')">
      <el-icon><Clock /></el-icon>
      <span>历史</span>
    </div>
    <div class="dropdown-bridge"></div>
    <HistoryDropdown v-show="showHistoryDropdown" @navigate="showHistoryDropdown = false" />
  </div>

  <div
    class="dropdown-container"
    @mouseenter="showDynamicDropdown = true"
    @mouseleave="showDynamicDropdown = false"
  >
    <div :class="['action-btn', { 'hide-text-on-small': isZoomed }]" @click="router.push('/dynamic')" @mouseenter="prefetch('/dynamic')">
      <div class="icon-with-badge">
        <el-icon><ChatDotRound /></el-icon>
        <span v-if="dynamicUnreadCount > 0" class="badge">{{ dynamicUnreadCount > 99 ? '99+' : dynamicUnreadCount }}</span>
      </div>
      <span>动态</span>
    </div>
    <div class="dropdown-bridge"></div>
    <DynamicDropdown v-show="showDynamicDropdown" />
  </div>

  <el-button type="primary" @click="router.push('/create-center')" @mouseenter="prefetch('/create-center')" class="upload-btn upload-btn-right">
    <el-icon><Upload /></el-icon>
    <span>投稿</span>
  </el-button>
</template>

<style scoped>
.dropdown-container {
  position: relative;
  display: inline-block;
  z-index: 1001;
}

.icon-with-badge {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background-color: #fb7299;
  color: #fff;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  font-weight: 500;
}

.dropdown-bridge {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  height: 8px;
  background: transparent;
  z-index: 999;
}

.dropdown-container :deep(.favorite-dropdown) {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-60%);
  margin-top: 8px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.dropdown-container :deep(.history-dropdown) {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-70%);
  margin-top: 8px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.dropdown-container :deep(.dynamic-dropdown) {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-80%);
  margin-top: 8px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 0;
  border: none;
  background: transparent;
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  min-width: 60px;
  height: 60px;
  text-align: center;
  width: 60px;
}

.action-btn:hover {
  background: transparent;
}

.action-btn .el-icon {
  font-size: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  transform: translateY(0) scale(1);
}

.action-btn:hover .el-icon {
  animation: bounce 0.6s ease-in-out;
}

@keyframes bounce {
  0% {
    transform: translateY(0) scale(1);
  }
  25% {
    transform: translateY(-10px) scale(1.1);
  }
  50% {
    transform: translateY(0) scale(1);
  }
  75% {
    transform: translateY(-5px) scale(1.05);
  }
  100% {
    transform: translateY(0) scale(1);
  }
}

.action-btn span {
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  margin-top: 0;
  width: 100%;
}

.upload-btn {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 6px;
  background-color: #fb7299;
  color: #fff;
  border: none;
  border-radius: 15px;
  padding: 0 20px;
  font-size: 14px;
  height: 50px;
  transition: all 0.3s ease;
}

.upload-btn:hover {
  background-color: #f75982;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(251, 114, 153, 0.3);
}

.upload-btn .el-icon {
  font-size: 18px;
}

.upload-btn span {
  font-size: 14px;
  margin-top: 0;
}

.upload-btn-right {
  margin-right: 16px;
}

.hide-text-on-small span {
  display: none;
}

.hide-text-on-small {
  min-width: 50px;
  width: 50px;
}
</style>