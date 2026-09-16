import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => ({
  userApi: {
    getUserById: vi.fn(),
  },
}))

vi.mock('@/utils/auth.ts', () => ({
  getCurrentUserId: vi.fn(),
  getStoredUser: vi.fn(),
  setAuthSession: vi.fn(),
}))

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
  },
}))

import { userApi } from '@/api/client'
import { getCurrentUserId, getStoredUser, setAuthSession } from '@/utils/auth.ts'
import { ElMessage } from 'element-plus'
import { refreshUserWithNotify } from './expNotify'

describe('refreshUserWithNotify', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('未登录时直接返回，不调接口也不提示', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(null)
    await refreshUserWithNotify()
    expect(userApi.getUserById).not.toHaveBeenCalled()
    expect(ElMessage.success).not.toHaveBeenCalled()
    expect(setAuthSession).not.toHaveBeenCalled()
  })

  it('升级时弹出恭喜升级提示并更新会话', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(1)
    vi.mocked(getStoredUser).mockReturnValue({ id: 1, level: 1, experience: 50 })
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 200,
      data: { id: 1, level: 2, experience: 80 },
    })

    await refreshUserWithNotify()

    expect(userApi.getUserById).toHaveBeenCalledWith(1)
    expect(ElMessage.success).toHaveBeenCalledWith(expect.stringContaining('LV2'))
    expect(setAuthSession).toHaveBeenCalledWith({ user: { id: 1, level: 2, experience: 80 } })
  })

  it('未升级但经验增加时弹出获得经验提示', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(2)
    vi.mocked(getStoredUser).mockReturnValue({ id: 2, level: 3, experience: 100 })
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 200,
      data: { id: 2, level: 3, experience: 150 },
    })

    await refreshUserWithNotify()

    expect(ElMessage.success).toHaveBeenCalledTimes(1)
    expect(ElMessage.success).toHaveBeenCalledWith('获得经验 +50')
    expect(setAuthSession).toHaveBeenCalledWith({ user: { id: 2, level: 3, experience: 150 } })
  })

  it('等级和经验都无变化时不弹提示但仍更新会话', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(3)
    vi.mocked(getStoredUser).mockReturnValue({ id: 3, level: 4, experience: 200 })
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 200,
      data: { id: 3, level: 4, experience: 200 },
    })

    await refreshUserWithNotify()

    expect(ElMessage.success).not.toHaveBeenCalled()
    expect(setAuthSession).toHaveBeenCalledWith({ user: { id: 3, level: 4, experience: 200 } })
  })

  it('旧用户信息为空时若新等级>0 视为升级', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(4)
    vi.mocked(getStoredUser).mockReturnValue(null)
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 200,
      data: { id: 4, level: 1, experience: 10 },
    })

    await refreshUserWithNotify()

    expect(ElMessage.success).toHaveBeenCalledWith('🎉 恭喜升级到 LV1！')
    expect(setAuthSession).toHaveBeenCalledWith({ user: { id: 4, level: 1, experience: 10 } })
  })

  it('接口返回非 200 时静默返回，不弹提示', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(5)
    vi.mocked(getStoredUser).mockReturnValue({ id: 5, level: 1, experience: 0 })
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 500,
      data: null,
    })

    await refreshUserWithNotify()

    expect(ElMessage.success).not.toHaveBeenCalled()
    expect(setAuthSession).not.toHaveBeenCalled()
  })

  it('接口抛错时被 catch 吞掉，不弹提示', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(6)
    vi.mocked(getStoredUser).mockReturnValue({ id: 6, level: 1, experience: 0 })
    vi.mocked(userApi.getUserById).mockRejectedValue(new Error('网络错误'))

    await expect(refreshUserWithNotify()).resolves.toBeUndefined()
    expect(ElMessage.success).not.toHaveBeenCalled()
    expect(setAuthSession).not.toHaveBeenCalled()
  })

  it('新等级未传时按 0 处理，与老等级比较', async () => {
    vi.mocked(getCurrentUserId).mockReturnValue(7)
    vi.mocked(getStoredUser).mockReturnValue({ id: 7, level: 0, experience: 0 })
    vi.mocked(userApi.getUserById).mockResolvedValue({
      code: 200,
      data: { id: 7 },
    })

    await refreshUserWithNotify()

    expect(ElMessage.success).not.toHaveBeenCalled()
    expect(setAuthSession).toHaveBeenCalledWith({ user: { id: 7 } })
  })
})