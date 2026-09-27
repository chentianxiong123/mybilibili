import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { getWapTheme, applyWapTheme, toggleWapTheme, initWapTheme } from '../theme'

describe('theme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.dataset.wapTheme = ''
  })

  afterEach(() => {
    localStorage.clear()
    document.documentElement.dataset.wapTheme = ''
  })

  describe('getWapTheme', () => {
    it('无设置 → 默认 "light"', () => {
      expect(getWapTheme()).toBe('light')
    })

    it('读取 dark', () => {
      localStorage.setItem('wap:theme', 'dark')
      expect(getWapTheme()).toBe('dark')
    })

    it('读取 light', () => {
      localStorage.setItem('wap:theme', 'light')
      expect(getWapTheme()).toBe('light')
    })

    it('非法值 → 默认 "light"', () => {
      localStorage.setItem('wap:theme', 'blue')
      expect(getWapTheme()).toBe('light')
    })

    it('旧键迁移（wap-theme → wap:theme）', () => {
      localStorage.setItem('wap-theme', 'dark')
      expect(getWapTheme()).toBe('dark')
      // 旧键已删，新键已写
      expect(localStorage.getItem('wap-theme')).toBeNull()
      expect(localStorage.getItem('wap:theme')).toBe('dark')
    })
  })

  describe('applyWapTheme', () => {
    it('写入 dark', () => {
      applyWapTheme('dark')
      expect(localStorage.getItem('wap:theme')).toBe('dark')
      expect(document.documentElement.dataset.wapTheme).toBe('dark')
    })

    it('写入 light', () => {
      applyWapTheme('light')
      expect(localStorage.getItem('wap:theme')).toBe('light')
      expect(document.documentElement.dataset.wapTheme).toBe('light')
    })
  })

  describe('toggleWapTheme', () => {
    it('light → dark', () => {
      applyWapTheme('light')
      expect(toggleWapTheme()).toBe('dark')
    })

    it('dark → light', () => {
      applyWapTheme('dark')
      expect(toggleWapTheme()).toBe('light')
    })

    it('无设置 → 默认 light，toggle 后变 dark', () => {
      expect(toggleWapTheme()).toBe('dark')
    })
  })

  describe('initWapTheme', () => {
    it('应用当前主题到 DOM', () => {
      applyWapTheme('dark')
      document.documentElement.dataset.wapTheme = ''
      initWapTheme()
      expect(document.documentElement.dataset.wapTheme).toBe('dark')
    })
  })
})
