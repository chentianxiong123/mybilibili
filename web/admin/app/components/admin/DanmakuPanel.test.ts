import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import DanmakuPanel from './DanmakuPanel.vue'

vi.mock('@/api/adminContent', () => ({
  getAdminDanmaku: vi.fn(),
  deleteAdminDanmaku: vi.fn(),
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

import { getAdminDanmaku, deleteAdminDanmaku } from '@/api/adminContent'

const getAdminDanmakuMock = getAdminDanmaku as unknown as ReturnType<typeof vi.fn>
const deleteAdminDanmakuMock = deleteAdminDanmaku as unknown as ReturnType<typeof vi.fn>

function makeWrapper() {
  return mount(DanmakuPanel, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-input': { template: '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(\'update:modelValue\', e.target.value)" @keyup.enter="$emit(\'keyup:enter\')" />', props: ['modelValue'] },
        'el-button': { template: '<button class="el-button-stub" @click="$emit(\'click\')"><slot /></button>' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': { template: '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>' },
        'el-avatar': { template: '<span class="el-avatar-stub"><slot /></span>' },
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-pagination': { template: '<div class="el-pagination-stub" />' },
      },
    },
  })
}

beforeEach(() => {
  vi.clearAllMocks()
  getAdminDanmakuMock.mockReset()
  deleteAdminDanmakuMock.mockReset()
  confirmMock.mockReset()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
})

describe('DanmakuPanel.vue', () => {
  it('onMounted 加载弹幕列表并写入 tableData / total', async () => {
    getAdminDanmakuMock.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 9, content: '前方高能', userName: '弹幕侠', time: 65 }], total: 1 },
    })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.$nextTick()
    expect(getAdminDanmakuMock).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, size: 10 })
    )
    expect(w.vm.tableData).toHaveLength(1)
    expect(w.vm.tableData[0].content).toBe('前方高能')
    expect(w.vm.total).toBe(1)
  })

  it('搜索时重置页码并携带 keyword', async () => {
    getAdminDanmakuMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.find('.el-input-stub').setValue('高能')
    await w.find('.el-button-stub').trigger('click')
    expect(getAdminDanmakuMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ keyword: '高能', page: 1 })
    )
  })

  it('删除：确认后调用 delete 并提示成功', async () => {
    confirmMock.mockResolvedValueOnce('confirm')
    deleteAdminDanmakuMock.mockResolvedValue({ code: 200, data: null })
    getAdminDanmakuMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })

    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleDelete({ id: 12 })

    expect(confirmMock).toHaveBeenCalled()
    expect(deleteAdminDanmakuMock).toHaveBeenCalledWith(12)
    expect(messageMocks.success).toHaveBeenCalledWith('删除成功')
  })

  it('删除：取消确认时不调用 delete', async () => {
    getAdminDanmakuMock.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
    confirmMock.mockRejectedValueOnce('cancel')
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleDelete({ id: 12 })
    expect(deleteAdminDanmakuMock).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('删除：接口失败时提示失败消息', async () => {
    confirmMock.mockResolvedValueOnce('confirm')
    deleteAdminDanmakuMock.mockResolvedValue({ code: 500, message: '删除失败', data: null })
    const w = makeWrapper()
    await w.vm.$nextTick()
    await w.vm.handleDelete({ id: 12 })
    expect(messageMocks.error).toHaveBeenCalledWith('删除失败')
  })

  it('formatTime 将秒转为 mm:ss', () => {
    const w = makeWrapper()
    expect(w.vm.formatTime(65)).toBe('01:05')
    expect(w.vm.formatTime(0)).toBe('00:00')
    expect(w.vm.formatTime(null)).toBe('00:00')
  })

  it('formatDateTime 对空值返回 "-"', () => {
    const w = makeWrapper()
    expect(w.vm.formatDateTime('')).toBe('-')
    expect(w.vm.formatDateTime('2024-01-02T03:04:05')).toBe('2024-01-02 03:04')
  })
})