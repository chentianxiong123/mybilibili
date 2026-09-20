import { describe, it, expect } from 'vitest'
import {
  linkify,
  emojiText,
  highlightKeyword,
  handleTime,
  returnSecond,
  handleNum,
  handleDate,
  handleDateTime,
  handleDateTime2,
  handleDateTime3,
  getNicknameLength,
  handleLevel,
  generateUUID,
} from '@/teriteri-src/utils/utils'

describe('linkify', () => {
  it('returns null/undefined as-is', () => {
    expect(linkify(null)).toBeNull()
    expect(linkify(undefined)).toBeUndefined()
    expect(linkify('')).toBe('')
  })

  it('wraps URLs in anchor tags', () => {
    const result = linkify('visit https://example.com today')
    expect(result).toContain('<a href="https://example.com"')
    expect(result).toContain('target="_blank"')
    expect(result).toContain('https://example.com</a>')
  })

  it('does not alter text without URLs', () => {
    expect(linkify('no links here')).toBe('no links here')
  })

  it('handles multiple URLs', () => {
    const result = linkify('a https://a.com and https://b.com')
    expect(result).toContain('https://a.com')
    expect(result).toContain('https://b.com')
  })
})

describe('highlightKeyword', () => {
  it('highlights keyword characters', () => {
    const result = highlightKeyword('abc', 'xabcy')
    expect(result).toContain('<em')
    expect(result).toContain('a</em>')
    expect(result).toContain('b</em>')
    expect(result).toContain('c</em>')
  })

  it('is case-insensitive', () => {
    const result = highlightKeyword('ab', 'xAB')
    expect(result).toContain('<em')
  })
})

describe('handleTime', () => {
  it('formats seconds to mm:ss', () => {
    expect(handleTime(0)).toBe('00:00')
    expect(handleTime(60)).toBe('01:00')
    expect(handleTime(90)).toBe('01:30')
    expect(handleTime(3661)).toBe('61:01')
  })

  it('pads single digits', () => {
    expect(handleTime(5)).toBe('00:05')
    expect(handleTime(65)).toBe('01:05')
  })

  it('handles string input', () => {
    expect(handleTime('120')).toBe('02:00')
  })
})

describe('returnSecond', () => {
  it('converts mm:ss back to seconds', () => {
    expect(returnSecond('00:00')).toBe(0)
    expect(returnSecond('01:00')).toBe(60)
    expect(returnSecond('01:30')).toBe(90)
  })
})

describe('handleNum', () => {
  it('returns numbers <= 10000 as-is', () => {
    expect(handleNum(0)).toBe(0)
    expect(handleNum(9999)).toBe(9999)
    expect(handleNum(10000)).toBe(10000)
  })

  it('formats large numbers with 万', () => {
    expect(handleNum(10001)).toBe('1.0万')
    expect(handleNum(198765)).toBe('19.9万')
    expect(handleNum(100000000)).toBe('10000.0万')
  })
})

describe('handleDate', () => {
  it('returns X分钟前 for < 1 hour', () => {
    const now = new Date()
    const thirtyMinAgo = new Date(now.getTime() - 30 * 60 * 1000)
    const result = handleDate(thirtyMinAgo)
    expect(result).toMatch(/\d+分钟前/)
  })

  it('returns X小时前 for 1-24 hours', () => {
    const now = new Date()
    const fiveHoursAgo = new Date(now.getTime() - 5 * 60 * 60 * 1000)
    const result = handleDate(fiveHoursAgo)
    expect(result).toBe('5小时前')
  })

  it('returns M-D for same year, older than 24h', () => {
    const now = new Date()
    const threeDaysAgo = new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000)
    const result = handleDate(threeDaysAgo)
    expect(result).toMatch(/^\d+-\d+$/)
  })

  it('returns YYYY-M-D for different year', () => {
    const result = handleDate(new Date('2023-06-15'))
    expect(result).toMatch(/^2023-\d+-\d+$/)
  })
})

describe('handleDateTime', () => {
  it('returns 今天 HH:MM for today', () => {
    const result = handleDateTime(new Date())
    expect(result).toMatch(/^今天 \d{2}:\d{2}$/)
  })

  it('returns 未知时间 for invalid date', () => {
    expect(handleDateTime('invalid')).toBe('未知时间')
  })
})

describe('handleDateTime2', () => {
  it('formats to M-D HH:MM', () => {
    const result = handleDateTime2(new Date('2025-03-15T14:30:00'))
    expect(result).toBe('03-15 14:30')
  })

  it('returns 未知时间 for invalid date', () => {
    expect(handleDateTime2('nope')).toBe('未知时间')
  })
})

describe('handleDateTime3', () => {
  it('returns 刚刚 for < 30s', () => {
    const result = handleDateTime3(new Date())
    expect(result).toBe('刚刚')
  })

  it('returns X分钟前 for 1-60 min', () => {
    const now = new Date()
    const tenMinAgo = new Date(now.getTime() - 10 * 60 * 1000)
    const result = handleDateTime3(tenMinAgo)
    expect(result).toBe('10分钟前')
  })

  it('returns X小时前 for 1-24 hours', () => {
    const now = new Date()
    const threeHoursAgo = new Date(now.getTime() - 3 * 60 * 60 * 1000)
    const result = handleDateTime3(threeHoursAgo)
    expect(result).toBe('3小时前')
  })

  it('returns YYYY-MM-DD HH:MM for older', () => {
    const result = handleDateTime3(new Date('2024-01-15T08:30:00'))
    expect(result).toBe('2024-01-15 08:30')
  })

  it('returns 未知时间 for invalid', () => {
    expect(handleDateTime3('garbage')).toBe('未知时间')
  })
})

describe('getNicknameLength', () => {
  it('counts ASCII as 1', () => {
    expect(getNicknameLength('abc')).toBe(3)
  })

  it('counts CJK as 2', () => {
    expect(getNicknameLength('你好')).toBe(4)
  })

  it('counts mixed correctly', () => {
    expect(getNicknameLength('a你b')).toBe(4) // 1+2+1
  })
})

describe('handleLevel', () => {
  it('returns 0 for exp < 50', () => {
    expect(handleLevel(0)).toBe(0)
    expect(handleLevel(49)).toBe(0)
  })

  it('returns level based on thresholds', () => {
    expect(handleLevel(50)).toBe(1)
    expect(handleLevel(200)).toBe(2)
    expect(handleLevel(1500)).toBe(3)
    expect(handleLevel(4500)).toBe(4)
    expect(handleLevel(10800)).toBe(5)
    expect(handleLevel(28800)).toBe(6)
  })
})

describe('generateUUID', () => {
  it('returns a UUID-like string', () => {
    const uuid = generateUUID()
    expect(uuid).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
  })

  it('generates unique values', () => {
    const a = generateUUID()
    const b = generateUUID()
    expect(a).not.toBe(b)
  })
})

describe('emojiText', () => {
  it('returns null/undefined as-is', () => {
    expect(emojiText(null)).toBeNull()
    expect(emojiText(undefined)).toBeUndefined()
  })

  it('returns text without emoji markers unchanged', () => {
    expect(emojiText('hello')).toBe('hello')
  })
})
