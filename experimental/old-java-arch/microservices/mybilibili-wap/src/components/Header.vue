<script setup>
import { ref, onMounted } from 'vue'
import noface from '../assets/noface.gif'
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
      <img :src="avatar || noface" alt="user" />
    </router-link>
    <router-link to="/m/search" class="search-box">
      <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <span class="placeholder">{{ placeholder }}</span>
    </router-link>
    <button class="game-icon" aria-label="游戏中心">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M6 10h12a4 4 0 0 1 3.7 5.5l-.8 2a2.5 2.5 0 0 1-4.2.8L14.8 16H9.2l-1.9 2.3a2.5 2.5 0 0 1-4.2-.8l-.8-2A4 4 0 0 1 6 10z" />
        <path d="M8 13v4M6 15h4M16 14h.01M19 16h.01" />
      </svg>
    </button>
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
  height: 56px;
  padding: 6px 16px 4px;
  background: $bg-white;
  gap: 12px;
  position: sticky;
  top: 0;
  z-index: 20;

  .user-avatar {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
    img { width: 100%; height: 100%; object-fit: cover; }
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    min-width: 0;
    height: 38px;
    background: #f4f4f4;
    border: 1px solid #d3d6dc;
    border-radius: 20px;
    padding: 0 14px;
    color: #757a82;
    font-size: 18px;

    .search-icon {
      width: 21px;
      height: 21px;
      color: #757a82;
      flex: 0 0 auto;
    }

    .placeholder {
      min-width: 0;
      margin-left: 8px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .game-icon,
  .msg-icon {
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    border: 0;
    padding: 0;
    background: transparent;
    color: #61666d;
    display: flex;
    align-items: center;
    justify-content: center;
    svg {
      width: 29px;
      height: 29px;
    }
  }
}

@media (max-width: 390px) {
  .mobile-header {
    gap: 8px;
    padding-left: 12px;
    padding-right: 12px;

    .user-avatar {
      width: 34px;
      height: 34px;
    }

    .search-box {
      height: 34px;
      font-size: 16px;
    }

    .game-icon,
    .msg-icon {
      width: 30px;
      height: 30px;

      svg {
        width: 26px;
        height: 26px;
      }
    }
  }
}
</style>
