import { computed, nextTick, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sameUserId } from '../utils/userId.js'

export const formatMeetingDuration = (seconds) => {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const isTextInputTarget = target => {
  const tagName = target?.tagName
  return tagName === 'INPUT' || tagName === 'TEXTAREA' || target?.isContentEditable
}

export function useMeetingSessionUi({
  me,
  currentRoom,
  hostUserId,
  remotePeers,
  localVideoRef,
  audioEnabled,
  videoEnabled,
  screenShareEnabled,
  sendMeetingMessage,
  broadcastSelfState,
  toggleAudio,
  toggleVideo,
  startScreenShare,
  documentRef = typeof document !== 'undefined' ? document : null,
  addEventListenerFn = (type, handler) => window.addEventListener(type, handler),
  removeEventListenerFn = (type, handler) => window.removeEventListener(type, handler),
  setTimeoutFn = (handler, delay) => setTimeout(handler, delay),
  clearTimeoutFn = timerId => clearTimeout(timerId),
  setIntervalFn = (handler, delay) => setInterval(handler, delay),
  clearIntervalFn = timerId => clearInterval(timerId),
  nextTickFn = nextTick,
  hideControlsDelayMs = 3000,
  logger = console
}) {
  const sidebarOpen = ref(true)
  const controlsVisible = ref(true)
  const handRaised = ref(false)
  const waitingUsers = ref([])
  const waitingForHost = ref(false)
  const editingName = ref(false)
  const editNameDraft = ref(me.value.name)
  const editNameInput = ref(null)
  const meetingDuration = ref(0)
  const spotlightId = ref(null)
  let hideControlsTimer = null
  let meetingTimer = null
  let keyboardRegistered = false

  const remotePeerIds = computed(() => Object.keys(remotePeers))
  const isHost = computed(() => currentRoom.value && sameUserId(me.value.id, hostUserId.value))
  const participantList = computed(() => {
    const list = [{
        userId: me.value.id,
        name: `${me.value.name}（我）`,
        audioEnabled: audioEnabled.value,
        videoEnabled: videoEnabled.value,
        screenShareEnabled: screenShareEnabled.value,
        isHost: currentRoom.value && sameUserId(me.value.id, hostUserId.value),
        self: true
      }]
    for (const id of Object.keys(remotePeers)) {
      const p = remotePeers[id]
      list.push({
        userId: Number(id),
        name: p.name || (`用户${id}`),
        audioEnabled: p.audioEnabled !== false,
        videoEnabled: p.videoEnabled !== false,
        screenShareEnabled: !!p.screenShareEnabled,
        handRaised: !!p.handRaised,
        isHost: currentRoom.value && sameUserId(id, hostUserId.value),
        self: false
      })
    }
    return list
  })
  const totalHandRaised = computed(() => {
    return participantList.value.filter(p => p.handRaised && !p.self).length
  })
  const screenSharerId = computed(() => {
    const entry = Object.entries(remotePeers).find(([, p]) => p.screenShareEnabled)
    return entry ? entry[0] : null
  })

  const clearHideControlsTimer = () => {
    if (!hideControlsTimer) return false
    clearTimeoutFn(hideControlsTimer)
    hideControlsTimer = null
    return true
  }

  const showControls = () => {
    controlsVisible.value = true
    clearHideControlsTimer()
    hideControlsTimer = setTimeoutFn(() => {
      controlsVisible.value = false
      hideControlsTimer = null
    }, hideControlsDelayMs)
  }

  const stopMeetingTimer = () => {
    if (!meetingTimer) return false
    clearIntervalFn(meetingTimer)
    meetingTimer = null
    return true
  }

  const restartMeetingTimer = () => {
    stopMeetingTimer()
    meetingDuration.value = 0
    meetingTimer = setIntervalFn(() => {
      meetingDuration.value++
    }, 1000)
    return true
  }

  const exitFullscreen = () => {
    if (documentRef?.fullscreenElement) {
      documentRef.exitFullscreen?.()
      return true
    }
    return false
  }

  const handleKeydown = (event) => {
    if (!currentRoom.value) return false
    if (isTextInputTarget(event.target)) return false
    switch (event.key) {
      case 'Escape':
        return exitFullscreen()
      case 'm':
      case 'M':
        return toggleAudio()
      case 'v':
      case 'V':
        return toggleVideo()
      case 's':
      case 'S':
        return startScreenShare()
      case 'h':
      case 'H':
        return toggleHandRaise()
      default:
        return false
    }
  }

  const registerKeyboardShortcuts = () => {
    if (keyboardRegistered) return false
    addEventListenerFn('keydown', handleKeydown)
    keyboardRegistered = true
    return true
  }

  const unregisterKeyboardShortcuts = () => {
    if (!keyboardRegistered) return false
    removeEventListenerFn('keydown', handleKeydown)
    keyboardRegistered = false
    return true
  }

  const setRemoteVideo = (peerId, el) => {
    if (!el) return false
    const peer = remotePeers[peerId]
    if (peer?.stream && el.srcObject !== peer.stream) {
      el.srcObject = peer.stream
      return true
    }
    return false
  }

  const togglePiP = async (peerId) => {
    let videoEl = null
    if (peerId === 'local') {
      videoEl = localVideoRef.value
    } else {
      const idx = remotePeerIds.value.indexOf(peerId)
      if (idx >= 0) {
        const cards = documentRef?.querySelectorAll?.('.video-card video') || []
        videoEl = cards[idx + 1]
      }
    }
    if (!videoEl) return false
    try {
      if (documentRef?.pictureInPictureElement) {
        await documentRef.exitPictureInPicture()
        return true
      }
      if (documentRef?.pictureInPictureEnabled) {
        await videoEl.requestPictureInPicture()
        return true
      }
      ElMessage.warning('当前浏览器不支持画中画')
      return false
    } catch (e) {
      logger.warn?.('PiP error:', e)
      return false
    }
  }

  const handleVideoDoubleClick = (peerId) => {
    return togglePiP(peerId)
  }

  const startEditName = () => {
    editNameDraft.value = me.value.name
    editingName.value = true
    nextTickFn(() => editNameInput.value?.select())
  }

  const saveName = () => {
    const newName = editNameDraft.value.trim()
    if (!newName) {
      editNameDraft.value = me.value.name
      editingName.value = false
      return false
    }
    me.value.name = newName
    editNameDraft.value = newName
    broadcastSelfState()
    ElMessage.success('昵称已更新')
    editingName.value = false
    return true
  }

  const toggleHandRaise = () => {
    const nextRaised = !handRaised.value
    if (!sendMeetingMessage({ type: 'hand-raised', data: { raised: nextRaised } })) return false
    handRaised.value = nextRaised
    ElMessage.info(handRaised.value ? '已举手，请等待主持人允许' : '已放下手')
    return true
  }

  const cleanupMeetingSessionUi = () => {
    stopMeetingTimer()
    clearHideControlsTimer()
    meetingDuration.value = 0
    controlsVisible.value = true
    handRaised.value = false
    waitingUsers.value = []
    waitingForHost.value = false
    editingName.value = false
    spotlightId.value = null
  }

  return {
    sidebarOpen,
    controlsVisible,
    handRaised,
    waitingUsers,
    waitingForHost,
    editingName,
    editNameDraft,
    editNameInput,
    meetingDuration,
    spotlightId,
    remotePeerIds,
    isHost,
    participantList,
    totalHandRaised,
    screenSharerId,
    formatDuration: formatMeetingDuration,
    showControls,
    restartMeetingTimer,
    stopMeetingTimer,
    handleKeydown,
    registerKeyboardShortcuts,
    unregisterKeyboardShortcuts,
    setRemoteVideo,
    togglePiP,
    handleVideoDoubleClick,
    startEditName,
    saveName,
    toggleHandRaise,
    cleanupMeetingSessionUi
  }
}
