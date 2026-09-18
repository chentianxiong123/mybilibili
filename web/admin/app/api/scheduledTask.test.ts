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
  getScheduledTasks, createScheduledTask, updateScheduledTask,
  toggleScheduledTask, triggerScheduledTask, deleteScheduledTask
} from './scheduledTask'
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

describe('scheduledTask api', () => {
  it('getScheduledTasks GET /admin/scheduled-tasks', async () => {
    requestMock.mockResolvedValue({ code: 200, data: [] })
    await getScheduledTasks()
    expect(requestMock).toHaveBeenCalledWith({ url: '/admin/scheduled-tasks', method: 'get' })
  })

  it('createScheduledTask POST 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: { id: 1 } })
    await createScheduledTask({ name: 't', cron: '* * * * *' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/scheduled-tasks',
      method: 'post',
      data: { name: 't', cron: '* * * * *' },
    })
  })

  it('updateScheduledTask PUT 透传 data', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await updateScheduledTask({ id: 1, name: 't2' })
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/scheduled-tasks',
      method: 'put',
      data: { id: 1, name: 't2' },
    })
  })

  it('toggleScheduledTask POST 带 {id, enabled}', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await toggleScheduledTask(5, true)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/scheduled-tasks/toggle',
      method: 'post',
      data: { id: 5, enabled: true },
    })
  })

  it('triggerScheduledTask POST 带 task_key', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await triggerScheduledTask('clean_logs')
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/scheduled-tasks/trigger',
      method: 'post',
      data: { task_key: 'clean_logs' },
    })
  })

  it('deleteScheduledTask DELETE 带 data: {id}', async () => {
    requestMock.mockResolvedValue({ code: 200, data: null })
    await deleteScheduledTask(7)
    expect(requestMock).toHaveBeenCalledWith({
      url: '/admin/scheduled-tasks',
      method: 'delete',
      data: { id: 7 },
    })
  })

  it('错误响应透传', async () => {
    requestMock.mockRejectedValue(new Error('boom'))
    await expect(getScheduledTasks()).rejects.toThrow('boom')
  })
})
