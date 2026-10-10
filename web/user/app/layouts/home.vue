<script setup lang="ts">
import { inject, onMounted, onBeforeUnmount, ref } from 'vue'
import TopBackground from './components/TopBackground.vue'
import CategoryTabs from './components/CategoryTabs.vue'
import HeaderBar from '../components/teriteri/headerBar/HeaderBar.vue'
import FeedbackFloat from '../components/FeedbackFloat.vue'

const showLoginDialog = inject('showLoginDialog')

const handleShowLogin = () => {
  if (showLoginDialog) {
    showLoginDialog.value = true
  }
}

const handleLogout = () => {
}

const isFixHeaderBar = ref(false)
const onScroll = () => {
  isFixHeaderBar.value = window.scrollY >= 64
}

onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onBeforeUnmount(() => window.removeEventListener('scroll', onScroll))
</script>

<template>
  <div class="layout-home">
    <HeaderBar :isFixHeaderBar="isFixHeaderBar" />
    <div class="home-hero">
      <TopBackground />
      <CategoryTabs />
    </div>
    <div class="layout-content">
      <slot />
    </div>
    <FeedbackFloat />
  </div>
</template>

<style scoped>
.layout-home {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #fff;
  max-width: 2560px;
  margin: 0 auto;
}

.home-hero {
  margin-top: -64px;
}

.layout-content {
  flex: 1;
  min-height: calc(100vh - 64px);
  padding-top: 0;
}
</style>