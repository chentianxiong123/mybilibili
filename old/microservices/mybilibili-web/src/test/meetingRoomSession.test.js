import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeetingRoomSession } from '../composables/useMeetingRoomSession.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
    info: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

const createHarness = (options = {}) => {
  const me = ref(options.me || { id: 1, name: '主持人' })
  const joinCode = ref(options.joinCode || 'ROOM01')
  const currentRoom = ref(options.currentRoom ?? null)
  const hostUserId = ref(options.hostUserId ?? null)
  const loading = ref(false)
  const waitingForHost = ref(false)
  let wsOptions = null
  const ws = {
    send: vi.fn(() => true),
    close: vi.fn(),
    isOpen: vi.fn(() => true)
  }
  const createWs = vi.fn(options => {
    wsOptions = options
    return ws
  })
  const api = options.api || {
    getRoom: vi.fn(async () => ({
      code: 200,
      data: { id: 1, roomCode: 'ROOM01', roomName: '例会', creatorId: 1, status: 0 }
    })),
    joinRoom: vi.fn(async () => ({ code: 200 })),
    leaveRoom: vi.fn(async () => ({ code: 200 })),
    endRoom: vi.fn(async () => ({ code: 200 }))
  }
  const startLocalStream = options.startLocalStream || vi.fn(async () => true)
  const restartMeetingTimer = options.restartMeetingTimer || vi.fn()
  const startQualityPolling = options.startQualityPolling || vi.fn()
  const loadMyRooms = options.loadMyRooms || vi.fn(async () => true)
  const onMessage = options.onMessage || vi.fn()
  const confirm = options.confirm || vi.fn(async () => true)
  const logger = options.logger || { error: vi.fn() }

  const session = useMeetingRoomSession({
    me,
    joinCode,
    currentRoom,
    hostUserId,
    loading,
    waitingForHost,
    startLocalStream,
    restartMeetingTimer,
    startQualityPolling,
    loadMyRooms,
    onMessage,
    api,
    createWs,
    getWsUrl: () => 'ws://meeting.example/ws',
    getJoinBlockMessage: options.getJoinBlockMessage || (() => ''),
    confirm,
    logger
  })

  return {
    me,
    joinCode,
    currentRoom,
    hostUserId,
    loading,
    waitingForHost,
    ws,
    createWs,
    api,
    startLocalStream,
    restartMeetingTimer,
    startQualityPolling,
    loadMyRooms,
    onMessage,
    confirm,
    logger,
    getWsOptions: () => wsOptions,
    ...session
  }
}

describe('useMeetingRoomSession', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('joins as host, opens the room socket, and sends meeting messages', async () => {
    const session = createHarness()

    expect(await session.joinRoomByCode('ROOM01')).toBe(true)

    expect(session.api.getRoom).toHaveBeenCalledWith('ROOM01')
    expect(session.api.joinRoom).toHaveBeenCalledWith('ROOM01')
    expect(session.currentRoom.value).toEqual(expect.objectContaining({ roomCode: 'ROOM01' }))
    expect(session.hostUserId.value).toBe(1)
    expect(session.startLocalStream).toHaveBeenCalled()
    expect(session.restartMeetingTimer).toHaveBeenCalled()
    expect(session.startQualityPolling).toHaveBeenCalled()
    expect(session.loading.value).toBe(false)
    expect(ElMessage.success).toHaveBeenCalledWith('加入成功')

    session.getWsOptions().onOpen()
    expect(session.ws.send).toHaveBeenCalledWith({
      type: 'join',
      roomCode: 'ROOM01',
      userId: 1,
      userName: '主持人'
    })

    expect(session.sendMeetingMessage({
      type: 'chat',
      targetUserId: 2,
      data: { text: 'hi' }
    })).toBe(true)
    expect(session.ws.send).toHaveBeenCalledWith({
      type: 'chat',
      roomCode: 'ROOM01',
      userId: 1,
      userName: '主持人',
      targetUserId: 2,
      data: { text: 'hi' }
    })
  })

  it('sends a join request for non-hosts and cancels waiting state cleanly', async () => {
    const session = createHarness({
      me: { id: 2, name: '访客' }
    })

    expect(await session.joinRoomByCode()).toBe(true)

    expect(session.api.joinRoom).not.toHaveBeenCalled()
    expect(session.waitingForHost.value).toBe(true)
    expect(session.currentRoom.value).toEqual(expect.objectContaining({ roomCode: 'ROOM01' }))
    expect(ElMessage.info).toHaveBeenCalledWith('正在等待主持人批准...')

    session.getWsOptions().onOpen()
    expect(session.ws.send).toHaveBeenCalledWith({
      type: 'join-request',
      roomCode: 'ROOM01',
      userId: 2,
      userName: '访客'
    })

    session.getWsOptions().onMessage({ type: 'admit-user' })
    expect(session.onMessage).toHaveBeenCalledWith({ type: 'admit-user' })

    expect(session.cancelWaitingForHost()).toBe(true)
    expect(session.ws.close).toHaveBeenCalled()
    expect(session.waitingForHost.value).toBe(false)
    expect(session.currentRoom.value).toBeNull()
    expect(session.hostUserId.value).toBeNull()
  })

  it('completes an approved guest join through the room socket', async () => {
    const session = createHarness({
      me: { id: 2, name: '访客' },
      currentRoom: { id: 1, roomCode: 'ROOM01', creatorId: 1 },
      hostUserId: 1
    })
    session.waitingForHost.value = true

    expect(await session.enterApprovedRoom()).toBe(true)

    expect(session.api.joinRoom).toHaveBeenCalledWith('ROOM01')
    expect(session.waitingForHost.value).toBe(false)
    expect(session.startLocalStream).toHaveBeenCalled()
    expect(session.restartMeetingTimer).toHaveBeenCalled()
    expect(session.startQualityPolling).toHaveBeenCalled()
    expect(ElMessage.success).toHaveBeenCalledWith('主持人已允许加入')

    session.getWsOptions().onOpen()
    expect(session.ws.send).toHaveBeenCalledWith({
      type: 'join',
      roomCode: 'ROOM01',
      userId: 2,
      userName: '访客'
    })
  })

  it('blocks invalid or unavailable joins without mutating room state', async () => {
    const session = createHarness({
      getJoinBlockMessage: () => '会议已结束'
    })

    expect(await session.joinRoomByCode('ROOM01')).toBe(false)

    expect(ElMessage.warning).toHaveBeenCalledWith('会议已结束')
    expect(session.currentRoom.value).toBeNull()
    expect(session.api.joinRoom).not.toHaveBeenCalled()

    session.joinCode.value = ''
    expect(await session.joinRoomByCode()).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('请输入会议室邀请码')
  })

  it('logs unexpected join failures and clears loading state', async () => {
    const error = new Error('network down')
    const session = createHarness({
      api: {
        getRoom: vi.fn(async () => { throw error }),
        joinRoom: vi.fn(),
        leaveRoom: vi.fn(),
        endRoom: vi.fn()
      }
    })

    expect(await session.joinRoomByCode('ROOM01')).toBe(false)

    expect(session.logger.error).toHaveBeenCalledWith(error)
    expect(ElMessage.error).toHaveBeenCalledWith('加入失败')
    expect(session.loading.value).toBe(false)
    expect(session.currentRoom.value).toBeNull()
  })

  it('leaves as a participant and ends as host through the session socket', async () => {
    const participant = createHarness({
      me: { id: 2, name: '访客' },
      currentRoom: { id: 1, roomCode: 'ROOM01', creatorId: 1 },
      hostUserId: 1
    })
    participant.connectRoom('ROOM01')

    expect(await participant.leaveRoom()).toBe(true)

    expect(participant.api.leaveRoom).toHaveBeenCalledWith(1)
    expect(participant.ws.send).toHaveBeenCalledWith({
      type: 'leave',
      roomCode: 'ROOM01',
      userId: 2
    })
    expect(participant.ws.close).toHaveBeenCalled()
    expect(participant.currentRoom.value).toBeNull()
    expect(participant.loadMyRooms).toHaveBeenCalled()

    const host = createHarness({
      currentRoom: { id: 7, roomCode: 'HOST01', creatorId: 1 },
      hostUserId: 1
    })
    host.connectRoom('HOST01')

    expect(await host.leaveRoom()).toBe(true)

    expect(host.confirm).toHaveBeenCalled()
    expect(host.api.endRoom).toHaveBeenCalledWith(7)
    expect(host.ws.send).toHaveBeenCalledWith({
      type: 'meeting-ended',
      roomCode: 'HOST01',
      userId: 1,
      data: {}
    })
    expect(host.currentRoom.value).toBeNull()
  })

  it('does not broadcast meeting-ended when the host end request fails', async () => {
    const host = createHarness({
      currentRoom: { id: 7, roomCode: 'HOST01', creatorId: 1 },
      hostUserId: 1,
      api: {
        getRoom: vi.fn(),
        joinRoom: vi.fn(),
        leaveRoom: vi.fn(async () => ({ code: 200 })),
        endRoom: vi.fn(async () => ({ code: 500, message: '无权限结束会议或会议不存在' }))
      }
    })
    host.connectRoom('HOST01')

    expect(await host.leaveRoom()).toBe(true)

    expect(host.api.endRoom).toHaveBeenCalledWith(7)
    expect(ElMessage.error).toHaveBeenCalledWith('无权限结束会议或会议不存在')
    expect(host.ws.send).not.toHaveBeenCalledWith({
      type: 'meeting-ended',
      roomCode: 'HOST01',
      userId: 1,
      data: {}
    })
    expect(host.ws.send).toHaveBeenCalledWith({
      type: 'leave',
      roomCode: 'HOST01',
      userId: 1
    })
    expect(host.currentRoom.value).toBeNull()
  })
})
