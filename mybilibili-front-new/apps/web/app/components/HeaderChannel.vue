<script setup lang="ts">
interface ChannelItem {
  id?: string | number
  name: string
  icon?: string
  color?: string
  children?: Array<{ id?: string | number; name: string }>
}

withDefaults(defineProps<{
  channels?: ChannelItem[]
  iconCards?: Array<{ name: string; icon: string; color?: string; href?: string }>
}>(), {
  channels: () => [],
  iconCards: () => [
    { name: '动态', icon: 'icon-fengche', color: 'var(--v-brand-pink, #FF6699)', href: '/dynamic' },
    { name: '热门', icon: 'icon-huo', color: 'var(--v-operate-orange, #FF7F24)', href: '/hot' }
  ]
})

const onCardClick = (href?: string) => {
  if (href && typeof window !== 'undefined') {
    window.location.href = href
  }
}
</script>

<template>
  <div class="v-channel-bar">
    <div class="v-channel-icons">
      <div
        v-for="card in iconCards"
        :key="card.name"
        class="v-channel-icon-item"
        @click="onCardClick(card.href)"
      >
        <div class="v-channel-icon-bg" :style="{ '--icon-bg': card.color }">
          <i :class="['iconfont', card.icon]"></i>
        </div>
        <span class="v-channel-icon-title">{{ card.name }}</span>
      </div>
    </div>

    <div class="v-channel-items">
      <div
        v-for="ch in channels"
        :key="ch.id ?? ch.name"
        class="v-channel-link"
      >
        <span>{{ ch.name }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.v-channel-bar {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 12px 0;
  border-bottom: 1px solid var(--v-line-light, #F1F2F3);
}

.v-channel-icons {
  display: flex;
  align-items: center;
  gap: 20px;
}

.v-channel-icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.v-channel-icon-item:hover {
  transform: translateY(-2px);
}

.v-channel-icon-bg {
  width: 40px;
  height: 40px;
  border-radius: var(--v-radius-md, 8px);
  background: var(--icon-bg, var(--v-brand-pink, #FF6699));
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--v-bg1, #FFFFFF);
  font-size: 20px;
  transition: filter 0.2s ease;
}

.v-channel-icon-item:hover .v-channel-icon-bg {
  filter: brightness(1.1);
}

.v-channel-icon-bg .iconfont {
  font-size: 20px;
  color: var(--v-bg1, #FFFFFF);
}

.v-channel-icon-title {
  font-size: 12px;
  color: var(--v-text1, #18191C);
  transition: color 0.2s ease;
}

.v-channel-icon-item:hover .v-channel-icon-title {
  color: var(--brand-pink, #FF6699);
}

.v-channel-items {
  display: flex;
  align-items: center;
  gap: 18px;
  flex: 1;
  overflow-x: auto;
}

.v-channel-link {
  font-size: 14px;
  color: var(--v-text1, #18191C);
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.2s ease;
}

.v-channel-link:hover {
  color: var(--v-brand-pink, #FF6699);
}
</style>
