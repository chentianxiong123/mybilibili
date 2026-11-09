import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeetingHostControls } from '../composables/useMeetingHostControls.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

const createControls = (options = {}) => {
  const me = options.me || ref({ id: 1, name: '主持人' })
  const hostUserId = options.hostUserId || ref(1)
  const isHost = options.isHost || computed(() => me.value.id === hostUserId.value)
  const remotePeers = reactive(options.remotePeers || {
    2: { name: 'Alice', handRaised: true }
  })
  const waitingUsers = options.waitingUsers || ref([
    { userId: '3', userName: 'Carol' },
    { userId: 4, userName: 'Dave' }
  ])
  const roomLocked = options.roomLocked || ref(false)
  const spotlightId = options.spotlightId || ref(null)
  const sendMeetingMessage = options.sendMeetingMessage || vi.fn(() => true)
  const removePeerConnection = options.removePeerConnection || vi.fn(peerId => {
    delete remotePeers[peerId]
  })

  const controls = useMeetingHostControls({
    me,
    isHost,
    remotePeers,
    waitingUsers,
    roomLocked,
    spotlightId,
    hostUserId,
    sendMeetingMessage,
    removePeerConnection
  })

  return {
    me,
    isHost,
    remotePeers,
    waitingUsers,
    roomLocked,
    spotlightId,
    hostUserId,
    sendMeetingMessage,
    removePeerConnection,
    ...controls
  }
}

describe('useMeetingHostControls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('updates local host state after successful hand, waiting-room, and moderation actions', () => {
    const meeting = createControls()

    expect(meeting.approveHandRaise(2)).toBe(true)
    expect(meeting.remotePeers[2].handRaised).toBe(false)
    expect(meeting.sendMeetingMessage).toHaveBeenCalledWith({
      type: 'hand-raised',
      targetUserId: 2,
      data: { raised: false }
    })
    expect(ElMessage.success).toHaveBeenCalledWith('已批准发言')

    expect(meeting.admitUser(3)).toBe(true)
    expect(meeting.waitingUsers.value).toEqual([
      { userId: 4, userName: 'Dave' }
    ])

    expect(meeting.rejectUser('4')).toBe(true)
    expect(meeting.waitingUsers.value).toEqual([])
  })

  it('controls spotlight, lock, host transfer, mute, and kick broadcasts', () => {
    const meeting = createControls()

    expect(meeting.spotlightUser(2)).toBe(true)
    expect(meeting.spotlightId.value).toBe(2)
    expect(meeting.sendMeetingMessage).toHaveBeenCalledWith({
      type: 'spotlight',
      data: { targetUserId: 2, active: true }
    })

    expect(meeting.spotlightUser('2')).toBe(true)
    expect(meeting.spotlightId.value).toBeNull()
    expect(meeting.sendMeetingMessage).toHaveBeenCalledWith({
      type: 'spotlight',
      data: { targetUserId: '2', active: false }
    })

    expect(meeting.toggleLockRoom()).toBe(true)
    expect(meeting.roomLocked.value).toBe(true)
    expect(ElMessage.success).toHaveBeenCalledWith('会议已锁定')

    expect(meeting.muteUser(2)).toBe(true)
    expect(meeting.muteAll()).toBe(true)
    expect(meeting.kickUser(2)).toBe(true)
    expect(meeting.removePeerConnection).toHaveBeenCalledWith(2)

    expect(meeting.transferHost(2)).toBe(true)
    expect(meeting.hostUserId.value).toBe(2)
    expect(ElMessage.success).toHaveBeenCalledWith('已将主持人转让')
  })

  it('handles incoming host-control messages and guards host-only sends', () => {
    const me = ref({ id: 2, name: 'Alice' })
    const meeting = createControls({
      me,
      hostUserId: ref(1),
      waitingUsers: ref([]),
      remotePeers: {
        7: { name: 'Bob', handRaised: false }
      }
    })

    expect(meeting.muteAll()).toBe(false)
    expect(meeting.sendMeetingMessage).not.toHaveBeenCalled()

    expect(meeting.handleJoinRequestMessage({ userId: 8, userName: 'Carol' })).toBe(false)
    expect(meeting.waitingUsers.value).toEqual([])

    expect(meeting.handleLockRoomMessage({ data: { locked: true } })).toBe(true)
    expect(meeting.roomLocked.value).toBe(true)
    expect(ElMessage.warning).toHaveBeenCalledWith('会议已被主持人锁定，请等待')

    expect(meeting.handleTransferHostMessage({ targetUserId: 2 })).toBe(true)
    expect(meeting.hostUserId.value).toBe(2)
    expect(ElMessage.success).toHaveBeenCalledWith('你已成为新的主持人')

    expect(meeting.handleHandRaisedMessage({
      userId: 7,
      userName: 'Bob',
      data: { raised: true }
    })).toBe(true)
    expect(meeting.remotePeers[7].handRaised).toBe(true)
  })

  it('deduplicates join requests for the host', () => {
    const meeting = createControls({ waitingUsers: ref([]) })

    expect(meeting.handleJoinRequestMessage({ userId: 8, userName: 'Carol' })).toBe(true)
    expect(meeting.handleJoinRequestMessage({ userId: '8', userName: 'Carol' })).toBe(true)

    expect(meeting.waitingUsers.value).toHaveLength(1)
    expect(meeting.waitingUsers.value[0]).toEqual(expect.objectContaining({
      userId: 8,
      userName: 'Carol'
    }))
    expect(ElMessage.info).toHaveBeenCalledTimes(1)
  })
})
