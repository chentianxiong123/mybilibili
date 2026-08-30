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
  padding: 0 4px;

  &::-webkit-scrollbar { display: none; }

  .tab-item {
    flex: none;
    padding: 0 16px;
    height: 44px;
    line-height: 44px;
    font-size: 15px;
    color: #757575;
    position: relative;
    cursor: pointer;

    &.active {
      color: #fb7299;
      font-weight: 500;
      &::after {
        content: '';
        position: absolute;
        bottom: 6px;
        left: 50%;
        transform: translateX(-50%);
        width: 16px;
        height: 3px;
        background: #fb7299;
        border-radius: 2px;
      }
    }
  }
}
</style>