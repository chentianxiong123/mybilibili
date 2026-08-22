import api from './client'

export function getMessages(page: number = 1) {
  return api.get(`/api/message/list?page=${page}`)
}

export function getChatList() {
  return api.get('/api/message/chat/list')
}

export function getChatMessages(userId: number, page: number = 1) {
  return api.get(`/api/message/chat/${userId}?page=${page}`)
}

export function sendMessage(toUserId: number, content: string) {
  return api.post('/api/message/send', { toUserId, content })
}