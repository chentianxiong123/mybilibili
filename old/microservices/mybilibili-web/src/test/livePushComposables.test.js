import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useWhipPublisher } from '../composables/useWhipPublisher.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn()
  }
}))

class FakeRTCPeerConnection {
  static instances = []

  constructor(config) {
    this.config = config
    this.tracks = []
    this.localDescription = null
    this.remoteDescription = null
    this.closed = false
    FakeRTCPeerConnection.instances.push(this)
  }

  addTrack(track, stream) {
    this.tracks.push({ track, stream })
  }

  async createOffer() {
    return { type: 'offer', sdp: 'offer-sdp' }
  }

  async setLocalDescription(description) {
    this.localDescription = description
  }

  async setRemoteDescription(description) {
    this.remoteDescription = description
  }

  close() {
    this.closed = true
  }
}

const createTrack = () => ({
  stop: vi.fn(),
  onended: null
})

const createStream = () => {
  const videoTrack = createTrack()
  return {
    videoTrack,
    getTracks: () => [videoTrack],
    getVideoTracks: () => [videoTrack]
  }
}

const createPublisher = (options = {}) => {
  const {
    streamKey: streamKeyValue = 'stream-1',
    captureStream = vi.fn().mockResolvedValue(createStream()),
    onPublished = vi.fn(),
    logger = { error: vi.fn() },
    ...publisherOptions
  } = options
  const streamKey = ref(streamKeyValue)
  const publisher = useWhipPublisher({
    streamKey,
    captureStream,
    onPublished,
    logger,
    getWhipUrl: key => `https://whip.example/${key}`,
    ...publisherOptions
  })

  return {
    streamKey,
    captureStream,
    onPublished,
    logger,
    ...publisher
  }
}

describe('useWhipPublisher', () => {
  beforeEach(() => {
    FakeRTCPeerConnection.instances = []
    vi.stubGlobal('RTCPeerConnection', FakeRTCPeerConnection)
    vi.stubGlobal('fetch', vi.fn())
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('warns when streamKey is missing', async () => {
    const publisher = createPublisher({ streamKey: '' })

    await publisher.startPublishing()

    expect(ElMessage.warning).toHaveBeenCalledWith('缺少直播流密钥')
    expect(publisher.captureStream).not.toHaveBeenCalled()
    expect(FakeRTCPeerConnection.instances).toHaveLength(0)
  })

  it('publishes camera stream and applies WHIP answer', async () => {
    const stream = createStream()
    const captureStream = vi.fn().mockResolvedValue(stream)
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: vi.fn().mockResolvedValue('answer-sdp')
    })
    vi.stubGlobal('fetch', fetchMock)
    const publisher = createPublisher({ captureStream })
    publisher.videoPreviewRef.value = { srcObject: null }

    await publisher.startPublishing()

    const pc = FakeRTCPeerConnection.instances[0]
    expect(publisher.isPublishing.value).toBe(true)
    expect(publisher.mediaStream.value).toEqual(stream)
    expect(publisher.videoPreviewRef.value.srcObject).toEqual(stream)
    expect(pc.tracks).toEqual([{ track: stream.videoTrack, stream }])
    expect(fetchMock).toHaveBeenCalledWith('https://whip.example/stream-1', {
      method: 'POST',
      headers: { 'Content-Type': 'application/sdp' },
      body: 'offer-sdp'
    })
    expect(pc.remoteDescription).toEqual({ type: 'answer', sdp: 'answer-sdp' })
    expect(publisher.onPublished).toHaveBeenCalled()
    expect(ElMessage.success).toHaveBeenCalledWith('摄像头推流中')
  })

  it('cleans up stream and connection when WHIP negotiation fails', async () => {
    const stream = createStream()
    const captureStream = vi.fn().mockResolvedValue(stream)
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      text: vi.fn()
    })
    vi.stubGlobal('fetch', fetchMock)
    const publisher = createPublisher({ captureStream })
    publisher.videoPreviewRef.value = { srcObject: null }

    await publisher.startPublishing()

    const pc = FakeRTCPeerConnection.instances[0]
    expect(ElMessage.error).toHaveBeenCalledWith('WHIP push failed')
    expect(publisher.logger.error).toHaveBeenCalledWith(expect.any(Error))
    expect(stream.videoTrack.stop).toHaveBeenCalled()
    expect(pc.closed).toBe(true)
    expect(publisher.isPublishing.value).toBe(false)
    expect(publisher.mediaStream.value).toBe(null)
    expect(publisher.videoPreviewRef.value.srcObject).toBe(null)
  })

  it('stops publishing when a screen track ends', async () => {
    const stream = createStream()
    const captureStream = vi.fn().mockResolvedValue(stream)
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: vi.fn().mockResolvedValue('answer-sdp')
    })
    vi.stubGlobal('fetch', fetchMock)
    const publisher = createPublisher({ captureStream })
    publisher.videoPreviewRef.value = { srcObject: null }
    publisher.sourceType.value = 'screen'
    await Promise.resolve()

    await publisher.startPublishing()
    stream.videoTrack.onended()

    const pc = FakeRTCPeerConnection.instances[0]
    expect(stream.videoTrack.stop).toHaveBeenCalled()
    expect(pc.closed).toBe(true)
    expect(publisher.isPublishing.value).toBe(false)
  })
})
