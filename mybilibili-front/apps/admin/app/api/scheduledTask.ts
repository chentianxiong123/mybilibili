import request from './client'

export const getScheduledTasks = () => {
  return request({
    url: '/admin/scheduled-tasks',
    method: 'get'
  })
}

export const createScheduledTask = (data) => {
  return request({
    url: '/admin/scheduled-tasks',
    method: 'post',
    data
  })
}

export const updateScheduledTask = (data) => {
  return request({
    url: '/admin/scheduled-tasks',
    method: 'put',
    data
  })
}

export const toggleScheduledTask = (id, enabled) => {
  return request({
    url: '/admin/scheduled-tasks/toggle',
    method: 'post',
    data: { id, enabled }
  })
}

export const triggerScheduledTask = (taskKey) => {
  return request({
    url: '/admin/scheduled-tasks/trigger',
    method: 'post',
    data: { task_key: taskKey }
  })
}

export const deleteScheduledTask = (id) => {
  return request({
    url: '/admin/scheduled-tasks',
    method: 'delete',
    data: { id }
  })
}