<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@element-plus/icons-vue'
import MessageDropdown from './dropdowns/MessageDropdown.vue'

defineProps({
  totalUnreadCount: {
    type: Number,
    default: 0
  },
  isZoomed: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()

const showMessageDropdown = ref(false)
</script>

<template>
  <div 
    class="dropdown-container"
    @mouseenter="showMessageDropdown = true"
    @mouseleave="showMessageDropdown = false"
  >
    <div :class="['action-btn', { 'hide-text-on-small': isZoomed }]" @click="router.push('/message')">
      <div class="icon-with-badge">
        <el-icon><Message /></el-icon>
        <span v-if="totalUnreadCount > 0" class="badge">{{ totalUnreadCount > 99 ? '99+' : totalUnreadCount }}</span>
      </div>
      <span>消息</span>
    </div>
    <div class="dropdown-bridge"></div>
    <MessageDropdown v-show="showMessageDropdown" />
  </div>
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

.dropdown-container :deep(.message-dropdown) {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
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

.hide-text-on-small span {
  display: none;
}

.hide-text-on-small {
  min-width: 50px;
  width: 50px;
}
</style>