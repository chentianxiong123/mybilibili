<template>
    <div class="drafts-page">
        <DraftsBox />
    </div>
</template>

<script setup lang="ts">
import DraftsBox from '@/components/createCenter/DraftsBox.vue'
</script>

<style scoped>
/* B站 创作中心内容列 .content-body 已是 1283 居中，页面级不再加内边距(原 8px 致 1267≠1283) */
.drafts-page {
    padding: 0;
}
</style>
