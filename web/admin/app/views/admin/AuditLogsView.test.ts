import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AuditLogsView from './AuditLogsView.vue'

const apis = vi.hoisted(() => ({ getAuditLogs: vi.fn(), getAuditLogDetail: vi.fn() }))
vi.mock('@/api/audit', () => apis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({ ElMessage: messageMocks }))

vi.mock('@element-plus/icons-vue', () => ({
  Refresh: { template: '<i />' },
  Search: { template: '<i />' },
}))

const emptySearch = () => ({
  operatorKeyword: '',
  module: '',
  action: '',
  result: null,
  targetKeyword: '',
  startTime: '',
  endTime: '',
})

const makeWrapper = () =>
  mount(AuditLogsView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
          emits: ['click'],
        },
        'el-input': {
          template:
            '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(`update:modelValue`, e.target.value)" />',
          props: ['modelValue'],
        },
        'el-select': { template: '<div class="el-select-stub"><slot /></div>' },
        'el-option': { template: '<div />' },
        'el-date-picker': { template: '<div />' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-pagination': { template: '<div />' },
        'el-descriptions': { template: '<div><slot /></div>' },
        'el-descriptions-item': { template: '<div><slot /></div>' },
        'el-dialog': { template: '<div class="el-dialog-stub"><slot /></div>' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  apis.getAuditLogs.mockReset()
  apis.getAuditLogDetail.mockReset()
  apis.getAuditLogs.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
})

describe('AuditLogsView.vue', () => {
  it('setup 阶段即拉取审计日志', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getAuditLogs).toHaveBeenCalledTimes(1)
    expect(apis.getAuditLogs).toHaveBeenCalledWith({ page: 1, size: 10, ...emptySearch() })
  })

  it('snake_case 字段归一化', async () => {
    apis.getAuditLogs.mockResolvedValue({
      code: 200,
      data: {
        list: [
          {
            id: 1,
            operator_id: 9,
            operator_name: 'root',
            operator_role: 'super',
            module: 'auth',
            action: 'admin_login',
            target_type: 'admin',
            target_id: 3,
            request_method: 'POST',
            request_uri: '/api/v1/admin/login',
            client_ip: '10.0.0.1',
            user_agent: 'UA',
            result: 1,
            message: 'ok',
            detail: '{}',
            created_at: '2026-01-01 00:00:00',
          },
        ],
        total: 1,
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0]).toMatchObject({
      id: 1,
      operatorId: 9,
      operatorName: 'root',
      operatorRole: 'super',
      targetType: 'admin',
      targetId: 3,
      requestMethod: 'POST',
      requestUri: '/api/v1/admin/login',
      clientIp: '10.0.0.1',
      userAgent: 'UA',
      result: 1,
      createdAt: '2026-01-01 00:00:00',
    })
    expect(w.vm.total).toBe(1)
  })

  it('camelCase 字段也可识别，数字型 0 不被吞', async () => {
    apis.getAuditLogs.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 2, operatorId: 0, targetId: 0, result: 0 }] },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0].operatorId).toBe(0)
    expect(w.vm.tableData[0].targetId).toBe(0)
  })

  it('data.list 缺失时兜底空数组、total 兜底 0', async () => {
    apis.getAuditLogs.mockResolvedValue({ code: 200, data: {} })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.total).toBe(0)
  })

  it('业务失败：提示后端 message 且不覆盖列表', async () => {
    apis.getAuditLogs.mockResolvedValue({ code: 500, message: '无权访问审计日志' })
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('无权访问审计日志')
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.loading).toBe(false)
  })

  it('业务失败且无 message：回落通用文案', async () => {
    apis.getAuditLogs.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('加载审计日志失败')
  })

  it('请求异常：提示错误、复位 loading，不产生未捕获拒绝', async () => {
    apis.getAuditLogs.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('加载审计日志失败')
    expect(w.vm.loading).toBe(false)
  })

  it('搜索：带全部条件并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 5
    Object.assign(w.vm.searchForm, {
      operatorKeyword: 'root',
      module: 'auth',
      action: 'admin_login',
      result: 1,
      targetKeyword: 'admin',
      startTime: '2026-01-01 00:00:00',
      endTime: '2026-01-31 23:59:59',
    })
    w.vm.handleSearch()
    await flushPromises()
    expect(w.vm.page).toBe(1)
    expect(apis.getAuditLogs).toHaveBeenLastCalledWith({
      page: 1,
      size: 10,
      operatorKeyword: 'root',
      module: 'auth',
      action: 'admin_login',
      result: 1,
      targetKeyword: 'admin',
      startTime: '2026-01-01 00:00:00',
      endTime: '2026-01-31 23:59:59',
    })
  })

  it('重置：清空全部条件并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 3
    Object.assign(w.vm.searchForm, { operatorKeyword: 'a', module: 'b', action: 'c', result: 1, targetKeyword: 'd', startTime: 'e', endTime: 'f' })
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.searchForm).toEqual(emptySearch())
    expect(w.vm.page).toBe(1)
  })

  it('翻页保留筛选条件', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.searchForm.module = 'user'
    w.vm.handlePageChange(4)
    await flushPromises()
    expect(apis.getAuditLogs).toHaveBeenLastCalledWith(expect.objectContaining({ page: 4, module: 'user' }))
  })

  it('改每页条数：同时重置回第 1 页，避免落在越界页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 9
    w.vm.handleSizeChange(50)
    await flushPromises()
    expect(w.vm.page).toBe(1)
    expect(apis.getAuditLogs).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, size: 50 }))
  })

  it('查看详情：按 id 拉取并归一化写入 currentDetail', async () => {
    apis.getAuditLogDetail.mockResolvedValue({ code: 200, data: { id: 4, operator_name: 'admin', action: 'role_delete' } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.showDetail({ id: 4 })
    expect(apis.getAuditLogDetail).toHaveBeenCalledWith(4)
    expect(w.vm.detailVisible).toBe(true)
    expect(w.vm.currentDetail.operatorName).toBe('admin')
    expect(w.vm.detailLoading).toBe(false)
  })

  it('查看详情失败：先清空旧详情再提示，detailLoading 复位', async () => {
    apis.getAuditLogDetail.mockResolvedValue({ code: 500, message: '详情已清理' })
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentDetail = { id: 99 }
    await w.vm.showDetail({ id: 4 })
    expect(w.vm.currentDetail).toBeNull()
    expect(messageMocks.error).toHaveBeenCalledWith('详情已清理')
    expect(w.vm.detailLoading).toBe(false)
  })

  it('查看详情抛异常：同样要提示，不能静默', async () => {
    apis.getAuditLogDetail.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.showDetail({ id: 4 })
    expect(messageMocks.error).toHaveBeenCalledWith('加载审计详情失败')
    expect(w.vm.detailLoading).toBe(false)
  })

  it('formatTime 空值占位', () => {
    const w = makeWrapper()
    expect(w.vm.formatTime('')).toBe('-')
    expect(w.vm.formatTime(null)).toBe('-')
  })

  it('formatAction：命中中文标签，未知/空值兜底', () => {
    const w = makeWrapper()
    expect(w.vm.formatAction('admin_login')).toBe('管理员登录')
    expect(w.vm.formatAction('manuscript_approve')).toBe('稿件审核通过')
    expect(w.vm.formatAction('pipeline_retry')).toBe('重试AI任务')
    expect(w.vm.formatAction('not_exist')).toBe('not_exist')
    expect(w.vm.formatAction('')).toBe('-')
  })

  it('resultType / resultText：1 为成功，其余为失败', () => {
    const w = makeWrapper()
    expect(w.vm.resultType(1)).toBe('success')
    expect(w.vm.resultType(0)).toBe('danger')
    expect(w.vm.resultText(1)).toBe('成功')
    expect(w.vm.resultText(0)).toBe('失败')
  })
})
