import { describe, it, expect, vi, beforeEach } from 'vitest'

const echartsStub = {
  use: vi.fn(),
  init: vi.fn(),
  resize: vi.fn(),
  dispose: vi.fn(),
  getInstanceByDom: vi.fn(),
  setOption: vi.fn(),
  registerTheme: vi.fn(),
  connect: vi.fn(),
  disconnect: vi.fn()
}

vi.mock('echarts/core', () => {
  const c = {
    use: vi.fn(),
    init: vi.fn(),
    resize: vi.fn(),
    dispose: vi.fn(),
    getInstanceByDom: vi.fn(),
    setOption: vi.fn(),
    registerTheme: vi.fn(),
    connect: vi.fn(),
    disconnect: vi.fn()
  }
  ;(globalThis as any).__echartsCoreMock = c
  return { default: c, ...c }
})

vi.mock('echarts/charts', () => ({
  BarChart: { __type: 'BarChart' },
  LineChart: { __type: 'LineChart' },
  PieChart: { __type: 'PieChart' }
}))

vi.mock('echarts/components', () => ({
  GridComponent: { __type: 'GridComponent' },
  LegendComponent: { __type: 'LegendComponent' },
  TooltipComponent: { __type: 'TooltipComponent' }
}))

vi.mock('echarts/renderers', () => ({
  CanvasRenderer: { __type: 'CanvasRenderer' }
}))

import echarts from './echartsCore'

const mock = (): any => (globalThis as any).__echartsCoreMock

beforeEach(() => {
  const c = mock()
  c.init.mockClear()
  c.resize.mockClear()
  c.dispose.mockClear()
  c.getInstanceByDom.mockClear()
  // Intentionally NOT clearing c.use — it is called exactly once at module load
})

describe('echartsCore default export', () => {
  it('is the echarts core instance with standard API', () => {
    expect(echarts).toBeDefined()
    expect(typeof echarts.use).toBe('function')
    expect(typeof echarts.init).toBe('function')
    expect(typeof echarts.resize).toBe('function')
    expect(typeof echarts.dispose).toBe('function')
    expect(typeof echarts.getInstanceByDom).toBe('function')
  })

  it('registers all chart/component/renderer modules on import', () => {
    const useMock = mock().use
    expect(useMock).toHaveBeenCalled()
    const callArgs = useMock.mock.calls[0]?.[0] || []
    const types = callArgs.map((m: any) => m.__type)
    expect(types).toEqual(
      expect.arrayContaining([
        'BarChart',
        'LineChart',
        'PieChart',
        'GridComponent',
        'LegendComponent',
        'TooltipComponent',
        'CanvasRenderer'
      ])
    )
    expect(callArgs.length).toBe(7)
  })

  it('init initializes an instance on a DOM element', () => {
    const initMock = mock().init
    const fakeDom = document.createElement('div')
    initMock.mockReturnValue({ id: 'fake-instance' })
    const inst = echarts.init(fakeDom, 'dark', { renderer: 'canvas' })
    expect(initMock).toHaveBeenCalledWith(fakeDom, 'dark', { renderer: 'canvas' })
    expect(inst).toEqual({ id: 'fake-instance' })
  })

  it('resize resizes the given chart instance', () => {
    const resizeMock = mock().resize
    const inst = { id: 'inst' }
    echarts.resize(inst, 800, 600)
    expect(resizeMock).toHaveBeenCalledWith(inst, 800, 600)
  })

  it('resize works without explicit width/height', () => {
    const resizeMock = mock().resize
    const inst = { id: 'inst' }
    echarts.resize(inst)
    expect(resizeMock).toHaveBeenCalledWith(inst)
  })

  it('dispose disposes the instance and frees resources', () => {
    const disposeMock = mock().dispose
    const inst = { id: 'inst' }
    echarts.dispose(inst)
    expect(disposeMock).toHaveBeenCalledWith(inst)
  })

  it('getInstanceByDom returns existing instance for a DOM node', () => {
    const getInstanceByDomMock = mock().getInstanceByDom
    const dom = document.createElement('div')
    getInstanceByDomMock.mockReturnValue({ id: 'existing' })
    const inst = echarts.getInstanceByDom(dom)
    expect(getInstanceByDomMock).toHaveBeenCalledWith(dom)
    expect(inst).toEqual({ id: 'existing' })
  })

  it('getInstanceByDom returns null/undefined when no instance attached', () => {
    const getInstanceByDomMock = mock().getInstanceByDom
    getInstanceByDomMock.mockReturnValueOnce(null)
    expect(echarts.getInstanceByDom(document.createElement('div'))).toBeNull()
  })
})

describe('module-level side effect', () => {
  it('echarts.use is called exactly once at module evaluation', () => {
    expect(mock().use).toHaveBeenCalledTimes(1)
  })
})

// ====== 补充：dispose / resize / 多次初始化 ======

describe('echartsCore 补充 - 多次初始化与 dispose 协同', () => {
  beforeEach(() => {
    mock().init.mockReset()
    mock().resize.mockReset()
    mock().dispose.mockReset()
    mock().getInstanceByDom.mockReset()
  })

  it('同一 DOM 上多次 init 返回独立实例', () => {
    const initMock = mock().init
    initMock.mockReturnValueOnce({ id: 'first' })
    initMock.mockReturnValueOnce({ id: 'second' })
    const dom = document.createElement('div')
    const a = echarts.init(dom)
    const b = echarts.init(dom)
    expect(a).toEqual({ id: 'first' })
    expect(b).toEqual({ id: 'second' })
    expect(initMock).toHaveBeenCalledTimes(2)
    expect(initMock.mock.calls[0][0]).toBe(dom)
    expect(initMock.mock.calls[1][0]).toBe(dom)
  })

  it('init 不传 theme 与 config 时只传 DOM', () => {
    const initMock = mock().init
    initMock.mockReturnValue({ id: 'inst' })
    const dom = document.createElement('div')
    echarts.init(dom)
    expect(initMock.mock.calls[0][0]).toBe(dom)
    expect(initMock.mock.calls[0].length).toBe(1)
  })

  it('dispose 之后允许再次 init 同一 DOM', () => {
    const initMock = mock().init
    initMock.mockReturnValue({ id: 'fresh' })
    const dom = document.createElement('div')
    const inst = echarts.init(dom)
    echarts.dispose(inst)
    const inst2 = echarts.init(dom)
    expect(inst2).toEqual({ id: 'fresh' })
    expect(mock().dispose).toHaveBeenCalledWith(inst)
  })

  it('resize 在 dispose 之前/之后均可调用（不抛错）', () => {
    const inst = { id: 'inst' }
    expect(() => echarts.resize(inst, 100, 100)).not.toThrow()
    echarts.dispose(inst)
    expect(() => echarts.resize(inst)).not.toThrow()
    expect(mock().resize).toHaveBeenCalledTimes(2)
  })
})

describe('echartsCore 补充 - dispose 边界', () => {
  beforeEach(() => {
    mock().dispose.mockReset()
  })

  it('dispose(null) 直接转发', () => {
    echarts.dispose(null as any)
    expect(mock().dispose).toHaveBeenCalledWith(null)
  })

  it('连续 dispose 多个不同实例', () => {
    const inst1 = { id: 'a' }
    const inst2 = { id: 'b' }
    echarts.dispose(inst1)
    echarts.dispose(inst2)
    expect(mock().dispose).toHaveBeenNthCalledWith(1, inst1)
    expect(mock().dispose).toHaveBeenNthCalledWith(2, inst2)
    expect(mock().dispose).toHaveBeenCalledTimes(2)
  })
})

describe('echartsCore 补充 - resize 边界', () => {
  beforeEach(() => {
    mock().resize.mockReset()
  })

  it('resize 单独传 width 不传 height', () => {
    const inst = { id: 'i' }
    echarts.resize(inst, 800)
    expect(mock().resize).toHaveBeenCalledTimes(1)
    expect(mock().resize.mock.calls[0][0]).toBe(inst)
    expect(mock().resize.mock.calls[0][1]).toBe(800)
    expect(mock().resize.mock.calls[0][2]).toBeUndefined()
  })

  it('resize 三个参数都传时全部转发', () => {
    const inst = { id: 'i' }
    echarts.resize(inst, 1920, 1080)
    expect(mock().resize).toHaveBeenCalledWith(inst, 1920, 1080)
  })
})

// keep reference so unused-var lint passes
void echartsStub
