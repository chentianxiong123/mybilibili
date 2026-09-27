import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ScrollToTop from '../ScrollToTop.vue'

describe('ScrollToTop.vue', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('默认不可见（scrollY < 500）', () => {
    Object.defineProperty(window, 'scrollY', { value: 100, writable: true })
    const wrapper = mount(ScrollToTop)
    expect(wrapper.find('.to-top').exists()).toBe(false)
  })

  it('scrollY > 500 时可见', async () => {
    const wrapper = mount(ScrollToTop)
    Object.defineProperty(window, 'scrollY', { value: 600, writable: true })
    window.dispatchEvent(new Event('scroll'))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.to-top').exists()).toBe(true)
  })

  it('scrollY < 500 时不可见', async () => {
    Object.defineProperty(window, 'scrollY', { value: 600, writable: true })
    const wrapper = mount(ScrollToTop)
    window.dispatchEvent(new Event('scroll'))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.to-top').exists()).toBe(true)

    Object.defineProperty(window, 'scrollY', { value: 100, writable: true })
    window.dispatchEvent(new Event('scroll'))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.to-top').exists()).toBe(false)
  })

  it('点击按钮 → scrollTo top', async () => {
    const scrollToSpy = vi.fn()
    window.scrollTo = scrollToSpy
    Object.defineProperty(window, 'scrollY', { value: 600, writable: true })
    const wrapper = mount(ScrollToTop)
    window.dispatchEvent(new Event('scroll'))
    await wrapper.vm.$nextTick()
    await wrapper.find('.to-top').trigger('click')
    expect(scrollToSpy).toHaveBeenCalledWith({ top: 0, behavior: 'smooth' })
  })
})
