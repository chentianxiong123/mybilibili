<template>
  <NuxtPage />
</template>

<script setup lang="ts">
definePageMeta({ layout: 'simple' })
if (import.meta.client) {
  const route = useRoute()
  if (route.path === '/message') {
    navigateTo('/message/private', { replace: true })
  }
}
</script>
