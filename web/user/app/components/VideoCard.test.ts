import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import VideoCard from './VideoCard.vue'

function makeWrapper(video: any = {}, showVideoCount = false) {
  return mount(VideoCard, {
    props: { video, showVideoCount },
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
      },
    },
  })
}

describe('VideoCard', () => {
  it('渲染标题与封面', () => {
    const w = makeWrapper({ id: 1, title: 'Hello', coverUrl: '/c.jpg' })
    expect(w.text()).toContain('Hello')
    expect(w.find('.video-cover img').attributes('src')).toBe('/c.jpg')
  })

  it('无 coverUrl 使用默认 placeholder', () => {
    const w = makeWrapper({ id: 1, title: 'x' })
    expect(w.find('.video-cover img').attributes('src')).toContain('placeholder-cover.svg')
  })

  it('manuscriptId 优先作为链接', () => {
    const w = makeWrapper({ id: 1, manuscriptId: 99, title: 'X' })
    expect(w.find('.video-cover-link').attributes('href')).toBe('/manuscript/99')
    expect(w.find('.video-title-text').attributes('href')).toBe('/manuscript/99')
  })

  it('无 manuscriptId 回退到 id', () => {
    const w = makeWrapper({ id: 5, title: 'X' })
    expect(w.find('.video-cover-link').attributes('href')).toBe('/manuscript/5')
  })

  it('sourceType=bilibili 显示 B 站徽标', () => {
    const w = makeWrapper({ id: 1, sourceType: 'bilibili' })
    expect(w.find('.source-badge').exists()).toBe(true)
    expect(w.text()).toContain('B站')
  })

  it('sourceType 非 bilibili 不显示徽标', () => {
    const w = makeWrapper({ id: 1, sourceType: 'self' })
    expect(w.find('.source-badge').exists()).toBe(false)
  })

  it('duration 显示', () => {
    const w = makeWrapper({ id: 1, duration: '05:30' })
    expect(w.find('.video-duration').text()).toBe('05:30')
  })

  it('无 duration 显示默认 00:00', () => {
    const w = makeWrapper({ id: 1 })
    expect(w.find('.video-duration').text()).toBe('00:00')
  })

  it('viewText < 10000 显示原值带千位', () => {
    const w = makeWrapper({ id: 1, viewCount: 1500 })
    expect(w.text()).toContain('1,500')
  })

  it('viewText >= 10000 显示万', () => {
    const w = makeWrapper({ id: 1, viewCount: 15000 })
    expect(w.text()).toContain('1.5万')
  })

  it('viewText >= 100000 显示整数万', () => {
    const w = makeWrapper({ id: 1, viewCount: 250000 })
    expect(w.text()).toContain('25万')
  })

  it('viewText 字符串数字透传', () => {
    const w = makeWrapper({ id: 1, viewCount: '5000' })
    expect(w.text()).toContain('5,000')
  })

  it('viewText undefined 显示 0', () => {
    const w = makeWrapper({ id: 1 })
    expect(w.text()).toContain('0')
  })

  it('commentText 千位', () => {
    const w = makeWrapper({ id: 1, commentCount: 1234 })
    expect(w.text()).toContain('1,234')
  })

  it('showVideoCount=true 且 videoCount>1 显示 N P', () => {
    const w = makeWrapper({ id: 1, videoCount: 5 }, true)
    expect(w.find('.video-count').exists()).toBe(true)
    expect(w.text()).toContain('5P')
  })

  it('showVideoCount=true 但 videoCount<=1 不显示', () => {
    const w = makeWrapper({ id: 1, videoCount: 1 }, true)
    expect(w.find('.video-count').exists()).toBe(false)
  })

  it('showVideoCount=false 不显示 P 标签', () => {
    const w = makeWrapper({ id: 1, videoCount: 5 })
    expect(w.find('.video-count').exists()).toBe(false)
  })

  it('uploader 有 id 时渲染链接', () => {
    const w = makeWrapper({ id: 1, uploader: { id: 7, name: 'Alice' } })
    const author = w.find('.video-author')
    expect(author.attributes('href')).toBe('/user/7')
    expect(author.text()).toBe('Alice')
  })

  it('uploader 无 id 时只显示文字', () => {
    const w = makeWrapper({ id: 1, uploader: { name: 'Bob' } })
    const author = w.find('.video-author')
    expect(author.attributes('href')).toBeUndefined()
    expect(author.text()).toBe('Bob')
  })

  it('无 uploader 时显示 "未知UP主"', () => {
    const w = makeWrapper({ id: 1 })
    expect(w.text()).toContain('未知UP主')
  })

  it('dateText 透传', () => {
    const w = makeWrapper({ id: 1, dateText: '3小时前' })
    expect(w.text()).toContain('3小时前')
  })

  it('无 id 无 manuscriptId 时链接 "#"', () => {
    const w = makeWrapper({ title: 'X' })
    expect(w.find('.video-cover-link').attributes('href')).toBe('#')
  })
})
