// 本地存储工具（保持旧接口兼容）
// 注：完整存储能力见 utils/storage_layer.ts（SQLite 优先 + localStorage 兜底 + 服务端同步）
// 此文件仅保留 vue 页面仍在使用的 PlayHistory 记写。
const VIEW_HISTORY = 'view_history'

export interface ViewHistory {
  aId: number
  title: string
  pic: string
  viewAt: number
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
      try {
        viewHistory = JSON.parse(item)
      } catch {
        viewHistory = []
      }
    }
    const findIndex = viewHistory.findIndex((view) => view.aId === history.aId)
    if (findIndex !== -1) {
      viewHistory.splice(findIndex, 1)
    }
    viewHistory.push(history)
    window.localStorage.removeItem(VIEW_HISTORY)
    window.localStorage.setItem(VIEW_HISTORY, JSON.stringify(viewHistory))
  }
}