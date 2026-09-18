import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import {
  manuscriptApi,
  getPendingManuscripts, getProcessingManuscripts, getAllManuscripts,
  getManuscriptDetail, approveManuscript, approveWithProcess, rejectManuscript,
  publishManuscript, unpublishManuscript, getManuscriptVideos,
  getManuscriptStatistics, retryManuscript, manualTranscode,
  manualExtractAudio, manualGenerateSubtitle, manualAiSummary,
  manualProcessAll, resetVideoStatus, getVideoSourceUrl
} from './manuscript'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
})

describe('manuscript api - 上传会话', () => {
  it('createUploadSession POST /manuscript/upload-session', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: { uploadId: 'u1' } })
    await manuscriptApi.createUploadSession({ foo: 'bar' })
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/upload-session', { foo: 'bar' }, expect.any(Object))
  })

  it('getUploadSessionStatus GET /manuscript/upload-session/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { status: 'uploading' } })
    await manuscriptApi.getUploadSessionStatus('u1')
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/upload-session/u1')
  })

  it('uploadChunk POST /manuscript/upload-chunk multipart 带 progress 回调', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const file = new File(['x'], 'a.bin')
    const onProgress = vi.fn()
    const fakeEvent = { loaded: 50, total: 100 }
    requestMock.post.mockImplementationOnce(async (_url: string, _data: any, config: any) => {
      config.onUploadProgress(fakeEvent)
      return { code: 200, data: null }
    })
    await manuscriptApi.uploadChunk({ uploadId: 'u1', partIndex: 0, chunkIndex: 0, totalChunks: 1, file }, onProgress)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/upload-chunk', expect.any(FormData), expect.objectContaining({
      timeout: 120000,
      headers: { 'Content-Type': 'multipart/form-data' },
    }))
    expect(onProgress).toHaveBeenCalledWith({ percent: 50, loaded: 50, total: 100 })
  })

  it('completeUploadSession POST /manuscript/upload-complete 带 cover', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const cover = new File(['x'], 'c.jpg')
    const onProgress = vi.fn()
    const fakeEvent = { loaded: 100, total: 100 }
    requestMock.post.mockImplementationOnce(async (_url: string, _data: any, config: any) => {
      config.onUploadProgress(fakeEvent)
      return { code: 200, data: null }
    })
    await manuscriptApi.completeUploadSession('u1', cover, onProgress)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/upload-complete', expect.any(FormData), expect.objectContaining({ timeout: 600000 }))
    expect(onProgress).toHaveBeenCalledWith(100)
  })

  it('cancelUploadSession DELETE /manuscript/upload-session/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await manuscriptApi.cancelUploadSession('u1')
    expect(requestMock.delete).toHaveBeenCalledWith('/manuscript/upload-session/u1')
  })

  it('uploadManuscript POST /manuscript/upload multipart 含 title/tags/videos', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const cover = new File(['c'], 'c.jpg')
    const videoFile = new File(['v'], 'v.mp4')
    await manuscriptApi.uploadManuscript({
      title: 'T',
      description: 'D',
      cover,
      categoryId: 1,
      tags: ['a', 'b'],
      videos: [{ file: videoFile, title: 'P1', sortOrder: 0, durationSeconds: 60 }],
    })
    expect(requestMock.post).toHaveBeenCalledWith(
      '/manuscript/upload',
      expect.any(FormData),
      expect.objectContaining({ timeout: 300000, headers: { 'Content-Type': 'multipart/form-data' } }),
    )
  })

  it('uploadManuscript 无 tags / 无 videos 仍能构造表单', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const cover = new File(['c'], 'c.jpg')
    await manuscriptApi.uploadManuscript({ title: 'T', cover, categoryId: 1 })
    expect(requestMock.post).toHaveBeenCalled()
  })

  it('uploadManuscript 自定义 onProgress 回调被调用', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const onProgress = vi.fn()
    const fakeEvent = { loaded: 50, total: 100 }
    requestMock.post.mockImplementationOnce(async (_u: string, _d: any, config: any) => {
      config.onUploadProgress(fakeEvent)
      return { code: 200, data: null }
    })
    await manuscriptApi.uploadManuscript({
      title: 'T',
      cover: new File(['c'], 'c.jpg'),
      categoryId: 1,
    }, onProgress)
    expect(onProgress).toHaveBeenCalledWith(50)
  })
})

describe('manuscript api - 列表/详情', () => {
  it('getManuscriptList 默认参数', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await manuscriptApi.getManuscriptList()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/list?page=1&size=10')
  })

  it('getManuscriptList 带 status 拼接', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await manuscriptApi.getManuscriptList(2, 20, 'pending')
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/list?page=2&size=20&status=pending')
  })

  it('getManuscriptById GET /manuscript/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { id: 5 } })
    await manuscriptApi.getManuscriptById(5)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/5')
  })

  it('getRecommendedManuscripts GET /manuscript/recommended', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await manuscriptApi.getRecommendedManuscripts()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/recommended')
  })

  it('updateManuscript PUT /manuscript/:id 仅含部分字段', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await manuscriptApi.updateManuscript(5, { title: 'New' })
    expect(requestMock.put).toHaveBeenCalledWith('/manuscript/5', expect.any(FormData))
  })

  it('deleteManuscript DELETE /manuscript/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await manuscriptApi.deleteManuscript(5)
    expect(requestMock.delete).toHaveBeenCalledWith('/manuscript/5')
  })

  it('getUserManuscripts 默认 status=null 不拼接', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await manuscriptApi.getUserManuscripts(3)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/user/3?page=1&size=10')
  })

  it('getUserManuscripts 带 status 拼接', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await manuscriptApi.getUserManuscripts(3, 2, 20, 'pending')
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/user/3?page=2&size=20&status=pending')
  })

  it('getManuscriptStats GET /manuscript/user/:id/stats', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: {} })
    await manuscriptApi.getManuscriptStats(3)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/user/3/stats')
  })
})

describe('manuscript api - 管理员操作', () => {
  it('getPendingManuscripts GET /manuscript/admin/pending', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getPendingManuscripts()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/pending')
  })

  it('getProcessingManuscripts GET /manuscript/admin/processing', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getProcessingManuscripts()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/processing')
  })

  it('getAllManuscripts GET /manuscript/admin/all', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getAllManuscripts()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/all')
  })

  it('getManuscriptDetail GET /manuscript/admin/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: {} })
    await getManuscriptDetail(7)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/7')
  })

  it('approveManuscript POST 带 reviewerId/reason', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await approveManuscript(7, 100, 'ok')
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/approve/7', { reviewerId: 100, reason: 'ok' })
  })

  it('approveWithProcess 默认 autoProcess=false', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await approveWithProcess(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/7/approve-with-process', { autoProcess: false })
  })

  it('approveWithProcess autoProcess=true', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await approveWithProcess(7, true)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/7/approve-with-process', { autoProcess: true })
  })

  it('rejectManuscript POST 带 reviewerId/reason', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await rejectManuscript(7, 100, '不符合')
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/reject/7', { reviewerId: 100, reason: '不符合' })
  })

  it('publishManuscript POST /manuscript/admin/publish/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await publishManuscript(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/publish/7')
  })

  it('unpublishManuscript POST /manuscript/admin/unpublish/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await unpublishManuscript(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/unpublish/7')
  })

  it('getManuscriptVideos GET /manuscript/admin/:id/videos', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getManuscriptVideos(7)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/7/videos')
  })

  it('getManuscriptStatistics GET /manuscript/admin/statistics', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: {} })
    await getManuscriptStatistics()
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/statistics')
  })

  it('retryManuscript POST /manuscript/admin/retry/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await retryManuscript(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/retry/7')
  })

  it('manualTranscode POST /manuscript/admin/transcode/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await manualTranscode(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/transcode/7')
  })

  it('manualExtractAudio POST /manuscript/admin/extract-audio/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await manualExtractAudio(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/extract-audio/7')
  })

  it('manualGenerateSubtitle POST /manuscript/admin/generate-subtitle/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await manualGenerateSubtitle(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/generate-subtitle/7')
  })

  it('manualAiSummary POST /manuscript/admin/ai-summary/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await manualAiSummary(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/ai-summary/7')
  })

  it('manualProcessAll POST /manuscript/admin/process-all/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await manualProcessAll(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/process-all/7')
  })

  it('resetVideoStatus POST /manuscript/admin/reset/:id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await resetVideoStatus(7)
    expect(requestMock.post).toHaveBeenCalledWith('/manuscript/admin/reset/7')
  })

  it('getVideoSourceUrl GET /manuscript/admin/video-source/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { url: 'a' } })
    await getVideoSourceUrl(7)
    expect(requestMock.get).toHaveBeenCalledWith('/manuscript/admin/video-source/7')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('boom'))
    await expect(getPendingManuscripts()).rejects.toThrow('boom')
  })
})
