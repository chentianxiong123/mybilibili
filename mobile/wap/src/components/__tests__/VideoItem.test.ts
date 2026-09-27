import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import VideoItem from '../VideoItem.vue'

vi.mock('../../utils/session', () => ({
  isLogin: vi.fn().mockReturnValue(true),
  getLocalUser: vi.fn().mockReturnValue({ id: 1 })
}))

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/m/video/:aId', component: { template: '<div />' } },
    { path: '/m/vertical/:aId', component: { template: '<div />' } }
  ]
})

describe('VideoItem.vue', () => {
  const mockVideo = {
    aId: 123,
    title: '测试视频标题',
    pic: 'https://example.com/cover.jpg',
    author: '测试UP主',
    play: 12345,
    videoReview: 678,
    duration: '10:30',
    isVertical: 0
  }

  it('渲染标题和作者', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.title').text()).toBe('测试视频标题')
    expect(wrapper.find('.author-name').text()).toBe('测试UP主')
  })

  it('渲染播放量（万单位）', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('1.2万')
  })

  it('渲染弹幕数', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.danmaku').text()).toContain('678')
  })

  it('渲染时长', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.duration').text()).toBe('10:30')
  })

  it('showStatistics=false 时隐藏统计', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo, showStatistics: false },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').exists()).toBe(false)
    expect(wrapper.find('.danmaku').exists()).toBe(false)
    // duration 仍然显示
    expect(wrapper.find('.duration').text()).toBe('10:30')
  })

  it('无作者时不渲染 author-row', () => {
    const videoNoAuthor = { ...mockVideo, author: '' }
    const wrapper = mount(VideoItem, {
      props: { video: videoNoAuthor },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.author-row').exists()).toBe(false)
  })

  it('竖屏视频链接到 /m/vertical/', () => {
    const verticalVideo = { ...mockVideo, isVertical: 1 }
    const wrapper = mount(VideoItem, {
      props: { video: verticalVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('a').attributes('href')).toBe('/m/vertical/123')
  })

  it('横屏视频链接到 /m/video/', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('a').attributes('href')).toBe('/m/video/123')
  })

  it('播放量=0 显示 "0"', () => {
    const video = { ...mockVideo, play: 0, videoReview: 0 }
    const wrapper = mount(VideoItem, {
      props: { video },
      global: { plugins: [router] }
    })
    expect(wrapper.find('.play').text()).toContain('0')
    expect(wrapper.find('.danmaku').text()).toContain('0')
  })

  it('图片设置 alt 属性', () => {
    const wrapper = mount(VideoItem, {
      props: { video: mockVideo },
      global: { plugins: [router] }
    })
    expect(wrapper.find('img').attributes('alt')).toBe('测试视频标题')
  })
})
