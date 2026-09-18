import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPut: vi.fn(),
  apiDelete: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: mocks.apiPut,
    delete: mocks.apiDelete,
  }),
}))

import {
  manuscriptApi,
  getPendingManuscripts,
  getProcessingManuscripts,
  getAllManuscripts,
  getManuscriptDetail,
  approveManuscript,
  approveWithProcess,
  rejectManuscript,
  publishManuscript,
  unpublishManuscript,
  getManuscriptVideos,
  getManuscriptStatistics,
  retryManuscript,
  manualTranscode,
  manualExtractAudio,
  manualGenerateSubtitle,
  manualAiSummary,
  manualProcessAll,
  resetVideoStatus,
  getVideoSourceUrl,
} from './manuscript'

describe('manuscript api - manuscriptApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uploadManuscript 用 multipart + onUploadProgress 透传', async () => {
    mocks.apiPost.mockImplementationOnce(async (_url: string, _data: any, cfg: any) => {
      cfg.onUploadProgress?.({ loaded: 50, total: 100 })
      return { code: 200 }
    })
    const onProgress = vi.fn()
    await manuscriptApi.uploadManuscript(
      {
        title: 'T',
        description: 'D',
        cover: new File(['c'], 'c.jpg'),
        categoryId: 2,
        tags: ['a'],
        videos: [{ file: new File(['v'], 'v.mp4'), title: 'P1', sortOrder: 0, durationSeconds: 10 }],
      },
      onProgress,
    )
    expect(mocks.apiPost).toHaveBeenCalled()
    const [url, , cfg] = mocks.apiPost.mock.calls[0]
    expect(url).toBe('/manuscript/upload')
    expect(cfg.timeout).toBe(300000)
    expect(cfg.headers).toEqual({ 'Content-Type': 'multipart/form-data' })
    expect(onProgress).toHaveBeenCalledWith(50)
  })

  it('uploadManuscript 无 onProgress 时不设置 onUploadProgress', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.uploadManuscript({
      title: 'T',
      cover: new File(['c'], 'c.jpg'),
      categoryId: 2,
    })
    const [, , cfg] = mocks.apiPost.mock.calls[0]
    expect(cfg.onUploadProgress).toBeUndefined()
  })

  it('createUploadSession', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.createUploadSession({ fileName: 'x.mp4' })
    expect(mocks.apiPost).toHaveBeenCalledWith(
      '/manuscript/upload-session',
      { fileName: 'x.mp4' },
      { headers: { 'Content-Type': 'application/json' } },
    )
  })

  it('getUploadSessionStatus', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getUploadSessionStatus('upload_1')
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/upload-session/upload_1')
  })

  it('uploadChunk 携带进度回调', async () => {
    mocks.apiPost.mockImplementationOnce(async (_u: string, _d: any, cfg: any) => {
      cfg.onUploadProgress?.({ loaded: 50, total: 200 })
      return { code: 200 }
    })
    const onProgress = vi.fn()
    await manuscriptApi.uploadChunk(
      {
        uploadId: 'u',
        partIndex: 0,
        chunkIndex: 0,
        totalChunks: 1,
        file: new File(['x'], 'a'),
      },
      onProgress,
    )
    expect(onProgress).toHaveBeenCalledWith({ percent: 25, loaded: 50, total: 200 })
    const [, , cfg] = mocks.apiPost.mock.calls[0]
    expect(cfg.timeout).toBe(120000)
  })

  it('uploadChunk 无 total 时回退到 file.size', async () => {
    mocks.apiPost.mockImplementationOnce(async (_u: string, _d: any, cfg: any) => {
      cfg.onUploadProgress?.({ loaded: 0, total: 0 })
      return { code: 200 }
    })
    const onProgress = vi.fn()
    await manuscriptApi.uploadChunk(
      {
        uploadId: 'u',
        partIndex: 0,
        chunkIndex: 0,
        totalChunks: 1,
        file: new File(['x'], 'a'),
      },
      onProgress,
    )
    expect(onProgress).toHaveBeenCalledWith({ percent: 0, loaded: 0, total: 1 })
  })

  it('completeUploadSession', async () => {
    mocks.apiPost.mockImplementationOnce(async (_u: string, _d: any, cfg: any) => {
      cfg.onUploadProgress?.({ loaded: 50, total: 100 })
      return { code: 200 }
    })
    const onProgress = vi.fn()
    const cover = new File(['c'], 'c.jpg')
    await manuscriptApi.completeUploadSession('upload_1', cover, onProgress)
    expect(mocks.apiPost).toHaveBeenCalled()
    const [url, form, cfg] = mocks.apiPost.mock.calls[0]
    expect(url).toBe('/manuscript/upload-complete')
    expect(form.get('uploadId')).toBe('upload_1')
    expect(form.get('cover')).toBe(cover)
    expect(cfg.timeout).toBe(600000)
    expect(onProgress).toHaveBeenCalledWith(50)
  })

  it('cancelUploadSession', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.cancelUploadSession('upload_1')
    expect(mocks.apiDelete).toHaveBeenCalledWith('/manuscript/upload-session/upload_1')
  })

  it('getManuscriptList 默认 page/size', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await manuscriptApi.getManuscriptList()
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/list?page=1&size=10')
  })

  it('getManuscriptList 带 status', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await manuscriptApi.getManuscriptList(2, 5, 'PUBLISHED')
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/list?page=2&size=5&status=PUBLISHED')
  })

  it('getManuscriptById', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getManuscriptById(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/7')
  })

  it('getRecommendedManuscripts', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getRecommendedManuscripts()
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/recommended')
  })

  it('updateManuscript 追加 tags', async () => {
    mocks.apiPut.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.updateManuscript(7, {
      title: 'T',
      description: 'D',
      categoryId: 2,
      cover: new File(['c'], 'c.jpg'),
      tags: ['a', 'b'],
    })
    const [url, form] = mocks.apiPut.mock.calls[0]
    expect(url).toBe('/manuscript/7')
    expect(form.get('title')).toBe('T')
    expect(form.get('tags')).not.toBeNull()
    expect(form.getAll('tags')).toEqual(['a', 'b'])
  })

  it('deleteManuscript', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.deleteManuscript(7)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/manuscript/7')
  })

  it('getUserManuscripts 拼接 userId', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getUserManuscripts(5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/user/5?page=1&size=10')
  })

  it('getUserManuscripts 带 status', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getUserManuscripts(5, 2, 5, 'PUBLISHED')
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/user/5?page=2&size=5&status=PUBLISHED')
  })

  it('getManuscriptStats', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await manuscriptApi.getManuscriptStats(5)
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/user/5/stats')
  })
})

describe('manuscript api - admin functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getPendingManuscripts / getProcessingManuscripts / getAllManuscripts', async () => {
    mocks.apiGet.mockResolvedValue({ code: 200 })
    await getPendingManuscripts()
    await getProcessingManuscripts()
    await getAllManuscripts()
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/manuscript/admin/pending')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/manuscript/admin/processing')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(3, '/manuscript/admin/all')
  })

  it('getManuscriptDetail', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await getManuscriptDetail(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/admin/7')
  })

  it('approveManuscript POST body', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await approveManuscript(7, 99, 'ok')
    expect(mocks.apiPost).toHaveBeenCalledWith('/manuscript/admin/approve/7', { reviewerId: 99, reason: 'ok' })
  })

  it('approveWithProcess autoProcess=true', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await approveWithProcess(7, true)
    expect(mocks.apiPost).toHaveBeenCalledWith('/manuscript/admin/7/approve-with-process', { autoProcess: true })
  })

  it('approveWithProcess 默认 autoProcess=false', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await approveWithProcess(7)
    expect(mocks.apiPost).toHaveBeenCalledWith('/manuscript/admin/7/approve-with-process', { autoProcess: false })
  })

  it('rejectManuscript', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await rejectManuscript(7, 99, 'no')
    expect(mocks.apiPost).toHaveBeenCalledWith('/manuscript/admin/reject/7', { reviewerId: 99, reason: 'no' })
  })

  it('publishManuscript / unpublishManuscript', async () => {
    mocks.apiPost.mockResolvedValue({ code: 200 })
    await publishManuscript(7)
    await unpublishManuscript(7)
    expect(mocks.apiPost).toHaveBeenNthCalledWith(1, '/manuscript/admin/publish/7')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(2, '/manuscript/admin/unpublish/7')
  })

  it('getManuscriptVideos / getManuscriptStatistics', async () => {
    mocks.apiGet.mockResolvedValue({ code: 200 })
    await getManuscriptVideos(7)
    await getManuscriptStatistics()
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/manuscript/admin/7/videos')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/manuscript/admin/statistics')
  })

  it('retryManuscript / manualTranscode / manualExtractAudio / manualGenerateSubtitle / manualAiSummary / manualProcessAll / resetVideoStatus', async () => {
    mocks.apiPost.mockResolvedValue({ code: 200 })
    await retryManuscript(1)
    await manualTranscode(2)
    await manualExtractAudio(3)
    await manualGenerateSubtitle(4)
    await manualAiSummary(5)
    await manualProcessAll(6)
    await resetVideoStatus(7)
    expect(mocks.apiPost).toHaveBeenNthCalledWith(1, '/manuscript/admin/retry/1')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(2, '/manuscript/admin/transcode/2')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(3, '/manuscript/admin/extract-audio/3')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(4, '/manuscript/admin/generate-subtitle/4')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(5, '/manuscript/admin/ai-summary/5')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(6, '/manuscript/admin/process-all/6')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(7, '/manuscript/admin/reset/7')
  })

  it('getVideoSourceUrl', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200 })
    await getVideoSourceUrl(9)
    expect(mocks.apiGet).toHaveBeenCalledWith('/manuscript/admin/video-source/9')
  })
})
