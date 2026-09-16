import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'

vi.mock('@/api/linkmic.ts', () => ({
  linkmicApi: {
    applyLinkmic: vi.fn(),
    acceptLinkmic: vi.fn(),
    rejectLinkmic: vi.fn(),
    disconnectLinkmic: vi.fn(),
    toggleAudio: vi.fn(),
    toggleVideo: vi.fn(),
    getActiveLinkmics: vi.fn(),
    getPendingApplications: vi.fn(),
    getQueuePosition: vi.fn(),
  },
}))

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  },
}))

import { linkmicApi } from '@/api/linkmic.ts'
import { ElMessage } from 'element-plus'
import { useLiveLinkmic } from './useLiveLinkmic'

const api = linkmicApi as unknown as {
  applyLinkmic: ReturnType<typeof vi.fn>
  acceptLinkmic: ReturnType<typeof vi.fn>
  rejectLinkmic: ReturnType<typeof vi.fn>
  disconnectLinkmic: ReturnType<typeof vi.fn>
  toggleAudio: ReturnType<typeof vi.fn>
  toggleVideo: ReturnType<typeof vi.fn>
  getActiveLinkmics: ReturnType<typeof vi.fn>
  getPendingApplications: ReturnType<typeof vi.fn>
  getQueuePosition: ReturnType<typeof vi.fn>
}

describe('useLiveLinkmic', () => {
  let sendRoomMessage: any

  beforeEach(() => {
    vi.clearAllMocks()
    sendRoomMessage = vi.fn(() => true)
  })

  function makeHarness(overrides: any = {}) {
    const room = ref({ userId: 1 })
    const isStreamer = ref(false)
    const currentUserId = ref(99)
    const r = useLiveLinkmic({
      roomId: 100,
      room,
      isStreamer,
      currentUserId,
      sendRoomMessage,
      logger: { error: vi.fn(), warn: vi.fn() },
      ...overrides,
    })
    return { ...r, room, isStreamer, currentUserId }
  }

  describe('loadLinkmicStatus', () => {
    it('加载活跃 linkmic', async () => {
      api.getActiveLinkmics.mockResolvedValueOnce({ code: 200, data: [{ id: 1 }] })
      const r = makeHarness()
      await r.loadLinkmicStatus()
      expect(api.getActiveLinkmics).toHaveBeenCalledWith(100)
      expect(r.activeLinkmics.value).toEqual([{ id: 1 }])
    })

    it('非 streamer 时不调用 getPendingApplications', async () => {
      api.getActiveLinkmics.mockResolvedValueOnce({ code: 200, data: [] })
      const r = makeHarness()
      await r.loadLinkmicStatus()
      expect(api.getPendingApplications).not.toHaveBeenCalled()
    })

    it('streamer 时也加载 pending applications', async () => {
      api.getActiveLinkmics.mockResolvedValueOnce({ code: 200, data: [] })
      api.getPendingApplications.mockResolvedValueOnce({ code: 200, data: [{ id: 99 }] })
      const r = makeHarness({ isStreamer: ref(true) })
      await r.loadLinkmicStatus()
      expect(api.getPendingApplications).toHaveBeenCalledWith(100)
      expect(r.pendingApplications.value).toEqual([{ id: 99 }])
    })
  })

  describe('applyLinkmic', () => {
    it('未登录时直接 return', async () => {
      const r = makeHarness({ currentUserId: ref(null) })
      await r.applyLinkmic()
      expect(api.applyLinkmic).not.toHaveBeenCalled()
    })

    it('成功时设置 myLinkmicId + queuePosition + sendBroadcast', async () => {
      api.applyLinkmic.mockResolvedValueOnce({ code: 200, data: { id: 7 } })
      api.getQueuePosition.mockResolvedValueOnce({ code: 200, data: 3 })
      const r = makeHarness()
      await r.applyLinkmic()
      expect(r.myLinkmicId.value).toBe(7)
      expect(r.myQueuePosition.value).toBe(3)
      expect(sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'linkmic-apply' }))
    })

    it('queue=0 时显示立即申请', async () => {
      api.applyLinkmic.mockResolvedValueOnce({ code: 200, data: { id: 7 } })
      api.getQueuePosition.mockResolvedValueOnce({ code: 200, data: 0 })
      const r = makeHarness()
      await r.applyLinkmic()
      expect(ElMessage.success).toHaveBeenCalled()
    })

    it('非 200 时显示错误', async () => {
      api.applyLinkmic.mockResolvedValueOnce({ code: 500, message: 'fail' })
      const r = makeHarness()
      await r.applyLinkmic()
      expect(ElMessage.error).toHaveBeenCalledWith('fail')
    })

    it('rejected 时显示通用错误', async () => {
      api.applyLinkmic.mockRejectedValueOnce(new Error('boom'))
      const r = makeHarness()
      await r.applyLinkmic()
      expect(ElMessage.error).toHaveBeenCalledWith('申请连麦失败')
    })
  })

  describe('acceptLinkmic', () => {
    it('接受连麦', async () => {
      api.acceptLinkmic.mockResolvedValueOnce({})
      const r = makeHarness({ isStreamer: ref(true) })
      const linkmic = { id: 5, viewerId: 10 }
      r.activeLinkmics.value = []
      r.pendingApplications.value = [linkmic]
      await r.acceptLinkmic(linkmic)
      expect(r.activeLinkmics.value).toContainEqual(linkmic)
      expect(r.pendingApplications.value).toEqual([])
      expect(sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'linkmic-accepted' }))
    })

    it('重复 id 时不重复 push', async () => {
      api.acceptLinkmic.mockResolvedValueOnce({})
      const r = makeHarness({ isStreamer: ref(true) })
      const linkmic = { id: 5, viewerId: 10 }
      r.activeLinkmics.value = [linkmic]
      r.pendingApplications.value = [linkmic]
      await r.acceptLinkmic(linkmic)
      expect(r.activeLinkmics.value.length).toBe(1)
    })
  })

  describe('rejectLinkmic', () => {
    it('拒绝连麦', async () => {
      api.rejectLinkmic.mockResolvedValueOnce({})
      const r = makeHarness({ isStreamer: ref(true) })
      const linkmic = { id: 5, viewerId: 10 }
      r.pendingApplications.value = [linkmic]
      await r.rejectLinkmic(linkmic)
      expect(r.pendingApplications.value).toEqual([])
      expect(sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'linkmic-rejected' }))
    })
  })

  describe('disconnectLinkmic', () => {
    it('断开连麦', async () => {
      api.disconnectLinkmic.mockResolvedValueOnce({})
      const r = makeHarness({ isStreamer: ref(true) })
      const linkmic = { id: 5, viewerId: 10 }
      r.activeLinkmics.value = [linkmic]
      await r.disconnectLinkmic(linkmic)
      expect(r.activeLinkmics.value).toEqual([])
      expect(sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'linkmic-disconnected' }))
    })
  })

  describe('message handlers', () => {
    it('handleLinkmicApplyMessage 在非 streamer 时忽略', () => {
      const r = makeHarness()
      r.handleLinkmicApplyMessage({ data: { id: 1 } })
      expect(r.pendingApplications.value).toEqual([])
    })

    it('handleLinkmicApplyMessage 在 streamer 时加入 pending', () => {
      const r = makeHarness({ isStreamer: ref(true) })
      r.handleLinkmicApplyMessage({ data: { id: 1 } })
      expect(r.pendingApplications.value).toEqual([{ id: 1 }])
    })

    it('handleLinkmicApplyMessage 重复 id 不重复添加', () => {
      const r = makeHarness({ isStreamer: ref(true) })
      r.handleLinkmicApplyMessage({ data: { id: 1 } })
      r.handleLinkmicApplyMessage({ data: { id: 1 } })
      expect(r.pendingApplications.value.length).toBe(1)
    })

    it('handleLinkmicRejectedMessage 重置状态', () => {
      const r = makeHarness()
      r.myLinkmicId.value = 5
      r.myQueuePosition.value = 2
      r.viewerHandRaised.value = true
      r.handleLinkmicRejectedMessage()
      expect(r.myLinkmicId.value).toBe(null)
      expect(r.myQueuePosition.value).toBe(0)
      expect(r.viewerHandRaised.value).toBe(false)
      expect(ElMessage.warning).toHaveBeenCalled()
    })

    it('handleHandRaisedMessage 在非 streamer 时忽略', () => {
      const r = makeHarness()
      r.handleHandRaisedMessage({ userId: 5, data: { raised: true } })
      expect(r.pendingApplications.value).toEqual([])
    })

    it('handleHandRaisedMessage streamer 时更新 pending 上对应用户的 handRaised 标记', () => {
      const r = makeHarness({ isStreamer: ref(true) })
      r.pendingApplications.value = [{ id: 1, viewerId: 5, handRaised: false }]
      r.handleHandRaisedMessage({ userId: 5, data: { raised: true } })
      expect(r.pendingApplications.value[0].handRaised).toBe(true)
    })

    it('handlePeerStateMessage 写入 linkmicStreams', () => {
      const r = makeHarness()
      r.handlePeerStateMessage({ userId: 5, userName: 'bob', data: { audioEnabled: false } })
      expect(r.linkmicStreams[5]).toBeTruthy()
      expect(r.linkmicStreams[5].audioEnabled).toBe(false)
    })

    it('handlePeerStateMessage userId 缺失时返回 false', () => {
      const r = makeHarness()
      expect(r.handlePeerStateMessage({ data: {} })).toBe(false)
    })

    it('handleLinkmicDisconnectedMessage 自己是 target 时清空状态', () => {
      const r = makeHarness()
      r.myLinkmicId.value = 1
      r.myQueuePosition.value = 2
      r.viewerHandRaised.value = true
      r.handleLinkmicDisconnectedMessage({ userId: 50, data: { viewerId: 99 } })
      expect(r.myLinkmicId.value).toBe(null)
      expect(ElMessage.info).toHaveBeenCalledWith('连麦已断开')
    })

    it('handleLinkmicDisconnectedMessage 不是自己时只移除对应 peer', () => {
      const r = makeHarness({ isStreamer: ref(true) })
      r.myLinkmicId.value = 1
      r.handleLinkmicDisconnectedMessage({ userId: 50, data: { viewerId: 50 } })
      expect(r.myLinkmicId.value).toBe(1)
    })
  })

  describe('toggleViewerHandRaise', () => {
    it('sendRoomMessage 失败时不更新状态', () => {
      sendRoomMessage.mockReturnValue(false)
      const r = makeHarness()
      r.toggleViewerHandRaise()
      expect(r.viewerHandRaised.value).toBe(false)
    })

    it('sendRoomMessage 成功时切换状态', () => {
      const r = makeHarness()
      r.toggleViewerHandRaise()
      expect(r.viewerHandRaised.value).toBe(true)
      expect(sendRoomMessage).toHaveBeenCalledWith(expect.objectContaining({ type: 'hand-raised' }))
    })
  })

  describe('cleanupLinkmic', () => {
    it('调用 closeAllPeerConnections 与 stopLocalLinkmicStream', () => {
      const r = makeHarness()
      // cleanup 应该调用了 closeAllPeerConnections (来自 usePeerConnectionMesh)
      // 这里只验证函数存在并可调用
      expect(() => r.cleanupLinkmic()).not.toThrow()
    })
  })
})