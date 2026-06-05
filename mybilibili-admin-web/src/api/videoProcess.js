import request from './request'

export const getCurrentTask = () => {
  return request.get('/ai/admin/process/current')
}

export const getQueueInfo = () => {
  return request.get('/ai/admin/process/queue')
}

export const getStatistics = () => {
  return request.get('/ai/admin/process/statistics')
}
