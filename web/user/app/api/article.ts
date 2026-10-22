import api from './client'

export interface ArticleSavePayload {
  id?: number
  title: string
  content: string
  cover_url?: string
  topics?: string
  collection_name?: string
  visibility?: number
  comment_enabled?: number
  featured_comment?: number
  is_original?: number
  ai_declaration?: number
  word_count?: number
  publish?: boolean
}

export const articleApi = {
  getDrafts: (keyword = '') => api.get('/article/drafts', { params: { keyword } }),
  getMine: (status: 'all' | '0' | '1' = 'all') => api.get('/article/mine', { params: { status } }),
  getUserArticles: (uid: number, page = 1, size = 20) =>
    api.get(`/article/user/${uid}`, { params: { page, size } }),
  getFeed: (page = 1, size = 20) => api.get('/article/feed', { params: { page, size } }),
  getArticle: (id: number) => api.get(`/article/${id}`),
  save: (payload: ArticleSavePayload) => api.post('/article/save', payload),
  remove: (id: number) => api.delete(`/article/${id}`),
  view: (id: number) => api.post(`/article/${id}/view`),
  getInteraction: (id: number) => api.get(`/article/${id}/interaction`),
  like: (id: number) => api.post(`/article/${id}/like`),
  coin: (id: number, count = 1) => api.post(`/article/${id}/coin`, { count }),
  favorite: (id: number) => api.post(`/article/${id}/favorite`),
  share: (id: number) => api.post(`/article/${id}/share`),
  getComments: (id: number, page = 1, size = 20) =>
    api.get(`/article/${id}/comments`, { params: { page, size } }),
  postComment: (id: number, payload: { content: string; parent_id?: number; reply_to_user_id?: number }) =>
    api.post(`/article/${id}/comments`, payload),
  deleteComment: (cid: number) => api.delete(`/article/comment/${cid}`),
  likeComment: (cid: number) => api.post(`/article/comment/${cid}/like`),
  getFolders: (articleId: number) => api.get('/article/folders', { params: { article_id: articleId } }),
  createFolder: (name: string) => api.post('/article/folders', { name }),
  collect: (id: number, adds: number[], removes: number[]) =>
    api.post(`/article/${id}/collect`, { adds, removes }),
  mentionSearch: (keyword: string) =>
    api.get('/article/mention-search', { params: { keyword } }),
  uploadImage: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post('/banner/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
  }
}

export default articleApi
