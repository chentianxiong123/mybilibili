import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { isNotificationEnabled, notifyMention } from '../utils/notification.js'
import { sameUserId } from '../utils/userId.js'

const DANMAKU_COLORS = ['#fff', '#ff6b81', '#fad34a', '#4fc3f7', '#a78bfa', '#34d399']
const REACTION_EMOJIS = ['❤️', '😂', '👍', '👏', '🔥', '🎉', '😮', '🤔']
const CHAT_EMOJIS = ['😀', '😂', '👍', '👏', '❤️', '🔥', '🤔', '😅', '🙌', '😎', '🤝', '👀', '🎉', '💯', '😱', '🤯']
const GIFT_OPTIONS = [
  { emoji: '👏', name: '鼓掌', cost: 10 },
  { emoji: '🌹', name: '鲜花', cost: 50 },
  { emoji: '🚀', name: '飞船', cost: 100 },
  { emoji: '⭐', name: '星星', cost: 200 },
  { emoji: '666', name: '666', cost: 66 },
  { emoji: '❤️', name: '爱心', cost: 30 }
]
const getMessageText = (msg) => typeof msg?.data?.text === 'string' ? msg.data.text : ''

export function useLiveRoomInteractions({
  room,
  me,
  currentUserId,
  sendRoomMessage,
  emitDanmaku,
  setTimeoutFn = (handler, delay) => setTimeout(handler, delay),
  clearTimeoutFn = timerId => clearTimeout(timerId)
}) {
  const danmakuList = ref([])
  const danmakuText = ref('')
  const showEmojiPicker = ref(false)
  const reactionEmojis = REACTION_EMOJIS
  const chatEmojis = CHAT_EMOJIS
  const reactions = ref([])
  const pinnedMessage = ref(null)
  const gifts = ref([])
  const showGiftPicker = ref(false)
  const giftOptions = GIFT_OPTIONS
  const reactionTimers = new Set()
  const giftTimers = new Set()

  const scheduleCleanup = (timers, handler, delay) => {
    const timerId = setTimeoutFn(() => {
      timers.delete(timerId)
      handler()
    }, delay)
    timers.add(timerId)
    return timerId
  }

  const clearTimers = (timers) => {
    for (const timerId of timers) {
      clearTimeoutFn(timerId)
    }
    timers.clear()
  }

  const pushDanmaku = ({ text, color, from, id = Date.now() + Math.random() }) => {
    const item = { id, text, color: color || DANMAKU_COLORS[0], from }
    danmakuList.value.push(item)
    if (danmakuList.value.length > 200) danmakuList.value.splice(0, 50)
    emitDanmaku(item.text, item.color)
    return item
  }

  const addReaction = (reaction) => {
    const id = Date.now() + Math.random()
    reactions.value.push({ id, ...reaction })
    if (reactions.value.length > 20) reactions.value.shift()
    scheduleCleanup(reactionTimers, () => {
      reactions.value = reactions.value.filter(r => r.id !== id)
    }, 1500)
  }

  const pushGift = ({ emoji, name, from }) => {
    const id = Date.now() + Math.random()
    gifts.value.push({ id, emoji, name, from })
    scheduleCleanup(giftTimers, () => {
      gifts.value = gifts.value.filter(g => g.id !== id)
    }, 3000)
  }

  const receiveGift = (msg) => {
    if (!msg.data?.emoji) return
    pushGift({
      emoji: msg.data.emoji,
      name: msg.data.name,
      from: msg.userName || '观众'
    })
  }

  const receiveReaction = (msg) => {
    if (!msg.data?.emoji) return
    addReaction({
      emoji: msg.data.emoji,
      x: msg.data.x,
      y: msg.data.y,
      from: msg.userName || ('用户' + msg.userId)
    })
  }

  const receivePinMessage = (msg) => {
    if (!msg.data?.text) return
    pinnedMessage.value = {
      text: msg.data.text,
      from: msg.userName || '主播',
      time: Date.now()
    }
  }

  const clearPinnedMessage = () => {
    pinnedMessage.value = null
  }

  const receiveDanmaku = (msg) => {
    const text = getMessageText(msg)
    if (!text) return null
    const item = pushDanmaku({
      text,
      color: msg.data.color || DANMAKU_COLORS[0],
      from: msg.userName || ('用户' + msg.userId)
    })
    if (isNotificationEnabled() && text.includes('@' + me.value.name) && !sameUserId(msg.userId, currentUserId.value)) {
      notifyMention(msg.userName || ('用户' + msg.userId), room.value?.roomName || '直播间', text)
    }
    return item
  }

  const sendReaction = (emoji) => {
    const x = Math.random() * 80 + 10
    const y = Math.random() * 60 + 20
    if (!sendRoomMessage({ type: 'reaction', data: { emoji, x, y } })) return
    addReaction({ emoji, x, y, from: me.value.name })
  }

  const toggleEmojiPicker = () => {
    showEmojiPicker.value = !showEmojiPicker.value
  }

  const insertChatEmoji = (emoji) => {
    danmakuText.value += emoji
    showEmojiPicker.value = false
  }

  const sendGift = (gift) => {
    if (!sendRoomMessage({
      type: 'gift',
      data: { emoji: gift.emoji, name: gift.name, cost: gift.cost }
    })) return
    pushGift({ emoji: gift.emoji, name: gift.name, from: me.value.name })
    ElMessage.success(`赠送了 ${gift.name}`)
    showGiftPicker.value = false
  }

  const pinMessage = (item) => {
    if (!sendRoomMessage({ type: 'pin-message', data: { text: item.text || item.from } })) return
    pinnedMessage.value = {
      text: item.text || item.from,
      from: item.from || '观众',
      time: Date.now()
    }
  }

  const unpinMessage = () => {
    if (!sendRoomMessage({ type: 'unpin-message' })) return
    pinnedMessage.value = null
  }

  const sendDanmaku = () => {
    const text = danmakuText.value.trim()
    if (!text) return
    const color = DANMAKU_COLORS[Math.floor(Math.random() * DANMAKU_COLORS.length)]
    sendRoomMessage({ type: 'chat', data: { text, color } })
    pushDanmaku({
      id: Date.now(),
      text,
      color,
      from: me.value.name
    })
    danmakuText.value = ''
  }

  const cleanupLiveRoomInteractions = () => {
    clearTimers(reactionTimers)
    clearTimers(giftTimers)
    reactions.value = []
    gifts.value = []
    showEmojiPicker.value = false
    showGiftPicker.value = false
  }

  return {
    danmakuList,
    danmakuText,
    showEmojiPicker,
    reactionEmojis,
    chatEmojis,
    reactions,
    pinnedMessage,
    gifts,
    showGiftPicker,
    giftOptions,
    sendReaction,
    receiveReaction,
    receiveGift,
    receivePinMessage,
    clearPinnedMessage,
    receiveDanmaku,
    toggleEmojiPicker,
    insertChatEmoji,
    sendGift,
    pinMessage,
    unpinMessage,
    sendDanmaku,
    cleanupLiveRoomInteractions
  }
}
