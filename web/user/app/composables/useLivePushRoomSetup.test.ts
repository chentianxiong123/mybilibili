import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/live.ts', () => ({
  liveApi: {
    getMyRoom: vi.fn(),
    createRoom: vi.fn(),
    updateRoomStatus: vi.fn(),
    uploadCover: vi.fn(),
  },
}))

vi.mock('@/utils/currentUser.ts', () => ({
  getCurrentUser: vi.fn(() => ({ id: 100, name: 'host' })),
}))

vi.mock('@/utils/clipboard.ts', () => ({
  copyTextToClipboard: vi.fn(),
}))

import { liveApi } from '@/api/live.ts'
import { useLivePushRoomSetup } from './useLivePushRoomSetup'
import { LIVE_ROOM_STATUS } from '@/utils/liveMeetingStatus.ts'

const api = liveApi as unknown as {
  getMyRoom: ReturnType<typeof vi.fn>
  createRoom: vi.fn
  updateRoomStatus: ReturnType<typeof vi.fn>
  uploadCover: ReturnType<typeof vi.fn>
}

describe('useLivePushRoomSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('初始', () => {
    it('loading=true，room=null，rtmpUrl/streamKey 为空', () => {
      const r = useLivePushRoomSetup()
      expect(r.loading.value).toBe(true)
      expect(r.room.value).toBe(null)
      expect(r.rtmpUrl.value).toBe('')
      expect(r.streamKey.value).toBe('')
    })

    it('me 来自 getCurrentUser', () => {
      const r = useLivePushRoomSetup()
      expect(r.me).toEqual({ id: 100, name: 'host' })
    })

    it('isLive 在 status 为 live 时为 true', () => {
      const r = useLivePushRoomSetup()
      r.room.value = { status: LIVE_ROOM_STATUS.LIVE }
      expect(r.isLive.value).toBe(true)
      r.room.value = { status: LIVE_ROOM_STATUS.OFFLINE }
      expect(r.isLive.value).toBe(false)
    })
  })

  describe('loadRoom', () => {
    it('getMyRoom 返回数据时 syncRoomState', async () => {
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: { id: 1, streamKey: 'k', status: 'offline' } })
      const r = useLivePushRoomSetup()
      const result = await r.loadRoom()
      expect(result).toEqual({ id: 1, streamKey: 'k', status: 'offline' })
      expect(r.streamKey.value).toBe('k')
      expect(r.rtmpUrl.value).toMatch(/^rtmp:\/\/.+\/live$/)
      expect(api.createRoom).not.toHaveBeenCalled()
      expect(r.loading.value).toBe(false)
    })

    it('getMyRoom 返回空数据时尝试 createRoom', async () => {
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      api.createRoom.mockResolvedValueOnce({ code: 200, data: { id: 2, streamKey: 'k2' } })
      const r = useLivePushRoomSetup()
      await r.loadRoom()
      expect(api.createRoom).toHaveBeenCalledWith('我的直播间')
    })

    it('createRoom 失败时 syncRoomState(null) 并返回 null', async () => {
      api.getMyRoom.mockResolvedValueOnce({ code: 200, data: null })
      api.createRoom.mockResolvedValueOnce({ code: 500, data: null })
      const r = useLivePushRoomSetup()
      const result = await r.loadRoom()
      expect(result).toBe(null)
      expect(r.room.value).toBe(null)
    })

    it('rejected 时 logger.error 被调用', async () => {
      api.getMyRoom.mockRejectedValueOnce(new Error('boom'))
      const message = { error: vi.fn(), success: vi.fn() }
      const logger = { error: vi.fn() }
      const r = useLivePushRoomSetup({ message, logger })
      const result = await r.loadRoom()
      expect(result).toBe(null)
      expect(logger.error).toHaveBeenCalled()
      expect(message.error).toHaveBeenCalledWith('获取直播间信息失败')
    })
  })

  describe('handleWebRtcPublished', () => {
    it('room 为 null 时返回 false', async () => {
      const r = useLivePushRoomSetup()
      expect(await r.handleWebRtcPublished()).toBe(false)
    })

    it('非 live 状态时调用 updateRoomStatus(LIVE)', async () => {
      api.updateRoomStatus.mockResolvedValueOnce({ code: 200 })
      const r = useLivePushRoomSetup()
      r.room.value = { id: 1, status: LIVE_ROOM_STATUS.OFFLINE }
      const ok = await r.handleWebRtcPublished()
      expect(ok).toBe(true)
      expect(api.updateRoomStatus).toHaveBeenCalledWith(1, LIVE_ROOM_STATUS.LIVE)
      expect(r.room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
    })

    it('已是 live 时不再调 API，直接设为 live', async () => {
      const r = useLivePushRoomSetup()
      r.room.value = { id: 1, status: LIVE_ROOM_STATUS.LIVE }
      const ok = await r.handleWebRtcPublished()
      expect(ok).toBe(true)
      expect(api.updateRoomStatus).not.toHaveBeenCalled()
    })

    it('updateRoomStatus 抛错时也设为 live', async () => {
      api.updateRoomStatus.mockRejectedValueOnce(new Error('boom'))
      const logger = { error: vi.fn() }
      const r = useLivePushRoomSetup({ logger })
      r.room.value = { id: 1, status: LIVE_ROOM_STATUS.OFFLINE }
      await r.handleWebRtcPublished()
      expect(r.room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
      expect(logger.error).toHaveBeenCalled()
    })

    it('updateRoomStatus 返回非 200 时仍设 live（兜底）', async () => {
      api.updateRoomStatus.mockResolvedValueOnce({ code: 500 })
      const r = useLivePushRoomSetup()
      r.room.value = { id: 1, status: LIVE_ROOM_STATUS.OFFLINE }
      await r.handleWebRtcPublished()
      expect(r.room.value.status).toBe(LIVE_ROOM_STATUS.LIVE)
    })
  })

  describe('formatJoinTime', () => {
    it('< 60 秒 → 秒前', () => {
      const r = useLivePushRoomSetup({ now: () => 100000 })
      expect(r.formatJoinTime(99000)).toBe('1 秒前')
      expect(r.formatJoinTime(95000)).toBe('5 秒前')
    })
    it('60-3600 秒 → 分钟前', () => {
      const r = useLivePushRoomSetup({ now: () => 100000 })
      expect(r.formatJoinTime(40000)).toBe('1 分钟前') // 60s
      expect(r.formatJoinTime(0)).toBe('1 分钟前') // 100s
      expect(r.formatJoinTime(-260000)).toBe('6 分钟前') // 360s
    })
    it('>= 3600 秒 → 小时前', () => {
      const r = useLivePushRoomSetup({ now: () => 100000 })
      // 需要 diff >= 3600000ms
      expect(r.formatJoinTime(-3600000)).toBe('1 小时前') // 3700s
      expect(r.formatJoinTime(-7200000)).toBe('2 小时前')
    })
  })

  describe('copyField', () => {
    it('调用 copyTextToClipboard', () => {
      const copyText = vi.fn(() => true)
      const message = { success: vi.fn() }
      const r = useLivePushRoomSetup({ copyText, message })
      const ok = r.copyField('hello')
      expect(copyText).toHaveBeenCalledWith('hello', expect.objectContaining({ message }))
    })
  })

  describe('beforeCoverUpload', () => {
    it('非图片返回 false', () => {
      const message = { error: vi.fn() }
      const r = useLivePushRoomSetup({ message })
      expect(r.beforeCoverUpload({ type: 'application/pdf', size: 100 } as any)).toBe(false)
      expect(message.error).toHaveBeenCalledWith('只能上传图片文件')
    })
    it('> 2MB 返回 false', () => {
      const message = { error: vi.fn() }
      const r = useLivePushRoomSetup({ message })
      expect(r.beforeCoverUpload({ type: 'image/png', size: 3 * 1024 * 1024 } as any)).toBe(false)
      expect(message.error).toHaveBeenCalledWith('图片大小不能超过 2MB')
    })
    it('合法文件返回 true', () => {
      const r = useLivePushRoomSetup()
      expect(r.beforeCoverUpload({ type: 'image/jpeg', size: 1024 * 1024 } as any)).toBe(true)
    })
  })

  describe('uploadCover', () => {
    it('room 为 null 返回 false', async () => {
      const r = useLivePushRoomSetup()
      expect(await r.uploadCover({ file: {} as any })).toBe(false)
    })
    it('成功时更新 coverUrl 并 message.success', async () => {
      api.uploadCover.mockResolvedValueOnce({ code: 200, data: '/u/c.jpg' })
      const message = { success: vi.fn(), error: vi.fn() }
      const r = useLivePushRoomSetup({ message })
      r.room.value = { id: 1 }
      const ok = await r.uploadCover({ file: {} as any })
      expect(ok).toBe(true)
      expect(r.room.value.coverUrl).toBe('/u/c.jpg')
      expect(message.success).toHaveBeenCalledWith('封面上传成功')
    })
    it('非 200 返回 false 并 message.error', async () => {
      api.uploadCover.mockResolvedValueOnce({ code: 500, message: 'fail' })
      const message = { success: vi.fn(), error: vi.fn() }
      const r = useLivePushRoomSetup({ message })
      r.room.value = { id: 1 }
      const ok = await r.uploadCover({ file: {} as any })
      expect(ok).toBe(false)
      expect(message.error).toHaveBeenCalledWith('fail')
    })
    it('rejected 时 logger.error', async () => {
      api.uploadCover.mockRejectedValueOnce(new Error('boom'))
      const message = { success: vi.fn(), error: vi.fn() }
      const logger = { error: vi.fn() }
      const r = useLivePushRoomSetup({ message, logger })
      r.room.value = { id: 1 }
      const ok = await r.uploadCover({ file: {} as any })
      expect(ok).toBe(false)
      expect(logger.error).toHaveBeenCalled()
      expect(message.error).toHaveBeenCalledWith('封面上传失败')
    })
  })

  describe('syncRoomState', () => {
    it('null 时清空所有字段', () => {
      const r = useLivePushRoomSetup()
      r.syncRoomState({ id: 1, streamKey: 'k' })
      expect(r.streamKey.value).toBe('k')
      r.syncRoomState(null)
      expect(r.room.value).toBe(null)
      expect(r.streamKey.value).toBe('')
      expect(r.rtmpUrl.value).toBe('')
    })
  })
})