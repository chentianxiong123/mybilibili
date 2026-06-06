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
  { path: '/m/dynamic', name: '关注' },
  { path: 'publish', name: '+' },
  { path: '/m/mall', name: '会员购' },
  { path: '/m/space', name: '我的', needAuth: true }
]

function showToast(msg: string) {
  const el = document.createElement('div')
  el.className = 'wap-toast'
  el.textContent = msg
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2500)
}

const handleTabClick = (tab: typeof tabs[0]) => {
  if (tab.path === 'publish') {
    if (!isLoggedIn.value) {
      router.push('/m/login')
      return
    }
    if (route.path !== '/m/dynamic') {
      router.push('/m/dynamic').then(() => {
        setTimeout(() => {
          window.dispatchEvent(new CustomEvent('open-publish-modal'))
        }, 300)
      })
    } else {
      window.dispatchEvent(new CustomEvent('open-publish-modal'))
    }
    return
  }

  if (tab.path === '/m/mall') {
    showToast('会员购功能正在火热开发中...')
    return
  }

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
      :class="['tab-item', { active: tab.path !== 'publish' && currentPath.startsWith(tab.path), 'publish-btn-container': tab.path === 'publish' }]"
      @click="handleTabClick(tab)"
    >
      <!-- 首页 -->
      <svg v-if="tab.path === '/m/index'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
        <polyline points="9 22 9 12 15 12 15 22"></polyline>
      </svg>
      <!-- 关注 -->
      <svg v-else-if="tab.path === '/m/dynamic'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <circle cx="12" cy="12" r="10"></circle>
        <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
        <path d="M2 12h20"></path>
      </svg>
      <!-- 发布 (+) -->
      <div v-else-if="tab.path === 'publish'" class="big-plus-btn">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
      </div>
      <!-- 会员购 -->
      <svg v-else-if="tab.path === '/m/mall'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"></path>
        <line x1="3" y1="6" x2="21" y2="6"></line>
        <path d="M16 10a4 4 0 0 1-8 0"></path>
      </svg>
      <!-- 我的 -->
      <svg v-else class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
        <line x1="8" y1="21" x2="16" y2="21"></line>
        <line x1="12" y1="17" x2="12" y2="21"></line>
      </svg>
      <span v-if="tab.path !== 'publish'" class="tab-name">{{ tab.name }}</span>
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
  height: 52px;
  background: #fff;
  border-top: 1px solid #e3e5e7;
  display: flex;
  z-index: 1000;
  padding-bottom: env(safe-area-inset-bottom, 0);
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  cursor: pointer;
  position: relative;

  &.active {
    .tab-icon { color: $theme-pink; }
    .tab-name { color: $theme-pink; font-weight: 500; }
  }

  &.publish-btn-container {
    justify-content: center;
    overflow: visible;
  }
}

.big-plus-btn {
  width: 40px;
  height: 28px;
  background-color: $theme-pink;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 0 2px 6px rgba(251, 114, 153, 0.35);
  transition: transform 0.2s, background-color 0.2s;

  &:active {
    transform: scale(0.9);
    background-color: $theme-pink-hover;
  }

  svg {
    width: 18px;
    height: 18px;
  }
}

.tab-icon {
  width: 20px;
  height: 20px;
  color: #61666d;
}

.tab-name {
  font-size: 10px;
  color: #61666d;
}
</style>
