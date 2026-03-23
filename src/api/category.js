import request from './request'

// 获取分类列表
export const getCategoryList = (params) => {
  return request({
    url: '/categories',
    method: 'get',
    params
  })
}

// 获取分类详情
export const getCategoryById = (id) => {
  return request({
    url: `/categories/${id}`,
    method: 'get'
  })
}

// 添加分类
export const addCategory = (data) => {
  return request({
    url: '/categories',
    method: 'post',
    params: data
  })
}

// 更新分类
export const updateCategory = (id, data) => {
  return request({
    url: `/categories/${id}`,
    method: 'put',
    params: data
  })
}

// 删除分类
export const deleteCategory = (id) => {
  return request({
    url: `/categories/${id}`,
    method: 'delete'
  })
}