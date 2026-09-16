import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('element-plus', () => {
  const confirm = vi.fn()
  const success = vi.fn()
  const error = vi.fn()
  return {
    ElMessageBox: { confirm },
    ElMessage: { success, error }
  }
})

import { ElMessage, ElMessageBox } from 'element-plus'
import { normalizePagedResult, formatDateTime, runConfirmedAction } from './adminPage'

const confirmMock = ElMessageBox.confirm as unknown as ReturnType<typeof vi.fn>
const successMock = ElMessage.success as unknown as ReturnType<typeof vi.fn>
const errorMock = ElMessage.error as unknown as ReturnType<typeof vi.fn>

beforeEach(() => {
  confirmMock.mockReset()
  successMock.mockReset()
  errorMock.mockReset()
})

describe('normalizePagedResult', () => {
  it('returns array as records with total=length', () => {
    const arr = [{ id: 1 }, { id: 2 }, { id: 3 }]
    expect(normalizePagedResult(arr)).toEqual({ records: arr, total: 3 })
  })

  it('returns records/total for {records,total} payload', () => {
    const records = [{ id: 1 }]
    expect(normalizePagedResult({ records, total: 100 })).toEqual({ records, total: 100 })
  })

  it('coerces total string to Number', () => {
    expect(normalizePagedResult({ records: [{ id: 1 }], total: '42' })).toEqual({
      records: [{ id: 1 }],
      total: 42
    })
  })

  it('falls back to records.length when total missing', () => {
    expect(normalizePagedResult({ records: [{ id: 1 }, { id: 2 }] })).toEqual({
      records: [{ id: 1 }, { id: 2 }],
      total: 2
    })
  })

  it('returns empty records for non-array data', () => {
    expect(normalizePagedResult(null)).toEqual({ records: [], total: 0 })
    expect(normalizePagedResult(undefined)).toEqual({ records: [], total: 0 })
    expect(normalizePagedResult({})).toEqual({ records: [], total: 0 })
  })

  it('returns empty records when records is not array', () => {
    expect(normalizePagedResult({ records: 'nope', total: 5 })).toEqual({
      records: [],
      total: 5
    })
  })
})

describe('formatDateTime', () => {
  it('returns "-" for empty/null/undefined/0', () => {
    expect(formatDateTime('')).toBe('-')
    expect(formatDateTime(null)).toBe('-')
    expect(formatDateTime(undefined)).toBe('-')
    expect(formatDateTime(0)).toBe('-')
  })

  it('replaces T with space for strings', () => {
    expect(formatDateTime('2024-01-02T03:04:05')).toBe('2024-01-02 03:04:05')
    expect(formatDateTime('2024-01-02')).toBe('2024-01-02')
  })

  it('formats non-string values via toLocaleString', () => {
    const out = formatDateTime(new Date('2024-01-02T03:04:05'))
    expect(typeof out).toBe('string')
    expect(out.length).toBeGreaterThan(0)
  })
})

describe('runConfirmedAction', () => {
  it('runs action and shows success message on confirm', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockResolvedValue(undefined)
    const onSuccess = vi.fn().mockResolvedValue(undefined)

    const ok = await runConfirmedAction({
      message: 'sure?',
      action,
      successMessage: 'done',
      onSuccess
    })

    expect(ok).toBe(true)
    expect(confirmMock).toHaveBeenCalledWith('sure?', '提示')
    expect(action).toHaveBeenCalledOnce()
    expect(successMock).toHaveBeenCalledWith('done')
    expect(onSuccess).toHaveBeenCalledOnce()
    expect(errorMock).not.toHaveBeenCalled()
  })

  it('uses custom title when provided', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockResolvedValue(undefined)

    await runConfirmedAction({
      message: 'm',
      title: '自定义标题',
      action
    })

    expect(confirmMock).toHaveBeenCalledWith('m', '自定义标题')
  })

  it('returns true without success message when not provided', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockResolvedValue(undefined)

    const ok = await runConfirmedAction({ message: 'm', action })
    expect(ok).toBe(true)
    expect(successMock).not.toHaveBeenCalled()
  })

  it('returns false on cancel and does not run action', async () => {
    confirmMock.mockRejectedValueOnce('cancel')
    const action = vi.fn()

    const ok = await runConfirmedAction({ message: 'm', action })
    expect(ok).toBe(false)
    expect(action).not.toHaveBeenCalled()
    expect(errorMock).not.toHaveBeenCalled()
  })

  it('shows error message on non-cancel failure', async () => {
    confirmMock.mockRejectedValueOnce(new Error('boom'))
    const action = vi.fn()

    const ok = await runConfirmedAction({
      message: 'm',
      action,
      errorMessage: '自定义失败'
    })
    expect(ok).toBe(false)
    expect(errorMock).toHaveBeenCalledWith('自定义失败')
  })

  it('uses default error message "操作失败"', async () => {
    confirmMock.mockRejectedValueOnce(new Error('boom'))
    const ok = await runConfirmedAction({ message: 'm', action: vi.fn() })
    expect(ok).toBe(false)
    expect(errorMock).toHaveBeenCalledWith('操作失败')
  })
})
