import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import VideoItem from '../VideoItem.vue'

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn().mockReturnValue(true)
}))

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/m/video/:aId', component: { template: '<div />' } },
    { path: '/m/vertical/:aId', component: { template: '<div />' } }
  ]
})

describe('VideoItem 格式化逻辑', () => {
  const makeVideo = (overrides = {}) => ({
    aId: 1,
    title: '标题',
    pic: 'pic.jpg',
    author: 'UP',
    play: 0,
    videoReview: 0,
    duration: '',
    isVertical: 0,
    ...overrides
  })

  it('formatNum: 0 → "0"', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ play: 0 }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('0')
  })

  it('formatNum: 9999 → "9999"（不转万）', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ play: 9999 }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('9999')
  })

  it('formatNum: 10000 → "1.0万"', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ play: 10000 }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('1.0万')
  })

  it('formatNum: 123456 → "12.3万"', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ play: 123456 }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('12.3万')
  })

  it('无 duration → 不渲染 duration span', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ duration: '' }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.duration').exists()).toBe(false)
  })

  it('有 duration → 渲染 duration span', () => {
    const wrapper = mount(VideoItem, {
      props: { video: makeVideo({ duration: '5:30' }) },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.duration').text()).toBe('5:30')
  })
})
