import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import TranscodeConfigView from './TranscodeConfigView.vue'

const getConfig = vi.hoisted(() => vi.fn())
const updateConfig = vi.hoisted(() => vi.fn())
vi.mock('@/api/transcodeConfig', () => ({
  transcodeConfigApi: { getConfig, updateConfig },
}))

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
}))

const makeWrapper = () =>
  mount(TranscodeConfigView, {
    global: {
      directives: { loading: {} },
      stubs: {
        'el-card': { template: '<div class="el-card-stub"><slot name="header" /><slot /></div>' },
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-select': { template: '<div class="el-select-stub"><slot /></div>' },
        'el-option': { template: '<div class="el-option-stub" />' },
        'el-descriptions': { template: '<div class="el-descriptions-stub"><slot /></div>' },
        'el-descriptions-item': {
          template: '<div class="el-descriptions-item-stub"><slot /></div>',
        },
        'el-tag': { template: '<span class="el-tag-stub"><slot /></span>' },
        'el-button': {
          template:
            '<button class="el-button-stub" :disabled="loading" @click="$emit(\'click\')"><slot /></button>',
          props: ['loading'],
        },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  getConfig.mockReset()
  updateConfig.mockReset()
  getConfig.mockResolvedValue({ data: { code: 200, data: { encoder: 'auto', vaapi: false } } })
})

describe('TranscodeConfigView.vue', () => {
  it('onMounted 拉取配置并回填表单', async () => {
    const w = makeWrapper()
    await flushPromises()
    expect(getConfig).toHaveBeenCalledTimes(1)
    expect(w.vm.encoder).toBe('auto')
    expect(w.vm.vaapi).toBe(false)
    expect(w.vm.loading).toBe(false)
  })

  it('兼容后端直接把配置挂在响应根上的情况', async () => {
    getConfig.mockResolvedValue({ encoder: 'x264', vaapi: true, vaapiDev: '/dev/dri/renderD129' })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.encoder).toBe('x264')
    expect(w.vm.vaapi).toBe(true)
    expect(w.vm.vaapiDev).toBe('/dev/dri/renderD129')
  })

  it('配置缺字段时回落默认值 auto + 三个可选编码器', async () => {
    getConfig.mockResolvedValue({ data: { code: 200, data: {} } })
    const w = makeWrapper()
    await flushPromises()
    expect(w.vm.encoder).toBe('auto')
    expect(w.vm.vaapiDev).toBe('')
    expect(w.vm.options).toEqual(['auto', 'vaapi', 'x264'])
  })

  it('拉取配置失败：提示错误并复位 loading', async () => {
    getConfig.mockRejectedValue(new Error('network'))
    const w = makeWrapper()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('获取转码配置失败')
    expect(w.vm.loading).toBe(false)
  })

  it('vaapiStatusText 随编码器与硬件可用性变化', async () => {
    const w = makeWrapper()
    await flushPromises()

    w.vm.encoder = 'vaapi'
    expect(w.vm.vaapiStatusText).toBe('硬编已启用')

    w.vm.encoder = 'x264'
    expect(w.vm.vaapiStatusText).toBe('硬编未启用（软件编码）')

    w.vm.encoder = 'auto'
    w.vm.vaapi = true
    expect(w.vm.vaapiStatusText).toBe('自动探测到可用，将优先硬编')

    w.vm.vaapi = false
    expect(w.vm.vaapiStatusText).toBe('未探测到，将回退软件编码')
  })

  it('vaapiTagType：x264 恒为 info，其余跟随硬件可用性', async () => {
    const w = makeWrapper()
    await flushPromises()

    w.vm.encoder = 'x264'
    w.vm.vaapi = true
    expect(w.vm.vaapiTagType).toBe('info')

    w.vm.encoder = 'auto'
    w.vm.vaapi = true
    expect(w.vm.vaapiTagType).toBe('success')

    w.vm.vaapi = false
    expect(w.vm.vaapiTagType).toBe('warning')
  })

  it('保存成功：只提交 encoder，并提示需重启转码服务', async () => {
    updateConfig.mockResolvedValue({ data: { code: 200 } })
    const w = makeWrapper()
    await flushPromises()
    w.vm.encoder = 'vaapi'
    await w.vm.saveConfig()
    expect(updateConfig).toHaveBeenCalledWith({ encoder: 'vaapi' })
    expect(messageMocks.success).toHaveBeenCalledWith('配置已保存，重启转码服务后生效')
    expect(w.vm.saving).toBe(false)
  })

  it('保存：根级 status=ok 也算成功', async () => {
    updateConfig.mockResolvedValue({ status: 'ok' })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.success).toHaveBeenCalled()
  })

  it('保存：业务失败时展示后端 message', async () => {
    updateConfig.mockResolvedValue({ data: { code: 500, message: '编码器不被支持' } })
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('编码器不被支持')
    expect(messageMocks.success).not.toHaveBeenCalled()
    expect(w.vm.saving).toBe(false)
  })

  it('保存：抛异常时提示通用失败文案', async () => {
    updateConfig.mockRejectedValue(new Error('boom'))
    const w = makeWrapper()
    await flushPromises()
    await w.vm.saveConfig()
    expect(messageMocks.error).toHaveBeenCalledWith('保存失败')
    expect(w.vm.saving).toBe(false)
  })
})
