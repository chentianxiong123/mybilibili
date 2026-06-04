<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import BottomTabBar from './components/BottomTabBar.vue'

const route = useRoute()

// 只在主要页面显示底部 TabBar
const showTabBar = computed(() => {
  const path = route.path
  return path === '/m/index' || path.startsWith('/m/live') || path.startsWith('/m/search') || path.startsWith('/m/space') || path.startsWith('/m/dynamic')
})

// 登录页不显示
const hideOnLogin = computed(() => route.path === '/m/login')
</script>

<template>
  <router-view />
  <BottomTabBar v-if="showTabBar && !hideOnLogin" />
</template>

<style>
/* 给有 TabBar 的页面底部留出空间 */
body {
  padding-bottom: 0;
  background-color: #f4f4f4; /* 匹配B站的灰色背景 */
  margin: 0;
}
</style>