<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { searchApi } from '../../api/search.ts'

const router = useRouter()

const searchText = ref('')
const showSearchDropdown = ref(false)
const showAllHistory = ref(false)
const isSearchFocused = ref(false)
const searchHistory = ref([])
const hotSearchList = ref([])
const suggestList = ref([])
let suggestTimer = null

watch(searchText, (val) => {
  clearTimeout(suggestTimer)
  if (!val || val.trim().length < 1) {
    suggestList.value = []
    return
  }
  suggestTimer = setTimeout(async () => {
    try {
      const res = await searchApi.getSearchSuggestions(val.trim())
      if (res.code === 200 && res.data) {
        suggestList.value = res.data.slice(0, 8)
      }
    } catch (e) {
      suggestList.value = []
    }
  }, 300)
})

const loadSearchHistory = () => {
  try {
    const stored = localStorage.getItem('searchHistory')
    if (stored) {
      searchHistory.value = JSON.parse(stored)
    }
  } catch (error) {
    console.error('加载搜索历史失败:', error)
    searchHistory.value = []
  }
}

const saveSearchHistory = () => {
  try {
    localStorage.setItem('searchHistory', JSON.stringify(searchHistory.value))
  } catch (error) {
    console.error('保存搜索历史失败:', error)
  }
}

const fetchHotSearch = async () => {
  try {
    const response = await searchApi.getHotSearch()
    if (response.code === 200 && response.data) {
      hotSearchList.value = response.data.map(item => ({
        rank: item.rank,
        keyword: item.keyword,
        score: item.score
      }))
    }
  } catch (error) {
    console.error('获取热搜榜失败:', error)
    hotSearchList.value = []
  }
}

const handleSearchFocus = () => {
  showSearchDropdown.value = true
  isSearchFocused.value = true
}

const handleSearchBlur = (event) => {
  const searchBox = document.querySelector('.search-box')
  if (searchBox && searchBox.contains(event.relatedTarget)) {
    return
  }
  
  setTimeout(() => {
    const searchBox = document.querySelector('.search-box')
    if (searchBox && !searchBox.contains(document.activeElement)) {
      showSearchDropdown.value = false
      isSearchFocused.value = false
    }
  }, 200)
}

const handleSearchContentMouseDown = (event) => {
  event.preventDefault()
}

const handleHistoryClick = (keyword) => {
  searchText.value = keyword
  handleSearch()
}

const handleHotSearchClick = (keyword) => {
  searchText.value = keyword
  handleSearch()
}

const handleSuggestClick = (keyword) => {
  searchText.value = keyword
  suggestList.value = []
  handleSearch()
}

const clearSearchHistory = () => {
  searchHistory.value = []
  saveSearchHistory()
}

const handleSearch = () => {
  if (searchText.value.trim()) {
    const keyword = searchText.value.trim()
    const index = searchHistory.value.indexOf(keyword)
    if (index > -1) {
      searchHistory.value.splice(index, 1)
    }
    searchHistory.value.unshift(keyword)
    if (searchHistory.value.length > 10) {
      searchHistory.value = searchHistory.value.slice(0, 10)
    }
    saveSearchHistory()
    router.push(`/search?keyword=${encodeURIComponent(keyword)}`)
    showSearchDropdown.value = false
  }
}

onMounted(() => {
  loadSearchHistory()
  fetchHotSearch()
})

onUnmounted(() => {
  if (suggestTimer) {
    clearTimeout(suggestTimer)
  }
})
</script>

<template>
  <div :class="['search-box', 'v-search-bar', { 'focused': isSearchFocused, 'is-focus': isSearchFocused }]">
    <el-input
      v-model="searchText"
      placeholder="搜索番剧、影视、UP主..."
      @keyup.enter="handleSearch"
      @focus="handleSearchFocus"
      @blur="handleSearchBlur"
      clearable
    >
      <template #suffix>
        <i class="iconfont icon-sousuo v-searchform-btn" @click="handleSearch"></i>
      </template>
    </el-input>
    <div class="search-dropdown v-search-panel" v-show="showSearchDropdown" @mousedown="handleSearchContentMouseDown">
      <div class="search-suggestions" v-if="suggestList.length > 0 && searchText.trim()">
        <div
          v-for="(item, index) in suggestList"
          :key="'suggest-' + index"
          class="suggest-item"
          @click="handleSuggestClick(item)"
        >
          <i class="iconfont icon-sousuo suggest-icon"></i>
          <span class="suggest-text">{{ item }}</span>
        </div>
      </div>
      <div class="search-history" v-if="searchHistory.length > 0 && !searchText.trim()">
        <div class="search-history-header v-search-panel-header">
          <span class="search-history-title">搜索历史</span>
          <span class="clear-history v-clear" @click="clearSearchHistory">清除</span>
        </div>
        <div class="search-history-list v-history-list" :class="{ 'expanded': showAllHistory }">
          <div
            v-for="(item, index) in searchHistory"
            :key="index"
            class="history-item v-history-item"
            @click="handleHistoryClick(item)"
          >
            {{ item }}
          </div>
        </div>
        <div class="history-more" v-if="searchHistory.length > 6">
          <span class="more-btn" @click="showAllHistory = !showAllHistory">
            {{ showAllHistory ? '收起' : '展开更多' }}
          </span>
        </div>
      </div>

      <div class="hot-search">
        <div class="hot-search-header v-search-panel-header">
          <span class="hot-search-title">热搜</span>
        </div>
        <div class="hot-search-list v-trending-list">
          <div
            v-for="item in hotSearchList"
            :key="item.rank"
            class="hot-item v-trending-item"
            @click="handleHotSearchClick(item.keyword)"
          >
            <span :class="['hot-rank', 'v-trending-rank', { 'top-three': item.rank <= 3 }]">{{ item.rank }}</span>
            <span class="hot-keyword v-trending-text">{{ item.keyword }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-box {
  width: 500px;
  display: flex;
  align-items: center;
  position: relative;
}

.search-box .el-input {
  border-radius: var(--v-radius, 8px);
}

.search-box :deep(.el-input__wrapper) {
  border-radius: var(--v-radius, 8px);
  background-color: var(--v-bg3, #F1F2F3);
  box-shadow: none;
  border: 1px solid transparent;
  height: 50px;
  transition: background-color 0.3s ease, box-shadow 0.3s ease, border-color 0.3s ease;
}

.search-box:hover :deep(.el-input__wrapper) {
  background-color: var(--v-bg1, #FFFFFF);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.search-box.focused :deep(.el-input__wrapper) {
  background-color: var(--v-bg1, #FFFFFF);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.12);
  border-color: var(--v-brand-pink, #FF6699);
}

.search-box .search-icon,
.search-box .v-searchform-btn {
  color: var(--v-text3, #9499A0);
  cursor: pointer;
  font-size: 18px;
}

.search-box .search-icon:hover,
.search-box .v-searchform-btn:hover {
  color: var(--v-brand-pink, #FF6699);
}

.search-suggestions {
  padding: 8px 0;
}

.suggest-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.suggest-item:hover {
  background: var(--v-bg2, #F6F7F8);
}

.suggest-icon {
  color: var(--v-text3, #9499A0);
  font-size: 14px;
  flex-shrink: 0;
}

.suggest-text {
  font-size: 14px;
  color: var(--v-text1, #18191C);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  width: 500px;
  background: var(--v-bg1, #FFFFFF);
  border-radius: 0 0 var(--v-radius, 8px) var(--v-radius, 8px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  margin-top: -1px;
  z-index: 1000;
  overflow: hidden;
  max-height: 500px;
  overflow-y: auto;
  border: 1px solid var(--v-line-light, #F1F2F3);
  border-top: none;
}

.search-history {
  padding: 16px;
  border-bottom: 1px solid var(--v-line-light, #F1F2F3);
}

.search-history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.search-history-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--v-text1, #18191C);
}

.clear-history {
  font-size: 12px;
  color: var(--v-text3, #9499A0);
  cursor: pointer;
  transition: color 0.3s;
}

.clear-history:hover {
  color: var(--v-brand-pink, #FF6699);
}

.search-history-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 70px;
  overflow: hidden;
  align-content: flex-start;
}

.search-history-list.expanded {
  max-height: none;
}

.history-item {
  padding: 6px 12px;
  background: var(--v-bg2, #F6F7F8);
  border-radius: var(--v-radius-sm, 4px);
  font-size: 13px;
  color: var(--v-text1, #18191C);
  cursor: pointer;
  transition: all 0.3s;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-item:hover {
  background: var(--v-bg3, #F1F2F3);
  color: var(--v-brand-pink, #FF6699);
}

.history-more {
  margin-top: 8px;
  text-align: center;
}

.more-btn {
  font-size: 12px;
  color: var(--v-brand-blue, #00AEEC);
  cursor: pointer;
  transition: color 0.3s;
}

.more-btn:hover {
  color: var(--v-brand-pink, #FF6699);
}

.hot-search {
  padding: 16px;
}

.hot-search-header {
  margin-bottom: 12px;
}

.hot-search-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--v-text1, #18191C);
}

.hot-search-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 16px;
}

.hot-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  cursor: pointer;
  transition: background 0.3s;
  border-radius: var(--v-radius-sm, 4px);
  min-width: 0;
}

.hot-item:hover {
  background: var(--v-bg2, #F6F7F8);
}

.hot-rank {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--v-text3, #9499A0);
  background: var(--v-bg2, #F6F7F8);
  border-radius: var(--v-radius-sm, 4px);
  flex-shrink: 0;
}

.hot-rank.top-three {
  background: var(--v-brand-pink, #FF6699);
  color: var(--v-bg1, #FFFFFF);
}

.hot-keyword {
  flex: 1;
  font-size: 13px;
  color: var(--v-text1, #18191C);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hot-item:hover .hot-keyword {
  color: var(--v-brand-pink, #FF6699);
}
</style>