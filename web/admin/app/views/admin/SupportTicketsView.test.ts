import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SupportTicketsView from './SupportTicketsView.vue'

const apis = vi.hoisted(() => ({ getTicketList: vi.fn(), processTicket: vi.fn(), deleteTicket: vi.fn() }))
vi.mock('@/api/supportTicket', () => apis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn(), warning: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
  ElMessageBox: { confirm: confirmMock },
}))

const makeWrapper = () =>
  mount(SupportTicketsView, {
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
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-dialog': { template: '<div class="el-dialog-stub"><slot /></div>' },
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-descriptions': { template: '<div><slot /></div>' },
        'el-descriptions-item': { template: '<div><slot /></div>' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  messageMocks.warning.mockReset()
  confirmMock.mockReset()
  for (const fn of Object.values(apis)) fn.mockReset()
  apis.getTicketList.mockResolvedValue({ code: 200, data: { list: [] } })
})

describe('SupportTicketsView.vue — 列表', () => {
  it('onMounted 无筛选时拉取全部工单', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getTicketList).toHaveBeenCalledWith({})
  })

  it('带状态筛选时把 status 传给接口', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.statusFilter = 'PENDING'
    w.vm.handleFilter()
    await flushPromises()
    expect(apis.getTicketList).toHaveBeenLastCalledWith({ status: 'PENDING' })
  })

  it('snake_case 字段归一化', async () => {
    apis.getTicketList.mockResolvedValue({
      code: 200,
      data: {
        list: [
          {
            id: 1,
            ticket_no: 'T-1',
            user_id: 5,
            session_id: 6,
            source: 'USER_FEEDBACK',
            category: 'ACCOUNT',
            priority: 'HIGH',
            status: 'PENDING',
            title: 't',
            content: 'c',
            entry_reply: 'e',
            admin_reply: '',
            assignee_admin_id: 2,
            processed_at: 'P',
            created_at: 'C',
          },
        ],
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.ticketList[0]).toMatchObject({
      id: 1,
      ticketNo: 'T-1',
      userId: 5,
      sessionId: 6,
      source: 'USER_FEEDBACK',
      category: 'ACCOUNT',
      priority: 'HIGH',
      status: 'PENDING',
      entryReply: 'e',
      assigneeAdminId: 2,
      processedAt: 'P',
      createdAt: 'C',
    })
  })

  it('camelCase 也可识别，数字 0 不被吞', async () => {
    apis.getTicketList.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 2, ticketNo: 'T-2', userId: 0, sessionId: 0, assigneeAdminId: 0 }] },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.ticketList[0].userId).toBe(0)
    expect(w.vm.ticketList[0].sessionId).toBe(0)
    expect(w.vm.ticketList[0].assigneeAdminId).toBe(0)
  })

  it('兼容 data 直接是数组的响应形状', async () => {
    apis.getTicketList.mockResolvedValue({ code: 200, data: [{ id: 3 }] })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.ticketList).toHaveLength(1)
  })

  it('data 非数组时列表置空，不崩', async () => {
    apis.getTicketList.mockResolvedValue({ code: 200, data: { list: 'oops' } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.ticketList).toEqual([])
  })

  it('业务失败：不覆盖列表', async () => {
    apis.getTicketList.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.ticketList).toEqual([])
    expect(w.vm.loading).toBe(false)
  })

  it('拉取异常：记日志并复位 loading，不产生未捕获拒绝', async () => {
    apis.getTicketList.mockRejectedValue(new Error('network'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    await flushPromises()
    expect(errSpy).toHaveBeenCalled()
    expect(w.vm.loading).toBe(false)
    errSpy.mockRestore()
  })
})

describe('SupportTicketsView.vue — 处理工单', () => {
  it('openProcessDialog：记录当前工单并清空回复框', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.openProcessDialog({ id: 9, ticketNo: 'T-9' })
    expect(w.vm.currentTicket).toEqual({ id: 9, ticketNo: 'T-9' })
    expect(w.vm.processForm).toEqual({ id: 9, adminReply: '' })
    expect(w.vm.processDialogVisible).toBe(true)
  })

  it('回复为空白：直接返回，不发请求也不提示', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.processForm = { id: 1, adminReply: '   ' }
    await w.vm.handleProcess()
    expect(apis.processTicket).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('处理成功：提示、关弹窗、清空当前工单并刷新列表', async () => {
    apis.processTicket.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.processForm = { id: 4, adminReply: '已处理' }
    w.vm.processDialogVisible = true
    w.vm.currentTicket = { id: 4 }
    apis.getTicketList.mockClear()
    await w.vm.handleProcess()
    expect(apis.processTicket).toHaveBeenCalledWith(4, '已处理')
    expect(messageMocks.success).toHaveBeenCalledWith('工单已处理')
    expect(w.vm.processDialogVisible).toBe(false)
    expect(w.vm.currentTicket).toBeNull()
    expect(apis.getTicketList).toHaveBeenCalled()
  })

  it('处理业务失败：展示后端 message，保留弹窗与输入', async () => {
    apis.processTicket.mockResolvedValue({ code: 500, message: '工单已被他人处理' })
    const w = makeWrapper()
    await flushPromises()
    w.vm.processForm = { id: 4, adminReply: '回复' }
    w.vm.processDialogVisible = true
    apis.getTicketList.mockClear()
    await w.vm.handleProcess()
    expect(messageMocks.error).toHaveBeenCalledWith('工单已被他人处理')
    expect(w.vm.processDialogVisible).toBe(true)
    expect(w.vm.processForm.adminReply).toBe('回复')
    expect(apis.getTicketList).not.toHaveBeenCalled()
  })

  it('处理业务失败且无 message：回落通用文案', async () => {
    apis.processTicket.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.processForm = { id: 4, adminReply: '回复' }
    await w.vm.handleProcess()
    expect(messageMocks.error).toHaveBeenCalledWith('处理工单失败')
  })

  it('处理抛异常：提示失败', async () => {
    apis.processTicket.mockRejectedValue(new Error('boom'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    await flushPromises()
    w.vm.processForm = { id: 4, adminReply: '回复' }
    await w.vm.handleProcess()
    expect(messageMocks.error).toHaveBeenCalledWith('处理工单失败')
    errSpy.mockRestore()
  })
})

describe('SupportTicketsView.vue — 删除工单', () => {
  it('确认文案带上工单号', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteTicket.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3, ticketNo: 'T-3' })
    expect(confirmMock.mock.calls[0][0]).toContain('T-3')
  })

  it('工单号缺失时确认文案回落到 id', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteTicket.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3, ticketNo: '' })
    expect(confirmMock.mock.calls[0][0]).toContain('3')
  })

  it('删除成功：提示并刷新', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteTicket.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    apis.getTicketList.mockClear()
    await w.vm.handleDelete({ id: 3 })
    expect(apis.deleteTicket).toHaveBeenCalledWith(3)
    expect(messageMocks.success).toHaveBeenCalledWith('工单已删除')
    expect(apis.getTicketList).toHaveBeenCalled()
  })

  it('取消删除：不调接口、不报错', async () => {
    confirmMock.mockRejectedValue('cancel')
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3 })
    expect(apis.deleteTicket).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('删除业务失败：展示后端 message，不刷新', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteTicket.mockResolvedValue({ code: 500, message: '工单处理中，暂不可删' })
    const w = makeWrapper()
    await flushPromises()
    apis.getTicketList.mockClear()
    await w.vm.handleDelete({ id: 3 })
    expect(messageMocks.error).toHaveBeenCalledWith('工单处理中，暂不可删')
    expect(apis.getTicketList).not.toHaveBeenCalled()
  })

  it('删除抛异常：提示失败（区别于 cancel）', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteTicket.mockRejectedValue(new Error('network'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3 })
    expect(messageMocks.error).toHaveBeenCalledWith('删除工单失败')
    errSpy.mockRestore()
  })
})

describe('SupportTicketsView.vue — 展示映射', () => {
  it('状态 / 来源 / 分类 / 优先级映射命中与兜底', () => {
    const w = makeWrapper()
    expect(w.vm.getStatusLabel('PENDING')).toBe('待处理')
    expect(w.vm.getStatusType('PROCESSED')).toBe('success')
    expect(w.vm.getStatusLabel('???')).toBe('???')
    expect(w.vm.getStatusLabel('')).toBe('-')
    expect(w.vm.getStatusType('???')).toBe('info')

    expect(w.vm.getSourceLabel('AI_CUSTOMER_SERVICE')).toBe('AI客服')
    expect(w.vm.getSourceLabel('???')).toBe('???')
    expect(w.vm.getSourceLabel('')).toBe('-')

    expect(w.vm.getCategoryLabel('CONTENT_REVIEW')).toBe('内容审核')
    expect(w.vm.getCategoryLabel('???')).toBe('???')

    expect(w.vm.getPriorityLabel('URGENT')).toBe('紧急')
    expect(w.vm.getPriorityType('URGENT')).toBe('danger')
    expect(w.vm.getPriorityType('???')).toBe('info')
    expect(w.vm.getPriorityLabel('')).toBe('-')
  })

  it('formatTime 空值占位', () => {
    const w = makeWrapper()
    expect(w.vm.formatTime('')).toBe('-')
  })
})
