import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('element-plus', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn() },
}))

vi.mock('@element-plus/icons-vue', () => ({
  User: { name: 'User', render: () => {} },
  Lock: { name: 'Lock', render: () => {} },
  Message: { name: 'Message', render: () => {} },
  VideoPlay: { name: 'VideoPlay', render: () => {} },
}))

vi.mock('@/api/client', () => ({
  userApi: {
    login: vi.fn(),
    register: vi.fn(),
  },
}))

// 动态导入 component，确保 mock 已生效
import LoginRegister from './LoginRegister.vue'
import { userApi } from '@/api/client'
import { ElMessage } from 'element-plus'

function createWrapper() {
  return mount(LoginRegister, {
    global: {
      stubs: {
        'el-input': { template: '<div><input :placeholder="placeholder" /></div>', props: ['modelValue', 'placeholder', 'type', 'prefixIcon'] },
        'el-checkbox': { template: '<div><input type="checkbox" /></div>', props: ['modelValue'] },
        'el-icon': { template: '<span><slot /></span>' },
      },
    },
  })
}

describe('LoginRegister.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('默认显示登录表单', () => {
    const wrapper = createWrapper()
    expect(wrapper.find('.login-register').exists()).toBe(true)
    expect(wrapper.find('.submit').exists()).toBe(true)
    expect(wrapper.text()).toContain('登录')
    expect(wrapper.text()).toContain('还没有账号')
  })

  it('点击注册链接切换到注册表单', async () => {
    const wrapper = createWrapper()
    await wrapper.find('.switch-btn').trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('注册')
    expect(wrapper.text()).toContain('已有账号')
    expect(wrapper.find('.register-title').exists()).toBe(true)
  })

  it('点击登录链接从注册切回登录', async () => {
    const wrapper = createWrapper()
    // 切到注册
    await wrapper.find('.switch-btn').trigger('click')
    await nextTick()
    // 切回登录
    const switchBtns = wrapper.findAll('.switch-btn')
    await switchBtns[switchBtns.length - 1].trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('登录')
    expect(wrapper.text()).toContain('还没有账号')
  })

  it('空用户名登录触发校验错误', async () => {
    const wrapper = createWrapper()
    await wrapper.find('.submit').trigger('click')
    expect(ElMessage.error).toHaveBeenCalled()
  })

  it('空字段注册触发校验错误', async () => {
    const wrapper = createWrapper()
    // 切到注册
    await wrapper.find('.switch-btn').trigger('click')
    await nextTick()
    // 点注册按钮（空字段）
    await wrapper.find('.submit').trigger('click')
    expect(ElMessage.error).toHaveBeenCalled()
  })
})
