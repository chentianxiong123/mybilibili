import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive, ref, toRaw } from 'vue'
import { ElMessage } from 'element-plus'
import { formatMeetingDuration, useMeetingSessionUi } from '../composables/useMeetingSessionUi.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

const createUi = (options = {}) => {
  const me = options.me || ref({ id: 1, name: '主持人' })
  const currentRoom = options.currentRoom || ref({ roomCode: 'ROOM01' })
  const hostUserId = options.hostUserId || ref(1)
  const remoteStream = { id: 'remote-stream' }
  const remotePeers = reactive(options.remotePeers || {
    2: {
      name: 'Alice',
      stream: remoteStream,
      audioEnabled: false,
      videoEnabled: true,
      screenShareEnabled: true,
      handRaised: true
    }
  })
  const localVideo = {
    srcObject: null,
    requestPictureInPicture: vi.fn(async () => true)
  }
  const remoteVideo = { srcObject: null }
  const localVideoRef = ref(localVideo)
  const audioEnabled = ref(true)
  const videoEnabled = ref(true)
  const screenShareEnabled = ref(false)
  const sendMeetingMessage = options.sendMeetingMessage || vi.fn(() => true)
  const broadcastSelfState = options.broadcastSelfState || vi.fn()
  const toggleAudio = options.toggleAudio || vi.fn(() => true)
  const toggleVideo = options.toggleVideo || vi.fn(() => true)
  const startScreenShare = options.startScreenShare || vi.fn(() => true)
  const logger = options.logger || { warn: vi.fn() }

  let keyboardHandler = null
  let hideControlsHandler = null
  let meetingTimerHandler = null
  const addEventListenerFn = vi.fn((type, handler) => { keyboardHandler = handler })
  const removeEventListenerFn = vi.fn()
  const setTimeoutFn = vi.fn(handler => {
    hideControlsHandler = handler
    return 'hide-timer'
  })
  const clearTimeoutFn = vi.fn()
  const setIntervalFn = vi.fn(handler => {
    meetingTimerHandler = handler
    return 'meeting-timer'
  })
  const clearIntervalFn = vi.fn()
  const documentRef = {
    fullscreenElement: null,
    exitFullscreen: vi.fn(),
    pictureInPictureElement: null,
    pictureInPictureEnabled: true,
    exitPictureInPicture: vi.fn(async () => true),
    querySelectorAll: vi.fn(() => [localVideo, remoteVideo])
  }

  const ui = useMeetingSessionUi({
    me,
    currentRoom,
    hostUserId,
    remotePeers,
    localVideoRef,
    audioEnabled,
    videoEnabled,
    screenShareEnabled,
    sendMeetingMessage,
    broadcastSelfState,
    toggleAudio,
    toggleVideo,
    startScreenShare,
    documentRef,
    addEventListenerFn,
    removeEventListenerFn,
    setTimeoutFn,
    clearTimeoutFn,
    setIntervalFn,
    clearIntervalFn,
    nextTickFn: fn => fn(),
    logger
  })

  return {
    me,
    currentRoom,
    hostUserId,
    remotePeers,
    remoteStream,
    localVideo,
    remoteVideo,
    localVideoRef,
    audioEnabled,
    videoEnabled,
    screenShareEnabled,
    sendMeetingMessage,
    broadcastSelfState,
    toggleAudio,
    toggleVideo,
    startScreenShare,
    logger,
    addEventListenerFn,
    removeEventListenerFn,
    setTimeoutFn,
    clearTimeoutFn,
    setIntervalFn,
    clearIntervalFn,
    documentRef,
    getKeyboardHandler: () => keyboardHandler,
    runHideControlsTimer: () => hideControlsHandler?.(),
    runMeetingTimer: () => meetingTimerHandler?.(),
    ...ui
  }
}

describe('useMeetingSessionUi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('formats duration and derives participant state', () => {
    const ui = createUi({ hostUserId: ref('1') })

    expect(formatMeetingDuration(3661)).toBe('01:01:01')
    expect(ui.remotePeerIds.value).toEqual(['2'])
    expect(ui.isHost.value).toBe(true)
    expect(ui.participantList.value).toEqual([
      expect.objectContaining({
        userId: 1,
        name: '主持人（我）',
        audioEnabled: true,
        videoEnabled: true,
        self: true,
        isHost: true
      }),
      expect.objectContaining({
        userId: 2,
        name: 'Alice',
        audioEnabled: false,
        videoEnabled: true,
        screenShareEnabled: true,
        handRaised: true,
        self: false
      })
    ])
    expect(ui.totalHandRaised.value).toBe(1)
    expect(ui.screenSharerId.value).toBe('2')
  })

  it('owns control hide and meeting duration timers', () => {
    const ui = createUi()

    ui.showControls()
    ui.showControls()
    expect(ui.controlsVisible.value).toBe(true)
    expect(ui.setTimeoutFn).toHaveBeenCalledWith(expect.any(Function), 3000)
    expect(ui.clearTimeoutFn).toHaveBeenCalledWith('hide-timer')

    ui.runHideControlsTimer()
    expect(ui.controlsVisible.value).toBe(false)

    ui.restartMeetingTimer()
    ui.runMeetingTimer()
    ui.runMeetingTimer()
    expect(ui.meetingDuration.value).toBe(2)

    ui.restartMeetingTimer()
    expect(ui.clearIntervalFn).toHaveBeenCalledWith('meeting-timer')
    expect(ui.meetingDuration.value).toBe(0)

    ui.showControls()
    ui.handRaised.value = true
    ui.waitingUsers.value = [{ userId: 3 }]
    ui.waitingForHost.value = true
    ui.spotlightId.value = 2
    ui.cleanupMeetingSessionUi()

    expect(ui.clearIntervalFn).toHaveBeenCalledWith('meeting-timer')
    expect(ui.clearTimeoutFn).toHaveBeenCalledWith('hide-timer')
    expect(ui.controlsVisible.value).toBe(true)
    expect(ui.handRaised.value).toBe(false)
    expect(ui.waitingUsers.value).toEqual([])
    expect(ui.waitingForHost.value).toBe(false)
    expect(ui.spotlightId.value).toBeNull()
  })

  it('registers keyboard shortcuts and ignores text inputs', () => {
    const ui = createUi()

    expect(ui.registerKeyboardShortcuts()).toBe(true)
    expect(ui.registerKeyboardShortcuts()).toBe(false)
    expect(ui.addEventListenerFn).toHaveBeenCalledWith('keydown', expect.any(Function))

    ui.getKeyboardHandler()({ key: 'm', target: { tagName: 'INPUT' } })
    expect(ui.toggleAudio).not.toHaveBeenCalled()

    ui.getKeyboardHandler()({ key: 'm', target: { tagName: 'DIV' } })
    ui.getKeyboardHandler()({ key: 'v', target: { tagName: 'DIV' } })
    ui.getKeyboardHandler()({ key: 's', target: { tagName: 'DIV' } })
    ui.getKeyboardHandler()({ key: 'h', target: { tagName: 'DIV' } })

    expect(ui.toggleAudio).toHaveBeenCalled()
    expect(ui.toggleVideo).toHaveBeenCalled()
    expect(ui.startScreenShare).toHaveBeenCalled()
    expect(ui.sendMeetingMessage).toHaveBeenCalledWith({
      type: 'hand-raised',
      data: { raised: true }
    })
    expect(ui.handRaised.value).toBe(true)

    expect(ui.unregisterKeyboardShortcuts()).toBe(true)
    expect(ui.removeEventListenerFn).toHaveBeenCalledWith('keydown', expect.any(Function))
  })

  it('edits the local display name and toggles hand raise state', () => {
    const ui = createUi()
    const select = vi.fn()
    ui.editNameInput.value = { select }

    ui.startEditName()
    expect(ui.editingName.value).toBe(true)
    expect(select).toHaveBeenCalled()

    ui.editNameDraft.value = '  新主持  '
    expect(ui.saveName()).toBe(true)
    expect(ui.me.value.name).toBe('新主持')
    expect(ui.broadcastSelfState).toHaveBeenCalled()
    expect(ElMessage.success).toHaveBeenCalledWith('昵称已更新')

    expect(ui.toggleHandRaise()).toBe(true)
    expect(ui.handRaised.value).toBe(true)
    expect(ElMessage.info).toHaveBeenCalledWith('已举手，请等待主持人允许')

    expect(ui.toggleHandRaise()).toBe(true)
    expect(ui.handRaised.value).toBe(false)
    expect(ElMessage.info).toHaveBeenCalledWith('已放下手')
  })

  it('binds remote video streams and controls picture-in-picture', async () => {
    const ui = createUi()

    expect(ui.setRemoteVideo('2', ui.remoteVideo)).toBe(true)
    expect(toRaw(ui.remoteVideo.srcObject)).toBe(ui.remoteStream)

    expect(await ui.togglePiP('local')).toBe(true)
    expect(ui.localVideo.requestPictureInPicture).toHaveBeenCalled()

    ui.documentRef.pictureInPictureElement = ui.localVideo
    expect(await ui.togglePiP('local')).toBe(true)
    expect(ui.documentRef.exitPictureInPicture).toHaveBeenCalled()

    ui.documentRef.pictureInPictureElement = null
    ui.documentRef.pictureInPictureEnabled = false
    expect(await ui.togglePiP('local')).toBe(false)
    expect(ElMessage.warning).toHaveBeenCalledWith('当前浏览器不支持画中画')
  })

  it('logs picture-in-picture failures without throwing', async () => {
    const error = new Error('pip failed')
    const localVideo = {
      srcObject: null,
      requestPictureInPicture: vi.fn(async () => { throw error })
    }
    const ui = createUi()
    ui.localVideoRef.value = localVideo

    await expect(ui.togglePiP('local')).resolves.toBe(false)

    expect(ui.logger.warn).toHaveBeenCalledWith('PiP error:', error)
  })
})
