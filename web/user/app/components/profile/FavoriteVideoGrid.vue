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
    return '/assets/teriteri/img/bilibili/video-placeholder.png'
  }
  return videos[0].coverUrl || videos[0].cover || '/assets/teriteri/img/bilibili/video-placeholder.png'
}

const allFolders = computed(() => [...props.collections, ...props.myFavorites])

const currentFolder = computed(() => {
  return allFolders.value.find(c => Number(c.id) === Number(props.activeFolderId))
})

const currentFolderCount = computed(() => {
  return currentFolder.value ? currentFolder.value.count : (props.videos?.length || 0)
})

const currentFolderDesc = computed(() => {
  return currentFolder.value?.description || ''
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

const rule = computed(() => {
  const idx = props.sortOptions.indexOf(props.activeSort)
  return idx + 1
})
</script>

<template>
  <div class="fav-main">
    <!-- 收藏夹信息头 -->
    <div class="favList-info">
      <div class="favInfo-box">
        <a class="favInfo-cover">
          <img v-if="videos.length" :src="getFavoriteFolderCover(videos)" alt="">
          <div class="cover-icon" v-else>
            <svg t="1711977711975" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4395" width="16" height="16"><path d="M865.49525114 335.25237508H158.50474886V246.87856197h706.99050228v88.37381311z m-88.37381311-265.12143802H246.87856197v88.3738118h530.24287606V70.13093706z m176.74762491 441.86906294v353.49525114c0 48.6055974-39.7682157 88.3738131-88.3738118 88.3738118H158.50474886c-48.6055974 0-88.3738131-39.7682157-88.3738118-88.3738118v-353.49525114c0-48.6055974 39.7682157-88.3738131 88.3738118-88.3738131h706.99050228c48.6055974 0 88.3738131 39.7682157 88.3738118 88.3738131z m-265.12143802 176.74762492l-265.12143802-145.8167909v287.21489161l265.12143802-141.39810071z" p-id="4396"></path></svg>
          </div>
        </a>
        <div class="favInfo-details">
          <a class="fav-name">{{ activeCategory }}</a>
          <div class="fav-meta" v-if="currentFolderDesc">
            <span class="fav-desc" :title="currentFolderDesc">{{ currentFolderDesc }}</span>
          </div>
          <div class="fav-meta">
            <span>{{ currentFolderCount }}个内容</span>
            <span class="dot">·</span>
            <span>创建者：{{ userInfo?.nickname || userInfo?.username || '' }}</span>
            <span class="dot">·</span>
            <span>公开</span>
          </div>
          <div class="fav-options">
            <div class="play-all" @click="emit('play-all')">
              <svg t="1711992851151" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4395" width="14" height="14"><path d="M847.52000001 429.80266667L344.50133334 101.152c-78.176-45.184-142.21866667-8.18133333-142.21866667 82.13333333v657.22666667c0 90.368 64.04266667 127.30666667 142.21866667 82.19733333L847.52000001 594.13333333c78.25066667-45.25866667 78.25066667-119.14666667 0-164.33066666z m0 0" fill="#FFFFFF" p-id="4396"></path></svg>
              <span>播放全部</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 排序 tab + 批量操作 -->
    <div class="fav-header">
      <div class="fav-info">
        <div class="be-tab">
          <ul class="be-tab-inner">
            <li
              v-for="(option, index) in sortOptions"
              :key="option"
              class="be-tab-item"
              :class="{'active': activeSort === option}"
              @click="emit('update:activeSort', option)"
            >
              <span>{{ option }}</span>
            </li>
          </ul>
          <div class="be-tab-cursor" :style="`transform: translateX(${48 * (rule - 1)}px); width: 48px;`"></div>
        </div>
      </div>
      <div class="fav-actions" v-if="isOwnSpace">
        <div class="batch-operations" v-if="!batchMode">
          <span class="batch-toggle" @click="emit('toggle-batch')">批量操作</span>
        </div>
        <div class="batch-bar" v-else>
          <el-checkbox
            :indeterminate="selectedFavorites.size > 0 && selectedFavorites.size < videos.length"
            :model-value="selectedFavorites.size === videos.length"
            @change="emit('select-all')"
          >全选</el-checkbox>
          <span class="batch-count">已选 {{ selectedFavorites.size }} 项</span>
          <el-button size="small" type="danger" :disabled="selectedFavorites.size === 0" :loading="batchDeleting" @click="emit('batch-delete')">批量删除</el-button>
          <el-button size="small" @click="emit('toggle-batch')">取消</el-button>
        </div>
      </div>
    </div>

    <!-- 搜索框 -->
    <div class="fav-search" v-if="isOwnSpace">
      <div class="search-box">
        <input type="text"
          :value="searchKeyword"
          @input="emit('update:searchKeyword', $event.target.value)"
          placeholder="在收藏夹内搜索"
          class="search-input"
        >
        <button class="search-btn">
          <el-icon><Search /></el-icon>
        </button>
      </div>
    </div>

    <!-- 视频内容 -->
    <div class="fav-content" v-if="filteredFavorites.length > 0">
      <ul class="fav-video-list clearfix">
        <li
          class="small-item"
          :class="{'batch-selected': batchMode && selectedFavorites.has(video.id)}"
          v-for="(video, index) in filteredFavorites"
          :key="index"
        >
          <div class="batch-checkbox" v-if="batchMode && isOwnSpace" @click.stop="emit('toggle-select', video.id)">
            <input type="checkbox" :checked="selectedFavorites.has(video.id)" readonly>
          </div>
          <a :href="`/video/${video.id}`" target="_blank" class="cover" @click.stop>
            <img v-if="video.coverUrl" loading="lazy" decoding="async" :src="video.coverUrl" alt="">
            <span v-if="video.duration" class="length">{{ video.duration }}</span>
            <div class="meta-mask">
              <div class="meta-info">
                <p class="view">播放：{{ formatViews(video.viewCount) }}</p>
                <p class="favorite">收藏：{{ formatViews(video.collectCount) }}</p>
                <p class="author">UP主：{{ video.uploader?.nickname || video.uploader?.username || 'UP主' }}</p>
              </div>
            </div>
          </a>
          <a :href="`/video/${video.id}`" target="_blank" class="title" :title="video.title">{{ video.title }}</a>
          <div class="meta pubdate">收藏于：{{ formatCollectTime(video.collectTime || video.collect_time) }}</div>
        </li>
      </ul>
    </div>
    <div v-else class="fav-empty">
      <img src="/assets/teriteri/img/nothing.png" alt="">
      <p>{{ loading ? '加载中...' : (searchKeyword ? '没有匹配的视频' : '暂无收藏') }}</p>
    </div>
  </div>
</template>

<style scoped>
.fav-main {
  flex: 1;
  min-width: 0;
  border-left: 1px solid #e5e9ef;
  display: flex;
  flex-direction: column;
}

.favList-info {
  padding: 20px 20px;
  box-sizing: border-box;
  border-bottom: 1px solid #e5e9ef;
}

.favInfo-box {
  height: 119px;
  display: flex;
}

.favInfo-cover {
  background: url('/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
  background-size: cover;
  width: 190px;
  height: 119px;
  overflow: hidden;
  position: relative;
  border-radius: 4px;
  display: block;
  cursor: pointer;
  flex-shrink: 0;
}

.favInfo-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-icon {
  position: absolute;
  right: 5px;
  bottom: 5px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.5);
  fill: #fff;
  height: 20px;
  width: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.favInfo-details {
  margin-left: 18px;
  min-width: 0;
}

.fav-name {
  display: block;
  font-size: 16px;
  font-weight: 500;
  color: #212121;
  line-height: 17px;
  margin-bottom: 12px;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fav-meta {
  display: block;
  color: #99a2aa;
  font-size: 0;
  letter-spacing: 0;
  line-height: 16px;
  margin-bottom: 6px;
}

.fav-meta span {
  display: inline-block;
  font-size: 12px;
  vertical-align: middle;
}

.fav-desc {
  width: 100%;
  height: 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fav-meta .dot {
  margin: 0 8px;
}

.fav-options {
  display: flex;
  align-items: center;
  margin-top: 14px;
}

.play-all {
  height: 32px;
  width: 120px;
  line-height: 32px;
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  color: #fff;
  background-color: #FF7DA1;
  border-radius: 4px;
  transition: all .2s;
  cursor: pointer;
}

.play-all:hover {
  background-color: #fb7299;
}

.play-all svg {
  margin-left: 16px;
  margin-right: 6px;
}

.fav-header {
  margin: 20px 20px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.fav-info {
  height: 30px;
}

.be-tab {
  display: inline-block;
  line-height: 18px;
  vertical-align: middle;
  position: relative;
}

.be-tab-inner {
  list-style: none;
  margin: 0;
  padding: 0;
}

.be-tab-item {
  float: left;
  font-size: 14px;
  margin-right: 20px;
  line-height: 27px;
  vertical-align: top;
  cursor: pointer;
  padding: 0 0 3px;
  font-size: 12px;
  color: #666;
}

.be-tab-item.active {
  color: #fb7299;
}

.be-tab-cursor {
  left: 0;
  border-bottom: 1px solid #fb7299;
  transform: translateX(0);
  transition: width .2s ease, transform .2s ease;
  position: absolute;
  bottom: 0;
  height: 0;
  width: 0;
}

.be-tab-cursor:after {
  content: "";
  left: 50%;
  transform: translateX(-50%);
  border-bottom: 3px solid #fb7299;
  border-top: 0;
  border-left: 3px solid transparent;
  border-right: 3px solid transparent;
  position: absolute;
  bottom: 0;
  height: 0;
  width: 0;
}

.fav-actions {
  margin-right: 0;
}

.batch-toggle {
  font-size: 12px;
  color: #999;
  cursor: pointer;
  transition: color .2s;
}

.batch-toggle:hover {
  color: #fb7299;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.batch-count {
  color: #666;
  font-size: 12px;
}

.fav-search {
  padding: 0 20px;
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.search-box {
  display: flex;
  align-items: center;
  position: relative;
  width: 220px;
}

.search-box .search-input {
  width: 100%;
  padding: 6px 32px 6px 12px;
  border: 1px solid #e5e9ef;
  border-radius: 16px;
  font-size: 13px;
  box-sizing: border-box;
  outline: none;
  transition: border-color .2s;
}

.search-box .search-input:focus {
  border-color: #fb7299;
}

.search-box .search-btn {
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  padding: 4px;
  background-color: transparent;
  color: #9499a0;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fav-content {
  position: relative;
  font-size: 12px;
  margin-top: 20px;
}

.fav-video-list {
  margin: 0 20px;
  list-style: none;
  padding: 0;
}

.clearfix:after {
  content: " ";
  display: table;
  clear: both;
}

.small-item {
  display: block;
  float: left;
  width: 160px;
  position: relative;
  padding: 0;
  margin: 0 20px 20px 0;
  border-radius: 4px;
  box-sizing: border-box;
  cursor: pointer;
}

.small-item.batch-selected .cover {
  outline: 2px solid #fb7299;
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
  accent-color: #fb7299;
}

.small-item .cover {
  background: url('/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
  background-size: cover;
  border-radius: 4px;
  display: block;
  width: 160px;
  height: 100px;
  overflow: hidden;
  position: relative;
}

.small-item .cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.small-item .length {
  background: rgba(0, 0, 0, .5);
  border-radius: 5px 0 0 0;
  color: #fff;
  line-height: 20px;
  transition: opacity .2s ease;
  padding: 0 6px;
  position: absolute;
  right: 0;
  bottom: 0;
  opacity: 1;
}

.cover:hover .length {
  opacity: 0;
}

.meta-mask {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 4;
  width: 100%;
  height: 100%;
  opacity: 0;
  background-color: rgba(0, 0, 0, .45);
  color: #e5e9ef;
  transition: opacity .2s ease;
}

.cover:hover .meta-mask {
  opacity: 1;
}

.meta-info {
  position: absolute;
  top: 50%;
  left: 15px;
  transform: translateY(-50%);
}

.meta-info p {
  line-height: 1.7em;
  margin: 0;
}

.meta-info .author {
  width: 125px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.small-item .title {
  display: block;
  line-height: 20px;
  height: 38px;
  margin-top: 6px;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
  line-break: anywhere;
  font-size: 12px;
  color: #212121;
  text-decoration: none;
}

.small-item .title:hover {
  color: #fb7299;
}

.small-item .meta {
  color: #999;
  white-space: nowrap;
  font-size: 0;
  margin-top: 5px;
  height: 16px;
  line-height: 16px;
}

.meta.pubdate {
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fav-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #99a2aa;
  font-size: 13px;
  padding: 60px 0;
}

.fav-empty img {
  width: 20%;
  max-width: 140px;
  margin-bottom: 12px;
}
</style>
