<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getMyInfo } from '../../api/user'
import { manuscriptApi, normalizeManuscriptList } from '../../api/manuscript'
import { formatTenThousand } from '../../utils/format'
import noface from '../../assets/noface.gif'

const router = useRouter()
const userInfo = ref<any>(null)
const stats = ref<Record<string, number>>({})
const manuscripts = ref<any[]>([])
const dismissTip = ref(false)

const manuscriptCount = computed(() => Object.values(stats.value).reduce((sum, n) => sum + Number(n || 0), 0))
const totalViews = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.viewCount || 0), 0))
const totalDanmaku = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.danmakuCount || 0), 0))
const totalComments = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.commentCount || 0), 0))
const totalLikes = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.likeCount || 0), 0))
const totalShares = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.shareCount || 0), 0))
const totalCollections = computed(() => manuscripts.value.reduce((sum, item) => sum + Number(item.collectCount || 0), 0))

const services = [
  { label: '稿件管理', icon: 'video', to: '/m/space/manuscripts' },
  { label: '互动管理', icon: 'chat' },
  { label: '粉丝管理', icon: 'fans' },
  { label: '收益中心', icon: 'money' },
  { label: '任务中心', icon: 'task' },
  { label: '流量奖励', icon: 'fire' },
  { label: '创作激励', icon: 'reward' },
  { label: '更多功能', icon: 'more' }
]

onMounted(async () => {
  const [userRes, statsRes, listRes] = await Promise.all([
    getMyInfo(),
    manuscriptApi.getMyStats().catch(() => ({})),
    manuscriptApi.getMyManuscripts({ page: 1, size: 100 }).catch(() => ({}))
  ])
  if (userRes.code === '1') userInfo.value = userRes.data
  stats.value = statsRes?.data || statsRes || {}
  manuscripts.value = normalizeManuscriptList(listRes).list
})

const goBack = () => router.back()
const goService = (service: any) => {
  if (service.to) router.push(service.to)
}
const goPublish = () => {
  const origin = window.location.origin.replace(':5174', ':5173')
  window.location.href = `${origin}/create-center/upload`
}
</script>

<template>
  <div class="creator-center-page">
    <section class="hero">
      <header class="creator-topbar">
        <button @click="goBack" aria-label="返回">
          <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
        </button>
        <h1>创作中心</h1>
        <button aria-label="任务">
          <svg viewBox="0 0 24 24"><path d="M7 3h11v18H7z" /><path d="M4 7h3M4 12h3M4 17h3M11 8h4M11 13h4" /></svg>
        </button>
      </header>

      <div class="decor-icons">
        <span class="bili">⌁</span>
        <span class="like">♡</span>
        <span class="star">☆</span>
      </div>
    </section>

    <main class="creator-body">
      <section class="profile-card">
        <div class="profile-row">
          <img :src="userInfo?.avatar || noface" alt="" />
          <h2>{{ userInfo?.nickname || userInfo?.username || '创作者' }}</h2>
        </div>
        <div class="benefit-row">
          <span class="active">权益中心</span>
          <span>合集</span>
          <span>评论精选</span>
          <span>关闭互动</span>
          <button>3/18项权益 <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></button>
        </div>
        <div class="metric-row">
          <div>
            <strong>{{ formatTenThousand(userInfo?.followerCount || 0) }}</strong>
            <span>粉丝量</span>
          </div>
          <div>
            <strong>{{ formatTenThousand(totalViews) }}</strong>
            <span>播放量</span>
          </div>
          <div>
            <strong>{{ formatTenThousand(totalDanmaku) }}</strong>
            <span>弹幕量</span>
          </div>
          <div>
            <strong>{{ manuscriptCount }}</strong>
            <span>稿件数</span>
          </div>
        </div>
      </section>

      <section v-if="!dismissTip" class="tip-card">
        <span>📣</span>
        <button @click="goPublish">体验花生创作，带水印投稿，最高赢10万积分&gt;&gt;</button>
        <i @click="dismissTip = true">×</i>
      </section>

      <section class="panel-card">
        <div class="panel-title">
          <h2>常用服务</h2>
          <div class="panel-corner">›</div>
        </div>
        <div class="service-grid">
          <button v-for="service in services" :key="service.label" @click="goService(service)">
            <span :class="['service-icon', service.icon]">
              <svg v-if="service.icon === 'video'" viewBox="0 0 24 24"><path d="M4 7h16v12H4z" /><path d="M10 10l5 3-5 3z" /></svg>
              <svg v-else-if="service.icon === 'chat'" viewBox="0 0 24 24"><path d="M4 5h16v12H8l-4 4z" /></svg>
              <svg v-else-if="service.icon === 'fans'" viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.5" /></svg>
              <svg v-else-if="service.icon === 'money'" viewBox="0 0 24 24"><path d="M12 2v20M17 6H9a3 3 0 0 0 0 6h6a3 3 0 0 1 0 6H7" /></svg>
              <svg v-else-if="service.icon === 'task'" viewBox="0 0 24 24"><path d="M8 4h10v16H6V4h2z" /><path d="M9 9h6M9 14h6" /></svg>
              <svg v-else-if="service.icon === 'fire'" viewBox="0 0 24 24"><path d="M12 22c4 0 7-3 7-7 0-5-4-7-5-12-3 3-5 6-5 9-1-1-2-2-2-4-2 2-3 4-3 7 0 4 4 7 8 7z" /></svg>
              <svg v-else-if="service.icon === 'reward'" viewBox="0 0 24 24"><path d="M12 2l3 7h7l-5.5 4.5L18 21l-6-4-6 4 1.5-7.5L2 9h7z" /></svg>
              <svg v-else viewBox="0 0 24 24"><circle cx="5" cy="12" r="2" /><circle cx="12" cy="12" r="2" /><circle cx="19" cy="12" r="2" /></svg>
            </span>
            <b>{{ service.label }}</b>
          </button>
        </div>
      </section>

      <section class="panel-card data-panel">
        <div class="panel-title blue">
          <h2>近30日数据 <small>更新至今天</small></h2>
          <div class="panel-corner">›</div>
        </div>
        <div class="data-grid">
          <div><span>涨粉</span><strong>{{ formatTenThousand(userInfo?.followerCount || 0) }}</strong><em>昨日 -</em></div>
          <div><span>视频播放</span><strong>{{ formatTenThousand(totalViews) }}</strong><em>昨日 {{ manuscripts[0]?.viewCount || '-' }}</em></div>
          <div><span>评论</span><strong>{{ formatTenThousand(totalComments) }}</strong><em>昨日 -</em></div>
          <div><span>弹幕</span><strong>{{ formatTenThousand(totalDanmaku) }}</strong><em>昨日 -</em></div>
          <div><span>收益</span><strong>0.00</strong><em>昨日 0.00</em></div>
          <div><span>点赞</span><strong>{{ formatTenThousand(totalLikes) }}</strong><em>昨日 -</em></div>
          <div><span>分享</span><strong>{{ formatTenThousand(totalShares) }}</strong><em>昨日 -</em></div>
          <div><span>专栏浏览</span><strong>{{ formatTenThousand(totalCollections) }}</strong><em>昨日 -</em></div>
        </div>
        <div class="data-tip">📣 近30天涨粉超过同类UP主</div>
      </section>
    </main>
  </div>
</template>

<style scoped lang="scss">
.creator-center-page {
  min-height: 100vh;
  background: #f6f7f9;
  color: #18191c;
}

.hero {
  position: relative;
  height: 190px;
  overflow: hidden;
  background:
    radial-gradient(circle at 86% 60%, rgba(255, 195, 244, 0.95), transparent 28%),
    radial-gradient(circle at 30% 10%, rgba(131, 168, 255, 0.8), transparent 38%),
    linear-gradient(128deg, #81c5ff 0%, #9cb7ff 42%, #f7c9ef 100%);
}

.creator-topbar {
  position: relative;
  z-index: 2;
  height: 82px;
  padding: 30px 18px 0;
  display: grid;
  grid-template-columns: 42px 1fr 42px;
  align-items: center;
  box-sizing: border-box;

  h1 {
    margin: 0;
    text-align: center;
    font-size: 24px;
    font-weight: 700;
  }

  button {
    border: 0;
    padding: 0;
    background: transparent;
    color: #61666d;
  }

  svg {
    width: 34px;
    height: 34px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.1;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.decor-icons {
  position: absolute;
  left: 70px;
  top: 72px;
  color: #fff;
  font-size: 40px;
  text-shadow: 0 0 12px rgba(255,255,255,0.65);

  span {
    position: absolute;
  }

  .bili {
    left: 44px;
    top: 0;
    font-size: 58px;
  }

  .like {
    left: 118px;
    top: 48px;
    color: #ffe175;
  }

  .star {
    left: 138px;
    top: 2px;
    color: #ffe175;
  }
}

.creator-body {
  position: relative;
  z-index: 3;
  margin-top: -38px;
  padding: 0 14px 24px;
}

.profile-card,
.panel-card,
.tip-card {
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 4px 14px rgba(0,0,0,0.04);
}

.profile-card {
  padding: 16px 18px 22px;
}

.profile-row {
  display: flex;
  align-items: center;
  gap: 14px;

  img {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    object-fit: cover;
  }

  h2 {
    margin: 0;
    font-size: 25px;
    font-weight: 700;
  }
}

.benefit-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
  overflow: hidden;

  span {
    flex: 0 0 auto;
    padding: 7px 12px;
    border-radius: 10px;
    background: #f6f7f9;
    color: #61666d;
    font-size: 16px;

    &.active {
      color: #8a5c27;
      background: #ffe3bd;
    }
  }

  button {
    margin-left: auto;
    border: 0;
    display: inline-flex;
    align-items: center;
    white-space: nowrap;
    background: transparent;
    color: #61666d;
    font-size: 16px;

    svg {
      width: 18px;
      height: 18px;
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
    }
  }
}

.metric-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  margin-top: 26px;
  text-align: center;

  strong {
    display: block;
    color: #18191c;
    font-size: 24px;
    font-weight: 800;
  }

  span {
    display: block;
    margin-top: 6px;
    color: #61666d;
    font-size: 15px;
  }
}

.tip-card {
  height: 48px;
  margin-top: 14px;
  padding: 0 14px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 8px;

  button {
    min-width: 0;
    border: 0;
    padding: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
    background: transparent;
    color: #d99100;
    font-size: 16px;
    font-weight: 600;
  }

  i {
    color: #9499a0;
    font-size: 28px;
    font-style: normal;
  }
}

.panel-card {
  margin-top: 18px;
  padding: 18px 18px 20px;
  overflow: hidden;
}

.panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;

  h2 {
    margin: 0;
    font-size: 26px;
    font-weight: 800;

    small {
      margin-left: 10px;
      color: #9499a0;
      font-size: 17px;
      font-weight: 500;
    }
  }
}

.panel-corner {
  min-width: 48px;
  height: 32px;
  border-radius: 20px 0 0 20px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(120deg, #ff82ba, #fb3b91);
  font-size: 30px;
  transform: translateX(18px);
}

.panel-title.blue .panel-corner {
  background: linear-gradient(120deg, #2bc8ff, #188edc);
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 28px 12px;

  button {
    min-width: 0;
    border: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    background: transparent;
  }

  b {
    color: #18191c;
    font-size: 16px;
    font-weight: 500;
  }
}

.service-icon {
  width: 46px;
  height: 46px;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 14px;
  background: linear-gradient(160deg, #ff91c0, #fb7299);

  &.task {
    background: linear-gradient(160deg, #ffbf72, #ff8a3d);
  }

  &.fire {
    background: linear-gradient(160deg, #ffad44, #ff5d34);
  }

  &.reward,
  &.money {
    background: linear-gradient(160deg, #a993ff, #795cff);
  }

  &.more {
    background: linear-gradient(160deg, #a789ff, #7d5cff);
  }

  svg {
    width: 29px;
    height: 29px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  row-gap: 24px;
  text-align: center;

  div {
    border-right: 1px solid #f1f2f3;

    &:nth-child(4n) {
      border-right: 0;
    }
  }

  span {
    display: block;
    color: #9499a0;
    font-size: 16px;
  }

  strong {
    display: block;
    margin-top: 10px;
    color: #18191c;
    font-size: 23px;
    font-weight: 800;
  }

  em {
    display: block;
    margin-top: 5px;
    color: #61666d;
    font-size: 14px;
    font-style: normal;
  }
}

.data-tip {
  height: 44px;
  margin-top: 24px;
  padding: 0 14px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  color: #61666d;
  background: #f4f7ff;
  font-size: 16px;
}

@media (max-width: 390px) {
  .benefit-row span:nth-of-type(n + 4) {
    display: none;
  }

  .panel-title h2 {
    font-size: 23px;
  }

  .metric-row strong,
  .data-grid strong {
    font-size: 20px;
  }
}
</style>
