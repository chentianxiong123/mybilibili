<script setup lang="ts">
interface CardUploader {
  id?: string | number
  name?: string
  avatar?: string
}

interface CardVideo {
  id?: string | number
  manuscriptId?: string | number
  title?: string
  coverUrl?: string
  viewCount?: number | string
  commentCount?: number | string
  duration?: string
  dateText?: string
  sourceType?: string
  videoCount?: number
  uploader?: CardUploader
}

const props = withDefaults(defineProps<{
  video: CardVideo
  showVideoCount?: boolean
}>(), {
  showVideoCount: false
})

const manuscriptHref = (v: CardVideo) => {
  const id = v.manuscriptId ?? v.id
  return id ? `/manuscript/${id}` : '#'
}

const viewText = (v: number | string | undefined) => {
  if (v === undefined || v === null) return '0'
  const n = typeof v === 'number' ? v : Number(v)
  if (!Number.isFinite(n)) return '0'
  if (n >= 10000) {
    const wan = n / 10000
    return wan >= 10 ? `${Math.floor(wan)}万` : `${wan.toFixed(1)}万`
  }
  return n.toLocaleString()
}

const commentText = (v: number | string | undefined) => {
  if (v === undefined || v === null) return '0'
  const n = typeof v === 'number' ? v : Number(v)
  if (!Number.isFinite(n)) return '0'
  return n.toLocaleString()
}
</script>

<template>
  <div class="video-item">
    <div class="video-cover">
      <a :href="manuscriptHref(props.video)" class="video-cover-link">
        <img
          loading="lazy"
          decoding="async"
          :src="props.video.coverUrl || '/assets/placeholder-cover.svg'"
          :alt="props.video.title || '视频封面'"
        >
      </a>
      <span v-if="props.video.sourceType === 'bilibili'" class="source-badge">B站</span>
      <div class="video-stats-overlay">
        <span class="stat-item">
          <el-icon><View /></el-icon>
          {{ viewText(props.video.viewCount) }}
        </span>
        <span class="stat-item">
          <el-icon><Star /></el-icon>
          {{ commentText(props.video.commentCount) }}
        </span>
      </div>
      <span class="video-duration">{{ props.video.duration || '00:00' }}</span>
      <span
        v-if="props.showVideoCount && (props.video.videoCount ?? 0) > 1"
        class="video-count"
      >{{ props.video.videoCount }}P</span>
    </div>
    <span class="video-title">
      <a :href="manuscriptHref(props.video)" class="video-title-text">{{ props.video.title }}</a>
    </span>
    <div class="video-meta">
      <a
        v-if="props.video.uploader?.id"
        :href="`/user/${props.video.uploader.id}`"
        class="video-author"
      >{{ props.video.uploader?.name || '未知UP主' }}</a>
      <span v-else class="video-author">{{ props.video.uploader?.name || '未知UP主' }}</span>
      <span class="video-separator"> · </span>
      <span class="video-date">{{ props.video.dateText }}</span>
    </div>
  </div>
</template>

<style scoped>
.video-item {
  position: relative;
  display: flex;
  flex-direction: column;
  cursor: pointer;
}

.video-cover {
  position: relative;
  width: 100%;
  padding-top: 56.25%;
  border-radius: var(--v-radius, 6px);
  overflow: hidden;
  background: var(--v-bg2, #F6F7F8);
}

.video-cover-link {
  position: absolute;
  inset: 0;
  display: block;
}

.video-cover img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.video-item:hover .video-cover img {
  transform: scale(1.08);
}

.source-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.6);
  color: #fff;
  font-size: 11px;
  padding: 1px 5px;
  border-radius: 3px;
  z-index: 2;
}

.video-stats-overlay {
  position: absolute;
  bottom: 6px;
  left: 6px;
  display: flex;
  gap: 8px;
  color: #fff;
  font-size: 12px;
  z-index: 2;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
}

.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.video-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.6);
  color: #fff;
  font-size: 12px;
  padding: 1px 5px;
  border-radius: 3px;
  z-index: 2;
}

.video-count {
  position: absolute;
  top: 6px;
  left: 6px;
  background: var(--v-brand-pink, #FF6699);
  color: #fff;
  font-size: 11px;
  padding: 1px 5px;
  border-radius: 3px;
  z-index: 2;
}

.video-title {
  margin-top: 8px;
  display: block;
  font-size: 14px;
  line-height: 1.4;
  color: var(--v-text1, #18191C);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.video-title-text {
  color: inherit;
  text-decoration: none;
  transition: color 0.2s ease;
}

.video-title-text:hover {
  color: var(--v-brand-pink, #FF6699);
}

.video-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--v-text3, #9499A0);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-author {
  color: inherit;
  text-decoration: none;
  transition: color 0.2s ease;
}

.video-author:hover {
  color: var(--v-brand-pink, #FF6699);
}

.video-separator {
  margin: 0 4px;
}
</style>
