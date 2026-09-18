import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import VAvatar from './VAvatar.vue'

describe('VAvatar', () => {
  it('默认 size=40 + rounded=true', () => {
    const w = mount(VAvatar)
    const root = w.find('.v-avatar')
    expect(root.exists()).toBe(true)
    expect(root.classes()).toContain('is-rounded')
    expect(root.attributes('style')).toContain('width: 40px')
    expect(root.attributes('style')).toContain('height: 40px')
  })

  it('无 src 时显示 placeholder', () => {
    const w = mount(VAvatar, { props: { src: '' } })
    expect(w.find('img').exists()).toBe(false)
    expect(w.find('.v-avatar-placeholder').exists()).toBe(true)
  })

  it('有 src 时显示 img 且 alt 透传', () => {
    const w = mount(VAvatar, { props: { src: '/a.png', alt: 'm' } })
    const img = w.find('img')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe('/a.png')
    expect(img.attributes('alt')).toBe('m')
  })

  it('size 透传到 style', () => {
    const w = mount(VAvatar, { props: { size: 80 } })
    expect(w.attributes('style')).toContain('width: 80px')
    expect(w.attributes('style')).toContain('height: 80px')
  })

  it('rounded=false 不加 is-rounded 类', () => {
    const w = mount(VAvatar, { props: { rounded: false } })
    expect(w.classes()).not.toContain('is-rounded')
  })

  it('auth=0 不显示徽标', () => {
    const w = mount(VAvatar, { props: { auth: 0 } })
    expect(w.find('.v-avatar-badge').exists()).toBe(false)
  })

  it('auth=1 黄色徽标', () => {
    const w = mount(VAvatar, { props: { auth: 1 } })
    const badge = w.find('.v-avatar-badge')
    expect(badge.exists()).toBe(true)
    expect(badge.attributes('style')).toContain('background: #FFC62E')
  })

  it('auth=2 蓝色徽标', () => {
    const w = mount(VAvatar, { props: { auth: 2 } })
    const badge = w.find('.v-avatar-badge')
    expect(badge.exists()).toBe(true)
    expect(badge.attributes('style')).toContain('background: #4AC7FF')
  })

  it('默认 alt 为 "avatar"', () => {
    const w = mount(VAvatar, { props: { src: '/x.png' } })
    expect(w.find('img').attributes('alt')).toBe('avatar')
  })
})
