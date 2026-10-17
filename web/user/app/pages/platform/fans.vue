<template>
    <div class="fans-page">
        <FansManager :active="true" />
    </div>
</template>

<script setup lang="ts">
import FansManager from '@/components/createCenter/FansManager.vue'
</script>

<style scoped>
.fans-page {
    padding: 8px;
}
</style>