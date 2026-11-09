import request from './request'

export const getCategoryList = (params) => {
  return request({
    url: '/category',
    method: 'get',
    params
  })
}

export const getCategoryById = (id) => {
  return request({
    url: `/category/${id}`,
    method: 'get'
  })
}

export const addCategory = (data) => {
  return request({
    url: '/category',
    method: 'post',
    data
  })
}

export const updateCategory = (id, data) => {
  return request({
    url: `/category/${id}`,
    method: 'put',
    data
  })
}

export const deleteCategory = (id) => {
  return request({
    url: `/category/${id}`,
    method: 'delete'
  })
}
