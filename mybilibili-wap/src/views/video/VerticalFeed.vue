<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getVideoInfo, getRecommendVides, getComments } from '../../api/video'
import { likeManuscript, collectManuscript, shareManuscript, followUser } from '../../api/interaction'
import { getToken } from '../../utils/session'

const route = useRoute()
const router = useRouter()
const startAId = parseInt(route.params.aId) || 0

const feed = ref([])
const activeIndex = ref(0)
const playing = ref(false)
const loading = ref(true)
const progress = ref(0)
const containerRef = ref(null)
const videoEls = ref([])

// 评论面板
const showComment = ref(false)
const comments = ref([])
const commentLoading = ref(false)
const isFullscreen = ref(false)
// 双击点赞心形动画
const heartShow = ref(false)
let heartTimer = null

const hlsMap = {}
let scrollRaf = 0

const fmt = (n) => {
  n = n || 0
  if (n >= 10000) return (Math.floor(n / 1000) / 10) + '万'
  return String(n)
}

async function loadDetail(aId) {
  const res = await getVideoInfo(aId)
  return res.code === '1' && res.data ? res.data : null
}

function makeItem(basic) {
  return {
    aId: basic.aId,
    title: basic.title || '',
    pic: basic.pic || '',
    author: basic.author || '',
    mid: basic.mid || 0,
    play: basic.play || 0,
    commentCount: basic.videoReview || 0,
    playUrl: basic.playUrl || '',
    playUrlHd: basic.playUrlHd || '',
    playUrlSd: basic.playUrlSd || '',
    playUrlLd: basic.playUrlLd || '',
    description: basic.description || '',
    isVertical: basic.isVertical ? 1 : 0,
    uploader: basic.uploader || null,
    likeCount: basic.likeCount || 0,
    liked: false,
    starred: false,
    following: false,
    resolved: !!(basic.playUrl || basic.playUrlHd)
  }
}

async function ensureLoaded(i) {
  const it = feed.value[i]
  if (!it || it.resolved) return
  const d = await loadDetail(it.aId)
  if (d) {
    Object.assign(it, {
      title: d.title, pic: d.pic, author: d.author, mid: d.mid,
      play: d.play, commentCount: d.videoReview || 0,
      playUrl: d.playUrl, playUrlHd: d.playUrlHd,
      playUrlSd: d.playUrlSd, playUrlLd: d.playUrlLd,
      description: d.description, isVertical: d.isVertical,
      uploader: d.uploader, likeCount: d.likeCount,
      resolved: true
    })
  }
}

function destroyHls(i) {
  const it = feed.value[i]
  if (it && hlsMap[it.aId]) {
    try { hlsMap[it.aId].destroy() } catch (e) {}
    delete hlsMap[it.aId]
  }
}

function setupVideo(i) {
  const it = feed.value[i]
  const el = videoEls.value[i]
  if (!it || !el) return
  const url = it.playUrlHd || it.playUrlSd || it.playUrlLd || it.playUrl
  if (!url) return
  destroyHls(i)
  if (url.endsWith('.m3u8')) {
    import('hls.js').then(({ default: Hls }) => {
      if (Hls.isSupported()) {
        const hls = new Hls()
        hls.loadSource(url)
        hls.attachMedia(el)
        hlsMap[it.aId] = hls
      } else {
        el.src = url
      }
    })
  } else {
    el.src = url
  }
}

function playActive() {
  const el = videoEls.value[activeIndex.value]
  if (!el) return
  el.play().then(() => { playing.value = true }).catch(() => {})
}

function pauseAll(except) {
  videoEls.value.forEach((el, i) => {
    if (el && i !== except) el.pause()
  })
  playing.value = false
}

async function activate(i) {
  i = Math.max(0, Math.min(feed.value.length - 1, i))
  activeIndex.value = i
  await ensureLoaded(i)
  await nextTick()
  pauseAll(i)
  setupVideo(i)
  playActive()
}

function onScroll() {
  cancelAnimationFrame(scrollRaf)
  scrollRaf = requestAnimationFrame(() => {
    const c = containerRef.value
    if (!c) return
    const idx = Math.round(c.scrollTop / c.clientHeight)
    if (idx !== activeIndex.value) activate(idx)
  })
}

function onTimeUpdate(e) {
  const el = e.target
  if (el && el.duration) progress.value = el.currentTime / el.duration
}

function onEnded() {
  playing.value = false
  progress.value = 0
}

function togglePlay() {
  const el = videoEls.value[activeIndex.value]
  if (!el) return
  if (el.paused) {
    el.play().then(() => { playing.value = true }).catch(() => {})
  } else {
    el.pause()
    playing.value = false
  }
}

function showHeart() {
  heartShow.value = false
  nextTick(() => { heartShow.value = true })
  clearTimeout(heartTimer)
  heartTimer = setTimeout(() => { heartShow.value = false }, 900)
}

// 单击 播放/暂停；300ms 内双击 = 点赞 + 心形动画（抖音交互）
let lastTap = 0
let tapTimer = null
function onOverlayTap() {
  const now = Date.now()
  if (now - lastTap < 300) {
    clearTimeout(tapTimer)
    lastTap = 0
    const it = feed.value[activeIndex.value]
    if (it && !it.liked) toggleLike(it)
    showHeart()
    return
  }
  lastTap = now
  clearTimeout(tapTimer)
  tapTimer = setTimeout(() => { togglePlay() }, 300)
}

async function toggleLike(it) {
  it.liked = !it.liked
  it.likeCount = Math.max(0, it.likeCount + (it.liked ? 1 : -1))
  await likeManuscript(it.aId, it.liked)
}

async function toggleStar(it) {
  it.starred = !it.starred
  await collectManuscript(it.aId, it.starred)
}

async function doFollow(it) {
  if (!getToken()) { router.push('/m/login'); return }
  it.following = !it.following
  await followUser(it.mid, it.following)
}

async function doShare(it) {
  try {
    const abs = router.resolve(`/m/video/${it.aId}`).href
    if (navigator.share) {
      await navigator.share({ title: it.title, url: `${location.origin}${abs}` })
    }
  } catch (e) {}
  await shareManuscript(it.aId)
}

async function openComments(it) {
  showComment.value = true
  commentLoading.value = true
  comments.value = []
  try {
    const res = await getComments(it.aId, 1, 30)
    if (res.code === '1') comments.value = res.data || []
  } catch (e) {}
  commentLoading.value = false
}

const goBack = () => {
  exitFullscreen()
  const it = feed.value[activeIndex.value]
  // 退出竖屏 → 回到普通视频详情页（与横屏一样的 16:9 + 左右补黑边状态）
  router.replace(`/m/video/${it?.aId || startAId}`)
}

function enterFullscreen() {
  const el = document.documentElement
  const rfs = el.requestFullscreen || el.webkitRequestFullscreen
  if (rfs) {
    const p = rfs.call(el)
    if (p && p.catch) p.catch(() => {})
  }
}

function exitFullscreen() {
  if (document.fullscreenElement) {
    document.exitFullscreen().catch(() => {})
  } else if (document.webkitFullscreenElement && document.webkitExitFullscreen) {
    document.webkitExitFullscreen()
  }
}

function toggleFullscreen() {
  if (isFullscreen.value) exitFullscreen()
  else enterFullscreen()
}

function onFsChange() {
  isFullscreen.value = !!(document.fullscreenElement || document.webkitFullscreenElement)
}

onMounted(async () => {
  document.addEventListener('fullscreenchange', onFsChange)
  document.addEventListener('webkitfullscreenchange', onFsChange)
  enterFullscreen()
  const first = await loadDetail(startAId)
  if (first) {
    feed.value.push(makeItem({ ...first, resolved: true }))
  } else {
    feed.value.push(makeItem({ aId: startAId, title: '视频加载失败', resolved: true }))
  }
  const rec = await getRecommendVides(startAId)
  if (rec.code === '1') {
    ;(rec.data || []).forEach(r => {
      if (r.aId && !feed.value.some(x => x.aId === r.aId)) feed.value.push(makeItem(r))
    })
  }
  loading.value = false
  await nextTick()
  setupVideo(0)
  playActive()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFsChange)
  document.removeEventListener('webkitfullscreenchange', onFsChange)
  Object.values(hlsMap).forEach(h => { try { h.destroy() } catch (e) {} })
  cancelAnimationFrame(scrollRaf)
  exitFullscreen()
})
</script>

<template>
  <div class="vf-page">
    <!-- 加载动画 -->
    <div v-if="loading" class="vf-loading">
      <div class="vf-spinner"></div>
      <div class="vf-loading-text">加载中...</div>
    </div>

    <!-- 顶部返回 -->
    <div class="vf-topbar">
      <div class="vf-back" @click="goBack">
        <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#fff" stroke-width="2.4">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </div>
      <div class="vf-brand">竖屏 · 短视频</div>
      <div class="vf-fs" @click.stop="toggleFullscreen">
        <svg v-if="!isFullscreen" viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3" />
        </svg>
      </div>
    </div>

    <!-- 视频流 -->
    <div ref="containerRef" class="vf-container" @scroll="onScroll">
      <div v-for="(it, i) in feed" :key="it.aId" class="vf-slide">
        <!-- 横屏视频：模糊封面铺满全屏作为背景（抖音处理横屏方式），竖屏视频直接 cover 铺满 -->
        <div
          v-if="!it.isVertical && it.pic"
          class="vf-bg"
          :style="{ backgroundImage: `url(${it.pic})` }"
        ></div>
        <video
          :ref="(el) => (videoEls[i] = el)"
          :class="['vf-video', it.isVertical ? 'vf-video-cover' : 'vf-video-contain']"
          playsinline
          webkit-playsinline
          preload="metadata"
          :poster="it.isVertical ? it.pic || undefined : undefined"
          @timeupdate="onTimeUpdate"
          @ended="onEnded"
        ></video>

        <!-- 覆盖层 UI -->
        <div class="vf-overlay" @click="onOverlayTap">
          <!-- 右侧操作栏 -->
          <div class="vf-rail">
            <div class="vf-rail-item" @click.stop="doFollow(it)">
              <div class="vf-avatar-wrap">
                <img class="vf-avatar" :src="it.uploader?.face || ''" />
                <span class="vf-follow-badge" :class="{ on: it.following }">{{ it.following ? '✓' : '+' }}</span>
              </div>
            </div>
            <div class="vf-rail-item" @click.stop="toggleLike(it)">
              <div class="vf-icon" :class="{ active: it.liked }">
                <svg viewBox="0 0 24 24" width="34" height="34" :fill="it.liked ? '#fb7299' : 'none'" stroke="#fff" stroke-width="1.6">
                  <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
                </svg>
              </div>
              <span class="vf-rail-num">{{ fmt(it.likeCount) }}</span>
            </div>
            <div class="vf-rail-item" @click.stop="openComments(it)">
              <div class="vf-icon">
                <svg viewBox="0 0 24 24" width="34" height="34" fill="none" stroke="#fff" stroke-width="1.6">
                  <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
                </svg>
              </div>
              <span class="vf-rail-num">{{ fmt(it.commentCount) }}</span>
            </div>
            <div class="vf-rail-item" @click.stop="toggleStar(it)">
              <div class="vf-icon" :class="{ active: it.starred }">
                <svg viewBox="0 0 24 24" width="34" height="34" :fill="it.starred ? '#ffd04b' : 'none'" stroke="#fff" stroke-width="1.6">
                  <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01z" />
                </svg>
              </div>
              <span class="vf-rail-num">{{ it.starred ? '已藏' : '收藏' }}</span>
            </div>
            <div class="vf-rail-item" @click.stop="doShare(it)">
              <div class="vf-icon">
                <svg viewBox="0 0 24 24" width="34" height="34" fill="none" stroke="#fff" stroke-width="1.6">
                  <circle cx="18" cy="5" r="3" /><circle cx="6" cy="12" r="3" /><circle cx="18" cy="19" r="3" />
                  <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" /><line x1="15.41" y1="6.51" x2="8.59" y2="10.49" />
                </svg>
              </div>
              <span class="vf-rail-num">分享</span>
            </div>
          </div>

          <!-- 底部信息 -->
          <div class="vf-bottom">
            <div class="vf-author">@{{ it.author || 'UP主' }}</div>
            <div class="vf-title">{{ it.title }}</div>
            <div v-if="it.description" class="vf-desc">{{ it.description }}</div>
            <div class="vf-music">
              <span class="vf-music-disc">♪</span>
              <span class="vf-music-text">@{{ it.author }} · 原创视频</span>
            </div>
          </div>

          <!-- 播放/暂停指示 -->
          <div v-if="!playing && i === activeIndex" class="vf-play-hint">
            <svg viewBox="0 0 24 24" width="60" height="60" fill="rgba(255,255,255,0.85)">
              <path d="M8 5v14l11-7z" />
            </svg>
          </div>
        </div>

        <!-- 双击点赞心形 -->
        <div v-if="i === activeIndex && heartShow" class="vf-heart">
          <svg viewBox="0 0 24 24" width="120" height="120" fill="#fb7299" stroke="#fff" stroke-width="0.5">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
          </svg>
        </div>

        <!-- 顶部进度条（仅当前视频） -->
        <div v-if="i === activeIndex" class="vf-progress">
          <div class="vf-progress-inner" :style="{ width: (progress * 100) + '%' }"></div>
        </div>
      </div>
    </div>

    <!-- 评论面板 -->
    <div v-if="showComment" class="vf-comment-mask" @click="showComment = false">
      <div class="vf-comment-sheet" @click.stop>
        <div class="vf-comment-head">
          <span>评论 ({{ feed[activeIndex]?.commentCount || 0 }})</span>
          <span class="vf-comment-close" @click="showComment = false">✕</span>
        </div>
        <div v-if="commentLoading" class="vf-comment-empty">加载中...</div>
        <div v-else-if="comments.length === 0" class="vf-comment-empty">还没有评论</div>
        <div v-else class="vf-comment-list">
          <div v-for="c in comments" :key="c.rpid" class="vf-comment-item">
            <img :src="c.user?.face || ''" class="vf-comment-avatar" />
            <div class="vf-comment-body">
              <div class="vf-comment-name">{{ c.user?.name || '匿名' }}</div>
              <div class="vf-comment-text">{{ c.content }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.vf-page {
  position: fixed;
  inset: 0;
  background: #000;
  overflow: hidden;
  z-index: 100;
}

.vf-loading {
  position: absolute;
  inset: 0;
  z-index: 300;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: #000;
}
.vf-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-top-color: #fb7299;
  border-radius: 50%;
  animation: vf-spin 0.8s linear infinite;
}
.vf-loading-text {
  color: #999;
  font-size: 13px;
}


.vf-topbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 14px;
  background: linear-gradient(to bottom, rgba(0,0,0,0.5), transparent);
}
.vf-back { display: flex; align-items: center; cursor: pointer; }
.vf-fs { margin-left: auto; display: flex; align-items: center; cursor: pointer; opacity: 0.9; }
.vf-brand {
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 1px;
}

.vf-container {
  height: 100%;
  overflow-y: auto;
  scroll-snap-type: y mandatory;
  -webkit-overflow-scrolling: touch;
}

.vf-slide {
  position: relative;
  height: 100vh;
  height: 100dvh;
  scroll-snap-align: start;
  background: #000;
}

.vf-video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}
/* 竖屏视频：抖音式，cover 裁切铺满整个屏幕，不露黑边 */
.vf-video-cover {
  object-fit: cover;
}
/* 横屏视频：居中 contain，背景由模糊封面铺满 */
.vf-video-contain {
  object-fit: contain;
}
.vf-bg {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  filter: blur(28px) brightness(0.45);
  transform: scale(1.3);
}

.vf-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, rgba(0,0,0,0) 45%, rgba(0,0,0,0.45) 100%);
}

.vf-rail {
  position: absolute;
  right: 12px;
  bottom: 110px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 22px;
  z-index: 2;
}
.vf-rail-item { display: flex; flex-direction: column; align-items: center; gap: 4px; cursor: pointer; }
.vf-avatar-wrap { position: relative; }
.vf-avatar { width: 48px; height: 48px; border-radius: 50%; border: 2px solid #fff; object-fit: cover; background: #333; }
.vf-follow-badge {
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 20px; height: 20px;
  border-radius: 50%;
  background: #fb7299;
  color: #fff;
  font-size: 14px;
  line-height: 20px;
  text-align: center;
  border: 1px solid #fff;
  &.on { background: #4caf50; }
}
.vf-icon { display: flex; }
.vf-icon.active { transform: scale(1.05); }
.vf-rail-num { color: #fff; font-size: 12px; text-shadow: 0 1px 3px rgba(0,0,0,0.5); }

.vf-bottom {
  position: absolute;
  left: 16px;
  right: 84px;
  bottom: 40px;
  color: #fff;
  z-index: 2;
  text-shadow: 0 1px 3px rgba(0,0,0,0.6);
}
.vf-author { font-size: 16px; font-weight: 700; margin-bottom: 8px; }
.vf-title { font-size: 15px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8px; }
.vf-desc { font-size: 13px; line-height: 1.5; opacity: 0.85; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 10px; }
.vf-music { display: flex; align-items: center; gap: 8px; font-size: 13px; opacity: 0.9; }
.vf-music-disc {
  display: inline-block;
  width: 22px; height: 22px;
  border-radius: 50%;
  background: #fb7299;
  color: #fff;
  text-align: center;
  line-height: 22px;
  font-size: 12px;
  animation: vf-spin 4s linear infinite;
}
@keyframes vf-spin { from { transform: rotate(0); } to { transform: rotate(360deg); } }

.vf-heart {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 6;
  pointer-events: none;
  animation: vf-heart-pop 0.9s ease-out forwards;
}
@keyframes vf-heart-pop {
  0% { transform: translate(-50%, -50%) scale(0.2); opacity: 0; }
  30% { transform: translate(-50%, -50%) scale(1.3); opacity: 1; }
  70% { transform: translate(-50%, -50%) scale(1); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(1); opacity: 0; }
}

.vf-play-hint {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 3;
  pointer-events: none;
}

.vf-progress {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: rgba(255,255,255,0.2);
  z-index: 4;
}
.vf-progress-inner {
  height: 100%;
  background: #fb7299;
  transition: width 0.2s linear;
}

/* 评论面板 */
.vf-comment-mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 200;
  display: flex;
  align-items: flex-end;
}
.vf-comment-sheet {
  width: 100%;
  height: 60%;
  background: #fff;
  border-radius: 16px 16px 0 0;
  display: flex;
  flex-direction: column;
  padding: 14px 16px;
  box-sizing: border-box;
}
.vf-comment-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  font-size: 15px;
  color: #333;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}
.vf-comment-close { cursor: pointer; font-size: 16px; color: #999; }
.vf-comment-empty { text-align: center; color: #999; padding: 40px 0; }
.vf-comment-list { flex: 1; overflow-y: auto; padding-top: 8px; }
.vf-comment-item { display: flex; gap: 10px; padding: 10px 0; border-bottom: 1px solid #f6f6f6; }
.vf-comment-avatar { width: 36px; height: 36px; border-radius: 50%; object-fit: cover; background: #eee; }
.vf-comment-body { flex: 1; }
.vf-comment-name { font-size: 13px; color: #999; margin-bottom: 4px; }
.vf-comment-text { font-size: 14px; color: #333; line-height: 1.5; }
</style>
