<template>
  <div></div>
</template>

<script setup lang="ts">
const route = useRoute()
const m = String(route.params.id || '').match(/(\d+)/)
navigateTo(m ? `/read/cv${m[1]}` : '/read/home', { replace: true })
</script>
