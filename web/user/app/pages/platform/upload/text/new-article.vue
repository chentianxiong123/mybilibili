<template>
  <div class="article-home">
    <UploadTypeTabs active="article" @change="switchUploadType" />

    <!-- 提示条 -->
    <div class="tips-row">
      <div class="tip-card">
        <svg viewBox="0 0 24 24" width="22" height="22"><rect x="3" y="4" width="18" height="16" rx="2" fill="none" stroke="#00aeec" stroke-width="1.6"/><circle cx="8.5" cy="9.5" r="1.6" fill="#00aeec"/><path d="m4 17 5-5 4 4 3-3 4 4" fill="none" stroke="#00aeec" stroke-width="1.6" stroke-linejoin="round"/></svg>
        <div>
          <div class="tip-title">图片格式</div>
          <div class="tip-desc">图片仅支持.jpg/.png/.webp/.gif/.heif</div>
          <div class="tip-desc">静态图片不超过20M，gif图片不超过8M</div>
        </div>
      </div>
      <div class="tip-card">
        <svg viewBox="0 0 24 24" width="22" height="22"><path d="M12 3l2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.7-5.2 2.7 1-5.8L3.5 9.2l5.9-.9z" fill="none" stroke="#f8a05a" stroke-width="1.5" stroke-linejoin="round"/></svg>
        <div>
          <div class="tip-title">可发专栏数 <span class="tag">需LV1+</span></div>
          <div class="tip-desc">账号等级LV1-3，每天可发布 1 篇专栏</div>
          <div class="tip-desc">账号等级LV4-6，每天可发布 5 篇专栏</div>
        </div>
      </div>
    </div>

    <button class="new-create-btn" @click="createNew">新的创作</button>

    <!-- 草稿箱 -->
    <div class="draft-panel">
      <div class="draft-head">
        <span class="draft-title">草稿箱</span>
        <span class="draft-count">({{ drafts.length }})</span>
        <input v-model="keyword" class="draft-search" placeholder="标题搜索" @keyup.enter="load" />
      </div>
      <div v-if="loading" class="draft-loading">加载中…</div>
      <div v-else-if="!drafts.length" class="draft-empty">暂无草稿</div>
      <div v-else class="draft-list">
        <div v-for="d in filteredDrafts" :key="d.id" class="draft-item">
          <div class="draft-main">
            <div class="draft-item-title">{{ d.title || '无标题' }}</div>
            <div class="draft-item-summary">{{ d.summary || '（暂无内容）' }}</div>
            <div class="draft-item-time">最后修改时间：{{ fmt(d.updated_at) }}</div>
          </div>
          <div class="draft-actions">
            <button @click="edit(d.id)">编辑</button>
            <button @click="remove(d.id)">删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import UploadTypeTabs from '@/components/site/uploadTypeTabs/UploadTypeTabs.vue'
import { articleApi } from '@/api/article'

const router = useRouter()
const drafts = ref<any[]>([])
const loading = ref(false)
const keyword = ref('')

const filteredDrafts = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return drafts.value
  return drafts.value.filter(d => (d.title || '').toLowerCase().includes(k))
})

function fmt(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return t.replace('T', ' ').slice(0, 19)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

function switchUploadType(type?: string) {
  if (type === 'video') router.push('/platform/upload/video')
  else if (type === 'dynamic') router.push('/platform/upload/dynamic')
}

function createNew() {
  router.push('/platform/upload/text/new-edit')
}
function edit(id: number) {
  router.push({ path: '/platform/upload/text/new-edit', query: { id } })
}
async function remove(id: number) {
  try {
    await ElMessageBox.confirm('确定删除这篇草稿吗？', '删除草稿', { confirmButtonText: '删除', cancelButtonText: '取消' })
  } catch {
    return
  }
  try {
    const res: any = await articleApi.remove(id)
    if (res.code === 200) {
      ElMessage.success('已删除')
      drafts.value = drafts.value.filter(d => d.id !== id)
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {
    ElMessage.error('删除失败')
  }
}

async function load() {
  loading.value = true
  try {
    const res: any = await articleApi.getDrafts(keyword.value.trim())
    drafts.value = res?.data || []
  } catch {
    drafts.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.tips-row { display: flex; gap: 16px; margin-bottom: 16px; }
.tip-card {
  flex: 1; display: flex; gap: 10px; align-items: flex-start;
  background: #f6f7f8; border-radius: 8px; padding: 14px 16px;
}
.tip-title { font-size: 14px; font-weight: 600; color: #18191c; margin-bottom: 4px; }
.tip-desc { font-size: 12px; color: #9499a0; line-height: 1.7; }
.tag { display: inline-block; margin-left: 6px; font-size: 11px; color: #f8a05a; border: 1px solid #f8cfa5; border-radius: 3px; padding: 0 4px; }

.new-create-btn {
  display: block; width: 100%; height: 44px; border: none; border-radius: 6px;
  background: #00aeec; color: #fff; font-size: 15px; cursor: pointer; margin-bottom: 20px;
}
.new-create-btn:hover { background: #0096c7; }

.draft-panel { border: 1px solid #f1f2f3; border-radius: 8px; }
.draft-head {
  display: flex; align-items: center; gap: 8px; padding: 14px 20px; border-bottom: 1px solid #f1f2f3;
}
.draft-title { font-size: 16px; font-weight: 600; color: #18191c; }
.draft-count { font-size: 14px; color: #9499a0; }
.draft-search {
  margin-left: auto; width: 200px; height: 30px; border: 1px solid #e3e5e7; border-radius: 4px; padding: 0 10px; font-size: 13px; outline: none;
}
.draft-search:focus { border-color: #00aeec; }
.draft-loading, .draft-empty { padding: 60px 0; text-align: center; color: #9499a0; font-size: 14px; }
.draft-item {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
  padding: 16px 20px; border-bottom: 1px solid #f6f7f8;
}
.draft-item:last-child { border-bottom: none; }
.draft-item-title { font-size: 15px; color: #18191c; margin-bottom: 6px; }
.draft-item-summary { font-size: 13px; color: #9499a0; margin-bottom: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 640px; }
.draft-item-time { font-size: 12px; color: #c9ccd0; }
.draft-actions { display: flex; gap: 10px; flex: 0 0 auto; }
.draft-actions button {
  height: 30px; padding: 0 16px; border: 1px solid #e3e5e7; border-radius: 4px; background: #fff; color: #18191c; font-size: 13px; cursor: pointer;
}
.draft-actions button:hover { border-color: #00aeec; color: #00aeec; }
</style>
