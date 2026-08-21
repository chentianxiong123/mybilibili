import request from './client'

export const transcodeConfigApi = {
  getConfig() {
    return request.get('/admin/transcode-config')
  },
  updateConfig(data: { encoder: string }) {
    return request.put('/admin/transcode-config', data)
  }
}