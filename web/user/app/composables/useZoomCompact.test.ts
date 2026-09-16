import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, computed } from 'vue'
import { compactLevel, desktopCompact, useZoomCompact } from './useZoomCompact'

describe('useZoomCompact', () => {
  beforeEach(() => {
    compactLevel.value = 0
  })

  describe('compactLevel / desktopCompact 全局响应式状态', () => {
    it('默认 compactLevel 为 0', () => {
      expect(compactLevel.value).toBe(0)
    })

    it('默认 desktopCompact 为 false', () => {
      expect(desktopCompact.value).toBe(false)
    })

    it('compactLevel = 1 时 desktopCompact = true', () => {
      compactLevel.value = 1
      expect(desktopCompact.value).toBe(true)
    })

    it('compactLevel = 0 重置后 desktopCompact = false', () => {
      compactLevel.value = 1
      expect(desktopCompact.value).toBe(true)
      compactLevel.value = 0
      expect(desktopCompact.value).toBe(false)
    })
  })

  describe('useZoomCompact 返回值', () => {
    it('返回 compactLevel 和 desktopCompact 引用', () => {
      const r = useZoomCompact()
      expect(r.compactLevel).toBe(compactLevel)
      expect(r.desktopCompact).toBe(desktopCompact)
    })

    it('返回的 compactLevel 是同一个 ref（共享状态）', () => {
      const r1 = useZoomCompact()
      const r2 = useZoomCompact()
      r1.compactLevel.value = 2
      expect(r2.compactLevel.value).toBe(2)
      expect(r2.desktopCompact.value).toBe(true)
    })
  })
})