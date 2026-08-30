<template>
  <div class="h-screen w-screen bg-background text-text-primary overflow-hidden">
    <!-- Vue 宿主容器，未来我们可以在这里加入 Vue 的 UI 组件、路由或 AI 工作流面板 -->
    <div class="w-full h-full relative">
      <ReactAdapter :component="ReactAppRaw" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { markRaw } from "vue";
import ReactAdapter from "./components/ReactAdapter.vue";
import ReactApp from "./ReactApp";

// 包装 ReactApp 以免 Vue 的响应式系统深度劫持它，提升性能
const ReactAppRaw = markRaw(ReactApp);
</script>

<style scoped>
/* 可以在这里添加 Vue 宿主的样式 */
</style>
