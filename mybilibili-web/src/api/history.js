import api from './index.js'

const authRequired = (data) => (
  localStorage.getItem('token')
    ? null
    : Promise.resolve({ code: 401, message: '请先登录', data })
)

export const historyApi = {
  getHistoryList: (page = 1, size = 10) => {
    return authRequired([]) || api.get('/history/list', { params: { page, size } })
  },

  clearHistory: () => {
    return authRequired(null) || api.delete('/history/clear')
  },

  deleteHistoryItem: (id) => {
    return authRequired(null) || api.delete(`/history/${id}`)
  }
}

export default historyApi
