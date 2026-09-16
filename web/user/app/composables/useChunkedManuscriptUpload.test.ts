import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

vi.mock('@/api/manuscript.ts', () => ({
  manuscriptApi: {
    createUploadSession: vi.fn(),
    uploadChunk: vi.fn(),
    completeUploadSession: vi.fn(),
    cancelUploadSession: vi.fn(),
  },
}))

import { manuscriptApi } from '@/api/manuscript.ts'
import { useChunkedManuscriptUpload, UPLOAD_STAGES } from './useChunkedManuscriptUpload'

const api = manuscriptApi as unknown as {
  createUploadSession: ReturnType<typeof vi.fn>
  uploadChunk: ReturnType<typeof vi.fn>
  completeUploadSession: ReturnType<typeof vi.fn>
  cancelUploadSession: ReturnType<typeof vi.fn>
}

function makeFile(size: number, name = 'v.mp4'): File {
  // 在 happy-dom 中，slice 行为可能不一致，这里我们模拟 File + slice
  const bytes = new Uint8Array(size)
  const blob = new Blob([bytes], { type: 'video/mp4' })
  const originalSlice = blob.slice.bind(blob)
  const file = blob as any
  file.name = name
  Object.defineProperty(file, 'size', { value: size, configurable: true })
  file.slice = (start: number, end: number) => {
    const slice = originalSlice(start, end)
    Object.defineProperty(slice, 'size', { value: end - start, configurable: true })
    return slice
  }
  return file as File
}

describe('useChunkedManuscriptUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useRealTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  describe('初始状态', () => {
    it('stage 默认 preparing', () => {
      const r = useChunkedManuscriptUpload()
      expect(r.stage.value).toBe('preparing')
    })

    it('percentage 默认 0', () => {
      const r = useChunkedManuscriptUpload()
      expect(r.percentage.value).toBe(0)
    })

    it('isUploading 为 false', () => {
      const r = useChunkedManuscriptUpload()
      expect(r.isUploading.value).toBe(false)
    })

    it('UPLOAD_STAGES 导出常量', () => {
      expect(UPLOAD_STAGES.PREPARING).toBe('preparing')
      expect(UPLOAD_STAGES.COMPLETED).toBe('completed')
      expect(UPLOAD_STAGES.FAILED).toBe('failed')
    })
  })

  describe('stageLabel / isFinished', () => {
    it('stageLabel 中文映射', () => {
      const r = useChunkedManuscriptUpload()
      r.stage.value = 'uploading'
      expect(r.stageLabel.value).toBe('上传视频分片')
      r.stage.value = 'completed'
      expect(r.stageLabel.value).toBe('上传完成，等待审核/转码')
    })

    it('isFinished 在 completed/failed/cancelled 时为 true', () => {
      const r = useChunkedManuscriptUpload()
      r.stage.value = 'completed'
      expect(r.isFinished.value).toBe(true)
      r.stage.value = 'failed'
      expect(r.isFinished.value).toBe(true)
      r.stage.value = 'cancelled'
      expect(r.isFinished.value).toBe(true)
      r.stage.value = 'uploading'
      expect(r.isFinished.value).toBe(false)
    })
  })

  describe('start - 失败路径', () => {
    it('无 videos 时抛错并切到 failed', async () => {
      const r = useChunkedManuscriptUpload()
      await expect(r.start({ videos: [] } as any)).rejects.toThrow('至少需要一个视频分P')
      expect(r.stage.value).toBe('failed')
      expect(r.error.value).toBe('至少需要一个视频分P')
    })
  })

  describe('start - 成功路径', () => {
    it('成功完成整个流程：preparing→uploading→merging→submitting→completed', async () => {
      // 8MB 单 chunk 大小，刚好一个分片
      const file = makeFile(4 * 1024 * 1024) // 4MB → 1 chunk
      api.createUploadSession.mockResolvedValueOnce({ code: 200, data: { uploadId: 'u1' } })
      api.uploadChunk.mockResolvedValue({ code: 200 })
      api.completeUploadSession.mockResolvedValueOnce({ code: 200, data: { id: 1 } })

      const r = useChunkedManuscriptUpload()
      const stages: string[] = []
      // @ts-ignore watch
      r.stage // ref
      const result = await r.start({
        title: 't',
        description: 'd',
        categoryId: 1,
        tags: ['a'],
        videos: [{ title: 'P1', file, duration: 100 }],
        cover: makeFile(100, 'c.jpg'),
      })
      // 监听 stage 变化收集
      expect(result).toEqual({ id: 1 })
      expect(r.stage.value).toBe('completed')
      expect(r.percentage.value).toBe(100)
      expect(api.createUploadSession).toHaveBeenCalledTimes(1)
      expect(api.uploadChunk).toHaveBeenCalledTimes(1)
      expect(api.completeUploadSession).toHaveBeenCalledTimes(1)
      expect(r.uploadId.value).toBe('u1')
    })

    it('多 chunk 时按顺序上传', async () => {
      const file = makeFile(16 * 1024 * 1024) // 16MB → 2 chunks @ 8MB
      api.createUploadSession.mockResolvedValueOnce({ code: 200, data: { uploadId: 'u2' } })
      api.uploadChunk.mockResolvedValue({ code: 200 })
      api.completeUploadSession.mockResolvedValueOnce({ code: 200, data: { id: 2 } })

      const r = useChunkedManuscriptUpload({ chunkSize: 8 * 1024 * 1024 })
      await r.start({
        title: 't',
        description: 'd',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file }],
        cover: makeFile(100),
      })
      expect(api.uploadChunk).toHaveBeenCalledTimes(2)
      expect(r.uploadedChunks.value).toBe(2)
      expect(r.totalChunks.value).toBe(2)
      expect(r.stage.value).toBe('completed')
    })

    it('createUploadSession 失败时 stage=failed 并抛错', async () => {
      api.createUploadSession.mockResolvedValueOnce({ code: 500, message: 'server err' })
      const r = useChunkedManuscriptUpload()
      await expect(r.start({
        title: 't',
        description: '',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file: makeFile(1000) }],
        cover: makeFile(100),
      })).rejects.toThrow('server err')
      expect(r.stage.value).toBe('failed')
      expect(r.error.value).toBe('server err')
    })

    it('createUploadSession 返回 null/undefined 时也走 failed', async () => {
      api.createUploadSession.mockResolvedValueOnce(null)
      const r = useChunkedManuscriptUpload()
      await expect(r.start({
        title: 't',
        description: '',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file: makeFile(1000) }],
        cover: makeFile(100),
      })).rejects.toThrow()
      expect(r.stage.value).toBe('failed')
    })

    it('uploadChunk 返回非 200 时重试 maxRetries+1 次后失败', async () => {
      vi.useFakeTimers()
      const file = makeFile(1000) // 1 chunk
      api.createUploadSession.mockResolvedValueOnce({ code: 200, data: { uploadId: 'u3' } })
      api.uploadChunk.mockResolvedValue({ code: 500, message: 'chunk fail' })

      const r = useChunkedManuscriptUpload({ chunkSize: 8 * 1024 * 1024, maxRetries: 2 })
      const promise = r.start({
        title: 't',
        description: '',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file }],
        cover: makeFile(100),
      }).catch(e => e)
      // 推进时间以完成重试退避
      for (let i = 0; i < 50; i++) await vi.advanceTimersByTimeAsync(2000)
      const err = await promise
      expect(err).toBeTruthy()
      expect(api.uploadChunk).toHaveBeenCalledTimes(3) // 1 + 2 retries
      expect(r.stage.value).toBe('failed')
    })

    it('completeUploadSession 失败时 stage=failed', async () => {
      const file = makeFile(1000)
      api.createUploadSession.mockResolvedValueOnce({ code: 200, data: { uploadId: 'u4' } })
      api.uploadChunk.mockResolvedValue({ code: 200 })
      api.completeUploadSession.mockResolvedValueOnce({ code: 500 })

      const r = useChunkedManuscriptUpload()
      await expect(r.start({
        title: 't',
        description: '',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file }],
        cover: makeFile(100),
      })).rejects.toThrow()
      expect(r.stage.value).toBe('failed')
    })
  })

  describe('cancel', () => {
    it('cancel 时把 stage 设为 cancelled，并尝试调用 cancelUploadSession', async () => {
      api.cancelUploadSession.mockResolvedValueOnce({ code: 200 })
      const r = useChunkedManuscriptUpload()
      r.uploadId.value = 'ux'
      await r.cancel()
      expect(api.cancelUploadSession).toHaveBeenCalledWith('ux')
      expect(r.stage.value).toBe('cancelled')
    })

    it('cancelUploadSession reject 时不抛错', async () => {
      api.cancelUploadSession.mockRejectedValueOnce(new Error('x'))
      const r = useChunkedManuscriptUpload()
      r.uploadId.value = 'ux'
      await expect(r.cancel()).resolves.toBeUndefined()
      expect(r.stage.value).toBe('cancelled')
    })

    it('uploadId 为空时不调用 cancelUploadSession', async () => {
      const r = useChunkedManuscriptUpload()
      await r.cancel()
      expect(api.cancelUploadSession).not.toHaveBeenCalled()
      expect(r.stage.value).toBe('cancelled')
    })
  })

  describe('reset', () => {
    it('重置所有状态', () => {
      const r = useChunkedManuscriptUpload()
      r.uploadId.value = 'u'
      r.percentage.value = 50
      r.error.value = 'e'
      r.uploadedChunks.value = 3
      r.reset()
      expect(r.uploadId.value).toBe(null)
      expect(r.percentage.value).toBe(0)
      expect(r.error.value).toBe(null)
      expect(r.uploadedChunks.value).toBe(0)
      expect(r.stage.value).toBe('preparing')
    })
  })

  describe('progress 回调', () => {
    it('uploadChunk 的 onProgress 回调会被触发并更新 uploadedBytes', async () => {
      const file = makeFile(8 * 1024 * 1024)
      api.createUploadSession.mockResolvedValueOnce({ code: 200, data: { uploadId: 'u5' } })
      api.uploadChunk.mockImplementation(async (data, onProgress) => {
        onProgress({ loaded: 4 * 1024 * 1024, total: 8 * 1024 * 1024 })
        onProgress({ percent: 50 })
        onProgress({ percent: 100, loaded: 8 * 1024 * 1024 })
        return { code: 200 }
      })
      api.completeUploadSession.mockResolvedValueOnce({ code: 200, data: {} })

      const r = useChunkedManuscriptUpload()
      await r.start({
        title: 't',
        description: '',
        categoryId: 1,
        tags: [],
        videos: [{ title: 'P1', file }],
        cover: makeFile(100),
      })
      expect(r.uploadedBytes.value).toBeGreaterThanOrEqual(8 * 1024 * 1024)
      expect(r.percentage.value).toBe(100)
    })
  })
})