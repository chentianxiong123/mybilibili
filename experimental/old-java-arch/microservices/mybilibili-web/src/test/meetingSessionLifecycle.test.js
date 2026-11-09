import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useMeetingSessionLifecycle } from '../composables/useMeetingSessionLifecycle.js'

const createLifecycle = (options = {}) => {
  const currentRoom = ref(options.currentRoom ?? { roomCode: 'ROOM01' })
  const hostUserId = ref(options.hostUserId ?? 1)
  const cleanupRecording = vi.fn()
  const cleanupMeetingSessionUi = vi.fn()
  const closeAllPeerConnections = vi.fn()
  const cleanupMedia = vi.fn()
  const cleanupPresenceSignals = vi.fn()
  const cleanupRoomSession = vi.fn()
  const leaveMeetingSession = options.leaveMeetingSession || vi.fn(async () => true)
  const copyText = options.copyText || vi.fn(async () => true)

  const lifecycle = useMeetingSessionLifecycle({
    currentRoom,
    hostUserId,
    cleanupRecording,
    cleanupMeetingSessionUi,
    closeAllPeerConnections,
    cleanupMedia,
    cleanupPresenceSignals,
    cleanupRoomSession,
    leaveMeetingSession,
    copyText,
    getOrigin: () => 'https://example.com'
  })

  return {
    currentRoom,
    hostUserId,
    cleanupRecording,
    cleanupMeetingSessionUi,
    closeAllPeerConnections,
    cleanupMedia,
    cleanupPresenceSignals,
    cleanupRoomSession,
    leaveMeetingSession,
    copyText,
    ...lifecycle
  }
}

describe('useMeetingSessionLifecycle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('cleans up room artifacts and clears the host marker', () => {
    const lifecycle = createLifecycle()

    expect(lifecycle.cleanupRoomArtifacts()).toBe(true)

    expect(lifecycle.cleanupRecording).toHaveBeenCalled()
    expect(lifecycle.cleanupMeetingSessionUi).toHaveBeenCalled()
    expect(lifecycle.closeAllPeerConnections).toHaveBeenCalled()
    expect(lifecycle.cleanupMedia).toHaveBeenCalled()
    expect(lifecycle.cleanupPresenceSignals).toHaveBeenCalled()
    expect(lifecycle.hostUserId.value).toBeNull()
  })

  it('runs full unmount cleanup including the room session', () => {
    const lifecycle = createLifecycle()

    expect(lifecycle.cleanup()).toBe(true)

    expect(lifecycle.cleanupRoomSession).toHaveBeenCalled()
    expect(lifecycle.hostUserId.value).toBeNull()
  })

  it('only cleans local artifacts after a successful leave', async () => {
    const failedLeave = createLifecycle({
      leaveMeetingSession: vi.fn(async () => false)
    })

    expect(await failedLeave.leaveRoom()).toBe(false)
    expect(failedLeave.cleanupRecording).not.toHaveBeenCalled()

    const successfulLeave = createLifecycle()

    expect(await successfulLeave.leaveRoom()).toBe(true)
    expect(successfulLeave.leaveMeetingSession).toHaveBeenCalled()
    expect(successfulLeave.cleanupRecording).toHaveBeenCalled()
    expect(successfulLeave.hostUserId.value).toBeNull()
  })

  it('copies invite links and room codes with meeting-specific messages', async () => {
    const lifecycle = createLifecycle()

    await lifecycle.shareRoom()
    await lifecycle.copyRoomCode('ROOM01')

    expect(lifecycle.copyText).toHaveBeenNthCalledWith(1, 'https://example.com/meeting?code=ROOM01', {
      successMessage: '邀请链接已复制，可直接发送给朋友加入会议'
    })
    expect(lifecycle.copyText).toHaveBeenNthCalledWith(2, 'ROOM01', {
      successMessage: '邀请码已复制'
    })

    lifecycle.currentRoom.value = null
    expect(lifecycle.shareRoom()).toBe(false)
  })
})
