import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: vi.fn(),
    delete: vi.fn(),
  }),
}))

import { linkmicApi } from './linkmic'

describe('linkmic api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('applyLinkmic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.applyLinkmic(1)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/apply/1')
  })

  it('acceptLinkmic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.acceptLinkmic(2)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/accept/2')
  })

  it('rejectLinkmic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.rejectLinkmic(3)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/reject/3')
  })

  it('disconnectLinkmic', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.disconnectLinkmic(4)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/disconnect/4')
  })

  it('toggleAudio 携带 enabled 参数', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.toggleAudio(5, true)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/toggle-audio/5', null, { params: { enabled: true } })
  })

  it('toggleVideo 携带 enabled=false', async () => {
    mocks.apiPost.mockResolvedValueOnce({ code: 200 })
    await linkmicApi.toggleVideo(6, false)
    expect(mocks.apiPost).toHaveBeenCalledWith('/live/linkmic/toggle-video/6', null, { params: { enabled: false } })
  })

  it('getActiveLinkmics', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await linkmicApi.getActiveLinkmics(7)
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/linkmic/active/7')
  })

  it('getPendingApplications', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: [] })
    await linkmicApi.getPendingApplications(8)
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/linkmic/pending/8')
  })

  it('getQueuePosition', async () => {
    mocks.apiGet.mockResolvedValueOnce({ code: 200, data: { position: 1 } })
    await linkmicApi.getQueuePosition(9)
    expect(mocks.apiGet).toHaveBeenCalledWith('/live/linkmic/queue-position/9')
  })
})
