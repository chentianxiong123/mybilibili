<template>
  <div class="read-home">
    <div class="read-hero">
      <h1 class="read-hero-title">专栏</h1>
      <p class="read-hero-sub">发现优质的图文创作</p>
    </div>

    <div class="read-container">
      <div v-if="loading" class="read-loading">加载中…</div>
      <div v-else-if="!list.length" class="read-empty">暂无专栏</div>
      <div v-else class="read-list">
        <article v-for="a in list" :key="a.id" class="read-card" @click="open(a.id)">
          <div class="read-card-cover">
            <img v-if="a.cover_url" :src="a.cover_url" :alt="a.title" />
            <div v-else class="read-card-cover-text">{{ a.title || '无标题' }}</div>
          </div>
          <div class="read-card-body">
            <h3 class="read-card-title">{{ a.title || '无标题' }}</h3>
            <p class="read-card-summary">{{ a.summary || '（暂无摘要）' }}</p>
            <div class="read-card-meta">
              <span class="read-card-author">
                <img v-if="a.author_avatar" :src="a.author_avatar" alt="" />
                {{ a.author_name }}
              </span>
              <span class="read-card-stats">
                <span>{{ fmtNum(a.view_count) }} 阅读</span>
                <span>{{ fmtNum(a.like_count) }} 点赞</span>
              </span>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { articleApi } from '@/api/article'

const router = useRouter()
const list = ref<any[]>([])
const loading = ref(false)

function fmtNum(n: number) {
  n = n || 0
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
function open(id: number) {
  router.push(`/read/cv${id}`)
}

onMounted(async () => {
  loading.value = true
  try {
    const res: any = await articleApi.getFeed(1, 20)
    list.value = res?.data?.list || []
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.read-home { background: #fff; min-height: 100vh; }
.read-hero {
  padding: 40px 24px 24px;
  text-align: center;
  background: linear-gradient(180deg, #f6f7f8, #fff);
}
.read-hero-title { font-size: 30px; font-weight: 700; color: #18191c; margin: 0; }
.read-hero-sub { font-size: 14px; color: #9499a0; margin: 8px 0 0; }
.read-container { max-width: 1000px; margin: 0 auto; padding: 16px 24px 60px; }
.read-loading, .read-empty { text-align: center; color: #9499a0; padding: 60px 0; }
.read-card {
  display: flex; gap: 20px; padding: 20px 0; border-bottom: 1px solid #f1f2f3; cursor: pointer;
}
.read-card:hover .read-card-title { color: #00aeec; }
.read-card-cover {
  flex: 0 0 200px; height: 120px; border-radius: 6px; overflow: hidden; background: #f6f7f8;
  display: flex; align-items: center; justify-content: center;
}
.read-card-cover img { width: 100%; height: 100%; object-fit: cover; }
.read-card-cover-text { color: #9499a0; font-size: 13px; padding: 12px; text-align: center; overflow: hidden; }
.read-card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.read-card-title { font-size: 18px; font-weight: 600; color: #18191c; margin: 0 0 8px; transition: color .2s; }
.read-card-summary {
  font-size: 13px; color: #61666d; line-height: 1.7; margin: 0 0 auto;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.read-card-meta { display: flex; align-items: center; justify-content: space-between; margin-top: 12px; font-size: 12px; color: #9499a0; }
.read-card-author { display: flex; align-items: center; gap: 6px; }
.read-card-author img { width: 20px; height: 20px; border-radius: 50%; }
.read-card-stats { display: flex; gap: 14px; }
</style>
