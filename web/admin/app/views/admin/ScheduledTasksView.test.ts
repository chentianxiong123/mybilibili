import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ScheduledTasksView from './ScheduledTasksView.vue'

const apis = vi.hoisted(() => ({
  getScheduledTasks: vi.fn(),
  createScheduledTask: vi.fn(),
  updateScheduledTask: vi.fn(),
  toggleScheduledTask: vi.fn(),
  triggerScheduledTask: vi.fn(),
  deleteScheduledTask: vi.fn(),
}))
vi.mock('@/api/scheduledTask', () => apis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({ ElMessage: messageMocks }))

vi.mock('@element-plus/icons-vue', () => ({
  Plus: { template: '<i class="plus-stub" />' },
  Refresh: { template: '<i class="refresh-stub" />' },
}))

const emptyForm = () => ({
  taskKey: '',
  taskName: '',
  description: '',
  cronExpr: '',
  taskType: 'hot_search_cleanup',
  taskConfig: '',
  timeoutSeconds: 30,
  maxRetries: 3,
})

const makeWrapper = () =>
  mount(ScheduledTasksView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
        },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-switch': {
          template:
            '<input type="checkbox" class="el-switch-stub" :checked="modelValue" @change="e=>$emit(`update:modelValue`, e.target.checked)" />',
          props: ['modelValue'],
        },
        'el-dialog': { template: '<div class="el-dialog-stub"><slot /></div>' },
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-input': {
          template:
            '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(`update:modelValue`, e.target.value)" />',
          props: ['modelValue'],
        },
        'el-input-number': {
          template:
            '<input class="el-input-number-stub" :value="modelValue" @input="e=>$emit(`update:modelValue`, Number(e.target.value))" />',
          props: ['modelValue'],
        },
        'el-select': { template: '<div class="el-select-stub"><slot /></div>' },
        'el-option': { template: '<div class="el-option-stub" />' },
        'el-popconfirm': { template: '<div class="el-popconfirm-stub"><slot /></div>' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  for (const fn of Object.values(apis)) fn.mockReset()
  apis.getScheduledTasks.mockResolvedValue({ code: 200, data: { list: [] } })
})

afterEach(() => {
  vi.useRealTimers()
})

describe('ScheduledTasksView.vue', () => {
  it('onMounted 拉取任务列表', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getScheduledTasks).toHaveBeenCalledTimes(1)
  })

  it('list 缺失时回落空数组', async () => {
    apis.getScheduledTasks.mockResolvedValue({ code: 200, data: {} })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.loading).toBe(false)
  })

  it('拉取失败：提示错误并复位 loading', async () => {
    apis.getScheduledTasks.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取任务列表失败')
    expect(w.vm.loading).toBe(false)
  })

  it('启停：已启用(1) 提交 0，本地同步翻转', async () => {
    apis.toggleScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    const row = { id: 1, enabled: 1 }
    await w.vm.handleToggle(row)
    expect(apis.toggleScheduledTask).toHaveBeenCalledWith(1, 0)
    expect(row.enabled).toBe(0)
    expect(w.vm.togglingId).toBeNull()
  })

  it('启停：已停用(0) 提交 1', async () => {
    apis.toggleScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    const row = { id: 2, enabled: 0 }
    await w.vm.handleToggle(row)
    expect(apis.toggleScheduledTask).toHaveBeenCalledWith(2, 1)
    expect(row.enabled).toBe(1)
  })

  it('启停失败：提示操作失败且不翻转本地状态', async () => {
    apis.toggleScheduledTask.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    const row = { id: 3, enabled: 1 }
    await w.vm.handleToggle(row)
    expect(messageMocks.error).toHaveBeenCalledWith('操作失败')
    expect(row.enabled).toBe(1)
    expect(w.vm.togglingId).toBeNull()
  })

  it('启停业务失败（非 200）：不翻转也不提示', async () => {
    apis.toggleScheduledTask.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    const row = { id: 4, enabled: 1 }
    await w.vm.handleToggle(row)
    expect(row.enabled).toBe(1)
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('手动触发：按 taskKey 提交并在成功后延迟刷新列表', async () => {
    vi.useFakeTimers()
    apis.triggerScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    apis.getScheduledTasks.mockClear()
    await w.vm.handleTrigger({ id: 9, taskKey: 'hot_search' })
    expect(apis.triggerScheduledTask).toHaveBeenCalledWith('hot_search')
    expect(messageMocks.success).toHaveBeenCalledWith('任务已触发')
    expect(apis.getScheduledTasks).not.toHaveBeenCalled()
    vi.advanceTimersByTime(2000)
    expect(apis.getScheduledTasks).toHaveBeenCalledTimes(1)
  })

  it('手动触发失败：提示触发失败且不排刷新', async () => {
    vi.useFakeTimers()
    apis.triggerScheduledTask.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    apis.getScheduledTasks.mockClear()
    await w.vm.handleTrigger({ id: 9, taskKey: 'hot_search' })
    expect(messageMocks.error).toHaveBeenCalledWith('触发失败')
    vi.advanceTimersByTime(2000)
    expect(apis.getScheduledTasks).not.toHaveBeenCalled()
  })

  it('编辑：把行数据搬进表单并打开弹窗', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleEdit({
      id: 11,
      taskKey: 'k',
      taskName: 'n',
      description: 'd',
      cronExpr: '0 * * * *',
      taskType: 'tt',
      taskConfig: 'cfg',
      timeoutSeconds: 60,
      maxRetries: 5,
    })
    expect(w.vm.editingId).toBe(11)
    expect(w.vm.showCreateDialog).toBe(true)
    expect(w.vm.form).toEqual({
      taskKey: 'k',
      taskName: 'n',
      description: 'd',
      cronExpr: '0 * * * *',
      taskType: 'tt',
      taskConfig: 'cfg',
      timeoutSeconds: 60,
      maxRetries: 5,
    })
  })

  it('删除成功：提示并重新拉取列表', async () => {
    apis.deleteScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    apis.getScheduledTasks.mockClear()
    await w.vm.handleDelete(7)
    expect(apis.deleteScheduledTask).toHaveBeenCalledWith(7)
    expect(messageMocks.success).toHaveBeenCalledWith('已删除')
    expect(apis.getScheduledTasks).toHaveBeenCalled()
  })

  it('删除失败：提示删除失败', async () => {
    apis.deleteScheduledTask.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete(7)
    expect(messageMocks.error).toHaveBeenCalledWith('删除失败')
  })

  it('新建：无 editingId 时走 create，成功后重置表单', async () => {
    apis.createScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.form = { ...emptyForm(), taskKey: 'nk', taskName: 'nn' }
    await w.vm.handleSubmit()
    expect(apis.createScheduledTask).toHaveBeenCalledWith({ ...emptyForm(), taskKey: 'nk', taskName: 'nn' })
    expect(apis.updateScheduledTask).not.toHaveBeenCalled()
    expect(messageMocks.success).toHaveBeenCalledWith('已创建')
    expect(w.vm.showCreateDialog).toBe(false)
    expect(w.vm.editingId).toBeNull()
    expect(w.vm.form).toEqual(emptyForm())
    expect(w.vm.submitting).toBe(false)
  })

  it('编辑保存：有 editingId 时走 update 并带上 id', async () => {
    apis.updateScheduledTask.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.editingId = 21
    w.vm.showCreateDialog = true
    w.vm.form = { ...emptyForm(), taskKey: 'ek' }
    await w.vm.handleSubmit()
    expect(apis.updateScheduledTask).toHaveBeenCalledWith({ id: 21, ...emptyForm(), taskKey: 'ek' })
    expect(apis.createScheduledTask).not.toHaveBeenCalled()
    expect(messageMocks.success).toHaveBeenCalledWith('已更新')
    expect(w.vm.editingId).toBeNull()
    expect(w.vm.form).toEqual(emptyForm())
  })

  it('提交失败：提示提交失败，保留弹窗与表单内容', async () => {
    apis.createScheduledTask.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    w.vm.showCreateDialog = true
    w.vm.form = { ...emptyForm(), taskKey: 'keep' }
    await w.vm.handleSubmit()
    expect(messageMocks.error).toHaveBeenCalledWith('提交失败')
    expect(w.vm.showCreateDialog).toBe(true)
    expect(w.vm.form.taskKey).toBe('keep')
    expect(w.vm.submitting).toBe(false)
  })
})
