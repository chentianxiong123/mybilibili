import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('../client', () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() }
}))

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn().mockReturnValue(true)
}))

import api from '../client'
import { isLogin } from '../../utils/session'
import { followUser, checkFollow, likeManuscript, coinManuscript, collectManuscript, getInteractionStatus, shareManuscript } from '../interaction'

const mockApi = vi.mocked(api)
const mockIsLogin = vi.mocked(isLogin)

describe('interaction API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(true)
  })

  it('followUser (follow=true) → POST /follow/:id', async () => {
    mockApi.post.mockResolvedValue({})
    const result = await followUser(42, true)
    expect(mockApi.post).toHaveBeenCalledWith('/follow/42')
    expect(result.code).toBe('1')
  })

  it('followUser (follow=false) → DELETE /follow/:id', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    mockApi.delete.mockResolvedValue({})
    const result = await followUser(42, false)
    expect(result.code).toBe('1')
  })

  it('checkFollow → GET /follow/check/:id', async () => {
    mockApi.get.mockResolvedValue({ data: { followed: true } })
    const result = await checkFollow(42)
    expect(mockApi.get).toHaveBeenCalledWith('/follow/check/42')
    expect(result.code).toBe('1')
  })

  it('likeManuscript (liked=true) → POST', async () => {
    mockApi.post.mockResolvedValue({})
    const result = await likeManuscript(100, true)
    expect(mockApi.post).toHaveBeenCalledWith('/manuscript/100/like')
    expect(result.code).toBe('1')
  })

  it('likeManuscript (liked=false) → DELETE', async () => {
    mockApi.delete = vi.fn().mockResolvedValue({})
    const result = await likeManuscript(100, false)
    expect(result.code).toBe('1')
  })

  it('coinManuscript → POST with coinCount', async () => {
    mockApi.post.mockResolvedValue({})
    const result = await coinManuscript(100, 2)
    expect(mockApi.post).toHaveBeenCalledWith('/manuscript/100/coin?coinCount=2')
    expect(result.code).toBe('1')
  })

  it('collectManuscript (collected=true) → POST', async () => {
    mockApi.post.mockResolvedValue({})
    const result = await collectManuscript(100, true)
    expect(mockApi.post).toHaveBeenCalledWith('/manuscript/100/collect')
    expect(result.code).toBe('1')
  })

  it('shareManuscript → POST', async () => {
    mockApi.post.mockResolvedValue({})
    const result = await shareManuscript(100)
    expect(mockApi.post).toHaveBeenCalledWith('/manuscript/100/share', { channel: 'wap' })
    expect(result.code).toBe('1')
  })

  it('getInteractionStatus → GET /manuscript/:id/status', async () => {
    mockApi.get.mockResolvedValue({ data: { liked: false, collected: true } })
    const result = await getInteractionStatus(100)
    expect(mockApi.get).toHaveBeenCalledWith('/manuscript/100/status')
    expect(result.code).toBe('1')
  })

  it('getInteractionStatus 未登录 → 返回 code 0', async () => {
    mockIsLogin.mockReturnValue(false)
    const result = await getInteractionStatus(100)
    expect(result).toEqual({ code: '0', data: null })
  })
})
