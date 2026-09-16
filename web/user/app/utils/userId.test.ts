import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { sameUserId } from './userId'

describe('sameUserId', () => {
  beforeEach(() => {
    vi.useRealTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('数字相同返回 true', () => {
    expect(sameUserId(1, 1)).toBe(true)
  })

  it('数字与字符串字面量相同返回 true', () => {
    expect(sameUserId(1, '1')).toBe(true)
    expect(sameUserId('1', 1)).toBe(true)
  })

  it('字符串与字符串相同返回 true', () => {
    expect(sameUserId('42', '42')).toBe(true)
  })

  it('不同 id 返回 false', () => {
    expect(sameUserId(1, 2)).toBe(false)
    expect(sameUserId('1', '2')).toBe(false)
  })

  it('left 为 null 时返回 false', () => {
    expect(sameUserId(null, 1)).toBe(false)
    expect(sameUserId(null, '1')).toBe(false)
    expect(sameUserId(null, null)).toBe(false)
  })

  it('right 为 null 时返回 false', () => {
    expect(sameUserId(1, null)).toBe(false)
    expect(sameUserId('1', null)).toBe(false)
  })

  it('left 为 undefined 时返回 false', () => {
    expect(sameUserId(undefined, 1)).toBe(false)
    expect(sameUserId(undefined, '1')).toBe(false)
  })

  it('right 为 undefined 时返回 false', () => {
    expect(sameUserId(1, undefined)).toBe(false)
  })

  it('两边都 undefined 返回 false', () => {
    expect(sameUserId(undefined, undefined)).toBe(false)
  })

  it('空字符串与空字符串相等', () => {
    expect(sameUserId('', '')).toBe(true)
  })

  it('0 与 "0" 视为相同', () => {
    expect(sameUserId(0, '0')).toBe(true)
  })

  it('0 与 1 不相同', () => {
    expect(sameUserId(0, 1)).toBe(false)
  })

  it('大数字字符串保持一致', () => {
    expect(sameUserId('9999999999', 9999999999)).toBe(true)
  })
})