import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SystemNotificationManagerView from './SystemNotificationManagerView.vue'

const broadcast = vi.hoisted(() => vi.fn())
vi.mock('@/api/message.ts', () => ({
  messageApi: { broadcastSystemNotification: broadcast },
}))

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn(), warning: vi.fn() }))
const confirmMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: messageMocks,
  ElMessageBox: { confirm: confirmMock },
}))

vi.mock('@element-plus/icons-vue', () => ({
  Bell: { template: '<i class="bell-stub" />' },
  Promotion: { template: '<i class="promotion-stub" />' },
}))

const makeWrapper = () =>
  mount(SystemNotificationManagerView, {
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-card': { template: '<div class="el-card-stub"><slot name="header" /><slot /></div>' },
        'el-input': {
          template:
            '<textarea class="el-input-stub" :value="modelValue" @input="e=>$emit(`update:modelValue`, e.target.value)" />',
          props: ['modelValue'],
        },
        'el-button': {
          template:
            '<button class="el-button-stub" :disabled="disabled" @click="$emit(`click`)"><slot /></button>',
          props: ['disabled'],
        },
      },
    },
  })

beforeEach(() => {
  vi.clearAllMocks()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
  messageMocks.warning.mockReset()
  confirmMock.mockReset()
  broadcast.mockReset()
})

describe('SystemNotificationManagerView.vue', () => {
  it('内容为空或纯空白：直接警告，不弹确认框也不发请求', async () => {
    const w = makeWrapper()
    await w.vm.handleBroadcast()
    expect(messageMocks.warning).toHaveBeenCalledWith('请输入通知内容')
    expect(confirmMock).not.toHaveBeenCalled()
    expect(broadcast).not.toHaveBeenCalled()

    w.vm.form.content = '    '
    await w.vm.handleBroadcast()
    expect(messageMocks.warning).toHaveBeenCalledTimes(2)
    expect(broadcast).not.toHaveBeenCalled()
  })

  it('发送成功：内容去空白提交、成功提示并清空表单', async () => {
    confirmMock.mockResolvedValue('confirm')
    broadcast.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    w.vm.form.content = '  系统维护通知  '
    await w.vm.handleBroadcast()
    expect(broadcast).toHaveBeenCalledWith({ content: '系统维护通知' })
    expect(messageMocks.success).toHaveBeenCalledWith('全站系统通知发送成功')
    expect(w.vm.form.content).toBe('')
    expect(w.vm.sending).toBe(false)
  })

  it('发送前必须经过二次确认，且确认文案点明不可撤回', async () => {
    confirmMock.mockResolvedValue('confirm')
    broadcast.mockResolvedValue({ code: 200 })
    const w = makeWrapper()
    w.vm.form.content = '公告'
    await w.vm.handleBroadcast()
    expect(confirmMock).toHaveBeenCalledTimes(1)
    expect(confirmMock.mock.calls[0][0]).toContain('不可撤回')
  })

  it('用户取消确认：不发请求、不报错', async () => {
    confirmMock.mockRejectedValue('cancel')
    const w = makeWrapper()
    w.vm.form.content = '公告'
    await w.vm.handleBroadcast()
    expect(broadcast).not.toHaveBeenCalled()
    expect(messageMocks.error).not.toHaveBeenCalled()
    expect(w.vm.sending).toBe(false)
  })

  it('业务失败：展示后端 message 且保留输入内容', async () => {
    confirmMock.mockResolvedValue('confirm')
    broadcast.mockResolvedValue({ code: 500, message: '发送频率过快' })
    const w = makeWrapper()
    w.vm.form.content = '公告'
    await w.vm.handleBroadcast()
    expect(messageMocks.error).toHaveBeenCalledWith('发送频率过快')
    expect(w.vm.form.content).toBe('公告')
  })

  it('业务失败且后端未给 message：回落通用文案', async () => {
    confirmMock.mockResolvedValue('confirm')
    broadcast.mockResolvedValue({ code: 500 })
    const w = makeWrapper()
    w.vm.form.content = '公告'
    await w.vm.handleBroadcast()
    expect(messageMocks.error).toHaveBeenCalledWith('发送失败')
  })

  it('请求抛异常：提示失败并复位 sending', async () => {
    confirmMock.mockResolvedValue('confirm')
    broadcast.mockRejectedValue(new Error('network'))
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const w = makeWrapper()
    w.vm.form.content = '公告'
    await w.vm.handleBroadcast()
    expect(messageMocks.error).toHaveBeenCalledWith('发送失败')
    expect(w.vm.sending).toBe(false)
    errSpy.mockRestore()
  })
})
