<template>
    <div></div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: [() => navigateTo('/platform/upload/video', { replace: true })] })
</script>
