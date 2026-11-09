import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref, toRaw } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeetingMediaControls } from '../composables/useMeetingMediaControls.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn()
  }
}))

const createTrack = (kind) => ({
  kind,
  enabled: true,
  stop: vi.fn(),
  onended: null,
  readyState: 'live'
})

const createStream = ({ audio = true, video = true } = {}) => {
  const tracks = []
  if (audio) tracks.push(createTrack('audio'))
  if (video) tracks.push(createTrack('video'))
  return {
    tracks,
    getTracks: () => tracks,
    getAudioTracks: () => tracks.filter(track => track.kind === 'audio'),
    getVideoTracks: () => tracks.filter(track => track.kind === 'video'),
    addTrack: vi.fn(track => tracks.push(track))
  }
}

const createControls = (options = {}) => {
  const localStream = ref(null)
  const localVideoRef = ref({ srcObject: null })
  const audioEnabled = ref(true)
  const videoEnabled = ref(true)
  const screenShareEnabled = ref(false)
  const virtualBgEnabled = ref(false)
  const virtualBgStream = ref(null)
  const broadcastSelfState = options.broadcastSelfState || vi.fn()
  const updateAudioActivity = options.updateAudioActivity || vi.fn()
  const replaceVideoTrack = options.replaceVideoTrack || vi.fn()
  const logger = options.logger || { error: vi.fn() }

  const controls = useMeetingMediaControls({
    localStream,
    localVideoRef,
    audioEnabled,
    videoEnabled,
    screenShareEnabled,
    virtualBgEnabled,
    virtualBgStream,
    replaceVideoTrack,
    broadcastSelfState,
    updateAudioActivity,
    captureUserMedia: options.captureUserMedia || vi.fn(async () => createStream()),
    captureDisplayMedia: options.captureDisplayMedia || vi.fn(async () => createStream({ audio: false, video: true })),
    createElement: options.createElement || vi.fn(() => ({
      width: 0,
      height: 0,
      getContext: vi.fn(() => ({
        filter: '',
        drawImage: vi.fn(),
        clearRect: vi.fn()
      })),
      captureStream: vi.fn(() => createStream({ audio: false, video: true }))
    })),
    createMediaStream: options.createMediaStream || vi.fn(() => createStream({ audio: false, video: false })),
    requestFrame: options.requestFrame || vi.fn(cb => 1),
    cancelFrame: options.cancelFrame || vi.fn(),
    logger
  })

  return {
    localStream,
    localVideoRef,
    audioEnabled,
    videoEnabled,
    screenShareEnabled,
    virtualBgEnabled,
    virtualBgStream,
    broadcastSelfState,
    updateAudioActivity,
    replaceVideoTrack,
    logger,
    ...controls
  }
}

describe('useMeetingMediaControls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('initializes local media and toggles audio/video state', async () => {
    const localStream = createStream()
    const captureUserMedia = vi.fn(async () => localStream)
    const media = createControls({ captureUserMedia })

    await media.startLocalStream()

    expect(toRaw(media.localStream.value)).toBe(localStream)
    expect(toRaw(media.localVideoRef.value.srcObject)).toBe(localStream)
    expect(media.audioEnabled.value).toBe(true)
    expect(media.videoEnabled.value).toBe(true)

    expect(media.toggleAudio()).toBe(true)
    expect(localStream.getAudioTracks()[0].enabled).toBe(false)
    expect(media.audioEnabled.value).toBe(false)
    expect(media.updateAudioActivity).toHaveBeenCalledWith('local', false)
    expect(media.broadcastSelfState).toHaveBeenCalled()

    expect(media.toggleVideo()).toBe(true)
    expect(localStream.getVideoTracks()[0].enabled).toBe(false)
    expect(media.videoEnabled.value).toBe(false)
  })

  it('logs local media capture failures without leaving active state', async () => {
    const error = new Error('media failed')
    const media = createControls({
      captureUserMedia: vi.fn(async () => { throw error })
    })

    await expect(media.startLocalStream()).resolves.toBe(false)

    expect(media.logger.error).toHaveBeenCalledWith(error)
    expect(ElMessage.error).toHaveBeenCalledWith('无法获取摄像头/麦克风')
    expect(media.localStream.value).toBeNull()
  })

  it('starts and stops screen share while restoring the camera track', async () => {
    const localStream = createStream()
    const screenStream = createStream({ audio: false, video: true })
    const captureDisplayMedia = vi.fn(async () => screenStream)
    const media = createControls({
      captureDisplayMedia,
      captureUserMedia: vi.fn(async () => localStream)
    })

    await media.startLocalStream()
    await media.startScreenShare()

    expect(captureDisplayMedia).toHaveBeenCalledWith({ video: true, audio: false })
    expect(media.screenShareEnabled.value).toBe(true)
    expect(toRaw(media.localVideoRef.value.srcObject)).toBe(screenStream)
    expect(media.replaceVideoTrack).toHaveBeenCalledWith(screenStream.getVideoTracks()[0])

    expect(media.stopScreenShare()).toBe(true)
    expect(media.screenShareEnabled.value).toBe(false)
    expect(toRaw(media.localVideoRef.value.srcObject)).toBe(localStream)
    expect(media.replaceVideoTrack).toHaveBeenLastCalledWith(localStream.getVideoTracks()[0])
  })

  it('rolls back virtual background state when frame processing setup fails', async () => {
    const localStream = createStream()
    const error = new Error('raf failed')
    const requestFrame = vi.fn(() => { throw error })
    const media = createControls({
      captureUserMedia: vi.fn(async () => localStream),
      requestFrame
    })

    await media.startLocalStream()

    await expect(media.applyVirtualBg()).resolves.toBe(false)

    expect(media.logger.error).toHaveBeenCalledWith(error)
    expect(media.virtualBgEnabled.value).toBe(false)
    expect(media.virtualBgStream.value).toBeNull()
    expect(toRaw(media.localVideoRef.value.srcObject)).toBe(localStream)
    expect(media.replaceVideoTrack).toHaveBeenLastCalledWith(localStream.getVideoTracks()[0])
    expect(ElMessage.error).toHaveBeenCalledWith('虚拟背景开启失败')
  })

  it('cleans up active media tracks and preview state', async () => {
    const localStream = createStream()
    const screenStream = createStream({ audio: false, video: true })
    const captureUserMedia = vi.fn(async () => localStream)
    const captureDisplayMedia = vi.fn(async () => screenStream)
    const media = createControls({
      captureUserMedia,
      captureDisplayMedia
    })

    await media.startLocalStream()
    await media.startScreenShare()

    media.cleanupMedia()

    expect(localStream.getTracks().every(track => track.stop.mock.calls.length > 0)).toBe(true)
    expect(screenStream.getTracks().every(track => track.stop.mock.calls.length > 0)).toBe(true)
    expect(media.localStream.value).toBeNull()
    expect(media.localVideoRef.value.srcObject).toBeNull()
    expect(media.screenShareEnabled.value).toBe(false)
    expect(media.virtualBgEnabled.value).toBe(false)
  })
})
