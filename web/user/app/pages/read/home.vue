<template>
  <div class="read-home">
    <div class="read-main">
      <div class="read-content">
        <div class="read-header">
          <div class="read-title">推荐文章</div>
          <div class="read-search-box">
            <div class="read-search-bar">
              <input
                v-model="keyword"
                class="read-search-input__field"
                type="text"
                placeholder="搜索文章"
                @keyup.enter="handleSearch"
                />
              <div class="read-search-bar-icon" @click="handleSearch">
                  <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <circle cx="7" cy="7" r="5" stroke="#9499a0" stroke-width="1.4" />
                  <path d="M11 11L14 14" stroke="#9499a0" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                </div>
            </div>
          </div>
        </div>

        <div v-if="loading" class="read-loading">加载中…</div>
        <div v-else-if="!list.length" class="read-empty">暂无专栏</div>
        <div v-else class="read-feed">
          <div v-for="a in list" :key="a.id" class="read-feed-item">
            <a class="read-article-item" @click.prevent="open(a.id)">
              <div class="read-article-item__left">
                <div class="read-article-item__details">
                  <div class="read-article-item__title">{{ a.title || '无标题' }}</div>
                  <div class="read-article-item__desc">{{ a.summary || '（暂无摘要）' }}</div>
                </div>
                <div class="read-article-item__meta">
                  <div class="read-article-item__author">
                    <div class="read-article-item__avatar">
                      <img v-if="a.author_avatar" :src="a.author_avatar" alt="" />
                    </div>
                    <div class="read-article-item__uname">{{ a.author_name || '佚名' }}</div>
                  </div>
                  <div class="read-article-item__label">{{ a.category_name || a.topics || '日常' }}</div>
                  <div class="read-article-item__stats">
                    <div class="read-article-item__stat read-article-item__stat--view">
                      <span>{{ fmtNum(a.view_count) }}</span>
                    </div>
                    <div class="read-article-item__stat read-article-item__stat--like">
                      <span>{{ fmtNum(a.like_count) }}</span>
                    </div>
                    <div class="read-article-item__stat read-article-item__stat--comment">
                      <span>{{ fmtNum(a.comment_count) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="read-article-item__right">
                <div class="read-article-item__cover">
                  <img v-if="a.cover_url" :src="a.cover_url" :alt="a.title || ''" />
                </div>
              </div>
            </a>
          </div>
        </div>
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
const keyword = ref('')

function fmtNum(n: number) {
  n = n || 0
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
function open(id: number) {
  router.push(`/read/cv${id}`)
}
function handleSearch() {
  // 预留：与 B 站一致的关键词过滤入口
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
/* 对齐 bilibili.com/read/home：680px 居中单列 + 推荐文章头 + 图文信息流 */
.read-home {
  min-height: 100vh;
  background: #fff;
  color: #18191c;
}

.read-main {
  display: flex;
  box-sizing: border-box;
  width: 680px;
  padding-top: 30px;
  margin: 0 auto;
}

.read-content {
  width: 680px;
}

/* 头部：推荐文章 + 搜索框 */
.read-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
}

.read-title {
  font-size: 18px;
  font-weight: 400;
  line-height: 25px;
  color: #18191c;
}

.read-search-box {
  width: 260px;
  height: 40px;
}

.read-search-bar {
  position: relative;
  width: 260px;
  height: 40px;
}

.read-search-input__field {
  display: block;
  box-sizing: border-box;
  width: 100%;
  height: 40px;
  padding: 0;
  border: 0;
  outline: none;
  background: transparent;
  color: #18191c;
  font-size: 16px;
}

.read-search-input__field::placeholder {
  color: #9499a0;
}

.read-search-bar-icon {
  position: absolute;
  top: 12px;
  left: 232px;
  display: block;
  width: 16px;
  height: 16px;
  line-height: 0;
  color: #9499a0;
  font-size: 16px;
  cursor: pointer;
}

/* 信息流 */
.read-feed {
  margin-top: 0;
}

.read-feed-item {
  height: 128px;
}

.read-article-item {
  display: flex;
  box-sizing: border-box;
  height: 128px;
  padding: 20px 0;
  color: #18191c;
  cursor: pointer;
}

.read-article-item:hover .read-article-item__title {
  color: #00aeec;
}

.read-article-item__left {
  display: flex;
  flex: 1 1 0%;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.read-article-item__right {
  flex: 0 0 auto;
  margin-left: 20px;
}

.read-article-item__details {
  height: 52px;
}

.read-article-item__title {
  height: 26px;
  overflow: hidden;
  font-size: 20px;
  font-weight: 400;
  line-height: 26px;
  color: #18191c;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.read-article-item__desc {
  height: 22px;
  margin: 4px 0 0;
  overflow: hidden;
  font-size: 12px;
  line-height: 22px;
  color: #9499a0;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.read-article-item__meta {
  display: flex;
  align-items: center;
  height: 24px;
  margin: 12px 0 0;
  font-size: 12px;
  color: #9499a0;
}

.read-article-item__author {
  display: flex;
  align-items: center;
  margin-right: 20px;
}

.read-article-item__avatar {
  width: 24px;
  height: 24px;
  margin-right: 5px;
  border-radius: 50%;
  overflow: hidden;
}

.read-article-item__avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.read-article-item__uname {
  max-width: 120px;
  overflow: hidden;
  font-size: 12px;
  line-height: 16px;
  color: #9499a0;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.read-article-item__label {
  box-sizing: content-box;
  padding: 0 5px;
  margin-right: 25px;
  font-size: 12px;
  line-height: 16px;
  color: #9499a0;
}

.read-article-item__stats {
  display: flex;
  align-items: center;
}

.read-article-item__stat {
  display: flex;
  align-items: center;
  margin-right: 25px;
  font-size: 12px;
  line-height: 16px;
  color: #9499a0;
}

.read-article-item__stat:last-child {
  margin-right: 0;
}

.read-article-item__cover {
  width: 117px;
  height: 88px;
  border-radius: 4px;
  overflow: hidden;
  background: #f1f2f3;
}

.read-article-item__cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.read-loading,
.read-empty {
  padding: 60px 0;
  text-align: center;
  color: #9499a0;
}
</style>