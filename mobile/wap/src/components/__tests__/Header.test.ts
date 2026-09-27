import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import Header from '../Header.vue'

vi.mock('../../utils/session', () => ({
  getLocalUser: vi.fn()
}))

import { getLocalUser } from '../../utils/session'
const mockGetLocalUser = vi.mocked(getLocalUser)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/index', component: { template: '<div />' } },
    { path: '/m/space', component: { template: '<div />' } },
    { path: '/m/search', component: { template: '<div />' } },
    { path: '/m/message', component: { template: '<div />' } }
  ]
})

describe('Header.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockGetLocalUser.mockReturnValue(null)
    router.push('/')
    await router.isReady()
  })

  it('渲染默认搜索占位符', async () => {
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    expect(wrapper.find('.placeholder').text()).toBe('搜索...')
  })

  it('自定义搜索占位符', async () => {
    const wrapper = mount(Header, {
      props: { placeholder: '搜索视频...' },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.placeholder').text()).toBe('搜索视频...')
  })

  it('有用户时显示头像', async () => {
    mockGetLocalUser.mockReturnValue({ id: 1, avatar: 'https://example.com/avatar.jpg' })
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    await flushPromises()
    const img = wrapper.find('.user-avatar img')
    expect(img.attributes('src')).toBe('https://example.com/avatar.jpg')
  })

  it('无用户时显示默认头像', async () => {
    mockGetLocalUser.mockReturnValue(null)
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    await flushPromises()
    const img = wrapper.find('.user-avatar img')
    expect(img.attributes('src')).toBeDefined()
    expect(img.attributes('src')).not.toBe('')
  })

  it('头像链接到 /m/space', async () => {
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    expect(wrapper.find('.user-avatar').attributes('href')).toBe('/m/space')
  })

  it('搜索框链接到 /m/search', async () => {
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    expect(wrapper.find('.search-box').attributes('href')).toBe('/m/search')
  })

  it('消息图标链接到 /m/message', async () => {
    const wrapper = mount(Header, {
      global: { plugins: [router] }
    })
    expect(wrapper.find('.msg-icon').attributes('href')).toBe('/m/message')
  })
})
