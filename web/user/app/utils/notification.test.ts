import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  requestNotificationPermission,
  isNotificationSupported,
  isNotificationEnabled,
  showNotification,
  notifyStreamLive,
  notifyMention,
} from './notification'

class FakeNotification {
  static permission: NotificationPermission = 'default'
  static requestPermission = vi.fn()
  title: string
  options: NotificationOptions
  onclick: ((this: Notification, ev: Event) => any) | null = null
  closed = false
  static instances: FakeNotification[] = []

  constructor(title: string, options: NotificationOptions = {}) {
    this.title = title
    this.options = options
    FakeNotification.instances.push(this)
  }

  close() {
    this.closed = true
  }
}

describe('notification utils', () => {
  beforeEach(() => {
    FakeNotification.permission = 'default'
    FakeNotification.requestPermission = vi.fn()
    FakeNotification.instances = []
    vi.stubGlobal('Notification', FakeNotification)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  describe('isNotificationSupported', () => {
    it('Notification 存在时返回 true', () => {
      expect(isNotificationSupported()).toBe(true)
    })

    it('Notification 不存在时返回 false', () => {
      const original = (window as any).Notification
      // @ts-expect-error testing absence
      delete (window as any).Notification
      expect(isNotificationSupported()).toBe(false)
      ; (window as any).Notification = original
    })
  })

  describe('isNotificationEnabled', () => {
    it('granted 时返回 true', () => {
      FakeNotification.permission = 'granted'
      expect(isNotificationEnabled()).toBe(true)
    })

    it('default 时返回 false', () => {
      FakeNotification.permission = 'default'
      expect(isNotificationEnabled()).toBe(false)
    })

    it('denied 时返回 false', () => {
      FakeNotification.permission = 'denied'
      expect(isNotificationEnabled()).toBe(false)
    })
  })

  describe('requestNotificationPermission', () => {
    it('不支持 Notification 时返回 false', async () => {
      const original = (window as any).Notification
      // @ts-expect-error testing absence
      delete (window as any).Notification
      const result = await requestNotificationPermission()
      expect(result).toBe(false)
      ;(window as any).Notification = original
    })

    it('已 granted 时直接返回 true', async () => {
      FakeNotification.permission = 'granted'
      const result = await requestNotificationPermission()
      expect(result).toBe(true)
      expect(FakeNotification.requestPermission).not.toHaveBeenCalled()
    })

    it('已 denied 时返回 false（不调用 requestPermission）', async () => {
      FakeNotification.permission = 'denied'
      const result = await requestNotificationPermission()
      expect(result).toBe(false)
      expect(FakeNotification.requestPermission).not.toHaveBeenCalled()
    })

    it('default 时调用 requestPermission 并按结果返回', async () => {
      FakeNotification.permission = 'default'
      FakeNotification.requestPermission.mockResolvedValue('granted')
      const result = await requestNotificationPermission()
      expect(FakeNotification.requestPermission).toHaveBeenCalledTimes(1)
      expect(result).toBe(true)
    })

    it('default 但用户拒绝时返回 false', async () => {
      FakeNotification.permission = 'default'
      FakeNotification.requestPermission.mockResolvedValue('denied')
      const result = await requestNotificationPermission()
      expect(result).toBe(false)
    })
  })

  describe('showNotification', () => {
    it('granted 时创建 Notification 并返回实例', () => {
      FakeNotification.permission = 'granted'
      const spy = vi.spyOn(window, 'focus').mockImplementation(() => {})
      const locSpy = vi.spyOn(window.location, 'href', 'set').mockImplementation(() => {})

      const notif = showNotification('hi', { body: 'world' })
      expect(notif).toBeInstanceOf(FakeNotification)
      expect(notif?.title).toBe('hi')
      expect(notif?.options.body).toBe('world')
      expect(FakeNotification.instances.length).toBe(1)

      spy.mockRestore()
      locSpy.mockRestore()
    })

    it('无权限时返回 null 且不创建实例', () => {
      FakeNotification.permission = 'default'
      const notif = showNotification('hi')
      expect(notif).toBeNull()
      expect(FakeNotification.instances.length).toBe(0)
    })

    it('options.data.url 时 onclick 会跳转并关闭', () => {
      FakeNotification.permission = 'granted'
      const focusSpy = vi.spyOn(window, 'focus').mockImplementation(() => {})
      const locSpy = vi.spyOn(window.location, 'href', 'set').mockImplementation(() => {})

      const notif = showNotification('msg', {
        data: { url: '/live/abc' },
      })
      notif?.onclick?.(new Event('click'))
      expect(notif?.closed).toBe(true)
      expect(focusSpy).toHaveBeenCalled()
      expect(locSpy).toHaveBeenCalledWith('/live/abc')

      focusSpy.mockRestore()
      locSpy.mockRestore()
    })

    it('8 秒后自动关闭（用 fake timers）', () => {
      FakeNotification.permission = 'granted'
      vi.useFakeTimers()
      const notif = showNotification('msg')
      expect(notif?.closed).toBe(false)
      vi.advanceTimersByTime(8000)
      expect(notif?.closed).toBe(true)
    })

    it('不传 options 时不抛错', () => {
      FakeNotification.permission = 'granted'
      const notif = showNotification('title')
      expect(notif).toBeInstanceOf(FakeNotification)
    })
  })

  describe('notifyStreamLive', () => {
    it('生成带正确 title/body/tag/data 的通知', () => {
      FakeNotification.permission = 'granted'
      const focusSpy = vi.spyOn(window, 'focus').mockImplementation(() => {})

      const notif = notifyStreamLive('room123', 'Alice')
      expect(notif?.title).toBe('Alice 开播了！')
      expect(notif?.options.body).toBe('room123 正在进行直播，点击进入')
      expect(notif?.options.tag).toBe('stream-live')
      expect((notif?.options as any).data?.url).toBe('/live/room123')

      focusSpy.mockRestore()
    })

    it('无权限时不抛', () => {
      FakeNotification.permission = 'denied'
      expect(() => notifyStreamLive('room', 'Alice')).not.toThrow()
    })
  })

  describe('notifyMention', () => {
    it('文本 <= 60 时完整传 body', () => {
      FakeNotification.permission = 'granted'
      const notif = notifyMention('Bob', 'room1', 'hello world')
      expect(notif?.title).toBe('Bob 在直播间提及你')
      expect(notif?.options.body).toBe('hello world')
      expect(notif?.options.tag).toBe('mention')
      expect((notif?.options as any).data?.url).toBe('/live/room1')
    })

    it('文本 > 60 时截断为 60 字 + ...', () => {
      FakeNotification.permission = 'granted'
      const long = 'x'.repeat(120)
      const notif = notifyMention('Bob', 'room1', long)
      expect((notif?.options.body as string).length).toBe(63) // 60 + '...'
      expect((notif?.options.body as string).endsWith('...')).toBe(true)
    })

    it('无权限时不抛', () => {
      FakeNotification.permission = 'default'
      expect(() => notifyMention('Bob', 'r', 'hi')).not.toThrow()
    })
  })
})