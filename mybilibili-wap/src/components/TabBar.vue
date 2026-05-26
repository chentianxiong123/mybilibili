<script setup>
defineProps({
  data: { type: Array, required: true },
  activeId: { type: Number, default: 0 }
})
const emit = defineEmits(['click'])
const handleClick = (tab) => emit('click', tab)
</script>

<template>
  <div class="tab-bar">
    <div
      v-for="tab in data"
      :key="tab.id"
      :class="['tab-item', { active: tab.id === activeId }]"
      @click="handleClick(tab)"
    >
      {{ tab.name }}
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.tab-bar {
  display: flex;
  overflow-x: auto;
  white-space: nowrap;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar { display: none; }

  .tab-item {
    flex-shrink: 0;
    padding: 0 14px;
    height: $tab-height;
    line-height: $tab-height;
    font-size: 14px;
    color: $text-secondary;

    &.active {
      color: $theme-pink;
      position: relative;
      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 50%;
        transform: translateX(-50%);
        width: 20px;
        height: 3px;
        background: $theme-pink;
        border-radius: 2px;
      }
    }
  }
}
</style>