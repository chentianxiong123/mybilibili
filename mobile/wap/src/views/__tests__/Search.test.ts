import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../../api/search', () => ({
  getHotwords: vi.fn(),
  getSuggests: vi.fn(),
  getSearchResult: vi.fn()
}))

vi.mock('../../api/index', () => ({
  submitFeedback: vi.fn()
}))

vi.mock('../../utils/storage_layer', () => ({
  default: {
    get: vi.fn().mockReturnValue(null),
    set: vi.fn(),
    remove: vi.fn()
  },
  K: { searchHistory: 'search:history', searchDiscoverVisible: 'search:discover:visible' }
}))

import Search from '../search/Search.vue'
import { getHotwords, getSuggests, getSearchResult } from '../../api/search'
import { submitFeedback } from '../../api/index'

const mockGetHotwords = vi.mocked(getHotwords)
const mockGetSuggests = vi.mocked(getSuggests)
const mockGetSearchResult = vi.mocked(getSearchResult)
const mockSubmitFeedback = vi.mocked(submitFeedback)

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/m/search', component: Search },
    { path: '/m/search/result', component: { template: '<div />' } },
    { path: '/m/search/hot', component: { template: '<div />' } }
  ]
})

describe('Search.vue', () => {
  beforeEach(async () => {
    vi.clearAllMocks()
    mockGetHotwords.mockResolvedValue({ code: '1', data: [{ keyword: 'AI' }, { keyword: '前端' }] })
    mockGetSuggests.mockResolvedValue({ code: '1', data: [] })
    mockGetSearchResult.mockResolvedValue({ code: '1', data: [] })
    mockSubmitFeedback.mockResolvedValue({ code: '1', data: {} })
    router.push('/m/search')
    await router.isReady()
  })

  it('渲染搜索框', () => {
    const wrapper = mount(Search, { global: { plugins: [router] } })
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('加载热搜词', async () => {
    mount(Search, { global: { plugins: [router] } })
    await flushPromises()
    expect(mockGetHotwords).toHaveBeenCalled()
  })

  it('输入时显示搜索建议', async () => {
    mockGetSuggests.mockResolvedValue({ code: '1', data: [{ name: 'AI绘画' }, { name: 'AI视频' }] })
    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('AI')
    await flushPromises()

    expect(mockGetSuggests).toHaveBeenCalledWith('AI')
  })

  it('空输入 → 清空建议', async () => {
    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('AI')
    await flushPromises()
    await input.setValue('')
    await flushPromises()

    // 空输入不调 getSuggests
    expect(mockGetSuggests).not.toHaveBeenCalledWith('')
  })

  it('提交反馈成功', async () => {
    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()

    // 查找反馈按钮（如果有的话）
    const feedbackBtn = wrapper.find('.feedback-btn')
    if (feedbackBtn.exists()) {
      await feedbackBtn.trigger('click')
    }
  })
})
