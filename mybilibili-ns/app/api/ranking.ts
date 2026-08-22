import api from './client'

export function getRanking(rId: number) {
  return api.get(`/api/ranking/${rId}`)
}