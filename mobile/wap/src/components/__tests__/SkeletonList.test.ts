import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SkeletonList from '../SkeletonList.vue'

describe('SkeletonList.vue', () => {
  it('渲染 grid 形态双列视频卡片骨架', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'grid', count: 4 } })
    expect(wrapper.find('.skeleton-grid').exists()).toBe(true)
    expect(wrapper.findAll('.sk-card').length).toBe(4)
    expect(wrapper.findAll('.sk-cover').length).toBe(4)
    expect(wrapper.findAll('.sk-line').length).toBeGreaterThan(0)
  })

  it('渲染 row 形态横向卡片骨架', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'row', count: 3 } })
    expect(wrapper.find('.skeleton-row-list').exists()).toBe(true)
    expect(wrapper.findAll('.sk-row').length).toBe(3)
    expect(wrapper.findAll('.sk-row__cover').length).toBe(3)
  })

  it('渲染 icon 形态图标网格骨架', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'icon', count: 8 } })
    expect(wrapper.find('.skeleton-icon-grid').exists()).toBe(true)
    expect(wrapper.findAll('.sk-icon').length).toBe(8)
    expect(wrapper.findAll('.sk-icon__circle').length).toBe(8)
  })

  it('渲染 avatar 形态头像列表骨架', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'avatar', count: 5 } })
    expect(wrapper.find('.skeleton-avatar-list').exists()).toBe(true)
    expect(wrapper.findAll('.sk-avatar-row').length).toBe(5)
    expect(wrapper.findAll('.sk-avatar-row__circle').length).toBe(5)
  })

  it('渲染 feed 形态动态卡片骨架', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'feed', count: 2 } })
    expect(wrapper.find('.skeleton-feed-list').exists()).toBe(true)
    expect(wrapper.findAll('.sk-feed').length).toBe(2)
    expect(wrapper.findAll('.sk-feed__image').length).toBe(6)
  })

  it('count 默认值为 6', () => {
    const wrapper = mount(SkeletonList, { props: { variant: 'grid' } })
    expect(wrapper.findAll('.sk-card').length).toBe(6)
  })

  it('默认形态为 grid', () => {
    const wrapper = mount(SkeletonList)
    expect(wrapper.find('.skeleton-grid').exists()).toBe(true)
  })
})
