<template>
  <div></div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: [() => navigateTo('/platform/upload/text/new-article', { replace: true })] })
</script>
