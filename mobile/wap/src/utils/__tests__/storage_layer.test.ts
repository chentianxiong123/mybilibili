import { describe, it, expect, beforeEach, vi } from 'vitest'
import storage from '../storage_layer'

describe('storage_layer', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('set/get string', () => {
    storage.set('test-str', 'hello')
    expect(storage.get('test-str')).toBe('hello')
  })

  it('set/get number', () => {
    storage.set('test-num', 42)
    expect(storage.get('test-num')).toBe(42)
  })

  it('set/get object', () => {
    const obj = { a: 1, b: 'two', c: [3, 4] }
    storage.set('test-obj', obj)
    expect(storage.get('test-obj')).toEqual(obj)
  })

  it('set/get boolean', () => {
    storage.set('test-bool', true)
    expect(storage.get('test-bool')).toBe(true)
  })

  it('null → 删除', () => {
    storage.set('test-del', 'exists')
    storage.set('test-del', null)
    expect(storage.get('test-del')).toBeNull()
  })

  it('get 不存在的 key → null', () => {
    expect(storage.get('nonexistent-key')).toBeNull()
  })

  it('remove', () => {
    storage.set('test-rm', 'val')
    storage.remove('test-rm')
    expect(storage.get('test-rm')).toBeNull()
  })

  it('键统一加 wap: 前缀', () => {
    storage.set('mykey', 'myval')
    expect(localStorage.getItem('wap:mykey')).toBe('myval')
  })

  it('bulkSet', () => {
    storage.bulkSet([
      { key: 'bulk1', value: 'a' },
      { key: 'bulk2', value: 123 },
      { key: 'bulk3', value: { x: true } }
    ])
    expect(storage.get('bulk1')).toBe('a')
    expect(storage.get('bulk2')).toBe(123)
    expect(storage.get('bulk3')).toEqual({ x: true })
  })

  it('multiGet', () => {
    storage.set('prefix:1', 'one')
    storage.set('prefix:2', 'two')
    storage.set('other:3', 'three')
    const result = storage.multiGet('prefix:')
    expect(result.length).toBeGreaterThanOrEqual(2)
    const keys = result.map(r => r.key)
    expect(keys).toContain('prefix:1')
    expect(keys).toContain('prefix:2')
    expect(keys).not.toContain('other:3')
  })

  it('clearAll 只删 wap: 前缀', () => {
    storage.set('keep1', 'a')
    localStorage.setItem('other_prefix', 'b')
    storage.clearAll()
    expect(storage.get('keep1')).toBeNull()
    expect(localStorage.getItem('other_prefix')).toBe('b')
  })

  it('序列化 JSON 对象往返正确', () => {
    const data = { name: '测试', list: [1, 2, 3], nested: { a: true } }
    storage.set('json-roundtrip', data)
    const got = storage.get<typeof data>('json-roundtrip')
    expect(got).toEqual(data)
  })
})
