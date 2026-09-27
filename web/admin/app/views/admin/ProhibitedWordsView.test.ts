import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ProhibitedWordsView from './ProhibitedWordsView.vue'

const wordApis = vi.hoisted(() => ({
  getProhibitedWordList: vi.fn(),
  addProhibitedWord: vi.fn(),
  updateProhibitedWord: vi.fn(),
  deleteProhibitedWord: vi.fn(),
  batchImportProhibitedWords: vi.fn(),
}))
vi.mock('@/api/prohibitedWord', () => wordApis)

const securityApis = vi.hoisted(() => ({ getSecuritySettings: vi.fn(), updateSecuritySettings: vi.fn() }))
vi.mock('@/api/securitySettings', () => securityApis)

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn(), warning: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
  ElMessageBox: { confirm: confirmMock },
}))

vi.mock('@element-plus/icons-vue', () => ({
  Plus: { template: '<i />' },
  Search: { template: '<i />' },
  Upload: { template: '<i />' },
}))

const emptyDialogForm = () => ({ id: null, word: '', matchType: 'CONTAINS', category: '', isEnabled: 1 })
const emptySecurityForm = () => ({
  commentMaxCount: 10,
  commentWindowSeconds: 60,
  replyMaxCount: 20,
  replyWindowSeconds: 60,
  cacheRefreshIntervalSeconds: 300,
})

const makeWrapper = () =>
  mount(ProhibitedWordsView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
          emits: ['click'],
        },
        'el-card': { template: '<div class="el-card-stub"><slot name="header" /><slot /></div>' },
        'el-row': { template: '<div><slot /></div>' },
        'el-col': { template: '<div><slot /></div>' },
        'el-divider': { template: '<hr />' },
        'el-icon': { template: '<span><slot /></span>' },
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
        'el-option': { template: '<div />' },
        'el-radio-group': { template: '<div><slot /></div>' },
        'el-radio': { template: '<div><slot /></div>' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': {
          template:
            '<div class="el-table-column-stub"><template v-if="$slots.default"><template v-for="row in $parent.$props.data"><slot :row="row" /></template></template></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-dialog': { template: '<div class="el-dialog-stub"><slot /></div>' },
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-upload': { template: '<div class="el-upload-stub" />' },
        'el-pagination': { template: '<div class="el-pagination-stub" />' },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  messageMocks.warning.mockReset()
  confirmMock.mockReset()
  for (const fn of Object.values(wordApis)) fn.mockReset()
  for (const fn of Object.values(securityApis)) fn.mockReset()

  wordApis.getProhibitedWordList.mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
  securityApis.getSecuritySettings.mockResolvedValue({ code: 200, data: {} })
})

describe('ProhibitedWordsView.vue — 列表与字段归一化', () => {
  it('onMounted 同时拉取违禁词列表与安全设置', async () => {
    makeWrapper()
    await flushPromises()
    expect(wordApis.getProhibitedWordList).toHaveBeenCalledWith({ page: 1, size: 10, keyword: '' })
    expect(securityApis.getSecuritySettings).toHaveBeenCalledTimes(1)
  })

  it('snake_case 字段归一化为表格字段，缺省值兜底', async () => {
    wordApis.getProhibitedWordList.mockResolvedValue({
      code: 200,
      data: {
        list: [{ id: 1, word: '违禁', match_type: 'EXACT', category: 'PORN', is_enabled: 0, created_at: 'T' }],
        total: 1,
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([
      { id: 1, word: '违禁', matchType: 'EXACT', category: 'PORN', isEnabled: 0, createdAt: 'T' },
    ])
    expect(w.vm.total).toBe(1)
  })

  it('camelCase 字段也可识别，缺失项回落默认值', async () => {
    wordApis.getProhibitedWordList.mockResolvedValue({
      code: 200,
      data: { list: [{ id: 2, matchType: 'EXACT', isEnabled: 0 }] },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData[0]).toEqual({
      id: 2,
      word: '',
      matchType: 'EXACT',
      category: '',
      isEnabled: 0,
      createdAt: '',
    })
  })

  it('兼容 data 直接是数组的响应形状', async () => {
    wordApis.getProhibitedWordList.mockResolvedValue({ code: 200, data: [{ id: 3, word: 'x' }] })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toHaveLength(1)
    expect(w.vm.tableData[0].word).toBe('x')
  })

  it('兼容 success 标记 + 根级 total 的响应形状', async () => {
    wordApis.getProhibitedWordList.mockResolvedValue({ success: true, data: [{ id: 4 }], total: 9 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toHaveLength(1)
    expect(w.vm.total).toBe(9)
  })

  it('业务失败时不覆盖列表，loading 复位', async () => {
    wordApis.getProhibitedWordList.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.tableData).toEqual([])
    expect(w.vm.loading).toBe(false)
  })

  it('拉取失败：提示错误并复位 loading', async () => {
    wordApis.getProhibitedWordList.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取违禁词列表失败')
    expect(w.vm.loading).toBe(false)
  })

  it('搜索：带 keyword 并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentPage = 4
    w.vm.keyword = '违禁'
    w.vm.handleSearch()
    await flushPromises()
    expect(w.vm.currentPage).toBe(1)
    expect(wordApis.getProhibitedWordList).toHaveBeenLastCalledWith({ page: 1, size: 10, keyword: '违禁' })
  })

  it('重置：清空 keyword 并回到第 1 页', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.currentPage = 3
    w.vm.keyword = 'abc'
    w.vm.handleReset()
    await flushPromises()
    expect(w.vm.keyword).toBe('')
    expect(w.vm.currentPage).toBe(1)
  })

  it('翻页：按新页码重新查询', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handlePageChange(6)
    await flushPromises()
    expect(wordApis.getProhibitedWordList).toHaveBeenLastCalledWith({ page: 6, size: 10, keyword: '' })
  })
})

describe('ProhibitedWordsView.vue — 安全设置', () => {
  it('用后端返回值覆盖安全设置', async () => {
    securityApis.getSecuritySettings.mockResolvedValue({
      code: 200,
      data: { commentMaxCount: 3, commentWindowSeconds: 30, replyMaxCount: 6, replyWindowSeconds: 45, cacheRefreshIntervalSeconds: 120 },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.securityForm).toEqual({
      commentMaxCount: 3,
      commentWindowSeconds: 30,
      replyMaxCount: 6,
      replyWindowSeconds: 45,
      cacheRefreshIntervalSeconds: 120,
    })
    expect(w.vm.securityLoading).toBe(false)
  })

  it('后端返回空对象时保持内置默认值', async () => {
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.securityForm).toEqual(emptySecurityForm())
  })

  it('⚠️ 契约不匹配：后端返回的是 password_policy/login_policy 嵌套结构，本表单的 5 个扁平字段拿不到值，只能显示内置默认值', async () => {
    securityApis.getSecuritySettings.mockResolvedValue({
      code: 200,
      data: {
        password_policy: { min_length: 8, require_upper: true },
        login_policy: { max_attempts: 5, lockout_minutes: 30 },
      },
    })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.securityForm).toEqual(emptySecurityForm())
  })

  it('⚠️ 0 会被 || 兜底打回默认值，无法把上限配成 0', async () => {
    securityApis.getSecuritySettings.mockResolvedValue({ code: 200, data: { commentMaxCount: 0, replyMaxCount: 0 } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.securityForm.commentMaxCount).toBe(10)
    expect(w.vm.securityForm.replyMaxCount).toBe(20)
  })

  it('拉取失败：提示错误并复位 securityLoading', async () => {
    securityApis.getSecuritySettings.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取安全设置失败')
    expect(w.vm.securityLoading).toBe(false)
  })

  it('保存安全设置成功：提交整份表单并提示', async () => {
    securityApis.updateSecuritySettings.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.securityForm.commentMaxCount = 5
    await w.vm.handleSaveSecurity()
    expect(securityApis.updateSecuritySettings).toHaveBeenCalledWith(expect.objectContaining({ commentMaxCount: 5 }))
    expect(messageMocks.success).toHaveBeenCalledWith('保存成功')
    expect(w.vm.securityLoading).toBe(false)
  })

  it('保存安全设置失败：提示失败并复位 loading', async () => {
    securityApis.updateSecuritySettings.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleSaveSecurity()
    expect(messageMocks.error).toHaveBeenCalledWith('保存失败')
    expect(w.vm.securityLoading).toBe(false)
  })
})

describe('ProhibitedWordsView.vue — 新增/编辑/删除', () => {
  it('handleAdd：打开弹窗并重置为新增态', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogForm.word = '脏数据'
    w.vm.handleAdd()
    expect(w.vm.dialogTitle).toBe('添加违禁词')
    expect(w.vm.dialogVisible).toBe(true)
    expect(w.vm.dialogForm).toEqual(emptyDialogForm())
  })

  it('handleEdit：把行数据搬进表单，标题切为编辑', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleEdit({ id: 7, word: '原词', matchType: 'EXACT', category: 'AD', isEnabled: 0 })
    expect(w.vm.dialogTitle).toBe('编辑违禁词')
    expect(w.vm.dialogForm).toEqual({ id: 7, word: '原词', matchType: 'EXACT', category: 'AD', isEnabled: 0 })
    expect(w.vm.dialogVisible).toBe(true)
  })

  it('handleEdit：matchType 缺失时兜底 CONTAINS', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleEdit({ id: 8, word: 'x', isEnabled: 1 })
    expect(w.vm.dialogForm.matchType).toBe('CONTAINS')
  })

  it('无 dialogFormRef 时保存直接返回，不发请求', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = null
    await w.vm.handleSave()
    expect(wordApis.addProhibitedWord).not.toHaveBeenCalled()
  })

  it('校验不通过：不发请求、不关弹窗', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(false) }
    w.vm.dialogVisible = true
    await w.vm.handleSave()
    await flushPromises()
    expect(wordApis.addProhibitedWord).not.toHaveBeenCalled()
    expect(w.vm.dialogVisible).toBe(true)
  })

  it('新增成功：走 add 接口、提示添加成功、关弹窗并刷新列表', async () => {
    wordApis.addProhibitedWord.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogForm = { ...emptyDialogForm(), word: '新词' }
    w.vm.dialogVisible = true
    wordApis.getProhibitedWordList.mockClear()
    await w.vm.handleSave()
    await flushPromises()
    expect(wordApis.addProhibitedWord).toHaveBeenCalledWith({ id: null, word: '新词', matchType: 'CONTAINS', category: '', isEnabled: 1 })
    expect(wordApis.updateProhibitedWord).not.toHaveBeenCalled()
    expect(messageMocks.success).toHaveBeenCalledWith('添加成功')
    expect(w.vm.dialogVisible).toBe(false)
    expect(wordApis.getProhibitedWordList).toHaveBeenCalled()
  })

  it('编辑成功：走 update 接口并带上 id，提示更新成功', async () => {
    wordApis.updateProhibitedWord.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    w.vm.dialogFormRef = { validate: (cb: any) => cb(true) }
    w.vm.dialogForm = { id: 5, word: '改过的词', matchType: 'EXACT', category: 'AD', isEnabled: 0 }
    await w.vm.handleSave()
    await flushPromises()
    expect(wordApis.updateProhibitedWord).toHaveBeenCalledWith(5, {
      id: 5,
      word: '改过的词',
      matchType: 'EXACT',
      category: 'AD',
      isEnabled: 0,
    })
    expect(messageMocks.success).toHaveBeenCalledWith('更新成功')
  })

  it('保存抛异常：提示操作失败并复位 loading', async () => {
    wordApis.addProhibitedWord.mockRejectedValue(new Error('boom'))
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

  it('删除成功：确认后调用接口、提示并刷新列表', async () => {
    confirmMock.mockResolvedValue('confirm')
    wordApis.deleteProhibitedWord.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    await flushPromises()
    wordApis.getProhibitedWordList.mockClear()
    await w.vm.handleDelete({ id: 3 })
    expect(confirmMock).toHaveBeenCalledTimes(1)
    expect(confirmMock.mock.calls[0][0]).toContain('不可恢复')
    expect(wordApis.deleteProhibitedWord).toHaveBeenCalledWith(3)
    expect(messageMocks.success).toHaveBeenCalledWith('删除成功')
    expect(wordApis.getProhibitedWordList).toHaveBeenCalled()
  })

  it('取消删除：不调接口、不提示', async () => {
    confirmMock.mockRejectedValue('cancel')
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3 })
    expect(wordApis.deleteProhibitedWord).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('删除接口返回业务失败：提示后端 message，不刷新列表', async () => {
    confirmMock.mockResolvedValue('confirm')
    wordApis.deleteProhibitedWord.mockResolvedValue({ code: 500, message: '词条被引用中' })
    const w = makeWrapper()
    await flushPromises()
    wordApis.getProhibitedWordList.mockClear()
    await w.vm.handleDelete({ id: 3 })
    expect(messageMocks.error).toHaveBeenCalledWith('词条被引用中')
    expect(wordApis.getProhibitedWordList).not.toHaveBeenCalled()
  })

  it('删除接口抛错：必须提示用户，不能静默吞掉', async () => {
    confirmMock.mockResolvedValue('confirm')
    wordApis.deleteProhibitedWord.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleDelete({ id: 3 })
    expect(messageMocks.error).toHaveBeenCalledWith('删除失败')
  })
})

describe('ProhibitedWordsView.vue — 批量导入', () => {
  it('未选文件：警告并直接返回', async () => {
    const w = makeWrapper()
    await flushPromises()
    await w.vm.handleImport()
    expect(messageMocks.warning).toHaveBeenCalledWith('请选择文件')
    expect(wordApis.batchImportProhibitedWords).not.toHaveBeenCalled()
  })

  it('导入成功：拼 FormData 并按后端统计提示，关闭弹窗后刷新', async () => {
    wordApis.batchImportProhibitedWords.mockResolvedValue({
      code: 200,
      data: { successCount: 8, failCount: 2 },
    })
    const w = makeWrapper()
    await flushPromises()
    w.vm.importForm = {
      file: new File(['a,b'], 'words.csv', { type: 'text/csv' }),
      matchType: 'EXACT',
      category: 'AD',
    }
    w.vm.importDialogVisible = true
    wordApis.getProhibitedWordList.mockClear()
    await w.vm.handleImport()
    const sent = wordApis.batchImportProhibitedWords.mock.calls[0][0] as FormData
    expect(sent).toBeInstanceOf(FormData)
    expect(sent.get('matchType')).toBe('EXACT')
    expect(sent.get('category')).toBe('AD')
    expect(sent.get('file')).toBeInstanceOf(File)
    expect(messageMocks.success).toHaveBeenCalledWith('导入完成：成功 8 条，失败 2 条')
    expect(w.vm.importDialogVisible).toBe(false)
    expect(wordApis.getProhibitedWordList).toHaveBeenCalled()
  })

  it('分类留空时 FormData 不带 category 字段', async () => {
    wordApis.batchImportProhibitedWords.mockResolvedValue({ code: 200, data: {} })
    const w = makeWrapper()
    await flushPromises()
    w.vm.importForm = { file: new File(['x'], 'w.txt'), matchType: 'CONTAINS', category: '' }
    await w.vm.handleImport()
    const sent = wordApis.batchImportProhibitedWords.mock.calls[0][0] as FormData
    expect(sent.get('category')).toBeNull()
    expect(messageMocks.success).toHaveBeenCalledWith('导入完成：成功 0 条，失败 0 条')
  })

  it('导入失败：提示失败、保留弹窗、复位 loading', async () => {
    wordApis.batchImportProhibitedWords.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    w.vm.importForm = { file: new File(['x'], 'w.txt'), matchType: 'CONTAINS', category: '' }
    w.vm.importDialogVisible = true
    await w.vm.handleImport()
    expect(messageMocks.error).toHaveBeenCalledWith('导入失败')
    expect(w.vm.importDialogVisible).toBe(true)
    expect(w.vm.loading).toBe(false)
  })

  it('handleOpenImport：重置导入表单与文件列表', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.importForm = { file: '脏数据', matchType: 'EXACT', category: 'AD' }
    w.vm.fileList = [1, 2]
    w.vm.handleOpenImport()
    expect(w.vm.importForm).toEqual({ file: null, matchType: 'CONTAINS', category: '' })
    expect(w.vm.fileList).toEqual([])
    expect(w.vm.importDialogVisible).toBe(true)
  })

  it('handleFileChange：取 el-upload 传回的 raw 文件', async () => {
    const w = makeWrapper()
    await flushPromises()
    w.vm.handleFileChange({ raw: new File(['x'], 'w.txt') })
    expect(w.vm.importForm.file.name).toBe('w.txt')
  })
})

describe('ProhibitedWordsView.vue — 展示辅助函数', () => {
  it('getMatchTypeLabel：命中选项转中文，未知值原样返回', () => {
    const w = makeWrapper()
    expect(w.vm.getMatchTypeLabel('CONTAINS')).toBe('包含匹配')
    expect(w.vm.getMatchTypeLabel('EXACT')).toBe('精确匹配')
    expect(w.vm.getMatchTypeLabel('WEIRD')).toBe('WEIRD')
  })

  it('getCategoryLabel：空值占位符，命中转中文，未知原样返回', () => {
    const w = makeWrapper()
    expect(w.vm.getCategoryLabel('')).toBe('-')
    expect(w.vm.getCategoryLabel('POLITICS')).toBe('政治')
    expect(w.vm.getCategoryLabel('PORN')).toBe('色情')
    expect(w.vm.getCategoryLabel('UNKNOWN')).toBe('UNKNOWN')
  })

  it('formatDateTime：空值占位，合法时间补零到 YYYY-MM-DD HH:mm', () => {
    const w = makeWrapper()
    expect(w.vm.formatDateTime('')).toBe('-')
    const d = new Date(2026, 0, 2, 3, 4)
    expect(w.vm.formatDateTime(d)).toBe('2026-01-02 03:04')
  })
})
