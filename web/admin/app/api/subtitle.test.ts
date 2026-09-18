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
  getVideosWithSubtitleInfo, getVideoSubtitles, uploadSubtitle,
  importSrtToMongo, setDefaultSubtitle, deleteSubtitle,
  getPendingSubtitles, approveSubtitle, rejectSubtitle,
  previewSubtitle, scanSystemSubtitles, importSystemSubtitle,
  subtitleApi
} from './subtitle'
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

describe('subtitle api - subtitleApi 对象', () => {
  it('getSubtitles GET /subtitle/video/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await subtitleApi.getSubtitles(7)
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/video/7')
  })

  it('getSubtitle GET /subtitle/video/:id/:lang', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await subtitleApi.getSubtitle(7, 'zh-CN')
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/video/7/zh-CN')
  })

  it('uploadSubtitle POST 透传 data', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await subtitleApi.uploadSubtitle({ foo: 'bar' })
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/upload', { foo: 'bar' })
  })

  it('uploadSrt POST multipart 带 FormData 含 video_id/file/language', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const file = new File(['x'], 'a.srt')
    await subtitleApi.uploadSrt(7, file, 'zh-CN', '中文', true)
    expect(requestMock.post).toHaveBeenCalledWith(
      '/subtitle/upload-srt',
      expect.any(FormData),
      expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } }),
    )
  })

  it('uploadSrt 非默认时不附加 is_default', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const file = new File(['x'], 'a.srt')
    await subtitleApi.uploadSrt(7, file, 'en', 'English', false)
    const fd = requestMock.post.mock.calls[0][1] as FormData
    expect(fd.has('is_default')).toBe(false)
  })

  it('deleteSubtitle DELETE /subtitle/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await subtitleApi.deleteSubtitle(5)
    expect(requestMock.delete).toHaveBeenCalledWith('/subtitle/5')
  })

  it('setDefaultSubtitle POST 透传 video_id 和 id', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await subtitleApi.setDefaultSubtitle(7, 'sub-1')
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/set-default', { video_id: 7, id: 'sub-1' })
  })
})

describe('subtitle api - 顶层函数', () => {
  it('getVideosWithSubtitleInfo GET /subtitle/videos', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getVideosWithSubtitleInfo()
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/videos')
  })

  it('getVideoSubtitles GET /subtitle/video/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getVideoSubtitles(7)
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/video/7')
  })

  it('uploadSubtitle 函数 POST multipart zh-CN 默认 language_name=中文', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const file = new File(['x'], 'a.srt')
    await uploadSubtitle(7, file, 'zh-CN', true)
    expect(requestMock.post).toHaveBeenCalledWith(
      '/subtitle/upload-srt',
      expect.any(FormData),
      expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } }),
    )
    const fd = requestMock.post.mock.calls[0][1] as FormData
    expect(fd.get('language_name')).toBe('中文')
  })

  it('uploadSubtitle 非中文时 language_name = language', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    const file = new File(['x'], 'a.srt')
    await uploadSubtitle(7, file, 'en', false)
    const fd = requestMock.post.mock.calls[0][1] as FormData
    expect(fd.get('language_name')).toBe('en')
  })

  it('importSrtToMongo POST /subtitle/import-srt 透传 video_id 和 srt', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await importSrtToMongo(7, '1\n00:00:00,000 --> 00:00:01,000\nhi', 'zh-CN', true)
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/import-srt', { video_id: 7, srt: '1\n00:00:00,000 --> 00:00:01,000\nhi' })
  })

  it('setDefaultSubtitle 函数 POST /subtitle/:id/set-default?video_id=', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await setDefaultSubtitle(5, 7)
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/5/set-default?video_id=7')
  })

  it('deleteSubtitle 函数 DELETE /subtitle/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteSubtitle(5)
    expect(requestMock.delete).toHaveBeenCalledWith('/subtitle/5')
  })

  it('getPendingSubtitles GET /subtitle/pending', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getPendingSubtitles()
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/pending')
  })

  it('approveSubtitle POST /subtitle/:id/approve', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await approveSubtitle(3)
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/3/approve')
  })

  it('rejectSubtitle POST /subtitle/:id/reject 带 reason', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await rejectSubtitle(3, '内容不当')
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/3/reject', { reason: '内容不当' })
  })

  it('previewSubtitle GET /subtitle/:id/preview', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: '' })
    await previewSubtitle(3)
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/3/preview')
  })

  it('scanSystemSubtitles GET /subtitle/scan/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await scanSystemSubtitles(7)
    expect(requestMock.get).toHaveBeenCalledWith('/subtitle/scan/7')
  })

  it('importSystemSubtitle POST /subtitle/import-system 透传 video_id 和 srt', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await importSystemSubtitle(7, 'srt-content')
    expect(requestMock.post).toHaveBeenCalledWith('/subtitle/import-system', { video_id: 7, srt: 'srt-content' })
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(getPendingSubtitles()).rejects.toThrow('500')
  })
})
