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

// keep reference so unused-var lint passes
void echartsStub
