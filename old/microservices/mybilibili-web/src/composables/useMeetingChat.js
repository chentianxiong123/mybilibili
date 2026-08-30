import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sameUserId } from '../utils/userId.js'

const DEFAULT_EMOJIS = ['😀', '😂', '👍', '👏', '❤️', '🔥', '🤔', '😅', '🙌', '😎', '🤝', '👀', '🎉', '💯', '😱', '🤯']

export function useMeetingChat({ me, participantList, sendMeetingMessage }) {
  const chatOpen = ref(false)
  const chatMessages = ref([])
  const chatText = ref('')
  const showEmojiPicker = ref(false)
  const emojis = DEFAULT_EMOJIS
  const showMentionDropdown = ref(false)
  const mentionQuery = ref('')
  const mentionSelected = ref(null)

  const mentionSuggestions = computed(() => {
    if (!showMentionDropdown.value) return []
    const q = mentionQuery.value.toLowerCase()
    return participantList.value.filter(p => (p.name || '').toLowerCase().includes(q) && !p.self)
  })

  const toggleEmojiPicker = () => {
    showEmojiPicker.value = !showEmojiPicker.value
  }

  const insertEmoji = (emoji) => {
    chatText.value += emoji
    showEmojiPicker.value = false
  }

  const handleChatInput = () => {
    const val = chatText.value
    const atIdx = val.lastIndexOf('@')
    if (atIdx >= 0 && atIdx === val.length - 1) {
      showMentionDropdown.value = true
      mentionQuery.value = ''
    } else if (atIdx >= 0) {
      mentionQuery.value = val.slice(atIdx + 1)
      showMentionDropdown.value = true
    } else {
      showMentionDropdown.value = false
    }
  }

  const selectMention = (p) => {
    const atIdx = chatText.value.lastIndexOf('@')
    chatText.value = chatText.value.slice(0, atIdx) + '@' + p.name + ' '
    mentionSelected.value = p.userId
    showMentionDropdown.value = false
  }

  const pushLocalMessage = (text, mentionId) => {
    chatMessages.value.push({
      id: Date.now(),
      from: me.value.name,
      text,
      self: true,
      mentionId
    })
    if (chatMessages.value.length > 200) chatMessages.value.splice(0, 50)
  }

  const sendChat = () => {
    const text = chatText.value.trim()
    if (!text) return false
    const mentionId = mentionSelected.value
    const ok = sendMeetingMessage({
      type: 'chat',
      targetUserId: mentionId,
      data: { text }
    })
    if (!ok) return false
    pushLocalMessage(text, mentionId)
    chatText.value = ''
    mentionSelected.value = null
    showMentionDropdown.value = false
    return true
  }

  const receiveChatMessage = (msg) => {
    const text = typeof msg?.data?.text === 'string' ? msg.data.text : ''
    if (!text) return false
    const fromSelf = sameUserId(msg.userId, me.value.id)
    const directlyMentioned = sameUserId(msg.targetUserId, me.value.id)
    const isMentioned = directlyMentioned || text.includes('@' + me.value.name)
    chatMessages.value.push({
      id: Date.now() + Math.random(),
      from: msg.userName || ('用户' + msg.userId),
      text,
      self: false,
      mentionId: directlyMentioned ? msg.userId : (isMentioned ? msg.userId : null)
    })
    if (chatMessages.value.length > 200) chatMessages.value.splice(0, 50)
    if (isMentioned && !fromSelf) {
      ElMessage.info(`${msg.userName || ('用户' + msg.userId)} 在会议中提到了你`)
    }
    return true
  }

  return {
    chatOpen,
    chatMessages,
    chatText,
    showEmojiPicker,
    emojis,
    showMentionDropdown,
    mentionQuery,
    mentionSelected,
    mentionSuggestions,
    toggleEmojiPicker,
    insertEmoji,
    handleChatInput,
    selectMention,
    sendChat,
    receiveChatMessage
  }
}
