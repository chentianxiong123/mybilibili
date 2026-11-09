export function getCurrentUser(options = {}) {
  const {
    fallbackId = 0,
    fallbackName = '游客',
    temporary = false,
    temporaryKey = 'tmpUserId',
    temporaryPrefix = '访客'
  } = options

  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    if (user && user.id) {
      return {
        id: Number(user.id),
        name: user.username || user.nickname || `用户${user.id}`
      }
    }
  } catch (e) {
    console.error('解析用户信息失败:', e)
  }

  if (temporary) {
    let tmp = sessionStorage.getItem(temporaryKey)
    if (!tmp) {
      tmp = String(Date.now() % 1000000)
      sessionStorage.setItem(temporaryKey, tmp)
    }
    return { id: Number(tmp), name: `${temporaryPrefix}${tmp}` }
  }

  return { id: fallbackId, name: fallbackName }
}
