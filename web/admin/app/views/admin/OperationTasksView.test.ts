import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import OperationTasksView from './OperationTasksView.vue'

const apis = vi.hoisted(() => ({ getOperationTasks: vi.fn(), getOperationTaskDetail: vi.fn() }))
vi.mock('@/api/operationTask', () => apis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({ ElMessage: messageMocks }))

vi.mock('@element-plus/icons-vue', () => ({
  Refresh: { template: '<i />' },
  Search: { template: '<i />' },
}))

const emptySearch = () => ({
  taskType: '',
  status: '',
  keyword: '',
  targetKeyword: '',
  startTime: '',
  endTime: '',
})

const makeWrapper = () =>
  mount(OperationTasksView, {
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
        'el-progress': { template: '<div class="el-progress-stub" />' },
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
  apis.getOperationTasks.mockReset()
  apis.getOperationTaskDetail.mockReset()
  apis.getOperationTasks.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
})

describe('OperationTasksView.vue', () => {
  it('setup 阶段即拉取任务列表', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getOperationTasks).toHaveBeenCalledTimes(1)
    expect(apis.getOperationTasks).toHaveBeenCalledWith({ page: 1, size: 10, ...emptySearch() })
  })

  it('snake_case 字段归一化', async () => {
    apis.getOperationTasks.mockResolvedValue({
      code: 200,
      data: {
        list: [
          {
            id: 1,
            task_key: 'k',
            task_type: 'VIDEO_PROCESS',
            task_name: '视频处理',
            target_type: 'manuscript',
            target_id: 8,
            status: 'RUNNING',
            progress: 45,
            stage: 'transcode',
            message: 'm',
            error_message: '',
            operator_id: 2,
            operator_name: 'admin',
            started_at: 'S',
            finished_at: '',
            created_at: 'C',
            updated_at: 'U',
          },
        ],
        total: 1,
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0]).toEqual({
      id: 1,
      taskKey: 'k',
      taskType: 'VIDEO_PROCESS',
      taskName: '视频处理',
      targetType: 'manuscript',
      targetId: 8,
      status: 'RUNNING',
      progress: 45,
      stage: 'transcode',
      message: 'm',
      errorMessage: '',
      operatorId: 2,
      operatorName: 'admin',
      startedAt: 'S',
      finishedAt: '',
      createdAt: 'C',
      updatedAt: 'U',
    })
  })

  it('camelCase 字段也可识别，progress 为 0 时不被 ?? 吞掉', async () => {
    apis.getOperationTasks.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 2, taskKey: 'k2', progress: 0, operatorId: 0, targetId: 0 }] },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0].progress).toBe(0)
    expect(w.vm.tableData[0].operatorId).toBe(0)
    expect(w.vm.tableData[0].targetId).toBe(0)
  })

  it('progress 缺失时兜底 0', async () => {
    apis.getOperationTasks.mockResolvedValue({ code: 200, data: { list: [{ id: 3 }] } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0].progress).toBe(0)
  })

  it('业务失败：提示后端 message，不覆盖列表', async () => {
    apis.getOperationTasks.mockResolvedValue({ code: 500, message: '任务服务不可用' })
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('任务服务不可用')
    expect(w.vm.tableData).toEqual([])
  })

  it('业务失败且无 message：回落通用文案', async () => {
    apis.getOperationTasks.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('加载任务失败')
  })

  it('请求异常：提示错误、复位 loading，不产生未捕获拒绝', async () => {
    apis.getOperationTasks.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('加载任务失败')
    expect(w.vm.loading).toBe(false)
  })

  it('搜索：带条件并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 7
    Object.assign(w.vm.searchForm, {
      taskType: 'VIDEO_PROCESS',
      status: 'FAILED',
      keyword: 'k',
      targetKeyword: 't',
      startTime: 's',
      endTime: 'e',
    })
    w.vm.handleSearch()
    await flushPromises()
    expect(w.vm.page).toBe(1)
    expect(apis.getOperationTasks).toHaveBeenLastCalledWith({
      page: 1,
      size: 10,
      taskType: 'VIDEO_PROCESS',
      status: 'FAILED',
      keyword: 'k',
      targetKeyword: 't',
      startTime: 's',
      endTime: 'e',
    })
  })

  it('重置：清空全部条件并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 2
    Object.assign(w.vm.searchForm, { taskType: 'a', status: 'b', keyword: 'c', targetKeyword: 'd', startTime: 'e', endTime: 'f' })
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.searchForm).toEqual(emptySearch())
    expect(w.vm.page).toBe(1)
  })

  it('翻页保留筛选条件', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.searchForm.status = 'RUNNING'
    w.vm.handlePageChange(3)
    await flushPromises()
    expect(apis.getOperationTasks).toHaveBeenLastCalledWith(expect.objectContaining({ page: 3, status: 'RUNNING' }))
  })

  it('改每页条数同时重置回第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.page = 8
    w.vm.handleSizeChange(20)
    await flushPromises()
    expect(w.vm.page).toBe(1)
    expect(apis.getOperationTasks).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, size: 20 }))
  })

  it('查看详情：按 id 拉取并归一化', async () => {
    apis.getOperationTaskDetail.mockResolvedValue({ code: 200, data: { id: 5, status: 'SUCCESS', progress: 100 } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.showDetail({ id: 5 })
    expect(apis.getOperationTaskDetail).toHaveBeenCalledWith(5)
    expect(w.vm.detailVisible).toBe(true)
    expect(w.vm.currentDetail.status).toBe('SUCCESS')
    expect(w.vm.detailLoading).toBe(false)
  })

  it('查看详情业务失败：清空旧详情并提示', async () => {
    apis.getOperationTaskDetail.mockResolvedValue({ code: 404, message: '任务已删除' })
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentDetail = { id: 99 }
    await w.vm.showDetail({ id: 5 })
    expect(w.vm.currentDetail).toBeNull()
    expect(messageMocks.error).toHaveBeenCalledWith('任务已删除')
  })

  it('查看详情抛异常：同样要提示', async () => {
    apis.getOperationTaskDetail.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.showDetail({ id: 5 })
    expect(messageMocks.error).toHaveBeenCalledWith('加载任务详情失败')
    expect(w.vm.detailLoading).toBe(false)
  })

  it('normalizeProgress：非数字与 null 归 0，并夹在 0~100', () => {
    const w = makeWrapper()
    expect(w.vm.normalizeProgress(null)).toBe(0)
    expect(w.vm.normalizeProgress(undefined)).toBe(0)
    expect(w.vm.normalizeProgress('abc')).toBe(0)
    expect(w.vm.normalizeProgress(-20)).toBe(0)
    expect(w.vm.normalizeProgress(150)).toBe(100)
    expect(w.vm.normalizeProgress('60')).toBe(60)
  })

  it('formatTaskType / formatStatus / statusTag / taskTypeTag：命中映射，未知兜底', () => {
    const w = makeWrapper()
    expect(w.vm.formatTaskType('UPLOAD')).toBe('上传稿件')
    expect(w.vm.formatTaskType('AI_PIPELINE')).toBe('AI任务')
    expect(w.vm.formatTaskType('UNKNOWN')).toBe('UNKNOWN')
    expect(w.vm.formatTaskType('')).toBe('-')

    expect(w.vm.formatStatus('PENDING')).toBe('待处理')
    expect(w.vm.formatStatus('CANCELLED')).toBe('已取消')
    expect(w.vm.formatStatus('')).toBe('-')

    expect(w.vm.statusTag('RUNNING')).toBe('warning')
    expect(w.vm.statusTag('SUCCESS')).toBe('success')
    expect(w.vm.statusTag('FAILED')).toBe('danger')
    expect(w.vm.statusTag('???')).toBe('info')

    expect(w.vm.taskTypeTag('UPLOAD')).toBe('primary')
    expect(w.vm.taskTypeTag('???')).toBe('info')
  })

  it('progressStatus：仅失败/成功有特殊态，其余空', () => {
    const w = makeWrapper()
    expect(w.vm.progressStatus('FAILED')).toBe('exception')
    expect(w.vm.progressStatus('SUCCESS')).toBe('success')
    expect(w.vm.progressStatus('RUNNING')).toBe('')
    expect(w.vm.progressStatus('PENDING')).toBe('')
  })

  it('formatTime 空值占位', () => {
    const w = makeWrapper()
    expect(w.vm.formatTime('')).toBe('-')
  })
})
