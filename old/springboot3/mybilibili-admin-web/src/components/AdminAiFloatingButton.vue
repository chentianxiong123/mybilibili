<script setup>
import { ChatDotRound } from '@element-plus/icons-vue'

defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['toggle'])

const handleClick = () => {
  emit('toggle')
}
</script>

<template>
  <button
    class="floating-button"
    :class="{ active: visible }"
    @click="handleClick"
    aria-label="AI Chat"
  >
    <el-icon :size="24"><ChatDotRound /></el-icon>
  </button>
</template>

<style scoped>
.floating-button {
  position: fixed;
  bottom: 80px;
  right: 24px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #00a1d6, #00d6b2);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  box-shadow: 0 4px 12px rgba(0, 161, 214, 0.4);
  transition: all 0.3s ease;
}

.floating-button:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(0, 161, 214, 0.5);
}

.floating-button.active {
  background: #f56c6c;
  transform: rotate(45deg);
  box-shadow: 0 4px 12px rgba(245, 108, 108, 0.4);
}

.floating-button.active:hover {
  box-shadow: 0 6px 20px rgba(245, 108, 108, 0.5);
}
</style>