<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import { commentApi, userApi } from '@/api/client'
import LevelBadge from '@/components/LevelBadge.vue'
import UserFloatCard from '@/components/UserFloatCard.vue'
import { CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  manuscriptId: {
    type: Number,
    required: true
  },
  commentCount: {
    type: Number,
    default: 0
  },
  uploaderId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['update:commentCount'])

const currentUser = ref(null)
onMounted(() => {
  const u = safeStorage.getItem('user')
  if (u) currentUser.value = JSON.parse(u)
})

const comments = ref([])
const loadingComments = ref(false)
const replyInputs = ref({})
const replyTargets = ref({})
const replyExpanded = ref({})
const replyPage = ref({})
const loadingReplies = ref({})
const commentSort = ref('hot')
const REPLY_PAGE_SIZE = 7
const newComment = ref('')
const commentInputWrapper = ref(null)
const isCommentInputCollapsed = ref(true)
const showEmojiPicker = ref(false)
const showReplyEmojiPicker = ref({})
const activeReplyEmojiCommentId = ref(null)

const currentUserAvatar = () => {
  const user = JSON.parse(safeStorage.getItem('user') || 'null')
  return user?.avatar || '/default-avatar.svg'
}

const toggleCommentInput = () => {
  isCommentInputCollapsed.value = !isCommentInputCollapsed.value
}

const handleClickOutside = event => {
  if (commentInputWrapper.value && !commentInputWrapper.value.contains(event.target)) {
    isCommentInputCollapsed.value = true
    showEmojiPicker.value = false
  }
}

const formatDate = dateStr => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const formatContentWithAtLinks = content => {
  if (!content) return ''
  const atRegex = /@([\u4e00-\u9fa5a-zA-Z0-9_]+)/g
  return content.replace(
    atRegex,
    '<a href="javascript:void(0)" class="user-link" onclick="window.open(\'/profile/\' + this.textContent.substring(1) + \'/home\', \'_blank\')">@$1</a>'
  )
}

const goToAuthor = authorId => {
  window.open(`/profile/${authorId}/home`, '_blank')
}

const emojiList = [
  '😀',
  '😃',
  '😄',
  '😁',
  '😆',
  '😅',
  '😂',
  '🤣',
  '😊',
  '😇',
  '🙂',
  '🙃',
  '😉',
  '😌',
  '😍',
  '🥰',
  '😘',
  '😗',
  '😙',
  '😚',
  '😋',
  '😛',
  '😜',
  '😝',
  '🤪',
  '🤨',
  '🧐',
  '🤓',
  '😎',
  '🤩',
  '🥳',
  '😏',
  '😒',
  '😞',
  '😔',
  '😟',
  '😕',
  '🙁',
  '☹️',
  '😣',
  '😖',
  '😫',
  '😩',
  '🥺',
  '😢',
  '😭',
  '😤',
  '😠',
  '😡',
  '🤬',
  '👍',
  '👎',
  '👏',
  '🙌',
  '🤝',
  '💪',
  '❤️',
  '🧡',
  '💛',
  '💚',
  '💙',
  '💜',
  '🖤',
  '🤍'
]

const selectEmoji = emoji => {
  newComment.value += emoji
}

const toggleReplyEmojiPicker = commentId => {
  showEmojiPicker.value = false
  if (activeReplyEmojiCommentId.value === commentId && showReplyEmojiPicker.value[commentId]) {
    showReplyEmojiPicker.value[commentId] = false
    activeReplyEmojiCommentId.value = null
  } else {
    showReplyEmojiPicker.value = {}
    showReplyEmojiPicker.value[commentId] = true
    activeReplyEmojiCommentId.value = commentId
  }
}

const selectReplyEmoji = (emoji, commentId) => {
  if (replyInputs.value[commentId] !== undefined) {
    replyInputs.value[commentId] += emoji
  }
  showReplyEmojiPicker.value[commentId] = false
  activeReplyEmojiCommentId.value = null
}

const isUploaderUser = userId => {
  const uploaderId = props.uploaderId
  return uploaderId !== undefined && uploaderId !== null && uploaderId !== '' && String(userId) === String(uploaderId)
}

const getUploaderReplies = comment => {
  return (comment.replies || []).filter(reply => isUploaderUser(reply.userId))
}

const getOtherReplies = comment => {
  return (comment.replies || []).filter(reply => !isUploaderUser(reply.userId))
}

const getReplyTotal = comment => {
  return Number(comment.replyCount ?? comment.replies?.length ?? 0)
}

const getReplyPageCount = comment => {
  return Math.max(1, Math.ceil(getReplyTotal(comment) / REPLY_PAGE_SIZE))
}

const shouldShowReplyExpand = comment => {
  return getReplyTotal(comment) > getUploaderReplies(comment).length
}

const submitComment = async () => {
  if (!newComment.value.trim()) {
    ElMessage.warning('评论内容不能为空')
    return
  }

  try {
    const response = await commentApi.postComment('VIDEO', props.manuscriptId, newComment.value)
    if (response.code === 200) {
      comments.value.unshift({
        id: response.data.id,
        userId: response.data.userId,
        author: response.data.userName,
        avatar: response.data.userAvatar || currentUserAvatar(),
        userLevel: response.data.userLevel || currentUser.value?.level || 0,
        content: response.data.content,
        time: formatDate(response.data.createTime || new Date()),
        likeCount: response.data.likeCount || 0,
        isLiked: response.data.liked || false,
        dislikeCount: 0,
        isDisliked: false,
        replyCount: response.data.replyCount || 0,
        replies: []
      })
      emit('update:commentCount', props.commentCount + 1)
      newComment.value = ''
      isCommentInputCollapsed.value = true
      showEmojiPicker.value = false
      ElMessage.success('评论发表成功，经验值+5')
    } else {
      ElMessage.error(response.message || '评论发表失败')
    }
  } catch (error) {
    console.error('评论发表失败:', error)
    ElMessage.error('评论发表失败，请稍后重试')
  }
}

const likeComment = async commentId => {
  const comment = comments.value.find(c => c.id === commentId)
  if (!comment) return

  try {
    if (comment.isLiked) {
      const response = await commentApi.unlikeComment(commentId)
      if (response.code === 200) {
        comment.likeCount = Math.max(0, comment.likeCount - 1)
        comment.isLiked = false
      } else {
        ElMessage.error(response.message || '取消点赞失败')
      }
    } else {
      const response = await commentApi.likeComment(commentId)
      if (response.code === 200) {
        comment.likeCount++
        comment.isLiked = true
      } else {
        ElMessage.error(response.message || '点赞失败')
      }
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  }
}

const dislikeComment = commentId => {
  if (!currentUser.value) {
    ElMessage.warning('请先登录')
    return
  }
  const comment = comments.value.find(c => c.id === commentId)
  if (!comment) return
  if (comment.isDisliked) {
    comment.dislikeCount = Math.max(0, (comment.dislikeCount || 0) - 1)
    comment.isDisliked = false
  } else {
    comment.dislikeCount = (comment.dislikeCount || 0) + 1
    comment.isDisliked = true
    if (comment.isLiked) {
      comment.isLiked = false
      comment.likeCount = Math.max(0, (comment.likeCount || 0) - 1)
    }
  }
}

const likeReply = async replyId => {
  try {
    let targetReply = null
    for (const comment of comments.value) {
      if (comment.replies) {
        const reply = comment.replies.find(r => r.id === replyId)
        if (reply) {
          targetReply = reply
          break
        }
      }
    }
    if (!targetReply) return

    if (targetReply.isLiked) {
      const response = await commentApi.unlikeReply(replyId)
      if (response.code === 200) {
        targetReply.likeCount = Math.max(0, targetReply.likeCount - 1)
        targetReply.isLiked = false
      } else {
        ElMessage.error(response.message || '取消点赞失败')
      }
    } else {
      const response = await commentApi.likeReply(replyId)
      if (response.code === 200) {
        targetReply.likeCount++
        targetReply.isLiked = true
      } else {
        ElMessage.error(response.message || '点赞失败')
      }
    }
  } catch (error) {
    console.error('点赞回复失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  }
}

const dislikeReply = replyId => {
  if (!currentUser.value) {
    ElMessage.warning('请先登录')
    return
  }
  for (const comment of comments.value) {
    if (comment.replies) {
      const reply = comment.replies.find(r => r.id === replyId)
      if (reply) {
        if (reply.isDisliked) {
          reply.dislikeCount = Math.max(0, (reply.dislikeCount || 0) - 1)
          reply.isDisliked = false
        } else {
          reply.dislikeCount = (reply.dislikeCount || 0) + 1
          reply.isDisliked = true
          if (reply.isLiked) {
            reply.isLiked = false
            reply.likeCount = Math.max(0, (reply.likeCount || 0) - 1)
          }
        }
        return
      }
    }
  }
}

const loadReplies = async (commentId, page = 1) => {
  try {
    loadingReplies.value[commentId] = true
    const response = await commentApi.getRepliesByCommentId(commentId, page, REPLY_PAGE_SIZE)
    if (response.code === 200) {
      const comment = comments.value.find(c => c.id === commentId)
      if (comment) {
        const formattedReplies = response.data.map(reply => ({
          id: reply.id,
          userId: reply.userId || reply.user?.id,
          author: reply.userName || reply.author || '未知用户',
          avatar: reply.userAvatar || reply.avatar || '/default-avatar.svg',
          userLevel: reply.userLevel || 0,
          content: reply.content,
          time: reply.time || formatDate(reply.createTime),
          likeCount: reply.likeCount || 0,
          dislikeCount: reply.dislikeCount || 0,
          isLiked: reply.liked || false,
          targetAuthor: reply.replyToUserName || reply.targetAuthor || reply.replyTo || reply.toUserName || null
        }))
        comment.replies = formattedReplies
        replyPage.value[commentId] = page
      }
    }
  } catch (error) {
    console.error('获取回复失败:', error)
    ElMessage.error('获取回复失败，请稍后重试')
  } finally {
    loadingReplies.value[commentId] = false
  }
}

const toggleReplyExpanded = commentId => {
  replyExpanded.value[commentId] = !replyExpanded.value[commentId]
  if (replyExpanded.value[commentId]) {
    loadReplies(commentId, 1)
  }
}

const replyComment = commentId => {
  const comment = comments.value.find(c => c.id === commentId)
  if (comment) {
    comment.showReplyInput = !comment.showReplyInput
    replyInputs.value[commentId] = ''
    replyTargets.value[commentId] = null
  }
}

const replyToReply = (commentId, replyId, targetAuthor) => {
  const comment = comments.value.find(c => c.id === commentId)
  if (comment) {
    comment.showReplyInput = true
    replyInputs.value[commentId] = ''
    replyTargets.value[commentId] = {
      replyId,
      targetAuthor
    }
  }
}

const submitReply = async commentId => {
  const replyContent = replyInputs.value[commentId]
  if (!replyContent || !replyContent.trim()) {
    ElMessage.warning('回复内容不能为空')
    return
  }

  try {
    const targetAuthor = replyTargets.value[commentId] ? replyTargets.value[commentId].targetAuthor : null
    let replyToUserId = null
    if (targetAuthor) {
      const comment = comments.value.find(c => c.id === commentId)
      if (comment && comment.author === targetAuthor) {
        replyToUserId = comment.userId
      } else if (comment && comment.replies) {
        const reply = comment.replies.find(r => r.author === targetAuthor)
        if (reply) {
          replyToUserId = reply.userId
        }
      }
    }

    const response = await commentApi.replyComment(commentId, replyContent, replyToUserId)
    if (response.code === 200 && response.data) {
      replyInputs.value[commentId] = ''
      replyTargets.value[commentId] = null
      const comment = comments.value.find(c => c.id === commentId)
      if (comment) {
        comment.showReplyInput = false
        replyExpanded.value[commentId] = true
        if (!comment.replies) {
          comment.replies = []
        }
        const newReply = {
          id: response.data.id,
          userId: response.data.userId,
          author: response.data.userName || '未知用户',
          avatar: response.data.userAvatar || currentUserAvatar(),
          userLevel: response.data.userLevel || currentUser.value?.level || 0,
          content: response.data.content,
          time: formatDate(response.data.createTime),
          likeCount: response.data.likeCount || 0,
          dislikeCount: 0,
          isLiked: response.data.liked || false,
          targetAuthor: response.data.replyToUserName || targetAuthor
        }
        comment.replyCount = (comment.replyCount || 0) + 1
        if (isUploaderUser(newReply.userId)) {
          comment.replies.unshift(newReply)
        } else {
          comment.replies.push(newReply)
        }
      }
      ElMessage.success('回复成功，经验值+2')
    } else {
      ElMessage.error(response.message || (response.data ? '回复失败' : '回复失败：未获取到数据'))
    }
  } catch (error) {
    console.error('回复失败:', error)
    ElMessage.error('回复失败，请稍后重试')
  }
}

const loadComments = async (sort = 'new') => {
  if (!props.manuscriptId) return

  try {
    loadingComments.value = true
    const commentResponse = await commentApi.getComments('VIDEO', props.manuscriptId, 1, 20, sort)
    if (commentResponse.code === 200) {
      const commentData = commentResponse.data
      comments.value = commentData.map(comment => ({
        id: comment.id,
        userId: comment.userId || comment.user?.id,
        author: comment.userName || comment.author || comment.user?.name || '未知用户',
        avatar: comment.userAvatar || comment.avatar || comment.user?.avatar || '/default-avatar.svg',
        userLevel: comment.userLevel || 0,
        content: comment.content,
        time: comment.time || formatDate(comment.createTime),
        likeCount: comment.likeCount || 0,
        dislikeCount: comment.dislikeCount || 0,
        isLiked: comment.liked || false,
        replyCount: comment.replyCount || 0,
        showReplyInput: false,
        replies: (comment.replies || []).map(reply => ({
          id: reply.id,
          userId: reply.userId || reply.user?.id,
          author: reply.userName || reply.author || '未知用户',
          avatar: reply.userAvatar || reply.avatar || '/default-avatar.svg',
          userLevel: reply.userLevel || 0,
          content: reply.content,
          time: reply.time || formatDate(reply.createTime),
          likeCount: reply.likeCount || 0,
          dislikeCount: reply.dislikeCount || 0,
          isLiked: reply.liked || false,
          targetAuthor: reply.replyToUserName || reply.targetAuthor || reply.replyTo || reply.toUserName || null
        }))
      }))
    }
  } catch (error) {
    console.error('获取评论失败:', error)
  } finally {
    loadingComments.value = false
  }
}

const showCommentUserCard = ref(false)
const commentAvatarRef = ref(null)
const currentCommentUser = ref({
  id: null,
  name: '',
  avatar: '',
  bio: '',
  following: false,
  followerCount: 0,
  followingCount: 0,
  likeCount: 0,
  level: 0
})
const commentCardTimer = ref(null)
const commentUserCache = ref({})

const handleCommentAvatarMouseEnter = async (event, comment) => {
  if (commentCardTimer.value) {
    clearTimeout(commentCardTimer.value)
    commentCardTimer.value = null
  }
  commentAvatarRef.value = event.target

  const cached = commentUserCache.value[comment.userId]
  currentCommentUser.value = cached || {
    id: comment.userId,
    name: comment.author,
    avatar: comment.avatar,
    bio: '',
    signature: '',
    following: false,
    followerCount: 0,
    followingCount: 0,
    likeCount: 0,
    level: comment.userLevel || 0
  }
  showCommentUserCard.value = true

  if (!cached && comment.userId) {
    try {
      const res = await userApi.getUserById(comment.userId)
      if (res.code === 200 && res.data) {
        const u = res.data
        const fullInfo = {
          id: u.id,
          name: u.nickname || u.username || comment.author,
          avatar: u.avatar || comment.avatar,
          bio: u.bio || u.signature || '',
          signature: u.signature || '',
          following: false,
          followerCount: u.followerCount || 0,
          followingCount: u.followingCount || 0,
          likeCount: u.totalLikeCount || u.likedCount || 0,
          level: u.level || comment.userLevel || 0
        }
        commentUserCache.value[comment.userId] = fullInfo
        if (currentCommentUser.value.id === comment.userId) {
          currentCommentUser.value = fullInfo
        }
      }
    } catch (e) {}
  }
}

const handleCommentAvatarMouseLeave = () => {
  commentCardTimer.value = setTimeout(() => {
    showCommentUserCard.value = false
  }, 300)
}

const handleCommentCardMouseEnter = () => {
  if (commentCardTimer.value) {
    clearTimeout(commentCardTimer.value)
    commentCardTimer.value = null
  }
}

const handleCommentCardMouseLeave = () => {
  commentCardTimer.value = setTimeout(() => {
    showCommentUserCard.value = false
  }, 300)
}

const handleCommentFollowChange = ({ userId, isFollowing: newFollowStatus }) => {
  if (currentCommentUser.value.id === userId) {
    currentCommentUser.value.following = newFollowStatus
  }
}

watch(commentSort, newSort => {
  if (props.manuscriptId) {
    loadComments(newSort)
  }
})

watch(
  () => props.manuscriptId,
  newId => {
    if (newId) {
      loadComments(commentSort.value)
    }
  }
)

onMounted(() => {
  loadComments(commentSort.value)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="comment-section">
    <div class="comment-header">
      <h3>评论 ({{ (commentCount || 0).toLocaleString() }})</h3>
      <div class="comment-sort">
        <span
          class="sort-item"
          :class="{ 'is-active': commentSort === 'hot' }"
          @click="commentSort = 'hot'"
        >最热</span>
        <span
          class="sort-item"
          :class="{ 'is-active': commentSort === 'new' }"
          @click="commentSort = 'new'"
        >最新</span>
      </div>
    </div>

    <div class="comment-input-wrapper" ref="commentInputWrapper">
      <img loading="lazy" decoding="async" :src="currentUserAvatar()" alt="用户头像" class="comment-input-avatar">

      <div
        class="comment-input-collapsed"
        @click.stop="toggleCommentInput"
        v-if="isCommentInputCollapsed"
      >
        <span class="placeholder-text">发一条友善的评论吧...</span>
      </div>

      <div class="comment-input-expanded" v-else @click.stop>
        <el-input
          v-model="newComment"
          type="textarea"
          :rows="4"
          placeholder="发一条友善的评论吧..."
          maxlength="500"
          show-word-limit
          resize="none"
        />

        <div class="emoji-picker" v-if="showEmojiPicker" @click.stop>
          <div
            v-for="emoji in emojiList"
            :key="emoji"
            class="emoji-item"
            @click="selectEmoji(emoji)"
          >
            {{ emoji }}
          </div>
        </div>

        <div class="comment-input-actions">
          <el-button
            text
            size="small"
            class="emoji-btn"
            @click="showEmojiPicker = !showEmojiPicker; showReplyEmojiPicker = {}; activeReplyEmojiCommentId = null"
          >
            😊 表情
          </el-button>
          <el-button type="primary" size="small" @click="submitComment">发表评论</el-button>
        </div>
      </div>
    </div>

    <div class="comment-list">
      <div v-if="loadingComments" class="loading-comments">
        <el-skeleton :rows="5" animated />
      </div>
      <div v-else-if="comments.length === 0" class="no-comments">
        <p>暂无评论，快来抢沙发吧！</p>
      </div>
      <div v-else>
        <div v-for="comment in comments" :key="comment.id" class="comment-item">
          <img loading="lazy" decoding="async"
            :src="comment.avatar || '/default-avatar.svg'"
            alt="用户头像"
            class="comment-avatar"
            @click="goToAuthor(comment.userId)"
            @mouseenter="(e) => handleCommentAvatarMouseEnter(e, comment)"
            @mouseleave="handleCommentAvatarMouseLeave"
          >
          <div class="comment-content">
            <div class="comment-info">
              <span class="comment-author" @click="goToAuthor(comment.userId)">{{ comment.author }}</span>
              <LevelBadge :level="comment.userLevel" />
            </div>
            <div class="comment-text" v-html="formatContentWithAtLinks(comment.content)"></div>
            <div class="comment-actions">
              <span class="comment-time">{{ comment.time }}</span>
              <el-button text size="small" :class="{ 'liked': comment.isLiked }" @click="likeComment(comment.id)">
                <el-icon><CircleCheck /></el-icon>
                {{ comment.likeCount }}
              </el-button>
              <el-button text size="small" :class="{ 'disliked': comment.isDisliked }" @click="dislikeComment(comment.id)">
                <el-icon><CircleClose /></el-icon>
              </el-button>
              <el-button text size="small" @click.stop="replyComment(comment.id)">回复</el-button>
            </div>

            <div class="replies-list" v-if="getReplyTotal(comment) > 0">
              <div v-for="reply in getUploaderReplies(comment)" :key="reply.id" class="reply-item">
                <img loading="lazy" decoding="async" :src="reply.avatar || '/default-avatar.svg'" alt="用户头像" class="reply-avatar" @click="goToAuthor(reply.userId)">
                <div class="reply-content">
                  <div class="reply-text">
                    <span class="reply-author" @click="goToAuthor(reply.userId)">{{ reply.author }}</span>
                    <LevelBadge :level="reply.userLevel" />
                    <span class="reply-colon">: </span>
                    <span v-html="formatContentWithAtLinks(reply.content)"></span>
                  </div>
                  <div class="reply-actions">
                    <span class="reply-time">{{ reply.time }}</span>
                    <el-button text size="small" :class="{ 'liked': reply.isLiked }" @click="likeReply(reply.id)">
                      <el-icon><CircleCheck /></el-icon>
                      {{ reply.likeCount }}
                    </el-button>
                    <el-button text size="small" :class="{ 'disliked': reply.isDisliked }" @click="dislikeReply(reply.id)">
                      <el-icon><CircleClose /></el-icon>
                    </el-button>
                    <el-button text size="small" @click.stop="replyToReply(comment.id, reply.id, reply.author)">回复</el-button>
                  </div>
                </div>
              </div>

              <div v-if="shouldShowReplyExpand(comment)">
                <div class="reply-collapse" v-if="!replyExpanded[comment.id]">
                  <el-button text size="small" @click="toggleReplyExpanded(comment.id)">
                    显示全部 {{ getReplyTotal(comment) }} 条回复
                  </el-button>
                </div>

                <div v-if="replyExpanded[comment.id]">
                  <div v-for="reply in getOtherReplies(comment)" :key="reply.id" class="reply-item">
                    <img loading="lazy" decoding="async" :src="reply.avatar || '/default-avatar.svg'" alt="用户头像" class="reply-avatar" @click="goToAuthor(reply.userId)">
                    <div class="reply-content">
                      <div class="reply-text">
                        <span class="reply-author" @click="goToAuthor(reply.userId)">{{ reply.author }}</span>
                        <LevelBadge :level="reply.userLevel" />
                        <span class="reply-colon">: </span>
                        <span v-html="formatContentWithAtLinks(reply.content)"></span>
                      </div>
                      <div class="reply-actions">
                        <span class="reply-time">{{ reply.time }}</span>
                        <el-button text size="small" :class="{ 'liked': reply.isLiked }" @click="likeReply(reply.id)">
                          <el-icon><CircleCheck /></el-icon>
                          {{ reply.likeCount }}
                        </el-button>
                        <el-button text size="small" :class="{ 'disliked': reply.isDisliked }" @click="dislikeReply(reply.id)">
                          <el-icon><CircleClose /></el-icon>
                        </el-button>
                        <el-button text size="small" @click.stop="replyToReply(comment.id, reply.id, reply.author)">回复</el-button>
                      </div>
                    </div>
                  </div>

                  <div class="reply-pagination">
                    <span class="page-info">共{{ getReplyPageCount(comment) }}页</span>
                    <span
                      v-for="page in getReplyPageCount(comment)"
                      :key="page"
                      class="page-number"
                      :class="{ 'active': replyPage[comment.id] === page }"
                      @click="loadReplies(comment.id, page)"
                    >
                      {{ page }}
                    </span>
                    <span class="page-control" @click="loadReplies(comment.id, (replyPage[comment.id] || 1) + 1)" v-if="(replyPage[comment.id] || 1) < getReplyPageCount(comment)">
                      下一页
                    </span>
                    <span class="page-control" @click="replyExpanded[comment.id] = false">
                      收起
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div class="reply-input-wrapper" v-if="comment.showReplyInput" @click.stop>
              <img loading="lazy" decoding="async" :src="currentUserAvatar()" alt="用户头像" class="reply-input-avatar">
              <div class="reply-input-content">
                <div v-if="replyTargets[comment.id]" class="reply-target-info">
                  回复 @{{ replyTargets[comment.id].targetAuthor }}
                </div>
                <el-input
                  v-model="replyInputs[comment.id]"
                  type="textarea"
                  :rows="3"
                  :placeholder="replyTargets[comment.id] ? `回复 @${replyTargets[comment.id].targetAuthor}...` : '回复评论...'"
                  maxlength="500"
                  show-word-limit
                  resize="none"
                />
                <div class="reply-input-actions">
                  <div class="emoji-picker reply-emoji-picker" v-if="showReplyEmojiPicker[comment.id]" @click.stop>
                    <div
                      v-for="emoji in emojiList"
                      :key="emoji"
                      class="emoji-item"
                      @click="selectReplyEmoji(emoji, comment.id)"
                    >
                      {{ emoji }}
                    </div>
                  </div>
                  <el-button
                    text
                    size="small"
                    class="emoji-btn"
                    @click="toggleReplyEmojiPicker(comment.id)"
                  >
                    😊 表情
                  </el-button>
                  <el-button type="primary" size="small" @click="submitReply(comment.id)">发表回复</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <UserFloatCard
    v-model:visible="showCommentUserCard"
    :trigger-ref="commentAvatarRef"
    :user-info="currentCommentUser"
    :auto-placement="true"
    @follow-change="handleCommentFollowChange"
    @mouseenter="handleCommentCardMouseEnter"
    @mouseleave="handleCommentCardMouseLeave"
  />
</template>

<style scoped>
.comment-section {
  background-color: #fff;
  padding: 20px 0;
  margin-top: 20px;
}

.comment-section .comment-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}

.comment-section .comment-header h3 {
  font-size: 18px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.comment-section .comment-sort {
  display: flex;
  gap: 20px;
}

.comment-section .sort-item {
  font-size: 14px;
  color: #666;
  cursor: pointer;
  transition: all 0.3s ease;
  user-select: none;
}

.comment-section .sort-item:hover {
  color: #00a1d6;
}

.reply-pagination {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
  font-size: 14px;
  color: #666;
}

.reply-pagination .page-info {
  margin-right: 5px;
}

.reply-pagination .page-number {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.reply-pagination .page-number:hover {
  background-color: #f0f0f0;
  color: #00a1d6;
}

.reply-pagination .page-number.active {
  background-color: #00a1d6;
  color: #fff;
}

.reply-pagination .page-control {
  cursor: pointer;
  color: #00a1d6;
  transition: color 0.3s ease;
}

.reply-pagination .page-control:hover {
  color: #0091c6;
  text-decoration: underline;
}

.comment-section .sort-item.is-active {
  font-weight: 600;
  color: #333;
}

.comment-section .comment-input-wrapper {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 30px;
}

.comment-section .comment-input-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.comment-section .comment-input-collapsed {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: #fff;
}

.comment-section .comment-input-collapsed:hover {
  border-color: #00a1d6;
  background-color: #f5f5f5;
}

.comment-section .comment-input-collapsed .placeholder-text {
  color: #999;
  font-size: 14px;
}

.comment-section .comment-input-expanded {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px;
  background-color: #fff;
}

.comment-section .comment-input-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.comment-section .emoji-btn {
  color: #666;
  font-size: 14px;
  padding: 4px 8px;
}

.comment-section .emoji-btn:hover {
  color: #00a1d6;
  background-color: #f5f5f5;
}

.comment-section .emoji-picker {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  margin-bottom: 12px;
  background-color: #fff;
  max-height: 200px;
  overflow-y: auto;
}

.comment-section .emoji-picker::-webkit-scrollbar {
  width: 6px;
}

.comment-section .emoji-picker::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.comment-section .emoji-picker::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.comment-section .emoji-item {
  font-size: 24px;
  cursor: pointer;
  text-align: center;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
  user-select: none;
}

.comment-section .emoji-item:hover {
  background-color: #f5f5f5;
  transform: scale(1.2);
}

.comment-section .comment-input-expanded :deep(.el-textarea__inner) {
  border: none;
  border-radius: 0;
  resize: none;
  font-size: 14px;
  padding: 0;
}

.comment-section .comment-input-expanded :deep(.el-textarea__inner):focus {
  border: none;
  box-shadow: none;
}

.comment-section .comment-input-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 10px;
}

.comment-section .comment-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.comment-section .comment-item {
  display: flex;
  gap: 12px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.comment-section .comment-item:last-child {
  border-bottom: none;
}

.comment-section .comment-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.comment-section .comment-content {
  flex: 1;
  min-width: 0;
}

.comment-section .comment-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.comment-section .comment-author {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.comment-section .comment-time {
  font-size: 12px;
  color: #999;
}

.comment-section .comment-text {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  margin-bottom: 10px;
  word-wrap: break-word;
}

.comment-section .comment-actions {
  display: flex;
  gap: 4px;
  align-items: center;
}

.comment-section .comment-actions .comment-time {
  font-size: 12px;
  color: #999;
  margin-right: 10px;
}

.comment-section .comment-actions .el-button {
  color: #999;
  font-size: 13px;
}

.comment-section .comment-actions .el-button:hover {
  color: #00a1d6;
}

.comment-section .comment-actions .el-button.liked {
  color: #00a1d6;
}

.comment-section .comment-actions .el-button.disliked {
  color: #f56c6c;
}

.comment-section .comment-actions .el-icon {
  margin-right: 4px;
}

.comment-section .replies-list {
  margin-top: 15px;
  padding-left: 60px;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.comment-section .reply-item {
  margin-left: -60px;
  padding-left: 60px;
}

.comment-section .reply-item {
  display: flex;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  align-items: flex-start;
}

.comment-section .reply-item:last-child {
  border-bottom: none;
}

.comment-section .reply-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  margin-top: 2px;
}

.comment-section .reply-content {
  flex: 1;
  min-width: 0;
}

.comment-section .reply-author {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  cursor: pointer;
}

.comment-section .reply-time {
  font-size: 12px;
  color: #999;
  margin-right: 10px;
}

.comment-section .reply-text {
  font-size: 13px;
  color: #333;
  line-height: 1.5;
  margin-bottom: 8px;
  word-wrap: break-word;
}

.comment-section .reply-actions {
  display: flex;
  gap: 4px;
  align-items: center;
  flex-wrap: nowrap;
}

.comment-section .reply-target {
  color: #00a1d6;
  margin-right: 5px;
}

.comment-section .user-link {
  color: #00a1d6;
  text-decoration: none;
}

.comment-section .user-link:hover {
  text-decoration: underline;
}

.comment-section .reply-collapse {
  margin-top: 10px;
  padding-left: 0;
  margin-left: -60px;
  text-align: left;
}

.comment-section .reply-load-more {
  margin-top: 15px;
  padding-left: 60px;
  margin-left: -60px;
  text-align: left;
}

.comment-section .reply-collapse .el-button,
.comment-section .reply-load-more .el-button {
  color: #00a1d6;
  font-size: 13px;
}

.comment-section .reply-collapse .el-button:hover,
.comment-section .reply-load-more .el-button:hover {
  color: #0091c6;
}

.comment-section .reply-actions .el-button {
  color: #999;
  font-size: 12px;
}

.comment-section .reply-actions .el-button:hover {
  color: #00a1d6;
}

.comment-section .reply-actions .el-button.liked {
  color: #00a1d6;
}

.comment-section .reply-actions .el-button.disliked {
  color: #f56c6c;
}

.comment-section .reply-input-wrapper {
  display: flex;
  gap: 12px;
  margin-top: 15px;
  padding: 15px;
  background-color: #f8f8f8;
  border-radius: 8px;
}

.comment-section .reply-input-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.comment-section .reply-input-content {
  flex: 1;
  min-width: 0;
}

.comment-section .reply-input-content :deep(.el-textarea__inner) {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  resize: none;
  font-size: 14px;
  padding: 12px;
}

.comment-section .reply-input-content :deep(.el-textarea__inner):focus {
  border-color: #00a1d6;
}

.comment-section .reply-input-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 10px;
}

.comment-section .reply-input-actions .emoji-btn {
  color: #666;
  font-size: 13px;
}

.comment-section .reply-input-actions .emoji-btn:hover {
  color: #00a1d6;
}

.comment-section .reply-input-actions .reply-emoji-picker {
  position: absolute;
  bottom: 100%;
  left: 0;
  margin-bottom: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
}

.comment-section .reply-input-actions {
  position: relative;
}

.comment-section .loading-comments {
  padding: 20px 0;
}

.comment-section .no-comments {
  text-align: center;
  padding: 40px 0;
  color: #999;
  font-size: 14px;
}
</style>