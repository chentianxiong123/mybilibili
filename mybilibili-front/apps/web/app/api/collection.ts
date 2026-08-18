import api from './client'

export const collectionApi = {
  getUserCollections: (userId: number, page = 1, size = 20) => {
    return api.get(`/collection/user/${userId}?page=${page}&size=${size}`)
  },

  getCollectionById: (collectionId: number) => {
    return api.get(`/collection/${collectionId}`)
  },

  createCollection: (collectionData: any) => {
    const formData = new FormData()
    formData.append('name', collectionData.name)
    if (collectionData.description) {
      formData.append('description', collectionData.description)
    }
    formData.append('isPublic', collectionData.isPublic === false ? 'false' : 'true')
    if (collectionData.manuscriptIds && collectionData.manuscriptIds.length > 0) {
      const plainArray = Array.from(collectionData.manuscriptIds)
      formData.append('manuscriptIds', JSON.stringify(plainArray))
    }
    return api.post('/collection', formData)
  },

  updateCollection: (collectionId: number, collectionData: any) => {
    const formData = new FormData()
    if (collectionData.name) {
      formData.append('name', collectionData.name)
    }
    if (collectionData.description !== undefined) {
      formData.append('description', collectionData.description)
    }
    if (collectionData.cover) {
      formData.append('cover', collectionData.cover)
    }
    if (collectionData.isPublic !== undefined) {
      formData.append('isPublic', collectionData.isPublic === false ? 'false' : 'true')
    }
    return api.put(`/collection/${collectionId}`, formData)
  },

  deleteCollection: (collectionId: number) => {
    return api.delete(`/collection/${collectionId}`)
  },

  addManuscriptToCollection: (collectionId: number, manuscriptId: number, sortOrder: number) => {
    return api.post(`/collection/${collectionId}/manuscript/${manuscriptId}?order=${sortOrder}`)
  },

  removeManuscriptFromCollection: (collectionId: number, manuscriptId: number) => {
    return api.delete(`/collection/${collectionId}/manuscript/${manuscriptId}`)
  },

  updateManuscriptOrder: (collectionId: number, manuscriptOrders: any) => {
    return api.put(`/collection/${collectionId}/manuscripts/order`, { manuscriptOrders })
  },

  getCollectionManuscripts: (collectionId: number, page = 1, size = 20) => {
    return api.get(`/collection/${collectionId}/manuscripts?page=${page}&size=${size}`)
  },

  getMyCollections: () => {
    return api.get('/collection/my')
  }
}

export default collectionApi