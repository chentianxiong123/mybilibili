<script setup>
defineProps({
  live: { type: Object, required: true }
})
</script>

<template>
  <router-link :to="`/m/live/${live.roomId}`" class="live-info-item">
    <div class="cover-wrap">
      <img :src="live.cover || '/images/nocontent.png'" class="cover" />
      <div class="live-tag" v-if="live.isLive === 1">
        <span class="live-dot" />
        LIVE
      </div>
      <span class="viewer-count" v-if="live.onlineNum">
        {{ live.onlineNum }}人在看
      </span>
    </div>
    <div class="live-info">
      <p class="live-title">{{ live.title }}</p>
      <p class="up-name">{{ live.uname }}</p>
    </div>
  </router-link>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.live-info-item { display: block; }

.cover-wrap {
  position: relative;
  border-radius: 6px;
  overflow: hidden;
  background: #000;
}

.cover {
  width: 100%;
  padding-top: 56.25%;
  display: block;
  object-fit: cover;
}

.live-tag {
  position: absolute;
  top: 6px;
  left: 6px;
  background: $theme-pink;
  color: #fff;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 3px;
}

.live-dot {
  width: 5px;
  height: 5px;
  background: #fff;
  border-radius: 50%;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.viewer-count {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0,0,0,0.7);
  color: #fff;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
}

.live-info { padding: 6px 2px; }

.live-title {
  font-size: 13px;
  color: $text-primary;
  line-height: 1.3;
  margin-bottom: 3px;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.up-name { font-size: 12px; color: $text-secondary; }
</style>
