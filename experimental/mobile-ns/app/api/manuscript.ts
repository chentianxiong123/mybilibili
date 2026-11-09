import api from './client'

export function getManuscripts(page: number = 1) {
  return api.get(`/api/manuscript/list?page=${page}`)
}