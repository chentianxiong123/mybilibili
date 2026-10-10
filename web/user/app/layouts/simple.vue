<script setup lang="ts">
import { inject } from 'vue'
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
</script>

<template>
  <div class="layout-simple">
    <HeaderBar :isFixHeaderBar="true" />
    <div class="layout-content">
      <slot />
    </div>
    <FeedbackFloat />
  </div>
</template>

<style scoped>
.layout-simple {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
  max-width: 2560px;
  margin: 0 auto;
}

.layout-content {
  flex: 1;
  min-height: calc(100vh - 64px);
}
</style>