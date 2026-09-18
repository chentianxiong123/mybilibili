import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import { getCategoryList, getCategoryById, addCategory, updateCategory, deleteCategory } from './category'
import request from '@/api/client'

const requestMock = request as unknown as ReturnType<typeof vi.fn> & {
  get: ReturnType<typeof vi.fn>
  post: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  requestMock.mockReset()
  requestMock.get.mockReset()
  requestMock.post.mockReset()
  requestMock.put.mockReset()
  requestMock.delete.mockReset()
})

describe('category api', () => {
  it('getCategoryList GET /category 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getCategoryList({ page: 1 })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/category',
      method: 'get',
      params: { page: 1 },
    })
  })

  it('getCategoryById GET /category/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 5, name: '科技' } })
    const res = await getCategoryById(5)
    expect(requestMock).toHaveBeenCalledWith({ url: '/category/5', method: 'get' })
    expect(res.data.name).toBe('科技')
  })

  it('addCategory POST /category 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 8 } })
    const data = { name: '新分类' }
    await addCategory(data)
    expect(requestMock).toHaveBeenCalledWith({ url: '/category', method: 'post', data })
  })

  it('updateCategory PUT /category/:id 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateCategory(5, { name: '改名' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/category/5',
      method: 'put',
      data: { name: '改名' },
    })
  })

  it('deleteCategory DELETE /category/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteCategory(9)
    expect(requestMock).toHaveBeenCalledWith({ url: '/category/9', method: 'delete' })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('500'))
    await expect(getCategoryList()).rejects.toThrow('500')
  })
})
