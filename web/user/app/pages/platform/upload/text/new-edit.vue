<template>
  <div class="article-edit-page">
    <UploadTypeTabs active="article" @change="switchUploadType" />

    <div class="edit-body">
      <!-- 目录 -->
      <aside class="edit-toc">
        <div class="toc-head">
          <span>目录</span>
          <span class="toc-toggle" @click="tocOpen = !tocOpen">{{ tocOpen ? '收起' : '目录' }}</span>
        </div>
        <div v-if="tocOpen" class="toc-list">
          <p v-if="!toc.length" class="toc-empty">对文档内容应用“标题”样式即可生成目录</p>
          <a
            v-for="(h, i) in toc"
            :key="i"
            class="toc-item"
            :class="`lvl-${h.level}`"
            @click="scrollToHeading(i)"
          >{{ h.text }}</a>
        </div>
      </aside>

      <!-- 主体 -->
      <div class="edit-main">
        <ArticleEditor
          v-model:title="title"
          v-model:content="content"
          @preview="showPreview = true"
        />
        <div class="edit-splitter"></div>
        <ArticleSettings v-model="settings" />

        <div class="publish-footer">
          <div class="footer-left">
            <button class="btn-grey" @click="showSettingsModal = true">
              发布设置
              <svg viewBox="0 0 12 12" width="12" height="12"><path d="M2 4l4 4 4-4" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg>
            </button>
          </div>
          <div class="footer-right">
            <button class="btn-plain" :disabled="saving" @click="saveDraft">保存为草稿</button>
            <button class="btn-blue" :disabled="saving" @click="publish">发布</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 预览浮层 -->
    <div v-if="showPreview" class="preview-mask" @click.self="showPreview = false">
      <div class="preview-phone">
        <div class="preview-bar">
          <span>手机预览</span>
          <span class="preview-close" @click="showPreview = false">✕</span>
        </div>
        <div class="preview-body">
          <h1 class="preview-title">{{ title || '请输入标题' }}</h1>
          <div class="preview-author">
            <span>{{ authorName }}</span>
          </div>
          <div class="preview-content" v-html="previewHtml"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '@/stores/app'
import UploadTypeTabs from '@/components/site/uploadTypeTabs/UploadTypeTabs.vue'
import ArticleEditor from '@/components/createCenter/article/ArticleEditor.vue'
import ArticleSettings, { type ArticleSettings as SettingsT } from '@/components/createCenter/article/ArticleSettings.vue'
import { articleApi } from '@/api/article'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const title = ref('')
const content = ref('')
const currentId = ref<number>(0)
const saving = ref(false)
const showPreview = ref(false)
const showSettingsModal = ref(false)
const tocOpen = ref(true)
const settings = ref<SettingsT>({
  visibility: 1,
  coverEnabled: false,
  coverUrl: '',
  commentEnabled: true,
  featuredComment: false,
  isOriginal: false,
  aiDeclaration: false,
  topics: '',
  collectionName: ''
})

const authorName = computed(() => (appStore.user as any)?.nickname || '我')

const toc = computed(() => {
  const list: { level: number; text: string }[] = []
  const re = /<h([1-3])[^>]*>(.*?)<\/h\1>/gi
  let m: RegExpExecArray | null
  while ((m = re.exec(content.value)) !== null) {
    const text = m[2].replace(/<[^>]+>/g, '').trim()
    if (text) list.push({ level: Number(m[1]), text })
  }
  return list
})
const previewHtml = computed(() => content.value || '<p style="color:#c9ccd0">请输入正文</p>')

function scrollToHeading(idx: number) {
  const nodes = document.querySelectorAll('.article-content .eva3-editor h1, .article-content .eva3-editor h2, .article-content .eva3-editor h3')
  ;(nodes[idx] as HTMLElement)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function switchUploadType(type?: string) {
  if (type === 'video') router.push('/platform/upload/video')
  else if (type === 'dynamic') router.push('/platform/upload/dynamic')
}

function textLength(html: string) {
  const t = html.replace(/<[^>]+>/g, '').replace(/\s+/g, '')
  return t.length
}

function buildPayload(publish: boolean) {
  return {
    id: currentId.value || undefined,
    title: title.value.trim(),
    content: content.value,
    cover_url: settings.value.coverUrl,
    topics: settings.value.topics,
    collection_name: settings.value.collectionName,
    visibility: settings.value.visibility,
    comment_enabled: settings.value.commentEnabled ? 1 : 0,
    featured_comment: settings.value.featuredComment ? 1 : 0,
    is_original: settings.value.isOriginal ? 1 : 0,
    ai_declaration: settings.value.aiDeclaration ? 1 : 0,
    word_count: textLength(content.value),
    publish
  }
}

async function saveDraft() {
  if (saving.value) return
  saving.value = true
  try {
    const res: any = await articleApi.save(buildPayload(false))
    if (res.code === 200) {
      const id = res.data?.id
      if (id) currentId.value = id
      ElMessage.success('已保存为草稿')
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function publish() {
  if (!title.value.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  if (textLength(content.value) === 0) {
    ElMessage.warning('请输入正文')
    return
  }
  try {
    await ElMessageBox.confirm('专栏发布后需通过审核，审核通过后即可在专栏区展示。确定发布吗？', '发布专栏', {
      confirmButtonText: '发布',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  saving.value = true
  try {
    const res: any = await articleApi.save(buildPayload(true))
    if (res.code === 200) {
      const id = res.data?.id || currentId.value
      await articleApi.view(id).catch(() => {})
      ElMessage.success('发布成功')
      router.push(`/read/cv${id}`)
    } else {
      ElMessage.error(res.message || '发布失败')
    }
  } catch {
    ElMessage.error('发布失败')
  } finally {
    saving.value = false
  }
}

async function loadArticle(id: number) {
  try {
    const res: any = await articleApi.getArticle(id)
    if (res.code === 200 && res.data) {
      const a = res.data
      currentId.value = a.id
      title.value = a.title || ''
      content.value = a.content || ''
      settings.value = {
        visibility: a.visibility ?? 1,
        coverEnabled: !!a.cover_url,
        coverUrl: a.cover_url || '',
        commentEnabled: (a.comment_enabled ?? 1) === 1,
        featuredComment: (a.featured_comment ?? 0) === 1,
        isOriginal: (a.is_original ?? 0) === 1,
        aiDeclaration: (a.ai_declaration ?? 0) === 1,
        topics: a.topics || '',
        collectionName: a.collection_name || ''
      }
    } else {
      ElMessage.error(res.message || '加载失败')
    }
  } catch {
    ElMessage.error('加载失败')
  }
}

onMounted(() => {
  const id = Number(route.query.id || 0)
  if (id > 0) loadArticle(id)
})
</script>

<style scoped>
.article-edit-page {
  display: flex;
  flex-direction: column;
}

.edit-body {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.edit-toc {
  flex: 0 0 150px;
  position: sticky;
  top: 16px;
  font-size: 13px;
  color: #61666d;
}
.toc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 8px;
}
.toc-toggle { font-size: 12px; color: #00a1d6; cursor: pointer; font-weight: 400; }
.toc-empty { font-size: 12px; color: #c9ccd0; line-height: 1.6; }
.toc-item {
  display: block;
  padding: 4px 0;
  color: #61666d;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.toc-item:hover { color: #00aeec; }
.toc-item.lvl-2 { padding-left: 10px; }
.toc-item.lvl-3 { padding-left: 20px; }

.edit-main {
  flex: 1;
  min-width: 0;
  border: 1px solid #f1f2f3;
  border-radius: 4px;
  overflow: hidden;
  background: #fff;
}
.edit-splitter { height: 1px; background: #f1f2f3; }

.publish-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-top: 1px solid #f1f2f3;
  background: #fff;
}
.footer-right { display: flex; align-items: center; gap: 12px; }
.btn-grey {
  display: inline-flex; align-items: center; gap: 6px;
  height: 32px; padding: 0 16px; border: none; border-radius: 16px;
  background: #f1f2f3; color: #61666d; font-size: 14px; cursor: pointer;
}
.btn-grey:hover { background: #e3e5e7; }
.btn-plain {
  height: 32px; padding: 0 20px; border: 1px solid #00aeec; border-radius: 4px;
  background: #fff; color: #00aeec; font-size: 14px; cursor: pointer;
}
.btn-plain:hover { background: rgba(0,174,236,.06); }
.btn-blue {
  height: 32px; padding: 0 28px; border: none; border-radius: 4px;
  background: #00aeec; color: #fff; font-size: 14px; cursor: pointer;
}
.btn-blue:hover { background: #0096c7; }
.btn-plain:disabled, .btn-blue:disabled { opacity: .6; cursor: not-allowed; }

/* 预览 */
.preview-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,.45); z-index: 3000;
  display: flex; align-items: center; justify-content: center;
}
.preview-phone {
  width: 380px; height: 80vh; background: #fff; border-radius: 12px; display: flex; flex-direction: column; overflow: hidden;
}
.preview-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 16px; border-bottom: 1px solid #f1f2f3; font-size: 14px; color: #18191c;
}
.preview-close { cursor: pointer; color: #9499a0; }
.preview-body { flex: 1; overflow-y: auto; padding: 20px; }
.preview-title { font-size: 20px; font-weight: 700; margin: 0 0 12px; color: #18191c; }
.preview-author { font-size: 12px; color: #9499a0; margin-bottom: 16px; }
.preview-content { font-size: 15px; line-height: 1.8; color: #2f3238; word-break: break-word; }
.preview-content :deep(img) { max-width: 100%; }
</style>
