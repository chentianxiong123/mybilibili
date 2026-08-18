import api from './client'

export default {
  getComments(dynamicId: number, page = 1, size = 10) {
    return api({
      url: '/dynamic/comment/list',
      method: 'get',
      params: { dynamicId, page, size }
    })
  },

  addComment(dynamicId: number, content: string, parentId: number | null = null, replyUserId: number | null = null) {
    return api({
      url: '/dynamic/comment/add',
      method: 'post',
      params: { dynamicId, content, parentId, replyUserId }
    })
  },

  deleteComment(commentId: number) {
    return api({
      url: `/dynamic/comment/delete/${commentId}`,
      method: 'delete'
    })
  },

  getReplies(commentId: number) {
    return api({
      url: '/dynamic/comment/replies',
      method: 'get',
      params: { commentId }
    })
  },

  likeComment(commentId: number) {
    return api({
      url: `/dynamic/comment/like/${commentId}`,
      method: 'post'
    })
  },

  unlikeComment(commentId: number) {
    return api({
      url: `/dynamic/comment/like/${commentId}`,
      method: 'delete'
    })
  }
}