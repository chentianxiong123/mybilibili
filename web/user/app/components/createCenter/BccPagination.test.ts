import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BccPagination from './BccPagination.vue'

describe('BccPagination', () => {
  it('renders single page without prev/next when only one page', () => {
    const wrapper = mount(BccPagination, {
      props: { total: 5, pageSize: 10, currentPage: 1 }
    })
    const items = wrapper.findAll('.bcc-pagination-item')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toBe('1')
    expect(items[0].classes()).toContain('selected')
  })

  it('renders prev next and numbers for multiple pages', () => {
    const wrapper = mount(BccPagination, {
      props: { total: 88, pageSize: 10, currentPage: 1 }
    })
    const texts = wrapper.findAll('.bcc-pagination-item').map(n => n.text())
    expect(texts).toContain('下一页')
    // 第一页应该没有"上一页"按钮
    expect(texts).not.toContain('上一页')
    // 数字页含首尾页
    expect(texts).toContain('1')
    expect(texts).toContain('9')
    // 省略号存在
    const ellipsis = wrapper.findAll('.bcc-pagination-item').map(n => n.text()).filter(t => t === '⋯')
    expect(ellipsis.length).toBeGreaterThan(0)
  })

  it('emits current-change when clicking a page number', async () => {
    const wrapper = mount(BccPagination, {
      props: { total: 88, pageSize: 10, currentPage: 5 }
    })
    // currentPage=5 时生成 1 ⋯ 4 5 6 ⋯ 9
    const pageBtn = wrapper.findAll('.bcc-pagination-item').find(n => n.text() === '6')
    expect(pageBtn).toBeTruthy()
    await pageBtn.trigger('click')
    const emitted = wrapper.emitted('current-change')
    expect(emitted).toBeTruthy()
    expect(emitted[emitted.length - 1]).toEqual([6])
  })

  it('shows total pages and jump input in extra', () => {
    const wrapper = mount(BccPagination, {
      props: { total: 88, pageSize: 10, currentPage: 2 }
    })
    const extra = wrapper.find('.bcc-pagination-extra').text()
    expect(extra).toContain('共9页')
    expect(extra).toContain('/ 88个')
    expect(wrapper.find('.bcc-pagination-input').exists()).toBe(true)
  })
})