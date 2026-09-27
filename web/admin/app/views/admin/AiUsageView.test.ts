import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AiUsageView from './AiUsageView.vue'

const apis = vi.hoisted(() => ({
  getAiUsageOverview: vi.fn(),
  getAiUsageFeatures: vi.fn(),
  getAiUsageDaily: vi.fn(),
}))
vi.mock('@/api/aiUsage', () => apis)

vi.mock('@element-plus/icons-vue', () => ({
  DataAnalysis: { template: '<i />' },
  Connection: { template: '<i />' },
  SuccessFilled: { template: '<i />' },
  Timer: { template: '<i />' },
}))

const setOption = vi.hoisted(() => vi.fn())
const dispose = vi.hoisted(() => vi.fn())
const echartsInit = vi.hoisted(() => vi.fn())
vi.mock('@/utils/echartsCore', () => ({
  default: {
    init: echartsInit,
  },
}))

const makeWrapper = () =>
  mount(AiUsageView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-icon': { template: '<span><slot /></span>' },
        'el-select': { template: '<div><slot /></div>' },
        'el-option': { template: '<div />' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  setOption.mockReset()
  dispose.mockReset()
  echartsInit.mockReset()
  echartsInit.mockReturnValue({ setOption, dispose })
  apis.getAiUsageOverview.mockReset()
  apis.getAiUsageFeatures.mockReset()
  apis.getAiUsageDaily.mockReset()
  apis.getAiUsageOverview.mockResolvedValue({ code: 200, data: { totalCount: 10, totalTokens: 100, avgDuration: 500, successCount: 8 } })
  apis.getAiUsageFeatures.mockResolvedValue({ code: 200, data: [{ feature: 'CHAT', count: 6 }] })
  apis.getAiUsageDaily.mockResolvedValue({ code: 200, data: [{ date: '2026-01-01', count: 4, successCount: 3, totalTokens: 40 }] })
})

describe('AiUsageView.vue', () => {
  it('onMounted 并发拉取概览/功能分布/每日趋势三个接口', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getAiUsageOverview).toHaveBeenCalledTimes(1)
    expect(apis.getAiUsageFeatures).toHaveBeenCalledTimes(1)
    expect(apis.getAiUsageDaily).toHaveBeenCalledWith(7)
  })

  it('三个接口同时回填到对应 ref', async () => {
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.overview).toEqual({ totalCount: 10, totalTokens: 100, avgDuration: 500, successCount: 8 })
    expect(w.vm.features).toEqual([{ feature: 'CHAT', count: 6 }])
    expect(w.vm.daily).toEqual([{ date: '2026-01-01', count: 4, successCount: 3, totalTokens: 40 }])
    expect(w.vm.loading).toBe(false)
  })

  it('单个接口 code !== 200 时跳过该项，不影响其他数据', async () => {
    apis.getAiUsageFeatures.mockResolvedValue({ code: 500, data: [{ feature: 'X', count: 1 }] })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.features).toEqual([])
    expect(w.vm.overview.totalCount).toBe(10)
    expect(w.vm.daily).toHaveLength(1)
  })

  it('拉取失败：记录错误并复位 loading，不抛出未捕获异常', async () => {
    apis.getAiUsageOverview.mockRejectedValue(new Error('network'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    await flushPromises()
    expect(errSpy).toHaveBeenCalled()
    expect(w.vm.loading).toBe(false)
    errSpy.mockRestore()
  })

  it('formatDuration：空/零值为 0ms，毫秒保留整数，秒级保留一位小数', () => {
    const w = makeWrapper()
    expect(w.vm.formatDuration(0)).toBe('0ms')
    expect(w.vm.formatDuration(null)).toBe('0ms')
    expect(w.vm.formatDuration(999)).toBe('999ms')
    expect(w.vm.formatDuration(1000)).toBe('1.0s')
    expect(w.vm.formatDuration(1536)).toBe('1.5s')
  })

  it('切换统计天数后重新拉取每日趋势', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.chartDays = 30
    await w.vm.loadData()
    expect(apis.getAiUsageDaily).toHaveBeenLastCalledWith(30)
  })

  it('图表节点不存在时不初始化 echarts，也不报错', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.daily = []
    w.vm.features = []
    expect(() => w.vm.renderCharts()).not.toThrow()
    expect(echartsInit).not.toHaveBeenCalled()
  })
})
