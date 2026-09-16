import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getCurrentUser } from './currentUser'

describe('getCurrentUser', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  it('returns user from localStorage when available', () => {
    localStorage.setItem('user', JSON.stringify({ id: 42, username: 'Alice' }))
    const user = getCurrentUser()
    expect(user.id).toBe(42)
    expect(user.name).toBe('Alice')
  })

  it('falls back when no user in localStorage and temporary=false', () => {
    const user = getCurrentUser({ fallbackId: 1, fallbackName: 'Guest' })
    expect(user.id).toBe(1)
    expect(user.name).toBe('Guest')
  })

  it('returns temporary user when no user and temporary=true', () => {
    const user = getCurrentUser({ temporary: true })
    expect(user.id).toBeGreaterThan(0)
    expect(user.name).toMatch(/^访客\d+$/)
  })

  it('returns fallback when localStorage JSON is invalid', () => {
    localStorage.setItem('user', 'invalid-json{{{')
    const user = getCurrentUser({ fallbackId: 99, fallbackName: 'Fallback' })
    expect(user.id).toBe(99)
    expect(user.name).toBe('Fallback')
  })
})
