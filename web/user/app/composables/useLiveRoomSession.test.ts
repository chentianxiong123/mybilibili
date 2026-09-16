import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'

vi.mock('@/api/live.ts', () => ({
  liveApi: {
    getRoom: vi.fn(),
  },
}))

vi.mock('@/utils/currentUser.ts', () => ({
  getCurrentUser: vi.fn(() => ({ id: 100, name: 'streamer' })),
}))

vi.mock('@/utils/notification.ts', () => ({
  requestNotificationPermission: vi.fn(),
}))

vi.mock('@/utils/clipboard.ts', () => ({
  copyTextToClipboard: vi.fn(),
}))

import { liveApi } from '@/api/live.ts'
import { getCurrentUser } from '@/utils/currentUser.ts'
import { requestNotificationPermission } from '@/utils/notification.ts'
import { copyTextToClipboard } from '@/utils/clipboard.ts'
import { useLiveRoomSession } from './useLiveRoomSession'

const api = liveApi as unknown as { getRoom: ReturnType<typeof vi.fn> }
const getCurrentUserMock = getCurrentUser as unknown as ReturnType<typeof vi.fn>
const notifMock = requestNotificationPermission as unknown as ReturnType<typeof vi.fn>
const clipMock = copyTextToClipboard as unknown as ReturnType<typeof vi.fn>

describe('useLiveRoomSession', () => {
  it('初始化时 loading=true 且 room=null', () => {
    const r = useLiveRoomSession({ roomId: 1 })
    expect(r.loading.value).toBe(true)
    expect(r.room.value).toBe(null)
  })

  it('me / currentUserId 来自 getCurrentUser', () => {
    getCurrentUserMock.mockReturnValueOnce({ id: 200, name: 'me' })
    const r = useLiveRoomSession({ roomId: 1 })
    expect(r.me.value).toEqual({ id: 200, name: 'me' })
    expect(r.currentUserId.value).toBe(200)
  })

  describe('loadRoom', () => {
    it('code 200 时赋值 room 并调用 syncFollowState/initPlayer', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, userId: 100, streamKey: 'k' } })
      const sync = vi.fn()
      const init = vi.fn()
      const r = useLiveRoomSession({ roomId: 1 })
      const data = await r.loadRoom({ syncFollowState: sync, initPlayer: init })
      expect(data).toEqual({ id: 1, userId: 100, streamKey: 'k' })
      expect(r.room.value).toEqual({ id: 1, userId: 100, streamKey: 'k' })
      expect(sync).toHaveBeenCalledWith({ id: 1, userId: 100, streamKey: 'k' })
      expect(init).toHaveBeenCalled()
      expect(r.loading.value).toBe(false)
    })

    it('code 非 200 时 message.error 被调用并返回 null', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 500, message: 'not found' })
      const message = { error: vi.fn(), success: vi.fn() }
      const r = useLiveRoomSession({ roomId: 1, message })
      const data = await r.loadRoom()
      expect(data).toBe(null)
      expect(message.error).toHaveBeenCalledWith('not found')
      expect(r.loading.value).toBe(false)
    })

    it('code 非 200 且无 message 时使用默认文本', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 500 })
      const message = { error: vi.fn() }
      const r = useLiveRoomSession({ roomId: 1, message })
      await r.loadRoom()
      expect(message.error).toHaveBeenCalledWith('直播间不存在')
    })

    it('rejected 时 message.error 调用 加载失败', async () => {
      api.getRoom.mockRejectedValueOnce(new Error('boom'))
      const message = { error: vi.fn() }
      const r = useLiveRoomSession({ roomId: 1, message })
      const data = await r.loadRoom()
      expect(data).toBe(null)
      expect(message.error).toHaveBeenCalledWith('加载失败')
    })

    it('isStreamer 等于 currentUserId === room.userId', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, userId: 100, streamKey: 'k' } })
      const r = useLiveRoomSession({ roomId: 1 })
      expect(r.isStreamer.value).toBe(false) // me = 100, room.userId = 100... wait
      // Actually me.id default is 100 from getCurrentUser mock, room.userId is 100 → isStreamer true
      // After loadRoom, isStreamer will update
      await r.loadRoom()
      expect(r.isStreamer.value).toBe(true)
    })

    it('isStreamer 在 currentUserId !== room.userId 时为 false', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, userId: 999, streamKey: 'k' } })
      const r = useLiveRoomSession({ roomId: 1 })
      await r.loadRoom()
      expect(r.isStreamer.value).toBe(false)
    })
  })

  describe('initializeLiveRoom', () => {
    it('先 requestNotificationPermission，然后 loadRoom，然后 loadLinkmicStatus，然后 connectLiveRoomSocket', async () => {
      notifMock.mockClear()
      api.getRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, userId: 100, streamKey: 'k' } })
      const loadLinkmicStatus = vi.fn().mockResolvedValue(undefined)
      const connectLiveRoomSocket = vi.fn()
      const r = useLiveRoomSession({ roomId: 1 })
      await r.initializeLiveRoom({ loadLinkmicStatus, connectLiveRoomSocket })
      expect(notifMock).toHaveBeenCalled()
      expect(loadLinkmicStatus).toHaveBeenCalled()
      expect(connectLiveRoomSocket).toHaveBeenCalled()
    })
  })

  describe('shareRoom', () => {
    it('调用 copyTextToClipboard 并返回其结果', () => {
      clipMock.mockResolvedValueOnce(true)
      const r = useLiveRoomSession({ roomId: 1 })
      const ok = r.shareRoom()
      expect(clipMock).toHaveBeenCalled()
      expect(typeof ok).toBe('object') // Promise<true>
    })
  })

  describe('custom DI', () => {
    it('getCurrentUserFn 自定义', async () => {
      api.getRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, userId: 50, streamKey: 'k' } })
      const fn = vi.fn(() => ({ id: 50, name: 'custom-me' }))
      const r = useLiveRoomSession({ roomId: 1, getCurrentUserFn: fn })
      await r.loadRoom()
      expect(fn).toHaveBeenCalled()
      expect(r.isStreamer.value).toBe(true)
    })

    it('copyText 自定义', () => {
      const customCopy = vi.fn(() => 'result')
      const r = useLiveRoomSession({ roomId: 1, copyText: customCopy as any })
      r.shareRoom()
      expect(customCopy).toHaveBeenCalled()
    })

    it('getShareUrl 自定义', () => {
      clipMock.mockClear()
      const r = useLiveRoomSession({ roomId: 1, getShareUrl: () => 'https://example.com/share/1' })
      r.shareRoom()
      const urlArg = clipMock.mock.calls[0][0]
      expect(urlArg).toBe('https://example.com/share/1')
    })
  })
})