<script setup>
import { ref } from 'vue'

defineProps({
  data: { type: Array, required: true }
})
const emit = defineEmits(['click'])
const visible = ref(false)

const show = () => { visible.value = true }
const hide = () => { visible.value = false }
const handleClick = (tab) => { hide(); emit('click', tab) }

defineExpose({ show })
</script>

<template>
  <transition name="drawer-fade">
    <div v-if="visible" class="drawer-overlay" @click="hide">
      <div class="drawer-panel" @click.stop>
        <div
          v-for="tab in data"
          :key="tab.id"
          class="drawer-item"
          @click="handleClick(tab)"
        >
          {{ tab.name }}
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 1000;
}

.drawer-panel {
  background: $bg-white;
  border-radius: 0 0 12px 12px;
  padding: 12px 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.drawer-item {
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 14px;
  background: $bg-color;
  color: $text-primary;
  cursor: pointer;

  &:hover, &:active {
    background: $theme-pink;
    color: #fff;
  }
}

.drawer-fade-enter-active, .drawer-fade-leave-active {
  transition: opacity 0.2s;
}
.drawer-fade-enter-from, .drawer-fade-leave-to {
  opacity: 0;
}
</style>