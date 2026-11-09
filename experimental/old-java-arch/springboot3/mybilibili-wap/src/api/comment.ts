import api from './client'

// 发表评论
export async function postComment(manuscriptId: number, content: string) {
  try {
    const res = await api.post('/comment/add', `manuscriptId=${manuscriptId}&content=${encodeURIComponent(content)}`, {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    })
    return { code: '1', data: res?.data || res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 回复评论
export async function replyComment(commentId: number, content: string, replyToUserId?: number) {
  try {
    let url = `commentId=${commentId}&content=${encodeURIComponent(content)}`
    if (replyToUserId) url += `&replyToUserId=${replyToUserId}`
    const res = await api.post('/comment/reply', url, {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    })
    return { code: '1', data: res?.data || res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 点赞评论
export async function likeComment(commentId: number) {
  try {
    const res = await api.post(`/comment/${commentId}/like`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 取消点赞评论
export async function unlikeComment(commentId: number) {
  try {
    const res = await api.delete(`/comment/${commentId}/like`)
    return { code: '1', data: res }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 获取回复列表
export async function getReplies(commentId: number, page = 1, size = 7) {
  try {
    const res = await api.get(`/comment/${commentId}/replies?page=${page}&size=${size}`)
    return { code: '1', data: res?.data || res || [] }
  } catch (e) {
    return { code: '0', data: [] }
  }
}