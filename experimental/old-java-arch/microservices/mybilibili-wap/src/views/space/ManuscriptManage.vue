<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { manuscriptApi, normalizeManuscriptList } from '../../api/manuscript'
import { formatTenThousand } from '../../utils/format'

type ManageTab = 'video' | 'article' | 'audio'

const router = useRouter()
const activeTab = ref<ManageTab>('video')
const loading = ref(false)
const manuscripts = ref<any[]>([])
const stats = ref<Record<string, number>>({})
const page = ref(1)
const totalPages = ref(1)
const actionItem = ref<any>(null)
const toastText = ref('')

const tabs = computed(() => [
  { key: 'video' as ManageTab, label: '视频', count: stats.value.total ?? manuscripts.value.length },
  { key: 'article' as ManageTab, label: '图文', count: 0 },
  { key: 'audio' as ManageTab, label: '音频', count: 0 }
])

const videoList = computed(() => manuscripts.value)

const showToast = (text: string) => {
  toastText.value = text
  window.setTimeout(() => {
    if (toastText.value === text) toastText.value = ''
  }, 1800)
}

const loadStats = async () => {
  try {
    const res = await manuscriptApi.getMyStats()
    stats.value = res?.data || res || {}
  } catch (e) {
    stats.value = {}
  }
}

const loadManuscripts = async (nextPage = 1) => {
  loading.value = true
  try {
    const res = await manuscriptApi.getMyManuscripts({ page: nextPage, size: 20 })
    const data = normalizeManuscriptList(res)
    manuscripts.value = nextPage > 1 ? manuscripts.value.concat(data.list) : data.list
    page.value = data.page
    totalPages.value = data.totalPages
    stats.value = { ...stats.value, total: data.total }
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  await Promise.all([loadStats(), loadManuscripts(1)])
}

const goBack = () => router.back()
const goDrafts = () => showToast('草稿箱接口还没有接入')
const goVideo = (item: any) => router.push(`/m/video/${item.manuscriptId || item.id}`)
const goData = () => showToast('移动端数据面板待接入，先到 Web 创作中心查看')

const getWebUrl = (path: string) => {
  const origin = window.location.origin.replace(':5174', ':5173')
  return `${origin}${path}`
}

const goEdit = (item: any) => {
  const id = item.manuscriptId || item.id
  window.location.href = getWebUrl(`/create-center/content?manuscriptId=${id}`)
}

const goPublish = () => {
  window.location.href = getWebUrl('/create-center/upload')
}

const shareItem = async (item: any) => {
  const id = item.manuscriptId || item.id
  const url = `${window.location.origin}/wap/m/video/${id}`
  try {
    if (navigator.share) {
      await navigator.share({ title: item.title || '稿件', url })
      return
    }
    await navigator.clipboard?.writeText(url)
    showToast('链接已复制')
  } catch (e) {
    showToast('分享已取消')
  }
}

const togglePublishState = async (item: any) => {
  const id = item.manuscriptId || item.id
  const isPublished = Number(item.status) === 3
  try {
    if (isPublished) {
      await manuscriptApi.unpublish(id)
      showToast('已下架')
    } else {
      await manuscriptApi.publish(id)
      showToast('已提交上架')
    }
    actionItem.value = null
    await refresh()
  } catch (e) {
    showToast('操作失败')
  }
}

const removeItem = async (item: any) => {
  const id = item.manuscriptId || item.id
  if (!window.confirm('确定删除这个稿件吗？删除后不可恢复。')) return
  try {
    await manuscriptApi.remove(id)
    showToast('已删除')
    actionItem.value = null
    await refresh()
  } catch (e) {
    showToast('删除失败')
  }
}

const formatDateTime = (value: any) => {
  if (!value) return '-'
  const date = new Date(String(value).replace(/-/g, '/'))
  if (Number.isNaN(date.getTime())) return String(value)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const formatDurationText = (item: any) => {
  const value = item.duration || item.durationSeconds
  if (!value) return ''
  if (typeof value === 'string' && value.includes(':')) return value
  const seconds = Number(value)
  if (!Number.isFinite(seconds)) return ''
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

const statusText = (item: any) => {
  const status = Number(item.status)
  if (status === 3) return '已发布'
  if (status === -1) return '已下架'
  if (status === 0) return '审核中'
  if (status === 1) return '处理中'
  if (status === 2) return '待发布'
  if (status === 4) return '未通过'
  if (status === 5) return '处理失败'
  return '草稿'
}

const statusClass = (item: any) => {
  const status = Number(item.status)
  if (status === 3) return 'published'
  if (status === 4 || status === 5) return 'error'
  if (status === -1) return 'muted'
  return 'pending'
}

onMounted(refresh)
</script>

<template>
  <div class="manuscript-manage-page">
    <header class="manage-header">
      <button class="back-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1>稿件管理</h1>
      <button class="draft-btn" @click="goDrafts">草稿箱</button>
    </header>

    <nav class="manage-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        :class="{ active: activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="manage-content">
      <section v-if="activeTab === 'video'">
        <div v-if="loading" class="state-block">加载中...</div>
        <div v-else-if="videoList.length === 0" class="state-block">
          <p>还没有视频稿件</p>
          <button @click="goPublish">发布视频</button>
        </div>

        <article v-for="item in videoList" v-else :key="item.manuscriptId || item.id" class="manuscript-card">
          <div class="summary-row" @click="goVideo(item)">
            <div class="cover-box">
              <img :src="item.coverUrl || item.cover" alt="" />
              <span v-if="statusText(item) !== '已发布'" class="visibility-tag">{{ statusText(item) }}</span>
              <b v-if="formatDurationText(item)">{{ formatDurationText(item) }}</b>
            </div>
            <div class="summary-main">
              <h2>{{ item.title || '未命名稿件' }}</h2>
              <time>{{ formatDateTime(item.uploadTime || item.createdAt || item.updatedAt) }}</time>
              <span :class="['status-pill', statusClass(item)]">{{ statusText(item) }}</span>
            </div>
          </div>

          <div class="stats-row">
            <span>
              <svg viewBox="0 0 24 24"><polygon points="6 4 20 12 6 20" /></svg>
              {{ formatTenThousand(item.viewCount || 0) }}
            </span>
            <span>
              <svg viewBox="0 0 24 24"><path d="M7 11v10H4V11h3zm4 10h7.2a2 2 0 0 0 1.96-1.61l1.05-5.26A2 2 0 0 0 19.25 11H14l1-5a2 2 0 0 0-2-2h-.5L8 11v10h3z" /></svg>
              {{ formatTenThousand(item.likeCount || 0) }}
            </span>
            <span>
              <svg viewBox="0 0 24 24"><path d="M4 5h16v12H7l-3 3V5z" /></svg>
              {{ item.danmakuCount ?? '-' }}
            </span>
            <span>
              <svg viewBox="0 0 24 24"><path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" /></svg>
              {{ item.commentCount ?? '-' }}
            </span>
            <span>
              <svg viewBox="0 0 24 24"><path d="M18 8a3 3 0 1 0-2.83-4H15a3 3 0 0 0 .17 1L8.91 8.6a3 3 0 1 0 0 4.8l6.26 3.6A3 3 0 1 0 16 15.8l-6.26-3.6a3.1 3.1 0 0 0 0-.4l6.26-3.6A3 3 0 0 0 18 8z" /></svg>
              {{ item.shareCount ?? '-' }}
            </span>
          </div>

          <div class="action-row">
            <button @click="shareItem(item)">分享</button>
            <button @click="goData">数据</button>
            <button @click="goEdit(item)">编辑</button>
            <button @click="actionItem = item">更多</button>
          </div>
        </article>

        <button v-if="page < totalPages && !loading" class="load-more" @click="loadManuscripts(page + 1)">
          加载更多
        </button>
      </section>

      <section v-else class="state-block">
        <p>{{ activeTab === 'article' ? '图文稿件' : '音频稿件' }}接口还没有接入</p>
      </section>
    </main>

    <button class="publish-fab" @click="goPublish" aria-label="发布">
      <svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" /></svg>
    </button>

    <div v-if="actionItem" class="sheet-mask" @click.self="actionItem = null">
      <div class="action-sheet">
        <h3>{{ actionItem.title || '稿件操作' }}</h3>
        <button @click="togglePublishState(actionItem)">
          {{ Number(actionItem.status) === 3 ? '下架稿件' : '上架稿件' }}
        </button>
        <button @click="goVideo(actionItem)">查看稿件</button>
        <button class="danger" @click="removeItem(actionItem)">删除稿件</button>
        <button class="cancel" @click="actionItem = null">取消</button>
      </div>
    </div>

    <div v-if="toastText" class="page-toast">{{ toastText }}</div>
  </div>
</template>

<style scoped lang="scss">
.manuscript-manage-page {
  min-height: 100vh;
  background: #f6f7f9;
  color: #18191c;
  padding-bottom: calc(18px + env(safe-area-inset-bottom));
}

.manage-header {
  position: sticky;
  top: 0;
  z-index: 40;
  height: 78px;
  padding: 22px 18px 0;
  display: grid;
  grid-template-columns: 42px 1fr 74px;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;
  box-sizing: border-box;

  h1 {
    margin: 0;
    text-align: left;
    color: #18191c;
    font-size: 24px;
    font-weight: 600;
  }
}

.back-btn,
.draft-btn {
  border: 0;
  background: transparent;
  color: #61666d;
}

.back-btn {
  width: 38px;
  height: 38px;
  padding: 0;
  display: grid;
  place-items: center;

  svg {
    width: 34px;
    height: 34px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.draft-btn {
  font-size: 18px;
}

.manage-tabs {
  position: sticky;
  top: 78px;
  z-index: 35;
  height: 56px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  background: #fff;
  box-shadow: 0 2px 7px rgba(0, 0, 0, 0.08);

  button {
    position: relative;
    border: 0;
    background: transparent;
    color: #61666d;
    font-size: 20px;
    font-weight: 500;

    &.active {
      color: #fb7299;

      &::after {
        content: '';
        position: absolute;
        left: 18%;
        right: 18%;
        bottom: 0;
        height: 4px;
        border-radius: 4px;
        background: #fb7299;
      }
    }
  }
}

.manage-content {
  padding-bottom: 70px;
}

.manuscript-card {
  margin-top: 10px;
  background: #fff;
  border-top: 1px solid #f1f2f3;
  border-bottom: 1px solid #f1f2f3;
}

.summary-row {
  display: grid;
  grid-template-columns: 42% 1fr;
  gap: 14px;
  padding: 14px;
}

.cover-box {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  border-radius: 4px;
  background: #e3e5e7;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  b,
  .visibility-tag {
    position: absolute;
    bottom: 6px;
    color: #fff;
    font-size: 13px;
    font-weight: 500;
  }

  b {
    right: 7px;
    padding: 1px 4px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.48);
  }

  .visibility-tag {
    left: 7px;
    padding: 2px 6px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.5);
  }
}

.summary-main {
  min-width: 0;

  h2 {
    min-height: 54px;
    margin: 0 0 14px;
    color: #18191c;
    font-size: 20px;
    font-weight: 500;
    line-height: 1.36;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  time {
    display: block;
    color: #9499a0;
    font-size: 17px;
    line-height: 1.35;
  }
}

.status-pill {
  display: inline-flex;
  align-items: center;
  height: 22px;
  margin-top: 8px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 12px;
  color: #fff;
  background: #9499a0;

  &.published {
    background: #20b86a;
  }

  &.pending {
    background: #ff9e1b;
  }

  &.error {
    background: #f65c5c;
  }

  &.muted {
    background: #9499a0;
  }
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  padding: 14px 16px;
  border-top: 1px solid #f1f2f3;
  color: #9499a0;
  font-size: 17px;

  span {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-width: 0;
  }

  svg {
    width: 21px;
    height: 21px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.9;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.action-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
  padding: 14px 28px 16px;
  background: #fff;
  border-top: 1px solid #f1f2f3;

  button {
    height: 38px;
    border: 1px solid #d3d6dc;
    border-radius: 5px;
    background: #fff;
    color: #61666d;
    font-size: 17px;
  }
}

.state-block {
  margin: 12px 0;
  padding: 72px 18px;
  text-align: center;
  color: #9499a0;
  background: #fff;

  p {
    margin: 0 0 16px;
    font-size: 17px;
  }

  button {
    min-width: 116px;
    height: 38px;
    border: 0;
    border-radius: 19px;
    color: #fff;
    background: #fb7299;
    font-size: 16px;
  }
}

.load-more {
  width: calc(100% - 28px);
  height: 42px;
  margin: 14px;
  border: 0;
  border-radius: 4px;
  color: #61666d;
  background: #fff;
  font-size: 16px;
}

.publish-fab {
  position: fixed;
  right: 18px;
  bottom: calc(22px + env(safe-area-inset-bottom));
  z-index: 30;
  width: 54px;
  height: 54px;
  border: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  background: #fb7299;
  box-shadow: 0 8px 18px rgba(251, 114, 153, 0.35);

  svg {
    width: 30px;
    height: 30px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
    stroke-linecap: round;
  }
}

.sheet-mask {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: flex-end;
  background: rgba(0, 0, 0, 0.38);
}

.action-sheet {
  width: 100%;
  padding: 16px 16px calc(16px + env(safe-area-inset-bottom));
  border-radius: 14px 14px 0 0;
  background: #fff;
  box-sizing: border-box;

  h3 {
    margin: 0 0 10px;
    color: #18191c;
    font-size: 16px;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  button {
    width: 100%;
    height: 48px;
    border: 0;
    border-top: 1px solid #f1f2f3;
    background: #fff;
    color: #18191c;
    font-size: 17px;

    &.danger {
      color: #f65c5c;
    }

    &.cancel {
      margin-top: 8px;
      border-top: 8px solid #f6f7f9;
      color: #9499a0;
    }
  }
}

.page-toast {
  position: fixed;
  left: 50%;
  bottom: 92px;
  z-index: 100;
  max-width: 76vw;
  padding: 9px 14px;
  border-radius: 18px;
  transform: translateX(-50%);
  color: #fff;
  background: rgba(0, 0, 0, 0.78);
  font-size: 14px;
}

@media (max-width: 390px) {
  .summary-main h2 {
    font-size: 18px;
  }

  .summary-main time,
  .stats-row,
  .action-row button {
    font-size: 15px;
  }

  .action-row {
    gap: 10px;
    padding-inline: 18px;
  }
}
</style>
