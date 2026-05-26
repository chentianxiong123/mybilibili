<script setup>
defineProps({
  video: { type: Object, required: true },
  showStatistics: { type: Boolean, default: true }
})

const formatNum = (n) => {
  if (!n) return '0'
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
</script>

<template>
  <router-link :to="'/m/video/' + video.aId" class="video-item">
    <div class="pic">
      <img :src="video.pic" :alt="video.title" loading="lazy" />
      <span v-if="video.duration" class="duration">{{ video.duration }}</span>
    </div>
    <div class="info">
      <p class="title">{{ video.title }}</p>
      <div v-if="showStatistics" class="stats">
        <span class="play">{{ formatNum(video.play) }}次播放</span>
        <span class="danmaku">{{ formatNum(video.videoReview) }}弹幕</span>
      </div>
    </div>
  </router-link>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.video-item {
  display: block;
  margin-bottom: 12px;

  .pic {
    position: relative;
    width: 100%;
    padding-top: 56.25%;
    border-radius: 6px;
    overflow: hidden;
    background: #e3e5e7;

    img {
      position: absolute;
      inset: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .duration {
      position: absolute;
      bottom: 4px;
      right: 4px;
      background: rgba(0,0,0,0.7);
      color: #fff;
      font-size: 11px;
      padding: 1px 6px;
      border-radius: 4px;
    }
  }

  .info {
    padding: 8px 4px;
  }

  .title {
    font-size: 14px;
    color: $text-primary;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .stats {
    margin-top: 6px;
    font-size: 12px;
    color: $text-secondary;
    display: flex;
    gap: 12px;
  }
}
</style>