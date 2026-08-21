<template>
  <UserProfileView v-if="!isNestedTab" />
  <NuxtPage v-else />
</template>

<script setup lang="ts">
import UserProfileView from '~/views/web/UserProfileView.vue'
definePageMeta({ layout: 'simple' })

const route = useRoute()

// /profile/:id/xxx 这类子标签路由由 NuxtPage 渲染子页面；
// 裸访问 /profile/:id 时直接渲染主视图。
const isNestedTab = computed(() => route.path.split('/').filter(Boolean).length >= 3)
</script>