import api from './client'

function hasToken() {
  return !!localStorage.getItem('token')
}

export async function getCollectedVideos() {
  if (!hasToken()) return { code: '0', data: [] }
  try {
    const res = await api.get('/manuscript/user/collections')
    return {
      code: '1',
      data: res?.data || res || []
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getFavoriteFolders() {
  if (!hasToken()) return { code: '0', data: [] }
  try {
    const res = await api.get('/manuscript/favorite/folders')
    return {
      code: '1',
      data: res?.data || res || []
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

export async function getFavoriteFolderVideos(folderId: number) {
  if (!hasToken()) return { code: '0', data: [] }
  try {
    const res = await api.get(`/manuscript/favorite/folders/${folderId}/videos`)
    return {
      code: '1',
      data: res?.data?.records || res?.data || res || []
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}
