<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getVideoInfo, getRecommendVides, getComments } from '../../api/video'
import { likeManuscript, collectManuscript, shareManuscript, followUser, getInteractionStatus, checkFollow } from '../../api/interaction'
import { getToken } from '../../utils/session'

const route = useRoute()
const router = useRouter()
const startAId = parseInt(route.params.aId) || 0

const feed = ref([])
const activeIndex = ref(0)
const playing = ref(false)
const loading = ref(true)
const progress = ref(0)
const buffering = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const muted = ref(false)
const dragging = ref(false)
const showProgressBar = ref(false)
const heartShow = ref(false)
const heartStyle = ref({})
const containerRef = ref(null)
const videoEls = ref([])
let currentHlsIndex = -1
let loadingMore = false

// 评论面板
const showComment = ref(false)
const comments = ref([])
const commentLoading = ref(false)

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
    collectCount: basic.collectCount || 0,
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
      collectCount: d.collectCount || 0,
      resolved: true
    })
    // 拉取当前用户真实的点赞/收藏/关注状态
    loadInteractionState(it)
  }
}

// 拉取真实互动状态（点赞、收藏、关注），不阻塞渲染
async function loadInteractionState(it) {
  if (!getToken() || !it) return
  try {
    const res = await getInteractionStatus(it.aId)
    if (res.code === '1' && res.data) {
      const d = res.data
      it.liked = !!d.liked || !!d.isLiked
      it.starred = !!d.collected || !!d.isCollected
    }
  } catch (e) {}
  if (it.mid) {
    try {
      const r = await checkFollow(it.mid)
      if (r.code === '1') {
        it.following = typeof r.data === 'object' ? !!r.data.following : !!r.data
      }
    } catch (e) {}
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
        currentHlsIndex = i
      } else {
        el.src = url
      }
    })
  } else {
    el.src = url
  }
  el.muted = muted.value
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
  // 回收非当前视频的 HLS 实例，避免滑多条后带宽/内存泄漏
  if (currentHlsIndex !== -1 && currentHlsIndex !== except) {
    destroyHls(currentHlsIndex)
    currentHlsIndex = -1
  }
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
  // 预缓存相邻视频，提升滑动丝滑度
  preloadNearby(i)
}

// 预加载当前前后各1条：提前拉详情 + 预创建 HLS，滑到时秒开
function preloadNearby(i) {
  ;[i - 1, i + 1].forEach((idx) => {
    if (idx < 0 || idx >= feed.value.length) return
    const it = feed.value[idx]
    if (!it) return
    if (!it.resolved) {
      ensureLoaded(idx).then(() => {
        // 详情就绪后预挂载视频源（不播放）
        nextTick(() => { if (idx !== activeIndex.value) setupVideoQuiet(idx) })
      })
    } else if (!videoEls.value[idx]?.src && !hlsMap[it.aId]) {
      setupVideoQuiet(idx)
    }
  })
  // 接近末尾预拉更多
  if (feed.value.length - i <= 3 && !loadingMore) loadMore()
}

// 预挂载源但不自动播放（仅缓存）
function setupVideoQuiet(i) {
  const it = feed.value[i]
  const el = videoEls.value[i]
  if (!it || !el || el.src || hlsMap[it.aId]) return
  const url = it.playUrlHd || it.playUrlSd || it.playUrlLd || it.playUrl
  if (!url) return
  if (url.endsWith('.m3u8')) {
    import('hls.js').then(({ default: Hls }) => {
      if (Hls.isSupported() && !hlsMap[it.aId]) {
        const hls = new Hls()
        hls.loadSource(url)
        hls.attachMedia(el)
        hlsMap[it.aId] = hls
      } else if (!el.src) {
        el.src = url
      }
    })
  } else if (!el.src) {
    el.src = url
  }
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
  if (el && el.duration && !dragging.value) {
    progress.value = el.currentTime / el.duration
    currentTime.value = el.currentTime
    duration.value = el.duration
  }
}

function onWaiting() { buffering.value = true }
function onPlaying() { buffering.value = false; playing.value = true }
function onPause() { playing.value = false }
function onLoadedMeta(e) { duration.value = e.target.duration || 0 }

function onEnded() {
  playing.value = false
  progress.value = 0
}

function toggleMute() {
  muted.value = !muted.value
  videoEls.value.forEach((el) => { if (el) el.muted = muted.value })
}

function fmtTime(s) {
  if (!s || isNaN(s)) return '00:00'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${String(m).padStart(2,'0')}:${String(sec).padStart(2,'0')}`
}

const progressPct = computed(() => (dragging.value ? progress.value : (duration.value ? currentTime.value / duration.value : 0)) * 100)

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

function showHeart(x, y) {
  heartShow.value = false
  // 心形位置：默认居中；双击时跟随点击位置
  heartStyle.value = (x != null && y != null)
    ? { left: x + 'px', top: y + 'px' }
    : { left: '50%', top: '50%' }
  nextTick(() => { heartShow.value = true })
  clearTimeout(heartTimer)
  heartTimer = setTimeout(() => { heartShow.value = false }, 900)
}

// 触摸交互：短点=单击（播放/暂停），双击=点赞，长按拖拽=进度条
let lastTap = 0
let tapTimer = null
let touchStart = null
let holdTimer = null
let moved = false

function onTouchStart(e) {
  if (e.touches.length !== 1) return
  const t = e.touches[0]
  touchStart = { x: t.clientX, y: t.clientY, time: Date.now() }
  moved = false
  // 长按 180ms 后进入拖拽进度模式
  holdTimer = setTimeout(() => {
    if (!moved && touchStart) {
      dragging.value = true
      showProgressBar.value = true
      seekTo(touchStart.x)
    }
  }, 180)
}

function onTouchMove(e) {
  if (!touchStart || e.touches.length !== 1) return
  const t = e.touches[0]
  const dx = t.clientX - touchStart.x
  const dy = t.clientY - touchStart.y
  // 水平移动为主视为拖拽进度；垂直移动大视为滑动翻页
  if (Math.abs(dx) > 10 || Math.abs(dy) > 10) {
    if (!dragging.value) {
      moved = true
      clearTimeout(holdTimer) // 移动了，取消点击/长按
    }
  }
  if (dragging.value) {
    e.preventDefault()
    seekTo(t.clientX)
  }
}

function onTouchEnd() {
  clearTimeout(holdTimer)
  if (dragging.value) {
    // 结束拖拽，隐藏进度条
    dragging.value = false
    setTimeout(() => { showProgressBar.value = false }, 200)
    touchStart = null
    return
  }
  if (!touchStart || moved) { touchStart = null; return }
  // 短点 → 单击/双击逻辑
  const x = touchStart.x, y = touchStart.y
  touchStart = null
  const now = Date.now()
  if (now - lastTap < 300) {
    clearTimeout(tapTimer)
    lastTap = 0
    const it = feed.value[activeIndex.value]
    if (it && !it.liked) toggleLike(it)
    showHeart(x, y)
    return
  }
  lastTap = now
  clearTimeout(tapTimer)
  tapTimer = setTimeout(() => { togglePlay() }, 300)
}

// 根据屏幕 X 坐标定位进度
function seekTo(clientX) {
  const el = videoEls.value[activeIndex.value]
  if (!el || !duration.value) return
  const w = window.innerWidth
  const ratio = Math.max(0, Math.min(1, clientX / w))
  progress.value = ratio
  el.currentTime = ratio * duration.value
  currentTime.value = ratio * duration.value
}

// 桌面端鼠标：与触摸一致，按下后需停留 180ms 才进入拖拽，否则为点击
function onMouseDown(e) {
  touchStart = { x: e.clientX, y: e.clientY, time: Date.now() }
  moved = false
  holdTimer = setTimeout(() => {
    if (!moved && touchStart) {
      dragging.value = true
      showProgressBar.value = true
      seekTo(touchStart.x)
    }
  }, 180)
}
function onMouseMove(e) {
  if (!touchStart) return
  const dx = e.clientX - touchStart.x
  const dy = e.clientY - touchStart.y
  if (Math.abs(dx) > 10 || Math.abs(dy) > 10) {
    if (!dragging.value) {
      moved = true
      clearTimeout(holdTimer)
    }
  }
  if (dragging.value) {
    seekTo(e.clientX)
  }
}
function onMouseUp() {
  clearTimeout(holdTimer)
  if (dragging.value) {
    dragging.value = false
    setTimeout(() => { showProgressBar.value = false }, 200)
    touchStart = null
    return
  }
  if (!touchStart || moved) { touchStart = null; return }
  // 短点 → 单击/双击逻辑
  const x = touchStart.x, y = touchStart.y
  touchStart = null
  const now = Date.now()
  if (now - lastTap < 300) {
    clearTimeout(tapTimer)
    lastTap = 0
    const it = feed.value[activeIndex.value]
    if (it && !it.liked) toggleLike(it)
    showHeart(x, y)
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

async function loadMore() {
  if (loadingMore) return
  const last = feed.value[feed.value.length - 1]
  if (!last) return
  loadingMore = true
  try {
    const rec = await getRecommendVides(last.aId)
    if (rec.code === '1') {
      ;(rec.data || []).forEach(r => {
        if (r.aId && !feed.value.some(x => x.aId === r.aId)) feed.value.push(makeItem(r))
      })
    }
  } catch (e) {}
  loadingMore = false
}

const goBack = () => {
  const it = feed.value[activeIndex.value]
  // 退出竖屏 → 回到普通视频详情页（与横屏一样的 16:9 + 左右补黑边状态）
  router.replace(`/m/video/${it?.aId || startAId}`)
}

onMounted(async () => {
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
  // 浏览器拦截有声自动播放时，静音播放（抖音式体验，不静默等待）
  const el = videoEls.value[0]
  if (el) {
    el.play().then(() => { playing.value = true }).catch(() => {
      muted.value = true
      el.muted = true
      el.play().then(() => { playing.value = true }).catch(() => {})
    })
  }
  // 预缓存下一条
  preloadNearby(0)
})

onUnmounted(() => {
  Object.values(hlsMap).forEach(h => { try { h.destroy() } catch (e) {} })
  cancelAnimationFrame(scrollRaf)
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
          @waiting="onWaiting"
          @playing="onPlaying"
          @pause="onPause"
          @loadedmetadata="onLoadedMeta"
          @ended="onEnded"
        ></video>

        <!-- 缓冲指示 -->
        <div v-if="i === activeIndex && buffering && !dragging" class="vf-buffer">
          <div class="vf-spinner"></div>
        </div>

        <!-- 静音提示 -->
        <div v-if="i === activeIndex && muted" class="vf-muted-tip" @click.stop="toggleMute">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="#fff" stroke-width="2"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" fill="#fff"/><line x1="23" y1="9" x2="17" y2="15"/><line x1="17" y1="9" x2="23" y2="15"/></svg>
          <span>点按开声音</span>
        </div>

        <!-- 覆盖层 UI：触摸交互（短点/双击/长按拖拽进度） -->
        <div class="vf-overlay"
          @touchstart="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
          @mousedown="onMouseDown"
          @mousemove="onMouseMove"
          @mouseup="onMouseUp"
          @mouseleave="onMouseUp"
        >
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
              <span class="vf-rail-num">{{ fmt(it.collectCount) }}</span>
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

        <!-- 双击点赞心形（跟随点击位置） -->
        <div v-if="i === activeIndex && heartShow" class="vf-heart" :style="heartStyle">
          <svg viewBox="0 0 24 24" width="120" height="120" fill="#fb7299" stroke="#fff" stroke-width="0.5">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
          </svg>
        </div>

        <!-- 底部进度条：默认隐藏，仅按住拖拽时出现 -->
        <div v-if="i === activeIndex && showProgressBar" class="vf-progress">
          <span class="vf-time vf-time-cur">{{ fmtTime(currentTime) }}</span>
          <div class="vf-progress-track">
            <div class="vf-progress-inner" :style="{ width: progressPct + '%' }"></div>
            <div class="vf-progress-thumb" :style="{ left: progressPct + '%' }"></div>
          </div>
          <span class="vf-time vf-time-dur">{{ fmtTime(duration) }}</span>
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
  z-index: 6;
  pointer-events: none;
  transform: translate(-50%, -50%);
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
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 44px;
  z-index: 4;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  box-sizing: border-box;
  background: linear-gradient(to bottom, rgba(0,0,0,0.5), transparent);
  padding-top: 8px;
}
.vf-time { color: #fff; font-size: 11px; text-shadow: 0 1px 2px rgba(0,0,0,0.6); min-width: 34px; }
.vf-time-cur { text-align: right; }
.vf-progress-track {
  position: relative;
  flex: 1;
  height: 3px;
  background: rgba(255,255,255,0.25);
  border-radius: 2px;
  cursor: pointer;
}
.vf-progress-inner {
  height: 100%;
  background: #fb7299;
  border-radius: 2px;
}
.vf-progress-thumb {
  position: absolute;
  top: 50%;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #fb7299;
  transform: translate(-50%, -50%);
  box-shadow: 0 0 0 3px rgba(251,114,153,0.25);
}
.vf-mute-btn { display: flex; cursor: pointer; opacity: 0.9; }

.vf-buffer {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  z-index: 5;
}
.vf-muted-tip {
  position: absolute;
  top: 60px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 16px;
  background: rgba(0,0,0,0.6);
  color: #fff;
  font-size: 12px;
  cursor: pointer;
}

.vf-progress-old { display: none; }

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
