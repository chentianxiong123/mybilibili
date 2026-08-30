import api from './client'

export async function getMyInfo() {
  try {
    const userStr = localStorage.getItem('user')
    if (!userStr) return { code: '0', data: null }
    const localUser = JSON.parse(userStr)
    // 兼容可能存在的 id 或 userId 或 user.id 嵌套结构
    const userId = localUser.id || localUser.userId || (localUser.user && localUser.user.id)
    if (!userId) return { code: '0', data: null }

    const res = await api.get(`/user/${userId}`)
    // 兼容后端 Result 统一封装格式
    const data = res?.data || res
    if (data) {
      localStorage.setItem('user', JSON.stringify(data))
    }
    return {
      code: '1',
      data: data
    }
  } catch (e) {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      return { code: '1', data: JSON.parse(userStr) }
    }
    return { code: '0', data: null }
  }
}
