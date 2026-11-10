<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, View, ChatDotRound } from '@element-plus/icons-vue'
import { userApi } from '@/api/client'

const props = defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  allVideos: {
    type: Array,
    default: () => []
  },
  videoSortOption: {
    type: String,
    default: '最新发布'
  },
  loading: {
    type: Object,
    required: true
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['sort-change'])

const router = useRouter()

const sortOptions = ['最新发布', '最多播放', '最多收藏']

// 置顶视频相关数据
const pinnedVideo = ref(null)
const showPinnedVideoDialog = ref(false)
const pinnedVideoSelection = ref(null)

// 封面缺省图
const getDefaultCover = () => '/favicon.ico'

// 格式化播放量 / 时长 / 日期
const formatView = (n) => {
  const v = Number(n) || 0
  if (v >= 100000000) return (v / 100000000).toFixed(1).replace(/\.0$/, '') + '亿'
  if (v >= 10000) return (v / 10000).toFixed(1).replace(/\.0$/, '') + '万'
  return String(v)
}
const formatDuration = (d) => {
  if (!d) return ''
  if (typeof d === 'string' && d.includes(':')) return d
  const s = Number(d)
  if (!Number.isFinite(s)) return String(d)
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${String(sec).padStart(2, '0')}`
}
const formatDate = (t) => {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return String(t)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

// 加载置顶视频
const loadPinnedVideo = async () => {
  if (!props.userId) return
  try {
    const response = await userApi.getPinnedVideo()
    if (response.code === 200 && response.data) {
      pinnedVideo.value = response.data
    } else {
      pinnedVideo.value = null
    }
  } catch (error) {
    console.warn('获取置顶视频失败（非关键功能）:', error)
    pinnedVideo.value = null
  }
}

// 打开置顶视频选择对话框
const openPinnedVideoDialog = () => {
  if (props.allVideos.length === 0) {
    ElMessage.warning('暂无可选的视频')
    return
  }
  pinnedVideoSelection.value = pinnedVideo.value ? { ...pinnedVideo.value } : null
  showPinnedVideoDialog.value = true
}

// 保存置顶视频
const savePinnedVideo = async () => {
  if (!pinnedVideoSelection.value) {
    ElMessage.warning('请选择一个视频作为置顶视频')
    return
  }
  try {
    const response = await userApi.setPinnedVideo(pinnedVideoSelection.value.id)
    if (response.code === 200) {
      pinnedVideo.value = pinnedVideoSelection.value
      showPinnedVideoDialog.value = false
      ElMessage.success('置顶视频设置成功')
    } else {
      ElMessage.error(response.message || '设置置顶视频失败')
    }
  } catch (error) {
    console.error('设置置顶视频失败:', error)
    ElMessage.error('设置置顶视频失败，请稍后重试')
  }
}

// 取消置顶视频
const removePinnedVideo = async () => {
  try {
    const response = await userApi.removePinnedVideo()
    if (response.code === 200) {
      pinnedVideo.value = null
      showPinnedVideoDialog.value = false
      ElMessage.success('已取消置顶视频')
    } else {
      ElMessage.error(response.message || '取消置顶失败')
    }
  } catch (error) {
    console.error('取消置顶视频失败:', error)
    ElMessage.error('取消置顶失败，请稍后重试')
  }
}

// 处理排序变化
const handleSortChange = (option) => {
  emit('sort-change', option)
}

// 播放TA的视频
const playAllVideos = () => {
  if (props.allVideos.length > 0) {
    router.push(`/manuscript/${props.allVideos[0].id}`)
  } else {
    ElMessage.info('暂无视频')
  }
}

// 跳转到投稿页面
const goToSubmissions = () => {
  router.push(`/profile/${props.userId}/submissions`)
}

const openVideo = (video) => {
  if (video && video.id) router.push(`/manuscript/${video.id}`)
}

onMounted(() => {
  loadPinnedVideo()
})

watch(() => props.userId, () => {
  loadPinnedVideo()
})
</script>

<template>
  <div class="space-home-videos">
    <!-- 置顶视频（站点自研功能，B 站 space 主页无此区，仅样式对齐 section-wrap 头部） -->
    <section class="video-section pinned-section">
      <div class="section-wrap__header">
        <div class="section-wrap__title">置顶视频</div>
        <div class="btns" v-if="isOwnSpace">
          <button class="vui_button" @click="openPinnedVideoDialog">设置置顶</button>
        </div>
      </div>
      <div class="section-wrap__content">
        <div v-if="!pinnedVideo" class="empty-state">暂无置顶视频</div>
        <div v-else class="pinned-video-card" @click="openVideo(pinnedVideo)">
          <div class="pinned-video-cover">
            <img loading="lazy" decoding="async" :src="pinnedVideo.coverUrl || getDefaultCover()" :alt="pinnedVideo.title" />
            <span class="pinned-video-duration">{{ pinnedVideo.duration }}</span>
            <div class="pinned-badge">置顶</div>
          </div>
          <div class="pinned-video-info">
            <p class="pinned-video-title">{{ pinnedVideo.title }}</p>
            <div class="pinned-video-stats">
              <span class="pinned-stat-item">
                <el-icon><View /></el-icon>
                {{ pinnedVideo.viewCount }}
              </span>
              <span class="pinned-stat-item">
                <el-icon><ChatDotRound /></el-icon>
                {{ pinnedVideo.commentCount || 0 }}
              </span>
              <span class="pinned-stat-item">{{ pinnedVideo.date }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 视频（对齐 B 站 space 主页 .video-section） -->
    <section class="video-section">
      <div class="section-wrap__header">
        <div class="section-wrap__title">视频</div>
        <div class="section-wrap__desc">·{{ allVideos.length }}</div>
        <div class="radio-filter">
          <span
            v-for="option in sortOptions"
            :key="option"
            :class="['radio-filter__item', { active: videoSortOption === option }]"
            @click="handleSortChange(option)"
          >{{ option }}</span>
        </div>
        <div class="btns">
          <button class="vui_button playall-btn" @click="playAllVideos">播放全部</button>
          <button class="vui_button" @click="goToSubmissions">查看更多</button>
        </div>
      </div>
      <div class="section-wrap__content">
        <div v-if="loading.videos" class="empty-state">加载中...</div>
        <div v-else-if="allVideos.length === 0" class="empty-state">暂无视频</div>
        <div v-else class="video-list">
          <div
            v-for="video in allVideos.slice(0, 6)"
            :key="video.id"
            class="list-video-item"
            @click="openVideo(video)"
          >
            <div class="bili-video-card">
              <div class="bili-video-card__wrap">
                <div class="bili-video-card__cover">
                  <a class="bili-cover-card">
                    <div class="bili-cover-card__thumbnail">
                      <img :src="video.coverUrl || getDefaultCover()" :alt="video.title" loading="lazy" decoding="async" />
                    </div>
                    <div class="bili-cover-card__stats">
                      <div class="bili-cover-card__stat">
                        <i class="iconfont icon-bofangfangkuai"></i>
                        <span>{{ formatView(video.viewCount) }}</span>
                      </div>
                      <div class="bili-cover-card__stat bili-cover-card__stat--duration">
                        <span>{{ formatDuration(video.duration) }}</span>
                      </div>
                    </div>
                  </a>
                </div>
                <div class="bili-video-card__details">
                  <div class="bili-video-card__title" :title="video.title">
                    <a>{{ video.title }}</a>
                  </div>
                  <div class="bili-video-card__subtitle">
                    <span>{{ formatDate(video.uploadTime) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 置顶视频选择对话框 -->
    <el-dialog
      v-model="showPinnedVideoDialog"
      title="设置置顶视频"
      width="700px"
    >
      <div class="pinned-video-selection">
        <div v-if="allVideos.length === 0" class="empty-state">
          <p>暂无可选的视频</p>
        </div>
        <div v-else class="video-selection-grid">
          <div
            v-for="video in allVideos"
            :key="video.id"
            :class="['video-selection-card', { 'is-selected': pinnedVideoSelection?.id === video.id }]"
            @click="pinnedVideoSelection = video"
          >
            <div class="video-selection-cover">
              <img loading="lazy" decoding="async" :src="video.coverUrl" :alt="video.title">
              <span class="video-selection-duration">{{ video.duration }}</span>
              <div v-if="pinnedVideoSelection?.id === video.id" class="selected-badge">
                <el-icon><Check /></el-icon>
              </div>
            </div>
            <div class="video-selection-title">{{ video.title }}</div>
            <div class="video-selection-stats">
              <span>{{ video.viewCount }}播放</span>
              <span>{{ video.date }}</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button v-if="pinnedVideo" @click="removePinnedVideo">取消置顶</el-button>
          <el-button @click="showPinnedVideoDialog = false">取消</el-button>
          <el-button type="primary" @click="savePinnedVideo">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* ===== 对齐 space.bilibili.com/{uid} 主页 tab 的 .video-section =====
   参考值来自真实 B 站 getComputedStyle 抓取（1920×1080 / dpr1）：
   section 1289@x140 透明无白卡 padding 0 0 24 margin 0 0 24
   header 1289x34 flex mb16 / title「视频」24-600-24 #18191c
   desc「·N」16-400-16 #61666d ml6 / radio-filter 药丸 86x34 pad0 15 r6
     激活 #00AEEC 白字 / 未激活 #f6f7f8 #61666d, gap12 ml36
   btns 右对齐 播放全部/查看更多 100x34 r8 白底 border #e3e5e7 fs14 #18191c
   content grid 5 列 x245 column-gap16 row-gap20
   card 245x204 cover 245x138 r6 #e3e5e7 / details mt10 title 13-500-18 #18191c
     subtitle 12-400-16 #9499a0 mt4 */
.space-home-videos {
  width: 100%;
}

.video-section {
  width: 1289px;
  max-width: 100%;
  padding: 0 0 24px;
  margin: 0 0 24px;
}

/* header */
.section-wrap__header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.section-wrap__title {
  font-size: 24px;
  font-weight: 600;
  line-height: 24px;
  color: #18191c;
}

.section-wrap__desc {
  margin-left: 6px;
  font-size: 16px;
  font-weight: 400;
  line-height: 16px;
  color: #61666d;
}

/* 排序药丸（B站 radio-filter） */
.radio-filter {
  display: flex;
  gap: 12px;
  margin-left: 36px;
}

.radio-filter__item {
  display: flex;
  align-items: center;
  height: 34px;
  padding: 0 15px;
  border-radius: 6px;
  background-color: #f6f7f8;
  color: #61666d;
  font-size: 14px;
  font-weight: 400;
  line-height: 14px;
  cursor: pointer;
}

.radio-filter__item.active {
  background-color: #00aeec;
  color: #fff;
}

/* 右侧按钮组 */
.btns {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.vui_button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 100px;
  height: 34px;
  padding: 0 12px;
  border-radius: 8px;
  background-color: #fff;
  border: 1px solid #e3e5e7;
  color: #18191c;
  font-size: 14px;
  font-weight: 400;
  line-height: 14px;
  white-space: nowrap;
  cursor: pointer;
}

.vui_button + .vui_button {
  margin-left: 12px;
}

.vui_button:hover {
  color: #00aeec;
  border-color: #00aeec;
}

/* 网格（5 列 x245 gap16/20） */
.video-list {
  display: grid;
  grid-template-columns: repeat(5, 245px);
  column-gap: 16px;
  row-gap: 20px;
  justify-content: start;
}

.list-video-item {
  cursor: pointer;
}

.bili-video-card,
.bili-video-card__wrap {
  width: 100%;
}

/* 封面 245x138 r6 底 #e3e5e7 */
.bili-video-card__cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 6px;
  overflow: hidden;
  background-color: #e3e5e7;
}

.bili-cover-card {
  display: block;
  width: 100%;
  height: 100%;
  text-decoration: none;
}

.bili-cover-card__thumbnail {
  width: 100%;
  height: 100%;
}

.bili-cover-card__thumbnail img {
  display: inline-block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
}

/* 封面底部统计（白字 12-12，左播放右时长） */
.bili-cover-card__stats {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
  padding: 0 8px 4px 8px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.6) 100%);
  color: #fff;
}

.bili-cover-card__stat {
  display: flex;
  align-items: center;
  font-size: 12px;
  font-weight: 400;
  line-height: 12px;
  color: #fff;
}

.bili-cover-card__stat .iconfont {
  font-size: 12px;
  margin-right: 2px;
}

/* 卡片详情（mt10 + 标题 + 日期） */
.bili-video-card__details {
  margin-top: 10px;
}

.bili-video-card__title {
  display: flow-root;
  max-height: 36px;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bili-video-card__title a {
  color: #18191c;
  font-size: 13px;
  font-weight: 500;
  line-height: 18px;
  text-decoration: none;
}

.list-video-item:hover .bili-video-card__title a {
  color: #00aeec;
}

.bili-video-card__subtitle {
  display: flex;
  margin-top: 4px;
  font-size: 12px;
  font-weight: 400;
  line-height: 16px;
  color: #9499a0;
}

.bili-video-card__subtitle span {
  color: #9499a0;
}

/* 空态 */
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}

/* ===== 置顶视频（自研功能，B 站无同构参考，仅保持与 section 头部一致） ===== */
.pinned-video-card {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  padding: 16px;
  border-radius: 6px;
  background-color: #fff;
  cursor: pointer;
}

.pinned-video-cover {
  position: relative;
  width: 240px;
  flex-shrink: 0;
  aspect-ratio: 16 / 9;
  border-radius: 6px;
  overflow: hidden;
  background-color: #e3e5e7;
}

.pinned-video-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.pinned-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.pinned-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  background: #00a1d6;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.pinned-video-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.pinned-video-title {
  font-size: 18px;
  font-weight: 600;
  color: #18191c;
  margin: 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pinned-video-stats {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: #9499a0;
}

.pinned-stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 置顶视频选择对话框 */
.pinned-video-selection {
  max-height: 500px;
  overflow-y: auto;
}

.video-selection-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.video-selection-card {
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: 8px;
  overflow: hidden;
}

.video-selection-card:hover {
  border-color: #00a1d6;
}

.video-selection-card.is-selected {
  border-color: #00a1d6;
  background-color: #f0f9ff;
}

.video-selection-cover {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
}

.video-selection-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-selection-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.selected-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  background: #00a1d6;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-selection-title {
  padding: 8px;
  font-size: 14px;
  color: #18191c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.video-selection-stats {
  padding: 0 8px 8px;
  font-size: 12px;
  color: #9499a0;
  display: flex;
  justify-content: space-between;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
