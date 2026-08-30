import request from './request'

// 首页轮播图
export const getHomeBanners = () => {
  return request.get('/banner-images/home')
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
export const getCategoryBanners = (categoryId) => {
  return request.get(`/banner-images/category/${categoryId}`)
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
export const getBackgroundImage = () => {
  return request.get('/banner-images/background')
}

export const saveBackgroundImage = (data) => {
  return request.post('/banner-images/background', data)
}

export const deleteBackgroundImage = () => {
  return request.delete('/banner-images/background')
}

// 用户主页背景图
export const getUserProfileBackground = () => {
  return request.get('/banner-images/user-profile')
}

export const saveUserProfileBackground = (data) => {
  return request.post('/banner-images/user-profile', data)
}

export const deleteUserProfileBackground = () => {
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
