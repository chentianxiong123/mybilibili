import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// 收集 onMounted / onUnmounted 回调
const mountedCbs: Array<() => void> = []
const unmountedCbs: Array<() => void> = []

// 拦截 vue 的 onMounted / onUnmounted（useTabsVideoAlign 直接 import 它们）
vi.mock('vue', async () => {
  const actual: any = await vi.importActual('vue')
  return {
    ...actual,
    onMounted: (cb: () => void) => { mountedCbs.push(cb) },
    onUnmounted: (cb: () => void) => { unmountedCbs.push(cb) },
  }
})

import { useTabsVideoAlign } from './useTabsVideoAlign'

class FakeResizeObserver {
  static instances: FakeResizeObserver[] = []
  cb: ResizeObserverCallback
  observed: Element[] = []
  constructor(cb: ResizeObserverCallback) {
    this.cb = cb
    FakeResizeObserver.instances.push(this)
  }
  observe(target: Element) { this.observed.push(target) }
  unobserve() {}
  disconnect() { this.observed = [] }
}

function makeEl(left: number, width: number, paddingLeft = 0) {
  return {
    style: {} as Record<string, string>,
    getBoundingClientRect: () => ({
      left, width, top: 0, bottom: 0, right: left + width, height: 100, x: left, y: 0, toJSON() { return {} }
    }),
    parentElement: { getBoundingClientRect: () => ({ left: 0, width: 1200, top: 0, bottom: 0, right: 1200, height: 100, x: 0, y: 0, toJSON() { return {} } }) },
  } as any
}

describe('useTabsVideoAlign', () => {
  beforeEach(() => {
    mountedCbs.length = 0
    unmountedCbs.length = 0
    FakeResizeObserver.instances = []
    vi.stubGlobal('ResizeObserver', FakeResizeObserver)
    ;(import.meta as any).server = false
    document.body.innerHTML = ''
  })
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
    vi.restoreAllMocks()
  })

  it('onMounted 注册 ResizeObserver 监听 documentElement', () => {
    useTabsVideoAlign()
    expect(mountedCbs.length).toBe(1)
    mountedCbs[0]()
    expect(FakeResizeObserver.instances.length).toBe(1)
    expect(FakeResizeObserver.instances[0].observed).toContain(document.documentElement)
  })

  it('onUnmounted 移除 ResizeObserver', () => {
    useTabsVideoAlign()
    mountedCbs[0]()
    const ro = FakeResizeObserver.instances[0]
    const disconnectSpy = vi.spyOn(ro, 'disconnect')
    unmountedCbs[0]()
    expect(disconnectSpy).toHaveBeenCalled()
  })

  it('apply 在没有 .tabs-main / .main-section 时安全 return', () => {
    const { apply } = useTabsVideoAlign()
    expect(() => apply()).not.toThrow()
  })

  it('apply 正确写入 style', () => {
    const tabs = makeEl(100, 1000, 16)
    const main = makeEl(0, 800, 0)
    const parentRect = { left: 0, width: 1200, top: 0, bottom: 0, right: 1200, height: 100, x: 0, y: 0, toJSON() { return {} } }
    tabs.parentElement = { getBoundingClientRect: () => parentRect } as any
    main.parentElement = { getBoundingClientRect: () => parentRect } as any
    const tabsMock = vi.spyOn(document, 'querySelector').mockImplementation((sel: string) => {
      if (sel === '.tabs-main-container') return tabs
      if (sel === '.main-section') return main
      return null
    })
    vi.spyOn(window, 'getComputedStyle').mockReturnValue({ paddingLeft: '16px' } as any)
    const { apply } = useTabsVideoAlign()
    apply()
    expect(main.style.width).toBe('1000px')
    expect(main.style.maxWidth).toBe('1000px')
    expect(main.style.marginLeft).toBe('100px')
    expect(main.style.marginRight).toBe('auto')
    expect(main.style.paddingLeft).toBe('16px')
    expect(main.style.paddingRight).toBe('16px')
    tabsMock.mockRestore()
  })

  it('apply 重复调用相同尺寸时不重写 style', () => {
    const tabs = makeEl(100, 1000, 16)
    const main = makeEl(0, 800, 0)
    const parentRect = { left: 0, width: 1200, top: 0, bottom: 0, right: 1200, height: 100, x: 0, y: 0, toJSON() { return {} } }
    tabs.parentElement = { getBoundingClientRect: () => parentRect } as any
    main.parentElement = { getBoundingClientRect: () => parentRect } as any
    const spy = vi.spyOn(document, 'querySelector').mockImplementation((sel: string) => {
      if (sel === '.tabs-main-container') return tabs
      if (sel === '.main-section') return main
      return null
    })
    vi.spyOn(window, 'getComputedStyle').mockReturnValue({ paddingLeft: '16px' } as any)
    const { apply } = useTabsVideoAlign()
    apply()
    const width1 = main.style.width
    const margin1 = main.style.marginLeft
    apply()
    expect(main.style.width).toBe(width1)
    expect(main.style.marginLeft).toBe(margin1)
    spy.mockRestore()
  })

  it('当 tabs / main 偏移和宽度同时 ~ 0 时清空之前的 style（需要先应用过非零值）', () => {
    const parentRect = { left: 0, width: 1200, top: 0, bottom: 0, right: 1200, height: 100, x: 0, y: 0, toJSON() { return {} } }
    const tabs = makeEl(100, 1000, 16)
    const main = makeEl(0, 800, 0)
    tabs.parentElement = { getBoundingClientRect: () => parentRect } as any
    main.parentElement = { getBoundingClientRect: () => parentRect } as any
    const spy = vi.spyOn(document, 'querySelector').mockImplementation((sel: string) => {
      if (sel === '.tabs-main-container') return tabs
      if (sel === '.main-section') return main
      return null
    })
    vi.spyOn(window, 'getComputedStyle').mockReturnValue({ paddingLeft: '16px' } as any)
    const { apply } = useTabsVideoAlign()
    apply()
    expect(main.style.width).toBe('1000px')

    // 现在切换为 0 尺寸
    tabs.getBoundingClientRect = () => ({ left: 0, width: 0, top: 0, bottom: 0, right: 0, height: 100, x: 0, y: 0, toJSON() { return {} } }) as any
    apply()
    expect(main.style.width).toBe('')
    expect(main.style.marginLeft).toBe('')
    spy.mockRestore()
  })

  it('不存在 ResizeObserver 时不会报错', () => {
    vi.unstubAllGlobals()
    // @ts-expect-error 强制删除
    delete (globalThis as any).ResizeObserver
    expect(() => useTabsVideoAlign()).not.toThrow()
  })

  it('返回的 apply 与 apply 内调用同步', () => {
    const r = useTabsVideoAlign()
    expect(typeof r.apply).toBe('function')
    r.apply()
  })
})