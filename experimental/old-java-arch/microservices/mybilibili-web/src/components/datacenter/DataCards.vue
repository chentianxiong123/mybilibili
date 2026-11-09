<template>
  <div class="data-cards">
    <div class="data-card" v-for="card in cards" :key="card.key" :class="{ 'card-highlight': card.highlight }">
      <div class="card-icon" :style="{ background: card.bgColor }">
        <component :is="card.icon" v-if="card.icon" />
      </div>
      <div class="card-info">
        <div class="card-label">{{ card.label }}</div>
        <div class="card-value">
          <span class="value-main">{{ formatNum(card.value) }}</span>
          <span v-if="card.increase !== undefined" class="value-increase" :class="card.increase >= 0 ? 'up' : 'down'">
            {{ card.increase >= 0 ? '+' : '' }}{{ formatNum(card.increase) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  cards: { type: Array, required: true }
})

function formatNum(num) {
  if (num == null) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + 'w'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'k'
  return String(num)
}
</script>

<style scoped>
.data-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
}
.data-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  transition: box-shadow 0.2s;
}
.data-card:hover {
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
}
.card-highlight {
  border-color: #00aeec20;
  background: #f0f9ff;
}
.card-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.card-info {
  flex: 1;
  min-width: 0;
}
.card-label {
  font-size: 13px;
  color: #999;
  margin-bottom: 2px;
}
.card-value {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.value-main {
  font-size: 20px;
  font-weight: 700;
  color: #333;
}
.value-increase {
  font-size: 12px;
}
.value-increase.up {
  color: #52c41a;
}
.value-increase.down {
  color: #ff4d4f;
}
</style>
