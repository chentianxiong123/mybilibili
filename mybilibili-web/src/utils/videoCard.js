export const toDurationSeconds = (value) => {
  if (value === null || value === undefined || value === '') return 0
  if (typeof value === 'number') return value

  const numericValue = Number(value)
  if (!Number.isNaN(numericValue)) return numericValue

  const parts = String(value).split(':').map(part => Number(part))
  if (parts.some(part => Number.isNaN(part))) return 0
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  return 0
}

export const formatDuration = (seconds) => {
  const totalSeconds = toDurationSeconds(seconds)
  if (!totalSeconds || totalSeconds <= 0) return '00:00'

  const hours = Math.floor(totalSeconds / 3600)
  const mins = Math.floor((totalSeconds % 3600) / 60)
  const secs = Math.floor(totalSeconds % 60)

  if (hours > 0) {
    return `${hours}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

export const formatMonthDay = (dateStr) => {
  if (!dateStr) return ''
  const date = parseDate(dateStr)
  if (Number.isNaN(date.getTime())) return ''

  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}-${day}`
}

const parseDate = (value) => {
  if (value instanceof Date) return value
  if (typeof value === 'number') return new Date(value)
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return new Date(NaN)
    if (/^\d+$/.test(trimmed)) return new Date(Number(trimmed))
    return new Date(trimmed.replace(' ', 'T'))
  }
  if (Array.isArray(value)) {
    const [year, month = 1, day = 1, hour = 0, minute = 0, second = 0] = value
    return new Date(year, month - 1, day, hour, minute, second)
  }
  if (typeof value === 'object') {
    if (value.date) return parseDate(value.date)
    if (value.time) return parseDate(value.time)
    if (value.timestamp) return parseDate(value.timestamp)
  }
  return new Date(NaN)
}

export const normalizeVideoCard = (item) => {
  if (!item) return null

  const firstVideo = Array.isArray(item.videos) && item.videos.length > 0 ? item.videos[0] : null
  const manuscriptId = item.manuscriptId || item.id
  if (!manuscriptId) return null

  const durationSeconds = toDurationSeconds(
    item.durationSeconds ?? firstVideo?.durationSeconds ?? item.duration ?? firstVideo?.duration
  )
  const uploader = item.uploader || {
    id: item.userId,
    name: item.userName || item.username || item.nickname || item.author || '未知UP主',
    avatar: item.userAvatar || item.avatar
  }

  const uploadTime = item.uploadTime || item.upload_time || item.publishDate || item.publishTime ||
    item.releaseTime || item.createTime || item.createdAt || item.updatedAt || item.reviewTime ||
    firstVideo?.uploadTime || firstVideo?.createdAt

  return {
    id: item.firstVideoId || item.videoId || firstVideo?.id || manuscriptId,
    title: item.title,
    coverUrl: item.coverUrl || item.cover,
    viewCount: item.viewCount || item.play || 0,
    commentCount: item.commentCount || item.videoReview || item.danmakuCount || 0,
    uploadTime,
    dateText: formatMonthDay(uploadTime),
    uploader,
    duration: item.duration || firstVideo?.duration || formatDuration(durationSeconds),
    durationSeconds,
    manuscriptId,
    manuscript: item
  }
}
