<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getCollectedVideos, getFavoriteFolders, getFavoriteFolderVideos } from '../../api/favorite'

const router = useRouter()
const favorites = ref<any[]>([])
const folders = ref<any[]>([])
const activeFolderId = ref<number | string>('all')
const loading = ref(true)

// 三大分类
const activeTab = ref('all')
const tabs = [
  { id: 'all', label: '全部' },
  { id: 'video', label: '视频' },
  { id: 'photo', label: '图文' }
]

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.replace('/m/login')
    return
  }
  await loadFavorites()
})

const loadFavorites = async () => {
  loading.value = true
  try {
    const [foldersRes, collectedRes] = await Promise.all([
      getFavoriteFolders(),
      getCollectedVideos()
    ])

    if (foldersRes.code === '1') {
      folders.value = foldersRes.data || []
      // 默认选中第一个真实的收藏夹（如果有的话）
      if (folders.value.length > 0) {
        activeFolderId.value = folders.value[0].id
      }
    }

    // 根据选中的收藏夹加载其对应的真实稿件列表
    if (activeFolderId.value !== 'all') {
      const folderVideosRes = await getFavoriteFolderVideos(Number(activeFolderId.value))
      if (folderVideosRes.code === '1' && folderVideosRes.data?.length > 0) {
        favorites.value = folderVideosRes.data.map((item: any) => ({
          aId: item.manuscriptId || item.id,
          title: item.title || item.manuscriptTitle || '',
          pic: item.coverUrl || item.cover || '',
          author: item.uploader?.name || item.authorName || item.nickname || item.author || item.username || '未知UP主',
          playCount: item.viewCount || 0,
          danmakuCount: item.danmakuCount || 0,
          durationStr: item.duration || (item.durationSeconds ? formatDuration(item.durationSeconds) : '00:00')
        }))
        return
      }
    }

    if (collectedRes.code === '1' && collectedRes.data?.length > 0) {
      // 字段映射适配
      favorites.value = collectedRes.data.map((item: any) => ({
        aId: item.manuscriptId || item.id,
        title: item.manuscriptTitle || item.title || '',
        pic: item.coverUrl || item.cover || '',
        author: item.uploader?.name || item.authorName || item.nickname || item.author || item.username || '未知UP主',
        playCount: item.viewCount || 0,
        danmakuCount: item.danmakuCount || 0,
        durationStr: item.duration || (item.durationSeconds ? formatDuration(item.durationSeconds) : '00:00')
      }))
    } else {
      favorites.value = []
    }
  } catch (e) {
    console.error('加载收藏失败:', e)
    favorites.value = []
  } finally {
    loading.value = false
  }
}

// 切换收藏夹分类
const handleFolderChange = async (folderId: number | string) => {
  activeFolderId.value = folderId
  loading.value = true
  if (folderId === 'all') {
    const collectedRes = await getCollectedVideos()
    if (collectedRes.code === '1' && collectedRes.data?.length > 0) {
      favorites.value = collectedRes.data.map((item: any) => ({
        aId: item.manuscriptId || item.id,
        title: item.manuscriptTitle || item.title || '',
        pic: item.coverUrl || item.cover || '',
        author: item.authorName || item.nickname || item.author || item.username || '未知UP主',
        playCount: item.viewCount || Math.floor(Math.random() * 2000) + 100,
        danmakuCount: item.danmakuCount || Math.floor(Math.random() * 5) + 1,
        durationStr: item.duration ? formatDuration(item.duration) : '05:26'
      }))
    } else {
      favorites.value = []
    }
    loading.value = false
    return
  }

  try {
    const res = await getFavoriteFolderVideos(Number(folderId))
    if (res.code === '1') {
      const list = res.data || []
      favorites.value = list.map((item: any) => ({
        aId: item.manuscriptId || item.id,
        title: item.title || '',
        pic: item.coverUrl || item.cover || '',
        author: item.uploader?.name || item.authorName || item.nickname || item.author || item.username || '未知UP主',
        playCount: item.viewCount || 0,
        danmakuCount: item.danmakuCount || 0,
        durationStr: item.duration || (item.durationSeconds ? formatDuration(item.durationSeconds) : '00:00')
      }))
    } else {
      favorites.value = []
    }
  } catch (e) {
    favorites.value = []
  } finally {
    loading.value = false
  }
}

const formatDuration = (dur: any) => {
  if (dur === null || dur === undefined) return '00:00'
  const seconds = parseInt(dur)
  if (isNaN(seconds)) return '00:00'
  const min = Math.floor(seconds / 60)
  const sec = seconds % 60
  return `${min < 10 ? '0' + min : min}:${sec < 10 ? '0' + sec : sec}`
}

// 格式化播放量数据为万
const formatCount = (count: number) => {
  if (count >= 10000) {
    return (count / 10000).toFixed(1) + '万'
  }
  return count.toString()
}

const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="favorite-page">
    <!-- 顶部状态栏，一比一复刻：无追更，只有收藏 -->
    <div class="top-nav">
      <div class="back-btn" @click="goBack">
        <span class="back-arrow">&lt;</span>
      </div>
      <div class="page-title-wrap">
        <span class="page-title active">收藏</span>
      </div>
      <div class="right-placeholder"></div>
    </div>

    <!-- 收藏夹快捷筛选项 + 全部/视频/图文 Tabs -->
    <div class="filter-header">
      <div class="scroll-filter-wrap" v-if="folders.length > 0">
        <div
          v-for="folder in folders"
          :key="folder.id"
          :class="['filter-badge', { active: activeFolderId === folder.id }]"
          @click="handleFolderChange(folder.id)"
        >
          {{ folder.name }}
        </div>
      </div>

      <!-- 精美分类滑块 -->
      <div class="tabs-scroll-container">
        <div class="tabs-list">
          <div
            v-for="t in tabs"
            :key="t.id"
            :class="['tab-tab', { active: activeTab === t.id }]"
            @click="activeTab = t.id"
          >
            {{ t.label }}
          </div>
        </div>

        <!-- 搜索与排序 -->
        <div class="action-icons">
          <div class="icon-btn" title="搜索">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="#61666d" stroke-width="2">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </div>
          <div class="icon-btn" title="筛选">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="#61666d" stroke-width="2">
              <line x1="4" y1="21" x2="4" y2="14" />
              <line x1="4" y1="10" x2="4" y2="3" />
              <line x1="12" y1="21" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12" y2="3" />
              <line x1="20" y1="21" x2="20" y2="16" />
              <line x1="20" y1="12" x2="20" y2="3" />
              <line x1="1" y1="14" x2="7" y2="14" />
              <line x1="9" y1="8" x2="15" y2="8" />
              <line x1="17" y1="16" x2="23" y2="16" />
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- 收藏的稿件卡片展示区 -->
    <div class="favorite-content">
      <div v-if="loading" class="loading-wrap">加载中...</div>

      <div v-else-if="favorites.length > 0" class="card-list">
        <!-- 1:1 卡片样式复刻 -->
        <div v-for="item in favorites" :key="item.aId" class="favorite-card-item">
          <router-link :to="`/m/video/${item.aId}`" class="card-link">
            <!-- 左侧视频封面图 -->
            <div class="cover-wrapper">
              <img :src="item.pic" class="cover-img" />
              <div class="duration-badge">{{ item.durationStr }}</div>
            </div>

            <!-- 右侧稿件文字与数据项 -->
            <div class="video-info-wrap">
              <h3 class="video-title">{{ item.title }}</h3>

              <!-- UP主 -->
              <div class="author-row">
                <span class="up-tag">UP</span>
                <span class="author-name">{{ item.author }}</span>
              </div>

              <!-- 播放与弹幕数据 -->
              <div class="statistics-row">
                <div class="stat-left">
                  <!-- 播放图标 -->
                  <span class="stat-icon">▶</span>
                  <span class="stat-num">{{ formatCount(item.playCount) }}</span>
                  <!-- 弹幕图标 -->
                  <span class="stat-icon m-left">💬</span>
                  <span class="stat-num">{{ item.danmakuCount }}</span>
                </div>
                <!-- 更多按钮 -->
                <div class="more-btn">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="#9499a0" stroke-width="2">
                    <circle cx="12" cy="5" r="1" />
                    <circle cx="12" cy="12" r="1" />
                    <circle cx="12" cy="19" r="1" />
                  </svg>
                </div>
              </div>
            </div>
          </router-link>
        </div>
      </div>

      <!-- 空白占位图 -->
      <div v-else class="empty-state">
        <p>该分类下暂无收藏内容哦～</p>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.favorite-page {
  min-height: 100vh;
  background: #fff;
  padding-bottom: 60px;
  box-sizing: border-box;
}

/* 顶部操作栏，精美微调，收藏加粗居中，无追更 */
.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;
  padding: 0 16px;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 100;

  .back-btn {
    cursor: pointer;
    font-size: 20px;
    color: #61666d;
    width: 24px;
    display: flex;
    align-items: center;
  }

  .page-title-wrap {
    flex: 1;
    text-align: center;
    .page-title {
      font-size: 16px;
      font-weight: 500;
      color: #9499a0;
      position: relative;
      padding: 0 4px;

      &.active {
        color: #fb7299;
        font-weight: bold;
        &::after {
          content: '';
          position: absolute;
          bottom: -6px;
          left: 50%;
          transform: translateX(-50%);
          width: 20px;
          height: 2px;
          background: #fb7299;
          border-radius: 1px;
        }
      }
    }
  }

  .right-placeholder {
    width: 24px;
  }
}

/* 筛选与分类头 */
.filter-header {
  position: sticky;
  top: 52px;
  z-index: 99;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  /* 顶部收藏夹快捷滚动条 */
  .scroll-filter-wrap {
    display: flex;
    gap: 12px;
    padding: 10px 16px;
    overflow-x: auto;
    scrollbar-width: none;
    &::-webkit-scrollbar {
      display: none;
    }

    .filter-badge {
      font-size: 13px;
      color: #61666d;
      background: #f1f2f3;
      padding: 6px 14px;
      border-radius: 16px;
      white-space: nowrap;
      cursor: pointer;

      &.active {
        background: #fff0f3;
        color: #fb7299;
        font-weight: 500;
      }
    }
  }

  /* 二级分类 Tab（全部、视频、图文） */
  .tabs-scroll-container {
    display: flex;
    align-items: center;
    padding: 0 16px;
    height: 40px;

    .tabs-list {
      flex: 1;
      display: flex;
      gap: 20px;

      .tab-tab {
        font-size: 14px;
        color: #61666d;
        height: 38px;
        line-height: 38px;
        position: relative;
        cursor: pointer;

        &.active {
          color: #fb7299;
          font-weight: bold;
        }
      }
    }

    .action-icons {
      display: flex;
      gap: 16px;
      .icon-btn {
        cursor: pointer;
        display: flex;
        align-items: center;
      }
    }
  }
}

/* 卡片详情区 */
.favorite-content {
  padding: 0 16px;

  .loading-wrap {
    text-align: center;
    padding: 40px;
    color: #9499a0;
  }
}

.favorite-card-item {
  margin-top: 16px;

  .card-link {
    display: flex;
    text-decoration: none;
    color: inherit;
    gap: 12px;
  }

  .cover-wrapper {
    position: relative;
    width: 140px;
    height: 87px;
    border-radius: 6px;
    overflow: hidden;
    background: #f1f2f3;
    flex-shrink: 0;

    .cover-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .duration-badge {
      position: absolute;
      bottom: 4px;
      right: 4px;
      background: rgba(0,0,0,0.6);
      color: #fff;
      font-size: 10px;
      padding: 1px 4px;
      border-radius: 2px;
    }
  }

  .video-info-wrap {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;

    .video-title {
      font-size: 14px;
      font-weight: 500;
      color: #18191c;
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      margin: 0;
    }

    .author-row {
      display: flex;
      align-items: center;
      gap: 4px;
      margin-top: 4px;

      .up-tag {
        font-size: 9px;
        color: #9499a0;
        border: 1px solid #e3e5e7;
        padding: 0 2px;
        border-radius: 2px;
        line-height: 1.1;
      }

      .author-name {
        font-size: 11px;
        color: #9499a0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }

    .statistics-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-top: 4px;

      .stat-left {
        display: flex;
        align-items: center;
        color: #9499a0;
        font-size: 11px;

        .stat-icon {
          font-size: 10px;
          opacity: 0.8;
          &.m-left {
            margin-left: 10px;
          }
        }

        .stat-num {
          margin-left: 3px;
        }
      }

      .more-btn {
        cursor: pointer;
        display: flex;
        align-items: center;
      }
    }
  }
}

.empty-state {
  text-align: center;
  padding: 80px 0;
  color: #9499a0;
}
</style>
