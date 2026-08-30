import { nextTick } from 'vue'
import { ElMessage } from 'element-plus'

const getFirstTrack = (stream, getterName) => {
  return stream?.[getterName]?.()[0] || null
}

const stopStreamTracks = (stream) => {
  if (!stream?.getTracks) return
  stream.getTracks().forEach(track => {
    if ('onended' in track) track.onended = null
    if (track.readyState !== 'ended') track.stop?.()
  })
}

export function useMeetingMediaControls({
  localStream,
  localVideoRef,
  audioEnabled,
  videoEnabled,
  screenShareEnabled,
  virtualBgEnabled,
  virtualBgStream,
  replaceVideoTrack,
  broadcastSelfState,
  updateAudioActivity,
  captureUserMedia = constraints => navigator.mediaDevices.getUserMedia(constraints),
  captureDisplayMedia = constraints => navigator.mediaDevices.getDisplayMedia(constraints),
  createElement = tag => document.createElement(tag),
  createMediaStream = () => new MediaStream(),
  requestFrame = callback => requestAnimationFrame(callback),
  cancelFrame = id => cancelAnimationFrame(id),
  logger = console
}) {
  let virtualBgRafId = null
  let screenStream = null

  const startLocalStream = async () => {
    try {
      localStream.value = await captureUserMedia({ video: true, audio: true })
      audioEnabled.value = getFirstTrack(localStream.value, 'getAudioTracks')?.enabled !== false
      videoEnabled.value = getFirstTrack(localStream.value, 'getVideoTracks')?.enabled !== false
      await nextTick()
      if (localVideoRef.value) {
        localVideoRef.value.srcObject = localStream.value
      }
      return true
    } catch (e) {
      logger.error?.(e)
      ElMessage.error('无法获取摄像头/麦克风')
      return false
    }
  }

  const toggleAudio = () => {
    const track = getFirstTrack(localStream.value, 'getAudioTracks')
    if (!track) return false
    track.enabled = !track.enabled
    audioEnabled.value = track.enabled
    broadcastSelfState()
    updateAudioActivity?.('local', track.enabled)
    return true
  }

  const toggleVideo = () => {
    const track = getFirstTrack(localStream.value, 'getVideoTracks')
    if (!track) return false
    track.enabled = !track.enabled
    videoEnabled.value = track.enabled
    broadcastSelfState()
    return true
  }

  const stopScreenShare = ({ broadcast = true } = {}) => {
    const wasEnabled = screenShareEnabled.value
    stopStreamTracks(screenStream)
    screenStream = null

    const cameraTrack = getFirstTrack(localStream.value, 'getVideoTracks')
    if (cameraTrack) {
      replaceVideoTrack(cameraTrack)
      if (localVideoRef.value) localVideoRef.value.srcObject = localStream.value
    }

    screenShareEnabled.value = false
    if (broadcast && wasEnabled) broadcastSelfState()
    return wasEnabled
  }

  const startScreenShare = async () => {
    if (screenShareEnabled.value) return stopScreenShare()
    try {
      screenStream = await captureDisplayMedia({ video: true, audio: false })
      const screenTrack = getFirstTrack(screenStream, 'getVideoTracks')
      if (!screenTrack) {
        stopStreamTracks(screenStream)
        screenStream = null
        ElMessage.error('屏幕共享失败')
        return false
      }
      if (virtualBgEnabled.value) stopVirtualBg({ restoreCamera: false })
      replaceVideoTrack(screenTrack)
      if (localVideoRef.value) localVideoRef.value.srcObject = screenStream
      screenShareEnabled.value = true
      broadcastSelfState()
      screenTrack.onended = () => stopScreenShare()
      return true
    } catch (e) {
      if (e.name !== 'NotAllowedError') ElMessage.error('屏幕共享失败')
      return false
    }
  }

  const applyVirtualBg = async () => {
    const videoTrack = getFirstTrack(localStream.value, 'getVideoTracks')
    if (!videoTrack) return false
    if (screenShareEnabled.value) stopScreenShare()
    try {
      const inputCanvas = createElement('canvas')
      inputCanvas.width = 640
      inputCanvas.height = 480
      const inputCtx = inputCanvas.getContext('2d')
      const blurCanvas = createElement('canvas')
      blurCanvas.width = 640
      blurCanvas.height = 480
      const blurCtx = blurCanvas.getContext('2d')
      blurCtx.filter = 'blur(14px)'
      blurCtx.drawImage(inputCanvas, 0, 0, 640, 480)

      const hiddenVideo = createElement('video')
      hiddenVideo.srcObject = localStream.value
      hiddenVideo.autoplay = true
      hiddenVideo.muted = true
      hiddenVideo.play?.()

      const outputStream = createMediaStream()
      const canvasStream = inputCanvas.captureStream(30)
      const canvasTrack = getFirstTrack(canvasStream, 'getVideoTracks')
      if (!canvasTrack) {
        ElMessage.error('虚拟背景开启失败')
        return false
      }
      outputStream.addTrack(canvasTrack)
      virtualBgStream.value = outputStream
      if (localVideoRef.value) localVideoRef.value.srcObject = outputStream
      replaceVideoTrack(canvasTrack)
      virtualBgEnabled.value = true

      const processFrame = () => {
        if (!virtualBgEnabled.value) return
        if (hiddenVideo.readyState >= 2) {
          inputCtx.drawImage(hiddenVideo, 0, 0, 640, 480)
          blurCtx.drawImage(inputCanvas, 0, 0, 640, 480)
          inputCtx.drawImage(blurCanvas, 0, 0, 640, 480)
          const cx = 320
          const cy = 240
          const cw = 200
          const ch = 300
          inputCtx.clearRect(cx - cw / 2, cy - ch / 2, cw, ch)
          inputCtx.drawImage(hiddenVideo, cx - cw / 2, cy - ch / 2, cw, ch, cx - cw / 2, cy - ch / 2, cw, ch)
        }
        virtualBgRafId = requestFrame(processFrame)
      }
      virtualBgRafId = requestFrame(processFrame)

      ElMessage.success('已开启背景模糊')
      return true
    } catch (e) {
      logger.error?.(e)
      virtualBgEnabled.value = false
      stopStreamTracks(virtualBgStream.value)
      virtualBgStream.value = null
      const camTrack = getFirstTrack(localStream.value, 'getVideoTracks')
      if (camTrack) replaceVideoTrack(camTrack)
      if (localVideoRef.value) localVideoRef.value.srcObject = localStream.value
      ElMessage.error('虚拟背景开启失败')
      return false
    }
  }

  const stopVirtualBg = ({ restoreCamera = true } = {}) => {
    const wasEnabled = virtualBgEnabled.value
    virtualBgEnabled.value = false
    if (virtualBgRafId) {
      cancelFrame(virtualBgRafId)
      virtualBgRafId = null
    }
    stopStreamTracks(virtualBgStream.value)
    virtualBgStream.value = null
    if (restoreCamera) {
      const camTrack = getFirstTrack(localStream.value, 'getVideoTracks')
      replaceVideoTrack(camTrack)
      if (localVideoRef.value) localVideoRef.value.srcObject = localStream.value
    }
    return wasEnabled
  }

  const toggleVirtualBg = () => {
    return virtualBgEnabled.value ? stopVirtualBg() : applyVirtualBg()
  }

  const cleanupMedia = () => {
    stopVirtualBg({ restoreCamera: false })
    stopScreenShare({ broadcast: false })
    stopStreamTracks(localStream.value)
    localStream.value = null
    if (localVideoRef.value) localVideoRef.value.srcObject = null
    screenShareEnabled.value = false
    virtualBgEnabled.value = false
  }

  return {
    startLocalStream,
    toggleAudio,
    toggleVideo,
    startScreenShare,
    stopScreenShare,
    applyVirtualBg,
    stopVirtualBg,
    toggleVirtualBg,
    cleanupMedia
  }
}
