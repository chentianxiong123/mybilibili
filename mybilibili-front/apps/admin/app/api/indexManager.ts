import request from './client'

export const indexManagerApi = {
  getStatus() {
    return request.get('/search/admin/index/status')
  },
  validate() {
    return request.post('/search/admin/index/validate')
  },
  rebuild() {
    return request.post('/search/admin/index/rebuild')
  }
}