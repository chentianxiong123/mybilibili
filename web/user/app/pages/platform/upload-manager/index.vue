<template>
    <div></div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: [() => navigateTo('/platform/upload-manager/manuscript', { replace: true })] })
</script>
