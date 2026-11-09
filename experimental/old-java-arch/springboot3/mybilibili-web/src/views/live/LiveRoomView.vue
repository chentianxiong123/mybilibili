<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Phone, PhoneFilled, Microphone, Mute, VideoCamera, VideoPause, Share } from '@element-plus/icons-vue'
import flvJs from 'flv.js'
import Artplayer from 'artplayer'
import ArtplayerPluginDanmuku from 'artplayer-plugin-danmuku'
import { liveApi } from '../../api/live.js'
import { linkmicApi } from '../../api/linkmic.js'
import { createReconnectingWS } from '../../utils/reconnectingWs.js'
import { requestNotificationPermission, isNotificationEnabled, notifyMention, notifyStreamLive } from '../../utils/notification.js'

const route = useRoute()
const roomId = route.params.roomId
const room = ref(null)
const playerRef = ref(null)
const loading = ref(true)
const isBuffering = ref(false)
let bufferingTimer = null
let art = null

const showBuffering = () => {
  clearTimeout(bufferingTimer)
  isBuffering.value = true
}
const hideBuffering = () => {
  bufferingTimer = setTimeout(() => { isBuffering.value = false }, 500)
}
const danmakuList = ref([])
const danmakuText = ref('')

const danmakuColors = ['#fff', '#ff6b81', '#fad34a', '#4fc3f7', '#a78bfa', '#34d399']
const reactionEmojis = ['❤️', '😂', '👍', '👏', '🔥', '🎉', '😮', '🤔']
const showEmojiPicker = ref(false)
const chatEmojis = ['😀', '😂', '👍', '👏', '❤️', '🔥', '🤔', '😅', '🙌', '😎', '🤝', '👀', '🎉', '💯', '😱', '🤯']

// 连麦相关
const myLinkmicId = ref(null)
const myQueuePosition = ref(0)
const activeLinkmics = ref([])
const pendingApplications = ref([])
const linkmicLoading = ref(false)
const reactions = ref([])  // { id, emoji, x, y, from }
const pinnedMessage = ref(null)  // { text, from, time }
const gifts = ref([])  // { id, emoji, name, from, count }
const showGiftPicker = ref(false)
const giftOptions = [
  { emoji: '👏', name: '鼓掌', cost: 10 },
  { emoji: '🌹', name: '鲜花', cost: 50 },
  { emoji: '🚀', name: '飞船', cost: 100 },
  { emoji: '⭐', name: '星星', cost: 200 },
  { emoji: '666', name: '666', cost: 66 },
  { emoji: '❤️', name: '爱心', cost: 30 }
]

// 回放相关
const showReplayTab = ref(false)
const replayList = ref([])  // [{ id, title, coverUrl, url, duration, createTime }]
const currentReplay = ref(null)
const replayPlayerRef = ref(null)
let replayArt = null

const loadReplays = async () => {
  // TODO: 后端提供回放列表 API
  // 模拟数据（后端接入后可替换为: const res = await liveApi.getReplays(roomId)）
  replayList.value = []
}

const playReplay = async (replay) => {
  currentReplay.value = replay
  await nextTick()
  if (replayPlayerRef.value && replay.url) {
    replayArt = new Artplayer({
      container: replayPlayerRef.value,
      url: replay.url,
      type: replay.url.includes('.m3u8') ? 'm3u8' : undefined,
      autoplay: true,
      setting: true,
      fullscreen: true,
      playbackRate: true,
    })
  }
}

const closeReplay = () => {
  currentReplay.value = null
  if (replayArt) {
    replayArt.destroy()
    replayArt = null
  }
}

// peerId(=对端 userId) -> { stream, name }
const linkmicStreams = reactive({})
const linkmicPeerConnections = {}
const pendingIce = {}
let linkmicWs = null
const realtimeViewerCount = ref(0)

// 当前用户：从 localStorage（登录态可靠）
const getCurrentUser = () => {
  try {
    const u = JSON.parse(localStorage.getItem('user') || '{}')
    if (u && u.id) return { id: Number(u.id), name: u.username || u.nickname || ('用户' + u.id) }
  } catch (e) {}
  return { id: 0, name: '游客' }
}
const me = ref(getCurrentUser())
const currentUserId = ref(me.value.id)
const isStreamer = ref(false)

// 观众端的本地媒体流（连麦后产生）
const localLinkmicStream = ref(null)
const localLinkmicVideoRef = ref(null)
const localLinkmicAudioEnabled = ref(true)
const localLinkmicVideoEnabled = ref(true)
const viewerHandRaised = ref(false)
const currentQuality = ref('自动')
const bufferStallCount = ref(0)
const isFollowing = ref(false)
const followerCount = ref(0)

onMounted(async () => {
  // 请求通知权限
  requestNotificationPermission()
  try {
    const res = await liveApi.getRoom(roomId)
    if (res.code === 200) {
      room.value = res.data
      isStreamer.value = currentUserId.value === room.value?.userId
      // 检查是否已关注
      try {
        const followed = JSON.parse(localStorage.getItem('followedRooms') || '[]')
        isFollowing.value = followed.includes(room.value.userId)
      } catch (e) {}
      followerCount.value = room.value.followerCount || 0
    } else {
      ElMessage.error(res.message || '直播间不存在')
    }
  } catch (e) {
    console.error('[LiveRoom] 加载失败:', e)
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
    await nextTick()
    if (room.value) initPlayer()
  }

  await loadLinkmicStatus()
  connectLinkmicWebSocket()
})

onUnmounted(() => {
  if (art) {
    try { art.destroy() } catch (e) {}
    art = null
  }
  cleanupLinkmic()
})

const shareRoom = () => {
  const url = window.location.href
  navigator.clipboard.writeText(url).then(() => {
    ElMessage.success('直播间链接已复制，可分享给好友！')
  })
}

const toggleFollow = () => {
  if (!room.value) return
  try {
    const followed = JSON.parse(localStorage.getItem('followedRooms') || '[]')
    if (isFollowing.value) {
      isFollowing.value = false
      const updated = followed.filter(id => id !== room.value.userId)
      localStorage.setItem('followedRooms', JSON.stringify(updated))
      ElMessage.success('已取消关注')
    } else {
      isFollowing.value = true
      followed.push(room.value.userId)
      localStorage.setItem('followedRooms', JSON.stringify(followed))
      followerCount.value++
      ElMessage.success('已关注主播')
    }
  } catch (e) {}
}

const initPlayer = () => {
  if (!room.value || !playerRef.value) return
  const flvUrl = `http://${window.location.hostname}:28080/live/${room.value.streamKey}.flv`

  art = new Artplayer({
    container: playerRef.value,
    url: flvUrl,
    type: 'flv',
    isLive: true,                  // 启用直播模式（隐藏进度条、跳跃等）
    autoplay: true,
    muted: true,                   // 浏览器策略：自动播放必须先静音
    autoSize: false,
    autoOrientation: true,
    setting: true,                 // 启用设置面板（含画质选择）
    playbackRate: false,
    aspectRatio: false,
    fullscreen: true,
    fullscreenWeb: true,
    pip: true,
    miniProgressBar: false,
    mutex: true,
    backdrop: true,
    theme: '#fb7299',
    customType: {
      flv: function (video, url) {
        if (flvJs.isSupported()) {
          const flvPlayer = flvJs.createPlayer({
            type: 'flv',
            url,
            isLive: true,
            config: {
              enableWorker: true,
              enableStashBuffer: false,
              stashInitialSize: 128,
              autoCleanupSourceBuffer: true,
              autoCleanupMaxDuration: 3,
              autoCleanupMinRecoverDuration: 2,
            }
          })
          flvPlayer.attachMediaElement(video)
          flvPlayer.load()
          flvPlayer.on(flvJs.Events.ERROR, (errType, errDetail) => {
            console.error('[LiveRoom] FLV 错误:', errType, errDetail)
          })
          // 把 flvPlayer 句柄挂到 art 上以便销毁时一起 cleanup
          art.flvPlayer = flvPlayer
        }
      }
    },
    plugins: [
      ArtplayerPluginDanmuku({
        danmuku: [],               // 直播弹幕实时推送，无需历史
        speed: 5,
        opacity: 1,
        fontSize: 22,
        synchronousPlayback: false
      })
    ]
  })

  art.on('video:waiting', () => showBuffering())
  art.on('video:playing', () => {
    hideBuffering()
    bufferStallCount.value = 0
  })
  art.on('video:stalled', () => {
    bufferStallCount.value++
    if (bufferStallCount.value >= 3) {
      currentQuality.value = '流畅'
      ElMessage.warning('当前网络不稳定，已自动切换到流畅模式')
    }
  })
  art.on('error', (e) => console.error('[LiveRoom] 播放器错误:', e))

  art.on('destroy', () => {
    if (art && art.flvPlayer) {
      try {
        art.flvPlayer.pause()
        art.flvPlayer.unload()
        art.flvPlayer.detachMediaElement()
        art.flvPlayer.destroy()
      } catch (e) {}
      art.flvPlayer = null
    }
  })
}

// ============ 连麦核心 ============
const loadLinkmicStatus = async () => {
  try {
    const res = await linkmicApi.getActiveLinkmics(roomId)
    if (res.code === 200) {
      activeLinkmics.value = res.data || []
    }
    if (isStreamer.value) {
      const pendingRes = await linkmicApi.getPendingApplications(roomId)
      if (pendingRes.code === 200) {
        pendingApplications.value = pendingRes.data || []
      }
    }
  } catch (e) {
    console.error('加载连麦状态失败:', e)
  }
}

const connectLinkmicWebSocket = () => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  linkmicWs = createReconnectingWS({
    url: `${proto}//${location.host}/ws/meeting`,
    onOpen: () => {
      // 用 streamKey 作为 roomCode 加入房间，便于主播/观众间路由（含弹幕）
      if (room.value?.streamKey && currentUserId.value) {
        linkmicWs.send({
          type: 'join',
          roomCode: room.value.streamKey,
          userId: currentUserId.value,
          userName: me.value.name
        })
      }
    },
    onMessage: async (msg) => {
      switch (msg.type) {
        case 'linkmic-apply':
          if (isStreamer.value && msg.data) {
            if (!pendingApplications.value.find(a => a.id === msg.data.id)) {
              pendingApplications.value.push(msg.data)
            }
          }
          break
        case 'linkmic-accepted':
          myLinkmicId.value = msg.data?.linkmicId
          myQueuePosition.value = 0
          ElMessage.success('连麦已接通，正在建立连接...')
          await viewerStartConnection(msg.userId)
          break
        case 'linkmic-rejected':
          ElMessage.warning('连麦被拒绝')
          myLinkmicId.value = null
          myQueuePosition.value = 0
          break
        case 'linkmic-disconnected':
          removeLinkmicStream(msg.data?.viewerId || msg.userId)
          break
        case 'hand-raised':
          // 主播看到观众的举手状态
          if (isStreamer.value && msg.data) {
            const viewerId = msg.userId
            const raised = !!msg.data.raised
            // 更新对应申请项的 handRaised 状态
            const app = pendingApplications.value.find(a => a.viewerId === viewerId)
            if (app) app.handRaised = raised
            if (raised) ElMessage.info(`${msg.userName || ('观众' + viewerId)} 举手了`)
          }
          break
        case 'gift':
          if (msg.data?.emoji) {
            const id = Date.now() + Math.random()
            gifts.value.push({ id, emoji: msg.data.emoji, name: msg.data.name, from: msg.userName || '观众' })
            setTimeout(() => { gifts.value = gifts.value.filter(g => g.id !== id) }, 3000)
          }
          break
        case 'offer':
          await handleOffer(msg.userId, msg.data)
          break
        case 'answer':
          await handleAnswer(msg.userId, msg.data)
          break
        case 'ice-candidate':
          await handleIceCandidate(msg.userId, msg.data)
          break
        case 'reaction':
          if (msg.data?.emoji) {
            addReaction({ emoji: msg.data.emoji, x: msg.data.x, y: msg.data.y, from: msg.userName || ('用户' + msg.userId) })
          }
          break
        case 'pin-message':
          if (msg.data?.text) {
            pinnedMessage.value = { text: msg.data.text, from: msg.userName || '主播', time: Date.now() }
          }
          break
        case 'unpin-message':
          pinnedMessage.value = null
          break
        case 'chat':
          if (msg.data?.text) {
            const item = {
              id: Date.now() + Math.random(),
              text: msg.data.text,
              color: msg.data.color || danmakuColors[0],
              from: msg.userName || ('用户' + msg.userId)
            }
            danmakuList.value.push(item)
            if (danmakuList.value.length > 200) danmakuList.value.splice(0, 50)
            // 同时把弹幕推到 artplayer 飘过画面
            emitDanmaku(item.text, item.color)
            // @提及通知（页面不可见时）
            if (isNotificationEnabled() && msg.data.text.includes('@' + me.value.name) && msg.userId !== currentUserId.value) {
              notifyMention(msg.userName || ('用户' + msg.userId), room.value.roomName, msg.data.text)
            }
          }
          break
        case 'mute-target':
          // 主播让指定观众静音
          if (msg.targetUserId === currentUserId.value || msg.data?.targetUserId === currentUserId.value) {
            if (localLinkmicAudioEnabled.value) {
              toggleLocalAudio()
              ElMessage.info('主播已将你静音')
            }
          }
          break
        case 'mute-video-target':
          if (msg.targetUserId === currentUserId.value || msg.data?.targetUserId === currentUserId.value) {
            if (localLinkmicVideoEnabled.value) {
              toggleLocalVideo()
              ElMessage.info('主播已关闭你的摄像头')
            }
          }
          break
        case 'peer-state':
          // 连麦双方音视频状态同步
          if (linkmicStreams[msg.userId]) {
            Object.assign(linkmicStreams[msg.userId], msg.data || {})
          } else {
            linkmicStreams[msg.userId] = { ...(msg.data || {}) }
          }
          break
        case 'viewer-count':
          if (msg.data?.count !== undefined) {
            realtimeViewerCount.value = msg.data.count
          }
          break
      }
    },
    onClose: () => console.log('[Linkmic WS] 断开（将自动重连）')
  })
}

// ============ WebRTC P2P ============
const createPC = (peerId) => {
  if (linkmicPeerConnections[peerId]) return linkmicPeerConnections[peerId]
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  })
  linkmicPeerConnections[peerId] = pc

  pc.ontrack = (event) => {
    linkmicStreams[peerId] = { stream: event.streams[0] }
  }

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      sendSignaling('ice-candidate', peerId, event.candidate)
    }
  }

  pc.onconnectionstatechange = () => {
    console.log('[Linkmic] PC状态', peerId, pc.connectionState)
  }

  pc.oniceconnectionstatechange = async () => {
    if (pc.iceConnectionState === 'failed') {
      console.warn('[Linkmic] ICE failed, 尝试 restart', peerId)
      try {
        // 只有发起方（观众端）能 restart；主播侧由对方触发
        if (localLinkmicStream.value) {
          const offer = await pc.createOffer({ iceRestart: true })
          await pc.setLocalDescription(offer)
          sendSignaling('offer', peerId, offer.sdp)
        }
      } catch (e) {
        console.error('[Linkmic] ICE restart 失败', e)
      }
    }
  }

  return pc
}

// 观众侧：主播同意后，由观众发起 offer（观众有摄像头要推给主播）
const viewerStartConnection = async (streamerId) => {
  try {
    if (!localLinkmicStream.value) {
      localLinkmicStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      await nextTick()
      if (localLinkmicVideoRef.value) {
        localLinkmicVideoRef.value.srcObject = localLinkmicStream.value
      }
    }
    // 如果旧 PC 还在（重新协商场景），先关掉
    if (linkmicPeerConnections[streamerId]) {
      removePC(streamerId)
    }
    const pc = createPC(streamerId)
    localLinkmicStream.value.getTracks().forEach(t => pc.addTrack(t, localLinkmicStream.value))
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    sendSignaling('offer', streamerId, offer.sdp)
  } catch (e) {
    console.error('[Linkmic] viewer start failed', e)
    ElMessage.error('无法获取摄像头/麦克风')
  }
}

const handleOffer = async (peerId, sdp) => {
  // 主播侧收到观众的 offer，旧 PC 还在则替换（重新协商）
  if (linkmicPeerConnections[peerId]) {
    removePC(peerId)
  }
  const pc = createPC(peerId)
  await pc.setRemoteDescription({ type: 'offer', sdp })
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  sendSignaling('answer', peerId, answer.sdp)
}

const handleAnswer = async (peerId, sdp) => {
  const pc = linkmicPeerConnections[peerId]
  if (!pc) return
  await pc.setRemoteDescription({ type: 'answer', sdp })
  if (pendingIce[peerId]) {
    for (const c of pendingIce[peerId]) await pc.addIceCandidate(c)
    delete pendingIce[peerId]
  }
}

const handleIceCandidate = async (peerId, candidate) => {
  const pc = linkmicPeerConnections[peerId]
  if (pc && pc.remoteDescription) {
    try { await pc.addIceCandidate(candidate) } catch (e) { console.warn(e) }
  } else {
    (pendingIce[peerId] = pendingIce[peerId] || []).push(candidate)
  }
}

const sendSignaling = (type, targetUserId, data) => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  linkmicWs.send({
    type,
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    targetUserId,
    data
  })
}

const sendBroadcast = (type, data) => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  linkmicWs.send({
    type,
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    data
  })
}

// ============ 业务动作 ============
const applyLinkmic = async () => {
  if (!currentUserId.value) {
    ElMessage.warning('请先登录')
    return
  }
  linkmicLoading.value = true
  try {
    const res = await linkmicApi.applyLinkmic(roomId)
    if (res.code === 200) {
      myLinkmicId.value = res.data.id
      // 申请通过后查询队列位置
      try {
        const qres = await linkmicApi.getQueuePosition(roomId)
        if (qres.code === 200) myQueuePosition.value = qres.data || 0
      } catch (e) {}
      if (myQueuePosition.value > 0) {
        ElMessage.success('连麦申请已发送，当前排队第 ' + myQueuePosition.value + ' 位')
      } else {
        ElMessage.success('连麦申请已发送，请等待主播同意')
      }
      sendBroadcast('linkmic-apply', res.data)
    } else {
      ElMessage.error(res.message || '申请失败')
    }
  } catch (e) {
    ElMessage.error('申请连麦失败')
  } finally {
    linkmicLoading.value = false
  }
}

const acceptLinkmic = async (linkmic) => {
  try {
    await linkmicApi.acceptLinkmic(linkmic.id)
    linkmic.status = 1
    activeLinkmics.value.push(linkmic)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)
    // 定向告诉这个观众："你被同意了"，观众收到后会发 offer
    sendSignaling('linkmic-accepted', linkmic.viewerId, { linkmicId: linkmic.id })
    ElMessage.success('已同意连麦')
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const rejectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.rejectLinkmic(linkmic.id)
    pendingApplications.value = pendingApplications.value.filter(a => a.id !== linkmic.id)
    sendSignaling('linkmic-rejected', linkmic.viewerId, {})
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const disconnectLinkmic = async (linkmic) => {
  try {
    await linkmicApi.disconnectLinkmic(linkmic.id)
    removeLinkmicStream(linkmic.viewerId)
    activeLinkmics.value = activeLinkmics.value.filter(l => l.id !== linkmic.id)
    sendSignaling('linkmic-disconnected', linkmic.viewerId, { viewerId: linkmic.viewerId })
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const viewerHangup = async () => {
  if (!myLinkmicId.value) return
  try {
    await linkmicApi.disconnectLinkmic(myLinkmicId.value)
  } catch (e) {}
  if (room.value?.userId) {
    sendSignaling('linkmic-disconnected', room.value.userId, { viewerId: currentUserId.value })
  }
  myLinkmicId.value = null
  myQueuePosition.value = 0
  if (localLinkmicStream.value) {
    localLinkmicStream.value.getTracks().forEach(t => t.stop())
    localLinkmicStream.value = null
  }
  // 关掉跟主播的 PC
  if (room.value?.userId) {
    removePC(room.value.userId)
  }
  ElMessage.info('已断开连麦')
}

const toggleViewerHandRaise = () => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  viewerHandRaised.value = !viewerHandRaised.value
  sendSignaling('hand-raised', room.value?.userId, { raised: viewerHandRaised.value })
  ElMessage.info(viewerHandRaised.value ? '已举手，请等待主播接受' : '已放下手')
}

const removePC = (peerId) => {
  const pc = linkmicPeerConnections[peerId]
  if (pc) { pc.close(); delete linkmicPeerConnections[peerId] }
  delete linkmicStreams[peerId]
  delete pendingIce[peerId]
}

const removeLinkmicStream = (viewerId) => {
  removePC(viewerId)
}

const cleanupLinkmic = () => {
  Object.keys(linkmicPeerConnections).forEach(removePC)
  if (localLinkmicStream.value) {
    localLinkmicStream.value.getTracks().forEach(t => t.stop())
    localLinkmicStream.value = null
  }
  if (linkmicWs) {
    try { linkmicWs.close() } catch (e) {}
    linkmicWs = null
  }
}

const setRemoteVideo = (peerId, el) => {
  if (!el) return
  const peer = linkmicStreams[peerId]
  if (peer && peer.stream && el.srcObject !== peer.stream) {
    el.srcObject = peer.stream
  }
}

// ============ 连麦音视频控制 ============
const toggleLocalAudio = async () => {
  if (!localLinkmicStream.value) return
  const track = localLinkmicStream.value.getAudioTracks()[0]
  if (!track) return
  track.enabled = !track.enabled
  localLinkmicAudioEnabled.value = track.enabled
  // 同步状态给房间内（主播能看到）
  sendBroadcast('peer-state', {
    audioEnabled: localLinkmicAudioEnabled.value,
    videoEnabled: localLinkmicVideoEnabled.value
  })
  if (myLinkmicId.value) {
    try { await linkmicApi.toggleAudio(myLinkmicId.value, track.enabled) } catch (e) {}
  }
}

const toggleLocalVideo = async () => {
  if (!localLinkmicStream.value) return
  const track = localLinkmicStream.value.getVideoTracks()[0]
  if (!track) return
  track.enabled = !track.enabled
  localLinkmicVideoEnabled.value = track.enabled
  sendBroadcast('peer-state', {
    audioEnabled: localLinkmicAudioEnabled.value,
    videoEnabled: localLinkmicVideoEnabled.value
  })
  if (myLinkmicId.value) {
    try { await linkmicApi.toggleVideo(myLinkmicId.value, track.enabled) } catch (e) {}
  }
}

// 主播：让指定连麦观众静音
const muteLinkmicAudio = (linkmic) => {
  sendSignaling('mute-target', linkmic.viewerId, { targetUserId: linkmic.viewerId })
  ElMessage.success('已请求 ' + linkmic.viewerName + ' 静音')
}
const muteLinkmicVideo = (linkmic) => {
  sendSignaling('mute-video-target', linkmic.viewerId, { targetUserId: linkmic.viewerId })
  ElMessage.success('已请求 ' + linkmic.viewerName + ' 关摄像头')
}

// ============ 观众反应 ============
const sendReaction = (emoji) => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  const x = Math.random() * 80 + 10
  const y = Math.random() * 60 + 20
  linkmicWs.send({
    type: 'reaction',
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    data: { emoji, x, y }
  })
  // 自己立即显示
  addReaction({ emoji, x, y, from: me.value.name })
}

const addReaction = (r) => {
  const id = Date.now() + Math.random()
  reactions.value.push({ id, ...r })
  if (reactions.value.length > 20) reactions.value.shift()
  // 动画 1.5s 后移除
  setTimeout(() => {
    reactions.value = reactions.value.filter(r2 => r2.id !== id)
  }, 1500)
}

const showReactionPicker = ref(false)

// ============ 弹幕 ============
const emitDanmaku = (text, color) => {
  if (!art) return
  const plugin = art.plugins?.artplayerPluginDanmuku
  if (plugin && typeof plugin.emit === 'function') {
    try {
      plugin.emit({ text, color, border: false })
    } catch (e) { console.warn('emit danmaku failed', e) }
  }
}

const toggleEmojiPicker = () => { showEmojiPicker.value = !showEmojiPicker.value }
const insertChatEmoji = (e) => { danmakuText.value += e; showEmojiPicker.value = false }
const sendGift = (gift) => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  linkmicWs.send({
    type: 'gift',
    roomCode: room.value?.streamKey,
    userId: currentUserId.value,
    userName: me.value.name,
    data: { emoji: gift.emoji, name: gift.name, cost: gift.cost }
  })
  const id = Date.now() + Math.random()
  gifts.value.push({ id, emoji: gift.emoji, name: gift.name, from: me.value.name })
  setTimeout(() => { gifts.value = gifts.value.filter(g => g.id !== id) }, 3000)
  ElMessage.success(`赠送了 ${gift.name}`)
  showGiftPicker.value = false
}
const pinMessage = (item) => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  linkmicWs.send({ type: 'pin-message', roomCode: room.value?.streamKey, userId: currentUserId.value, userName: me.value.name, data: { text: item.text || item.from } })
  pinnedMessage.value = { text: item.text || item.from, from: item.from || '观众', time: Date.now() }
}
const unpinMessage = () => {
  if (!linkmicWs || !linkmicWs.isOpen()) return
  linkmicWs.send({ type: 'unpin-message', roomCode: room.value?.streamKey, userId: currentUserId.value })
  pinnedMessage.value = null
}

const sendDanmaku = () => {
  const text = danmakuText.value.trim()
  if (!text) return
  const color = danmakuColors[Math.floor(Math.random() * danmakuColors.length)]
  // 通过 ws 广播，其他观众和自己都会收到 'chat' 推送回来再渲染
  if (linkmicWs && linkmicWs.isOpen()) {
    linkmicWs.send({
      type: 'chat',
      roomCode: room.value?.streamKey,
      userId: currentUserId.value,
      userName: me.value.name,
      data: { text, color }
    })
    // 自己立即看见——侧边聊天 + 画面飘过
    danmakuList.value.push({
      id: Date.now(),
      text,
      color,
      from: me.value.name
    })
    emitDanmaku(text, color)
  } else {
    danmakuList.value.push({
      id: Date.now(),
      text,
      color,
      from: me.value.name
    })
    emitDanmaku(text, color)
  }
  danmakuText.value = ''
}
</script>

<template>
  <div class="live-room-page">
    <template v-if="loading">
      <div class="loading-state">
        <span class="buffering-icon" />
        <span style="color:#fff;font-size:14px;">加载中...</span>
      </div>
    </template>
    <template v-else-if="room">
      <div class="room-layout">
        <div class="video-section">
          <div class="video-wrapper" :class="{ buffering: isBuffering }">
            <div ref="playerRef" class="live-player"></div>
            <!-- 观众反应动画层 -->
            <div class="reaction-overlay">
              <transition-group name="reaction-anim">
                <span
                  v-for="r in reactions"
                  :key="r.id"
                  class="reaction-burst"
                  :style="{ left: r.x + '%', top: r.y + '%' }"
                >{{ r.emoji }}</span>
              </transition-group>
            </div>
            <!-- 礼物动画 -->
            <div class="gift-overlay">
              <transition-group name="gift-anim">
                <div v-for="g in gifts" :key="g.id" class="gift-burst">
                  <span class="gift-emoji">{{ g.emoji }}</span>
                  <span class="gift-text">{{ g.from }} 赠送了 {{ g.name }}</span>
                </div>
              </transition-group>
            </div>
            <transition name="fade">
              <div v-if="isBuffering" class="buffering-overlay">
                <span class="buffering-icon" />
                <span>缓冲中...</span>
              </div>
            </transition>
            <transition name="fade">
              <div v-if="room.status === 'offline'" class="stream-ended-overlay">
                <div v-if="currentReplay" class="replay-player-wrap">
                  <div class="replay-header">
                    <span class="replay-title">{{ currentReplay.title }}</span>
                    <el-button size="small" @click="closeReplay">返回列表</el-button>
                  </div>
                  <div ref="replayPlayerRef" class="replay-player"></div>
                </div>
                <div v-else class="ended-content">
                  <div class="ended-tabs">
                    <span :class="['ended-tab', { active: !showReplayTab }]" @click="showReplayTab = false">
                      主播已离开
                    </span>
                    <span :class="['ended-tab', { active: showReplayTab }]" @click="showReplayTab = true; loadReplays()">
                      直播回放
                      <span v-if="replayList.length" class="replay-count">{{ replayList.length }}</span>
                    </span>
                  </div>
                  <div v-if="!showReplayTab" class="ended-default">
                    <div class="ended-icon">🎮</div>
                    <h3>直播已结束</h3>
                    <p>主播已下播，感谢观看！</p>
                    <el-button type="primary" @click="$router.push('/live')">返回直播间列表</el-button>
                  </div>
                  <div v-else class="ended-replays">
                    <div v-if="replayList.length === 0" class="replay-empty">
                      <p>暂无回放内容</p>
                      <el-button type="primary" @click="$router.push('/live')">返回直播间列表</el-button>
                    </div>
                    <div v-else class="replay-grid">
                      <div v-for="r in replayList" :key="r.id" class="replay-item" @click="playReplay(r)">
                        <div class="replay-cover">
                          <img :src="r.coverUrl || '/live-placeholder.svg'" alt="cover" />
                          <span class="replay-duration">{{ r.duration }}</span>
                        </div>
                        <div class="replay-info">
                          <div class="replay-name">{{ r.title }}</div>
                          <div class="replay-time">{{ r.createTime }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </transition>
            <div class="video-info">
              <div class="room-title">{{ room.roomName }}</div>
            <div class="room-meta">
                <span class="live-status-dot" />
                <span>直播中</span>
                <span class="viewer">{{ realtimeViewerCount || room.viewerCount || 0 }} 人观看</span>
                <el-tag v-if="currentQuality !== '自动'" size="small" type="warning">{{ currentQuality }}</el-tag>
                <el-button size="small" circle @click="shareRoom" title="分享直播间">
                  <el-icon><Share /></el-icon>
                </el-button>
                <el-button v-if="!isStreamer" size="small" :type="isFollowing ? 'warning' : ''" circle @click="toggleFollow" :title="isFollowing ? '取消关注' : '关注主播'">
                  {{ isFollowing ? '❤️' : '🤍' }}
                </el-button>
              </div>
            </div>

            <!-- 连麦视频区域（显示其他连麦者） -->
            <div v-if="activeLinkmics.length > 0 || (myLinkmicId && !isStreamer)" class="linkmic-videos">
              <!-- 观众端本地预览 -->
              <div v-if="!isStreamer && myLinkmicId" class="linkmic-video-card">
                <video ref="localLinkmicVideoRef" autoplay playsinline muted />
                <div class="linkmic-name">我（连麦中）</div>
                <div class="linkmic-self-controls">
                  <el-button size="small" :type="localLinkmicAudioEnabled ? '' : 'danger'" circle @click="toggleLocalAudio">
                    <el-icon><Microphone v-if="localLinkmicAudioEnabled" /><Mute v-else /></el-icon>
                  </el-button>
                  <el-button size="small" :type="localLinkmicVideoEnabled ? '' : 'danger'" circle @click="toggleLocalVideo">
                    <el-icon><VideoCamera v-if="localLinkmicVideoEnabled" /><VideoPause v-else /></el-icon>
                  </el-button>
                </div>
              </div>
              <!-- 远端连麦者画面 -->
              <div v-for="linkmic in activeLinkmics" :key="linkmic.id" class="linkmic-video-card">
                <video :ref="el => setRemoteVideo(linkmic.viewerId, el)" autoplay playsinline />
                <div class="linkmic-name">
                  {{ linkmic.viewerName }}
                  <el-icon v-if="linkmicStreams[linkmic.viewerId]?.audioEnabled === false" class="state-off"><Mute /></el-icon>
                  <el-icon v-if="linkmicStreams[linkmic.viewerId]?.videoEnabled === false" class="state-off"><VideoPause /></el-icon>
                </div>
                <div v-if="isStreamer" class="linkmic-actions">
                  <el-button size="small" link @click="muteLinkmicAudio(linkmic)" title="远程静音">
                    <el-icon><Mute /></el-icon>
                  </el-button>
                  <el-button size="small" link @click="muteLinkmicVideo(linkmic)" title="远程关摄像头">
                    <el-icon><VideoPause /></el-icon>
                  </el-button>
                  <el-button size="small" type="danger" @click="disconnectLinkmic(linkmic)">
                    断开
                  </el-button>
                </div>
              </div>
            </div>
          </div>

          <!-- 观众端连麦按钮 -->
          <div v-if="!isStreamer && !myLinkmicId" class="linkmic-apply-bar">
            <el-button type="primary" size="large" round @click="applyLinkmic" :loading="linkmicLoading">
              <el-icon><Phone /></el-icon>
              申请连麦
            </el-button>
            <el-button v-if="!viewerHandRaised" size="large" round plain @click="toggleViewerHandRaise" style="margin-left:8px">
              ✋ 举手发言
            </el-button>
            <el-tag v-else type="warning" size="large" style="margin-left:8px">
              ✋ 举手中...
            </el-tag>
          </div>
          <div v-else-if="!isStreamer && myLinkmicId && myQueuePosition > 0" class="linkmic-apply-bar">
            <el-tag type="warning" effect="dark" size="large">
              排队中，当前第 {{ myQueuePosition }} 位
            </el-tag>
            <el-button type="danger" size="small" round @click="viewerHangup" style="margin-left:8px">
              取消
            </el-button>
          </div>
          <div v-else-if="!isStreamer && myLinkmicId" class="linkmic-apply-bar">
            <el-button type="danger" size="large" round @click="viewerHangup">
              <el-icon><PhoneFilled /></el-icon>
              挂断连麦
            </el-button>
          </div>

          <!-- 主播端连麦管理 -->
          <div v-if="isStreamer && pendingApplications.length > 0" class="linkmic-pending-bar">
            <div class="pending-title">连麦申请 ({{ pendingApplications.length }})</div>
            <div class="pending-list">
              <div v-for="(app, idx) in pendingApplications" :key="app.id" class="pending-item">
                <span class="queue-no">{{ idx + 1 }}.</span>
                <span>{{ app.viewerName }}</span>
                <el-tag v-if="app.handRaised" size="small" type="warning">✋ 举手</el-tag>
                <el-button size="small" type="success" @click="acceptLinkmic(app)">同意</el-button>
                <el-button size="small" type="danger" @click="rejectLinkmic(app)">拒绝</el-button>
              </div>
            </div>
          </div>
        </div>
        <div class="chat-section">
          <div class="chat-header">互动聊天</div>
          <!-- 反应按钮 -->
          <div class="reaction-bar">
            <span
              v-for="e in reactionEmojis"
              :key="e"
              class="reaction-btn"
              @click="sendReaction(e)"
              :title="'发送 ' + e"
            >{{ e }}</span>
            <el-button size="small" round @click="showGiftPicker = !showGiftPicker" title="送礼物">🎁</el-button>
            <transition name="fade">
              <div v-if="showGiftPicker" class="gift-picker">
                <div v-for="g in giftOptions" :key="g.name" class="gift-option" @click="sendGift(g)">
                  <span class="gift-opt-emoji">{{ g.emoji }}</span>
                  <span class="gift-opt-name">{{ g.name }}</span>
                  <span class="gift-opt-cost">{{ g.cost }}瓜子</span>
                </div>
              </div>
            </transition>
          </div>
          <div v-if="pinnedMessage" class="pinned-msg">
              <span class="pinned-label">📌 置顶</span>
              <span class="pinned-text">{{ pinnedMessage.from }}: {{ pinnedMessage.text }}</span>
              <el-button v-if="isStreamer" size="small" link type="danger" @click="unpinMessage">取消</el-button>
            </div>
          <div class="chat-messages">
            <div v-for="d in danmakuList" :key="d.id" class="chat-msg" :class="{ pinned: pinnedMessage?.text === d.text }">
              <span class="chat-name">{{ d.from || '观众' }}: </span>
              <span>{{ d.text }}</span>
              <el-button v-if="isStreamer" size="small" link @click="pinMessage(d)" title="置顶">📌</el-button>
            </div>
          </div>
          <div class="chat-input">
            <div class="chat-emoji-wrap">
              <el-button circle @click="toggleEmojiPicker" title="表情">😀</el-button>
              <transition name="fade">
                <div v-if="showEmojiPicker" class="chat-emoji-picker">
                  <span v-for="e in chatEmojis" :key="e" class="chat-emoji-item" @click="insertChatEmoji(e)">{{ e }}</span>
                </div>
              </transition>
            </div>
            <el-input
              v-model="danmakuText"
              placeholder="发个弹幕~"
              @keydown.enter="sendDanmaku"
            />
            <el-button type="primary" @click="sendDanmaku">发送</el-button>
          </div>
        </div>
      </div>
    </template>
    <div v-else class="empty-state">
      <el-empty description="直播间不存在" />
    </div>
  </div>
</template>

<style scoped>
.live-room-page { height: calc(100vh - 64px); background: #0a0a0a; }
.room-layout { display: flex; height: 100%; }
.video-section { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.video-wrapper { flex: 1; position: relative; background: #000; }
.video-wrapper.buffering .live-player { opacity: 0.5; }
.buffering-overlay {
  position: absolute; inset: 0;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: #fff; gap: 12px; font-size: 14px;
  pointer-events: none;
  z-index: 5;
}
.buffering-icon {
  width: 36px; height: 36px; border: 3px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.live-player {
  width: 100%;
  height: 100%;
  background: #000;
}
/* 观众反应动画 */
.reaction-overlay {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  z-index: 9;
}
.reaction-burst {
  position: absolute;
  font-size: 32px;
  animation: reaction-float 1.5s ease-out forwards;
  transform: translate(-50%, -50%);
}
@keyframes reaction-float {
  0% { opacity: 1; transform: translate(-50%, -50%) scale(0.5); }
  50% { opacity: 1; }
  100% { opacity: 0; transform: translate(-50%, -150%) scale(1.2); }
}
.reaction-anim-enter-active { animation: reaction-float 1.5s ease-out; }
.reaction-anim-leave-active { display: none; }
.stream-ended-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0,0,0,0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20;
}
.ended-content {
  text-align: center;
  color: #fff;
}
.ended-icon { font-size: 64px; margin-bottom: 16px; }
.ended-content h3 { font-size: 24px; margin: 0 0 12px; }
.ended-content p { font-size: 14px; color: #9499a0; margin: 0 0 24px; }
.ended-tabs {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 24px;
}
.ended-tab {
  padding: 6px 20px;
  border-radius: 20px;
  font-size: 14px;
  color: #9499a0;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}
.ended-tab:hover { color: #fff; }
.ended-tab.active { background: #fb7299; color: #fff; font-weight: 500; }
.replay-count {
  background: rgba(255,255,255,0.2);
  border-radius: 10px;
  padding: 0 6px;
  font-size: 11px;
}
.ended-replays { width: 100%; max-width: 600px; margin: 0 auto; }
.replay-empty { padding: 40px 0; }
.replay-empty p { color: #9499a0; margin-bottom: 16px; }
.replay-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; }
.replay-item {
  background: rgba(255,255,255,0.1);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: background 0.2s;
}
.replay-item:hover { background: rgba(255,255,255,0.2); }
.replay-cover { position: relative; aspect-ratio: 16/9; }
.replay-cover img { width: 100%; height: 100%; object-fit: cover; }
.replay-duration {
  position: absolute; bottom: 6px; right: 6px;
  background: rgba(0,0,0,0.7); color: #fff;
  font-size: 11px; padding: 2px 6px; border-radius: 4px;
}
.replay-info { padding: 10px 12px; }
.replay-name { font-size: 13px; color: #fff; margin-bottom: 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.replay-time { font-size: 11px; color: #9499a0; }
.replay-player-wrap { width: 100%; }
.replay-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.replay-title { font-size: 16px; color: #fff; font-weight: 500; }
.replay-player { width: 100%; aspect-ratio: 16/9; background: #000; border-radius: 8px; overflow: hidden; }
.video-info {
  position: absolute; top: 0; left: 0; right: 0;
  background: linear-gradient(rgba(0,0,0,0.7), transparent);
  padding: 16px 20px 40px;
  pointer-events: none;
  z-index: 8;
}
.room-title { font-size: 18px; font-weight: 600; color: #fff; margin-bottom: 8px; }
.room-meta { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #ccc; }
.live-status-dot { width: 8px; height: 8px; background: #f04040; border-radius: 50%; animation: pulse 1.5s infinite; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.viewer { color: #9499a0; }

/* 连麦视频 */
.linkmic-videos {
  position: absolute;
  top: 60px;
  right: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  z-index: 10;
}
.linkmic-video-card {
  width: 160px;
  height: 120px;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}
.linkmic-video-card video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.linkmic-name {
  position: absolute;
  bottom: 4px;
  left: 8px;
  font-size: 12px;
  color: #fff;
}
.linkmic-actions {
  position: absolute;
  top: 4px;
  right: 4px;
  display: flex;
  gap: 2px;
}
.linkmic-self-controls {
  position: absolute;
  bottom: 4px;
  right: 4px;
  display: flex;
  gap: 4px;
}
.linkmic-name .state-off {
  margin-left: 4px;
  color: #ff5252;
  font-size: 12px;
}

/* 连麦申请栏 */
.linkmic-apply-bar {
  position: absolute;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.linkmic-pending-bar {
  background: rgba(0,0,0,0.8);
  padding: 12px 16px;
  border-top: 1px solid #333;
}
.pending-title {
  color: #fff;
  font-size: 14px;
  margin-bottom: 8px;
}
.pending-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pending-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 13px;
}
.queue-no {
  color: #fad34a;
  font-weight: 600;
  min-width: 18px;
}

.chat-section { width: 320px; background: #fff; display: flex; flex-direction: column; border-left: 1px solid #e3e5e7; }
.chat-header { padding: 16px; font-weight: 600; border-bottom: 1px solid #e3e5e7; }
.chat-messages { flex: 1; overflow-y: auto; padding: 12px; }
.chat-msg { margin-bottom: 8px; font-size: 14px; }
.chat-name { color: #00a1d6; }
.chat-input { display: flex; gap: 8px; padding: 12px; border-top: 1px solid #e3e5e7; }
.chat-input .el-input { flex: 1; }
.reaction-bar {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid #f6f7f8;
  flex-wrap: wrap;
  position: relative;
}
.reaction-btn {
  font-size: 18px;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  transition: background 0.15s;
}
.reaction-btn:hover { background: #f6f7f8; }
/* 礼物 */
.gift-overlay {
  position: absolute;
  top: 60px;
  right: 20px;
  width: 220px;
  pointer-events: none;
  z-index: 15;
}
.gift-burst {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  margin-bottom: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.15);
}
.gift-emoji { font-size: 24px; }
.gift-text { font-size: 12px; color: #856404; }
.gift-anim-enter-active { animation: gift-in 0.3s ease-out; }
.gift-anim-leave-active { display: none; }
@keyframes gift-in {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.gift-picker {
  position: absolute;
  bottom: 40px;
  left: 0;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 8px;
  width: 200px;
  z-index: 20;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
}
.gift-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 6px;
}
.gift-option:hover { background: #f6f7f8; }
.gift-opt-emoji { font-size: 20px; }
.gift-opt-name { flex: 1; font-size: 13px; color: #18191c; }
.gift-opt-cost { font-size: 11px; color: #9499a0; }
.chat-emoji-wrap { position: relative; }
.chat-emoji-picker {
  position: absolute;
  bottom: 40px;
  left: 0;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 8px;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 2px;
  width: 220px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 10;
}
.chat-emoji-item { font-size: 18px; cursor: pointer; padding: 4px; border-radius: 4px; }
.chat-emoji-item:hover { background: #f6f7f8; }
.pinned-msg {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff3cd;
  border-bottom: 1px solid #ffeeba;
  font-size: 13px;
}
.pinned-label { font-size: 12px; color: #856404; white-space: nowrap; }
.pinned-text { flex: 1; color: #856404; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chat-msg.pinned { background: #fff9e6; }
.loading-state { display: flex; align-items: center; justify-content: center; height: 60vh; }
.empty-state { display: flex; align-items: center; justify-content: center; height: 60vh; background: #0a0a0a; }
</style>