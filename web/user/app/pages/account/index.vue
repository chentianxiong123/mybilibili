<template>
  <div></div>
</template>

<script setup lang="ts">
// 布局（顶部 banner + 左侧导航 + 右侧内容区）已上移到父路由 app/pages/account.vue，
// /account 作为默认子路由直接落到「首页」卡片，避免出现空的右侧内容区。
definePageMeta({ middleware: [() => navigateTo('/account/home', { replace: true })] })
</script>
