import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useLiveRoomSession } from '../composables/useLiveRoomSession.js'

const createSession = (options = {}) => {
  const api = options.api || {
    getRoom: vi.fn(async () => ({
      code: 200,
      data: options.room || { id: 1, userId: 7, streamKey: 'stream-1' }
    }))
  }
  const message = options.message || {
    error: vi.fn(),
    success: vi.fn()
  }
  const requestNotificationPermission = options.requestNotificationPermission || vi.fn()
  const copyText = options.copyText || vi.fn(async () => true)
  const getCurrentUserFn = options.getCurrentUserFn || vi.fn(() => ({ id: options.userId ?? 7, name: '观众A' }))
  const logger = options.logger || { error: vi.fn() }
  const order = []
  const nextTickFn = options.nextTickFn || vi.fn(async () => {
    order.push('nextTick')
  })

  const session = useLiveRoomSession({
    roomId: options.roomId || 'room-1',
    api,
    message,
    requestNotificationPermission,
    copyText,
    getCurrentUserFn,
    getShareUrl: () => options.shareUrl || 'https://example.com/live/room-1',
    nextTickFn,
    logger
  })

  return {
    api,
    message,
    requestNotificationPermission,
    copyText,
    getCurrentUserFn,
    logger,
    nextTickFn,
    order,
    ...session
  }
}

describe('useLiveRoomSession', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads a room, detects the streamer, syncs follow state, and initializes after next tick', async () => {
    const syncFollowState = vi.fn(() => {})
    const initPlayer = vi.fn(() => {})
    const session = createSession({
      room: { id: 1, userId: 7, streamKey: 'stream-1' },
      userId: 7
    })

    const room = await session.loadRoom({ syncFollowState, initPlayer })

    expect(session.api.getRoom).toHaveBeenCalledWith('room-1')
    expect(room).toEqual({ id: 1, userId: 7, streamKey: 'stream-1' })
    expect(session.room.value).toEqual(room)
    expect(session.isStreamer.value).toBe(true)
    expect(syncFollowState).toHaveBeenCalledWith(room)
    expect(session.loading.value).toBe(false)
    expect(session.nextTickFn).toHaveBeenCalled()
    expect(initPlayer).toHaveBeenCalled()
  })

  it('runs the full initialization sequence including notification, linkmic status, and websocket connect', async () => {
    const syncFollowState = vi.fn()
    const initPlayer = vi.fn()
    const loadLinkmicStatus = vi.fn(async () => {})
    const connectLiveRoomSocket = vi.fn()
    const session = createSession()

    await session.initializeLiveRoom({
      syncFollowState,
      initPlayer,
      loadLinkmicStatus,
      connectLiveRoomSocket
    })

    expect(session.requestNotificationPermission).toHaveBeenCalled()
    expect(syncFollowState).toHaveBeenCalledWith(session.room.value)
    expect(initPlayer).toHaveBeenCalled()
    expect(loadLinkmicStatus).toHaveBeenCalled()
    expect(connectLiveRoomSocket).toHaveBeenCalled()
  })

  it('reports load failures and still finishes the post-load initialization hooks', async () => {
    const api = {
      getRoom: vi.fn(async () => ({ code: 404, message: '不存在' }))
    }
    const initPlayer = vi.fn()
    const loadLinkmicStatus = vi.fn(async () => {})
    const connectLiveRoomSocket = vi.fn()
    const session = createSession({ api })

    await session.initializeLiveRoom({
      initPlayer,
      loadLinkmicStatus,
      connectLiveRoomSocket
    })

    expect(session.room.value).toBeNull()
    expect(session.loading.value).toBe(false)
    expect(session.message.error).toHaveBeenCalledWith('不存在')
    expect(initPlayer).not.toHaveBeenCalled()
    expect(loadLinkmicStatus).toHaveBeenCalled()
    expect(connectLiveRoomSocket).toHaveBeenCalled()
  })

  it('logs thrown load errors and copies the share url with the live-room message', async () => {
    const api = {
      getRoom: vi.fn(async () => {
        throw new Error('network')
      })
    }
    const session = createSession({ api, shareUrl: 'https://example.com/live/123' })

    await session.loadRoom()
    await session.shareRoom()

    expect(session.logger.error).toHaveBeenCalledWith('[LiveRoom] 加载失败:', expect.any(Error))
    expect(session.message.error).toHaveBeenCalledWith('加载失败')
    expect(session.copyText).toHaveBeenCalledWith('https://example.com/live/123', {
      successMessage: '直播间链接已复制，可分享给好友！',
      message: session.message
    })
  })
})
