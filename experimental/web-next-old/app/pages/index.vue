<template>
  <HomeView />
</template>

<script setup lang="ts">
import HomeView from '~/views/web/HomeView.vue'
definePageMeta({ layout: 'home', keepalive: true })
</script>
