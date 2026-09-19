<template>
  <div class="favorite-dropdown">
    <div class="dropdown-content">
      <!-- 收藏夹列表 -->
      <div class="folder-list">
        <div v-if="favoriteFolders.length > 0" v-for="folder in favoriteFolders" :key="folder.id" :class="['folder-item', { active: activeFolderId === folder.id }]" @click="handleFolderClick(folder)">
          <span class="folder-name">{{ folder.name }}</span>
          <span class="folder-count">{{ folder.count }}</span>
        </div>
        <div v-else class="empty-state">
          <p>暂无收藏夹</p>
        </div>
      </div>

      <!-- 视频列表 -->
      <div class="video-list">
        <div v-if="loading" class="loading-state">
          <el-icon class="loading-icon"><Loading /></el-icon>
          <p>加载中...</p>
        </div>
        <div v-else-if="favoriteVideos.length > 0" v-for="video in favoriteVideos" :key="video.id" class="video-item" @click="handleVideoClick(video)">
          <div class="thumbnail-wrapper">
            <img loading="lazy" decoding="async" :src="video.thumbnail" class="video-thumbnail" />
            <span class="video-duration">{{ video.duration }}</span>
          </div>
          <div class="video-info">
            <div class="video-title">{{ video.title }}</div>
            <div class="video-uploader">{{ video.uploader }}</div>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>{{ activeFolderId ? '暂无收藏视频' : '请选择收藏夹' }}</p>
        </div>
      </div>
    </div>

    <!-- 底部按钮 -->
    <div class="dropdown-footer">
      <span class="view-all-btn" @click="handleViewAll">查看全部</span>
      <el-button type="primary" class="play-all-btn" :disabled="favoriteVideos.length === 0" @click="handlePlayAll">
        <el-icon><VideoPlay /></el-icon>
        播放全部
      </el-button>
    </div>
  </div>
</template>

<script lang="ts">
import { collectionApi } from '@/api/collection'
import { getCurrentUserId } from '@/utils/auth'

export default {
  name: 'FavoriteDropdown',
  data() {
    return {
      activeFolderId: null as number | null,
      favoriteFolders: [] as any[],
      favoriteVideos: [] as any[],
      loading: false
    }
  },
  watch: {
    activeFolderId(newId: number | null) {
      if (newId) {
        this.fetchFavoriteVideos(newId)
      }
    }
  },
  mounted() {
    this.fetchFavoriteFolders()
  },
  methods: {
    async fetchFavoriteFolders() {
      const userId = getCurrentUserId()
      if (!userId) return
      try {
        const response = await collectionApi.getUserCollections(userId)
        const raw = response?.data ?? response
        const data = raw?.list || raw
        if (Array.isArray(data)) {
          this.favoriteFolders = data.map((folder: any) => ({
            id: folder.id,
            name: folder.title || folder.name,
            count: folder.manuscript_count || folder.manuscriptCount || 0
          }))
          if (this.favoriteFolders.length > 0 && !this.activeFolderId) {
            this.activeFolderId = this.favoriteFolders[0].id
          }
        }
      } catch (error) {
        console.error('获取收藏夹列表失败:', error)
      }
    },
    async fetchFavoriteVideos(folderId: number) {
      if (!folderId) {
        this.favoriteVideos = []
        return
      }

      this.loading = true
      try {
        const response = await collectionApi.getCollectionManuscripts(folderId, 1, 5)
        const raw = response?.data ?? response
        const list = Array.isArray(raw) ? raw : raw?.list || []
        const storeNickname = this.$store?.state?.user?.nickname || this.$store?.state?.user?.username || '我'
        this.favoriteVideos = list.map((item: any) => ({
          id: item.id || item.manuscriptId,
          title: item.title || '未知标题',
          thumbnail: item.cover_url || item.coverUrl || item.cover || '/assets/placeholder-cover.svg',
          duration: item.duration || '00:00',
          uploader: storeNickname
        }))
      } catch (error) {
        console.error('获取收藏夹视频失败:', error)
        this.favoriteVideos = []
      } finally {
        this.loading = false
      }
    },
    handleFolderClick(folder: any) {
      this.activeFolderId = folder.id
    },
    handleViewAll() {
      window.location.href = '/profile/favorites'
    },
    handleVideoClick(video: any) {
      window.location.href = `/manuscript/${video.id}`
    },
    handlePlayAll() {
      if (this.favoriteVideos.length > 0) {
        window.location.href = `/manuscript/${this.favoriteVideos[0].id}`
      }
    }
  }
}
</script>

<style scoped>
.favorite-dropdown {
  width: 600px;
  max-height: 500px;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 8px;
}

.dropdown-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.folder-list {
  width: 180px;
  background-color: #f5f7fa;
  overflow-y: auto;
  flex-shrink: 0;
}

.folder-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.folder-item:hover {
  background-color: rgba(0, 161, 214, 0.1);
}

.folder-item.active {
  background-color: #00a1d6;
  color: #fff;
}

.folder-name {
  font-size: 14px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.folder-count {
  font-size: 12px;
  color: #9499a0;
  flex-shrink: 0;
  margin-left: 8px;
}

.folder-item.active .folder-count {
  color: #fff;
}

.video-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.video-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.video-item:hover {
  background-color: #f5f7fa;
}

.thumbnail-wrapper {
  position: relative;
  flex-shrink: 0;
}

.video-thumbnail {
  width: 160px;
  height: 90px;
  object-fit: cover;
  border-radius: 6px;
}

.video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 2px;
}

.video-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.video-title {
  font-size: 14px;
  color: #18191c;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.video-uploader {
  font-size: 12px;
  color: #9499a0;
}

.dropdown-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-top: 1px solid #e3e5e7;
}

.view-all-btn {
  color: #00a1d6;
  font-size: 14px;
  cursor: pointer;
}

.view-all-btn:hover {
  color: #0091c6;
}

.play-all-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 16px;
  background-color: #00a1d6;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}

.play-all-btn:hover {
  background-color: #0091c6;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #9499a0;
}

.empty-state p {
  font-size: 14px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #9499a0;
}

.loading-icon {
  font-size: 32px;
  margin-bottom: 8px;
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-state p {
  font-size: 14px;
}
</style>
