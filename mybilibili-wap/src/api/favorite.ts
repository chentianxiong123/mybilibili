import api from './client'

export async function getCollectedVideos() {
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
