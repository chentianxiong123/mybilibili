<template>
    <div class="fans-page">
        <FansManager :active="true" />
    </div>
</template>

<script setup lang="ts">
import FansManager from '@/components/createCenter/FansManager.vue'
</script>

<style scoped>
/* B站 .fans_wrap cc-content-body pos-center 无内边距，内容列 1283 通宽 */
.fans-page {
    padding: 0;
}
</style>