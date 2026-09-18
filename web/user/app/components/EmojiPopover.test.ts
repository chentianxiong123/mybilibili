import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmojiPopover from './EmojiPopover.vue'

function queryAll(selector: string) {
  return Array.from(document.body.querySelectorAll(selector))
}

describe('EmojiPopover', () => {
  it('默认 visible=false 时 popover 存在但隐藏', () => {
    const w = mount(EmojiPopover, {
      attachTo: document.body,
      props: { visible: false },
    })
    const popover = queryAll('.emoji-popover')
    expect(popover.length).toBe(1)
    expect(popover[0].getAttribute('style') || '').toContain('display: none')
    w.unmount()
  })

  it('visible=true 显示 64 个 emoji', () => {
    const w = mount(EmojiPopover, {
      attachTo: document.body,
      props: { visible: true },
    })
    const items = queryAll('.emoji-item')
    expect(items.length).toBe(64)
    expect(items[0].textContent).toBe('😀')
    w.unmount()
  })

  it('点击 emoji 触发 select 事件并关闭', async () => {
    const w = mount(EmojiPopover, {
      attachTo: document.body,
      props: { visible: true },
    })
    const items = queryAll('.emoji-item')
    items[0].click()
    await new Promise(r => setTimeout(r, 0))
    expect(w.emitted('select')).toBeTruthy()
    expect((w.emitted('select') as any[])[0][0]).toBe('😀')
    expect((w.emitted('update:visible') as any[])[0]).toEqual([false])
    w.unmount()
  })

  it('click outside 触发 update:visible=false', async () => {
    const trigger = document.createElement('div')
    document.body.appendChild(trigger)
    const w = mount(EmojiPopover, {
      attachTo: document.body,
      props: { visible: true, triggerRef: trigger },
    })
    // 派发 click 到 body 的非 popover/trigger 区域
    const outside = document.createElement('div')
    document.body.appendChild(outside)
    outside.click()
    await new Promise(r => setTimeout(r, 0))
    expect(w.emitted('update:visible')).toBeTruthy()
    w.unmount()
    document.body.removeChild(trigger)
    document.body.removeChild(outside)
  })

  it('emojiList 包含表情', () => {
    const w = mount(EmojiPopover, {
      attachTo: document.body,
      props: { visible: true },
    })
    const items = queryAll('.emoji-item')
    expect(items.length).toBe(64)
    // 抽查几个
    const texts = items.map(i => i.textContent)
    expect(texts).toContain('❤️')
    expect(texts).toContain('👍')
    expect(texts).toContain('🤝')
    w.unmount()
  })
})
