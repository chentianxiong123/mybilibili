<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUserInfo, getUserVideos } from '../../api/up-user'
import { checkFollow, followUser } from '../../api/interaction'
import { getFavoriteFolders } from '../../api/favorite'
import dynamicApi from '../../api/dynamic'
import { getLocalUserId } from '../../api/user'
import { formatDate, formatTenThousand } from '../../utils/format'
import noface from '../../assets/noface.gif'

type SpaceTab = 'home' | 'dynamic' | 'archive' | 'favorite' | 'bangumi'

const route = useRoute()
const router = useRouter()
const mId = computed(() => Number(route.params.mId || 0))
const upUser = ref<any>(null)
const videos = ref<any[]>([])
const dynamics = ref<any[]>([])
const favoriteFolders = ref<any[]>([])
const loading = ref(true)
const activeTab = ref<SpaceTab>((route.query.tab as SpaceTab) || 'home')
const isFollowing = ref(false)
const showAllArchives = ref(false)
const sortLabel = ref('最新发布')

const tabs = [
  { key: 'home' as SpaceTab, label: '主页' },
  { key: 'dynamic' as SpaceTab, label: '动态' },
  { key: 'archive' as SpaceTab, label: '投稿' },
  { key: 'favorite' as SpaceTab, label: '收藏' },
  { key: 'bangumi' as SpaceTab, label: '追番' }
]

const archiveFilters = ['视频', '图文', '扫雷', '教程']
const isSelf = computed(() => Number(getLocalUserId()) === Number(mId.value))
const displayName = computed(() => upUser.value?.name || '哔哩哔哩用户')
const signature = computed(() => upUser.value?.signature || upUser.value?.sign || '这个人暂时还没有填写简介')
const homeVideos = computed(() => videos.value.slice(0, 4))
const archiveVideos = computed(() => showAllArchives.value ? videos.value : videos.value.slice(0, 20))
const levelProgress = computed(() => {
  const exp = Number(upUser.value?.experience || 0)
  return Math.max(8, Math.min(96, Math.round((exp % 28800) / 288)))
})

const loadSpace = async () => {
  if (!mId.value) return
  loading.value = true
  try {
    const [userRes, videoRes] = await Promise.all([
      getUserInfo(mId.value),
      getUserVideos(mId.value, 1, 50)
    ])
    if (userRes.code === '1') upUser.value = userRes.data
    if (videoRes.code === '1') videos.value = videoRes.data || []
    await Promise.all([
      loadFollowState(),
      loadDynamics(),
      loadFavorites()
    ])
  } finally {
    loading.value = false
  }
}

const loadFollowState = async () => {
  if (isSelf.value || !mId.value) return
  const res = await checkFollow(mId.value)
  isFollowing.value = typeof res.data === 'object' ? !!res.data?.following : !!res.data
}

const loadDynamics = async () => {
  try {
    const res = await dynamicApi.getUserDynamics(mId.value, 1, 10)
    if (res?.code === 200) {
      dynamics.value = res.data?.list || res.data || []
    }
  } catch (e) {
    dynamics.value = []
  }
}

const loadFavorites = async () => {
  if (!isSelf.value) return
  const res = await getFavoriteFolders()
  if (res.code === '1') favoriteFolders.value = res.data || []
}

const switchTab = (key: SpaceTab) => {
  activeTab.value = key
  router.replace({ path: route.path, query: key === 'home' ? {} : { tab: key } })
}

const goBack = () => router.back()
const goSearch = () => router.push('/m/search')
const goEdit = () => router.push('/m/space/profile/edit')
const goFriends = (tab: 'following' | 'followers') => router.push(`/m/space/${mId.value}/friends?tab=${tab}`)
const goVideo = (id: number) => router.push(`/m/video/${id}`)

const toggleFollow = async () => {
  if (isSelf.value) return
  if (!getLocalUserId()) {
    router.push('/m/login')
    return
  }
  const next = !isFollowing.value
  const res = await followUser(mId.value, next)
  if (res.code === '1') {
    isFollowing.value = next
    if (upUser.value) {
      upUser.value.follower = Math.max(0, Number(upUser.value.follower || 0) + (next ? 1 : -1))
    }
  }
}

const formatUploadDate = (value: string) => {
  if (!value) return ''
  return formatDate(value)
}

const formatDuration = (value: any) => {
  if (!value) return ''
  if (typeof value === 'string' && value.includes(':')) return value
  const n = Number(value)
  if (!Number.isFinite(n)) return ''
  const min = Math.floor(n / 60)
  const sec = Math.floor(n % 60)
  return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`
}

onMounted(loadSpace)

watch(() => route.params.mId, loadSpace)
watch(() => route.query.tab, (tab) => {
  activeTab.value = (tab as SpaceTab) || 'home'
})
</script>

<template>
  <div class="space-profile-page">
    <header :class="['profile-topbar', { compact: activeTab !== 'home' }]">
      <button class="icon-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1 v-if="activeTab !== 'home'">{{ displayName }}</h1>
      <div class="top-actions">
        <button class="icon-btn" @click="goSearch" aria-label="搜索">
          <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></svg>
        </button>
        <button class="icon-btn" aria-label="更多">
          <svg viewBox="0 0 24 24" class="fill-icon"><circle cx="12" cy="5" r="1.6" /><circle cx="12" cy="12" r="1.6" /><circle cx="12" cy="19" r="1.6" /></svg>
        </button>
      </div>
    </header>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-else-if="upUser">
      <section v-if="activeTab === 'home'" class="hero-section">
        <div class="cover-art"></div>
        <div class="profile-card">
          <div class="avatar-stats">
            <img :src="upUser.face || noface" class="big-avatar" alt="" />
            <div class="stats-cluster">
              <button @click="goFriends('followers')">
                <strong>{{ formatTenThousand(upUser.follower || 0) }}</strong>
                <span>粉丝</span>
              </button>
              <i></i>
              <button @click="goFriends('following')">
                <strong>{{ formatTenThousand(upUser.following || 0) }}</strong>
                <span>关注</span>
              </button>
              <i></i>
              <button>
                <strong>{{ formatTenThousand(upUser.totalLikeCount || 0) }}</strong>
                <span>获赞</span>
              </button>
            </div>
          </div>

          <button v-if="isSelf" class="edit-profile-btn" @click="goEdit">编辑资料</button>
          <button v-else :class="['edit-profile-btn', { following: isFollowing }]" @click="toggleFollow">
            {{ isFollowing ? '已关注' : '+ 关注' }}
          </button>

          <div class="identity-block">
            <div class="name-row">
              <h2>{{ displayName }}</h2>
              <span class="vip-badge">大会员</span>
              <span class="fans-badge">粉丝勋章</span>
            </div>
            <div class="level-row">
              <span class="level-pill">LV{{ upUser.level || 1 }}</span>
              <div class="level-track"><span :style="{ width: levelProgress + '%' }"></span></div>
              <span class="exp-text">{{ upUser.experience || 0 }}/28800</span>
            </div>
            <p class="signature">{{ signature }}</p>
            <div class="meta-row">
              <span>IP属地：广东</span>
              <button>+ 添加学校信息</button>
              <a>详情</a>
            </div>
          </div>
        </div>
      </section>

      <nav class="space-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
        </button>
      </nav>

      <main class="tab-content">
        <section v-if="activeTab === 'home'" class="home-panel">
          <div class="section-heading">
            <h3>视频 <span>{{ videos.length }}</span></h3>
            <button @click="switchTab('archive')">查看更多 <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></button>
          </div>
          <div v-if="homeVideos.length" class="home-video-grid">
            <article v-for="video in homeVideos" :key="video.aId" class="home-video-card" @click="goVideo(video.aId)">
              <div class="cover-wrap">
                <img :src="video.pic" alt="" />
                <div class="cover-stats">
                  <span>▶ {{ formatTenThousand(video.play || 0) }}</span>
                  <span>▣ {{ formatTenThousand(video.videoReview || 0) }}</span>
                  <b>{{ formatDuration(video.duration) }}</b>
                </div>
              </div>
              <h4>{{ video.title }}</h4>
              <p v-if="isSelf" class="privacy">🔒 仅自己可见</p>
            </article>
          </div>
          <div v-else class="empty">这里还没有视频</div>

          <div class="section-heading favorite-heading">
            <h3>图文 <span>0</span></h3>
            <button>查看更多 <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></button>
          </div>
          <div class="text-card">新年穿新鞋，多威战神4Pro 薄底竞速碳板</div>
        </section>

        <section v-else-if="activeTab === 'dynamic'" class="dynamic-panel">
          <article v-if="dynamics.length === 0" class="empty-card">暂无动态</article>
          <article v-for="item in dynamics" v-else :key="item.id" class="dynamic-card">
            <div class="dynamic-author">
              <img :src="upUser.face || noface" alt="" />
              <div>
                <strong>{{ displayName }}</strong>
                <span>{{ formatUploadDate(item.createdAt) }}</span>
              </div>
              <button>置顶</button>
            </div>
            <p class="dynamic-text">{{ item.content || '转发动态' }}</p>
            <div v-if="item.imageUrls?.length" class="dynamic-images">
              <img v-for="url in item.imageUrls.slice(0, 3)" :key="url" :src="url" alt="" />
            </div>
            <div class="dynamic-actions">
              <span>↗ 转发</span>
              <span>○ 评论</span>
              <span class="liked">♥ {{ item.likeCount || 0 }}</span>
            </div>
          </article>
        </section>

        <section v-else-if="activeTab === 'archive'" class="archive-panel">
          <div class="archive-filter">
            <button v-for="filter in archiveFilters" :key="filter" :class="{ active: filter === '视频' }">{{ filter }}</button>
            <button class="all-filter">全部 <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></button>
          </div>
          <div class="archive-toolbar">
            <button class="play-all">
              <svg viewBox="0 0 24 24"><polygon points="8 5 19 12 8 19" /></svg>
              播放全部
            </button>
            <button class="sort-btn">
              <svg viewBox="0 0 24 24"><path d="M4 7h13M4 12h10M4 17h7" /></svg>
              {{ sortLabel }}
            </button>
          </div>
          <div v-if="archiveVideos.length" class="archive-list">
            <article v-for="video in archiveVideos" :key="video.aId" class="archive-row" @click="goVideo(video.aId)">
              <div class="archive-cover">
                <img :src="video.pic" alt="" />
                <span v-if="isSelf" class="private-tag">🔒 仅自己可见</span>
                <b>{{ formatDuration(video.duration) }}</b>
              </div>
              <div class="archive-info">
                <h3>{{ video.title }}</h3>
                <time>{{ formatUploadDate(video.uploadTime) }}</time>
                <div class="archive-stats">
                  <span>▶ {{ formatTenThousand(video.play || 0) }}</span>
                  <span>▣ {{ formatTenThousand(video.videoReview || 0) }}</span>
                </div>
              </div>
              <button class="row-menu" aria-label="更多">
                <svg viewBox="0 0 24 24"><circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" /></svg>
              </button>
            </article>
          </div>
          <div v-else class="empty">这里还没有投稿</div>
        </section>

        <section v-else-if="activeTab === 'favorite'" class="favorite-panel">
          <div class="favorite-title">⌃ 我创建的收藏夹 <span>{{ favoriteFolders.length }}</span></div>
          <article v-for="folder in favoriteFolders" :key="folder.id" class="folder-row">
            <div class="folder-cover">
              <img :src="folder.coverUrl || videos[0]?.pic || noface" alt="" />
              <span>▶</span>
            </div>
            <div>
              <h3>{{ folder.name || '默认收藏夹' }}</h3>
              <p>{{ folder.videoCount || folder.count || 0 }}个内容 · {{ folder.visibility || '公开' }}</p>
            </div>
          </article>
          <div v-if="favoriteFolders.length === 0" class="empty">暂无公开收藏夹</div>
        </section>

        <section v-else class="empty-card">追番功能正在准备中</section>
      </main>
    </template>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.space-profile-page {
  min-height: 100vh;
  background: #f6f7f9;
  color: #18191c;
  padding-bottom: calc(14px + env(safe-area-inset-bottom));
}

.profile-topbar {
  position: sticky;
  top: 0;
  z-index: 50;
  height: 72px;
  padding: 18px 12px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(10px);

  h1 {
    position: absolute;
    left: 72px;
    right: 92px;
    margin: 0;
    text-align: center;
    font-size: 20px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.top-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 38px;
  height: 38px;
  border: 0;
  padding: 0;
  display: grid;
  place-items: center;
  background: transparent;
  color: #61666d;

  svg {
    width: 29px;
    height: 29px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .fill-icon {
    fill: currentColor;
    stroke: none;
  }
}

.hero-section {
  margin-top: -72px;
}

.cover-art {
  height: 190px;
  background:
    linear-gradient(rgba(255,255,255,0.08), rgba(246,247,249,0.98)),
    radial-gradient(circle at 18% 42%, rgba(157,190,235,0.72), transparent 32%),
    radial-gradient(circle at 75% 30%, rgba(222,229,240,0.9), transparent 36%),
    #eef3fb;
}

.profile-card {
  position: relative;
  margin-top: -52px;
  padding: 0 14px 14px;
}

.avatar-stats {
  display: flex;
  align-items: flex-end;
  gap: 16px;
}

.big-avatar {
  width: 92px;
  height: 92px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #fff;
  background: #e3e5e7;
}

.stats-cluster {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1px 1fr 1px 1fr;
  align-items: center;
  padding-bottom: 2px;

  button {
    border: 0;
    background: transparent;
    color: #9499a0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
  }

  strong {
    color: #18191c;
    font-size: 20px;
    font-weight: 600;
  }

  span {
    font-size: 13px;
  }

  i {
    width: 1px;
    height: 30px;
    background: #e3e5e7;
  }
}

.edit-profile-btn {
  width: calc(100% - 112px);
  height: 38px;
  margin: 12px 0 0 112px;
  border: 1px solid #ff6aa2;
  border-radius: 5px;
  background: transparent;
  color: #ff6aa2;
  font-size: 17px;
  font-weight: 600;

  &.following {
    border-color: #d3d6dc;
    color: #9499a0;
  }
}

.identity-block {
  margin-top: 18px;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;

  h2 {
    margin: 0;
    max-width: 42vw;
    color: #18191c;
    font-size: 28px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.vip-badge,
.fans-badge,
.level-pill {
  flex: 0 0 auto;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 12px;
  color: #fff;
  background: #9da1aa;
}

.fans-badge {
  color: #9ca0a8;
  background: #f1f2f3;
}

.level-row {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 12px;
  color: #8f939c;
}

.level-pill {
  padding: 1px 5px;
  border-radius: 2px;
  background: #ff7b2d;
  font-weight: 700;
}

.level-track {
  width: 115px;
  height: 3px;
  border-radius: 3px;
  background: #e3e5e7;

  span {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: #ff7b2d;
  }
}

.exp-text {
  font-size: 13px;
}

.signature {
  min-height: 20px;
  margin: 16px 0 0;
  color: #61666d;
  font-size: 15px;
  line-height: 1.45;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 14px;
  color: #9499a0;
  font-size: 13px;

  button {
    border: 1px dashed #d3d6dc;
    border-radius: 14px;
    padding: 2px 8px;
    background: transparent;
    color: #9499a0;
  }

  a {
    margin-left: auto;
    color: #18a7c7;
  }
}

.space-tabs {
  position: sticky;
  top: 72px;
  z-index: 45;
  height: 50px;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  background: #fff;
  border-top: 1px solid #f1f2f3;
  border-bottom: 1px solid #f1f2f3;

  button {
    position: relative;
    border: 0;
    background: transparent;
    color: #61666d;
    font-size: 18px;
    font-weight: 600;

    &.active {
      color: #ff6aa2;

      &::after {
        content: '';
        position: absolute;
        left: 50%;
        bottom: 0;
        width: 42px;
        height: 3px;
        border-radius: 3px;
        transform: translateX(-50%);
        background: #ff6aa2;
      }
    }
  }
}

.tab-content {
  background: #f6f7f9;
}

.home-panel {
  padding: 16px 14px 24px;
}

.section-heading {
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;

  h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;

    span {
      margin-left: 5px;
      color: #82868f;
      font-size: 16px;
      font-weight: 500;
    }
  }

  button {
    border: 0;
    background: transparent;
    color: #8f939c;
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-size: 15px;

    svg {
      width: 20px;
      height: 20px;
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
    }
  }
}

.home-video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px 12px;
}

.home-video-card {
  background: #fff;

  h4 {
    height: 43px;
    margin: 8px 8px 3px;
    color: #18191c;
    font-size: 16px;
    font-weight: 500;
    line-height: 1.35;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .privacy {
    margin: 0 8px 8px;
    color: #9499a0;
    font-size: 14px;
  }
}

.cover-wrap {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: #e3e5e7;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.cover-stats {
  position: absolute;
  inset: auto 0 0;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 20px 7px 6px;
  background: linear-gradient(to top, rgba(0,0,0,0.8), transparent);
  color: #fff;
  font-size: 13px;

  b {
    margin-left: auto;
    font-weight: 500;
  }
}

.favorite-heading {
  margin-top: 26px;
}

.text-card {
  min-height: 52px;
  padding: 14px;
  background: #fff;
  color: #18191c;
  font-size: 18px;
}

.archive-panel {
  padding-bottom: 20px;
}

.archive-filter {
  position: sticky;
  top: 122px;
  z-index: 30;
  display: flex;
  gap: 12px;
  padding: 10px 14px;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  button {
    min-width: 66px;
    height: 34px;
    border: 1px solid #e3e5e7;
    border-radius: 5px;
    background: transparent;
    color: #61666d;
    font-size: 16px;

    &.active {
      border-color: rgba(255,106,162,0.16);
      background: rgba(255,106,162,0.14);
      color: #ff6aa2;
    }
  }

  .all-filter {
    margin-left: auto;
    min-width: auto;
    border-color: transparent;
    color: #9499a0;
    display: inline-flex;
    align-items: center;

    svg {
      width: 18px;
      height: 18px;
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
    }
  }
}

.archive-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px;

  button {
    border: 0;
    background: transparent;
    color: #9499a0;
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-size: 16px;
  }

  .play-all {
    color: #18191c;
    font-size: 20px;
    font-weight: 600;

    svg {
      width: 30px;
      height: 30px;
      border: 2px solid #ff6aa2;
      border-radius: 50%;
      fill: #ff6aa2;
      stroke: #ff6aa2;
    }
  }

  svg {
    width: 19px;
    height: 19px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }
}

.archive-row {
  display: grid;
  grid-template-columns: 154px 1fr 26px;
  gap: 12px;
  padding: 10px 14px 14px;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;
}

.archive-cover {
  position: relative;
  height: 88px;
  overflow: hidden;
  border-radius: 4px;
  background: #e3e5e7;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .private-tag,
  b {
    position: absolute;
    bottom: 6px;
    color: #fff;
    font-size: 14px;
    font-weight: 500;
  }

  .private-tag {
    left: 8px;
    padding: 1px 5px;
    border-radius: 3px;
    background: rgba(0,0,0,0.42);
  }

  b {
    right: 7px;
  }
}

.archive-info {
  min-width: 0;

  h3 {
    height: 46px;
    margin: 1px 0 7px;
    color: #18191c;
    font-size: 18px;
    font-weight: 500;
    line-height: 1.28;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  time {
    color: #9499a0;
    font-size: 14px;
  }
}

.archive-stats {
  display: flex;
  gap: 42px;
  margin-top: 9px;
  color: #9499a0;
  font-size: 14px;
}

.row-menu {
  border: 0;
  background: transparent;
  color: #9499a0;
  align-self: center;

  svg {
    width: 22px;
    height: 22px;
    fill: currentColor;
  }
}

.dynamic-panel,
.favorite-panel {
  padding-bottom: 20px;
}

.dynamic-card,
.empty-card {
  background: #fff;
  border-bottom: 1px solid #f1f2f3;
  padding: 16px 14px;
}

.dynamic-author {
  display: flex;
  align-items: center;
  gap: 10px;

  img {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    object-fit: cover;
  }

  div {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  strong {
    color: #18191c;
    font-size: 18px;
  }

  span {
    color: #9499a0;
    font-size: 14px;
  }

  button {
    border: 0;
    border-radius: 3px;
    padding: 3px 8px;
    background: rgba(255,106,162,0.14);
    color: #ff6aa2;
  }
}

.dynamic-text {
  margin: 17px 0;
  color: #18191c;
  font-size: 20px;
  line-height: 1.45;
}

.dynamic-images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 5px;

  img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    border-radius: 4px;
  }
}

.dynamic-actions {
  display: flex;
  justify-content: space-around;
  margin-top: 16px;
  color: #9499a0;
  font-size: 16px;

  .liked {
    color: #ff6aa2;
  }
}

.favorite-title {
  padding: 17px 14px;
  color: #18191c;
  font-size: 18px;
  font-weight: 600;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  span {
    color: #9499a0;
    font-weight: 500;
  }
}

.folder-row {
  display: grid;
  grid-template-columns: 138px 1fr;
  gap: 14px;
  padding: 14px;
  background: #fff;

  h3 {
    margin: 5px 0 28px;
    color: #18191c;
    font-size: 20px;
    font-weight: 600;
  }

  p {
    margin: 0;
    color: #9499a0;
    font-size: 18px;
  }
}

.folder-cover {
  position: relative;
  height: 78px;
  overflow: hidden;
  border-radius: 4px;
  background: #e3e5e7;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  span {
    position: absolute;
    right: 6px;
    bottom: 6px;
    color: #fff;
  }
}

.loading,
.empty {
  padding: 60px 16px;
  color: #9499a0;
  text-align: center;
}

.empty-card {
  color: #9499a0;
  text-align: center;
}
</style>
