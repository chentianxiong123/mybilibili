import { describe, it, expect, beforeEach, vi } from 'vitest'

// happy-dom 18 在某些环境下没有把 localStorage 挂到 window/globalThis，
// 这里手动 polyfill（与 auth.test.ts 保持一致）。
const store = new Map<string, string>()
const localStorageMock: Storage = {
  getItem: (k: string) => (store.has(k) ? store.get(k)! : null),
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
  key: (i: number) => Array.from(store.keys())[i] ?? null,
  get length() { return store.size }
} as Storage

if (typeof window !== 'undefined') (window as any).localStorage = localStorageMock
;(globalThis as any).localStorage = localStorageMock

import { safeStorage } from './safeStorage'

describe('safeStorage - client mode', () => {
  beforeEach(() => {
    store.clear()
  })

  it('getItem returns null for nonexistent key', () => {
    expect(safeStorage.getItem('nonexistent')).toBeNull()
  })

  it('setItem then getItem retrieves the value', () => {
    safeStorage.setItem('testKey', 'testValue')
    expect(safeStorage.getItem('testKey')).toBe('testValue')
  })

  it('setItem overwrites previous value', () => {
    safeStorage.setItem('k', 'v1')
    safeStorage.setItem('k', 'v2')
    expect(safeStorage.getItem('k')).toBe('v2')
  })

  it('removeItem removes the value', () => {
    safeStorage.setItem('toDelete', 'value')
    expect(safeStorage.getItem('toDelete')).toBe('value')
    safeStorage.removeItem('toDelete')
    expect(safeStorage.getItem('toDelete')).toBeNull()
  })

  it('removeItem on nonexistent key is a no-op', () => {
    expect(() => safeStorage.removeItem('missing')).not.toThrow()
  })

  it('preserves JSON strings verbatim (serialization is the caller’s responsibility)', () => {
    const json = JSON.stringify({ a: 1, b: [2, 3] })
    safeStorage.setItem('json', json)
    expect(safeStorage.getItem('json')).toBe(json)
    expect(JSON.parse(safeStorage.getItem('json')!)).toEqual({ a: 1, b: [2, 3] })
  })

  it('returns raw value for corrupted JSON strings (does not throw)', () => {
    safeStorage.setItem('corrupt', '{not json')
    expect(safeStorage.getItem('corrupt')).toBe('{not json')
  })

  it('stores empty string values', () => {
    safeStorage.setItem('empty', '')
    expect(safeStorage.getItem('empty')).toBe('')
  })
})

describe('safeStorage - server mode (no window)', () => {
  it('getItem returns null when window is undefined', async () => {
    const originalWindow = (globalThis as any).window
    try {
      ;(globalThis as any).window = undefined
      vi.resetModules()
      const mod = await import('./safeStorage?server')
      expect(mod.safeStorage.getItem('any')).toBeNull()
    } finally {
      ;(globalThis as any).window = originalWindow
      vi.resetModules()
    }
  })

  it('setItem / removeItem are no-ops when window is undefined', async () => {
    const originalWindow = (globalThis as any).window
    try {
      ;(globalThis as any).window = undefined
      vi.resetModules()
      const mod = await import('./safeStorage?server2')
      expect(() => mod.safeStorage.setItem('k', 'v')).not.toThrow()
      expect(() => mod.safeStorage.removeItem('k')).not.toThrow()
      expect(mod.safeStorage.getItem('k')).toBeNull()
    } finally {
      ;(globalThis as any).window = originalWindow
      vi.resetModules()
    }
  })
})
