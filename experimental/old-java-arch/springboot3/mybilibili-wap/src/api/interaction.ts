import api from './client'

// 关注/取消关注
export async function followUser(userId: number, follow: boolean) {
  try {
    const res = follow
      ? await api.post(`/follow/${userId}`)
      : await api.delete(`/follow/${userId}`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 检查是否关注
export async function checkFollow(userId: number) {
  try {
    const res = await api.get(`/follow/check/${userId}`)
    return { code: '1', data: res?.data || res }
  } catch (e) {
    return { code: '0', data: false }
  }
}

// 点赞/取消点赞稿件
export async function likeManuscript(manuscriptId: number, liked: boolean) {
  try {
    const res = liked
      ? await api.post(`/manuscript/${manuscriptId}/like`)
      : await api.delete(`/manuscript/${manuscriptId}/like`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 投币
export async function coinManuscript(manuscriptId: number, coinCount = 1) {
  try {
    const res = await api.post(`/manuscript/${manuscriptId}/coin?coinCount=${coinCount}`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 收藏/取消收藏
export async function collectManuscript(manuscriptId: number, collected: boolean) {
  try {
    const res = collected
      ? await api.post(`/manuscript/${manuscriptId}/collect`)
      : await api.delete(`/manuscript/${manuscriptId}/collect`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 分享稿件
export async function shareManuscript(manuscriptId: number) {
  try {
    const res = await api.post(`/manuscript/${manuscriptId}/share`, { channel: 'wap' })
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 获取互动状态
export async function getInteractionStatus(manuscriptId: number) {
  try {
    const res = await api.get(`/manuscript/${manuscriptId}/status`)
    return { code: '1', data: res?.data || res }
  } catch (e) {
    return { code: '0', data: null }
  }
}