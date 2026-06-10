<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { followUser } from '../../api/interaction'
import { getFollowerList, getFollowingList, getLocalUserId } from '../../api/user'
import { formatTenThousand } from '../../utils/format'
import noface from '../../assets/noface.gif'

type FriendTab = 'following' | 'followers'

const route = useRoute()
const router = useRouter()
const userId = computed(() => Number(route.params.mId || getLocalUserId() || 0))
const activeTab = ref<FriendTab>(route.query.tab === 'followers' ? 'followers' : 'following')
const keyword = ref('')
const loading = ref(false)
const followingList = ref<any[]>([])
const followerList = ref<any[]>([])

const tabs = [
  { key: 'following' as FriendTab, label: '关注' },
  { key: 'followers' as FriendTab, label: '粉丝' }
]

const currentList = computed(() => activeTab.value === 'following' ? followingList.value : followerList.value)
const filteredList = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return currentList.value
  return currentList.value.filter((u) => getDisplayName(u).toLowerCase().includes(k))
})

const getDisplayName = (user: any) => user?.nickname || user?.username || `用户${user?.id || ''}`
const getAvatar = (user: any) => user?.avatar || noface
const getSign = (user: any) => user?.bio || user?.signature || '这个人暂时还没有填写简介'

const normalizeFollowingUsers = (list: any[]) => list.map((u) => ({
  ...u,
  following: true
}))

const loadData = async () => {
  if (!userId.value) return
  loading.value = true
  try {
    const [followingRes, followerRes] = await Promise.all([
      getFollowingList(userId.value),
      getFollowerList(userId.value)
    ])
    followingList.value = normalizeFollowingUsers(followingRes.data || [])
    followerList.value = (followerRes.data || []).map((u: any) => ({ ...u, following: false }))
    await markFollowerFollowingState()
  } finally {
    loading.value = false
  }
}

const markFollowerFollowingState = async () => {
  const myId = getLocalUserId()
  if (!myId || activeTab.value !== 'followers') return
  const followingIds = new Set(followingList.value.map((u) => Number(u.id)))
  followerList.value = followerList.value.map((u) => ({
    ...u,
    following: followingIds.has(Number(u.id))
  }))
}

const switchTab = (tab: FriendTab) => {
  activeTab.value = tab
  router.replace({ path: route.path, query: { tab } })
}

const handleFollow = async (user: any) => {
  const myId = getLocalUserId()
  if (!myId) {
    router.push('/m/login')
    return
  }
  if (Number(user.id) === Number(myId)) return
  const next = !user.following
  const res = await followUser(user.id, next)
  if (res.code === '1') {
    user.following = next
    if (next && !followingList.value.some((u) => Number(u.id) === Number(user.id))) {
      followingList.value.unshift({ ...user, following: true })
    }
  }
}

const goBack = () => router.back()
const goSpace = (id: number) => router.push(`/m/space/${id}`)

onMounted(loadData)

watch(() => route.query.tab, (tab) => {
  activeTab.value = tab === 'followers' ? 'followers' : 'following'
  markFollowerFollowingState()
})

watch(() => route.params.mId, () => {
  loadData()
})
</script>

<template>
  <div class="friends-page">
    <header class="friends-header">
      <button class="icon-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <button class="icon-btn close-btn" @click="goBack" aria-label="关闭">
        <svg viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12" /></svg>
      </button>
      <h1>我的好友</h1>
    </header>

    <nav class="top-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        :class="{ active: activeTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <section v-if="activeTab === 'following'" class="search-section">
      <div class="search-box">
        <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></svg>
        <input v-model="keyword" placeholder="搜索我的关注" />
      </div>
      <div class="list-summary">
        <span>我的关注 <b>{{ formatTenThousand(followingList.length) }}</b>人</span>
        <span class="sort-label">
          <svg viewBox="0 0 24 24"><path d="M4 7h13M4 12h10M4 17h7" /></svg>
          最近关注
        </span>
      </div>
      <div class="chips">
        <button class="active">全部</button>
        <button>特别关注</button>
        <button>未分组</button>
        <button class="fold">
          <svg viewBox="0 0 24 24"><path d="M6 9l6 6 6-6" /></svg>
        </button>
      </div>
    </section>

    <section v-else class="fans-summary">
      共 <b>{{ formatTenThousand(followerList.length) }}</b> 个粉丝
    </section>

    <main class="friend-list">
      <div v-if="loading" class="state">加载中...</div>
      <div v-else-if="filteredList.length === 0" class="state">暂无{{ activeTab === 'following' ? '关注' : '粉丝' }}</div>
      <article v-for="user in filteredList" v-else :key="user.id" class="friend-row">
        <img :src="getAvatar(user)" class="avatar" alt="" @click="goSpace(user.id)" />
        <div class="friend-main" @click="goSpace(user.id)">
          <h2>{{ getDisplayName(user) }}</h2>
          <p>{{ getSign(user) }}</p>
        </div>
        <button
          :class="['follow-state', { outline: activeTab === 'followers' && !user.following }]"
          @click="handleFollow(user)"
        >
          <svg v-if="activeTab === 'following' || user.following" viewBox="0 0 24 24"><path d="M4 8h16M4 12h16M4 16h16" /></svg>
          {{ activeTab === 'followers' && !user.following ? '回关' : '已关注' }}
        </button>
        <button v-if="activeTab === 'followers'" class="more-btn" aria-label="更多">
          <svg viewBox="0 0 24 24"><circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" /></svg>
        </button>
      </article>
    </main>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.friends-page {
  min-height: 100vh;
  background: #f6f7f9;
  color: #18191c;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
}

.friends-header {
  position: sticky;
  top: 0;
  z-index: 20;
  height: 76px;
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 18px 18px 0;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  h1 {
    margin: 0;
    font-size: 21px;
    font-weight: 700;
    letter-spacing: 0;
  }
}

.icon-btn,
.more-btn {
  width: 34px;
  height: 34px;
  border: 0;
  padding: 0;
  background: transparent;
  color: #61666d;
  display: grid;
  place-items: center;

  svg {
    width: 27px;
    height: 27px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.close-btn svg { width: 25px; height: 25px; }

.top-tabs {
  position: sticky;
  top: 76px;
  z-index: 19;
  display: grid;
  grid-template-columns: 1fr 1fr;
  height: 48px;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  button {
    position: relative;
    border: 0;
    background: transparent;
    color: #61666d;
    font-size: 17px;

    &.active {
      color: #ff6aa2;
      font-weight: 600;

      &::after {
        content: '';
        position: absolute;
        left: 50%;
        bottom: 0;
        width: 36px;
        height: 3px;
        transform: translateX(-50%);
        border-radius: 2px;
        background: #ff6aa2;
      }
    }
  }
}

.search-section {
  background: #fff;
  padding: 12px 12px 0;
}

.search-box {
  height: 38px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 0 13px;
  background: #f1f2f3;
  color: #9499a0;

  svg {
    width: 20px;
    height: 20px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  input {
    min-width: 0;
    flex: 1;
    border: 0;
    outline: 0;
    background: transparent;
    color: #18191c;
    font-size: 15px;
  }
}

.list-summary,
.fans-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 55px;
  font-size: 18px;
  color: #18191c;

  b {
    color: #9499a0;
    font-weight: 500;
  }
}

.fans-summary {
  padding: 0 14px;
  justify-content: flex-start;
  gap: 5px;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;
}

.sort-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #9499a0;
  font-size: 15px;

  svg {
    width: 18px;
    height: 18px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }
}

.chips {
  display: flex;
  align-items: center;
  gap: 14px;
  height: 46px;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  button {
    height: 32px;
    border: 0;
    border-radius: 0;
    padding: 0 10px;
    background: transparent;
    color: #61666d;
    font-size: 16px;

    &.active {
      color: #ff6aa2;
      background: rgba(255,106,162,0.12);
    }
  }

  .fold {
    margin-left: auto;
    padding-right: 2px;

    svg {
      width: 22px;
      height: 22px;
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
    }
  }
}

.friend-list {
  padding-left: 14px;
  background: #fff;
}

.friend-row {
  display: flex;
  align-items: center;
  min-height: 86px;
  gap: 13px;
  border-bottom: 1px solid #f1f2f3;
}

.avatar {
  width: 58px;
  height: 58px;
  border-radius: 50%;
  object-fit: cover;
  background: #e3e5e7;
  flex: 0 0 auto;
}

.friend-main {
  min-width: 0;
  flex: 1;

  h2 {
    margin: 0;
    color: #18191c;
    font-size: 18px;
    font-weight: 500;
    line-height: 1.35;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  p {
    margin: 4px 0 0;
    color: #9499a0;
    font-size: 14px;
    line-height: 1.35;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.follow-state {
  flex: 0 0 auto;
  height: 32px;
  min-width: 82px;
  border: 0;
  border-radius: 18px;
  padding: 0 12px;
  background: #f1f2f3;
  color: #9499a0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 15px;

  &.outline {
    border: 1px solid #ff6aa2;
    background: transparent;
    color: #ff6aa2;
  }

  svg {
    width: 17px;
    height: 17px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }
}

.more-btn {
  width: 34px;
  margin-right: 7px;
  color: #9499a0;

  svg {
    fill: currentColor;
    stroke: none;
  }
}

.state {
  padding: 72px 16px;
  text-align: center;
  color: #9499a0;
}
</style>
