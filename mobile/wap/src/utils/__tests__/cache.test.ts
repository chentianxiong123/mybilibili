import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { readCache, writeCache, swr, invalidateCache } from '../cache'
import storage from '../storage_layer'

describe('cache', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    localStorage.clear()
  })

  describe('writeCache / readCache', () => {
    it('写入后可读取', () => {
      writeCache('list', 'test-key', { items: [1, 2, 3] })
      const result = readCache('list', 'test-key')
      expect(result).toEqual({ items: [1, 2, 3] })
    })

    it('不存在的 key → null', () => {
      expect(readCache('list', 'nonexistent')).toBeNull()
    })

    it('过期后返回 null（list 默认 TTL 5 分钟）', () => {
      writeCache('list', 'expire-test', 'data')
      vi.advanceTimersByTime(5 * 60 * 1000 + 1)
      expect(readCache('list', 'expire-test')).toBeNull()
    })

    it('未过期仍可读', () => {
      writeCache('list', 'fresh-test', 'data')
      vi.advanceTimersByTime(4 * 60 * 1000)
      expect(readCache('list', 'fresh-test')).toBe('data')
    })

    it('detail TTL 30 分钟', () => {
      writeCache('detail', 'd1', 'info')
      vi.advanceTimersByTime(30 * 60 * 1000 - 1)
      expect(readCache('detail', 'd1')).toBe('info')
      vi.advanceTimersByTime(2)
      expect(readCache('detail', 'd1')).toBeNull()
    })

    it('覆盖写入', () => {
      writeCache('list', 'ow', 'old')
      writeCache('list', 'ow', 'new')
      expect(readCache('list', 'ow')).toBe('new')
    })
  })

  describe('invalidateCache', () => {
    it('删除后读不到', () => {
      writeCache('list', 'inv', 'data')
      expect(readCache('list', 'inv')).toBe('data')
      invalidateCache('list', 'inv')
      expect(readCache('list', 'inv')).toBeNull()
    })

    it('默认 key 空串', () => {
      writeCache('list', '', 'data')
      invalidateCache('list')
      expect(readCache('list', '')).toBeNull()
    })
  })

  describe('swr', () => {
    it('网络成功 → local 为 null，fresh 有值', async () => {
      const fetcher = vi.fn().mockResolvedValue({ id: 1 })
      const result = await swr('list', 'swr1', fetcher)
      expect(result.local).toBeNull()
      expect(result.fresh).toEqual({ id: 1 })
      expect(fetcher).toHaveBeenCalledOnce()
    })

    it('有本地缓存 → local 有值', () => {
      writeCache('list', 'swr2', 'cached')
      const fetcher = vi.fn().mockResolvedValue('fresh')
      return swr('list', 'swr2', fetcher).then(result => {
        expect(result.local).toBe('cached')
        expect(result.fresh).toBe('fresh')
      })
    })

    it('网络失败 → local 保留，fresh 为 null', async () => {
      writeCache('list', 'swr3', 'offline-data')
      const fetcher = vi.fn().mockRejectedValue(new Error('network'))
      const result = await swr('list', 'swr3', fetcher)
      expect(result.local).toBe('offline-data')
      expect(result.fresh).toBeNull()
    })

    it('无缓存 + 网络失败 → 都为 null', async () => {
      const fetcher = vi.fn().mockRejectedValue(new Error('network'))
      const result = await swr('list', 'swr4', fetcher)
      expect(result.local).toBeNull()
      expect(result.fresh).toBeNull()
    })
  })
})
