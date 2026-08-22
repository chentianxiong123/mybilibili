import api from './client'

export function getHomeContent() {
  return api.get('/api/home/content')
}

export function getBanners() {
  return api.get('/api/banner-images')
}

export function getVideosByCategory(categoryId: number) {
  return api.get(`/api/videos/category/${categoryId}`)
}