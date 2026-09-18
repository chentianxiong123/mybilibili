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
  subtitleApi,
  getVideosWithSubtitleInfo,
  getVideoSubtitles,
  uploadSubtitle as uploadSubtitleFn,
  importSrtToMongo,
  setDefaultSubtitle as setDefaultSubtitleFn,
  deleteSubtitle as deleteSubtitleFn,
  getPendingSubtitles,
  approveSubtitle,
  rejectSubtitle,
  previewSubtitle,
  scanSystemSubtitles,
  importSystemSubtitle,
} from './subtitle'

describe('subtitle api - subtitleApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getSubtitles', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await subtitleApi.getSubtitles(1)
    expect(mocks.apiGet).toHaveBeenCalledWith('/subtitle/video/1')
  })

  it('getSubtitle 携带 language', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await subtitleApi.getSubtitle(1, 'zh-CN')
    expect(mocks.apiGet).toHaveBeenCalledWith('/subtitle/video/1/zh-CN')
  })

  it('uploadSubtitle POST body', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await subtitleApi.uploadSubtitle({ videoId: 1, content: 'x' })
    expect(mocks.apiPost).toHaveBeenCalledWith('/subtitle/upload', { videoId: 1, content: 'x' })
  })

  it('uploadSrt 用 multipart 拼接所有字段', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 's.srt')
    await subtitleApi.uploadSrt(9, file, 'en', 'English', true)
    const [url, form, cfg] = mocks.apiPost.mock.calls[0]
    expect(url).toBe('/subtitle/upload-srt')
    expect(form.get('video_id')).toBe('9')
    expect(form.get('file')).toBe(file)
    expect(form.get('language')).toBe('en')
    expect(form.get('language_name')).toBe('English')
    expect(form.get('is_default')).toBe('true')
    expect(cfg).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } })
  })

  it('uploadSrt isDefault=false 不追加 is_default', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 's.srt')
    await subtitleApi.uploadSrt(9, file, 'en', 'English', false)
    const [, form] = mocks.apiPost.mock.calls[0]
    expect(form.get('is_default')).toBeNull()
  })

  it('deleteSubtitle', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await subtitleApi.deleteSubtitle(5)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/subtitle/5')
  })

  it('setDefaultSubtitle POST body', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await subtitleApi.setDefaultSubtitle(3, 'sub_id_1')
    expect(mocks.apiPost).toHaveBeenCalledWith('/subtitle/set-default', { video_id: 3, id: 'sub_id_1' })
  })
})

describe('subtitle api - 模块级函数', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getVideosWithSubtitleInfo / getVideoSubtitles', async () => {
    mocks.apiGet.mockResolvedValue({ code: 200, data: [] })
    await getVideosWithSubtitleInfo()
    await getVideoSubtitles(2)
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/subtitle/videos')
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/subtitle/video/2')
  })

  it('uploadSubtitle zh-CN 语言名用"中文"', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 's.srt')
    await uploadSubtitleFn(7, file, 'zh-CN', true)
    const [, form] = mocks.apiPost.mock.calls[0]
    expect(form.get('language_name')).toBe('中文')
    expect(form.get('is_default')).toBe('true')
  })

  it('uploadSubtitle 非中文保持原值且 isDefault=false 时不追加', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 's.srt')
    await uploadSubtitleFn(7, file, 'ja', false)
    const [, form] = mocks.apiPost.mock.calls[0]
    expect(form.get('language_name')).toBe('ja')
    expect(form.get('is_default')).toBeNull()
  })

  it('importSrtToMongo POST body', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await importSrtToMongo(1, '/a/b.srt', 'en', true)
    expect(mocks.apiPost).toHaveBeenCalledWith('/subtitle/import-srt', {
      video_id: 1,
      srt: '/a/b.srt',
    })
  })

  it('setDefaultSubtitle 模块版携带 query string', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await setDefaultSubtitleFn(5, 9)
    expect(mocks.apiPost).toHaveBeenCalledWith('/subtitle/5/set-default?video_id=9')
  })

  it('deleteSubtitle 模块版', async () => {
    mocks.apiDelete.mockResolvedValueOnce({ code: 200 })
    await deleteSubtitleFn(8)
    expect(mocks.apiDelete).toHaveBeenCalledWith('/subtitle/8')
  })

  it('getPendingSubtitles / approveSubtitle / rejectSubtitle / previewSubtitle', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: {} })
    await getPendingSubtitles()
    await approveSubtitle(1)
    await rejectSubtitle(2, 'bad')
    await previewSubtitle(3)
    expect(mocks.apiGet).toHaveBeenNthCalledWith(1, '/subtitle/pending')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(1, '/subtitle/1/approve')
    expect(mocks.apiPost).toHaveBeenNthCalledWith(2, '/subtitle/2/reject', { reason: 'bad' })
    expect(mocks.apiGet).toHaveBeenNthCalledWith(2, '/subtitle/3/preview')
  })

  it('scanSystemSubtitles / importSystemSubtitle', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await scanSystemSubtitles(11)
    await importSystemSubtitle(11, 'srt-content')
    expect(mocks.apiGet).toHaveBeenCalledWith('/subtitle/scan/11')
    expect(mocks.apiPost).toHaveBeenCalledWith('/subtitle/import-system', {
      video_id: 11,
      srt: 'srt-content',
    })
  })
})
