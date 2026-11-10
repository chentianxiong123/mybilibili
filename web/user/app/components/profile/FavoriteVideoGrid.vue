<script setup lang="ts">
import { computed } from 'vue'
import { Search } from '@element-plus/icons-vue'

const props = defineProps({
  videos: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  batchMode: {
    type: Boolean,
    default: false
  },
  selectedFavorites: {
    type: Set,
    default: () => new Set()
  },
  activeCategory: {
    type: String,
    required: true
  },
  activeSort: {
    type: String,
    required: true
  },
  sortOptions: {
    type: Array,
    required: true
  },
  searchKeyword: {
    type: String,
    default: ''
  },
  userInfo: {
    type: Object,
    required: true
  },
  collections: {
    type: Array,
    required: true
  },
  myFavorites: {
    type: Array,
    required: true
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  },
  activeFolderId: {
    type: [String, Number],
    default: null
  }
})

const emit = defineEmits([
  'toggle-batch',
  'toggle-select',
  'select-all',
  'batch-delete',
  'play-all',
  'update:activeSort',
  'update:searchKeyword',
  'navigate'
])

const filteredFavorites = computed(() => {
  const keyword = (props.searchKeyword || '').trim().toLowerCase()
  if (!keyword) return props.videos
  return props.videos.filter(v => (v.title || '').toLowerCase().includes(keyword))
})

const getFavoriteFolderCover = (videos) => {
  if (!videos || videos.length === 0) {
    return '/assets/site/img/bilibili/video-placeholder.png'
  }
  return videos[0].coverUrl || videos[0].cover || '/assets/site/img/bilibili/video-placeholder.png'
}

const allFolders = computed(() => [...props.collections, ...props.myFavorites])

const currentFolder = computed(() => {
  return allFolders.value.find(c => Number(c.id) === Number(props.activeFolderId))
})

const currentFolderCount = computed(() => {
  return currentFolder.value ? currentFolder.value.count : (props.videos?.length || 0)
})

// 播放量合计（对齐 B站 favlist-info-detail__data）
const totalViews = computed(() => {
  return (props.videos || []).reduce((sum, v) => sum + (Number(v.viewCount) || 0), 0)
})

// 格式化收藏时间：x天前 / 今天 / x小时前
const formatCollectTime = (t) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return ''
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`
  return date.toLocaleDateString()
}

// 格式化播放数
const formatViews = (n) => {
  if (!n) return 0
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return n
}

// 卡片副标题：UP主名 · 收藏于x前（对齐 B站 bili-video-card__subtitle）
const cardSubtitle = (video) => {
  const up = video.uploader?.nickname || video.uploader?.username || ''
  const ct = formatCollectTime(video.collectTime || video.collect_time)
  return [up, ct ? `收藏于${ct}` : ''].filter(Boolean).join(' · ')
}
</script>

<template>
  <div class="favlist-main">
    <!-- 收藏夹信息头 -->
    <div class="favlist-info">
      <div class="favlist-info-file-card">
        <div class="folder-cover-card__cover">
          <img v-if="videos.length" :src="getFavoriteFolderCover(videos)" alt="">
        </div>
      </div>
      <div class="favlist-info-detail">
        <div class="favlist-info-detail__title">
          <div class="favlist-info-detail__title-row">
            <span class="vui_ellipsis">{{ activeCategory }}</span>
          </div>
        </div>
        <div class="favlist-info-detail__data">
          <span>视频数: {{ currentFolderCount }}</span>
          <span class="data-sep">播放量: {{ formatViews(totalViews) }}</span>
        </div>
        <div class="favlist-info-detail__actions">
          <button class="vui_button action-btn blue" @click="emit('play-all')">
            <svg class="action-icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="14" height="14"><path d="M847.52 429.80266667L344.50133334 101.152c-78.176-45.184-142.21866667-8.18133333-142.21866667 82.13333333v657.22666667c0 90.368 64.04266667 127.30666667 142.21866667 82.19733333L847.52000001 594.13333333c78.25066667-45.25866667 78.25066667-119.14666667 0-164.33066666z m0 0" fill="#FFFFFF"></path></svg>
            <span>播放全部</span>
          </button>
          <button v-if="isOwnSpace && !batchMode" class="vui_button favlist-info-batch" @click="emit('toggle-batch')">
            <span>批量操作</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 排序 radio-filter + 右侧 搜索/批量栏 -->
    <div class="fav-list-header">
      <div class="fav-list-header-filter">
        <div class="radio-filter">
          <div
            v-for="option in sortOptions"
            :key="option"
            class="radio-filter__item"
            :class="{'radio-filter__item--active': activeSort === option}"
            @click="emit('update:activeSort', option)"
          >{{ option }}</div>
        </div>
      </div>
      <div class="fav-list-header-filter__right" v-if="isOwnSpace">
        <div class="batch-bar" v-if="batchMode">
          <el-checkbox
            :indeterminate="selectedFavorites.size > 0 && selectedFavorites.size < videos.length"
            :model-value="selectedFavorites.size === videos.length"
            @change="emit('select-all')"
          >全选</el-checkbox>
          <span class="batch-count">已选 {{ selectedFavorites.size }} 项</span>
          <el-button size="small" type="danger" :disabled="selectedFavorites.size === 0" @click="emit('batch-delete')">批量删除</el-button>
          <button class="vui_button batch-cancel" @click="emit('toggle-batch')">取消</button>
        </div>
        <div class="vui_input fav-list-header-filter__search" v-else>
          <div class="vui_input-wrapper vui_input-wrapper--prepped">
            <input
              type="text"
              class="vui_input__input"
              :value="searchKeyword"
              @input="emit('update:searchKeyword', $event.target.value)"
              placeholder="在收藏夹内搜索"
            >
            <el-icon class="search-icon"><Search /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 视频内容 -->
    <div class="items" v-if="filteredFavorites.length > 0">
      <div
        class="items__item"
        v-for="(video, index) in filteredFavorites"
        :key="index"
        @click="emit('navigate', video.id)"
      >
        <div class="bili-video-card" :class="{'batch-selected': batchMode && selectedFavorites.has(video.id)}">
          <div class="bili-video-card__cover" @click.stop>
            <img v-if="video.coverUrl" loading="lazy" decoding="async" :src="video.coverUrl" alt="">
            <span v-if="video.duration" class="bili-video-card__duration">{{ video.duration }}</span>
            <div class="batch-checkbox" v-if="batchMode && isOwnSpace" @click.stop="emit('toggle-select', video.id)">
              <input type="checkbox" :checked="selectedFavorites.has(video.id)" readonly>
            </div>
          </div>
          <div class="bili-video-card__details">
            <div class="bili-video-card__title" :title="video.title">{{ video.title }}</div>
            <div class="bili-video-card__subtitle">{{ cardSubtitle(video) }}</div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="fav-empty">
      <img src="/assets/site/img/nothing.png" alt="">
      <p>{{ loading ? '加载中...' : (searchKeyword ? '没有匹配的视频' : '暂无收藏') }}</p>
    </div>
  </div>
</template>

<style scoped>
.favlist-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  margin: 30px 0 0;
  padding: 0 0 40px;
}

/* 收藏夹信息头 */
.favlist-info {
  display: flex;
  padding: 0 0 24px;
}

.favlist-info-file-card {
  width: 178px;
  flex-shrink: 0;
  padding: 12px 0 0;
}

.folder-cover-card__cover {
  width: 178px;
  height: 100px;
  border-radius: 6px;
  background-color: #e3e5e7;
  overflow: hidden;
}

.folder-cover-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.favlist-info-detail {
  flex: 1;
  min-width: 0;
  margin-left: 18px;
}

.favlist-info-detail__title {
  line-height: 28px;
}

.favlist-info-detail__title-row {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: 500;
  color: #18191c;
}

.favlist-info-detail__title-row .vui_ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.favlist-info-detail__data {
  margin-top: 4px;
  font-size: 14px;
  line-height: 17px;
  color: #9499a0;
}

.favlist-info-detail__data .data-sep {
  margin-left: 24px;
}

.favlist-info-detail__actions {
  display: flex;
  align-items: center;
  margin-top: 8px;
}

.vui_button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  font-size: 14px;
  line-height: 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: all .2s ease;
}

.action-btn.blue {
  padding: 0 15px;
  background-color: #00aeec;
  color: #fff;
  border: 1px solid #00aeec;
}

.action-btn.blue .action-icon {
  margin-right: 6px;
  flex-shrink: 0;
}

.action-btn.blue:hover {
  background-color: #00a9e6;
  border-color: #00a9e6;
}

.favlist-info-batch {
  padding: 0 12px;
  background-color: #fff;
  color: #18191c;
  border: 1px solid #e3e5e7;
  margin-left: 8px;
}

.favlist-info-batch:hover {
  color: #00aeec;
  border-color: #00aeec;
}

/* 排序 + 搜索 头 */
.fav-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 24px 0 0;
  height: 34px;
}

.radio-filter {
  display: flex;
  gap: 12px;
}

.radio-filter__item {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 15px;
  border-radius: 6px;
  font-size: 14px;
  line-height: 14px;
  background-color: #f6f7f8;
  color: #61666d;
  cursor: pointer;
  transition: all .2s ease;
}

.radio-filter__item:hover {
  color: #00aeec;
}

.radio-filter__item--active {
  background-color: #00aeec;
  color: #fff;
}

.radio-filter__item--active:hover {
  color: #fff;
}

.fav-list-header-filter__right {
  display: flex;
  align-items: center;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}

.batch-count {
  color: #61666d;
  font-size: 13px;
}

.batch-cancel {
  padding: 0 12px;
  background-color: #fff;
  color: #18191c;
  border: 1px solid #e3e5e7;
}

.vui_input {
  width: 211px;
}

.vui_input-wrapper--prepped {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
  height: 34px;
  padding: 0 10px;
  background-color: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 0 6px 6px 0;
}

.vui_input__input {
  flex: 1;
  min-width: 0;
  height: 32px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: #18191c;
  padding: 0;
}

.vui_input__input::placeholder {
  color: #9499a0;
}

.search-icon {
  color: #9499a0;
  font-size: 14px;
  margin-left: 6px;
}

/* 视频网格 */
.items {
  display: grid;
  grid-template-columns: repeat(6, 210px);
  gap: 20px 16px;
  margin: 24px 0 0;
}

.items__item {
  cursor: pointer;
}

.bili-video-card {
  display: block;
}

.bili-video-card__cover {
  position: relative;
  display: block;
  width: 210px;
  height: 118px;
  border-radius: 6px;
  background-color: #e3e5e7;
  overflow: hidden;
}

.bili-video-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.bili-video-card__duration {
  position: absolute;
  right: 0;
  bottom: 0;
  padding: 0 6px;
  background: rgba(0, 0, 0, .5);
  border-radius: 6px 0 0 0;
  color: #fff;
  font-size: 12px;
  line-height: 18px;
  z-index: 2;
}

.bili-video-card.batch-selected .bili-video-card__cover {
  outline: 2px solid #00aeec;
  outline-offset: 2px;
}

.batch-checkbox {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 10;
  cursor: pointer;
}

.batch-checkbox input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #00aeec;
}

.bili-video-card__details {
  display: flex;
  flex-direction: column;
  margin-top: 10px;
}

.bili-video-card__title {
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  color: #18191c;
  padding-right: 30px;
  min-height: 36px;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
}

.bili-video-card__title a,
.bili-video-card__title {
  text-decoration: none;
}

.bili-video-card__subtitle {
  margin-top: 4px;
  font-size: 12px;
  line-height: 16px;
  color: #9499a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fav-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #9499a0;
  font-size: 13px;
  padding: 60px 0;
}

.fav-empty img {
  width: 20%;
  max-width: 140px;
  margin-bottom: 12px;
}
</style>
