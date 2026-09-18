import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/api/client', () => {
  const fn: any = vi.fn()
  fn.get = vi.fn()
  fn.post = vi.fn()
  fn.put = vi.fn()
  fn.delete = vi.fn()
  return { default: fn }
})

import {
  getProhibitedWordList, getProhibitedWordById, addProhibitedWord,
  updateProhibitedWord, deleteProhibitedWord, batchImportProhibitedWords
} from './prohibitedWord'
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

describe('prohibitedWord api', () => {
  it('getProhibitedWordList GET /moderation/admin/prohibited-words 带 params', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getProhibitedWordList({ page: 1, keyword: 'a' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words',
      method: 'get',
      params: { page: 1, keyword: 'a' },
    })
  })

  it('getProhibitedWordById GET /moderation/admin/prohibited-words/:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 7, word: '违禁' } })
    await getProhibitedWordById(7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words/7',
      method: 'get',
    })
  })

  it('addProhibitedWord POST 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    const data = { word: '违禁词', category: 'politics' }
    await addProhibitedWord(data)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words',
      method: 'post',
      data,
    })
  })

  it('updateProhibitedWord PUT 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateProhibitedWord(3, { word: '改' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words/3',
      method: 'put',
      data: { word: '改' },
    })
  })

  it('deleteProhibitedWord DELETE /:id', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteProhibitedWord(4)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words/4',
      method: 'delete',
    })
  })

  it('batchImportProhibitedWords POST multipart/form-data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { imported: 3 } })
    const fd = new FormData()
    fd.append('file', new File(['x'], 'a.txt'))
    await batchImportProhibitedWords(fd)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/moderation/admin/prohibited-words/batch-import',
      method: 'post',
      data: fd,
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  })
})
