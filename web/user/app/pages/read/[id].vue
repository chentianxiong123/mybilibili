<template>
  <div class="opus-page">
    <div v-if="loading" class="cv-loading">加载中…</div>
    <div v-else-if="!article" class="cv-empty">
      <p>专栏不存在或已被删除</p>
      <button @click="$router.push('/read/home')">返回专栏首页</button>
    </div>

    <template v-else>
      <!-- 右侧悬浮互动栏 -->
      <div class="side-toolbar">
        <div class="side-toolbar__action like" :class="{ active: liked }" @click="onLike">
          <svg viewBox="0 0 24 24" width="26" height="26"><path d="M7 22V10l5-8 1.5.5c.8.3 1.2 1.1 1 1.9L13 10h5.5c1.2 0 2.1 1 2 2.2l-1 7c-.1 1-1 1.8-2 1.8H7zM7 10 3 10v12h4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>
          <span class="side-toolbar__action__text">{{ fmtNum(article.like_count) }}</span>
        </div>
        <div class="side-toolbar__action coin" :class="{ active: myCoin > 0 }" @click="onCoin">
          <svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="8.5" fill="none" stroke="currentColor" stroke-width="1.6"/><path d="M12 8v8M8 12h8" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
          <span class="side-toolbar__action__text">{{ fmtNum(article.coin_count) }}</span>
        </div>
        <div class="side-toolbar__action favorite" :class="{ active: favorited }" @click="onFavorite">
          <svg viewBox="0 0 24 24" width="26" height="26"><path d="M6 3h12a1 1 0 0 1 1 1v17l-7-4-7 4V4a1 1 0 0 1 1-1z" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>
          <span class="side-toolbar__action__text">{{ fmtNum(article.favorite_count) }}</span>
        </div>
        <div class="side-toolbar__action forward" @click="onShare">
          <svg viewBox="0 0 24 24" width="26" height="26"><path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7M12 16V3m0 0L8 7m4-4 4 4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/></svg>
          <span class="side-toolbar__action__text">{{ fmtNum(article.share_count) }}</span>
        </div>
        <div class="side-toolbar__action comment" @click="scrollToComments">
          <svg viewBox="0 0 24 24" width="26" height="26"><path d="M4 5h16a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H9l-5 4V6a1 1 0 0 1 1-1z" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>
          <span class="side-toolbar__action__text">{{ fmtNum(article.comment_count) }}</span>
        </div>
      </div>

      <div class="opus-container">
        <!-- 标题 -->
        <h1 class="opus-module-title">{{ article.title || '无标题' }}</h1>

        <!-- 作者 -->
        <div class="opus-module-author">
          <img v-if="article.author_avatar" :src="article.author_avatar" alt="" class="opus-avatar" />
          <div v-else class="opus-avatar opus-avatar--empty"></div>
          <div class="opus-author-center">
            <span class="opus-author-name">{{ article.author_name }}</span>
            <span class="opus-author-pub">
              <span class="opus-author-pub__text">{{ fmt(article.publish_time || article.created_at) }}</span>
              <span v-if="article.is_original" class="opus-badge">原创</span>
              <span v-if="article.ai_declaration" class="opus-badge">AI辅助</span>
            </span>
          </div>
        </div>

        <!-- 正文 -->
        <div class="opus-module-content" v-html="article.content"></div>

        <!-- 底部分享 -->
        <div class="opus-module-bottom">
          <div class="opus-module-bottom__share">
            <span class="share-label">分享到：</span>
            <span class="share-item dynamic" @click="onShare">动态</span>
            <span class="share-item weibo" @click="onShare">微博</span>
            <span class="share-item qzone" @click="onShare">QQ空间</span>
            <span class="share-item qq" @click="onShare">QQ</span>
          </div>
          <div class="opus-module-bottom__feedback">举报</div>
        </div>

        <!-- 评论 -->
        <div ref="commentRef" class="comment-wrap">
          <div class="comment-head">
            <span class="comment-title">评论</span>
            <span class="comment-total">{{ commentsTotal }}</span>
          </div>

          <div class="comment-editor">
            <img v-if="myAvatar" :src="myAvatar" alt="" class="comment-avatar" />
            <div v-else class="comment-avatar comment-avatar--empty"></div>
            <div class="comment-editor__main">
              <div v-if="replyTo.id" class="reply-hint">
                回复 <b>@{{ replyTo.name }}</b>
                <span class="reply-cancel" @click="cancelReply">取消</span>
              </div>
              <div class="comment-box">
                <textarea
                  ref="taRef"
                  v-model="commentText"
                  class="comment-textarea"
                  :placeholder="isLogin ? (replyTo.id ? `回复 @${replyTo.name}` : '发一条友善的评论吧') : '请先登录后发表评论'"
                  maxlength="1000"
                  :disabled="!isLogin"
                  @keydown="onTextareaKeydown"
                ></textarea>
                <div v-if="mentionOpen" class="mention-pop">
                  <div class="mention-pop__tip">选择或输入你想@的人</div>
                  <input
                    v-model="mentionKeyword"
                    class="mention-pop__input"
                    placeholder="输入昵称搜索"
                  />
                  <div class="mention-pop__list">
                    <div v-if="mentionLoading" class="mention-pop__empty">搜索中…</div>
                    <div v-else-if="!mentionList.length" class="mention-pop__empty">没有找到相关用户</div>
                    <div
                      v-for="u in mentionList"
                      :key="u.uid"
                      class="mention-pop__item"
                      @mousedown.prevent="pickMention(u)"
                    >
                      <img v-if="u.avatar" :src="u.avatar" class="mention-pop__avatar" />
                      <span v-else class="mention-pop__avatar mention-pop__avatar--empty"></span>
                      <span class="mention-pop__name">{{ u.nickname }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="comment-editor__foot">
                <span class="comment-count">{{ commentText.length }}/1000</span>
                <span class="comment-at" :class="{ active: mentionOpen }" @click="toggleMention">@</span>
                <button class="comment-submit" :disabled="!isLogin || submitting || !commentText.trim()" @click="submitComment">发布</button>
              </div>
            </div>
          </div>

          <div v-if="loadingComments" class="comment-empty">加载中…</div>
          <div v-else-if="!comments.length" class="comment-empty">还没有评论，快来抢沙发吧~</div>
          <div v-else class="comment-list">
            <div v-for="c in comments" :key="c.id" class="comment-item">
              <img v-if="c.author_avatar" :src="c.author_avatar" alt="" class="comment-avatar" />
              <div v-else class="comment-avatar comment-avatar--empty"></div>
              <div class="comment-item__body">
                <div class="comment-item__name">{{ c.author_name }}</div>
                <div class="comment-item__text">{{ c.content }}</div>
                <div class="comment-item__foot">
                  <span class="comment-item__time">{{ fmt(c.created_at) }}</span>
                  <span class="comment-item__btn like" :class="{ active: c.my_liked }" @click="likeComment(c)">
                    <svg viewBox="0 0 24 24" width="15" height="15"><path d="M7 22V10l5-8 1.5.5c.8.3 1.2 1.1 1 1.9L13 10h5.5c1.2 0 2.1 1 2 2.2l-1 7c-.1 1-1 1.8-2 1.8H7zM7 10 3 10v12h4" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
                    <span v-if="c.like_count">{{ c.like_count }}</span>
                  </span>
                  <span class="comment-item__btn" @click="startReply(c)">回复</span>
                  <span v-if="c.user_id === myUid" class="comment-item__btn danger" @click="removeComment(c)">删除</span>
                </div>

                <div v-if="c.replies && c.replies.length" class="comment-replies">
                  <div v-for="r in c.replies" :key="r.id" class="comment-reply">
                    <img v-if="r.author_avatar" :src="r.author_avatar" alt="" class="comment-avatar comment-avatar--sm" />
                    <div v-else class="comment-avatar comment-avatar--sm comment-avatar--empty"></div>
                    <div class="comment-item__body">
                      <div class="comment-item__name">
                        {{ r.author_name }}
                        <template v-if="r.reply_to_name"> <span class="reply-arrow">回复</span> {{ r.reply_to_name }}</template>
                      </div>
                      <div class="comment-item__text">{{ r.content }}</div>
                      <div class="comment-item__foot">
                        <span class="comment-item__time">{{ fmt(r.created_at) }}</span>
                        <span class="comment-item__btn like" :class="{ active: r.my_liked }" @click="likeComment(r)">
                          <svg viewBox="0 0 24 24" width="15" height="15"><path d="M7 22V10l5-8 1.5.5c.8.3 1.2 1.1 1 1.9L13 10h5.5c1.2 0 2.1 1 2 2.2l-1 7c-.1 1-1 1.8-2 1.8H7zM7 10 3 10v12h4" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
                          <span v-if="r.like_count">{{ r.like_count }}</span>
                        </span>
                        <span class="comment-item__btn" @click="startReply(r, c.id)">回复</span>
                        <span v-if="r.user_id === myUid" class="comment-item__btn danger" @click="removeComment(r)">删除</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <ArticleFavorite v-model="showFav" :article-id="parseId()" @changed="onFavChanged" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '@/stores/app'
import { articleApi } from '@/api/article'
import ArticleFavorite from '@/components/site/article/ArticleFavorite.vue'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const article = ref<any>(null)
const loading = ref(false)
const comments = ref<any[]>([])
const commentsTotal = ref(0)
const loadingComments = ref(false)
const commentText = ref('')
const submitting = ref(false)
const liked = ref(false)
const favorited = ref(false)
const myCoin = ref(0)
const replyTo = ref<{ id: number; name: string; userId: number }>({ id: 0, name: '', userId: 0 })
const commentRef = ref<HTMLElement | null>(null)
const taRef = ref<HTMLTextAreaElement | null>(null)
const showFav = ref(false)

const mentionOpen = ref(false)
const mentionKeyword = ref('')
const mentionList = ref<any[]>([])
const mentionLoading = ref(false)
const mentionCaret = ref(0)
let mentionTimer: any = null

const myUid = computed(() => Number((appStore.user as any)?.uid || 0))
const myAvatar = computed(() => (appStore.user as any)?.avatar || '')
const isLogin = computed(() => !!appStore.isLogin)

function parseId(): number {
  const m = String(route.params.id || '').match(/(\d+)/)
  return m ? Number(m[1]) : 0
}
function fmtNum(n: number) {
  n = n || 0
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
function fmt(t: string) {
  if (!t) return ''
  const d = new Date(t)
  if (isNaN(d.getTime())) return String(t).replace('T', ' ').slice(0, 16)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

function requireLogin() {
  if (!isLogin.value) {
    ElMessage.warning('请先登录')
    return false
  }
  return true
}

async function load() {
  const id = parseId()
  if (!id) {
    article.value = null
    return
  }
  loading.value = true
  article.value = null
  comments.value = []
  commentsTotal.value = 0
  try {
    const res: any = await articleApi.getArticle(id)
    if (res.code === 200 && res.data) {
      article.value = res.data
      liked.value = !!res.data.my_liked
      favorited.value = !!res.data.my_favorited
      myCoin.value = Number(res.data.my_coin || 0)
      articleApi.view(id).catch(() => {})
      loadComments()
    }
  } catch {
    article.value = null
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  const id = parseId()
  if (!id) return
  loadingComments.value = true
  try {
    const res: any = await articleApi.getComments(id, 1, 20)
    if (res.code === 200) {
      comments.value = res.data.list || []
      commentsTotal.value = res.data.total || 0
    }
  } catch {
    comments.value = []
  } finally {
    loadingComments.value = false
  }
}

async function onLike() {
  if (!requireLogin()) return
  try {
    const res: any = await articleApi.like(parseId())
    if (res.code === 200) {
      liked.value = res.data.active
      article.value.like_count = res.data.count
    }
  } catch {
    ElMessage.error('操作失败')
  }
}
async function onFavorite() {
  if (!requireLogin()) return
  showFav.value = true
}
function onFavChanged(p: { favorited: boolean; count: number }) {
  favorited.value = p.favorited
  if (article.value) article.value.favorite_count = p.count
}
async function onCoin() {
  if (!requireLogin()) return
  if (myCoin.value >= 2) {
    ElMessage.info('最多投 2 个币')
    return
  }
  try {
    const res: any = await articleApi.coin(parseId(), 1)
    if (res.code === 200) {
      myCoin.value = res.data.my_coin
      article.value.coin_count = res.data.count
      ElMessage.success('投币成功')
    } else {
      ElMessage.info(res.message || '投币失败')
    }
  } catch {
    ElMessage.error('操作失败')
  }
}
async function onShare() {
  try {
    await navigator.clipboard.writeText(window.location.href)
    const res: any = await articleApi.share(parseId())
    if (res?.code === 200) article.value.share_count = res.data.count
    ElMessage.success('链接已复制，快去分享吧')
  } catch {
    ElMessage.info(window.location.href)
  }
}

function scrollToComments() {
  commentRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function startReply(c: any, parentId?: number) {
  if (!requireLogin()) return
  replyTo.value = { id: parentId || c.id, name: c.author_name, userId: c.user_id }
  scrollToComments()
}
function cancelReply() {
  replyTo.value = { id: 0, name: '', userId: 0 }
}

async function submitComment() {
  if (!requireLogin()) return
  const content = commentText.value.trim()
  if (!content) return
  submitting.value = true
  try {
    const payload: any = { content }
    if (replyTo.value.id) {
      payload.parent_id = replyTo.value.id
      if (replyTo.value.userId) payload.reply_to_user_id = replyTo.value.userId
    }
    const res: any = await articleApi.postComment(parseId(), payload)
    if (res.code === 200) {
      const c = res.data
      if (c.parent_id) {
        const root = comments.value.find(item => item.id === c.parent_id)
        if (root) {
          root.replies = root.replies || []
          root.replies.push(c)
        }
      } else {
        comments.value.unshift(c)
      }
      commentsTotal.value += 1
      if (article.value) article.value.comment_count = (article.value.comment_count || 0) + 1
      commentText.value = ''
      cancelReply()
      ElMessage.success('评论成功')
    } else {
      ElMessage.error(res.message || '评论失败')
    }
  } catch {
    ElMessage.error('评论失败')
  } finally {
    submitting.value = false
  }
}

async function removeComment(c: any) {
  try {
    await ElMessageBox.confirm('确定删除这条评论吗？', '删除评论', { confirmButtonText: '删除', cancelButtonText: '取消' })
  } catch {
    return
  }
  try {
    const res: any = await articleApi.deleteComment(c.id)
    if (res.code === 200) {
      comments.value = comments.value.filter(item => item.id !== c.id)
      comments.value.forEach(item => {
        if (item.replies) item.replies = item.replies.filter((r: any) => r.id !== c.id)
      })
      commentsTotal.value = Math.max(0, commentsTotal.value - 1)
      if (article.value && typeof res.data?.count === 'number') article.value.comment_count = res.data.count
      ElMessage.success('已删除')
    }
  } catch {
    ElMessage.error('删除失败')
  }
}

async function likeComment(c: any) {
  if (!requireLogin()) return
  try {
    const res: any = await articleApi.likeComment(c.id)
    if (res.code === 200) {
      c.my_liked = res.data.active
      c.like_count = res.data.count
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

function insertMention(u: any) {
  const text = commentText.value
  const caret = Math.min(mentionCaret.value || text.length, text.length)
  const before = text.slice(0, caret)
  const after = text.slice(caret)
  const prefix = before && !before.endsWith(' ') ? ' ' : ''
  const insert = `${prefix}@${u.nickname} `
  commentText.value = before + insert + after
  mentionOpen.value = false
  mentionKeyword.value = ''
  nextTick(() => {
    const pos = (before + insert).length
    taRef.value?.focus()
    taRef.value?.setSelectionRange(pos, pos)
  })
}

function openMention() {
  if (!requireLogin()) return
  mentionCaret.value = taRef.value?.selectionStart ?? commentText.value.length
  mentionOpen.value = true
  if (!mentionKeyword.value) mentionList.value = []
}

function toggleMention() {
  if (mentionOpen.value) {
    mentionOpen.value = false
  } else {
    openMention()
  }
}

function pickMention(u: any) {
  insertMention(u)
}

function onTextareaKeydown(e: KeyboardEvent) {
  if (e.key === '@') {
    e.preventDefault()
    mentionCaret.value = taRef.value?.selectionStart ?? commentText.value.length
    mentionOpen.value = true
  } else if (e.key === 'Escape') {
    mentionOpen.value = false
  }
}

watch(mentionKeyword, (kw) => {
  if (mentionTimer) clearTimeout(mentionTimer)
  const key = kw.trim()
  if (!mentionOpen.value || !key) {
    mentionList.value = []
    return
  }
  mentionLoading.value = true
  mentionTimer = setTimeout(async () => {
    try {
      const res: any = await articleApi.mentionSearch(key)
      if (res.code === 200) mentionList.value = res.data || []
    } catch {
      mentionList.value = []
    } finally {
      mentionLoading.value = false
    }
  }, 250)
})

onMounted(load)
watch(() => route.params.id, load)
</script>

<style scoped>
.opus-page {
  background: #fff;
  min-height: 100vh;
  padding: 24px 0 80px;
}
.cv-loading,
.cv-empty { text-align: center; color: #9499a0; padding: 100px 0; }
.cv-empty button {
  margin-top: 12px; height: 32px; padding: 0 20px; border: 1px solid #00aeec; border-radius: 4px;
  background: #fff; color: #00aeec; font-size: 14px; cursor: pointer;
}

.opus-container {
  max-width: 720px;
  margin: 0 auto;
  padding: 0 24px;
}

.opus-module-title {
  font-size: 24px;
  font-weight: 700;
  color: #18191c;
  line-height: 1.4;
  margin: 0 0 24px;
}

.opus-module-author {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f1f2f3;
  margin-bottom: 28px;
}
.opus-avatar { width: 48px; height: 48px; border-radius: 50%; object-fit: cover; }
.opus-avatar--empty { background: #e3e5e7; }
.opus-author-center { display: flex; flex-direction: column; gap: 4px; }
.opus-author-name { font-size: 15px; color: #18191c; font-weight: 500; }
.opus-author-pub { display: flex; align-items: center; gap: 8px; font-size: 12px; color: #9499a0; }
.opus-badge {
  font-size: 11px; color: #00aeec; border: 1px solid #8fd8f4; border-radius: 3px; padding: 0 4px; line-height: 15px;
}

.opus-module-content { font-size: 16px; line-height: 1.9; color: #2f3238; word-break: break-word; }
.opus-module-content :deep(img) { max-width: 100%; border-radius: 6px; display: block; margin: 16px auto; }
.opus-module-content :deep(h1),
.opus-module-content :deep(h2),
.opus-module-content :deep(h3) { color: #18191c; margin: 28px 0 14px; line-height: 1.4; }
.opus-module-content :deep(h1) { font-size: 22px; }
.opus-module-content :deep(h2) { font-size: 19px; }
.opus-module-content :deep(h3) { font-size: 17px; }
.opus-module-content :deep(p) { margin: 0 0 16px; }
.opus-module-content :deep(blockquote) {
  border-left: 3px solid #00aeec; background: #f6f7f8; margin: 16px 0; padding: 10px 16px; color: #61666d;
}
.opus-module-content :deep(a) { color: #00aeec; }
.opus-module-content :deep(ul),
.opus-module-content :deep(ol) { padding-left: 24px; margin: 0 0 16px; }
.opus-module-content :deep(hr) { border: none; border-top: 1px solid #f1f2f3; margin: 24px 0; }

.opus-module-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 32px 0 8px;
  padding: 16px 0;
  border-top: 1px solid #f1f2f3;
  border-bottom: 1px solid #f1f2f3;
}
.opus-module-bottom__share { display: flex; align-items: center; gap: 14px; font-size: 13px; color: #61666d; }
.share-label { color: #9499a0; }
.share-item { cursor: pointer; }
.share-item:hover { color: #00aeec; }
.opus-module-bottom__feedback { font-size: 13px; color: #9499a0; cursor: pointer; }

/* 侧边悬浮栏 */
.side-toolbar {
  position: fixed;
  right: calc(50% - 360px - 80px);
  top: 200px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  z-index: 10;
}
.side-toolbar__action {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: #61666d;
  cursor: pointer;
  transition: color .2s;
}
.side-toolbar__action:hover { color: #00aeec; }
.side-toolbar__action.active { color: #00aeec; }
.side-toolbar__action__text { font-size: 12px; }

/* 评论 */
.comment-wrap { margin-top: 32px; }
.comment-head { display: flex; align-items: baseline; gap: 8px; margin-bottom: 16px; }
.comment-title { font-size: 18px; font-weight: 600; color: #18191c; }
.comment-total { font-size: 13px; color: #9499a0; }

.comment-editor { display: flex; gap: 12px; margin-bottom: 28px; }
.comment-avatar { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; flex: 0 0 auto; }
.comment-avatar--sm { width: 28px; height: 28px; }
.comment-avatar--empty { background: #e3e5e7; }
.comment-editor__main { flex: 1; min-width: 0; }
.reply-hint { font-size: 13px; color: #61666d; margin-bottom: 6px; }
.reply-cancel { color: #00aeec; cursor: pointer; margin-left: 8px; }
.comment-textarea {
  width: 100%; min-height: 72px; border: 1px solid #e3e5e7; border-radius: 6px; padding: 10px 12px;
  font-size: 14px; resize: vertical; outline: none; box-sizing: border-box; font-family: inherit;
}
.comment-textarea:focus { border-color: #00aeec; }
.comment-editor__foot { display: flex; align-items: center; justify-content: flex-end; gap: 12px; margin-top: 8px; }
.comment-count { font-size: 12px; color: #c9ccd0; }
.comment-submit {
  height: 32px; padding: 0 22px; border: none; border-radius: 4px; background: #00aeec; color: #fff; font-size: 14px; cursor: pointer;
}
.comment-submit:disabled { background: #c9ccd0; cursor: not-allowed; }

.comment-empty { text-align: center; color: #9499a0; font-size: 14px; padding: 40px 0; }
.comment-list { display: flex; flex-direction: column; }
.comment-item { display: flex; gap: 12px; padding: 16px 0; border-bottom: 1px solid #f6f7f8; }
.comment-item__body { flex: 1; min-width: 0; }
.comment-item__name { font-size: 13px; color: #00aeec; font-weight: 500; margin-bottom: 6px; }
.reply-arrow { color: #9499a0; font-weight: 400; }
.comment-item__text { font-size: 14px; color: #18191c; line-height: 1.7; word-break: break-word; white-space: pre-wrap; }
.comment-item__foot { display: flex; align-items: center; gap: 16px; margin-top: 8px; font-size: 12px; color: #9499a0; }
.comment-item__btn { cursor: pointer; }
.comment-item__btn:hover { color: #00aeec; }
.comment-item__btn.danger:hover { color: #f56c6c; }
.comment-item__btn.like { display: inline-flex; align-items: center; gap: 4px; }
.comment-item__btn.like.active { color: #00aeec; }

.comment-box { position: relative; }
.comment-at {
  font-size: 16px; font-weight: 600; color: #61666d; cursor: pointer; padding: 0 6px; user-select: none;
}
.comment-at:hover,
.comment-at.active { color: #00aeec; }
.mention-pop {
  position: absolute; left: 0; bottom: calc(100% + 8px); width: 320px; background: #fff;
  border: 1px solid #f1f2f3; border-radius: 8px; box-shadow: 0 4px 16px rgba(0,0,0,.12);
  padding: 10px 12px; z-index: 20;
}
.mention-pop__tip { font-size: 12px; color: #9499a0; margin-bottom: 8px; }
.mention-pop__input {
  width: 100%; height: 30px; border: 1px solid #e3e5e7; border-radius: 4px; padding: 0 10px;
  font-size: 13px; outline: none; box-sizing: border-box; margin-bottom: 8px;
}
.mention-pop__input:focus { border-color: #00aeec; }
.mention-pop__list { max-height: 200px; overflow-y: auto; }
.mention-pop__item {
  display: flex; align-items: center; gap: 8px; padding: 6px 4px; border-radius: 4px; cursor: pointer;
}
.mention-pop__item:hover { background: #f6f7f8; }
.mention-pop__avatar { width: 24px; height: 24px; border-radius: 50%; object-fit: cover; flex: 0 0 auto; }
.mention-pop__avatar--empty { background: #e3e5e7; }
.mention-pop__name { font-size: 13px; color: #18191c; }
.mention-pop__empty { text-align: center; color: #9499a0; font-size: 12px; padding: 14px 0; }
.comment-replies { margin-top: 12px; background: #f6f7f8; border-radius: 6px; padding: 8px 12px; }
.comment-reply { display: flex; gap: 10px; padding: 8px 0; }
</style>
