<script setup>
import { inject } from 'vue'
import AppHeader from './components/AppHeader.vue'

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
    <AppHeader mode="white" @show-login="handleShowLogin" @logout="handleLogout" />
    <div class="layout-content">
      <slot />
    </div>
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
  min-height: calc(100vh - 80px);
  padding-top: 80px;
}
</style>