<script setup>
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
    return 'https://picsum.photos/id/1025/400/225'
  }
  return videos[0].coverUrl || videos[0].cover || 'https://picsum.photos/id/1025/400/225'
}

const allFolders = computed(() => [...props.collections, ...props.myFavorites])

const currentFolderCount = computed(() => {
  const folder = allFolders.value.find(c => c.name === props.activeCategory)
  return folder ? folder.count : 0
})


</script>

<template>
  <div class="favorites-content">
    <div class="favorite-header">
      <div class="favorite-header-cover">
        <img loading="lazy" decoding="async" :src="getFavoriteFolderCover(videos)" :alt="activeCategory" class="favorite-header-img">
      </div>
      <div class="favorite-header-info">
        <div class="favorite-header-title">{{ activeCategory }}</div>
        <div class="favorite-header-meta">
          <span class="meta-item">创建者：{{ userInfo.username }}</span>
          <span class="meta-item">{{ currentFolderCount }}个内容</span>
          <span class="meta-item">公开</span>
        </div>
        <div class="favorite-header-actions">
          <button class="action-btn play-all-btn" @click="emit('play-all')">
            <span class="play-icon">▶</span>
            播放全部视频
          </button>
        </div>
      </div>
    </div>

    <div class="sort-filter">
      <div class="batch-operations" v-if="!batchMode">
        <el-button size="small" @click="emit('toggle-batch')">批量操作</el-button>
      </div>
      <div class="batch-operations active" v-else>
        <el-checkbox
          :indeterminate="selectedFavorites.size > 0 && selectedFavorites.size < videos.length"
          :model-value="selectedFavorites.size === videos.length"
          @change="emit('select-all')"
        >
          全选
        </el-checkbox>
        <span class="batch-selected-count">已选 {{ selectedFavorites.size }} 项</span>
        <el-button
          size="small"
          type="danger"
          :disabled="selectedFavorites.size === 0"
          @click="emit('batch-delete')"
        >
          批量删除
        </el-button>
        <el-button size="small" @click="emit('toggle-batch')">取消</el-button>
      </div>
      <div class="sort-options">
        <select :value="activeSort" class="sort-select" @change="emit('update:activeSort', $event.target.value)">
          <option v-for="option in sortOptions" :key="option" :value="option">{{ option }}</option>
        </select>
      </div>
      <div class="search-box">
        <input type="text"
          :value="searchKeyword"
          @input="emit('update:searchKeyword', $event.target.value)"
          placeholder="输入关键词"
          class="search-input"
        >
        >
        <button class="search-btn">
          <el-icon><Search /></el-icon>
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <p>加载中...</p>
    </div>
    <div v-else-if="filteredFavorites.length === 0" class="empty-state">
      <p>{{ searchKeyword ? '没有匹配的视频' : '暂无收藏' }}</p>
    </div>
    <div v-else class="videos-grid">
      <div
        v-for="video in filteredFavorites"
        :key="video.id"
        :class="['video-item', { 'video-item-selected': selectedFavorites.has(video.id), 'batch-mode': batchMode }]"
        @click="batchMode ? emit('toggle-select', video.id) : emit('navigate', video.id)"
      >
        <div v-if="batchMode" class="video-checkbox" @click.stop>
          <el-checkbox
            :model-value="selectedFavorites.has(video.id)"
            @change="emit('toggle-select', video.id)"
          />
        </div>
        <div class="video-cover">
          <img loading="lazy" decoding="async" :src="video.coverUrl || video.cover" :alt="video.title" class="video-cover-img">
          <div class="video-duration">{{ video.duration }}</div>
        </div>
        <div class="video-title">{{ video.title }}</div>
        <div class="video-meta">
          <span class="video-views">{{ video.viewCount }}</span>
          <span class="video-date">{{ video.date }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.favorite-header {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background-color: #fff;
  border-radius: 8px;
}

.favorite-header-cover {
  width: 160px;
  height: 100px;
  border-radius: 4px;
  overflow: hidden;
  flex-shrink: 0;
  background-color: #f0f0f0;
}

.favorite-header-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.favorite-header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
}

.favorite-header-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.favorite-header-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #9499a0;
  margin-bottom: 12px;
}

.meta-item {
  position: relative;
}

.meta-item:not(:last-child)::after {
  content: '·';
  position: absolute;
  right: -10px;
}

.favorite-header-actions {
  display: flex;
  gap: 10px;
}

.favorite-header-actions .play-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #00aeec;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.favorite-header-actions .play-all-btn:hover {
  background-color: #0095d9;
}

.play-icon {
  font-size: 12px;
}

.favorites-content {
  flex: 1;
}

.sort-filter {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  font-size: 14px;
  color: #666;
}

.batch-operations {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-operations.active {
  cursor: default;
}

.batch-operations:hover {
  color: #00aeec;
}

.batch-selected-count {
  font-size: 13px;
  color: #666;
  margin-left: 4px;
}

.sort-options {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sort-select {
  padding: 4px 8px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 14px;
  color: #666;
}

.search-box {
  display: flex;
  align-items: center;
  position: relative;
  width: 200px;
}

.search-box .search-input {
  width: 100%;
  padding: 6px 32px 6px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 16px;
  font-size: 14px;
  box-sizing: border-box;
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
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-box .search-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: #333;
}

.videos-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

.video-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  position: relative;
  transition: opacity 0.2s;
}

.video-item.batch-mode {
  cursor: pointer;
}

.video-item.batch-mode .video-cover {
  opacity: 0.9;
}

.video-item-selected {
  border-radius: 8px;
  outline: 2px solid #00aeec;
  outline-offset: 2px;
}

.video-checkbox {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 10;
}

.video-checkbox :deep(.el-checkbox__inner) {
  background-color: #fff;
  border-color: #00aeec;
}

.video-checkbox :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: #00aeec;
  border-color: #00aeec;
}

.video-cover {
  position: relative;
  width: 100%;
  padding-bottom: 56.25%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f0f0f0;
}

.video-cover-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.video-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.video-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #999;
}

.video-views {
  font-size: 12px;
  color: #999;
  margin: 0;
}

.video-date {
  font-size: 12px;
  color: #999;
}

.loading-state,
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}
</style>