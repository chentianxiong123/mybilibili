import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 获取 UP 主信息
export async function getUserInfo(mId: number) {
  const res = await axios.get(`${API_BASE}/up/${mId}`)
  return res.data
}

// 获取 UP 主投稿视频
export async function getUserVideos(mId: number, p: number, size: number) {
  const res = await axios.get(`${API_BASE}/up/video`, { params: { uId: mId, p, size } })
  return res.data
}