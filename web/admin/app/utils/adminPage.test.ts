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

// ====== 补充：跳转/取消 行为 + 鉴权失败/异常路径 ======

describe('adminPage 补充 - cancel 跳转/取消分支', () => {
  it('confirm 抛字符串 cancel 不触发 error 提示', async () => {
    confirmMock.mockRejectedValueOnce('cancel')
    const action = vi.fn()
    const ok = await runConfirmedAction({ message: 'm', action })
    expect(ok).toBe(false)
    expect(action).not.toHaveBeenCalled()
    expect(errorMock).not.toHaveBeenCalled()
  })

  it('confirm 抛字符串 cancel 但提供 errorMessage 时也不触发 error', async () => {
    confirmMock.mockRejectedValueOnce('cancel')
    const action = vi.fn()
    const ok = await runConfirmedAction({ message: 'm', action, errorMessage: '不应触发' })
    expect(ok).toBe(false)
    expect(errorMock).not.toHaveBeenCalled()
  })

  it('action 抛异常时捕获并返回 false 调用 error', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockRejectedValue(new Error('鉴权失败'))
    const ok = await runConfirmedAction({
      message: 'm',
      action,
      errorMessage: '权限不足',
    })
    expect(ok).toBe(false)
    expect(errorMock).toHaveBeenCalledWith('权限不足')
  })

  it('confirm 抛非 Error 也走 error 分支', async () => {
    confirmMock.mockRejectedValueOnce(new TypeError('token expired'))
    const ok = await runConfirmedAction({ message: 'm', action: vi.fn(), errorMessage: '会话过期' })
    expect(ok).toBe(false)
    expect(errorMock).toHaveBeenCalledWith('会话过期')
  })

  it('onSuccess 抛异常时被外层 catch 捕获并返回 false', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockResolvedValue(undefined)
    const onSuccess = vi.fn().mockRejectedValue(new Error('refresh failed'))
    const ok = await runConfirmedAction({
      message: 'm',
      action,
      onSuccess,
      errorMessage: '后续失败',
    })
    expect(ok).toBe(false)
    expect(onSuccess).toHaveBeenCalledOnce()
    expect(errorMock).toHaveBeenCalledWith('后续失败')
  })

  it('successMessage 与 onSuccess 同时提供都执行', async () => {
    confirmMock.mockResolvedValueOnce(undefined)
    const action = vi.fn().mockResolvedValue(undefined)
    const onSuccess = vi.fn().mockResolvedValue(undefined)
    await runConfirmedAction({ message: 'm', action, successMessage: 'OK', onSuccess })
    expect(successMock).toHaveBeenCalledWith('OK')
    expect(onSuccess).toHaveBeenCalledOnce()
  })
})

describe('adminPage 补充 - normalizePagedResult 边界', () => {
  it('records 为 undefined 时返回空数组 total=0', () => {
    expect(normalizePagedResult({ records: undefined, total: 5 })).toEqual({ records: [], total: 5 })
  })

  it('records 为数字时返回空数组 total 走 records.length', () => {
    expect(normalizePagedResult({ records: 42 })).toEqual({ records: [], total: 0 })
  })

  it('空数组返回 records=[] total=0', () => {
    expect(normalizePagedResult([])).toEqual({ records: [], total: 0 })
  })

  it('total 缺失且 records 数组时 total=length', () => {
    expect(normalizePagedResult({ records: [1, 2, 3] })).toEqual({ records: [1, 2, 3], total: 3 })
  })

  it('total 为 0 字符串被 Number 转 0', () => {
    expect(normalizePagedResult({ records: [], total: '0' })).toEqual({ records: [], total: 0 })
  })
})

describe('adminPage 补充 - formatDateTime 边界', () => {
  it('空字符串返回 "-"', () => {
    expect(formatDateTime('')).toBe('-')
  })

  it('ISO 字符串转换为本地时间格式', () => {
    expect(formatDateTime('2024-01-02T03:04:05')).toBe('2024-01-02 03:04:05')
  })

  it('带 Z 的 UTC 字符串转换为本地时间格式', () => {
    const out = formatDateTime('2024-01-02T03:04:05Z')
    expect(out).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
  })

  it('无法解析的字符串原样返回', () => {
    expect(formatDateTime('2024-01-02T03:04:05T00:00:00')).toBe('2024-01-02T03:04:05T00:00:00')
  })

  it('数字时间戳走 Date 转换路径', () => {
    const out = formatDateTime(0 as any)
    expect(typeof out).toBe('string')
  })

  it('boolean 类型走 Date 转换不抛错', () => {
    expect(() => formatDateTime(true as any)).not.toThrow()
  })
})
