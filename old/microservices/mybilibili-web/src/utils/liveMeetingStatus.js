export const LIVE_ROOM_STATUS = Object.freeze({
  LIVE: 'live',
  OFFLINE: 'offline'
})

export const MEETING_STATUS = Object.freeze({
  NOT_STARTED: 0,
  IN_PROGRESS: 1,
  ENDED: 2,
  PENDING_APPROVAL: 3,
  REJECTED: 4
})

const meetingStatusLabels = {
  [MEETING_STATUS.NOT_STARTED]: '未开始',
  [MEETING_STATUS.IN_PROGRESS]: '进行中',
  [MEETING_STATUS.ENDED]: '已结束',
  [MEETING_STATUS.PENDING_APPROVAL]: '待审批',
  [MEETING_STATUS.REJECTED]: '已拒绝'
}

const meetingStatusTypes = {
  [MEETING_STATUS.NOT_STARTED]: 'info',
  [MEETING_STATUS.IN_PROGRESS]: 'success',
  [MEETING_STATUS.ENDED]: 'default',
  [MEETING_STATUS.PENDING_APPROVAL]: 'warning',
  [MEETING_STATUS.REJECTED]: 'danger'
}

export function isLiveRoomStatus(status) {
  return status === LIVE_ROOM_STATUS.LIVE
}

export function isOfflineRoomStatus(status) {
  return status === LIVE_ROOM_STATUS.OFFLINE
}

export function getMeetingStatusText(status) {
  return meetingStatusLabels[status] || '未知'
}

export function getMeetingStatusType(status) {
  return meetingStatusTypes[status] || 'default'
}

export function getMeetingJoinBlockMessage(status) {
  if (status === MEETING_STATUS.PENDING_APPROVAL) return '会议预约待审批，暂不能加入'
  if (status === MEETING_STATUS.REJECTED) return '会议预约已被拒绝'
  if (status === MEETING_STATUS.ENDED) return '会议已结束'
  return ''
}
