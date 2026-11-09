<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Result from './Result.vue'
import { getHotwords, getSuggests } from '../../api/search'

const router = useRouter()
const route = useRoute()
const searchValue = ref('')
const keyword = ref('')
const hotwords = ref([])
const suggestList = ref([])
const showSuggest = ref(false)
const searchHistories = ref([])

const discoveryWords = computed(() => {
  const fromHot = hotwords.value.map(item => item.keyword).filter(Boolean)
  const source = [...fromHot, ...searchHistories.value]
  return [...new Set(source)].slice(0, 12)
})

onMounted(async () => {
  try {
    const res = await getHotwords()
    if (res.code === '1') hotwords.value = res.data || []
  } catch (e) {
    hotwords.value = []
  }
  try {
    searchHistories.value = JSON.parse(localStorage.getItem('searchHistories') || '[]')
  } catch (e) {
    searchHistories.value = []
  }
  if (route.query.keyword) {
    onSearch(route.query.keyword)
  }
})

const onInput = async () => {
  const v = searchValue.value.trim()
  if (!v) {
    suggestList.value = []
    showSuggest.value = false
    return
  }
  try {
    const res = await getSuggests(v)
    if (res.code === '1') {
      suggestList.value = res.data || []
      showSuggest.value = true
    }
  } catch (e) {
    suggestList.value = []
  }
}

const onSearch = (value = searchValue.value) => {
  const v = String(value || '').trim()
  if (!v) return
  searchValue.value = v
  keyword.value = v
  const next = searchHistories.value.filter(x => x !== v)
  next.unshift(v)
  searchHistories.value = next.slice(0, 20)
  localStorage.setItem('searchHistories', JSON.stringify(searchHistories.value))
  showSuggest.value = false
}

const clearHistory = () => {
  searchHistories.value = []
  localStorage.removeItem('searchHistories')
}

const goBack = () => {
  if (keyword.value) {
    keyword.value = ''
    return
  }
  router.back()
}

const goHot = () => router.push('/m/search/hot')
</script>

<template>
  <div class="search-page">
    <header class="search-topbar">
      <button class="back-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <div class="search-box">
        <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></svg>
        <input
          v-model="searchValue"
          placeholder="ReLife_AnyTime · 昨天更新"
          maxlength="33"
          @input="onInput"
          @keydown.enter="onSearch()"
        />
        <button v-if="searchValue" @click="searchValue = ''; keyword = ''; showSuggest = false" aria-label="清空">×</button>
      </div>
      <button class="search-btn" @click="onSearch()">搜索</button>
    </header>

    <Result v-if="keyword" :keyword="keyword" />

    <main v-else class="search-home">
      <section class="hot-section">
        <div class="section-heading">
          <h2>bilibili热搜</h2>
          <button @click="goHot">完整榜单 <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></button>
        </div>
        <div class="hot-grid">
          <button
            v-for="(item, index) in hotwords.slice(0, 10)"
            :key="item.keyword + index"
            @click="onSearch(item.keyword)"
          >
            <span>{{ item.keyword }}</span>
            <i v-if="index === 0 || index === 4">新</i>
            <i v-else-if="index === 3" class="hot">热</i>
            <i v-else-if="index % 3 === 2" class="exclusive">独家</i>
          </button>
        </div>
      </section>

      <section v-if="showSuggest && suggestList.length" class="suggest-section">
        <button v-for="item in suggestList" :key="item.value || item.name" @click="onSearch(item.value || item.name)">
          {{ item.name || item.value }}
        </button>
      </section>

      <section class="history-section">
        <div class="section-heading">
          <h2>搜索历史</h2>
          <button class="icon-only" @click="clearHistory" aria-label="清空历史">
            <svg viewBox="0 0 24 24"><path d="M3 6h18M8 6V4h8v2M6 6l1 15h10l1-15" /></svg>
          </button>
        </div>
        <div v-if="searchHistories.length" class="pill-grid history-grid">
          <button v-for="h in searchHistories.slice(0, 8)" :key="h" @click="onSearch(h)">{{ h }}</button>
        </div>
        <p v-else class="empty-text">暂无搜索历史</p>
      </section>

      <section class="discover-section">
        <div class="section-heading">
          <h2>搜索发现</h2>
          <div class="heading-actions">
            <svg viewBox="0 0 24 24"><path d="M21 12a9 9 0 0 1-9 9 9 9 0 0 1-8.5-6M3 12a9 9 0 0 1 15.5-6" /><path d="M3 4v6h6M21 20v-6h-6" /></svg>
            <span></span>
            <svg viewBox="0 0 24 24"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7S1 12 1 12z" /><circle cx="12" cy="12" r="3" /></svg>
          </div>
        </div>
        <div v-if="discoveryWords.length" class="pill-grid discover-grid">
          <button v-for="word in discoveryWords" :key="word" @click="onSearch(word)">{{ word }}</button>
        </div>
        <p v-else class="empty-text">暂无发现词</p>
      </section>

      <button class="feedback-btn">反馈 <svg viewBox="0 0 24 24"><path d="M12 20h9" /><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" /></svg></button>
    </main>
  </div>
</template>

<style scoped lang="scss">
.search-page {
  min-height: 100vh;
  background: #fff;
  color: #18191c;
}

.search-topbar {
  position: sticky;
  top: 0;
  z-index: 40;
  height: 62px;
  display: grid;
  grid-template-columns: 42px 1fr 58px;
  align-items: center;
  gap: 8px;
  padding: 10px 18px 8px;
  background: #fff;
  box-sizing: border-box;
}

.back-btn,
.search-btn,
.search-box button {
  border: 0;
  background: transparent;
}

.back-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  color: #61666d;

  svg {
    width: 32px;
    height: 32px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.search-box {
  height: 40px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  border: 1.5px solid #9499a0;
  border-radius: 22px;
  box-sizing: border-box;

  svg {
    width: 22px;
    height: 22px;
    flex: 0 0 auto;
    fill: none;
    stroke: #7a7f87;
    stroke-width: 2;
  }

  input {
    min-width: 0;
    flex: 1;
    border: 0;
    outline: none;
    color: #18191c;
    font-size: 18px;
    background: transparent;
  }

  button {
    color: #9499a0;
    font-size: 20px;
  }
}

.search-btn {
  color: #fb7299;
  font-size: 20px;
  font-weight: 600;
}

.search-home {
  padding: 16px 28px 40px;
}

.section-heading {
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  h2 {
    margin: 0;
    color: #18191c;
    font-size: 20px;
    font-weight: 700;
  }

  button {
    border: 0;
    background: transparent;
    color: #9499a0;
    display: inline-flex;
    align-items: center;
    font-size: 17px;

    svg {
      width: 21px;
      height: 21px;
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
    }
  }
}

.hot-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px 42px;
  margin-bottom: 42px;

  button {
    min-width: 0;
    border: 0;
    padding: 0;
    display: flex;
    align-items: center;
    gap: 7px;
    background: transparent;
    color: #18191c;
    text-align: left;
    font-size: 20px;
  }

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  i {
    flex: 0 0 auto;
    min-width: 19px;
    height: 19px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    color: #fff;
    background: #ffbd2f;
    font-size: 13px;
    font-style: normal;
    font-weight: 700;

    &.hot {
      background: #f65c5c;
    }

    &.exclusive {
      min-width: 32px;
      background: #f25d9c;
    }
  }
}

.suggest-section {
  margin: -24px 0 30px;
  border-radius: 10px;
  overflow: hidden;
  background: #f6f7f9;

  button {
    width: 100%;
    height: 42px;
    border: 0;
    padding: 0 14px;
    text-align: left;
    color: #18191c;
    background: transparent;
    font-size: 16px;
  }
}

.history-section,
.discover-section {
  margin-top: 28px;
}

.icon-only svg,
.heading-actions svg {
  width: 24px;
  height: 24px;
  fill: none;
  stroke: #9499a0;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.heading-actions {
  display: flex;
  align-items: center;
  gap: 12px;

  span {
    width: 1px;
    height: 18px;
    background: #e3e5e7;
  }
}

.pill-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;

  button {
    min-width: 0;
    height: 46px;
    border: 0;
    padding: 0 16px;
    background: #f6f7f9;
    color: #18191c;
    text-align: left;
    font-size: 19px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.history-grid button {
  height: 42px;
}

.empty-text {
  margin: 0;
  color: #9499a0;
  font-size: 15px;
}

.feedback-btn {
  height: 36px;
  margin: 34px auto 0;
  padding: 0 16px;
  border: 1px solid #e3e5e7;
  border-radius: 18px;
  display: flex;
  align-items: center;
  gap: 5px;
  background: #fff;
  color: #61666d;
  font-size: 16px;

  svg {
    width: 19px;
    height: 19px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }
}

@media (max-width: 390px) {
  .search-home {
    padding-inline: 20px;
  }

  .hot-grid {
    column-gap: 22px;
  }

  .hot-grid button,
  .pill-grid button {
    font-size: 17px;
  }
}
</style>
