<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
:root {
  --bili-pink: #fb7299;
  --bili-pink-hover: #fc8bab;
  --bili-bg: #f4f5f7;
  --bili-text: #212121;
  --bili-text-secondary: #999;
  --bili-border: #e3e5e7;
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
