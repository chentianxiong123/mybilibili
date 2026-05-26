import { liveApi } from '../../api/live.js'

const handleRes = (res) => {
  if (res && res.code === 200) return res.data || {}
  return {}
}

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

export async function getAreas() {
  try {
    const res = await liveApi.getLiveList()
    return {
      code: '1',
      data: (handleRes(res) || []).map(l => ({
        id: l.id,
        name: l.areaName || l.roomName || '未分类',
        parentId: l.parentAreaId || 0
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getLiveIndexData() {
  try {
    const res = await liveApi.getLiveList()
    const list = handleRes(res) || []
    const liveItems = list.map(adaptLive)
    return {
      code: '1',
      data: {
        bannerList: [],
        itemList: [{
          areaName: '热门直播',
          areaId: 0,
          lives: liveItems.slice(0, 6)
        }]
      }
    }
  } catch (e) {
    return { code: '0', data: { bannerList: [], itemList: [] } }
  }
}

export async function getLiveListData(params) {
  try {
    const res = await liveApi.getLiveList()
    return {
      code: '1',
      data: {
        areaName: params?.areaName || '直播列表',
        list: (handleRes(res) || []).map(adaptLive)
      }
    }
  } catch (e) {
    return { code: '0', data: { list: [] } }
  }
}

export async function getRoomInfo(roomId) {
  try {
    const res = await liveApi.getRoom(roomId)
    const data = handleRes(res)
    return {
      code: '1',
      data: {
        ...adaptLive(data),
        uId: data.userId,
        title: data.roomName || data.title,
        cover: data.coverUrl,
        playUrl: data.streamUrl
      }
    }
  } catch (e) {
    return { code: '0', data: null }
  }
}

export async function getDanMuConfig(roomId) {
  // 弹幕 WebSocket 配置 - 使用项目的 ws 端点
  return {
    code: '1',
    data: {
      host: location.host,
      port: ''
    }
  }
}