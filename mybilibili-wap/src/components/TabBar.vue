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
  width: 100%;
  box-sizing: border-box;
  padding: 0;

  .tab-item {
    flex: 1;
    text-align: center;
    height: 48px;
    line-height: 48px;
    font-size: 18px;
    color: #61666d;
    position: relative;
    cursor: pointer;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.active {
      color: #fb7299;
      font-weight: 700;
      &::after {
        content: '';
        position: absolute;
        bottom: 4px;
        left: 50%;
        transform: translateX(-50%);
        width: 26px;
        height: 4px;
        background: #fb7299;
        border-radius: 4px;
      }
    }
  }
}

@media (max-width: 390px) {
  .tab-bar .tab-item {
    font-size: 16px;
  }
}
</style>
