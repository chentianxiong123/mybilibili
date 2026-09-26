import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LoginView from './LoginView.vue'

vi.mock('@/stores/admin', async () => {
  const { defineStore } = await vi.importActual<any>('pinia')
  const loginMock = vi.fn()
  const useAdminStore = defineStore('admin-test', {
    state: () => ({ role: '', permissions: [] as string[], userInfo: null }),
    actions: { login: loginMock, logout: vi.fn() },
  })
  return { useAdminStore, __loginMock: loginMock }
})

const loginMock = (await import('@/stores/admin')).__loginMock

const routerPushSpy = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: routerPushSpy }),
}))

const messageMocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({
  ElMessage: { error: messageMocks.error, success: messageMocks.success, warning: vi.fn() },
}))

import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/stores/admin'

function makeWrapper() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const store = useAdminStore(pinia)
  store.$patch({ role: '', permissions: [] })
  const w = mount(LoginView, {
    global: {
      plugins: [pinia],
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-form': {
          template: '<form @submit.prevent><slot /></form>',
          methods: { validate: () => {} },
        },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-input': {
          template: '<input class="el-input-stub" :value="modelValue" @input="e=>$emit(\'update:modelValue\', e.target.value)" @keyup.enter="$emit(\'keyup:enter\')" />',
          props: ['modelValue'],
        },
        'el-button': { template: '<button class="el-button-stub" :disabled="loading" @click="$emit(\'click\')"><slot /></button>' },
      },
    },
  })
  return w
}

beforeEach(() => {
  vi.clearAllMocks()
  loginMock.mockReset()
  loginMock.mockResolvedValue({ success: true })
  routerPushSpy.mockReset()
  messageMocks.error.mockReset()
  messageMocks.success.mockReset()
})

describe('LoginView.vue', () => {
  it('渲染登录表单标题', () => {
    const w = makeWrapper()
    expect(w.text()).toContain('管理后台')
    expect(w.text()).toContain('默认账号: admin / admin123')
  })

  it('登录成功 → 提示成功并跳转 firstAllowedPath', async () => {
    loginMock.mockResolvedValueOnce({ success: true })
    const w = makeWrapper()
    w.vm.loginForm.username = 'admin'
    w.vm.loginForm.password = 'admin123'
    // 绕过 el-form validate，直接驱动内部逻辑
    w.vm.loginFormRef = { validate: (cb: any) => cb(true) }
    await w.vm.handleLogin()
    await flushPromises()
    expect(loginMock).toHaveBeenCalledWith({ username: 'admin', password: 'admin123' })
    expect(messageMocks.success).toHaveBeenCalledWith('登录成功')
    expect(routerPushSpy).toHaveBeenCalledWith('/no-permission')
  })

  it('登录失败 → 提示错误，不跳转', async () => {
    loginMock.mockResolvedValueOnce({ success: false, message: '用户名或密码错误' })
    const w = makeWrapper()
    w.vm.loginFormRef = { validate: (cb: any) => cb(true) }
    await w.vm.handleLogin()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('用户名或密码错误')
    expect(routerPushSpy).not.toHaveBeenCalled()
  })

  it('登录抛异常 → 提示通用错误', async () => {
    loginMock.mockRejectedValueOnce(new Error('network'))
    const w = makeWrapper()
    w.vm.loginFormRef = { validate: (cb: any) => cb(true) }
    await w.vm.handleLogin()
    await flushPromises()
    expect(messageMocks.error).toHaveBeenCalledWith('登录失败，请稍后重试')
  })

  it('validate 校验失败（valid=false）→ 不发登录请求', async () => {
    const w = makeWrapper()
    w.vm.loginFormRef = { validate: (cb: any) => cb(false) }
    await w.vm.handleLogin()
    expect(loginMock).not.toHaveBeenCalled()
  })

  it('无 loginFormRef 时直接返回', async () => {
    const w = makeWrapper()
    w.vm.loginFormRef = null
    await w.vm.handleLogin()
    expect(loginMock).not.toHaveBeenCalled()
  })
})