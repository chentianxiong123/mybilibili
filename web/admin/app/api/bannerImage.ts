import request from './client'

// ?history=1 时返回含已归档（status=-1）的历史记录
const getWithHistory = (url, history) =>
  history ? request.get(url, { params: { history: 1 } }) : request.get(url)

// 首页轮播图
export const getHomeBanners = (history) => {
  return getWithHistory('/banner-images/home', history)
}

export const addHomeBanner = (data) => {
  return request.post('/banner-images/home', data)
}

export const updateHomeBanner = (id, data) => {
  return request.put(`/banner-images/home/${id}`, data)
}

export const deleteHomeBanner = (id) => {
  return request.delete(`/banner-images/home/${id}`)
}

// 分类轮播图
export const getCategoryBanners = (categoryId, history) => {
  return getWithHistory(`/banner-images/category/${categoryId}`, history)
}

export const addCategoryBanner = (categoryId, data) => {
  return request.post(`/banner-images/category/${categoryId}`, data)
}

export const updateCategoryBanner = (categoryId, id, data) => {
  return request.put(`/banner-images/category/${categoryId}/${id}`, data)
}

export const deleteCategoryBanner = (categoryId, id) => {
  return request.delete(`/banner-images/category/${categoryId}/${id}`)
}

// 顶部背景图
export const getBackgroundImage = (history) => {
  return getWithHistory('/banner-images/background', history)
}

export const saveBackgroundImage = (data) => {
  return request.post('/banner-images/background', data)
}

export const deleteBackgroundImage = () => {
  return request.delete('/banner-images/background')
}

// 用户主页背景图
export const getUserProfileBackground = (history) => {
  return getWithHistory('/banner-images/user-profile', history)
}

export const saveUserProfileBackground = (data) => {
  return request.post('/banner-images/user-profile', data)
}

export const deleteUserProfileBackground = (id?: number) => {
  if (id) {
    return request.delete(`/banner-images/user-profile/${id}`)
  }
  return request.delete('/banner-images/user-profile')
}

// 图片上传
export const uploadBannerImage = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/banner-images/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
