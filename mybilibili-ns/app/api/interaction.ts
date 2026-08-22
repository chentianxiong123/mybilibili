import api from './client'

export function followUser(mId: number, follow: boolean) {
  return api.post('/api/interaction/follow', { mId, follow })
}

export function checkFollow(mId: number) {
  return api.get(`/api/interaction/follow/check/${mId}`)
}

export function likeManuscript(aId: number, like: boolean) {
  return api.post('/api/interaction/like', { aId, like })
}

export function coinManuscript(aId: number, count: number) {
  return api.post('/api/interaction/coin', { aId, count })
}

export function collectManuscript(aId: number, collect: boolean) {
  return api.post('/api/interaction/collect', { aId, collect })
}

export function shareManuscript(aId: number) {
  return api.post('/api/interaction/share', { aId })
}

export function getInteractionStatus(aId: number) {
  return api.get(`/api/interaction/status/${aId}`)
}