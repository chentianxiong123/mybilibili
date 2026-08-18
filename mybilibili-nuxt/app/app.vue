<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup lang="ts">
import { provide } from 'vue'

const showLoginDialog = useState('showLoginDialog', () => false)
provide('showLoginDialog', showLoginDialog)
</script>

<style>
:root {
  --bili-pink: #fb7299;
  --bili-pink-hover: #fc8bab;
  --bili-bg: #f4f5f7;
  --bili-text: #212121;
  --bili-text-secondary: #999;
  --bili-border: #e3e5e7;

  --el-color-primary: #fb7299;
  --el-color-primary-light-3: #fc8bab;
  --el-color-primary-light-5: #fda3bc;
  --el-color-primary-light-7: #febbd0;
  --el-color-primary-light-8: #ffd3e0;
  --el-color-primary-light-9: #ffebf0;
  --el-color-primary-dark-2: #c95b7a;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  background-color: var(--bili-bg);
  color: var(--bili-text);
}

a {
  text-decoration: none;
  color: inherit;
}
</style>