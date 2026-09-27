import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CategoriesView from './CategoriesView.vue'

const apis = vi.hoisted(() => ({
  getCategoryList: vi.fn(),
  addCategory: vi.fn(),
  updateCategory: vi.fn(),
  deleteCategory: vi.fn(),
}))
vi.mock('@/api/category', () => apis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn(), warning: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
  ElMessageBox: { confirm: confirmMock },
}))

vi.mock('@element-plus/icons-vue', () => ({
  Plus: { template: '<i />' },
  Search: { template: '<i />' },
}))

const makeWrapper = () =>
  mount(CategoriesView, {
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
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-pagination': { template: '<div />' },
        'el-dialog': { template: '<div class="el-dialog-stub"><slot /></div>' },
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
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
  apis.getCategoryList.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
})

describe('CategoriesView.vue', () => {
  it('onMounted 拉取分区列表', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getCategoryList).toHaveBeenCalledWith({ page: 1, size: 10, keyword: '' })
  })

  it('归一化 created_at / createdAt / create_time 三种时间字段', async () => {
    apis.getCategoryList.mockResolvedValue({
      code: 200,
      data: {
        list: [
          { id: 1, name: '动画', created_at: 'A' },
          { id: 2, name: '音乐', createdAt: 'B' },
          { id: 3, name: '游戏', create_time: 'C' },
          { id: 4 },
        ],
        total: 4,
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData.map((r: any) => r.createdAt)).toEqual(['A', 'B', 'C', ''])
    expect(w.vm.tableData[3].name).toBe('')
    expect(w.vm.total).toBe(4)
  })

  it('兼容 data 直接是数组的响应形状', async () => {
    apis.getCategoryList.mockResolvedValue({ code: 200, data: [{ id: 1, name: 'x' }] })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toHaveLength(1)
  })

  it('兼容 success 标记 + 根级 total', async () => {
    apis.getCategoryList.mockResolvedValue({ success: true, data: [{ id: 1 }], total: 6 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.total).toBe(6)
  })

  it('data 不是数组时列表置空，不崩', async () => {
    apis.getCategoryList.mockResolvedValue({ code: 200, data: { list: 'not-an-array' } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
  })

  it('业务失败：不覆盖列表', async () => {
    apis.getCategoryList.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.loading).toBe(false)
  })

  it('拉取失败：提示错误并复位 loading', async () => {
    apis.getCategoryList.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取分区列表失败')
    expect(w.vm.loading).toBe(false)
  })

  it('搜索：带 keyword 并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentPage = 5
    w.vm.keyword = '动画'
    w.vm.handleSearch()
    await flushPromises()
    expect(w.vm.currentPage).toBe(1)
    expect(apis.getCategoryList).toHaveBeenLastCalledWith({ page: 1, size: 10, keyword: '动画' })
  })

  it('重置：清空 keyword 并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentPage = 3
    w.vm.keyword = 'x'
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.keyword).toBe('')
    expect(w.vm.currentPage).toBe(1)
  })

  it('翻页按新页码查询', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handlePageChange(4)
    await flushPromises()
    expect(apis.getCategoryList).toHaveBeenLastCalledWith({ page: 4, size: 10, keyword: '' })
  })

  it('handleAdd：打开弹窗并重置表单', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogForm.name = '脏数据'
    w.vm.handleAdd()
    expect(w.vm.dialogTitle).toBe('添加分区')
    expect(w.vm.dialogVisible).toBe(true)
    expect(w.vm.dialogForm).toEqual({ id: null, name: '' })
  })

  it('handleEdit：回填行数据并切标题', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleEdit({ id: 7, name: '番剧' })
    expect(w.vm.dialogTitle).toBe('编辑分区')
    expect(w.vm.dialogForm).toEqual({ id: 7, name: '番剧' })
    expect(w.vm.dialogVisible).toBe(true)
  })

  it('无 dialogFormRef 时保存直接返回', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = null
    await w.vm.handleSave()
    expect(apis.addCategory).not.toHaveBeenCalled()
  })

  it('校验不通过：不发请求、不关弹窗', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(false) }
    w.vm.dialogVisible = true
    await w.vm.handleSave()
    await flushPromises()
    expect(apis.addCategory).not.toHaveBeenCalled()
    expect(w.vm.dialogVisible).toBe(true)
  })

  it('新增成功：只提交 name，提示并关弹窗刷新', async () => {
    apis.addCategory.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogForm = { id: null, name: '新分区' }
    w.vm.dialogVisible = true
    apis.getCategoryList.mockClear()
    await w.vm.handleSave()
    await flushPromises()
    expect(apis.addCategory).toHaveBeenCalledWith({ name: '新分区' })
    expect(apis.updateCategory).not.toHaveBeenCalled()
    expect(messageMocks.success).toHaveBeenCalledWith('添加成功')
    expect(w.vm.dialogVisible).toBe(false)
    expect(apis.getCategoryList).toHaveBeenCalled()
  })

  it('编辑成功：走 update 并带上 id', async () => {
    apis.updateCategory.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogForm = { id: 3, name: '改名' }
    await w.vm.handleSave()
    await flushPromises()
    expect(apis.updateCategory).toHaveBeenCalledWith(3, { name: '改名' })
    expect(messageMocks.success).toHaveBeenCalledWith('更新成功')
  })

  it('保存抛异常：提示失败、保留弹窗、复位 loading', async () => {
    apis.addCategory.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogVisible = true
    await w.vm.handleSave()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('操作失败')
    expect(w.vm.dialogVisible).toBe(true)
    expect(w.vm.loading).toBe(false)
  })

  it('删除成功：确认后调接口、提示并刷新', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteCategory.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    apis.getCategoryList.mockClear()
    await w.vm.handleDelete({ id: 2, name: '音乐' })
    expect(confirmMock.mock.calls[0][0]).toContain('不可恢复')
    expect(apis.deleteCategory).toHaveBeenCalledWith(2)
    expect(messageMocks.success).toHaveBeenCalledWith('删除成功')
    expect(apis.getCategoryList).toHaveBeenCalled()
  })

  it('取消删除：不调接口、不报错', async () => {
    confirmMock.mockRejectedValue('cancel')
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 2 })
    expect(apis.deleteCategory).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('删除业务失败：展示后端 message，不刷新列表', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteCategory.mockResolvedValue({ code: 500, message: '分区下还有稿件' })
    const w = makeWrapper()
    await flushPromises()
    apis.getCategoryList.mockClear()
    await w.vm.handleDelete({ id: 2 })
    expect(messageMocks.error).toHaveBeenCalledWith('分区下还有稿件')
    expect(apis.getCategoryList).not.toHaveBeenCalled()
  })

  it('删除接口抛错：必须提示用户，不能静默吞掉', async () => {
    confirmMock.mockResolvedValue('confirm')
    apis.deleteCategory.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 2 })
    expect(messageMocks.error).toHaveBeenCalledWith('删除失败')
  })
})
