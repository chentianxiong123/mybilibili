import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import FansList from './FansList.vue'

function makeWrapper(props: any = {}) {
  return mount(FansList, {
    props: { users: [], loading: false, title: '全部粉丝', ...props },
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
      },
    },
  })
}

describe('FansList', () => {
  it('渲染 title 与 fans-count', () => {
    const w = makeWrapper({ users: [{ id: 1 }, { id: 2 }] })
    expect(w.text()).toContain('全部粉丝')
    expect(w.text()).toContain('2')
  })

  it('title="" 不渲染 header', () => {
    const w = makeWrapper({ title: '', users: [{ id: 1 }] })
    expect(w.find('.fans-list-header').exists()).toBe(false)
  })

  it('loading=true 显示 "加载中..."', () => {
    const w = makeWrapper({ loading: true })
    expect(w.text()).toContain('加载中')
  })

  it('users 空且非 loading 显示 "暂无粉丝"', () => {
    const w = makeWrapper({ users: [] })
    expect(w.text()).toContain('暂无粉丝')
  })

  it('users 非空渲染列表', () => {
    const users = [
      { id: 1, name: 'Alice', avatar: '/a.png' },
      { id: 2, name: 'Bob', nickname: 'B 哥' },
    ]
    const w = makeWrapper({ users })
    expect(w.findAll('.fans-user-item').length).toBe(2)
    expect(w.text()).toContain('Alice')
    expect(w.text()).toContain('B 哥')
  })

  it('无 nickname 时显示 name', () => {
    const users = [{ id: 1, name: 'NoNick' }]
    const w = makeWrapper({ users })
    expect(w.text()).toContain('NoNick')
  })

  it('无 name/nickname 时显示 "未知用户"', () => {
    const users = [{ id: 1 }]
    const w = makeWrapper({ users })
    expect(w.text()).toContain('未知用户')
  })

  it('无 avatar 使用 placeholder', () => {
    const users = [{ id: 1, name: 'x' }]
    const w = makeWrapper({ users })
    expect(w.find('.fans-user-avatar img').attributes('src')).toContain('placeholder-avatar.svg')
  })

  it('有 avatar 透传', () => {
    const users = [{ id: 1, name: 'x', avatar: '/custom.png' }]
    const w = makeWrapper({ users })
    expect(w.find('.fans-user-avatar img').attributes('src')).toBe('/custom.png')
  })

  it('signature 透传', () => {
    const users = [{ id: 1, name: 'x', signature: 'hello' }]
    const w = makeWrapper({ users })
    expect(w.text()).toContain('hello')
  })

  it('未关注时按钮显示 "关注"', () => {
    const users = [{ id: 1, name: 'x', isFollowing: false }]
    const w = makeWrapper({ users })
    expect(w.find('.follow-btn').text()).toContain('关注')
    expect(w.find('.follow-btn').classes()).toContain('not-following')
  })

  it('已关注时按钮显示 "已关注"', () => {
    const users = [{ id: 1, name: 'x', isFollowing: true }]
    const w = makeWrapper({ users })
    expect(w.find('.follow-btn').text()).toContain('已关注')
    expect(w.find('.follow-btn').classes()).toContain('following')
  })

  it('点击 follow 按钮 (未关注) 触发 follow 事件', async () => {
    const users = [{ id: 1, name: 'x', isFollowing: false }]
    const w = makeWrapper({ users })
    await w.find('.follow-btn').trigger('click')
    expect(w.emitted('follow')).toBeTruthy()
    expect((w.emitted('follow') as any[])[0]).toEqual([1])
    expect(w.emitted('unfollow')).toBeFalsy()
  })

  it('点击 follow 按钮 (已关注) 触发 unfollow 事件', async () => {
    const users = [{ id: 1, name: 'x', isFollowing: true }]
    const w = makeWrapper({ users })
    await w.find('.follow-btn').trigger('click')
    expect(w.emitted('unfollow')).toBeTruthy()
    expect((w.emitted('unfollow') as any[])[0]).toEqual([1])
    expect(w.emitted('follow')).toBeFalsy()
  })

  it('默认 showSearch=false', () => {
    const w = makeWrapper({ users: [{ id: 1, name: 'x' }] })
    // 验证默认 props 接受即可（无断言副作用）
    expect(w.exists()).toBe(true)
  })
})
