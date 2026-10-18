<template>
  <div class="home-wrap">
    <!-- ===== 活动区 ===== -->
    <div class="activity-row">
      <div class="carousel-box">
        <el-carousel
          v-if="banners.length"
          height="228px"
          :interval="5000"
          :autoplay="banners.length > 1"
          arrow="hover"
        >
          <el-carousel-item v-for="(b, i) in banners" :key="i">
            <a class="banner-link" :href="bannerTarget(b)" target="_blank" rel="noopener">
              <img :src="b.imageUrl || b.image_url" :alt="b.title || '活动 banner'">
            </a>
          </el-carousel-item>
        </el-carousel>
        <div v-else class="carousel-empty">暂无活动</div>
      </div>
      <div class="notice-box">
        <div class="notice-wrap">
          <div class="notice-title">
            <span class="notice-txt">热门活动</span>
            <a class="card-more" @click="comingSoon('热门活动')">更多<i class="iconfont icon-youjiantou"></i></a>
          </div>
          <ul v-if="activities.length" class="notice-body">
            <li v-for="(a, i) in activities" :key="i" class="hot-item">
              <span class="point-circle"></span>
              <p>{{ a }}</p>
            </li>
          </ul>
          <div v-else class="notice-empty">暂无活动</div>
        </div>
      </div>
    </div>

    <!-- ===== 数据区 ===== -->
    <div class="data-wrap">
      <div class="data-card">
        <!-- 视频数据卡 -->
        <div class="data-wrp" v-loading="loading.overview">
          <div class="tab-header">
            <div class="left-block">
              <div class="navs">
                <a class="data-tab" :class="{ current: dataTab === 'video' }" @click="dataTab = 'video'">视频数据</a>
                <a class="data-tab" :class="{ current: dataTab === 'article' }" @click="dataTab = 'article'">专栏数据</a>
              </div>
              <div class="tips">统计截至：{{ overview.updateTime }}（每日12点更新）</div>
            </div>
            <div class="right-block">
              <button class="query-btn" @click="goData">
                <i class="iconfont icon-shuju"></i>实时数据查询
              </button>
              <a class="rights-center" @click="comingSoon('权益中心')">权益中心</a>
            </div>
          </div>

          <div v-if="dataTab === 'video'" class="section-video">
            <div class="askb-row">
              <div class="askb-capsule" @click="goData">
                <i class="iconfont icon-shuju askb-icon"></i>
                <span class="askb-text">如何优化我的账号数据？</span>
              </div>
              <div class="askb-capsule" @click="goFans">
                <i class="iconfont icon-uper askb-icon"></i>
                <span class="askb-text">我的粉丝喜欢看什么？</span>
              </div>
            </div>

            <div class="metric-row metric-row--first">
              <div v-for="item in metricsRow1" :key="item.key" class="metric-col metric-col--3">
                <div class="metric-card">
                  <div class="metric-top">
                    <div class="metric-name">
                      <i class="iconfont" :class="item.icon"></i>
                      <span>{{ item.label }}</span>
                    </div>
                    <div class="metric-diff" :class="{ up: item.increase > 0, down: item.increase < 0 }">
                      <span v-if="item.increase > 0" class="delta-arrow">▲</span>
                      <span v-else-if="item.increase < 0" class="delta-arrow">▼</span>
                      <span>{{ formatNumber(item.increase) }}</span>
                    </div>
                  </div>
                  <div class="metric-value">{{ formatNumber(item.value) }}</div>
                </div>
              </div>
            </div>

            <div class="metric-row">
              <div v-for="item in metricsRow2" :key="item.key" class="metric-col metric-col--4">
                <div class="metric-card">
                  <div class="metric-top">
                    <div class="metric-name">
                      <i class="iconfont" :class="item.icon"></i>
                      <span>{{ item.label }}</span>
                    </div>
                    <div class="metric-diff" :class="{ up: item.increase > 0, down: item.increase < 0 }">
                      <span v-if="item.increase > 0" class="delta-arrow">▲</span>
                      <span v-else-if="item.increase < 0" class="delta-arrow">▼</span>
                      <span>{{ formatNumber(item.increase) }}</span>
                    </div>
                  </div>
                  <div class="metric-value">{{ formatNumber(item.value) }}</div>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="section-empty">
            <el-empty description="暂无专栏数据" :image-size="80" />
          </div>
        </div>

        <!-- 粉丝概览卡 -->
        <div class="data-right">
          <div class="fan-overview">
            <div class="fan-header">
              <span class="fan-title">粉丝概览</span>
              <a class="fan-enter" @click="goFans"><i class="iconfont icon-youjiantou"></i></a>
            </div>
            <div class="fan-item">
              <div class="fan-label">粉丝总数</div>
              <div class="fan-num">{{ formatNumber(overview.totalFollowers) }}</div>
            </div>
            <div class="fan-item">
              <div class="fan-label">今日粉丝数变化</div>
              <div class="fan-num">{{ overview.followersIncrease > 0 ? '+' + formatNumber(overview.followersIncrease) : formatNumber(overview.followersIncrease) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 评论 / 弹幕 ===== -->
    <div class="interact-bill-wrap">
      <div class="interact-wrap">
        <header class="card-header">
          <ul class="tabs-header">
            <li class="tabs-item" :class="{ active: commentTab === 'comment' }" @click="switchCommentTab('comment')"><span>评论</span></li>
            <li class="tabs-item" :class="{ active: commentTab === 'danmu' }" @click="switchCommentTab('danmu')"><span>弹幕</span></li>
          </ul>
          <a class="card-more" @click="goInteract">更多<i class="iconfont icon-youjiantou"></i></a>
        </header>

        <div class="interact-content" v-loading="commentTab === 'comment' ? loading.comment : loading.danmu">
          <!-- 评论 -->
          <div v-if="commentTab === 'comment'" class="comment-wrap">
            <div v-for="c in topComments" :key="c.id" class="comment-item">
              <div class="left-block">
                <el-avatar :size="28" :src="c.avatar" class="row-avatar">
                  {{ (c.username || '?').slice(0, 1) }}
                </el-avatar>
              </div>
              <div class="right-block">
                <div class="replier-time">
                  <a class="replier">{{ c.username || '未知用户' }}</a>
                  <span class="pointer"></span>
                  <p class="time">{{ c.time || formatDate(c.createTime) }}</p>
                </div>
                <div class="reply-content">{{ c.content }}</div>
                <div v-if="c.manuscriptId" class="origin">
                  <a class="comment-origin" @click="openManuscript(c.manuscriptId)">
                    <i class="iconfont icon-video"></i>
                    <span>{{ c.manuscriptTitle || '查看视频' }}</span>
                  </a>
                </div>
              </div>
              <button class="item-del" title="删除" @click="handleDeleteComment(c)">删除</button>
            </div>
            <el-empty v-if="!loading.comment && !latestComments.length" description="暂无评论" :image-size="70" />
          </div>

          <!-- 弹幕 -->
          <div v-else class="comment-wrap">
            <div v-for="d in danmakuList" :key="d.id" class="comment-item danmu-item">
              <div class="danmu-block">
                <div class="reply-content">{{ d.content }}</div>
                <div class="origin">
                  <a class="comment-origin" @click="openVideo(d.video_id, d.time)">
                    <i class="iconfont icon-video"></i>
                    <span>{{ d.manuscript_title || ('视频 #' + d.video_id) }}</span>
                  </a>
                  <span class="time">{{ d.created_at ? formatDate(d.created_at) : '' }}</span>
                </div>
              </div>
              <button class="item-del" title="删除" @click="handleDeleteDanmaku(d.id)">删除</button>
            </div>
            <el-empty v-if="!loading.danmu && !danmakuList.length" description="暂无弹幕" :image-size="70" />
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 排行 ===== -->
    <div class="bill-rank-wrap">
      <div class="rank-wrap" v-loading="loading.rank">
        <header class="card-header">
          <ul class="tabs-header">
            <li class="tabs-item" :class="{ active: rankTab === 'view' }" @click="rankTab = 'view'"><span>观看排行</span></li>
            <li class="tabs-item" :class="{ active: rankTab === 'interaction' }" @click="rankTab = 'interaction'"><span>互动排行</span></li>
          </ul>
          <a class="card-more" @click="goData">全部排行<i class="iconfont icon-youjiantou"></i></a>
        </header>

        <div class="rank-content">
          <div class="rank-list">
            <a v-for="(u, i) in currentRank" :key="u.id" class="rank-item" @click="openSpace(u.id)">
              <div class="rank-avatar-wrap">
                <el-avatar :size="56" :src="u.avatar" class="rank-avatar">
                  {{ (u.username || '?').slice(0, 1) }}
                </el-avatar>
                <span class="rank-badge" :class="'rank-badge--' + (i + 1)">{{ i + 1 }}</span>
              </div>
              <span class="rank-name" :title="u.username">{{ u.username }}</span>
            </a>
            <el-empty v-if="!loading.rank && !currentRank.length" description="暂无排行数据" :image-size="70" />
          </div>
          <div class="give-share">
            <button class="bili-btn give" @click="comingSoon('赠送礼物')">赠送礼物</button>
            <button class="bili-btn share" @click="comingSoon('分享排行')">分享排行</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 创作素材 ===== -->
    <div class="material-wrp">
      <header class="material-header">
        <div class="material-tag">创作素材</div>
        <div class="material-op">
          <button class="up-btn" @click="comingSoon('素材投稿')">素材投稿</button>
          <a class="card-more" @click="comingSoon('更多素材')">更多<i class="iconfont icon-youjiantou"></i></a>
        </div>
      </header>
      <div class="material-content">
        <ul class="tabs-header material-tabs">
          <li
            v-for="t in materialTabs"
            :key="t.key"
            class="tabs-item"
            :class="{ active: materialTab === t.key }"
            @click="materialTab = t.key"
          >
            <span>{{ t.label }}</span>
          </li>
        </ul>
        <div class="material-row">
          <div v-for="n in 4" :key="n" class="material-col">
            <div class="material-card">
              <i class="iconfont icon-video"></i>
              <span>敬请期待</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi, statsApi } from '@/api/creator'
import { getHomeBanners } from '@/api/banner'

interface Overview {
  totalFollowers: number
  totalViews: number
  totalComments: number
  totalDanmaku: number
  totalLikes: number
  totalShares: number
  totalCollections: number
  totalCoins: number
  followersIncrease: number
  viewsIncrease: number
  commentsIncrease: number
  danmakuIncrease: number
  likesIncrease: number
  sharesIncrease: number
  collectionsIncrease: number
  coinsIncrease: number
  updateTime: string
}

interface CommentItem {
  id: number
  content: string
  manuscriptId?: number
  manuscriptTitle?: string
  username?: string
  avatar?: string
  time?: string
  createTime?: string
}

interface DanmakuItem {
  id: number
  video_id: number
  content: string
  time: number
  mode: number
  created_at: string
  manuscript_title?: string
}

interface RankItem {
  id: number
  username: string
  avatar?: string
  interactionCount?: number
}

const router = useRouter()

const loading = reactive({
  overview: false,
  comment: false,
  danmu: false,
  rank: false
})

const overview = ref<Overview>({
  totalFollowers: 0,
  totalViews: 0,
  totalComments: 0,
  totalDanmaku: 0,
  totalLikes: 0,
  totalShares: 0,
  totalCollections: 0,
  totalCoins: 0,
  followersIncrease: 0,
  viewsIncrease: 0,
  commentsIncrease: 0,
  danmakuIncrease: 0,
  likesIncrease: 0,
  sharesIncrease: 0,
  collectionsIncrease: 0,
  coinsIncrease: 0,
  updateTime: ''
})

// ===== 活动区 =====
const banners = ref<any[]>([])
const activities = ref<string[]>([])

// ===== 视频数据 / 专栏数据 =====
const dataTab = ref<'video' | 'article'>('video')

const metrics = computed(() => [
  { key: 'views', label: '播放量', icon: 'icon-bofang', value: overview.value.totalViews, increase: overview.value.viewsIncrease },
  { key: 'comments', label: '评论', icon: 'icon-pinglun', value: overview.value.totalComments, increase: overview.value.commentsIncrease },
  { key: 'danmaku', label: '弹幕', icon: 'icon-danmushu', value: overview.value.totalDanmaku, increase: overview.value.danmakuIncrease },
  { key: 'likes', label: '点赞', icon: 'icon-dianzan', value: overview.value.totalLikes, increase: overview.value.likesIncrease },
  { key: 'shares', label: '分享', icon: 'icon-zhuanfa', value: overview.value.totalShares, increase: overview.value.sharesIncrease },
  { key: 'collections', label: '收藏', icon: 'icon-shoucang', value: overview.value.totalCollections, increase: overview.value.collectionsIncrease },
  { key: 'coins', label: '投币', icon: 'icon-toubi', value: overview.value.totalCoins, increase: overview.value.coinsIncrease }
])
const metricsRow1 = computed(() => metrics.value.slice(0, 3))
const metricsRow2 = computed(() => metrics.value.slice(3))

// ===== 评论 / 弹幕 =====
const commentTab = ref<'comment' | 'danmu'>('comment')
const latestComments = ref<CommentItem[]>([])
const topComments = computed(() => latestComments.value.slice(0, 3))

const danmakuList = ref<DanmakuItem[]>([])
let danmakuLoaded = false

// ===== 排行 =====
const rankTab = ref<'view' | 'interaction'>('view')
const viewRanking = ref<RankItem[]>([])
const interactionRanking = ref<RankItem[]>([])
const currentRank = computed(() => (rankTab.value === 'view' ? viewRanking.value : interactionRanking.value))

// ===== 创作素材 =====
const materialTab = ref('video')
const materialTabs = [
  { key: 'video', label: '视频素材' },
  { key: 'sticker', label: '贴纸素材' },
  { key: 'avatar', label: '虚拟形象' }
]

const formatNumber = (n: number | undefined) => (n ?? 0).toLocaleString('en-US')

const formatDate = (t?: string) => {
  if (!t) return ''
  const date = new Date(t)
  if (isNaN(date.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

const bannerTarget = (b: any) => {
  const url = b?.link_url || b?.linkUrl || ''
  return url ? url.replace('/manuscript/', '/video/') : '#'
}

const loadBanners = async () => {
  try {
    const res = await getHomeBanners()
    const list = res?.data?.data || res?.data || []
    banners.value = (Array.isArray(list) ? list : []).slice(0, 6)
  } catch (e) {
    console.error('加载轮播失败:', e)
  }
}

const loadOverview = async () => {
  loading.overview = true
  try {
    const res = await statsApi.getOverview()
    if (res?.code === 200 && res.data) {
      overview.value = { ...overview.value, ...res.data }
    }
  } catch (e) {
    console.error('加载数据概览失败:', e)
  } finally {
    loading.overview = false
  }
}

const loadLatestComments = async () => {
  loading.comment = true
  try {
    const res = await statsApi.getLatestComments(3)
    if (res?.code === 200 && res.data) {
      latestComments.value = res.data
    }
  } catch (e) {
    console.error('加载最新评论失败:', e)
  } finally {
    loading.comment = false
  }
}

const loadRanking = async () => {
  loading.rank = true
  try {
    const [viewRes, interactionRes] = await Promise.all([
      statsApi.getFansRanking('view', 3),
      statsApi.getFansRanking('interaction', 3)
    ])
    if (viewRes?.code === 200 && viewRes.data) viewRanking.value = viewRes.data
    if (interactionRes?.code === 200 && interactionRes.data) interactionRanking.value = interactionRes.data
  } catch (e) {
    console.error('加载排行榜失败:', e)
  } finally {
    loading.rank = false
  }
}

const fetchDanmaku = async () => {
  loading.danmu = true
  try {
    const res = await creatorApi.getDanmakuList({ page: 1, size: 3 })
    if (res?.code === 200 && res.data) {
      danmakuList.value = res.data.list || []
    }
  } catch (e) {
    console.error('加载弹幕失败:', e)
  } finally {
    loading.danmu = false
  }
}

const switchCommentTab = (tab: 'comment' | 'danmu') => {
  commentTab.value = tab
  if (tab === 'danmu' && !danmakuLoaded) {
    danmakuLoaded = true
    fetchDanmaku()
  }
}

const handleDeleteComment = async (comment: CommentItem) => {
  try {
    await ElMessageBox.confirm('确定要删除这条评论吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await creatorApi.deleteComment(comment.id)
    if (res?.code === 200) {
      ElMessage.success('删除成功')
      loadLatestComments()
      loadOverview()
    } else {
      ElMessage.error(res?.message || '删除失败')
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error('删除评论失败:', e)
      ElMessage.error('删除失败')
    }
  }
}

const handleDeleteDanmaku = async (danmakuId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这条弹幕吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await creatorApi.deleteDanmaku(danmakuId)
    if (res?.code === 200) {
      ElMessage.success('删除成功')
      fetchDanmaku()
      loadOverview()
    } else {
      ElMessage.error(res?.message || '删除失败')
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error('删除弹幕失败:', e)
      ElMessage.error('删除失败')
    }
  }
}

const openManuscript = (manuscriptId?: number) => {
  if (!manuscriptId) return
  window.open(`/manuscript/${manuscriptId}`, '_blank')
}

const openVideo = (videoId?: number, time?: number) => {
  if (!videoId) return
  const t = time ? `?t=${Math.floor(time)}` : ''
  window.open(`/video/${videoId}${t}`, '_blank')
}

const openSpace = (userId?: number) => {
  if (!userId) return
  router.push(`/space/${userId}`)
}

const goInteract = () => {
  router.push(commentTab.value === 'comment' ? '/platform/comment' : '/platform/danmu')
}
const goData = () => {
  router.push('/platform/data-up')
}
const goFans = () => {
  router.push('/platform/fans')
}
const comingSoon = (name: string) => {
  ElMessage.info(`${name}功能敬请期待`)
}

onMounted(() => {
  loadBanners()
  loadOverview()
  loadLatestComments()
  loadRanking()
})
</script>

<style scoped>
.home-wrap {
  color: var(--text1, #18191c);
}

/* ===== 通用卡片头 / 更多 ===== */
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tabs-header {
  display: flex;
  align-items: center;
  margin: 0;
  padding: 0;
  list-style: none;
}

.tabs-item {
  position: relative;
  margin-right: 24px;
  font-size: 16px;
  line-height: 22px;
  color: #18191c;
  cursor: pointer;
  transition: color .2s;
}

.tabs-item:last-child {
  margin-right: 0;
}

.tabs-item:hover {
  color: #00aeec;
}

.tabs-item.active {
  color: #00aeec;
  font-weight: 700;
}

.card-more {
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  color: #999;
  cursor: pointer;
  transition: color .2s;
}

.card-more:hover {
  color: #00aeec;
}

.card-more .iconfont {
  margin-left: 2px;
  font-size: 12px;
}

/* ===== 活动区 ===== */
.activity-row {
  display: flex;
}

.carousel-box {
  flex: 0 0 644px;
  max-width: calc(100% - 15px);
  height: 228px;
  border-radius: 8px;
  overflow: hidden;
}

.banner-link,
.banner-link img {
  display: block;
  width: 100%;
  height: 228px;
}

.carousel-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 228px;
  border-radius: 8px;
  background: #fff;
  font-size: 14px;
  color: var(--text3, #9499a0);
}

.notice-box {
  flex: 1;
  min-width: 0;
  margin-left: 15px;
}

.notice-wrap {
  height: 228px;
  box-sizing: border-box;
  padding: 16px 22px 16px 24px;
  border-radius: 2px;
  background: #fff;
}

.notice-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.notice-txt {
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
}

.notice-body {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
}

.hot-item {
  display: flex;
  align-items: center;
  margin-top: 12px;
  font-size: 13px;
  color: var(--text2, #61666d);
}

.point-circle {
  flex: none;
  width: 4px;
  height: 4px;
  margin-right: 8px;
  border-radius: 50%;
  background: #00aeec;
}

.hot-item p {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 160px;
  font-size: 13px;
  color: var(--text3, #9499a0);
}

/* ===== 数据区 ===== */
.data-wrap {
  margin-top: 16px;
}

.data-card {
  display: flex;
  gap: 16px;
}

.data-wrp {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  padding: 24px;
  border-radius: 8px;
  background: #fff;
}

.tab-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tab-header .left-block {
  display: flex;
  align-items: center;
  min-width: 0;
}

.navs {
  display: flex;
  align-items: center;
}

.data-tab {
  margin-right: 24px;
  font-size: 16px;
  line-height: 20px;
  color: #18191c;
  cursor: pointer;
  transition: color .2s;
}

.data-tab.current {
  margin-right: 35px;
  color: #00aeec;
  font-weight: 700;
}

.tips {
  font-size: 12px;
  color: #9499a0;
  white-space: nowrap;
}

.tab-header .right-block {
  display: flex;
  align-items: center;
  flex: none;
}

.query-btn {
  display: inline-flex;
  align-items: center;
  height: 32px;
  margin-right: 12px;
  padding: 0 14px;
  border: 1px solid var(--line_regular, #e3e5e7);
  border-radius: 4px;
  background: #fff;
  font-size: 13px;
  color: #18191c;
  cursor: pointer;
  transition: border-color .2s, color .2s;
}

.query-btn:hover {
  border-color: #00aeec;
  color: #00aeec;
}

.query-btn .iconfont {
  margin-right: 4px;
  font-size: 14px;
}

.rights-center {
  display: inline-flex;
  align-items: center;
  height: 32px;
  font-size: 13px;
  color: #00a1d6;
  cursor: pointer;
}

.rights-center:hover {
  text-decoration: underline;
}

.section-video {
  margin-top: 24px;
}

.askb-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.askb-capsule {
  display: flex;
  align-items: center;
  height: 36px;
  padding: 0 12px;
  border-radius: 18px;
  background: #f6f7f8;
  cursor: pointer;
  transition: background-color .2s;
}

.askb-capsule:hover {
  background: #eef1f3;
}

.askb-icon {
  margin-right: 6px;
  font-size: 20px;
  color: #00aeec;
}

.askb-text {
  font-size: 13px;
  color: #18191c;
}

.metric-row {
  display: flex;
  margin: 0 -8px;
}

.metric-row--first {
  margin-bottom: 16px;
}

.metric-col {
  box-sizing: border-box;
  padding: 0 8px;
}

.metric-col--3 {
  flex: 0 0 33.3333%;
  max-width: 33.3333%;
}

.metric-col--4 {
  flex: 0 0 25%;
  max-width: 25%;
}

.metric-card {
  height: 128px;
  box-sizing: border-box;
  padding: 16px 16px 20px;
  border-radius: 8px;
  background: rgba(0, 174, 236, .02);
}

.metric-top {
  display: flex;
  align-items: center;
}

.metric-name {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #61666d;
}

.metric-name .iconfont {
  margin-right: 4px;
  font-size: 15px;
  color: #61666d;
}

.metric-diff {
  display: flex;
  align-items: center;
  margin-left: 5px;
  font-size: 14px;
  font-weight: 500;
  color: #ff4684;
}

.metric-diff.down {
  color: #00b578;
}

.delta-arrow {
  margin-right: 2px;
  font-size: 9px;
}

.metric-value {
  margin-top: 14px;
  font-size: 26px;
  font-weight: 500;
  line-height: 22px;
  color: #00aeec;
}

.section-empty {
  margin-top: 24px;
  padding: 12px 0 24px;
}

/* 粉丝概览卡 */
.data-right {
  display: flex;
  flex-direction: column;
  flex: 0 0 281px;
}

.fan-overview {
  flex: 1;
  box-sizing: border-box;
  padding: 16px 20px;
  border-radius: 8px;
  background: #fff;
}

.fan-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.fan-title {
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
}

.fan-enter {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  color: #00a1d6;
  cursor: pointer;
}

.fan-item {
  margin-top: 16px;
}

.fan-item + .fan-item {
  margin-top: 24px;
}

.fan-label {
  font-size: 12px;
  color: #61666d;
}

.fan-num {
  margin-top: 8px;
  font-size: 26px;
  line-height: 31px;
  color: #18191c;
}

/* ===== 评论 / 弹幕 ===== */
.interact-bill-wrap {
  margin-top: 16px;
}

.interact-wrap {
  box-sizing: border-box;
  padding: 24px;
  border-radius: 2px;
  background: #fff;
}

.interact-content {
  margin-top: 24px;
  min-height: 132px;
}

.comment-wrap {
  display: flex;
  gap: 16px;
}

.comment-item {
  position: relative;
  flex: 1;
  min-width: 0;
  display: flex;
}

.comment-item .left-block {
  flex: none;
  margin-right: 12px;
}

.row-avatar {
  flex: none;
  background: #f1f2f3;
  color: #9499a0;
  font-size: 13px;
}

.comment-item .right-block {
  flex: 1;
  min-width: 0;
}

.replier-time {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #757575;
}

.replier {
  margin-right: 8px;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #505050;
  cursor: pointer;
}

.replier:hover {
  color: #00aeec;
}

.pointer {
  flex: none;
  width: 3px;
  height: 3px;
  margin-right: 8px;
  border-radius: 50%;
  background: #757575;
}

.replier-time .time {
  margin: 0;
  font-size: 13px;
  color: #757575;
  white-space: nowrap;
}

.reply-content {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.5;
  color: #212121;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.origin {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--line_light, #f1f2f3);
}

.comment-origin {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  font-size: 12px;
  color: #757575;
  cursor: pointer;
}

.comment-origin .iconfont {
  flex: none;
  margin-right: 4px;
  font-size: 13px;
}

.comment-origin span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.comment-origin:hover {
  color: #00aeec;
}

.origin .time {
  flex: none;
  font-size: 12px;
  color: #757575;
}

.danmu-item {
  padding: 0;
}

.danmu-block {
  flex: 1;
  min-width: 0;
}

.item-del {
  position: absolute;
  top: 0;
  right: 0;
  padding: 0 6px;
  border: none;
  border-radius: 4px;
  background: transparent;
  font-size: 12px;
  color: #9499a0;
  cursor: pointer;
  opacity: 0;
  transition: opacity .2s, color .2s, background-color .2s;
}

.comment-item:hover .item-del {
  opacity: 1;
}

.item-del:hover {
  color: #f85a54;
  background: #feecea;
}

/* ===== 排行 ===== */
.bill-rank-wrap {
  margin-top: 16px;
}

.rank-wrap {
  box-sizing: border-box;
  padding: 24px;
  border-radius: 2px;
  background: #fff;
}

.rank-content {
  display: flex;
  margin-top: 24px;
}

.rank-list {
  flex: 1;
  min-width: 0;
  display: flex;
}

.rank-item {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.rank-avatar-wrap {
  position: relative;
}

.rank-avatar {
  background: #f1f2f3;
  color: #9499a0;
  transition: transform .2s;
}

.rank-item:hover .rank-avatar {
  transform: translateY(-2px);
}

.rank-badge {
  position: absolute;
  left: 50%;
  bottom: -8px;
  transform: translateX(-50%);
  min-width: 34px;
  height: 20px;
  padding: 0 6px;
  box-sizing: border-box;
  border-radius: 10px;
  font-size: 12px;
  line-height: 20px;
  text-align: center;
  font-weight: 700;
  color: #fff;
  background: #c9ccd0;
}

.rank-badge--1 {
  background: #fed443;
  color: #7a5600;
}

.rank-badge--2 {
  background: #d9dde3;
  color: #5a6169;
}

.rank-badge--3 {
  background: #e8a76a;
  color: #6b3d10;
}

.rank-name {
  margin-top: 20px;
  max-width: 90%;
  font-size: 14px;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.give-share {
  flex: 0 0 287px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.bili-btn {
  width: 184px;
  height: 32px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color .2s, color .2s;
}

.bili-btn.give {
  border: none;
  background: #00a1d6;
  color: #fff;
}

.bili-btn.give:hover {
  background: #0096c7;
}

.bili-btn.share {
  margin-top: 16px;
  border: 1px solid #00a1d6;
  background: #fff;
  color: #00a1d6;
}

.bili-btn.share:hover {
  background: rgba(0, 161, 214, .06);
}

/* ===== 创作素材 ===== */
.material-wrp {
  margin-top: 16px;
  box-sizing: border-box;
  padding: 24px;
  border-radius: 2px;
  background: #fff;
}

.material-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.material-tag {
  font-size: 16px;
  color: #212121;
}

.material-op {
  display: flex;
  align-items: center;
}

.up-btn {
  margin-right: 20px;
  padding: 6px 16px;
  border: 1px solid #20aae2;
  border-radius: 18px;
  background: #fff;
  font-size: 12px;
  color: #20aae2;
  cursor: pointer;
  transition: background-color .2s;
}

.up-btn:hover {
  background: rgba(32, 170, 226, .08);
}

.material-content {
  margin-top: 10px;
}

.material-tabs {
  display: flex;
}

.material-row {
  display: flex;
  margin: 24px -8px 0;
}

.material-col {
  flex: 0 0 25%;
  max-width: 25%;
  box-sizing: border-box;
  padding: 0 8px 8px;
}

.material-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 240px;
  border-radius: 8px;
  background: #f6f7f8;
  font-size: 13px;
  color: #c9ccd0;
}

.material-card .iconfont {
  font-size: 34px;
  color: #dcdfe3;
}

@media (max-width: 1280px) {
  .activity-row {
    flex-direction: column;
  }

  .carousel-box {
    flex: none;
    max-width: 100%;
  }

  .notice-box {
    margin: 16px 0 0;
  }

  .data-card {
    flex-direction: column;
  }

  .data-right {
    flex: none;
  }

  .comment-wrap {
    flex-direction: column;
  }

  .rank-content {
    flex-direction: column;
  }

  .give-share {
    flex: none;
    margin-top: 24px;
  }

  .material-col {
    flex: 0 0 50%;
    max-width: 50%;
  }
}
</style>
