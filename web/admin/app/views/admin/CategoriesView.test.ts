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
  ElTag: { template: '<span class="el-tag-stub"><slot /></span>', props: ['type', 'size'] },
  ElSelect: { template: '<select><slot /></select>' },
  ElOption: { template: '<option />', props: ['value', 'label', 'disabled'] },
  ElInputNumber: { template: '<input />', props: ['modelValue'] },
}))

vi.mock('@element-plus/icons-vue', () => ({
  Plus: { template: '<i />' },
  Search: { template: '<i />' },
  Refresh: { template: '<i />' },
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
  it('onMounted 拉取全量分区列表（树形管理不需要分页）', async () => {
    makeWrapper()
    await flushPromises()
    expect(apis.getCategoryList).toHaveBeenCalledWith({})
  })

  it('构建树：一级为根、二级挂到 parent_id 下', async () => {
    apis.getCategoryList.mockResolvedValue({
      code: 200,
      data: [
        { id: 1, name: '科技', parent_id: 0, sort_order: 0 },
        { id: 95, name: '计算机技术', parent_id: 188, sort_order: 0 },
        { id: 188, name: '科技(主)', parent_id: 0, sort_order: 1 },
      ],
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.rawCategories).toHaveLength(3)
    expect(w.vm.groups.parents.map(p => p.id)).toEqual([1, 188])
    const techParent = w.vm.categoryTree.find(c => c.id === 188)
    expect(techParent.children.map(c => c.id)).toEqual([95])
  })

  it('兼容 parentId 驼峰与 success 标记响应形状', async () => {
    apis.getCategoryList.mockResolvedValue({
      success: true,
      data: [{ id: 1, name: '动画', parentId: 0, sortOrder: 2 }],
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.rawCategories[0]).toEqual({ id: 1, name: '动画', parent_id: 0, sort_order: 2 })
  })

  it('搜索命中保留父节点（子分区命中显示其父）', async () => {
    apis.getCategoryList.mockResolvedValue({
      code: 200,
      data: [
        { id: 1, name: '科技', parent_id: 0, sort_order: 0 },
        { id: 95, name: '计算机技术', parent_id: 188, sort_order: 0 },
        { id: 188, name: '科技(主)', parent_id: 0, sort_order: 1 },
      ],
    })
    const w = makeWrapper()
    await flushPromises()
    w.vm.keyword = '计算机'
    const filtered = w.vm.filteredTree
    expect(filtered).toHaveLength(1)
    expect(filtered[0].id).toBe(188)
    expect(filtered[0].children.map(c => c.id)).toEqual([95])
  })

  it('重置：清空 keyword（前端过滤，不再分页请求）', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.keyword = 'x'
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.keyword).toBe('')
  })

  it('handleAdd：打开弹窗并初始化带 parent_id/sort_order 的表单', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleAdd()
    expect(w.vm.dialogTitle).toBe('添加分区')
    expect(w.vm.dialogVisible).toBe(true)
    expect(w.vm.dialogForm).toMatchObject({ id: null, name: '', parent_id: 0 })
  })

  it('handleEdit：回填行数据含父分区与排序', async () => {
    apis.getCategoryList.mockResolvedValue({
      code: 200,
      data: [{ id: 7, name: '番剧', parent_id: 0, sort_order: 3 }],
    })
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleEdit({ id: 7, name: '番剧', parent_id: 0, sort_order: 3 })
    expect(w.vm.dialogTitle).toBe('编辑分区')
    expect(w.vm.dialogForm).toEqual({ id: 7, name: '番剧', parent_id: 0, sort_order: 3 })
    expect(w.vm.dialogVisible).toBe(true)
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

  it('新增成功：提交 name+parent_id+sort_order，提示并关弹窗刷新', async () => {
    apis.addCategory.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogForm = { id: null, name: '新分区', parent_id: 12, sort_order: 1 }
    w.vm.dialogVisible = true
    apis.getCategoryList.mockClear()
    await w.vm.handleSave()
    await flushPromises()
    expect(apis.addCategory).toHaveBeenCalledWith({ name: '新分区', parent_id: 12, sort_order: 1 })
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
    w.vm.dialogForm = { id: 3, name: '改名', parent_id: 0, sort_order: 0 }
    await w.vm.handleSave()
    await flushPromises()
    expect(apis.updateCategory).toHaveBeenCalledWith(3, { name: '改名', parent_id: 0, sort_order: 0 })
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