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

const liveStatusLabels = {
  [LIVE_ROOM_STATUS.LIVE]: '直播中',
  [LIVE_ROOM_STATUS.OFFLINE]: '离线'
}

const liveStatusTypes = {
  [LIVE_ROOM_STATUS.LIVE]: 'success',
  [LIVE_ROOM_STATUS.OFFLINE]: 'info'
}

const meetingStatusLabels = {
  [MEETING_STATUS.NOT_STARTED]: '未开始',
  [MEETING_STATUS.IN_PROGRESS]: '进行中',
  [MEETING_STATUS.ENDED]: '已结束',
  [MEETING_STATUS.PENDING_APPROVAL]: '待审批',
  [MEETING_STATUS.REJECTED]: '已拒绝'
}

const meetingStatusTypes = {
  [MEETING_STATUS.NOT_STARTED]: 'warning',
  [MEETING_STATUS.IN_PROGRESS]: 'success',
  [MEETING_STATUS.ENDED]: 'info',
  [MEETING_STATUS.PENDING_APPROVAL]: 'primary',
  [MEETING_STATUS.REJECTED]: 'danger'
}

export function isLiveRoomStatus(status) {
  return status === LIVE_ROOM_STATUS.LIVE
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

export function getMeetingStatusText(status) {
  return meetingStatusLabels[status] || '未知'
}

export function getMeetingStatusType(status) {
  return meetingStatusTypes[status] || 'info'
}

export function canForceEndMeeting(status) {
  return status === MEETING_STATUS.NOT_STARTED || status === MEETING_STATUS.IN_PROGRESS
}
