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
  { path: '/m/index', name: '首页', icon: '🏠' },
  { path: '/m/live', name: '直播', icon: '📺' },
  { path: '/m/search', name: '搜索', icon: '🔍' },
  { path: '/m/space', name: '我的', icon: '👤', needAuth: true }
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
      <span class="tab-icon">{{ tab.icon }}</span>
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
  background: $bg-white;
  border-top: 1px solid $border-color;
  display: flex;
  z-index: 1000;
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  cursor: pointer;

  &.active {
    .tab-name {
      color: $theme-pink;
      font-weight: 600;
    }
  }
}

.tab-icon {
  font-size: 18px;
}

.tab-name {
  font-size: 10px;
  color: $text-secondary;
}
</style>