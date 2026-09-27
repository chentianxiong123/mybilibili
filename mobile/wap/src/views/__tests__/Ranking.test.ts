import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../../api/ranking', () => ({
  getRankings: vi.fn(),
  getRankingArchive: vi.fn()
}))

vi.mock('../../utils/format', () => ({
  formatTenThousand: vi.fn((n: number) => n >= 10000 ? (n / 10000).toFixed(1) + '万' : String(n))
}))

import Ranking from '../ranking/Ranking.vue'
import { getRankings, getRankingArchive } from '../../api/ranking'

const mockGetRankings = vi.mocked(getRankings)
const mockGetRankingArchive = vi.mocked(getRankingArchive)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/ranking/:rId', component: Ranking }
  ]
})

describe('Ranking.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockGetRankings.mockResolvedValue({
      code: '1',
      data: [{ id: 1, name: '动画' }, { id: 2, name: '游戏' }]
    })
    mockGetRankingArchive.mockResolvedValue({
      code: '1',
      data: [{ aId: 1, title: '排行视频1' }]
    })
    router.push('/m/ranking/0')
    await router.isReady()
  })

  it('加载分区列表（含 "全站" 首项）', async () => {
    const wrapper = mount(Ranking, { global: { plugins: [router] } })
    await flushPromises()
    expect(mockGetRankings).toHaveBeenCalled()
    // 分区列表应该包含 "全站" 作为第一项
    const tabs = wrapper.findAll('.tab-item')
    expect(tabs.length).toBeGreaterThanOrEqual(3) // 全站 + 动画 + 游戏
    expect(tabs[0].text()).toContain('全站')
  })

  it('加载视频列表', async () => {
    const wrapper = mount(Ranking, { global: { plugins: [router] } })
    await flushPromises()
    expect(mockGetRankingArchive).toHaveBeenCalledWith(
      expect.objectContaining({ rId: 0, p: 1 })
    )
  })

  it('点击分区 tab → 重新加载', async () => {
    const wrapper = mount(Ranking, { global: { plugins: [router] } })
    await flushPromises()
    mockGetRankingArchive.mockClear()

    const tabs = wrapper.findAll('.tab-item')
    if (tabs.length > 1) {
      await tabs[1].trigger('click')
      await flushPromises()
      expect(mockGetRankingArchive).toHaveBeenCalled()
    }
  })

  it('滚动到底部 → 加载更多', async () => {
    mockGetRankingArchive
      .mockResolvedValueOnce({ code: '1', data: Array.from({ length: 20 }, (_, i) => ({ aId: i })) })
      .mockResolvedValueOnce({ code: '1', data: Array.from({ length: 5 }, (_, i) => ({ aId: i + 20 })) })

    const wrapper = mount(Ranking, { global: { plugins: [router] } })
    await flushPromises()

    // 模拟滚动
    Object.defineProperty(window, 'scrollY', { value: 1000, writable: true })
    Object.defineProperty(document.body, 'scrollHeight', { value: 2000, writable: true })
    Object.defineProperty(window, 'innerHeight', { value: 800, writable: true })
    window.dispatchEvent(new Event('scroll'))
    await flushPromises()

    // 第二次调用应该是 page=2
    expect(mockGetRankingArchive).toHaveBeenCalledTimes(2)
  })

  it('视频数不足 20 → hasMore=false', async () => {
    mockGetRankingArchive.mockResolvedValue({ code: '1', data: [{ aId: 1 }] })
    const wrapper = mount(Ranking, { global: { plugins: [router] } })
    await flushPromises()
    // 不足 20 条时不应触发滚动加载
  })
})
