import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn()
}))

vi.mock('../../api/message', () => ({
  getConversations: vi.fn(),
  getUnreadCounts: vi.fn()
}))

import Message from '../message/Message.vue'
import { isLogin } from '../../utils/session'
import { getConversations, getUnreadCounts } from '../../api/message'

const mockIsLogin = vi.mocked(isLogin)
const mockGetConversations = vi.mocked(getConversations)
const mockGetUnreadCounts = vi.mocked(getUnreadCounts)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/message', component: Message },
    { path: '/m/login', component: { template: '<div />' } },
    { path: '/m/message/chat/:id', component: { template: '<div />' } },
    { path: '/m/message/notify/:type', component: { template: '<div />' } }
  ]
})

describe('Message.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockIsLogin.mockReturnValue(true)
    mockGetConversations.mockResolvedValue({ code: '1', data: [] })
    mockGetUnreadCounts.mockResolvedValue({ code: '1', data: { reply: 0, at: 0, like: 0, system: 0 } })
    router.push('/m/message')
    await router.isReady()
  })

  it('未登录 → 跳转 /m/login', async () => {
    mockIsLogin.mockReturnValue(false)
    mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/login')
  })

  it('已登录 → 加载对话列表', async () => {
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, targetUserNickname: '张三', lastMessageContent: '你好' }]
    })
    mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(mockGetConversations).toHaveBeenCalled()
    expect(mockGetUnreadCounts).toHaveBeenCalled()
  })

  it('有未读消息 → 显示 badge', async () => {
    mockGetUnreadCounts.mockResolvedValue({
      code: '1',
      data: { reply: 3, at: 2, like: 0, system: 0 }
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    const badge = wrapper.find('.badge')
    expect(badge.exists()).toBe(true)
    expect(badge.text()).toBe('5') // reply(3) + at(2)
  })

  it('无未读消息 → 不显示 badge', async () => {
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('.badge').exists()).toBe(false)
  })

  it('点击对话 → 跳转 chat 页面', async () => {
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, targetUserId: 42, targetUserNickname: '张三', lastMessageContent: 'hi' }]
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()

    await wrapper.find('.chat-item').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/message/chat/42')
  })

  it('formatTime: 今天 → 显示时分', async () => {
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, lastMessageTime: new Date().toISOString() }]
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    const time = wrapper.find('.time')
    // 今天的时间应该包含 ':' (如 "14:30")
    expect(time.text()).toMatch(/\d{1,2}:\d{2}/)
  })

  it('formatTime: 昨天 → "昨天"', async () => {
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, lastMessageTime: yesterday.toISOString() }]
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('.time').text()).toBe('昨天')
  })

  it('formatTime: 更早 → 月日格式', async () => {
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, lastMessageTime: '2025-06-15T10:00:00Z' }]
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('.time').text()).toMatch(/\d+月\d+日/)
  })

  it('formatTime: 空字符串 → 空', async () => {
    mockGetConversations.mockResolvedValue({
      code: '1',
      data: [{ id: 1, lastMessageTime: '' }]
    })
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('.time').text()).toBe('')
  })

  it('对话列表为空 → 显示 "暂无消息"', async () => {
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.find('.empty-state').text()).toContain('暂无消息')
  })

  it('点击 "回复与@" → 跳转 /m/message/notify/reply', async () => {
    const wrapper = mount(Message, { global: { plugins: [router] } })
    await flushPromises()
    const tabs = wrapper.findAll('.tab-item')
    await tabs[0].trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/m/message/notify/reply')
  })
})
