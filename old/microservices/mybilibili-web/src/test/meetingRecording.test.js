import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { formatRecordingTime, useMeetingRecording } from '../composables/useMeetingRecording.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn()
  }
}))

const createTrack = (id) => ({ id })

const createStream = (...tracks) => ({
  tracks,
  getTracks: () => tracks
})

const createRecorderHarness = (options = {}) => {
  const localStream = ref(options.localStream || null)
  const remotePeers = options.remotePeers || {}
  const combinedStream = {
    tracks: [],
    addTrack: vi.fn(track => combinedStream.tracks.push(track))
  }
  const recorder = {
    state: 'inactive',
    start: vi.fn(timeslice => {
      recorder.state = 'recording'
      recorder.timeslice = timeslice
    }),
    stop: vi.fn(() => {
      recorder.state = 'inactive'
    }),
    ondataavailable: null
  }
  const anchor = {
    href: '',
    download: '',
    click: vi.fn()
  }
  let timerHandler = null
  const createBlob = vi.fn((chunks, blobOptions) => ({ chunks, blobOptions }))
  const createObjectUrl = vi.fn(() => 'blob:meeting-recording')
  const revokeObjectUrl = vi.fn()
  const setTimer = vi.fn(handler => {
    timerHandler = handler
    return 'recording-timer'
  })
  const clearTimer = vi.fn()
  const logger = options.logger || { error: vi.fn() }

  const recording = useMeetingRecording({
    localStream,
    remotePeers,
    createMediaStream: options.createMediaStream || vi.fn(() => combinedStream),
    createMediaRecorder: options.createMediaRecorder || vi.fn(() => recorder),
    createBlob,
    createObjectUrl,
    revokeObjectUrl,
    createAnchor: vi.fn(() => anchor),
    nowLabel: () => '2026-05-31 05:10:00',
    setTimer,
    clearTimer,
    logger
  })

  return {
    localStream,
    remotePeers,
    combinedStream,
    recorder,
    anchor,
    createBlob,
    createObjectUrl,
    revokeObjectUrl,
    setTimer,
    clearTimer,
    logger,
    tick: () => timerHandler?.(),
    ...recording
  }
}

describe('useMeetingRecording', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('formats recording time as mm:ss', () => {
    expect(formatRecordingTime(0)).toBe('00:00')
    expect(formatRecordingTime(65)).toBe('01:05')
    expect(formatRecordingTime(600)).toBe('10:00')
  })

  it('warns and does not create a recorder when no media streams exist', () => {
    const recording = createRecorderHarness()

    expect(recording.startRecording()).toBe(false)

    expect(ElMessage.warning).toHaveBeenCalledWith('没有可录制的媒体流')
    expect(recording.combinedStream.addTrack).not.toHaveBeenCalled()
    expect(recording.isRecording.value).toBe(false)
  })

  it('combines local and remote tracks, records chunks, and downloads on stop', () => {
    const localAudio = createTrack('local-audio')
    const localVideo = createTrack('local-video')
    const remoteVideo = createTrack('remote-video')
    const recording = createRecorderHarness({
      localStream: createStream(localAudio, localVideo),
      remotePeers: {
        2: { stream: createStream(remoteVideo) },
        3: { name: 'no stream yet' }
      }
    })

    expect(recording.startRecording()).toBe(true)

    expect(recording.combinedStream.addTrack).toHaveBeenCalledWith(localAudio)
    expect(recording.combinedStream.addTrack).toHaveBeenCalledWith(localVideo)
    expect(recording.combinedStream.addTrack).toHaveBeenCalledWith(remoteVideo)
    expect(recording.recorder.start).toHaveBeenCalledWith(1000)
    expect(recording.isRecording.value).toBe(true)
    expect(ElMessage.success).toHaveBeenCalledWith('开始录制')

    recording.tick()
    recording.tick()
    expect(recording.recordingTime.value).toBe(2)

    const chunk = { size: 16, payload: 'webm' }
    recording.recorder.ondataavailable({ data: chunk })

    expect(recording.stopRecording()).toBe(true)

    expect(recording.recorder.stop).toHaveBeenCalled()
    expect(recording.clearTimer).toHaveBeenCalledWith('recording-timer')
    expect(recording.createBlob).toHaveBeenCalledWith([chunk], { type: 'video/webm' })
    expect(recording.createObjectUrl).toHaveBeenCalled()
    expect(recording.anchor.href).toBe('blob:meeting-recording')
    expect(recording.anchor.download).toBe('会议录制_2026-05-31 05:10:00.webm')
    expect(recording.anchor.click).toHaveBeenCalled()
    expect(recording.revokeObjectUrl).toHaveBeenCalledWith('blob:meeting-recording')
    expect(recording.isRecording.value).toBe(false)
    expect(ElMessage.success).toHaveBeenCalledWith('录制已保存')
  })

  it('reports startup failures and cleanup is safe when inactive', () => {
    const logger = { error: vi.fn() }
    const recording = createRecorderHarness({
      localStream: createStream(createTrack('local-video')),
      createMediaRecorder: vi.fn(() => {
        throw new Error('unsupported')
      }),
      logger
    })

    expect(recording.startRecording()).toBe(false)
    expect(recording.isRecording.value).toBe(false)
    expect(ElMessage.error).toHaveBeenCalledWith('录制启动失败')

    recording.cleanupRecording()
    expect(recording.createBlob).not.toHaveBeenCalled()
    expect(logger.error).toHaveBeenCalledWith(expect.any(Error))
  })
})
