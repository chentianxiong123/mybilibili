import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://112.74.99.5:3000/web/api'

// 获取直播分区
export async function getAreas() {
  const res = await axios.get(`${API_BASE}/live/area`)
  return res.data
}

// 获取直播首页数据
export async function getLiveIndexData() {
  const res = await axios.get(`${API_BASE}/live/data`)
  return res.data
}

// 获取直播房间列表
export async function getLiveListData(params: {
  parentAreaId: number
  areaId: number
  page?: number
  pageSize?: number
}) {
  const res = await axios.get(`${API_BASE}/live/room/list`, { params })
  return res.data
}

// 获取直播间信息
export async function getRoomInfo(roomId: number) {
  const res = await axios.get(`${API_BASE}/live/room/info`, { params: { roomId } })
  return res.data
}

// 获取弹幕配置
export async function getDanMuConfig(roomId: number) {
  const res = await axios.get(`${API_BASE}/live/room/danmu_config`, { params: { roomId } })
  return res.data
}
