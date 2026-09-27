import { describe, it, expect } from 'vitest'
import { formatTenThousand, formatDuration, formatDate } from '../format'

describe('formatTenThousand', () => {
  it('0 → "0"', () => {
    expect(formatTenThousand(0)).toBe('0')
  })

  it('falsy → "0"', () => {
    expect(formatTenThousand(NaN)).toBe('0')
    expect(formatTenThousand(undefined as any)).toBe('0')
  })

  it('< 10000 → 原数字字符串', () => {
    expect(formatTenThousand(1)).toBe('1')
    expect(formatTenThousand(9999)).toBe('9999')
    expect(formatTenThousand(500)).toBe('500')
  })

  it('≥ 10000 → 保留一位小数 + 万', () => {
    expect(formatTenThousand(10000)).toBe('1.0万')
    expect(formatTenThousand(12345)).toBe('1.2万')
    expect(formatTenThousand(99999)).toBe('10.0万')
    expect(formatTenThousand(100000000)).toBe('10000.0万')
  })
})

describe('formatDuration', () => {
  it('0 秒 → "0:00"', () => {
    expect(formatDuration(0)).toBe('0:00')
  })

  it('秒数 → mm:ss', () => {
    expect(formatDuration(65)).toBe('1:05')
    expect(formatDuration(3600)).toBe('60:00')
    expect(formatDuration(59)).toBe('0:59')
    expect(formatDuration(1)).toBe('0:01')
  })
})

describe('formatDate', () => {
  it('≥ 30 天 → 月日格式', () => {
    const d = '2025-01-15 12:00:00'
    expect(formatDate(d)).toBe('1月15日')
  })

  it('几分钟前', () => {
    const now = Date.now()
    const minutesAgo = 5
    const ts = new Date(now - minutesAgo * 60000)
    const dateStr = `${ts.getFullYear()}-${String(ts.getMonth() + 1).padStart(2, '0')}-${String(ts.getDate()).padStart(2, '0')} ${String(ts.getHours()).padStart(2, '0')}:${String(ts.getMinutes()).padStart(2, '0')}:${String(ts.getSeconds()).padStart(2, '0')}`
    const result = formatDate(dateStr)
    expect(result).toBe(`${minutesAgo}分钟前`)
  })

  it('几小时前', () => {
    const now = Date.now()
    const hoursAgo = 3
    const ts = new Date(now - hoursAgo * 3600000)
    const dateStr = `${ts.getFullYear()}-${String(ts.getMonth() + 1).padStart(2, '0')}-${String(ts.getDate()).padStart(2, '0')} ${String(ts.getHours()).padStart(2, '0')}:${String(ts.getMinutes()).padStart(2, '0')}:${String(ts.getSeconds()).padStart(2, '0')}`
    const result = formatDate(dateStr)
    expect(result).toBe(`${hoursAgo}小时前`)
  })

  it('几天前', () => {
    const now = Date.now()
    const daysAgo = 7
    const ts = new Date(now - daysAgo * 86400000)
    const dateStr = `${ts.getFullYear()}-${String(ts.getMonth() + 1).padStart(2, '0')}-${String(ts.getDate()).padStart(2, '0')} ${String(ts.getHours()).padStart(2, '0')}:${String(ts.getMinutes()).padStart(2, '0')}:${String(ts.getSeconds()).padStart(2, '0')}`
    const result = formatDate(dateStr)
    expect(result).toBe(`${daysAgo}天前`)
  })
})
