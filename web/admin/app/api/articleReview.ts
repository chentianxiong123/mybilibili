import api from './client'

// 专栏审核（对齐视频稿件审核：待审 → 通过/拒绝）
export const getArticleReviewList = (params: { page?: number; size?: number; status?: string; keyword?: string }) =>
  api.get('/article/admin/all', { params })

export const getPendingArticles = (params: { page?: number; size?: number; keyword?: string } = {}) =>
  api.get('/article/admin/pending', { params })

export const getArticleReviewStatistics = () => api.get('/article/admin/statistics')

export const reviewArticle = (id: number, approved: boolean, reason = '') =>
  api.post(`/article/admin/review/${id}`, { approved, reason })

export const getArticleDetail = (id: number) => api.get(`/article/${id}`)
