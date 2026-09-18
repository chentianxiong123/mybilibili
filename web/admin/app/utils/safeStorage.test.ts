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

// ====== 补充：多 key 隔离 / 损坏数据 / 边界 ======

describe('safeStorage 补充 - 多 key 隔离', () => {
  beforeEach(() => { store.clear() })

  it('多个 key 互相独立读取', () => {
    safeStorage.setItem('user', 'alice')
    safeStorage.setItem('role', 'admin')
    safeStorage.setItem('count', '5')
    expect(safeStorage.getItem('user')).toBe('alice')
    expect(safeStorage.getItem('role')).toBe('admin')
    expect(safeStorage.getItem('count')).toBe('5')
    expect(safeStorage.getItem('nonexistent')).toBeNull()
  })

  it('删除单个 key 不影响其他 key', () => {
    safeStorage.setItem('a', '1')
    safeStorage.setItem('b', '2')
    safeStorage.setItem('c', '3')
    safeStorage.removeItem('b')
    expect(safeStorage.getItem('a')).toBe('1')
    expect(safeStorage.getItem('b')).toBeNull()
    expect(safeStorage.getItem('c')).toBe('3')
  })

  it('顺序覆写与最新值一致', () => {
    for (let i = 0; i < 10; i++) {
      safeStorage.setItem('counter', String(i))
    }
    expect(safeStorage.getItem('counter')).toBe('9')
  })

  it('长字符串值能完整存读', () => {
    const long = 'x'.repeat(5000)
    safeStorage.setItem('big', long)
    expect(safeStorage.getItem('big')).toBe(long)
  })
})

describe('safeStorage 补充 - 损坏/边界数据', () => {
  beforeEach(() => { store.clear() })

  it('存储 null 字符串能正确读回', () => {
    safeStorage.setItem('k', 'null')
    expect(safeStorage.getItem('k')).toBe('null')
  })

  it('存储 undefined 字符串能正确读回', () => {
    safeStorage.setItem('k', 'undefined')
    expect(safeStorage.getItem('k')).toBe('undefined')
  })

  it('存储带换行的多行字符串保留原文', () => {
    const multi = 'line1\nline2\r\nline3\twith tab'
    safeStorage.setItem('multi', multi)
    expect(safeStorage.getItem('multi')).toBe(multi)
  })

  it('存储 unicode/emoji 字符串不损坏', () => {
    const emoji = '用户名🚀emoji 中文'
    safeStorage.setItem('u', emoji)
    expect(safeStorage.getItem('u')).toBe(emoji)
  })

  it('数字类型值在 setItem 调用时被强转字符串', () => {
    safeStorage.setItem('n' as any, 123 as any)
    expect(safeStorage.getItem('n')).toBe('123')
  })

  it('读损坏 JSON 时不抛异常透传原文', () => {
    safeStorage.setItem('bad', '{user: "x"')
    expect(() => safeStorage.getItem('bad')).not.toThrow()
    expect(safeStorage.getItem('bad')).toBe('{user: "x"')
  })
})
