import { describe, it, expect, beforeEach } from 'vitest'
import { safeStorage } from './safeStorage'

describe('safeStorage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('getItem returns null for nonexistent key', () => {
    expect(safeStorage.getItem('nonexistent')).toBeNull()
  })

  it('setItem then getItem retrieves the value', () => {
    safeStorage.setItem('testKey', 'testValue')
    expect(safeStorage.getItem('testKey')).toBe('testValue')
  })

  it('removeItem removes the value', () => {
    safeStorage.setItem('toDelete', 'value')
    expect(safeStorage.getItem('toDelete')).toBe('value')
    safeStorage.removeItem('toDelete')
    expect(safeStorage.getItem('toDelete')).toBeNull()
  })
})
