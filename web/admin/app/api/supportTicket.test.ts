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
  getTicketList, getTicketById, processTicket, deleteTicket
} from './supportTicket'
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

describe('supportTicket api', () => {
  it('getTicketList GET /support/admin/tickets 带 params', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getTicketList({ page: 1, status: 'open' })
    expect(requestMock.get).toHaveBeenCalledWith('/support/admin/tickets', { params: { page: 1, status: 'open' } })
  })

  it('getTicketById GET /support/admin/tickets/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: { id: 3 } })
    await getTicketById(3)
    expect(requestMock.get).toHaveBeenCalledWith('/support/admin/tickets/3')
  })

  it('processTicket PUT /support/admin/tickets/:id/process 带 adminReply', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await processTicket(3, '已处理')
    expect(requestMock.put).toHaveBeenCalledWith('/support/admin/tickets/3/process', { adminReply: '已处理' })
  })

  it('deleteTicket DELETE /support/admin/tickets/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteTicket(5)
    expect(requestMock.delete).toHaveBeenCalledWith('/support/admin/tickets/5')
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('500'))
    await expect(getTicketList({})).rejects.toThrow('500')
  })
})
