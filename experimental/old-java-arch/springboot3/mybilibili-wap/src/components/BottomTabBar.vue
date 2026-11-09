<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const currentPath = computed(() => route.path)

const isLoggedIn = computed(() => {
  return !!localStorage.getItem('token')
})

const tabs = [
  { path: '/m/index', name: '首页' },
  { path: '/m/dynamic', name: '动态' },
  { path: '/m/live', name: '直播' },
  { path: '/m/space', name: '我的', needAuth: true }
]

const handleTabClick = (tab: typeof tabs[0]) => {
  if (tab.needAuth && !isLoggedIn.value) {
    router.push('/m/login')
    return
  }
  router.push(tab.path)
}
</script>

<template>
  <div class="bottom-tab-bar">
    <div
      v-for="tab in tabs"
      :key="tab.path"
      :class="['tab-item', { active: currentPath.startsWith(tab.path) }]"
      @click="handleTabClick(tab)"
    >
      <!-- 首页 -->
      <svg v-if="tab.path === '/m/index'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
        <polyline points="9 22 9 12 15 12 15 22"></polyline>
      </svg>
      <!-- 动态 -->
      <svg v-else-if="tab.path === '/m/dynamic'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
        <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
      </svg>
      <!-- 直播 -->
      <svg v-else-if="tab.path === '/m/live'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
        <line x1="8" y1="21" x2="16" y2="21"></line>
        <line x1="12" y1="17" x2="12" y2="21"></line>
      </svg>
      <!-- 我的 -->
      <svg v-else class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
        <circle cx="12" cy="7" r="4"></circle>
      </svg>
      <span class="tab-name">{{ tab.name }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.bottom-tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: #fff;
  border-top: 1px solid #e3e5e7;
  display: flex;
  z-index: 1000;
  padding-bottom: env(safe-area-inset-bottom, 0);
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  cursor: pointer;

  &.active {
    .tab-icon { color: $theme-pink; }
    .tab-name { color: $theme-pink; font-weight: 500; }
  }
}

.tab-icon {
  width: 22px;
  height: 22px;
  color: #999;
}

.tab-name {
  font-size: 10px;
  color: #999;
}
</style>