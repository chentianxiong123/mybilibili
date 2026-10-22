<template>
  <div></div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: [() => navigateTo('/read/home', { replace: true })] })
</script>
