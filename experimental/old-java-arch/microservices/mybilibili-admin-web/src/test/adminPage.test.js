import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDateTime, normalizePagedResult, runConfirmedAction } from '../utils/adminPage.js'

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

describe('adminPage utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('normalizes array and paged data', () => {
    expect(normalizePagedResult([{ id: 1 }, { id: 2 }])).toEqual({
      records: [{ id: 1 }, { id: 2 }],
      total: 2
    })
    expect(normalizePagedResult({ records: [{ id: 3 }], total: '5' })).toEqual({
      records: [{ id: 3 }],
      total: 5
    })
  })

  it('formats timestamps for display', () => {
    expect(formatDateTime('2026-05-31T08:09:00')).toBe('2026-05-31 08:09:00')
    expect(formatDateTime(null)).toBe('-')
  })

  it('runs confirmed actions and reports success', async () => {
    ElMessageBox.confirm.mockResolvedValueOnce()
    const action = vi.fn()
    const onSuccess = vi.fn()

    await expect(runConfirmedAction({
      message: '确认删除？',
      title: '审批',
      action,
      successMessage: '已完成',
      onSuccess
    })).resolves.toBe(true)

    expect(ElMessageBox.confirm).toHaveBeenCalledWith('确认删除？', '审批')
    expect(action).toHaveBeenCalled()
    expect(onSuccess).toHaveBeenCalled()
    expect(ElMessage.success).toHaveBeenCalledWith('已完成')
  })

  it('cancels without showing an error message', async () => {
    ElMessageBox.confirm.mockRejectedValueOnce('cancel')
    const action = vi.fn()

    await expect(runConfirmedAction({
      message: '确认删除？',
      action,
      errorMessage: '失败'
    })).resolves.toBe(false)

    expect(action).not.toHaveBeenCalled()
    expect(ElMessage.error).not.toHaveBeenCalled()
  })

  it('shows an error when confirmation fails unexpectedly', async () => {
    ElMessageBox.confirm.mockRejectedValueOnce(new Error('boom'))

    await expect(runConfirmedAction({
      message: '确认删除？',
      action: vi.fn(),
      errorMessage: '失败'
    })).resolves.toBe(false)

    expect(ElMessage.error).toHaveBeenCalledWith('失败')
  })
})
