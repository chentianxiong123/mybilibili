import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 获取视频详情
export async function getVideoInfo(aId: number) {
  const res = await axios.get(`${API_BASE}/av/${aId}`)
  return res.data
}

// 获取推荐视频
export async function getRecommendVides(aId: number) {
  const res = await axios.get(`${API_BASE}/av/recommend/${aId}`)
  return res.data
}

// 获取评论
export async function getComments(aId: number, p: number) {
  const res = await axios.get(`${API_BASE}/av/replay`, { params: { aId, p } })
  return res.data
}

// 获取弹幕
export async function getBarrages(cId: number) {
  const res = await axios.get(`${API_BASE}/av/barrage/${cId}`)
  return res.data
}