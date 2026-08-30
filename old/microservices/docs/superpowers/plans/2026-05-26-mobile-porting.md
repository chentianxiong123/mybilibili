# 移动端 B站 React → Vue 3 移植计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 react-bilibili 移动端页面移植到 mybilibili-web 中，作为独立的移动端模块运行

**Architecture:** 在 mybilibili-web/src 下新建 `mobile/` 目录，所有移动端页面独立于现有桌面端。路由统一加 `/m` 前缀。共用现有的 API 层（axios）、状态管理（Pinia）、播放器（Artplayer）基础设施，UI 从 react-bilibili 的 Stylus + JSX 转为 Vue 3 SFC + SCSS。

**Tech Stack:** Vue 3 + Composition API + Pinia + Vue Router + SCSS + Artplayer + axios

**参考源:** `D:\files\mybilibili-next\react-bilibili\src\`

---

## 文件结构

```
mybilibili-web/src/
├── mobile/
│   ├── api/                    # 移动端 API 封装（调用已有 live.js/video.js 等）
│   │   └── index.ts
│   ├── components/             # 移动端通用组件
│   │   ├── Header.vue          # 顶部导航
│   │   ├── TabBar.vue          # 分类 Tab 切换
│   │   ├── Drawer.vue          # 分类抽屉
│   │   ├── VideoItem.vue       # 视频卡片
│   │   ├── ScrollToTop.vue     # 回到顶部
│   │   ├── VideoPlayer.vue     # 播放器封装（基于 Artplayer）
│   │   └── Barrage.vue         # 弹幕层
│   ├── views/
│   │   ├── home/
│   │   │   └── Index.vue       # 首页（轮播+推荐+排行榜）
│   │   ├── channel/
│   │   │   ├── Channel.vue     # 分类分区页
│   │   │   ├── Partition.vue   # 子分区
│   │   │   └── VideoLatest.vue # 最新视频
│   │   ├── ranking/
│   │   │   └── Ranking.vue     # 排行榜
│   │   ├── video/
│   │   │   ├── Detail.vue      # 视频详情
│   │   │   └── VideoPlayer.vue # 视频播放
│   │   ├── search/
│   │   │   ├── Search.vue      # 搜索页
│   │   │   └── Result.vue      # 搜索结果
│   │   ├── space/
│   │   │   ├── Space.vue       # 个人空间（历史记录）
│   │   │   ├── History.vue     # 历史记录
│   │   │   └── UpUser.vue      # UP 主主页
│   │   └── live/
│   │       ├── Index.vue       # 直播首页
│   │       ├── List.vue        # 直播列表
│   │       ├── Room.vue        # 直播间
│   │       └── Area.vue        # 直播分类
│   ├── router/
│   │   └── index.ts            # 移动端路由（/m/*）
│   ├── stores/
│   │   └── mobile.ts           # 移动端状态（可选）
│   ├── styles/
│   │   ├── variables.scss      # 移动端样式变量
│   │   ├── mixins.scss         # 1pxborder、px2rem 等
│   │   └── global.scss         # 全局移动端样式
│   └── utils/
│       ├── format.ts           # 数字格式化等工具
│       └── image.ts            # 图片处理工具
```

**计划分 12 个 Task 执行，每个 Task 完成后提交一次。**

---

### Task 1: 移动端基础架子搭建

**文件:**
- Create: `src/mobile/styles/variables.scss`
- Create: `src/mobile/styles/global.scss`
- Create: `src/mobile/router/index.ts`
- Modify: `src/router/index.js`（添加 /m 路由入口）

**Step 1: 创建移动端样式变量**

`src/mobile/styles/variables.scss`:
```scss
// 主题色
$theme-pink: #fb7299;
$theme-pink-hover: #f5729a;
$bg-color: #f4f4f4;
$bg-white: #fff;
$text-primary: #18191c;
$text-secondary: #9499a0;
$text-third: #c0c0c0;
$border-color: #e3e5e7;
$header-height: 46px;
$tab-height: 40px;
```

**Step 2: 创建全局样式**

`src/mobile/styles/global.scss`:
```scss
@import './variables';

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 14px;
  color: $text-primary;
  background: $bg-color;
  -webkit-font-smoothing: antialiased;
  // 移动端禁止缩放
  -webkit-text-size-adjust: 100%;
}

a {
  color: inherit;
  text-decoration: none;
}

.clear::after {
  content: '';
  display: block;
  clear: both;
}
```

**Step 3: 创建移动端路由**

`src/mobile/router/index.ts`:
```ts
import type { RouteRecordRaw } from 'vue-router'

const mobileRoutes: RouteRecordRaw[] = [
  {
    path: '/m/index',
    name: 'mobile-home',
    component: () => import('../views/home/Index.vue'),
    meta: { title: '哔哩哔哩', mobile: true }
  },
  {
    path: '/m/channel/:rId',
    name: 'mobile-channel',
    component: () => import('../views/channel/Channel.vue'),
    meta: { title: '分区', mobile: true }
  },
  {
    path: '/m/ranking/:rId',
    name: 'mobile-ranking',
    component: () => import('../views/ranking/Ranking.vue'),
    meta: { title: '排行榜', mobile: true }
  },
  {
    path: '/m/video/:aId',
    name: 'mobile-video',
    component: () => import('../views/video/Detail.vue'),
    meta: { title: '视频详情', mobile: true }
  },
  {
    path: '/m/search',
    name: 'mobile-search',
    component: () => import('../views/search/Search.vue'),
    meta: { title: '搜索', mobile: true }
  },
  {
    path: '/m/search/result',
    name: 'mobile-search-result',
    component: () => import('../views/search/Result.vue'),
    meta: { title: '搜索结果', mobile: true }
  },
  {
    path: '/m/space',
    name: 'mobile-space',
    component: () => import('../views/space/Space.vue'),
    meta: { title: '我的', mobile: true },
    children: [
      {
        path: '',
        redirect: '/m/space/history'
      },
      {
        path: 'history',
        component: () => import('../views/space/History.vue'),
        meta: { title: '历史记录', mobile: true }
      },
      {
        path: ':mId',
        component: () => import('../views/space/UpUser.vue'),
        meta: { title: 'UP主主页', mobile: true }
      }
    ]
  },
  {
    path: '/m/live',
    name: 'mobile-live',
    component: () => import('../views/live/Index.vue'),
    meta: { title: '直播', mobile: true }
  },
  {
    path: '/m/live/list',
    name: 'mobile-live-list',
    component: () => import('../views/live/List.vue'),
    meta: { title: '直播列表', mobile: true }
  },
  {
    path: '/m/live/areas',
    name: 'mobile-live-areas',
    component: () => import('../views/live/Area.vue'),
    meta: { title: '直播分类', mobile: true }
  },
  {
    path: '/m/live/:roomId',
    name: 'mobile-live-room',
    component: () => import('../views/live/Room.vue'),
    meta: { title: '直播间', mobile: true }
  }
]

export default mobileRoutes
```

**Step 4: 在主路由中添加移动端入口**

在 `src/router/index.js` 末尾，`export default router` 之前：
```js
import mobileRoutes from '../mobile/router/index.ts'

// 注册移动端路由
mobileRoutes.forEach(route => router.addRoute(route))
```

**Step 5: 初始化移动端目录**

创建移动端目录结构（空 views 目录占位）。

**Step 6: commit**

```bash
cd D:\files\mybilibili-next\mybilibili-cloud\mybilibili-web
git add src/mobile/styles/ src/mobile/router/ src/router/index.js
git commit -m "feat: add mobile entry structure with routing"
```

---

### Task 2: 移动端通用组件

**文件:**
- Create: `src/mobile/components/Header.vue`
- Create: `src/mobile/components/TabBar.vue`
- Create: `src/mobile/components/Drawer.vue`
- Create: `src/mobile/components/VideoItem.vue`
- Create: `src/mobile/components/ScrollToTop.vue`

**Step 1: Header 组件**

`src/mobile/components/Header.vue`:
```vue
<script setup>
// 顶部栏：搜索入口 + Logo
// 参考 react-bilibili: components/header/Header.tsx
</script>

<template>
  <div class="mobile-header">
    <a href="/m/index" class="logo">
      <img src="../assets/images/logo.png" alt="bilibili" />
    </a>
    <router-link to="/m/search" class="search-box">
      <i class="icon-search" />
      <span class="placeholder">搜索...</span>
    </router-link>
    <router-link to="/m/space" class="user-avatar">
      <img src="../assets/images/noface.gif" alt="user" />
    </router-link>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.mobile-header {
  display: flex;
  align-items: center;
  height: $header-height;
  padding: 0 12px;
  background: $theme-pink;

  .logo {
    width: 72px;
    margin-right: 12px;
    img { width: 100%; }
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    height: 30px;
    background: rgba(255,255,255,0.2);
    border-radius: 15px;
    padding: 0 12px;
    color: rgba(255,255,255,0.8);
    font-size: 13px;

    .placeholder { margin-left: 6px; }
  }

  .user-avatar {
    width: 30px;
    height: 30px;
    margin-left: 12px;
    border-radius: 50%;
    overflow: hidden;
    img { width: 100%; height: 100%; }
  }
}
</style>
```

**Step 2: TabBar 组件**

`src/mobile/components/TabBar.vue` — 横向滚动 Tab 条：
```vue
<script setup>
defineProps({
  data: { type: Array, required: true },
  activeId: { type: Number, default: 0 }
})
const emit = defineEmits(['click'])

const handleClick = (tab) => emit('click', tab)
</script>

<template>
  <div class="tab-bar">
    <div
      v-for="tab in data"
      :key="tab.id"
      :class="['tab-item', { active: tab.id === activeId }]"
      @click="handleClick(tab)"
    >
      {{ tab.name }}
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.tab-bar {
  display: flex;
  overflow-x: auto;
  white-space: nowrap;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar { display: none; }

  .tab-item {
    flex-shrink: 0;
    padding: 0 14px;
    height: $tab-height;
    line-height: $tab-height;
    font-size: 14px;
    color: $text-secondary;

    &.active {
      color: $theme-pink;
      position: relative;
      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 50%;
        transform: translateX(-50%);
        width: 20px;
        height: 3px;
        background: $theme-pink;
        border-radius: 2px;
      }
    }
  }
}
</style>
```

**Step 3: Drawer 组件**

`src/mobile/components/Drawer.vue` — 分区抽屉面板：
```vue
<script setup>
import { ref } from 'vue'

defineProps({
  data: { type: Array, required: true }
})
const emit = defineEmits(['click'])
const visible = ref(false)

const show = () => { visible.value = true }
const hide = () => { visible.value = false }
const handleClick = (tab) => { hide(); emit('click', tab) }

defineExpose({ show })
</script>

<template>
  <transition name="drawer-fade">
    <div v-if="visible" class="drawer-overlay" @click="hide">
      <div class="drawer-panel" @click.stop>
        <div
          v-for="tab in data"
          :key="tab.id"
          class="drawer-item"
          @click="handleClick(tab)"
        >
          {{ tab.name }}
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 1000;
}

.drawer-panel {
  background: $bg-white;
  border-radius: 0 0 12px 12px;
  padding: 12px 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.drawer-item {
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 14px;
  background: $bg-color;
  color: $text-primary;
  cursor: pointer;

  &:hover, &:active {
    background: $theme-pink;
    color: #fff;
  }
}

.drawer-fade-enter-active, .drawer-fade-leave-active {
  transition: opacity 0.2s;
}
.drawer-fade-enter-from, .drawer-fade-leave-to {
  opacity: 0;
}
</style>
```

**Step 4: VideoItem 组件**

`src/mobile/components/VideoItem.vue` — 视频卡片：
```vue
<script setup>
defineProps({
  video: { type: Object, required: true },
  showStatistics: { type: Boolean, default: true }
})

const formatNum = (n) => {
  if (!n) return '0'
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
</script>

<template>
  <router-link :to="'/m/video/' + video.aId" class="video-item">
    <div class="pic">
      <img :src="video.pic" :alt="video.title" loading="lazy" />
      <span v-if="video.duration" class="duration">{{ video.duration }}</span>
    </div>
    <div class="info">
      <p class="title">{{ video.title }}</p>
      <div v-if="showStatistics" class="stats">
        <span class="play">{{ formatNum(video.play) }}次播放</span>
        <span class="danmaku">{{ formatNum(video.videoReview) }}弹幕</span>
      </div>
    </div>
  </router-link>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.video-item {
  display: block;
  margin-bottom: 12px;

  .pic {
    position: relative;
    width: 100%;
    padding-top: 56.25%;
    border-radius: 6px;
    overflow: hidden;
    background: #e3e5e7;

    img {
      position: absolute;
      inset: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .duration {
      position: absolute;
      bottom: 4px;
      right: 4px;
      background: rgba(0,0,0,0.7);
      color: #fff;
      font-size: 11px;
      padding: 1px 6px;
      border-radius: 4px;
    }
  }

  .info {
    padding: 8px 4px;
  }

  .title {
    font-size: 14px;
    color: $text-primary;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .stats {
    margin-top: 6px;
    font-size: 12px;
    color: $text-secondary;
    display: flex;
    gap: 12px;
  }
}
</style>
```

**Step 5: ScrollToTop 组件**

`src/mobile/components/ScrollToTop.vue`:
```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const visible = ref(false)
let scrollHandler = null

onMounted(() => {
  scrollHandler = () => {
    visible.value = window.scrollY > 500
  }
  window.addEventListener('scroll', scrollHandler)
})

onUnmounted(() => {
  if (scrollHandler) window.removeEventListener('scroll', scrollHandler)
})

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>

<template>
  <transition name="fade">
    <div v-if="visible" class="to-top" @click="scrollToTop">
      <i class="icon-arrow-up" />
    </div>
  </transition>
</template>

<style scoped lang="scss">
.to-top {
  position: fixed;
  bottom: 60px;
  right: 16px;
  width: 40px;
  height: 40px;
  background: rgba(0,0,0,0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  z-index: 100;
  cursor: pointer;
}

.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
```

**Step 6: commit**

```bash
git add src/mobile/components/
git commit -m "feat: add mobile common components (Header, TabBar, Drawer, VideoItem, ScrollToTop)"
```

---

### Task 3: 移动端工具函数

**文件:**
- Create: `src/mobile/utils/format.ts`
- Create: `src/mobile/utils/image.ts`

**Step 1: 格式化工具**

`src/mobile/utils/format.ts`:
```ts
// 万单位格式化
export function formatTenThousand(num: number): string {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return String(num)
}

// 时间格式化：秒 → mm:ss
export function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

// 日期格式化
export function formatDate(dateStr: string): string {
  const d = new Date(dateStr.replace(/-/g, '/'))
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return minutes + '分钟前'
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return hours + '小时前'
  const days = Math.floor(hours / 24)
  if (days < 30) return days + '天前'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}
```

**Step 2: 图片工具**

`src/mobile/utils/image.ts`:
```ts
// 获取浏览器支持的图片后缀
export function getPicSuffix(): string {
  const canvas = document.createElement('canvas')
  if (canvas.getContext && canvas.toDataURL('image/webp').indexOf('image/webp') === 0) {
    return '.webp'
  }
  return '.jpg'
}
```

**Step 3: commit**

```bash
git add src/mobile/utils/
git commit -m "feat: add mobile utility functions"
```

---

### Task 4: 首页 Index

**文件:**
- Create: `src/mobile/views/home/Index.vue`
- Create: `src/mobile/api/index.ts`（首页 API）

**Step 1: 首页 API**

`src/mobile/api/index.ts`:
```ts
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 首页推荐视频 + 分区
export async function getHomeContent() {
  const res = await axios.get(`${API_BASE}/index`)
  return res.data
}

// 轮播图
export async function getBanners() {
  const res = await axios.get(`${API_BASE}/round-sowing`)
  return res.data
}
```

**Step 2: 首页页面**

`src/mobile/views/home/Index.vue`:
```vue
<script setup>
import { ref, onMounted } from 'vue'
import Header from '../../components/Header.vue'
import TabBar from '../../components/TabBar.vue'
import Drawer from '../../components/Drawer.vue'
import VideoItem from '../../components/VideoItem.vue'
import ScrollToTop from '../../components/ScrollToTop.vue'
import { getHomeContent, getBanners } from '../../api/index'

const banners = ref([])
const partitions = ref([])
const rankingVideos = ref([])
const additionalVideos = ref([])
const activeTabId = ref(0)
const drawerRef = ref(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const [contentRes, bannerRes] = await Promise.all([
      getHomeContent(),
      getBanners()
    ])
    if (contentRes.code === '1') {
      partitions.value = contentRes.data.oneLevelPartitions || []
      additionalVideos.value = contentRes.data.additionalVideos || []
      rankingVideos.value = contentRes.data.rankingVideos || []
    }
    if (bannerRes.code === '1') {
      banners.value = bannerRes.data || []
    }
    // 初始化轮播
    initSwiper()
  } catch (e) {
    console.error('首页加载失败:', e)
  } finally {
    loading.value = false
  }
})

const initSwiper = () => {
  // 使用简单的轮播效果，不引入 swiper 库
  let idx = 0
  const items = document.querySelectorAll('.swiper-slide')
  if (!items.length) return
  const showSlide = (i) => {
    items.forEach((el, index) => {
      el.style.display = index === i ? 'block' : 'none'
    })
  }
  showSlide(0)
  setInterval(() => {
    idx = (idx + 1) % items.length
    showSlide(idx)
  }, 3000)
}

const handleTabClick = (tab) => {
  if (tab.id === -1) {
    window.location.href = '/m/live'
    return
  }
  if (tab.id === 0) {
    window.location.href = '/m/index'
  } else {
    window.location.href = `/m/channel/${tab.id}`
  }
}

const tabBarData = computed(() => {
  return [
    { id: 0, name: '首页' },
    ...partitions.value,
    { id: -1, name: '直播' }
  ]
})

import { computed } from 'vue'

// 过滤重复视频
const displayVideos = computed(() => {
  let videos = [...rankingVideos.value]
  if (additionalVideos.value.length > 0) {
    const addIds = new Set(additionalVideos.value.map(v => v.aId))
    videos = videos.filter(v => !addIds.has(v.aId))
  }
  return videos
})
</script>

<template>
  <div class="mobile-index">
    <Header />
    <div class="partition-bar">
      <TabBar :data="tabBarData" :active-id="activeTabId" @click="handleTabClick" />
      <div class="switch-btn" @click="drawerRef?.show()">
        <i class="icon-arrow-down" />
      </div>
    </div>
    <Drawer :data="tabBarData" ref="drawerRef" @click="handleTabClick" />

    <div class="content" v-if="!loading">
      <!-- 轮播 -->
      <div v-if="banners.length" class="banner-slider">
        <div class="swiper-container">
          <div v-for="b in banners" :key="b.id" class="swiper-slide">
            <a :href="b.url">
              <img :src="b.pic" width="100%" height="100%" alt="" />
            </a>
          </div>
        </div>
      </div>

      <!-- 视频列表 -->
      <div class="video-section">
        <div v-if="additionalVideos.length" class="section-title">推荐</div>
        <div class="video-grid">
          <VideoItem
            v-for="v in additionalVideos"
            :key="v.aId"
            :video="v"
            :show-statistics="false"
          />
        </div>
      </div>

      <div class="video-section">
        <div class="section-title">排行榜</div>
        <div class="video-grid">
          <VideoItem
            v-for="v in displayVideos"
            :key="v.aId"
            :video="v"
            :show-statistics="true"
          />
        </div>
      </div>
    </div>

    <div v-else class="loading">加载中...</div>
    <ScrollToTop />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.mobile-index {
  padding-bottom: 20px;
}

.partition-bar {
  display: flex;
  align-items: center;
  background: $bg-white;
  border-bottom: 1px solid $border-color;

  .tab-bar { flex: 1; }

  .switch-btn {
    width: 40px;
    text-align: center;
    color: $text-secondary;
    font-size: 16px;
    cursor: pointer;
  }
}

.banner-slider {
  .swiper-slide {
    width: 100%;
    img { width: 100%; display: block; }
    display: none;
    &:first-child { display: block; }
  }
}

.video-section {
  padding: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
  color: $text-primary;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: $text-secondary;
}
</style>
```

**Step 3: commit**

```bash
git add src/mobile/api/index.ts src/mobile/views/home/
git commit -m "feat: add mobile home page with banner, partitions, video grid"
```

---

### Task 5: 分类频道 Channel 页面

**文件:**
- Create: `src/mobile/views/channel/Channel.vue`
- Create: `src/mobile/views/channel/Partition.vue`
- Create: `src/mobile/views/channel/VideoLatest.vue`

这一步建议和 Task 6-12 一起由子代理并行分发执行。由于需要完整的 API 对接和页面渲染逻辑，每个页面的实现模式与首页类似。

但为了计划完整，这里罗列需要创建的文件清单和每个页面的核心逻辑：

| 页面 | 源文件 | 核心逻辑 |
|------|--------|----------|
| Channel | `react-bilibili/src/views/channel/Channel.tsx` | 分区子分类切换 + 分类视频列表 |
| Ranking | `react-bilibili/src/views/ranking/Ranking.tsx` | 排行榜分区选择 + 视频排行列表 |
| Video Detail | `react-bilibili/src/views/video/Detail.tsx` + `VideoPlayer.tsx` | 视频播放 + 推荐 + 评论 + 弹幕 |
| Search | `react-bilibili/src/views/search/Search.tsx` + `Result.tsx` | 搜索热词 + 搜索建议 + 结果展示 |
| Space/History | `react-bilibili/src/views/space/Space.tsx` + `History.tsx` | 历史记录列表 |
| UpUser | `react-bilibili/src/views/space/UpUser.tsx` | UP 主信息 + 投稿列表 |
| Live Index | `react-bilibili/src/views/live/index/Index.tsx` | 直播首页 + 推荐直播间 |
| Live List | `react-bilibili/src/views/live/list/List.tsx` | 直播房间列表 |
| Live Room | `react-bilibili/src/views/live/room/Room.tsx` | 直播间播放 + 弹幕 + 主播信息 |

实现模式统一为：
1. 创建对应 `.vue` 文件，使用 Composition API (`<script setup>`)
2. 读取对应的 React 源文件，将 JSX 转换为 Vue template
3. 将 Redux 状态管理替换为组件内 `ref/reactive` 或 Pinia
4. 将 Stylus 样式转换为 SCSS
5. 调用现有 API 层 (mybilibili-web/src/api/) 获取数据

**注意:** 直播间的弹幕系统在 react-bilibili 中使用 WebSocket 直接连接 B 站弹幕服务器。在我们的项目中，应该复用现有的 `src/utils/reconnectingWs.js` 和弹幕逻辑。

---

### Task N: 移动端图片资源

**文件:**
- Copy: `react-bilibili/src/assets/images/` → `src/mobile/assets/images/`

**Step 1: 复制图片资源**

将 react-bilibili 的 `src/assets/images/` 目录下的图标和占位图复制到 `src/mobile/assets/images/`。

```bash
cp -r D:\files\mybilibili-next\react-bilibili\src\assets\images D:\files\mybilibili-next\mybilibili-cloud\mybilibili-web\src\mobile\assets\
```

**Step 2: commit**

```bash
git add src/mobile/assets/images/
git commit -m "chore: add mobile image assets"
```

---

### Task N+1: 验证移动端可访问

**Step 1: 启动 dev server 验证**

```bash
cd mybilibili-web
npm run dev
```

访问 `http://localhost:5173/m/index`，确认首页可以正常渲染。

**Step 2: 检查路由是否注册成功**

检查控制台是否有路由报错，确保所有 `/m/*` 路由都正确注册。

---

## API 适配说明

react-bilibili 使用的 API 地址是 `http://112.74.99.5:3000/web/api`，接口格式和我们的项目不同。在移植时需要：

1. **如果那台 API 服务器还活着** — 可以在 `.env.development` 中配置 `VITE_API_BASE=http://112.74.99.5:3000/web/api`，直接调用
2. **如果想对接自己的后端** — 需要修改 API 调用逻辑，适配自己后端的接口格式（code: 200 而非 code: "1" 等）

建议先用方案1快速跑通，后续再逐步对接自己的后端。

---

## 总计

| Task | 内容 | 文件数 | 预计时长 |
|------|------|--------|----------|
| 1 | 基础架子搭建 | 5 | 20min |
| 2 | 通用组件 | 5 | 30min |
| 3 | 工具函数 | 2 | 10min |
| 4 | 首页 Index | 2 | 30min |
| 5 | 频道 Channel | 3 | 20min |
| 6 | 排行榜 Ranking | 1 | 15min |
| 7 | 视频详情 + 播放器 | 2 | 30min |
| 8 | 搜索 | 2 | 20min |
| 9 | 空间/历史/UP主 | 3 | 20min |
| 10 | 直播首页/列表 | 2 | 20min |
| 11 | 直播间 | 1 | 30min |
| 12 | 图片资源 | 1 | 5min |
| — | **合计** | **~29** | **~4h** |