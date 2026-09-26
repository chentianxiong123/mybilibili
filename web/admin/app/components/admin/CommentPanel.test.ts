import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import CommentPanel from './CommentPanel.vue'

vi.mock('@/api/adminContent', () => ({
  getAdminComments: vi.fn(),
  deleteAdminComment: vi.fn(),
  restoreAdminComment: vi.fn(),
}))

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('element-plus', async () => {
  const actual = await vi.importActual<any>('element-plus')
  return {
    ...actual,
    ElMessage: { error: messageMocks.error, success: messageMocks.success, warning: vi.fn() },
    ElMessageBox: { confirm: confirmMock },
  }
})

import { getAdminComments, deleteAdminComment, restoreAdminComment } from '@/api/adminContent'
import { ElMessageBox } from 'element-plus'

const getAdminCommentsMock = getAdminComments as unknown as ReturnType<typeof vi.fn>
const deleteAdminCommentMock = deleteAdminComment as unknown as ReturnType<typeof vi.fn>
const restoreAdminCommentMock = restoreAdminComment as unknown as ReturnType<typeof vi.fn>

function makeWrapper() {
  return mount(CommentPanel, {
    global: {
      stubs: {
        'el-select': { template: '<div class="el-select-stub"><slot /></div>' },
        'el-option': { template: '<div class="el-option-stub"><slot /></div>' },
        'el-input': { template: '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(\'update:modelValue\', e.target.value)" @keyup.enter="$emit(\'keyup:enter\')" />', props: ['modelValue'] },
        'el-button': { template: '<button class="el-button-stub" @click="$emit(\'click\')"><slot /></button>' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': { template: '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>' },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-avatar': { template: '<span class="el-avatar-stub"><slot /></span>' },
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-pagination': { template: '<div class="el-pagination-stub" />' },
      },
    },
  })
}

beforeEach(() => {
  vi.clearAllMocks()
  getAdminCommentsMock.mockReset()
  deleteAdminCommentMock.mockReset()
  restoreAdminCommentMock.mockReset()
  confirmMock.mockReset()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
})

describe('CommentPanel.vue', () => {
  it('onMounted 加载评论列表并写入 tableData', async () => {
    getAdminCommentsMock.mockResolvedValue({
      code: 200,
      data: {
        list: [{ id: 1, type: 'comment', content: '好看', userName: '小明', status: '0', createdAt: '2024-01-02T03:04:05' }],
        total: 1,
      },
    })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.$nextTick()
    expect(getAdminCommentsMock).toHaveBeenCalled()
    expect(w.vm.tableData).toHaveLength(1)
    expect(w.vm.tableData[0].content).toBe('好看')
    expect(w.vm.tableData[0].userName).toBe('小明')
    expect(w.vm.total).toBe(1)
  })

  it('搜索时重置页码并重新加载', async () => {
    getAdminCommentsMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.find('.el-input-stub').setValue('关键')
    await w.find('.el-button-stub').trigger('click')
    expect(getAdminCommentsMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ keyword: '关键', page: 1 })
    )
  })

  it('切换类型后重置状态筛选并加载 reply 类型', async () => {
    getAdminCommentsMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
    const w = makeWrapper()
    await w.vm.$nextTick()
    w.vm.targetType = 'reply'
    await w.vm.handleTypeChange()
    expect(getAdminCommentsMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ type: 'reply', status: undefined })
    )
  })

  it('下架：确认后调用 delete 并刷新', async () => {
    confirmMock.mockResolvedValueOnce('confirm')
    deleteAdminCommentMock.mockResolvedValue({ code: 200, data: null })
    getAdminCommentsMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })

    const w = makeWrapper()
    await w.vm.$nextTick()
    w.vm.targetType = 'comment'
    getAdminCommentsMock.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 5, type: 'comment', content: 'x', status: '0' }], total: 1 },
    })
    await w.vm.handleDelete({ id: 5, type: 'comment' })

    expect(confirmMock).toHaveBeenCalled()
    expect(deleteAdminCommentMock).toHaveBeenCalledWith('comment', 5)
    expect(messageMocks.success).toHaveBeenCalledWith('下架成功')
  })

  it('下架：取消确认时不调 delete', async () => {
    confirmMock.mockRejectedValueOnce('cancel')
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleDelete({ id: 5, type: 'comment' })
    expect(deleteAdminCommentMock).not.toHaveBeenCalled()
  })

  it('下架：接口报错时提示失败', async () => {
    confirmMock.mockResolvedValueOnce('confirm')
    deleteAdminCommentMock.mockResolvedValue({ code: 500, message: '服务不可用', data: null })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleDelete({ id: 5, type: 'comment' })
    expect(messageMocks.error).toHaveBeenCalledWith('服务不可用')
  })

  it('恢复：调用 restore 并刷新', async () => {
    restoreAdminCommentMock.mockResolvedValue({ code: 200, data: null })
    getAdminCommentsMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })

    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleRestore({ id: 3, type: 'reply' })

    expect(restoreAdminCommentMock).toHaveBeenCalledWith('reply', 3)
    expect(messageMocks.success).toHaveBeenCalledWith('恢复成功')
  })

  it('恢复失败时提示错误', async () => {
    restoreAdminCommentMock.mockResolvedValue({ code: 500, message: '失败', data: null })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleRestore({ id: 3, type: 'reply' })
    expect(messageMocks.error).toHaveBeenCalledWith('失败')
  })

  it('isRemoved 对 reply 用 REMOVED，对 comment 用 1', () => {
    const w = makeWrapper()
    expect(w.vm.isRemoved({ type: 'reply', status: 'REMOVED' })).toBe(true)
    expect(w.vm.isRemoved({ type: 'reply', status: 'NORMAL' })).toBe(false)
    expect(w.vm.isRemoved({ type: 'comment', status: '1' })).toBe(true)
    expect(w.vm.isRemoved({ type: 'comment', status: '0' })).toBe(false)
  })

  it('formatDateTime 对空值返回 "-"，ISO 转为本地分钟格式（组件只到分钟）', () => {
    const w = makeWrapper()
    expect(w.vm.formatDateTime('')).toBe('-')
    expect(w.vm.formatDateTime('2024-01-02T03:04:05')).toBe('2024-01-02 03:04')
  })
})