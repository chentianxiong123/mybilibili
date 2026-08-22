import api from './client'

export function postComment(aId: number, content: string) {
  return api.post('/api/comment/post', { aId, content })
}

export function replyComment(aId: number, rpId: number, content: string) {
  return api.post('/api/comment/reply', { aId, rpId, content })
}

export function likeComment(rpId: number) {
  return api.post('/api/comment/like', { rpId })
}