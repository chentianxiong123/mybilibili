<template>
  <Page class="page">
    <ActionBar title="404" />
    <StackLayout class="not-found-content">
      <Label text="404" class="not-found-code" />
      <Label text="页面不存在" class="not-found-text" />
      <Label text="你访问的页面可能已被删除或移动" class="not-found-hint" />
      <Button text="返回首页" class="back-home-btn" @tap="goHome" />
    </StackLayout>
  </Page>
</template>

<script lang="ts" setup>
import { $navigateTo } from 'nativescript-vue'

function goHome() {
  $navigateTo(require('./home/Index.vue').default, { clearHistory: true })
}
</script>

<style scoped>
.not-found-content {
  padding: 80 20;
  align-items: center;
}

.not-found-code {
  font-size: 72;
  font-weight: bold;
  color: #fb7299;
  margin-bottom: 16;
}

.not-found-text {
  font-size: 20;
  font-weight: bold;
  color: #18191c;
  margin-bottom: 8;
}

.not-found-hint {
  font-size: 14;
  color: #9499a0;
  margin-bottom: 40;
  text-align: center;
}

.back-home-btn {
  width: 200;
  height: 44;
  background-color: #fb7299;
  color: white;
  font-size: 15;
  font-weight: bold;
  border-radius: 22;
}
</style>