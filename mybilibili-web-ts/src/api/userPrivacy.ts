import api from './client'

const defaultPrivacySettings = {
  code: 200,
  message: '接口不存在，返回默认设置',
  data: {
    publicCollection: true,
    publicBirthdayTags: false,
    publicCoinVideos: false,
    publicLikeVideos: false,
    publicFollowingList: false,
    publicFollowersList: false,
    tags: []
  }
}

export const userPrivacyApi = {
  getPrivacySettings: async () => {
    try {
      return await api.get('/user/privacy/settings')
    } catch (err: any) {
      if (err?.response?.status === 404) {
        return defaultPrivacySettings
      }
      throw err
    }
  },
  updatePrivacySettings: (data: any) => api.put('/user/privacy/settings', data),
  getUserTags: () => api.get('/user/privacy/tags'),
  addUserTag: (tagName: string) => api.post('/user/privacy/tags', null, { params: { tagName } }),
  removeUserTag: (tagName: string) => api.delete('/user/privacy/tags', { params: { tagName } })
}

export default userPrivacyApi