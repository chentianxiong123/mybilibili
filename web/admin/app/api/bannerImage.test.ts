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
  getHomeBanners, addHomeBanner, updateHomeBanner, deleteHomeBanner,
  getCategoryBanners, addCategoryBanner, updateCategoryBanner, deleteCategoryBanner,
  getBackgroundImage, saveBackgroundImage, deleteBackgroundImage,
  getUserProfileBackground, saveUserProfileBackground, deleteUserProfileBackground,
  uploadBannerImage
} from './bannerImage'
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

describe('bannerImage api - 首页轮播', () => {
  it('getHomeBanners GET /banner-images/home', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getHomeBanners()
    expect(requestMock.get).toHaveBeenCalledWith('/banner-images/home')
  })

  it('addHomeBanner POST 透传 data', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: { id: 1 } })
    const data = { imageUrl: 'a.jpg', link: 'http://x' }
    await addHomeBanner(data)
    expect(requestMock.post).toHaveBeenCalledWith('/banner-images/home', data)
  })

  it('updateHomeBanner PUT /banner-images/home/:id', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await updateHomeBanner(7, { title: '新' })
    expect(requestMock.put).toHaveBeenCalledWith('/banner-images/home/7', { title: '新' })
  })

  it('deleteHomeBanner DELETE /banner-images/home/:id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteHomeBanner(8)
    expect(requestMock.delete).toHaveBeenCalledWith('/banner-images/home/8')
  })
})

describe('bannerImage api - 分类轮播', () => {
  it('getCategoryBanners GET /banner-images/category/:id', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: [] })
    await getCategoryBanners(5)
    expect(requestMock.get).toHaveBeenCalledWith('/banner-images/category/5')
  })

  it('addCategoryBanner POST 拼接 categoryId', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await addCategoryBanner(5, { url: 'x.jpg' })
    expect(requestMock.post).toHaveBeenCalledWith('/banner-images/category/5', { url: 'x.jpg' })
  })

  it('updateCategoryBanner PUT 拼接两个 id', async () => {
    requestMock.put.mockResolvedValue({ code: 200, data: null })
    await updateCategoryBanner(5, 6, { url: 'y.jpg' })
    expect(requestMock.put).toHaveBeenCalledWith('/banner-images/category/5/6', { url: 'y.jpg' })
  })

  it('deleteCategoryBanner DELETE 拼接两个 id', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteCategoryBanner(5, 6)
    expect(requestMock.delete).toHaveBeenCalledWith('/banner-images/category/5/6')
  })
})

describe('bannerImage api - 背景图', () => {
  it('getBackgroundImage GET /banner-images/background', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await getBackgroundImage()
    expect(requestMock.get).toHaveBeenCalledWith('/banner-images/background')
  })

  it('saveBackgroundImage POST', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await saveBackgroundImage({ url: 'b.jpg' })
    expect(requestMock.post).toHaveBeenCalledWith('/banner-images/background', { url: 'b.jpg' })
  })

  it('deleteBackgroundImage DELETE', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteBackgroundImage()
    expect(requestMock.delete).toHaveBeenCalledWith('/banner-images/background')
  })
})

describe('bannerImage api - 用户主页背景', () => {
  it('getUserProfileBackground GET /banner-images/user-profile', async () => {
    requestMock.get.mockResolvedValue({ code: 200, data: null })
    await getUserProfileBackground()
    expect(requestMock.get).toHaveBeenCalledWith('/banner-images/user-profile')
  })

  it('saveUserProfileBackground POST', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: null })
    await saveUserProfileBackground({ url: 'u.jpg' })
    expect(requestMock.post).toHaveBeenCalledWith('/banner-images/user-profile', { url: 'u.jpg' })
  })

  it('deleteUserProfileBackground DELETE', async () => {
    requestMock.delete.mockResolvedValue({ code: 200, data: null })
    await deleteUserProfileBackground()
    expect(requestMock.delete).toHaveBeenCalledWith('/banner-images/user-profile')
  })
})

describe('bannerImage api - 上传', () => {
  it('uploadBannerImage 使用 multipart/form-data', async () => {
    requestMock.post.mockResolvedValue({ code: 200, data: { url: 'x' } })
    const file = new File(['x'], 'a.jpg', { type: 'image/jpeg' })
    await uploadBannerImage(file)
    expect(requestMock.post).toHaveBeenCalledWith(
      '/banner-images/upload',
      expect.any(FormData),
      expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } }),
    )
  })

  it('错误响应透传', async () => {
    requestMock.get.mockRejectedValue(new Error('403'))
    await expect(getHomeBanners()).rejects.toThrow('403')
  })
})
