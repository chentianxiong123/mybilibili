export const LIVE_ROOM_STATUS = Object.freeze({
  LIVE: 'live',
  OFFLINE: 'offline'
})

const liveStatusLabels = {
  [LIVE_ROOM_STATUS.LIVE]: '直播中',
  [LIVE_ROOM_STATUS.OFFLINE]: '离线'
}

const liveStatusTypes = {
  [LIVE_ROOM_STATUS.LIVE]: 'success',
  [LIVE_ROOM_STATUS.OFFLINE]: 'info'
}

export function isLiveRoomStatus(status) {
  return status === LIVE_ROOM_STATUS.LIVE
}

export function isOfflineRoomStatus(status) {
  return status === LIVE_ROOM_STATUS.OFFLINE
}

export function getLiveStatusText(status) {
  return liveStatusLabels[status] || '未知'
}

export function getLiveStatusType(status) {
  return liveStatusTypes[status] || 'info'
}

export function getNextLiveStatus(status) {
  return isLiveRoomStatus(status) ? LIVE_ROOM_STATUS.OFFLINE : LIVE_ROOM_STATUS.LIVE
}
