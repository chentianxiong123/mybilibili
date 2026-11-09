import request from './request'

export const recommendConfigApi = {
  getConfig() {
    return request.get('/search/admin/recommend-config')
  },
  updateConfig(data) {
    return request.put('/search/admin/recommend-config', data)
  },
  resetConfig() {
    return request.post('/search/admin/recommend-config/reset')
  }
}
