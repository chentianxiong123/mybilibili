<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../../components/Header.vue'
import Result from './Result.vue'
import { getHotwords, getSuggests } from '../../api/search'

const router = useRouter()
const searchValue = ref('')
const keyword = ref('')
const hotwords = ref([])
const suggestList = ref([])
const showSuggest = ref(false)
const searchHistories = ref([])

onMounted(async () => {
  try {
    const res = await getHotwords()
    if (res.code === '1') {
      hotwords.value = res.data || []
    }
  } catch (e) {
    console.error(e)
  }
  // 从 localStorage 加载历史
  try {
    const h = JSON.parse(localStorage.getItem('searchHistories') || '[]')
    searchHistories.value = h
  } catch (e) {}
})

const onInput = async (e) => {
  const v = e.target.value.trim()
  searchValue.value = v
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
  } catch (e) {}
}

const onSearch = () => {
  const v = searchValue.value.trim()
  if (!v) return
  keyword.value = v
  // 保存历史
  const h = searchHistories.value.filter(x => x !== v)
  h.unshift(v)
  searchHistories.value = h.slice(0, 20)
  localStorage.setItem('searchHistories', JSON.stringify(searchHistories.value))
  showSuggest.value = false
}

const clearHistory = () => {
  searchHistories.value = []
  localStorage.removeItem('searchHistories')
}

const onSuggestClick = (item) => {
  searchValue.value = item.value || item.name
  onSearch()
}
</script>

<template>
  <div class="search-page">
    <div class="search-header">
      <input
        class="search-input"
        v-model="searchValue"
        placeholder="搜索视频、UP主或AV号"
        maxlength="33"
        @input="onInput"
        @keydown.enter="onSearch"
      />
      <span v-if="searchValue" class="clear-btn" @click="searchValue = ''">✕</span>
      <span class="cancel-btn" @click="$router.back()">取消</span>
    </div>

    <div v-if="!keyword" class="search-body">
      <!-- 热搜 -->
      <div v-if="hotwords.length" class="hotwords-section">
        <div class="section-title">大家都在搜</div>
        <div class="hotwords">
          <span
            v-for="w in hotwords"
            :key="w.keyword"
            class="hotword-tag"
            @click="searchValue = w.keyword; onSearch()"
          >
            {{ w.keyword }}
          </span>
        </div>
      </div>

      <!-- 搜索建议 -->
      <div v-if="showSuggest && suggestList.length" class="suggest-list">
        <div
          v-for="item in suggestList"
          :key="item.value || item.name"
          class="suggest-item"
          @click="onSuggestClick(item)"
        >
          {{ item.name || item.value }}
        </div>
      </div>

      <!-- 历史记录 -->
      <div v-if="searchHistories.length" class="history-section">
        <div class="section-header">
          <span class="section-title">搜索历史</span>
          <span class="clear-history" @click="clearHistory">清除</span>
        </div>
        <div class="history-list">
          <span
            v-for="h in searchHistories"
            :key="h"
            class="history-tag"
            @click="searchValue = h; onSearch()"
          >
            {{ h }}
          </span>
        </div>
      </div>
    </div>

    <!-- 搜索结果 -->
    <Result v-else :keyword="keyword" />
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.search-page { min-height: 100vh; background: $bg-color; }

.search-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: $theme-pink;
  gap: 8px;
}

.search-input {
  flex: 1;
  height: 32px;
  border: none;
  border-radius: 16px;
  padding: 0 12px;
  font-size: 14px;
  outline: none;
}

.clear-btn {
  color: #999;
  font-size: 14px;
  cursor: pointer;
}

.cancel-btn {
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}

.search-body { padding: 12px; }

.hotwords-section, .history-section { margin-bottom: 16px; }

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.section-title { font-size: 15px; font-weight: 600; }

.clear-history {
  font-size: 12px;
  color: $text-secondary;
  cursor: pointer;
}

.hotwords, .history-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.hotword-tag, .history-tag {
  padding: 4px 12px;
  border-radius: 14px;
  background: $bg-white;
  font-size: 13px;
  color: $text-primary;
  cursor: pointer;

  &:active { background: $theme-pink; color: #fff; }
}

.suggest-list {
  background: $bg-white;
  border-radius: 8px;
  overflow: hidden;
}

.suggest-item {
  padding: 10px 14px;
  font-size: 14px;
  border-bottom: 1px solid $border-color;
  cursor: pointer;

  &:last-child { border-bottom: none; }
  &:active { background: $bg-color; }
}
</style>