import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import BottomTabBar from '../BottomTabBar.vue'

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn()
}))

import { isLogin } from '../../utils/session'
const mockIsLogin = vi.mocked(isLogin)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/m/index', component: { template: '<div />' } },
    { path: '/m/dynamic', component: { template: '<div />' } },
    { path: '/m/space', component: { template: '<div />' } },
    { path: '/m/login', component: { template: '<div />' } },
    { path: '/m/mall', component: { template: '<div />' } }
  ]
})

describe('BottomTabBar.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(false)
    router.push('/m/index')
    await router.isReady()
  })

  it('渲染 5 个 tab', () => {
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    expect(wrapper.findAll('.tab-item').length).toBe(5)
  })

  it('首页 tab 显示 "首页"', () => {
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const tabs = wrapper.findAll('.tab-name')
    expect(tabs[0].text()).toBe('首页')
  })

  it('未登录点"我的"→ 跳 /m/login', async () => {
    mockIsLogin.mockReturnValue(false)
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const lastTab = wrapper.findAll('.tab-item')[4]
    await lastTab.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/login')
  })

  it('已登录点"我的"→ 跳 /m/space', async () => {
    mockIsLogin.mockReturnValue(true)
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const lastTab = wrapper.findAll('.tab-item')[4]
    await lastTab.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/space')
  })

  it('点会员购 tab → 不跳转（提示开发中）', async () => {
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const mallTab = wrapper.findAll('.tab-item')[3]
    await mallTab.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/index')
  })

  it('当前路径是 /m/index 时，首页 tab 有 active 样式', () => {
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const firstTab = wrapper.findAll('.tab-item')[0]
    expect(firstTab.classes()).toContain('active')
  })

  it('发布按钮渲染 big-plus-btn', () => {
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    expect(wrapper.find('.big-plus-btn').exists()).toBe(true)
  })

  it('发布按钮未登录 → 跳 /m/login', async () => {
    mockIsLogin.mockReturnValue(false)
    const wrapper = mount(BottomTabBar, {
      global: { plugins: [router] }
    })
    const publishBtn = wrapper.find('.publish-btn-container')
    await publishBtn.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/login')
  })
})
