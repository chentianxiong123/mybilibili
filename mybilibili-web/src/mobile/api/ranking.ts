import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 获取排行榜数据
export async function getRankings(rId: number) {
  const res = await axios.get(`${API_BASE}/ranking/${rId}`)
  return res.data
}

// 获取分区排行榜
export async function getRankingRegion(params: { rId: number, day?: number }) {
  const res = await axios.get(`${API_BASE}/ranking/region`, { params })
  return res.data
}

// 获取最新分区排行
export async function getRankingArchive(params: { rId: number, p?: number, size?: number }) {
  const res = await axios.get(`${API_BASE}/ranking/archive`, { params })
  return res.data
}