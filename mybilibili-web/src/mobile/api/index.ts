import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 首页推荐视频 + 分区
export async function getHomeContent() {
  const res = await axios.get(`${API_BASE}/index`)
  return res.data
}

// 轮播图
export async function getBanners() {
  const res = await axios.get(`${API_BASE}/round-sowing`)
  return res.data
}