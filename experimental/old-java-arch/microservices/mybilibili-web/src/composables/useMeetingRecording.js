import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const DEFAULT_MIME_TYPE = 'video/webm;codecs=vp9'

const getTrackList = (stream) => {
  return typeof stream?.getTracks === 'function' ? stream.getTracks() : []
}

export function formatRecordingTime(seconds) {
  const m = Math.floor(seconds / 60)
  const sec = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

export function useMeetingRecording({
  localStream,
  remotePeers,
  createMediaStream = () => new MediaStream(),
  createMediaRecorder = (stream, options) => new MediaRecorder(stream, options),
  createBlob = (chunks, options) => new Blob(chunks, options),
  createObjectUrl = blob => URL.createObjectURL(blob),
  revokeObjectUrl = url => URL.revokeObjectURL(url),
  createAnchor = () => document.createElement('a'),
  nowLabel = () => new Date().toLocaleString(),
  setTimer = (handler, delay) => setInterval(handler, delay),
  clearTimer = timerId => clearInterval(timerId),
  logger = console
}) {
  const isRecording = ref(false)
  const recordingTime = ref(0)
  let recordingTimer = null
  let mediaRecorder = null
  let recordedChunks = []

  const collectStreams = () => {
    const streams = []
    if (localStream.value) streams.push(localStream.value)
    Object.values(remotePeers).forEach(peer => {
      if (peer.stream) streams.push(peer.stream)
    })
    return streams
  }

  const clearRecordingTimer = () => {
    if (!recordingTimer) return
    clearTimer(recordingTimer)
    recordingTimer = null
  }

  const downloadRecording = () => {
    const blob = createBlob(recordedChunks, { type: 'video/webm' })
    const url = createObjectUrl(blob)
    const anchor = createAnchor()
    anchor.href = url
    anchor.download = `会议录制_${nowLabel()}.webm`
    anchor.click()
    revokeObjectUrl(url)
  }

  const startRecording = () => {
    if (isRecording.value) return false
    const streams = collectStreams()
    if (streams.length === 0) {
      ElMessage.warning('没有可录制的媒体流')
      return false
    }

    try {
      const combined = createMediaStream()
      streams.forEach(stream => {
        getTrackList(stream).forEach(track => combined.addTrack(track))
      })
      recordedChunks = []
      mediaRecorder = createMediaRecorder(combined, { mimeType: DEFAULT_MIME_TYPE })
      mediaRecorder.ondataavailable = event => {
        if (event.data?.size > 0) recordedChunks.push(event.data)
      }
      mediaRecorder.start(1000)
      isRecording.value = true
      recordingTime.value = 0
      clearRecordingTimer()
      recordingTimer = setTimer(() => { recordingTime.value++ }, 1000)
      ElMessage.success('开始录制')
      return true
    } catch (e) {
      logger.error?.(e)
      ElMessage.error('录制启动失败')
      return false
    }
  }

  const stopRecording = ({ download = true } = {}) => {
    if (!mediaRecorder || mediaRecorder.state === 'inactive') return false
    mediaRecorder.stop()
    clearRecordingTimer()
    isRecording.value = false
    if (download) downloadRecording()
    ElMessage.success('录制已保存')
    return true
  }

  const cleanupRecording = () => {
    stopRecording()
    clearRecordingTimer()
  }

  return {
    isRecording,
    recordingTime,
    startRecording,
    stopRecording,
    cleanupRecording,
    formatRecordingTime
  }
}
