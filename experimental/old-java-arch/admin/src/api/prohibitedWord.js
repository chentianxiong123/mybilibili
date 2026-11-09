import request from './request'

// 获取违禁词列表
export const getProhibitedWordList = (params) => {
  return request({
    url: '/admin/prohibited-words',
    method: 'get',
    params
  })
}

// 获取违禁词详情
export const getProhibitedWordById = (id) => {
  return request({
    url: `/admin/prohibited-words/${id}`,
    method: 'get'
  })
}

// 添加违禁词
export const addProhibitedWord = (data) => {
  return request({
    url: '/admin/prohibited-words',
    method: 'post',
    data
  })
}

// 更新违禁词
export const updateProhibitedWord = (id, data) => {
  return request({
    url: `/admin/prohibited-words/${id}`,
    method: 'put',
    data
  })
}

// 删除违禁词
export const deleteProhibitedWord = (id) => {
  return request({
    url: `/admin/prohibited-words/${id}`,
    method: 'delete'
  })
}

// 批量导入违禁词
export const batchImportProhibitedWords = (formData) => {
  return request({
    url: '/admin/prohibited-words/batch-import',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
