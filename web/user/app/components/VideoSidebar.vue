<script setup lang="ts">
import { ref } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'
import { ElSkeleton } from 'element-plus'

const props = defineProps<{
  danmuList: any[]
  loadingDanmus: boolean
  manuscriptInfo: any
  currentVideoIndex: number
  relatedVideos: any[]
  loadingRelatedVideos: boolean
  isDanmuListCollapsed: boolean
  isVideoPartsCollapsed: boolean
}>()

const emit = defineEmits(['toggleDanmuList', 'toggleVideoParts', 'jumpToDanmuTime', 'switchVideoPart', 'goToVideo', 'goToAuthor'])

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const danmuScrollRef = ref<HTMLElement | null>(null)
</script>

<template>
  <div class="right-section">
    <!-- 弹幕列表 -->
    <div class="side-danmu-list">
      <div class="danmu-list-header" @click="emit('toggleDanmuList')">
        <h3>弹幕列表</h3>
        <el-icon class="collapse-icon" :class="{ 'is-collapsed': isDanmuListCollapsed }">
          <ArrowDown />
        </el-icon>
      </div>
      <div ref="danmuScrollRef" class="danmu-items" :class="{ 'is-hidden': isDanmuListCollapsed }">
        <div v-if="loadingDanmus" class="loading-danmus">
          <el-skeleton :rows="5" animated />
        </div>
        <div v-else-if="danmuList.length === 0" class="no-danmus">
          <p>暂无弹幕</p>
        </div>
        <template v-else>
          <div class="danmu-header">
            <span class="header-time">时间</span>
            <span class="header-content">弹幕内容</span>
            <span class="header-send-time">发送时间</span>
          </div>
          <div class="danmu-list-content">
            <div
              v-for="(danmaku, index) in danmuList"
              :key="index"
              class="danmu-item"
              @click="emit('jumpToDanmuTime', danmaku.time)"
            >
              <span class="danmu-time">{{ formatTime(danmaku.time) }}</span>
              <span class="danmu-text">{{ danmaku.text }}</span>
              <span class="danmu-send-time">{{ danmaku.sendTime }}</span>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 视频分P列表 -->
    <div v-if="manuscriptInfo.videos && manuscriptInfo.videos.length > 1" class="video-parts-section">
      <div class="video-parts-header" @click="emit('toggleVideoParts')">
        <h3>视频分P</h3>
        <span class="video-parts-count">共{{ manuscriptInfo.videos.length }}P</span>
        <el-icon class="collapse-icon" :class="{ 'is-collapsed': isVideoPartsCollapsed }">
          <ArrowDown />
        </el-icon>
      </div>
      <div class="video-parts-list" :class="{ 'is-hidden': isVideoPartsCollapsed }">
        <div
          v-for="(part, index) in manuscriptInfo.videos"
          :key="part.id"
          :class="['video-part-item', { active: currentVideoIndex === index }]"
          @click="emit('switchVideoPart', Number(index))"
        >
          <span class="part-index">P{{ Number(index) + 1 }}</span>
          <span class="part-title">{{ part.title }}</span>
          <span class="part-duration">{{ part.duration }}</span>
        </div>
      </div>
    </div>

    <!-- 推荐视频 -->
    <div class="related-videos">
      <h3>推荐视频</h3>
      <div v-if="loadingRelatedVideos" class="loading-related">
        <el-skeleton :rows="4" animated />
      </div>
      <div v-else>
        <div v-for="video in relatedVideos" :key="video.id" class="related-video-item">
          <div class="video-cover">
            <a :href="video.manuscriptId ? '/manuscript/' + video.manuscriptId : '/manuscript/' + video.id" class="video-cover-link">
              <img loading="lazy" decoding="async" :src="video.cover || '/assets/placeholder-cover.svg'" alt="视频封面">
            </a>
            <span class="video-duration">{{ video.duration }}</span>
          </div>
          <div class="video-info">
            <h4 class="video-title">
              <span class="video-title-text" @click="emit('goToVideo', video)">{{ video.title }}</span>
            </h4>
            <div class="video-meta">
              <span class="video-author" @click="emit('goToAuthor', video.authorId)">{{ video.author }}</span>
              <div class="video-stats">
                <span class="video-view">{{ (video.viewCount || 0).toLocaleString() }}次播放</span>
                <span class="video-comment">{{ video.commentCount || 0 }}条评论</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!loadingRelatedVideos && relatedVideos.length === 0" class="no-related">
        <p>暂无相关视频推荐</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.right-section {
  width: 350px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.side-danmu-list {
  width: 100%;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.side-danmu-list .danmu-list-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  background-color: #f9f9f9;
}
.side-danmu-list .danmu-list-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
}
.side-danmu-list .danmu-items {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}
.side-danmu-list .danmu-items.is-hidden {
  display: none;
}
.collapse-icon {
  transition: transform 0.3s ease;
  font-size: 16px;
  color: #666;
}
.collapse-icon.is-collapsed {
  transform: rotate(-90deg);
}
.danmu-items {
  max-height: 400px;
}
.danmu-list-content {
  position: relative;
  width: 100%;
}
.danmu-header {
  display: flex;
  gap: 10px;
  padding: 8px 16px;
  background-color: #f5f5f5;
  font-size: 13px;
  font-weight: 600;
  color: #666;
  position: sticky;
  top: 0;
  z-index: 10;
}
.danmu-header .header-time {
  min-width: 50px;
}
.danmu-header .header-content {
  flex: 1;
}
.danmu-header .header-send-time {
  min-width: 140px;
  text-align: right;
}
.danmu-item {
  display: flex;
  gap: 10px;
  font-size: 14px;
  padding: 8px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.danmu-item:hover {
  background-color: #f0f0f0;
}
.danmu-time {
  color: #999;
  min-width: 50px;
  font-weight: 500;
}
.danmu-text {
  color: #333;
  flex: 1;
}
.danmu-send-time {
  color: #999;
  min-width: 140px;
  font-size: 12px;
  text-align: right;
}
.loading-danmus, .no-danmus {
  padding: 20px 0;
  text-align: center;
  color: #999;
  font-size: 14px;
}
.video-parts-section {
  width: 100%;
  background-color: #fff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
}
.video-parts-section .video-parts-header {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  background-color: #f9f9f9;
}
.video-parts-section .video-parts-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
}
.video-parts-section .video-parts-count {
  font-size: 12px;
  color: #999;
  margin-right: auto;
  margin-left: 8px;
}
.video-parts-section .video-parts-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 8px;
}
.video-parts-section .video-parts-list.is-hidden {
  display: none;
}
.video-parts-section .video-part-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  gap: 10px;
}
.video-parts-section .video-part-item:hover {
  background-color: #f5f5f5;
}
.video-parts-section .video-part-item.active {
  background-color: #e3f2fd;
  color: #00a1d6;
}
.video-parts-section .video-part-item .part-index {
  font-size: 12px;
  color: #999;
  min-width: 30px;
}
.video-parts-section .video-part-item .part-title {
  flex: 1;
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.video-parts-section .video-part-item .part-duration {
  font-size: 12px;
  color: #999;
}
.related-videos {
  background-color: #fff;
  padding: 20px 0;
}
.related-videos h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 15px;
  color: #333;
}
.loading-related {
  padding: 20px 0;
}
.no-related {
  text-align: center;
  padding: 40px 0;
  color: #999;
  font-size: 14px;
}
.related-video-item {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  align-items: flex-start;
}
.related-video-item:last-child {
  margin-bottom: 0;
}
.related-video-item .video-cover {
  position: relative;
  width: 160px;
  height: 90px;
  flex-shrink: 0;
  overflow: hidden;
  border-radius: 6px;
}
.related-video-item .video-cover-link {
  text-decoration: none;
  color: inherit;
  display: inline-block;
}
.related-video-item .video-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.related-video-item .video-cover:hover img {
  transform: scale(1.05);
  transition: transform 0.3s;
}
.related-video-item .video-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}
.related-video-item .video-info h4 {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
  pointer-events: none;
}
.related-video-item .video-title-text {
  cursor: pointer;
  transition: color 0.3s;
  pointer-events: auto;
}
.related-video-item .video-title-text:hover {
  color: #00aeec;
}
.related-video-item .video-meta {
  font-size: 12px;
  color: #999;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.related-video-item .video-author {
  font-size: 12px;
  color: #999;
  cursor: pointer;
  transition: color 0.3s;
}
.related-video-item .video-author:hover {
  color: #00aeec;
}
.related-video-item .video-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #999;
}
.related-video-item .video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}
</style>