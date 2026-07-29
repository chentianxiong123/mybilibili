import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useChunkedManuscriptUpload, UPLOAD_STAGES } from '../composables/useChunkedManuscriptUpload.js'

function createMockFile(name, size) {
  const chunks = []
  for (let i = 0; i < size; i += 4) {
    chunks.push(new Uint8Array([0x48, 0x65, 0x6c, 0x6c]))
  }
  return new File(chunks, name, { type: 'video/mp4' })
}

function createMockApi(overrides = {}) {
  return {
    createUploadSession: vi.fn().mockResolvedValue({
      code: 200,
      data: { uploadId: 'test-upload-id', totalParts: 1 }
    }),
    uploadChunk: vi.fn().mockResolvedValue({
      code: 200,
      data: { complete: true }
    }),
    completeUploadSession: vi.fn().mockResolvedValue({
      code: 200,
      data: { id: 1, title: 'test' }
    }),
    cancelUploadSession: vi.fn().mockResolvedValue({ code: 200 }),
    ...overrides
  }
}

function createManuscriptData(videoCount = 1, fileSize = 100) {
  return {
    title: 'Test Manuscript',
    description: 'A test',
    categoryId: 1,
    tags: ['tag1'],
    cover: new File(['cover'], 'cover.jpg', { type: 'image/jpeg' }),
    videos: Array.from({ length: videoCount }, (_, i) => ({
      file: createMockFile('video' + i + '.mp4', fileSize),
      title: 'Part ' + (i + 1),
      sortOrder: i,
      duration: 30
    }))
  }
}

describe('useChunkedManuscriptUpload', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('completes single-part upload through full lifecycle', async () => {
    const api = createMockApi()
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 50, maxConcurrent: 1 })

    expect(upload.stage.value).toBe(UPLOAD_STAGES.PREPARING)
    expect(upload.isFinished.value).toBe(false)

    const data = createManuscriptData(1, 100)
    const promise = upload.start(data)
    await vi.advanceTimersByTimeAsync(0)
    await promise

    expect(upload.stage.value).toBe(UPLOAD_STAGES.COMPLETED)
    expect(upload.percentage.value).toBe(100)
    expect(upload.isFinished.value).toBe(true)
    expect(upload.uploadId.value).toBe('test-upload-id')

    expect(api.createUploadSession).toHaveBeenCalledTimes(1)
    expect(api.uploadChunk).toHaveBeenCalledTimes(2)
    expect(api.completeUploadSession).toHaveBeenCalledWith('test-upload-id', data.cover)
  })

  it('tracks multi-part progress with concurrency', async () => {
    const api = createMockApi()
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 50, maxConcurrent: 2 })

    const data = createManuscriptData(2, 100)
    const promise = upload.start(data)
    await vi.advanceTimersByTimeAsync(0)
    await promise

    expect(upload.stage.value).toBe(UPLOAD_STAGES.COMPLETED)
    expect(api.uploadChunk).toHaveBeenCalledTimes(4)
    expect(upload.partProgress.value).toHaveLength(2)
  })

  it('tracks in-flight chunk byte progress before the chunk request completes', async () => {
    let resolveChunk
    const api = createMockApi({
      uploadChunk: vi.fn().mockImplementation((chunkData, onProgress) => {
        onProgress({ loaded: 24, total: 48, percent: 50 })
        return new Promise(resolve => {
          resolveChunk = () => resolve({ code: 200, data: {} })
        })
      })
    })
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 48, maxConcurrent: 1 })

    const data = createManuscriptData(1, 48)
    const promise = upload.start(data)
    await vi.advanceTimersByTimeAsync(0)

    expect(upload.uploadedBytes.value).toBe(24)
    expect(upload.percentage.value).toBe(50)

    resolveChunk()
    await vi.advanceTimersByTimeAsync(0)
    await promise

    expect(upload.stage.value).toBe(UPLOAD_STAGES.COMPLETED)
    expect(upload.uploadedBytes.value).toBe(48)
    expect(upload.percentage.value).toBe(100)
  })

  it('sets error state when createSession fails', async () => {
    const api = createMockApi({
      createUploadSession: vi.fn().mockResolvedValue({
        code: 500,
        message: '创建失败'
      })
    })
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 50 })

    const data = createManuscriptData(1, 100)
    await expect(upload.start(data)).rejects.toThrow()
    expect(upload.stage.value).toBe(UPLOAD_STAGES.FAILED)
    expect(upload.error.value).toBe('创建失败')
  })

  it('retries failed chunks and eventually succeeds', async () => {
    let callCount = 0
    const api = createMockApi({
      uploadChunk: vi.fn().mockImplementation(async () => {
        callCount++
        if (callCount === 1) throw new Error('network error')
        return { code: 200, data: {} }
      })
    })
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 50, maxRetries: 2 })

    const data = createManuscriptData(1, 100)
    const promise = upload.start(data)
    await vi.advanceTimersByTimeAsync(5000)
    await promise

    expect(upload.stage.value).toBe(UPLOAD_STAGES.COMPLETED)
    expect(upload.retryCount.value).toBe(1)
  })

  it('cancels upload and marks as cancelled', async () => {
    let resolveChunk
    const api = createMockApi({
      uploadChunk: vi.fn().mockImplementation(() => new Promise(r => { resolveChunk = r }))
    })
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 50, maxConcurrent: 1 })

    const data = createManuscriptData(1, 100)
    const startPromise = upload.start(data).catch(e => e)
    await vi.advanceTimersByTimeAsync(0)

    expect(upload.stage.value).toBe(UPLOAD_STAGES.UPLOADING)
    expect(api.uploadChunk).toHaveBeenCalledTimes(1)

    await upload.cancel()
    resolveChunk({ code: 200, data: {} })

    const err = await startPromise
    expect(err).toBeInstanceOf(Error)
    expect(upload.stage.value).toBe(UPLOAD_STAGES.CANCELLED)
    expect(api.cancelUploadSession).toHaveBeenCalledWith('test-upload-id')
  })

  it('computes chunk counts correctly', async () => {
    const api = createMockApi()
    const upload = useChunkedManuscriptUpload({ api, chunkSize: 32 })

    const data = createManuscriptData(1, 100)
    const promise = upload.start(data)
    await vi.advanceTimersByTimeAsync(0)
    await promise

    expect(api.uploadChunk).toHaveBeenCalledTimes(4)
  })
})
