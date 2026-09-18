import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import VideoDescription from './VideoDescription.vue'

function makeWrapper(props: any = {}) {
  return mount(VideoDescription, {
    props: { description: 'desc', tags: [], ...props },
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
      },
    },
  })
}

describe('VideoDescription', () => {
  it('渲染 description 文本', () => {
    const w = makeWrapper({ description: 'video desc' })
    expect(w.text()).toContain('video desc')
  })

  it('空 description 时显示 "该视频暂无简介"', () => {
    const w = makeWrapper({ description: '' })
    expect(w.text()).toContain('该视频暂无简介')
  })

  it('默认 is-collapsed class 存在', () => {
    const w = makeWrapper({ description: 'x' })
    expect(w.find('.description-content').classes()).toContain('is-collapsed')
    expect(w.text()).toContain('展开')
  })

  it('点击 toggle 切换折叠状态', async () => {
    const w = makeWrapper({ description: 'x' })
    await w.find('.description-toggle').trigger('click')
    expect(w.find('.description-content').classes()).not.toContain('is-collapsed')
    expect(w.text()).toContain('收起')
  })

  it('再次 toggle 折叠回去', async () => {
    const w = makeWrapper({ description: 'x' })
    await w.find('.description-toggle').trigger('click')
    await w.find('.description-toggle').trigger('click')
    expect(w.find('.description-content').classes()).toContain('is-collapsed')
  })

  it('无 tags 时不渲染 tags 区', () => {
    const w = makeWrapper({ description: 'x', tags: [] })
    expect(w.find('.video-tags').exists()).toBe(false)
  })

  it('tags 渲染每个 tag', () => {
    const w = makeWrapper({ description: 'x', tags: ['a', 'b', 'c'] })
    const tagEls = w.findAll('.tag-item')
    expect(tagEls.length).toBe(3)
    expect(tagEls[0].text()).toBe('a')
    expect(tagEls[1].text()).toBe('b')
  })

  it('点击 tag 触发 tagSearch 事件', async () => {
    const w = makeWrapper({ description: 'x', tags: ['music', 'pop'] })
    await w.findAll('.tag-item')[0].trigger('click')
    expect(w.emitted('tagSearch')).toBeTruthy()
    expect((w.emitted('tagSearch') as any[])[0]).toEqual(['music'])
  })

  it('多个 tag 点击分别触发对应事件', async () => {
    const w = makeWrapper({ description: 'x', tags: ['x', 'y'] })
    await w.findAll('.tag-item')[1].trigger('click')
    expect((w.emitted('tagSearch') as any[])[0]).toEqual(['y'])
  })
})
