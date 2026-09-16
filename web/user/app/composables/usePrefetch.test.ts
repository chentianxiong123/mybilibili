import { describe, it, expect, vi } from 'vitest'

// 通过 vi.mock 拦截 vue-router 的 useRouter
const fakeRouter = {
  resolve: vi.fn(),
}

vi.mock('vue-router', () => ({
  useRouter: () => fakeRouter,
}))

import { usePrefetch } from './usePrefetch'

describe('usePrefetch', () => {
  beforeEach(() => {
    fakeRouter.resolve.mockReset()
  })

  it('返回 prefetch 函数', () => {
    const r = usePrefetch()
    expect(typeof r.prefetch).toBe('function')
  })

  it('string target 触发 router.resolve(target)', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    r.prefetch('/foo')
    expect(fakeRouter.resolve).toHaveBeenCalledWith('/foo')
  })

  it('相同 string target 第二次调用被跳过', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    r.prefetch('/foo')
    r.prefetch('/foo')
    expect(fakeRouter.resolve).toHaveBeenCalledTimes(1)
  })

  it('不同 string target 触发多次 resolve', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    r.prefetch('/foo')
    r.prefetch('/bar')
    expect(fakeRouter.resolve).toHaveBeenCalledTimes(2)
  })

  it('object target 用 name + params 作 key', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    r.prefetch({ name: 'user', params: { id: 1 } })
    r.prefetch({ name: 'user', params: { id: 1 } })
    expect(fakeRouter.resolve).toHaveBeenCalledTimes(1)
    r.prefetch({ name: 'user', params: { id: 2 } })
    expect(fakeRouter.resolve).toHaveBeenCalledTimes(2)
  })

  it('matched 中 components.default 为函数时调用它', async () => {
    const compFn = vi.fn().mockResolvedValue({})
    fakeRouter.resolve.mockReturnValue({
      matched: [{ components: { default: compFn } }]
    })
    const r = usePrefetch()
    r.prefetch('/foo')
    await Promise.resolve()
    await Promise.resolve()
    expect(compFn).toHaveBeenCalledTimes(1)
  })

  it('matched 中 components.default 是对象时不调用', () => {
    const obj = {}
    fakeRouter.resolve.mockReturnValue({
      matched: [{ components: { default: obj } }]
    })
    const r = usePrefetch()
    expect(() => r.prefetch('/foo')).not.toThrow()
    expect(fakeRouter.resolve).toHaveBeenCalled()
  })

  it('组件加载失败被 .catch 吞掉，不抛错', async () => {
    const failingComp = vi.fn().mockRejectedValue(new Error('chunk fail'))
    fakeRouter.resolve.mockReturnValue({
      matched: [{ components: { default: failingComp } }]
    })
    const r = usePrefetch()
    expect(() => r.prefetch('/foo')).not.toThrow()
    await Promise.resolve()
    await Promise.resolve()
    expect(failingComp).toHaveBeenCalled()
  })

  it('matched 为空数组时不调用任何组件', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    expect(() => r.prefetch('/foo')).not.toThrow()
  })

  it('unknown target 类型不做任何事', () => {
    fakeRouter.resolve.mockReturnValue({ matched: [] })
    const r = usePrefetch()
    // @ts-expect-error
    r.prefetch(123)
    expect(fakeRouter.resolve).not.toHaveBeenCalled()
  })
})