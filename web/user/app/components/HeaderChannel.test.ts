import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import HeaderChannel from './HeaderChannel.vue'

describe('HeaderChannel', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('默认渲染 2 个 iconCards: 动态/热门', () => {
    const w = mount(HeaderChannel)
    const items = w.findAll('.v-channel-icon-item')
    expect(items.length).toBe(2)
    expect(items[0].text()).toContain('动态')
    expect(items[1].text()).toContain('热门')
  })

  it('iconCards 自定义', () => {
    const iconCards = [
      { name: '番剧', icon: 'icon-test', color: '#000', href: '/bangumi' },
    ]
    const w = mount(HeaderChannel, { props: { iconCards } })
    expect(w.findAll('.v-channel-icon-item').length).toBe(1)
    expect(w.text()).toContain('番剧')
  })

  it('channels 透传渲染', () => {
    const channels = [
      { id: 1, name: '动画' },
      { id: 2, name: '音乐' },
      { id: 3, name: '游戏' },
    ]
    const w = mount(HeaderChannel, { props: { channels } })
    const items = w.findAll('.v-channel-link')
    expect(items.length).toBe(3)
    expect(items[0].text()).toBe('动画')
  })

  it('channel id 缺省用 name 作 key', () => {
    const channels = [{ name: 'no-id' }]
    const w = mount(HeaderChannel, { props: { channels } })
    expect(w.findAll('.v-channel-link').length).toBe(1)
  })

  it('icon card click 触发 window.location.href', async () => {
    const originalLocation = window.location
    // happy-dom 中 location 是只读 — 使用 stub
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, href: '' },
      writable: true,
      configurable: true,
    })
    const w = mount(HeaderChannel)
    await w.find('.v-channel-icon-item').trigger('click')
    expect(window.location.href).toBe('/dynamic')
    Object.defineProperty(window, 'location', { value: originalLocation, configurable: true })
  })

  it('icon card 无 href 不跳转', async () => {
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, href: '' },
      writable: true,
      configurable: true,
    })
    const w = mount(HeaderChannel, { props: { iconCards: [{ name: 'x', icon: 'i' }] } })
    await w.find('.v-channel-icon-item').trigger('click')
    expect(window.location.href).toBe('')
    Object.defineProperty(window, 'location', { value: originalLocation, configurable: true })
  })

  it('icon color 通过 css var 透传', () => {
    const iconCards = [{ name: 'x', icon: 'i', color: '#123456', href: '/a' }]
    const w = mount(HeaderChannel, { props: { iconCards } })
    const style = w.find('.v-channel-icon-bg').attributes('style') || ''
    expect(style).toContain('--icon-bg: #123456')
  })

  it('channels 默认空数组', () => {
    const w = mount(HeaderChannel)
    expect(w.findAll('.v-channel-link').length).toBe(0)
  })

  it('iconCards 默认 color 含 css var', () => {
    const w = mount(HeaderChannel)
    const style = w.findAll('.v-channel-icon-bg')[0].attributes('style') || ''
    expect(style).toContain('--icon-bg:')
  })

  it('channels 与 iconCards 同时存在都渲染', () => {
    const w = mount(HeaderChannel, {
      props: {
        iconCards: [{ name: 'a', icon: 'i' }],
        channels: [{ id: 1, name: 'A' }, { id: 2, name: 'B' }],
      },
    })
    expect(w.findAll('.v-channel-icon-item').length).toBe(1)
    expect(w.findAll('.v-channel-link').length).toBe(2)
  })
})
