// 直播 API - 复用 mybilibili-web 的接口
// /live/room/list → 直播间列表
// /live/room/{id} → 直播间详情
import api from './client'

const adaptLive = (l) => ({
  roomId: l.id,
  title: l.roomName || l.title || '',
  cover: l.coverUrl || l.cover,
  isLive: l.status === 'live' ? 1 : 0,
  onlineNum: l.viewerCount || 0,
  playUrl: l.streamUrl || l.playUrl,
  uname: l.anchorName || l.username || l.nickname || '',
  description: l.description || ''
})

// 直播间列表 - 复用 liveApi.getLiveList() → /live/room/list
export async function getAreas() {
  try {
    const res = await api.get('/live/room/list')
    const list = res?.data || res || []
    // 从直播间列表中提取分���信息
    const areaMap = new Map()
    list.forEach(l => {
      if (l.areaId && !areaMap.has(l.areaId)) {
        areaMap.set(l.areaId, { id: l.areaId, name: l.areaName || '未分类', parentId: l.parentAreaId || 0 })
      }
    })
    return {
      code: '1',
      data: Array.from(areaMap.values())
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 直播首页 - 复用 liveApi.getLiveList() → /live/room/list
export async function getLiveIndexData() {
  try {
    const res = await api.get('/live/room/list')
    const list = (res?.data || res || []).map(adaptLive)
    return {
      code: '1',
      data: {
        bannerList: [],
        itemList: [{
          areaName: '热门直播',
          areaId: 0,
          lives: list.slice(0, 6)
        }]
      }
    }
  } catch (e) {
    return { code: '0', data: { bannerList: [], itemList: [] } }
  }
}

// 直播列表 - 复用 liveApi.getLiveList() → /live/room/list
export async function getLiveListData() {
  try {
    const res = await api.get('/live/room/list')
    const list = (res?.data || res || []).map(adaptLive)
    return {
      code: '1',
      data: {
        areaName: '直播列表',
        list
      }
    }
  } catch (e) {
    return { code: '0', data: { list: [] } }
  }
}

// 直播间详情 - 复用 liveApi.getRoom() → /live/room/{id}
export async function getRoomInfo(roomId) {
  try {
    const res = await api.get(`/live/room/${roomId}`)
    const data = res?.data || res
    if (!data) return { code: '0', data: null }
    return {
      code: '1',
      data: {
        ...adaptLive(data),
        uId: data.userId,
        followerCount: data.followerCount || 0
      }
    }
  } catch (e) {
    return { code: '0', data: null }
  }
}

// 弹幕 WebSocket 配置
export function getDanMuConfig(roomId) {
  return {
    code: '1',
    data: {
      host: location.host,
      port: ''
    }
  }
}