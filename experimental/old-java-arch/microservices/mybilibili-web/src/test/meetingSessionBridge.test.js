import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useMeetingSessionBridge } from '../composables/useMeetingSessionBridge.js'

describe('useMeetingSessionBridge', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('delegates meeting session actions through attachable references', async () => {
    const bridge = useMeetingSessionBridge()
    const session = {
      sendMeetingMessage: vi.fn(() => true),
      sendSignaling: vi.fn(() => true),
      joinRoomByCode: vi.fn(() => true),
      cancelWaitingForHost: vi.fn(() => true),
      leaveRoom: vi.fn(async () => true),
      cleanupRoomSession: vi.fn(() => true),
      enterApprovedRoom: vi.fn(() => true)
    }
    const router = {
      handleSignaling: vi.fn(async () => true)
    }

    expect(bridge.sendMeetingMessage({ type: 'chat' })).toBe(false)
    expect(await bridge.handleSignaling({ type: 'peer-state' })).toBe(false)

    bridge.attachMeetingSession(session)
    bridge.attachSignalingRouter(router)

    expect(bridge.sendMeetingMessage({ type: 'chat' })).toBe(true)
    expect(bridge.sendSignaling('chat', 2, { text: 'hi' })).toBe(true)
    expect(bridge.joinRoomByCode('ROOM01')).toBe(true)
    expect(bridge.cancelWaitingForHost()).toBe(true)
    expect(await bridge.leaveRoom()).toBe(true)
    expect(bridge.cleanupRoomSession()).toBe(true)
    expect(bridge.enterApprovedRoom()).toBe(true)
    expect(await bridge.handleSignaling({ type: 'peer-state' })).toBe(true)

    expect(session.sendMeetingMessage).toHaveBeenCalledWith({ type: 'chat' })
    expect(session.sendSignaling).toHaveBeenCalledWith('chat', 2, { text: 'hi' })
    expect(session.joinRoomByCode).toHaveBeenCalledWith('ROOM01')
    expect(session.cancelWaitingForHost).toHaveBeenCalled()
    expect(session.leaveRoom).toHaveBeenCalled()
    expect(session.cleanupRoomSession).toHaveBeenCalled()
    expect(session.enterApprovedRoom).toHaveBeenCalled()
    expect(router.handleSignaling).toHaveBeenCalledWith({ type: 'peer-state' })
  })
})
