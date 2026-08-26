<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listDrafts, deleteDraft, clearAllDrafts, type ManuscriptDraft } from '../../utils/drafts'

const router = useRouter()
const drafts = ref<ManuscriptDraft[]>([])
const toastText = ref('')
const confirmingClear = ref(false)

const showToast = (text: string) => {
  toastText.value = text
  window.setTimeout(() => {
    if (toastText.value === text) toastText.value = ''
  }, 1800)
}

const reload = () => {
  drafts.value = listDrafts()
}

const isEmpty = computed(() => drafts.value.length === 0)

const goBack = () => router.back()

const formatTime = (ts: number) => {
  const d = new Date(ts)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const partCountText = (d: ManuscriptDraft) => {
  const n = d.videoParts?.length || 0
  return n ? `${n} 个分集` : '无分集'
}

/** 跳到 web 创作中心继续编辑（wap 无自己的发布页） */
const continueEdit = (d: ManuscriptDraft) => {
  const origin = window.location.origin.replace(':5174', ':5173')
  const url = `${origin}/create-center/upload?draftId=${d.id}`
  window.location.href = url
}

const removeDraft = (d: ManuscriptDraft) => {
  if (deleteDraft(d.id)) {
    reload()
    showToast('已删除草稿')
  } else {
    showToast('删除失败')
  }
}

const onClearAll = () => {
  if (!confirmingClear.value) {
    confirmingClear.value = true
    showToast('再次点击确认清空')
    window.setTimeout(() => (confirmingClear.value = false), 2500)
    return
  }
  clearAllDrafts()
  reload()
  confirmingClear.value = false
  showToast('已清空全部草稿')
}

onMounted(reload)
</script>

<template>
  <div class="drafts-page">
    <header class="drafts-header">
      <button class="back-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1>草稿箱</h1>
      <button
        v-if="!isEmpty"
        class="clear-btn"
        :class="{ confirming: confirmingClear }"
        @click="onClearAll"
      >
        {{ confirmingClear ? '确认清空' : '清空' }}
      </button>
    </header>

    <main class="drafts-content">
      <p v-if="isEmpty" class="empty-state">
        草稿箱空空如也<br />
        <span class="hint">在投稿时点击「存草稿」可暂存稿件</span>
      </p>

      <article
        v-for="d in drafts"
        :key="d.id"
        class="draft-card"
      >
        <div class="draft-cover" :class="{ placeholder: !d.coverPreview }">
          <img v-if="d.coverPreview" :src="d.coverPreview" alt="封面" />
          <span v-else class="cover-fallback">草稿</span>
        </div>
        <div class="draft-info">
          <div class="draft-title">{{ d.title || '未命名稿件' }}</div>
          <div class="draft-meta">
            <span>{{ partCountText(d) }}</span>
            <span v-if="d.hasLocalVideoFiles" class="warn-tag">含未上传视频</span>
          </div>
          <div class="draft-desc" v-if="d.description">{{ d.description }}</div>
          <div class="draft-time">{{ formatTime(d.updatedAt) }}</div>
        </div>
        <div class="draft-actions">
          <button class="act-btn edit" @click="continueEdit(d)">继续</button>
          <button class="act-btn del" @click="removeDraft(d)">删除</button>
        </div>
      </article>
    </main>

    <transition name="toast">
      <div v-if="toastText" class="toast">{{ toastText }}</div>
    </transition>
  </div>
</template>

<style scoped lang="scss">

.drafts-page {
  min-height: 100dvh;
  background: #f4f4f4;
  padding-bottom: 24px;
}

.drafts-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 8px 0 4px;
  background: #fff;
  border-bottom: 1px solid #e3e5e7;

  h1 {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 600;
    color: #18191c;
  }

  .back-btn {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    svg {
      width: 22px;
      height: 22px;
      fill: none;
      stroke: #18191c;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
  }

  .clear-btn {
    padding: 6px 12px;
    font-size: 13px;
    color: #9499a0;
    background: none;
    border: none;
    &.confirming {
      color: #ff4d4f;
      font-weight: 600;
    }
  }
}

.drafts-content {
  padding: 12px;
}

.empty-state {
  margin-top: 80px;
  text-align: center;
  color: #9499a0;
  font-size: 15px;
  line-height: 1.8;
  .hint {
    font-size: 12px;
    color: #c0c0c0;
  }
}

.draft-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 10px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.draft-cover {
  flex: 0 0 96px;
  width: 96px;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
  background: #e3e5e7;
  display: flex;
  align-items: center;
  justify-content: center;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .cover-fallback {
    font-size: 12px;
    color: #c0c0c0;
  }
}

.draft-info {
  flex: 1 1 auto;
  min-width: 0;
}

.draft-title {
  font-size: 14px;
  font-weight: 600;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.draft-meta {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 4px;
  font-size: 11px;
  color: #9499a0;
  .warn-tag {
    color: #fa8c16;
    background: rgba(250, 140, 22, 0.1);
    padding: 1px 6px;
    border-radius: 4px;
  }
}

.draft-desc {
  font-size: 12px;
  color: #9499a0;
  margin-top: 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.draft-time {
  font-size: 11px;
  color: #c0c0c0;
  margin-top: 4px;
}

.draft-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.act-btn {
  padding: 5px 12px;
  font-size: 12px;
  border-radius: 14px;
  border: none;
  &.edit {
    background: #fb7299;
    color: #fff;
  }
  &.del {
    background: #e3e5e7;
    color: #9499a0;
  }
}

.toast {
  position: fixed;
  left: 50%;
  bottom: 72px;
  transform: translateX(-50%);
  padding: 8px 16px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  font-size: 13px;
  border-radius: 20px;
  z-index: 100;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.2s;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
}
</style>
