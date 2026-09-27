import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import LoginLogsView from './LoginLogsView.vue'

const getLoginLogs = vi.hoisted(() => vi.fn())
vi.mock('@/api/securitySettings.ts', () => ({
  adminLoginLogApi: { getLoginLogs },
}))

const errorMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: { error: errorMock, success: vi.fn(), warning: vi.fn() },
  ElMessageBox: { confirm: vi.fn() },
}))

const makeWrapper = () =>
  mount(LoginLogsView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-input': {
          template:
            '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(\'update:modelValue\', e.target.value)" />',
          props: ['modelValue'],
        },
        'el-select': { template: '<div class="el-select-stub"><slot /></div>' },
        'el-option': { template: '<div class="el-option-stub" />' },
        'el-date-picker': { template: '<div class="el-date-picker-stub" />' },
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(\'click\')"><slot /></button>',
        },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-pagination': { template: '<div class="el-pagination-stub" />' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  errorMock.mockReset()
  getLoginLogs.mockReset()
  getLoginLogs.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
})

describe('LoginLogsView.vue', () => {
  it('setup 阶段即拉取一次日志列表', async () => {
    makeWrapper()
    await flushPromises()
    expect(getLoginLogs).toHaveBeenCalledTimes(1)
    expect(getLoginLogs).toHaveBeenCalledWith({
      page: 1,
      size: 10,
      ip: '',
      userId: null,
      status: null,
      startTime: '',
      endTime: '',
    })
  })

  it('snake_case 字段归一化为表格用的 camelCase', async () => {
    getLoginLogs.mockResolvedValue({
      code: 200,
      data: {
        list: [
          {
            id: 7,
            user_id: 42,
            ip: '10.0.0.1',
            login_time: '2026-01-02 03:04:05',
            status: 1,
            user_agent: 'UA-snake',
          },
        ],
        total: 1,
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([
      {
        id: 7,
        userId: 42,
        ip: '10.0.0.1',
        loginTime: '2026-01-02 03:04:05',
        status: 1,
        userAgent: 'UA-snake',
      },
    ])
    expect(w.vm.total).toBe(1)
  })

  it('camelCase 字段也可识别，缺失字段回落为空值', async () => {
    getLoginLogs.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 1, userId: 9, loginTime: 'T', userAgent: 'UA-camel' }], total: 3 },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0]).toMatchObject({
      id: 1,
      userId: 9,
      loginTime: 'T',
      userAgent: 'UA-camel',
    })
    // ip 未提供 → 归一化为空串而非 undefined
    expect(w.vm.tableData[0].ip).toBe('')
    expect(w.vm.total).toBe(3)
  })

  it('code !== 200 时不覆盖列表，loading 仍被复位', async () => {
    getLoginLogs.mockResolvedValue({ code: 500, message: 'boom' })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.total).toBe(0)
    expect(w.vm.loading).toBe(false)
  })

  it('data.list 缺失时兜底为空数组', async () => {
    getLoginLogs.mockResolvedValue({ code: 200, data: {} })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
  })

  it('搜索：带条件查询并重置回第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 4
    w.vm.searchForm.ip = '192.168.1.1'
    w.vm.searchForm.userId = 5
    w.vm.searchForm.status = 0
    w.vm.searchForm.startTime = '2026-01-01 00:00:00'
    w.vm.searchForm.endTime = '2026-01-31 23:59:59'
    w.vm.handleSearch()
    await flushPromises()
    expect(w.vm.page).toBe(1)
    expect(getLoginLogs).toHaveBeenLastCalledWith({
      page: 1,
      size: 10,
      ip: '192.168.1.1',
      userId: 5,
      status: 0,
      startTime: '2026-01-01 00:00:00',
      endTime: '2026-01-31 23:59:59',
    })
  })

  it('重置：清空全部搜索条件并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 3
    w.vm.searchForm.ip = '1.1.1.1'
    w.vm.searchForm.userId = 8
    w.vm.searchForm.status = 1
    w.vm.searchForm.startTime = 'x'
    w.vm.searchForm.endTime = 'y'
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.searchForm).toEqual({
      ip: '',
      userId: null,
      status: null,
      startTime: '',
      endTime: '',
    })
    expect(w.vm.page).toBe(1)
    expect(getLoginLogs).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, ip: '' }))
  })

  it('分页：翻页与改每页条数都会带新参数重新查询', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handlePageChange(5)
    await flushPromises()
    expect(getLoginLogs).toHaveBeenLastCalledWith(expect.objectContaining({ page: 5, size: 10 }))

    w.vm.handleSizeChange(50)
    await flushPromises()
    expect(getLoginLogs).toHaveBeenLastCalledWith(expect.objectContaining({ page: 5, size: 50 }))
  })

  it('formatTime：空值占位，非空走本地化格式', () => {
    const w = makeWrapper()
    expect(w.vm.formatTime('')).toBe('-')
    expect(w.vm.formatTime(null)).toBe('-')
    expect(w.vm.formatTime('2026-01-02T03:04:05Z')).not.toBe('-')
  })

  it('formatStatus：1 为成功，其余为失败', () => {
    const w = makeWrapper()
    expect(w.vm.formatStatus(1)).toBe('成功')
    expect(w.vm.formatStatus(0)).toBe('失败')
    expect(w.vm.formatStatus(undefined)).toBe('失败')
  })

  it('请求异常：提示错误、复位 loading，且不再产生未捕获拒绝', async () => {
    getLoginLogs.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(errorMock).toHaveBeenCalledWith('获取登录日志失败')
    expect(w.vm.loading).toBe(false)
  })
})
