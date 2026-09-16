import { describe, it, expect, vi, beforeEach } from 'vitest'
import { toDurationSeconds, formatDuration, formatMonthDay, normalizeVideoCard } from './videoCard'

describe('toDurationSeconds', () => {
  it('returns 0 for null', () => {
    expect(toDurationSeconds(null)).toBe(0)
  })

  it('returns 0 for undefined', () => {
    expect(toDurationSeconds(undefined)).toBe(0)
  })

  it('returns 0 for empty string', () => {
    expect(toDurationSeconds('')).toBe(0)
  })

  it('returns the number directly', () => {
    expect(toDurationSeconds(120)).toBe(120)
    expect(toDurationSeconds(0)).toBe(0)
  })

  it('converts numeric string', () => {
    expect(toDurationSeconds('3600')).toBe(3600)
    expect(toDurationSeconds('0')).toBe(0)
  })

  it('parses mm:ss format', () => {
    expect(toDurationSeconds('1:30')).toBe(90)
    expect(toDurationSeconds('10:00')).toBe(600)
    expect(toDurationSeconds('0:00')).toBe(0)
  })

  it('parses hh:mm:ss format', () => {
    expect(toDurationSeconds('1:00:00')).toBe(3600)
    expect(toDurationSeconds('2:30:15')).toBe(9015)
    expect(toDurationSeconds('0:01:30')).toBe(90)
  })

  it('returns 0 for invalid string', () => {
    expect(toDurationSeconds('abc')).toBe(0)
    expect(toDurationSeconds('1:2:3:4')).toBe(0)
  })
})

describe('formatDuration', () => {
  it('returns 00:00 for 0', () => {
    expect(formatDuration(0)).toBe('00:00')
  })

  it('formats seconds less than 1 hour', () => {
    expect(formatDuration(65)).toBe('01:05')
    expect(formatDuration(3599)).toBe('59:59')
    expect(formatDuration(1)).toBe('00:01')
  })

  it('formats seconds >= 1 hour', () => {
    expect(formatDuration(3600)).toBe('1:00:00')
    expect(formatDuration(9015)).toBe('2:30:15')
    expect(formatDuration(3661)).toBe('1:01:01')
  })

  it('handles edge values', () => {
    expect(formatDuration(59)).toBe('00:59')
    expect(formatDuration(60)).toBe('01:00')
  })
})

describe('formatMonthDay', () => {
  it('returns empty string for empty value', () => {
    expect(formatMonthDay('')).toBe('')
    expect(formatMonthDay(null)).toBe('')
    expect(formatMonthDay(undefined)).toBe('')
  })

  it('returns empty string for invalid date', () => {
    expect(formatMonthDay('not-a-date')).toBe('')
  })

  it('returns MM-DD for valid date', () => {
    expect(formatMonthDay('2025-01-15')).toBe('01-15')
    expect(formatMonthDay('2025-12-31')).toBe('12-31')
  })
})

describe('normalizeVideoCard', () => {
  it('returns null for null input', () => {
    expect(normalizeVideoCard(null)).toBeNull()
  })

  it('returns null when no manuscriptId', () => {
    expect(normalizeVideoCard({ title: 'test' })).toBeNull()
  })

  it('normalizes a valid object', () => {
    const item = {
      manuscriptId: 123,
      title: 'Test Video',
      viewCount: 1000,
      commentCount: 50,
      uploadTime: '2025-06-01',
      coverUrl: 'http://example.com/cover.jpg',
      duration: '3:45',
    }
    const result = normalizeVideoCard(item)
    expect(result).not.toBeNull()
    expect(result.manuscriptId).toBe(123)
    expect(result.title).toBe('Test Video')
    expect(result.viewCount).toBe(1000)
    expect(result.commentCount).toBe(50)
    expect(result.coverUrl).toBe('http://example.com/cover.jpg')
  })

  it('falls back to snake_case field names', () => {
    const item = {
      manuscript_id: 456,
      view_count: 200,
      comment_count: 10,
      user_name: 'TestUser',
      user_avatar: 'avatar.jpg',
      cover_url: 'cover.jpg',
    }
    const result = normalizeVideoCard(item)
    expect(result).not.toBeNull()
    expect(result.manuscriptId).toBe(456)
    expect(result.viewCount).toBe(200)
    expect(result.commentCount).toBe(10)
    expect(result.uploader.name).toBe('TestUser')
    expect(result.uploader.avatar).toBe('avatar.jpg')
  })

  it('uses uploader object directly if provided', () => {
    const item = {
      manuscriptId: 789,
      uploader: { id: 1, name: 'UP主', avatar: 'a.png' },
    }
    const result = normalizeVideoCard(item)
    expect(result.uploader.name).toBe('UP主')
  })
})
