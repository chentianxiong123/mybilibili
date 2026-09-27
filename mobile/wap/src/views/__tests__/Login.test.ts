import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../../utils/session', () => ({
  saveSession: vi.fn(),
  getLocalUser: vi.fn()
}))

vi.mock('../../api/user', () => ({
  getMyInfo: vi.fn()
}))

import Login from '../Login.vue'
import { saveSession, getLocalUser } from '../../utils/session'
import { getMyInfo } from '../../api/user'

const mockSaveSession = vi.mocked(saveSession)
const mockGetMyInfo = vi.mocked(getMyInfo)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/index', component: { template: '<div />' } },
    { path: '/m/login', component: { template: '<div />' } }
  ]
})

describe('Login.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    vi.stubGlobal('fetch', vi.fn())
    mockGetMyInfo.mockResolvedValue({ code: '0', data: null })
    router.push('/m/login')
    await router.isReady()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  const mountLogin = () => mount(Login, { global: { plugins: [router] } })

  it('渲染登录表单', () => {
    const wrapper = mountLogin()
    expect(wrapper.find('input[placeholder="用户名"]').exists()).toBe(true)
    expect(wrapper.find('input[placeholder="密码"]').exists()).toBe(true)
    expect(wrapper.find('.login-btn').text()).toBe('登录')
  })

  it('空用户名 → 显示错误', async () => {
    const wrapper = mountLogin()
    await wrapper.find('.login-btn').trigger('click')
    expect(wrapper.find('.error-msg').text()).toBe('请输入用户名和密码')
  })

  it('空密码 → 显示错误', async () => {
    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('.login-btn').trigger('click')
    expect(wrapper.find('.error-msg').text()).toBe('请输入用户名和密码')
  })

  it('空白用户名 → 显示错误', async () => {
    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('   ')
    await wrapper.find('.login-btn').trigger('click')
    expect(wrapper.find('.error-msg').text()).toBe('请输入用户名和密码')
  })

  it('登录成功 → 跳转 /m/index', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ code: 200, data: { user: { id: 1 }, token: 'tok' } })
    })
    vi.stubGlobal('fetch', mockFetch)

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('123456')
    await wrapper.find('.login-btn').trigger('click')
    await flushPromises()

    expect(mockSaveSession).toHaveBeenCalled()
    expect(router.currentRoute.value.path).toBe('/m/index')
  })

  it('登录失败（非200）→ 显示服务端 message', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ code: 401, message: '账号或密码错误' })
    })
    vi.stubGlobal('fetch', mockFetch)

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('wrong')
    await wrapper.find('.login-btn').trigger('click')
    await flushPromises()

    expect(wrapper.find('.error-msg').text()).toBe('账号或密码错误')
  })

  it('登录失败无 message → 显示 "登录失败"', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ code: 401 })
    })
    vi.stubGlobal('fetch', mockFetch)

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('wrong')
    await wrapper.find('.login-btn').trigger('click')
    await flushPromises()

    expect(wrapper.find('.error-msg').text()).toBe('登录失败')
  })

  it('网络异常 → 显示 "网络错误"', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network')))

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('123')
    await wrapper.find('.login-btn').trigger('click')
    await flushPromises()

    expect(wrapper.find('.error-msg').text()).toBe('网络错误')
  })

  it('登录成功后拉取用户信息并合并', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ code: 200, data: { user: { id: 1, name: 'a' }, token: 'tok' } })
    })
    vi.stubGlobal('fetch', mockFetch)
    mockGetMyInfo.mockResolvedValue({ code: '1', data: { id: 1, nickname: '新昵称' } })

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('123456')
    await wrapper.find('.login-btn').trigger('click')
    await flushPromises()

    expect(mockGetMyInfo).toHaveBeenCalled()
    // saveSession 被调用两次：登录时 + 拉取信息后
    expect(mockSaveSession).toHaveBeenCalledTimes(2)
  })

  it('loading 状态切换', async () => {
    let resolveFetch: any
    vi.stubGlobal('fetch', vi.fn().mockImplementation(() =>
      new Promise(resolve => { resolveFetch = resolve })
    ))

    const wrapper = mountLogin()
    await wrapper.find('input[placeholder="用户名"]').setValue('admin')
    await wrapper.find('input[placeholder="密码"]').setValue('123')
    await wrapper.find('.login-btn').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('.login-btn').text()).toBe('登录中...')
    expect(wrapper.find('.login-btn').attributes('disabled')).toBeDefined()

    resolveFetch({ json: () => Promise.resolve({ code: 401 }) })
    await flushPromises()
    expect(wrapper.find('.login-btn').text()).toBe('登录')
  })
})
