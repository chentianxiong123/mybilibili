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
  // 标签是独立资源，实现在后端 /api/v1/user/tags（handleTags）。
  // 原来自定义在 /user/privacy/tags 是错位：旧 handlePrivacy 只处理
  // GET(隐私JSON)/PUT，POST/DELETE 直接空响应，标签从没生效过。
  getUserTags: () => api.get('/user/tags'),
  addUserTag: (tagName: string) => api.post('/user/tags', null, { params: { tagName } }),
  removeUserTag: (tagName: string) => api.delete('/user/tags', { params: { tagName } })
}

export default userPrivacyApi