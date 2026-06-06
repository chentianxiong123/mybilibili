import request from './request'

export const getCurrentTask = () => {
  return request.get('/video/process/admin/current')
}

export const getQueueInfo = () => {
  return request.get('/video/process/admin/queue')
}

export const getStatistics = () => {
  return request.get('/video/process/admin/statistics')
}
