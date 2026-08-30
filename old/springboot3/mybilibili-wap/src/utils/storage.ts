// 本地存储工具 - 翻译自 react-bilibili/src/util/storage.ts
const VIEW_HISTORY = 'view_history'
const SEARCH_HISTORY = 'search_history'

export interface ViewHistory {
  aId: number
  title: string
  pic: string
  viewAt: number
}

export interface SearchHistory {
  value: string
  timestamp: number
}

export default {
  /** 获取播放历史 */
  getViewHistory(): ViewHistory[] {
    const item = window.localStorage.getItem(VIEW_HISTORY)
    return item ? JSON.parse(item) : []
  },

  /** 添加播放历史 */
  setViewHistory(history: ViewHistory): void {
    let viewHistory: ViewHistory[] = []
    const item = window.localStorage.getItem(VIEW_HISTORY)
    if (item) {
      viewHistory = JSON.parse(item)
    }
    const findIndex = viewHistory.findIndex((view) => view.aId === history.aId)
    if (findIndex !== -1) {
      viewHistory.splice(findIndex, 1)
    }
    viewHistory.push(history)
    window.localStorage.removeItem(VIEW_HISTORY)
    window.localStorage.setItem(VIEW_HISTORY, JSON.stringify(viewHistory))
  },

  /** 获取搜索历史 */
  getSearchHistory(): SearchHistory[] {
    const item = window.localStorage.getItem(SEARCH_HISTORY)
    return item ? JSON.parse(item) : []
  },

  /** 添加搜索历史 */
  setSearchHistory(history: SearchHistory): void {
    let searchHistory: SearchHistory[] = []
    const item = window.localStorage.getItem(SEARCH_HISTORY)
    if (item) {
      searchHistory = JSON.parse(item)
    }
    if (searchHistory.findIndex((view) => view.value === history.value) === -1) {
      searchHistory.push(history)
      this.clearSearchHistory()
      window.localStorage.setItem(SEARCH_HISTORY, JSON.stringify(searchHistory))
    }
  },

  /** 清空搜索历史 */
  clearSearchHistory(): void {
    window.localStorage.removeItem(SEARCH_HISTORY)
  }
}
