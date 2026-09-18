import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DynamicCard from './DynamicCard.vue'

vi.mock('@/components/CommentSystem.vue', () => ({
  default: { name: 'CommentSystemStub', template: '<div class="comment-system-stub" />' },
}))

vi.mock('@element-plus/icons-vue', () => ({
  MoreFilled: { template: '<i />' },
  Share: { template: '<i />' },
  ChatDotRound: { template: '<i />' },
  Star: { template: '<i />' },
  VideoPlay: { template: '<i />' },
  Clock: { template: '<i />' },
  View: { template: '<i />' },
}))

function makeItem(overrides: any = {}) {
  const now = new Date()
  return {
    id: 1,
    userId: 9,
    content: 'hello world',
    createdAt: now.toISOString(),
    commentCount: 3,
    shareCount: 5,
    user: { username: 'alice', avatar: '/a.png' },
    stats: { isLiked: false, likeCount: 0 },
    ...overrides,
  }
}

function makeWrapper(item: any) {
  return mount(DynamicCard, {
    props: { item },
    global: {
      stubs: {
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-collapse-transition': { template: '<div class="collapse-stub"><slot /></div>' },
      },
    },
  })
}

describe('DynamicCard', () => {
  it('渲染用户名/内容/默认头像', () => {
    const w = makeWrapper(makeItem())
    expect(w.text()).toContain('alice')
    expect(w.text()).toContain('hello world')
    expect(w.find('.dynamic-avatar').attributes('src')).toBe('/a.png')
  })

  it('无 user.avatar 时使用 dicebear 默认头像', () => {
    const w = makeWrapper(makeItem({ user: { username: 'bob' } }))
    expect(w.find('.dynamic-avatar').attributes('src')).toContain('dicebear')
  })

  it('无 username 时显示 "用户"', () => {
    const w = makeWrapper(makeItem({ user: {} }))
    expect(w.text()).toContain('用户')
  })

  it('formatTime 刚刚', () => {
    const w = makeWrapper(makeItem({ createdAt: new Date().toISOString() }))
    expect(w.text()).toContain('刚刚')
  })

  it('formatTime 分钟前', () => {
    const w = makeWrapper(makeItem({ createdAt: new Date(Date.now() - 5 * 60 * 1000).toISOString() }))
    expect(w.text()).toContain('5分钟前')
  })

  it('formatTime 天前', () => {
    const w = makeWrapper(makeItem({ createdAt: new Date(Date.now() - 2 * 86400000).toISOString() }))
    expect(w.text()).toContain('2天前')
  })

  it('有 imageUrls 时渲染图片', () => {
    const w = makeWrapper(makeItem({ imageUrls: ['/a.jpg', '/b.jpg'] }))
    expect(w.findAll('.dynamic-image').length).toBe(2)
  })

  it('无 imageUrls 不渲染图片区', () => {
    const w = makeWrapper(makeItem())
    expect(w.findAll('.dynamic-image').length).toBe(0)
  })

  it('refVideo 有 cover 时显示封面', () => {
    const w = makeWrapper(makeItem({ refManuscriptId: 7, refVideo: { title: 'T', cover: '/c.jpg' } }))
    expect(w.find('.video-cover').exists()).toBe(true)
    expect(w.find('.video-cover').attributes('src')).toBe('/c.jpg')
  })

  it('refVideo 无 cover 时显示占位', () => {
    const w = makeWrapper(makeItem({ refManuscriptId: 7, refVideo: { title: 'T' } }))
    expect(w.find('.video-cover-placeholder').exists()).toBe(true)
  })

  it('formatNumber > 10000 显示万', () => {
    const w = makeWrapper(makeItem({ refManuscriptId: 7, refVideo: { title: 'T', viewCount: 15000 } }))
    expect(w.text()).toContain('1.5万')
  })

  it('操作按钮触发对应事件', async () => {
    const w = makeWrapper(makeItem())
    const buttons = w.findAll('.action-btn')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')
    expect(w.emitted('forward')).toBeTruthy()
    expect(w.emitted('toggle-comment')).toBeTruthy()
    expect(w.emitted('like')).toBeTruthy()
  })

  it('go-to-user 事件携带 userId', async () => {
    const w = makeWrapper(makeItem())
    await w.find('.dynamic-username').trigger('click')
    expect(w.emitted('go-to-user')).toBeTruthy()
    expect((w.emitted('go-to-user') as any[])[0]).toEqual([9])
  })

  it('go-to-manuscript 事件携带 refManuscriptId', async () => {
    const w = makeWrapper(makeItem({ refManuscriptId: 12, refVideo: { title: 'X' } }))
    await w.find('.video-card').trigger('click')
    expect(w.emitted('go-to-manuscript')).toBeTruthy()
    expect((w.emitted('go-to-manuscript') as any[])[0]).toEqual([12])
  })

  it('showComments=true 时渲染 CommentSystem', () => {
    const w = makeWrapper(makeItem({ showComments: true }))
    expect(w.find('.comment-system-stub').exists()).toBe(true)
  })

  it('like 按钮在 liked=true 时有 liked class', () => {
    const w = makeWrapper(makeItem({ stats: { isLiked: true, likeCount: 7 } }))
    const likeBtn = w.findAll('.action-btn')[2]
    expect(likeBtn.classes()).toContain('liked')
    expect(w.text()).toContain('7')
  })

  it('comment 按钮在 showComments=true 时有 active class', () => {
    const w = makeWrapper(makeItem({ showComments: true }))
    const commentBtn = w.findAll('.action-btn')[1]
    expect(commentBtn.classes()).toContain('active')
  })
})
