<template>
  <VideoView :manuscript-id="String(route.params.id)" :p="pValue" />
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import VideoView from '~/views/web/VideoView.vue'

const route = useRoute()
const pValue = Number.parseInt(String(route.query.p)) || 1

definePageMeta({ layout: 'simple' })
</script>