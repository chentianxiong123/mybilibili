<template>
  <span v-if="level > 0" class="level-badge" :class="levelClass" :title="'LV' + level">
    LV{{ level }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  level: {
    type: Number,
    default: 0
  }
})

const levelClass = computed(() => {
  if (props.level >= 6) return 'level-high'
  if (props.level >= 4) return 'level-mid'
  return 'level-low'
})
</script>

<style scoped>
.level-badge {
  display: inline-flex;
  align-items: center;
  font-size: 10px;
  font-weight: 600;
  color: #fff;
  padding: 1px 4px;
  border-radius: 3px;
  line-height: 1.2;
  white-space: nowrap;
  flex-shrink: 0;
}

.level-low {
  background: linear-gradient(135deg, #ff6b6b, #ee5a24);
}

.level-mid {
  background: linear-gradient(135deg, #f9ca24, #f0932b);
}

.level-high {
  background: linear-gradient(135deg, #6c5ce7, #a55eea);
}
</style>
