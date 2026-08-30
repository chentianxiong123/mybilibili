<script setup>
import { ref, onMounted } from 'vue'
defineProps({
  placeholder: { type: String, default: '搜索...' }
})

const avatar = ref('')
onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      avatar.value = user.avatar || user.avatarUrl || ''
    } catch (e) {}
  }
})
</script>

<template>
  <div class="mobile-header">
    <router-link to="/m/space" class="user-avatar">
      <img :src="avatar || '../assets/noface.gif'" alt="user" />
    </router-link>
    <router-link to="/m/search" class="search-box">
      <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <span class="placeholder">{{ placeholder }}</span>
    </router-link>
    <router-link to="/m/message" class="msg-icon">
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="2" y="4" width="20" height="16" rx="2" />
        <polyline points="2,4 12,13 22,4" />
      </svg>
    </router-link>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.mobile-header {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 16px 0 16px;
  background: $bg-white;
  gap: 16px;
  position: sticky;
  top: 0;
  z-index: 10;

  .user-avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
    img { width: 100%; height: 100%; object-fit: cover; }
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    height: 28px;
    background: #f4f4f4;
    border-radius: 14px;
    padding: 0 12px;
    color: #999;
    font-size: 13px;

    .search-icon {
      width: 14px;
      height: 14px;
      color: #999;
    }

    .placeholder { margin-left: 6px; }
  }

  .msg-icon {
    flex-shrink: 0;
    color: #a0a0a0;
    display: flex;
    align-items: center;
    svg {
      width: 24px;
      height: 24px;
    }
  }
}
</style>