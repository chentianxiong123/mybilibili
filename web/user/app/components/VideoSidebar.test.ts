import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import VideoSidebar from './VideoSidebar.vue'

vi.mock('element-plus', () => ({
  ElSkeleton: { template: '<div class="el-skeleton-stub" />' },
}))

vi.mock('@element-plus/icons-vue', () => ({
  ArrowDown: { template: '<i />' },
}))

function makeWrapper(overrides: any = {}) {
  const props = {
    danmuList: [],
    loadingDanmus: false,
    manuscriptInfo: { videos: [] },
    currentVideoIndex: 0,
    relatedVideos: [],
    loadingRelatedVideos: false,
    isDanmuListCollapsed: false,
    isVideoPartsCollapsed: false,
    ...overrides,
  }
  return mount(VideoSidebar, {
    props,
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
      },
    },
  })
}

describe('VideoSidebar', () => {
  it('渲染右侧三块: 弹幕/分P/推荐', () => {
    const w = makeWrapper()
    expect(w.find('.side-danmu-list').exists()).toBe(true)
    expect(w.find('.related-videos').exists()).toBe(true)
    expect(w.text()).toContain('弹幕列表')
    expect(w.text()).toContain('推荐视频')
  })

  it('loadingDanmus=true 渲染 skeleton', () => {
    const w = makeWrapper({ loadingDanmus: true })
    expect(w.find('.loading-danmus .el-skeleton-stub').exists()).toBe(true)
  })

  it('danmuList 空且未 loading 显示 "暂无弹幕"', () => {
    const w = makeWrapper({ danmuList: [] })
    expect(w.text()).toContain('暂无弹幕')
  })

  it('danmuList 有数据时渲染 header 与列表', () => {
    const danmuList = [
      { time: 65, text: 'hello', sendTime: '2024-01-01' },
      { time: 130, text: 'world', sendTime: '2024-01-02' },
    ]
    const w = makeWrapper({ danmuList })
    expect(w.find('.danmu-header').exists()).toBe(true)
    expect(w.findAll('.danmu-item').length).toBe(2)
    expect(w.text()).toContain('hello')
    expect(w.text()).toContain('world')
  })

  it('formatTime 65 → "1:05"', () => {
    const danmuList = [{ time: 65, text: 'x', sendTime: 't' }]
    const w = makeWrapper({ danmuList })
    expect(w.text()).toContain('1:05')
  })

  it('formatTime 3661 → "61:01"', () => {
    const danmuList = [{ time: 3661, text: 'x', sendTime: 't' }]
    const w = makeWrapper({ danmuList })
    expect(w.text()).toContain('61:01')
  })

  it('点击弹幕 jumpToDanmuTime 触发', async () => {
    const danmuList = [{ time: 30, text: 'h', sendTime: 't' }]
    const w = makeWrapper({ danmuList })
    await w.find('.danmu-item').trigger('click')
    expect(w.emitted('jumpToDanmuTime')).toBeTruthy()
    expect((w.emitted('jumpToDanmuTime') as any[])[0]).toEqual([30])
  })

  it('点击弹幕列表头 toggleDanmuList 触发', async () => {
    const w = makeWrapper()
    await w.find('.danmu-list-header').trigger('click')
    expect(w.emitted('toggleDanmuList')).toBeTruthy()
  })

  it('videos 长度 > 1 显示分P区', () => {
    const w = makeWrapper({
      manuscriptInfo: { videos: [{ id: 1, title: 'P1' }, { id: 2, title: 'P2' }] },
    })
    expect(w.find('.video-parts-section').exists()).toBe(true)
    expect(w.text()).toContain('共2P')
    expect(w.findAll('.video-part-item').length).toBe(2)
  })

  it('videos 长度 <= 1 不显示分P区', () => {
    const w = makeWrapper({
      manuscriptInfo: { videos: [{ id: 1, title: 'P1' }] },
    })
    expect(w.find('.video-parts-section').exists()).toBe(false)
  })

  it('点击分P 触发 switchVideoPart', async () => {
    const w = makeWrapper({
      manuscriptInfo: { videos: [{ id: 1, title: 'P1' }, { id: 2, title: 'P2' }] },
    })
    await w.findAll('.video-part-item')[1].trigger('click')
    expect(w.emitted('switchVideoPart')).toBeTruthy()
    expect((w.emitted('switchVideoPart') as any[])[0]).toEqual([1])
  })

  it('currentVideoIndex 匹配时 active', () => {
    const w = makeWrapper({
      manuscriptInfo: { videos: [{ id: 1, title: 'P1' }, { id: 2, title: 'P2' }] },
      currentVideoIndex: 1,
    })
    const items = w.findAll('.video-part-item')
    expect(items[1].classes()).toContain('active')
    expect(items[0].classes()).not.toContain('active')
  })

  it('分P header 点击触发 toggleVideoParts', async () => {
    const w = makeWrapper({
      manuscriptInfo: { videos: [{ id: 1, title: 'P1' }, { id: 2, title: 'P2' }] },
    })
    await w.find('.video-parts-header').trigger('click')
    expect(w.emitted('toggleVideoParts')).toBeTruthy()
  })

  it('loadingRelatedVideos=true 渲染 skeleton', () => {
    const w = makeWrapper({ loadingRelatedVideos: true })
    expect(w.find('.loading-related .el-skeleton-stub').exists()).toBe(true)
  })

  it('relatedVideos 渲染', () => {
    const relatedVideos = [
      { id: 1, title: 'A', author: 'x', authorId: 7, viewCount: 1000, commentCount: 5 },
      { id: 2, title: 'B', author: 'y', authorId: 8, viewCount: 2000, commentCount: 10 },
    ]
    const w = makeWrapper({ relatedVideos })
    expect(w.findAll('.related-video-item').length).toBe(2)
    expect(w.text()).toContain('A')
    expect(w.text()).toContain('x')
    expect(w.text()).toContain('1,000次播放')
  })

  it('relatedVideos 空显示 "暂无相关视频推荐"', () => {
    const w = makeWrapper({ relatedVideos: [] })
    expect(w.text()).toContain('暂无相关视频推荐')
  })

  it('点击标题触发 goToVideo', async () => {
    const relatedVideos = [{ id: 1, title: 'A', author: 'x', authorId: 7 }]
    const w = makeWrapper({ relatedVideos })
    await w.find('.video-title-text').trigger('click')
    expect(w.emitted('goToVideo')).toBeTruthy()
    expect((w.emitted('goToVideo') as any[])[0][0]).toMatchObject({ id: 1 })
  })

  it('点击作者触发 goToAuthor', async () => {
    const relatedVideos = [{ id: 1, title: 'A', author: 'x', authorId: 7 }]
    const w = makeWrapper({ relatedVideos })
    await w.find('.video-author').trigger('click')
    expect(w.emitted('goToAuthor')).toBeTruthy()
    expect((w.emitted('goToAuthor') as any[])[0]).toEqual([7])
  })

  it('relatedVideo cover 缺失使用默认占位', () => {
    const relatedVideos = [{ id: 1, title: 'A', author: 'x' }]
    const w = makeWrapper({ relatedVideos })
    expect(w.find('.video-cover img').attributes('src')).toContain('placeholder-cover.svg')
  })

  it('relatedVideo 有 manuscriptId 拼链接', () => {
    const relatedVideos = [{ id: 1, manuscriptId: 99, title: 'A', author: 'x' }]
    const w = makeWrapper({ relatedVideos })
    expect(w.find('.video-cover-link').attributes('href')).toBe('/manuscript/99')
  })

  it('relatedVideo 无 manuscriptId 用 id 拼链接', () => {
    const relatedVideos = [{ id: 11, title: 'A', author: 'x' }]
    const w = makeWrapper({ relatedVideos })
    expect(w.find('.video-cover-link').attributes('href')).toBe('/manuscript/11')
  })

  it('isDanmuListCollapsed=true 给 items 加 is-hidden', () => {
    const w = makeWrapper({ isDanmuListCollapsed: true })
    expect(w.find('.danmu-items').classes()).toContain('is-hidden')
  })
})
