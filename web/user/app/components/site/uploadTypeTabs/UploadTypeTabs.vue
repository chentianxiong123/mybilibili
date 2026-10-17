<template>
    <div class="upload-type-tabs">
        <div
            class="upload-type-tab"
            :class="{ 'is-active': activeType === 'video' }"
            @click="goTo('video')"
        >
            <span class="tab-text">视频投稿</span>
            <span class="tab-line"></span>
        </div>
        <div
            class="upload-type-tab"
            :class="{ 'is-active': activeType === 'text' }"
            @click="goTo('text')"
        >
            <span class="tab-text">图文投稿</span>
            <span class="tab-line"></span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
    active: {
        type: String,
        default: 'video',
    },
})

const emit = defineEmits(['change'])

const activeType = computed(() => props.active)

function goTo(type: string) {
    if (type === props.active) return;
    emit('change', type);
}
</script>

<style scoped>
.upload-type-tabs {
    display: flex;
    align-items: flex-end;
    margin-bottom: 18px;
    border-bottom: 1px solid #e7e7e7;
}

.upload-type-tab {
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 14px 28px 0;
}

.tab-text {
    color: #18191c;
    font-size: 15px;
    font-weight: 500;
    padding-bottom: 14px;
}

.tab-line {
    height: 3px;
    width: 66px;
    border-radius: 2px 2px 0 0;
    background: transparent;
}

.upload-type-tab:hover .tab-text {
    color: var(--brand_pink);
}

.upload-type-tab.is-active .tab-text {
    color: var(--brand_pink);
}

.upload-type-tab.is-active .tab-line {
    background: var(--brand_pink);
}
</style>