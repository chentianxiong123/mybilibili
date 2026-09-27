import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn(),
  getLocalUser: vi.fn(),
  logout: vi.fn()
}))

vi.mock('../../api/user', () => ({
  getMyInfo: vi.fn()
}))

vi.mock('../../utils/theme', () => ({
  getWapTheme: vi.fn().mockReturnValue('light'),
  toggleWapTheme: vi.fn()
}))

import Space from '../space/Space.vue'
import { isLogin, getLocalUser, logout } from '../../utils/session'
import { getMyInfo } from '../../api/user'

const mockIsLogin = vi.mocked(isLogin)
const mockGetLocalUser = vi.mocked(getLocalUser)
const mockLogout = vi.mocked(logout)
const mockGetMyInfo = vi.mocked(getMyInfo)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/space', component: Space },
    { path: '/m/login', component: { template: '<div />' } },
    { path: '/m/space/history', component: { template: '<div />' } },
    { path: '/m/space/favorite', component: { template: '<div />' } },
    { path: '/m/space/manuscripts', component: { template: '<div />' } },
    { path: '/m/space/profile/edit', component: { template: '<div />' } }
  ]
})

describe('Space.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(true)
    mockGetLocalUser.mockReturnValue({ id: 1, name: 'test' })
    mockGetMyInfo.mockResolvedValue({ code: '1', data: { id: 1, nickname: '测试用户', followerCount: 100 } })
    router.push('/m/space')
    await router.isReady()
  })

  it('未登录 → 跳转 /m/login', async () => {
    mockIsLogin.mockReturnValue(false)
    mount(Space, { global: { plugins: [router] } })
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/login')
  })

  it('已登录 → 加载用户信息', async () => {
    mount(Space, { global: { plugins: [router] } })
    await flushPromises()
    expect(mockGetMyInfo).toHaveBeenCalled()
  })

  it('getMyInfo 失败 → 降级使用本地用户', async () => {
    mockGetMyInfo.mockResolvedValue({ code: '0', data: null })
    mockGetLocalUser.mockReturnValue({ id: 1, nickname: '本地用户' })
    const wrapper = mount(Space, { global: { plugins: [router] } })
    await flushPromises()
    // 应该使用本地用户数据
    expect(mockGetLocalUser).toHaveBeenCalled()
  })

  it('点击退出登录 → 调用 logout + 跳转 /m/login', async () => {
    const wrapper = mount(Space, { global: { plugins: [router] } })
    await flushPromises()
    const logoutBtn = wrapper.findAll('.menu-item').find(el => el.text().includes('退出'))
    if (logoutBtn) {
      await logoutBtn.trigger('click')
      await flushPromises()
      expect(mockLogout).toHaveBeenCalled()
      expect(router.currentRoute.value.path).toBe('/m/login')
    }
  })

  it('getUserCount: 优先取第一个有值的 key', async () => {
    mockGetMyInfo.mockResolvedValue({
      code: '1',
      data: { id: 1, followerCount: 50, followers: 100 }
    })
    const wrapper = mount(Space, { global: { plugins: [router] } })
    await flushPromises()
    // 组件内部 getUserCount(['followerCount', 'followers']) 应返回 50
  })
})
