import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import RecommendConfigView from './RecommendConfigView.vue'

const getConfig = vi.hoisted(() => vi.fn())
const updateConfig = vi.hoisted(() => vi.fn())
const resetConfig = vi.hoisted(() => vi.fn())
vi.mock('@/api/recommendConfig', () => ({
  recommendConfigApi: { getConfig, updateConfig, resetConfig },
}))

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
  ElMessageBox: { confirm: confirmMock },
}))

const makeWrapper = () =>
  mount(RecommendConfigView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-card': { template: '<div class="el-card-stub"><slot name="header" /><slot /></div>' },
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
        'el-switch': {
          template:
            '<input type="checkbox" class="el-switch-stub" :checked="modelValue" @change="e=>$emit(`update:modelValue`, e.target.checked)" />',
          props: ['modelValue'],
        },
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
        },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  confirmMock.mockReset()
  getConfig.mockReset()
  updateConfig.mockReset()
  resetConfig.mockReset()
  getConfig.mockResolvedValue({ data: { code: 200, data: { userInterestWeight: 0.7 } } })
})

describe('RecommendConfigView.vue', () => {
  it('onMounted 拉取配置并写入 config', async () => {
    const w = makeWrapper()
    await flushPromises()
    expect(getConfig).toHaveBeenCalledTimes(1)
    expect(w.vm.config).toEqual({ userInterestWeight: 0.7 })
    expect(w.vm.loading).toBe(false)
  })

  it('拉取失败：提示错误并复位 loading，config 保持空', async () => {
    getConfig.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取推荐配置失败')
    expect(w.vm.loading).toBe(false)
    expect(w.vm.config).toEqual({})
  })

  it('code !== 200 时不覆盖 config', async () => {
    getConfig.mockResolvedValue({ data: { code: 500, message: 'x' } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.config).toEqual({})
  })

  it('保存成功：以服务端返回为准并提示', async () => {
    updateConfig.mockResolvedValue({ data: { code: 200, data: { userInterestWeight: 0.9 } } })
    const w = makeWrapper()
    await flushPromises()
    w.vm.config.userInterestWeight = 0.9
    await w.vm.saveConfig()
    expect(updateConfig).toHaveBeenCalledWith({ userInterestWeight: 0.9 })
    expect(w.vm.config).toEqual({ userInterestWeight: 0.9 })
    expect(messageMocks.success).toHaveBeenCalledWith('保存成功')
  })

  it('保存业务失败：展示后端 message，不提示成功', async () => {
    updateConfig.mockResolvedValue({ data: { code: 500, message: '权重超出范围' } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('权重超出范围')
    expect(messageMocks.success).not.toHaveBeenCalled()
  })

  it('保存业务失败且无 message：回落通用文案', async () => {
    updateConfig.mockResolvedValue({ data: { code: 500 } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('保存失败')
  })

  it('保存抛异常：提示失败', async () => {
    updateConfig.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('保存失败')
  })

  it('重置：确认后调用重置接口并回填默认值', async () => {
    confirmMock.mockResolvedValue('confirm')
    resetConfig.mockResolvedValue({ data: { code: 200, data: { userInterestWeight: 0.5 } } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.resetConfig()
    expect(confirmMock).toHaveBeenCalledTimes(1)
    expect(resetConfig).toHaveBeenCalled()
    expect(w.vm.config).toEqual({ userInterestWeight: 0.5 })
    expect(messageMocks.success).toHaveBeenCalledWith('已重置为默认值')
  })

  it('重置被取消：不调接口、不报错', async () => {
    confirmMock.mockRejectedValue('cancel')
    const w = makeWrapper()
    await flushPromises()
    await w.vm.resetConfig()
    expect(resetConfig).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
  })

  it('重置接口抛异常：提示重置失败', async () => {
    confirmMock.mockResolvedValue('confirm')
    resetConfig.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.resetConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('重置失败')
  })

  it('formatUpdatedAt：空值返回空串，非法时间原样返回，合法时间补零格式化', () => {
    const w = makeWrapper()
    expect(w.vm.formatUpdatedAt('')).toBe('')
    expect(w.vm.formatUpdatedAt(null)).toBe('')
    expect(w.vm.formatUpdatedAt('not-a-date')).toBe('not-a-date')

    const d = new Date(2026, 0, 2, 3, 4)
    expect(w.vm.formatUpdatedAt(d)).toBe('2026-01-02 03:04')
  })
})
