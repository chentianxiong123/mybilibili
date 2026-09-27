import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import DashboardView from './DashboardView.vue'

const apis = vi.hoisted(() => ({
  getOverviewStatistics: vi.fn(),
  getManuscriptStatusStatistics: vi.fn(),
  getRecentManuscripts: vi.fn(),
}))
vi.mock('@/api/statistics', () => apis)

const routerPush = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({ useRouter: () => ({ push: routerPush }) }))

vi.mock('@element-plus/icons-vue', () => {
  const stub = { template: '<i />' }
  return {
    User: stub,
    Document: stub,
    VideoCamera: stub,
    View: stub,
    Folder: stub,
    Warning: stub,
    DocumentChecked: stub,
    DataAnalysis: stub,
    Picture: stub,
  }
})

const makeWrapper = () =>
  mount(DashboardView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
          emits: ['click'],
        },
        'el-card': { template: '<div class="el-card-stub"><slot /></div>' },
        'el-icon': { template: '<span><slot /></span>' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-progress': { template: '<div class="el-progress-stub" />' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  routerPush.mockReset()
  for (const fn of Object.values(apis)) fn.mockReset()
  apis.getOverviewStatistics.mockResolvedValue({ code: 200, data: {} })
  apis.getManuscriptStatusStatistics.mockResolvedValue({ code: 200, data: {} })
  apis.getRecentManuscripts.mockResolvedValue({ code: 200, data: [] })
})

describe('DashboardView.vue', () => {
  it('onMounted 并发拉取三个统计接口，最近稿件取 5 条', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getOverviewStatistics).toHaveBeenCalledTimes(1)
    expect(apis.getManuscriptStatusStatistics).toHaveBeenCalledTimes(1)
    expect(apis.getRecentManuscripts).toHaveBeenCalledWith(5)
  })

  it('概览：snake_case 归一化', async () => {
    apis.getOverviewStatistics.mockResolvedValue({
      code: 200,
      data: { user_count: 10, manuscript_count: 20, video_count: 30, view_count: 40, pending_count: 5 },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.statistics).toEqual({
      userCount: 10,
      manuscriptCount: 20,
      videoCount: 30,
      viewCount: 40,
      pendingManuscriptCount: 5,
    })
  })

  it('概览：camelCase 也可识别，0 值不被 ?? 吞掉', async () => {
    apis.getOverviewStatistics.mockResolvedValue({
      code: 200,
      data: { userCount: 0, manuscriptCount: 0, videoCount: 0, viewCount: 0, pendingManuscriptCount: 0 },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.statistics.userCount).toBe(0)
    expect(w.vm.statistics.pendingManuscriptCount).toBe(0)
  })

  it('概览：兼容 success 标记', async () => {
    apis.getOverviewStatistics.mockResolvedValue({ success: true, data: { user_count: 7 } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.statistics.userCount).toBe(7)
  })

  it('稿件状态：对象结构被转成 [{status,count}] 便于渲染', async () => {
    apis.getManuscriptStatusStatistics.mockResolvedValue({
      code: 200,
      data: { pending_review: 3, published: 12 },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.manuscriptStatus).toEqual([
      { status: 'pending_review', count: 3 },
      { status: 'published', count: 12 },
    ])
  })

  it('稿件状态：data 缺失时置空数组', async () => {
    apis.getManuscriptStatusStatistics.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.manuscriptStatus).toEqual([])
  })

  it('最近稿件：兼容 upload_time / uploadTime / created_at / createdAt 四种时间字段', async () => {
    apis.getRecentManuscripts.mockResolvedValue({
      code: 200,
      data: [
        { id: 1, title: 'a', status: 0, upload_time: 'T1' },
        { id: 2, title: 'b', status: 1, uploadTime: 'T2' },
        { id: 3, title: 'c', status: 2, created_at: 'T3' },
        { id: 4, title: 'd', status: 3, createdAt: 'T4' },
        { id: 5, status: 4 },
      ],
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.recentManuscripts.map((m: any) => m.uploadTime)).toEqual(['T1', 'T2', 'T3', 'T4', ''])
    expect(w.vm.recentManuscripts[4].title).toBe('')
  })

  it('最近稿件：data 非数组时置空数组，不崩', async () => {
    apis.getRecentManuscripts.mockResolvedValue({ code: 200, data: { list: [] } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.recentManuscripts).toEqual([])
  })

  it('业务失败：各自独立跳过，不影响其他数据', async () => {
    apis.getOverviewStatistics.mockResolvedValue({ code: 500 })
    apis.getRecentManuscripts.mockResolvedValue({ code: 200, data: [{ id: 1, title: 'x' }] })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.statistics.userCount).toBe(0)
    expect(w.vm.recentManuscripts).toHaveLength(1)
  })

  it('接口异常：记日志、复位 loading，不产生未捕获拒绝', async () => {
    apis.getOverviewStatistics.mockRejectedValue(new Error('network'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    await flushPromises()
    expect(errSpy).toHaveBeenCalled()
    expect(w.vm.loading).toBe(false)
    errSpy.mockRestore()
  })

  it('getStatusText：字符串态与数字态都映射，未知回落到"未知"', () => {
    const w = makeWrapper()
    expect(w.vm.getStatusText('pending_review')).toBe('待审核')
    expect(w.vm.getStatusText('published')).toBe('已上架')
    expect(w.vm.getStatusText(3)).toBe('已上架')
    expect(w.vm.getStatusText(-1)).toBe('已下架')
    expect(w.vm.getStatusText('brand_new')).toBe('未知')
  })

  it('getStatusType：命中返回标签色，未知返回空串', () => {
    const w = makeWrapper()
    expect(w.vm.getStatusType('approved')).toBe('success')
    expect(w.vm.getStatusType(4)).toBe('danger')
    expect(w.vm.getStatusType('???')).toBe('')
  })

  it('formatDateTime：空值与非法时间都返回占位符', () => {
    const w = makeWrapper()
    expect(w.vm.formatDateTime('')).toBe('-')
    expect(w.vm.formatDateTime('not-a-date')).toBe('-')
    expect(w.vm.formatDateTime(new Date(2026, 0, 2, 3, 4, 5))).not.toBe('-')
  })

  it('goToManuscripts：跳转到稿件管理路由', () => {
    const w = makeWrapper()
    w.vm.goToManuscripts()
    expect(routerPush).toHaveBeenCalledWith('/manuscripts')
  })
})
