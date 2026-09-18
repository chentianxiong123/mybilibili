import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  apiGet: vi.fn(),
  apiPost: vi.fn(),
  apiPut: vi.fn(),
  apiDelete: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: Object.assign(vi.fn(), {
    get: mocks.apiGet,
    post: mocks.apiPost,
    put: mocks.apiPut,
    delete: mocks.apiDelete,
  }),
}))

import { collectionApi } from './collection'

const mockGet = mocks.apiGet
const mockPost = mocks.apiPost
const mockPut = mocks.apiPut
const mockDelete = mocks.apiDelete

describe('collection api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getUserCollections 拼接分页参数', async () => {
    mockGet.mockResolvedValueOnce({ code: 200, data: [] })
    await collectionApi.getUserCollections(10, 2, 30)
    expect(mockGet).toHaveBeenCalledWith('/collection/user/10?page=2&size=30')
  })

  it('getCollectionById', async () => {
    mockGet.mockResolvedValueOnce({ code: 200, data: { id: 5 } })
    await collectionApi.getCollectionById(5)
    expect(mockGet).toHaveBeenCalledWith('/collection/5')
  })

  it('createCollection 把字段塞进 FormData', async () => {
    mockPost.mockResolvedValueOnce({ code: 200, data: { id: 1 } })
    await collectionApi.createCollection({
      name: 'name',
      description: 'desc',
      isPublic: true,
      manuscriptIds: [1, 2, 3],
    })
    expect(mockPost).toHaveBeenCalled()
    const [url, form] = mockPost.mock.calls[0]
    expect(url).toBe('/collection')
    expect(form.get('name')).toBe('name')
    expect(form.get('description')).toBe('desc')
    expect(form.get('isPublic')).toBe('true')
    expect(form.get('manuscriptIds')).toBe(JSON.stringify([1, 2, 3]))
  })

  it('createCollection isPublic=false 转为字符串 false', async () => {
    mockPost.mockResolvedValueOnce({ code: 200, data: { id: 2 } })
    await collectionApi.createCollection({ name: 'p', isPublic: false })
    const [, form] = mockPost.mock.calls[0]
    expect(form.get('isPublic')).toBe('false')
  })

  it('updateCollection 包含 cover 时追加', async () => {
    mockPut.mockResolvedValueOnce({ code: 200 })
    const file = new File(['x'], 'c.jpg')
    await collectionApi.updateCollection(7, { name: 'new', cover: file })
    expect(mockPut).toHaveBeenCalled()
    const [url, form] = mockPut.mock.calls[0]
    expect(url).toBe('/collection/7')
    expect(form.get('name')).toBe('new')
    expect(form.get('cover')).toBe(file)
  })

  it('deleteCollection', async () => {
    mockDelete.mockResolvedValueOnce({ code: 200 })
    await collectionApi.deleteCollection(8)
    expect(mockDelete).toHaveBeenCalledWith('/collection/8')
  })

  it('addManuscriptToCollection 携带 order', async () => {
    mockPost.mockResolvedValueOnce({ code: 200 })
    await collectionApi.addManuscriptToCollection(1, 2, 3)
    expect(mockPost).toHaveBeenCalledWith('/collection/1/manuscript/2?order=3')
  })

  it('removeManuscriptFromCollection', async () => {
    mockDelete.mockResolvedValueOnce({ code: 200 })
    await collectionApi.removeManuscriptFromCollection(1, 2)
    expect(mockDelete).toHaveBeenCalledWith('/collection/1/manuscript/2')
  })

  it('updateManuscriptOrder 传 manuscriptOrders body', async () => {
    mockPut.mockResolvedValueOnce({ code: 200 })
    const orders = [{ manuscriptId: 1, order: 0 }]
    await collectionApi.updateManuscriptOrder(1, orders)
    expect(mockPut).toHaveBeenCalledWith('/collection/1/manuscripts/order', { manuscriptOrders: orders })
  })

  it('getCollectionManuscripts', async () => {
    mockGet.mockResolvedValueOnce({ code: 200, data: [] })
    await collectionApi.getCollectionManuscripts(9)
    expect(mockGet).toHaveBeenCalledWith('/collection/9/manuscripts?page=1&size=20')
  })

  it('getMyCollections', async () => {
    mockGet.mockResolvedValueOnce({ code: 200, data: [] })
    await collectionApi.getMyCollections()
    expect(mockGet).toHaveBeenCalledWith('/collection/my')
  })
})
