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
      <div v-if="showStatistics" class="stats-overlay">
        <span class="play">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21"></polygon></svg>
          {{ formatNum(video.play) }}
        </span>
        <span class="danmaku">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>
          {{ formatNum(video.videoReview) }}
        </span>
        <span v-if="video.duration" class="duration">{{ video.duration }}</span>
      </div>
      <div v-else-if="video.duration" class="stats-overlay" style="justify-content: flex-end;">
        <span class="duration">{{ video.duration }}</span>
      </div>
    </div>
    <div class="info">
      <p class="title">{{ video.title }}</p>
      <div v-if="video.author" class="author-row">
        <span class="up-badge">UP</span>
        <span class="author-name">{{ video.author }}</span>
      </div>
    </div>
  </router-link>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.video-item {
  display: block;
  background: #fff;
  border-radius: 6px;
  overflow: hidden;

  .pic {
    position: relative;
    width: 100%;
    padding-top: 56.25%; /* 16:9比例 */
    background: #f3f3f3;

    img {
      position: absolute;
      inset: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .stats-overlay {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 16px 6px 6px;
      background: linear-gradient(to top, rgba(0,0,0,0.85), transparent);
      color: #fff;
      font-size: 11px;
      display: flex;
      align-items: center;

      .play, .danmaku {
        display: flex;
        align-items: center;
        gap: 4px;
        margin-right: 12px;
      }

      svg {
        width: 14px;
        height: 14px;
      }

      .duration {
        margin-left: auto;
      }
    }
  }

  .info {
    padding: 8px 8px 9px;
  }

  .title {
    font-size: 14px;
    color: #18191c;
    line-height: 1.4;
    height: 39px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    margin-bottom: 8px;
    font-weight: 500;
  }

  .author-row {
    display: flex;
    align-items: center;
    min-width: 0;
    height: 16px;
    color: #9499a0;
    font-size: 12px;
    line-height: 16px;

    .up-badge {
      flex: 0 0 auto;
      margin-right: 4px;
      padding: 0 2px;
      border: 1px solid #c9ccd0;
      border-radius: 3px;
      color: #9499a0;
      font-size: 9px;
      line-height: 12px;
      transform: scale(0.92);
      transform-origin: left center;
    }

    .author-name {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}
</style>
