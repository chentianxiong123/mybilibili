import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 获取排行榜分区
export async function getRankingPartitions() {
  const res = await axios.get(`${API_BASE}/ranking/partitions`)
  return res.data
}

// 获取分区排行榜（7天）
export async function getRankingRegion(rId: number) {
  const res = await axios.get(`${API_BASE}/ranking/region`, { params: { rId, day: 7 } })
  return res.data
}

// 获取视频分区列表
export async function getPartitions() {
  const res = await axios.get(`${API_BASE}/partitions`)
  return res.data
}