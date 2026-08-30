import { nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const DEFAULT_ICE_SERVERS = [{ urls: 'stun:stun.l.google.com:19302' }]

const getCaptureStream = (type) => {
  if (type === 'screen') {
    return navigator.mediaDevices.getDisplayMedia({ video: true, audio: false })
  }
  return navigator.mediaDevices.getUserMedia({ video: true, audio: false })
}

export function useWhipPublisher({
  streamKey,
  onPublished,
  iceServers = DEFAULT_ICE_SERVERS,
  captureStream = getCaptureStream,
  getWhipUrl = key => `http://${location.hostname}:19854/rtc/v1/whip/stream/live/${key}`,
  logger = console
}) {
  const sourceType = ref('camera')
  const mediaStream = ref(null)
  const isPublishing = ref(false)
  const videoPreviewRef = ref(null)
  let peerConnection = null

  const stopPublishing = () => {
    if (videoPreviewRef.value) {
      videoPreviewRef.value.srcObject = null
    }
    if (peerConnection) {
      peerConnection.close()
      peerConnection = null
    }
    if (mediaStream.value) {
      mediaStream.value.getTracks().forEach(track => track.stop())
      mediaStream.value = null
    }
    isPublishing.value = false
  }

  const startPublishing = async () => {
    if (isPublishing.value) return
    if (!streamKey.value) {
      ElMessage.warning('缺少直播流密钥')
      return
    }

    try {
      isPublishing.value = true
      const stream = await captureStream(sourceType.value)
      mediaStream.value = stream
      await nextTick()
      if (videoPreviewRef.value) {
        videoPreviewRef.value.srcObject = stream
      }

      peerConnection = new RTCPeerConnection({ iceServers })
      stream.getTracks().forEach(track => peerConnection.addTrack(track, stream))

      const offer = await peerConnection.createOffer()
      await peerConnection.setLocalDescription(offer)
      const res = await fetch(getWhipUrl(streamKey.value), {
        method: 'POST',
        headers: { 'Content-Type': 'application/sdp' },
        body: offer.sdp
      })
      if (!res.ok) throw new Error('WHIP push failed')
      const answer = await res.text()
      await peerConnection.setRemoteDescription({ type: 'answer', sdp: answer })

      if (sourceType.value === 'screen') {
        const screenTrack = stream.getVideoTracks()[0]
        if (screenTrack) screenTrack.onended = () => stopPublishing()
      }

      if (onPublished) {
        await onPublished()
      }
      ElMessage.success(sourceType.value === 'screen' ? '屏幕共享推流中' : '摄像头推流中')
    } catch (e) {
      logger.error?.(e)
      ElMessage.error(e.message || '推流失败，请确保 SRS 已启动')
      stopPublishing()
    }
  }

  watch(sourceType, () => {
    if (!isPublishing.value) return
    stopPublishing()
    startPublishing()
  })

  return {
    sourceType,
    mediaStream,
    isPublishing,
    videoPreviewRef,
    startPublishing,
    stopPublishing
  }
}
