<template>
  <div class="th" :style="{ height: track.height + 'px' }">
    <div class="th-label">{{ track.label }}</div>
    <div class="th-kind">{{ track.kind }}</div>
  </div>
</template>

<script setup>
defineProps({ track: { type: Object, required: true } })
</script>

<style scoped>
.th {
  width: 60px; flex: 0 0 60px;
  background: #1a1d24;
  border-right: 1px solid #232733;
  border-bottom: 1px solid #1c2030;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  font-size: 11px; color: #b8bcc7;
  position: sticky; left: 0; z-index: 3;
}
.th-label { font-weight: 600; }
.th-kind { font-size: 9px; color: #8a8f9b; text-transform: uppercase; }
</style>

