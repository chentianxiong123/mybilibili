import { describe, it, expect } from 'vitest'
import {
  LIVE_ROOM_STATUS,
  isLiveRoomStatus,
  isOfflineRoomStatus,
  getLiveStatusText,
  getLiveStatusType,
  getNextLiveStatus,
} from './liveMeetingStatus'

describe('LIVE_ROOM_STATUS', () => {
  it('has correct values', () => {
    expect(LIVE_ROOM_STATUS.LIVE).toBe('live')
    expect(LIVE_ROOM_STATUS.OFFLINE).toBe('offline')
  })

  it('is frozen', () => {
    expect(Object.isFrozen(LIVE_ROOM_STATUS)).toBe(true)
  })
})

describe('isLiveRoomStatus', () => {
  it('returns true for live', () => {
    expect(isLiveRoomStatus('live')).toBe(true)
  })

  it('returns false for offline', () => {
    expect(isLiveRoomStatus('offline')).toBe(false)
  })

  it('returns false for unknown', () => {
    expect(isLiveRoomStatus('unknown')).toBe(false)
  })
})

describe('isOfflineRoomStatus', () => {
  it('returns true for offline', () => {
    expect(isOfflineRoomStatus('offline')).toBe(true)
  })

  it('returns false for live', () => {
    expect(isOfflineRoomStatus('live')).toBe(false)
  })
})

describe('getLiveStatusText', () => {
  it('returns 直播中 for live', () => {
    expect(getLiveStatusText('live')).toBe('直播中')
  })

  it('returns 离线 for offline', () => {
    expect(getLiveStatusText('offline')).toBe('离线')
  })

  it('returns 未知 for unknown status', () => {
    expect(getLiveStatusText('other')).toBe('未知')
  })
})

describe('getLiveStatusType', () => {
  it('returns success for live', () => {
    expect(getLiveStatusType('live')).toBe('success')
  })

  it('returns info for offline', () => {
    expect(getLiveStatusType('offline')).toBe('info')
  })

  it('returns info for unknown', () => {
    expect(getLiveStatusType('other')).toBe('info')
  })
})

describe('getNextLiveStatus', () => {
  it('returns offline when live', () => {
    expect(getNextLiveStatus('live')).toBe('offline')
  })

  it('returns live when offline', () => {
    expect(getNextLiveStatus('offline')).toBe('live')
  })
})
