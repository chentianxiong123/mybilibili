import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Drawer from '../Drawer.vue'

describe('Drawer.vue', () => {
  const mockData = [
    { id: 1, name: '选项1' },
    { id: 2, name: '选项2' },
    { id: 3, name: '选项3' }
  ]

  it('默认不可见', () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    expect(wrapper.find('.drawer-overlay').exists()).toBe(false)
  })

  it('调用 show() 后可见', async () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    wrapper.vm.show()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.drawer-overlay').exists()).toBe(true)
  })

  it('渲染所有选项', async () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    wrapper.vm.show()
    await wrapper.vm.$nextTick()
    const items = wrapper.findAll('.drawer-item')
    expect(items.length).toBe(3)
    expect(items[0].text()).toBe('选项1')
    expect(items[2].text()).toBe('选项3')
  })

  it('点击选项 → emit click 事件 + 关闭', async () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    wrapper.vm.show()
    await wrapper.vm.$nextTick()
    await wrapper.findAll('.drawer-item')[1].trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click')![0]).toEqual([mockData[1]])
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.drawer-overlay').exists()).toBe(false)
  })

  it('点击遮罩层 → 关闭', async () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    wrapper.vm.show()
    await wrapper.vm.$nextTick()
    await wrapper.find('.drawer-overlay').trigger('click')
    expect(wrapper.find('.drawer-overlay').exists()).toBe(false)
  })

  it('点击面板内容 → 不关闭（stop propagation）', async () => {
    const wrapper = mount(Drawer, { props: { data: mockData } })
    wrapper.vm.show()
    await wrapper.vm.$nextTick()
    await wrapper.find('.drawer-panel').trigger('click')
    expect(wrapper.find('.drawer-overlay').exists()).toBe(true)
  })
})
