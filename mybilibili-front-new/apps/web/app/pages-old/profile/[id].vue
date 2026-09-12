<template>
  <UserProfileView v-if="!isNestedTab" />
  <NuxtPage v-else />
</template>

<script setup lang="ts">
import UserProfileView from '~/views/web/UserProfileView.vue'
definePageMeta({ layout: 'simple' })

const route = useRoute()

// /profile/:id 裸路径直接渲染主视图；
// /profile/:id/xxx 子标签路由交给 NuxtPage 渲染对应子页面。
const isNestedTab = computed(() => route.path.split('/').filter(Boolean).length >= 3)
</script>