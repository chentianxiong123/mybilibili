import { describe, expect, it } from 'vitest'
import { sameUserId } from '../utils/userId.js'

describe('sameUserId', () => {
  it('matches numeric and string forms of the same user id', () => {
    expect(sameUserId(1, '1')).toBe(true)
    expect(sameUserId('42', 42)).toBe(true)
  })

  it('does not match missing or different user ids', () => {
    expect(sameUserId(null, null)).toBe(false)
    expect(sameUserId(undefined, 1)).toBe(false)
    expect(sameUserId(1, 2)).toBe(false)
  })
})
