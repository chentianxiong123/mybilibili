import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import LevelBadge from './LevelBadge.vue'

describe('LevelBadge', () => {
  it('renders LV{n} for level > 0', () => {
    const w = mount(LevelBadge, { props: { level: 5 } })
    expect(w.text()).toBe('LV5')
  })

  it('does not render when level <= 0', () => {
    const w = mount(LevelBadge, { props: { level: 0 } })
    expect(w.find('span').exists()).toBe(false)
  })

  it('does not render when level is negative', () => {
    const w = mount(LevelBadge, { props: { level: -1 } })
    expect(w.find('span').exists()).toBe(false)
  })

  it('uses level-low for level 1-3', () => {
    const w = mount(LevelBadge, { props: { level: 1 } })
    expect(w.classes()).toContain('level-badge')
    expect(w.classes()).toContain('level-low')
  })

  it('uses level-mid for level 4-5', () => {
    const w4 = mount(LevelBadge, { props: { level: 4 } })
    const w5 = mount(LevelBadge, { props: { level: 5 } })
    expect(w4.classes()).toContain('level-mid')
    expect(w5.classes()).toContain('level-mid')
  })

  it('uses level-high for level >= 6', () => {
    const w = mount(LevelBadge, { props: { level: 6 } })
    expect(w.classes()).toContain('level-high')
  })

  it('title attribute is "LV{n}"', () => {
    const w = mount(LevelBadge, { props: { level: 7 } })
    expect(w.attributes('title')).toBe('LV7')
  })
})
