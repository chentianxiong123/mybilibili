import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AdminAiFloatingButton from './AdminAiFloatingButton.vue'

function makeWrapper(props: any = {}) {
  return mount(AdminAiFloatingButton, {
    props: { visible: false, ...props },
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
      },
    },
  })
}

describe('AdminAiFloatingButton.vue', () => {
  it('渲染悬浮按钮（含 aria-label）', () => {
    const w = makeWrapper()
    const btn = w.find('.floating-button')
    expect(btn.exists()).toBe(true)
    expect(btn.attributes('aria-label')).toBe('打开管理助手')
  })

  it('visible=false 时无 active class', () => {
    const w = makeWrapper({ visible: false })
    expect(w.find('.floating-button').classes()).not.toContain('active')
  })

  it('visible=true 时有 active class', () => {
    const w = makeWrapper({ visible: true })
    expect(w.find('.floating-button').classes()).toContain('active')
  })

  it('默认 visible=false', () => {
    const w = mount(AdminAiFloatingButton, { global: { stubs: { 'el-icon': { template: '<span />' } } } })
    expect(w.find('.floating-button').classes()).not.toContain('active')
  })

  it('点击时 emit update:visible（取反）', async () => {
    const w = makeWrapper({ visible: false })
    await w.find('.floating-button').trigger('click')
    expect(w.emitted('update:visible')).toBeTruthy()
    expect((w.emitted('update:visible') as any[])[0]).toEqual([true])
  })

  it('visible=true 时点击 emit false', async () => {
    const w = makeWrapper({ visible: true })
    await w.find('.floating-button').trigger('click')
    expect((w.emitted('update:visible') as any[])[0]).toEqual([false])
  })

  it('点击两次分别 emit true/false（v-model 下 toggle）', async () => {
    const w = makeWrapper({ visible: false })
    const btn = w.find('.floating-button')
    await btn.trigger('click')
    expect((w.emitted('update:visible') as any[])[0]).toEqual([true])
    // 模拟父组件 v-model 回写
    await w.setProps({ visible: true })
    await btn.trigger('click')
    const emits = w.emitted('update:visible') as any[]
    expect(emits).toHaveLength(2)
    expect(emits[1]).toEqual([false])
  })
})