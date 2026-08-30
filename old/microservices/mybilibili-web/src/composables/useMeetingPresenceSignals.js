import { reactive, ref } from 'vue'

export function useMeetingPresenceSignals({
  currentRoom,
  peerConnections,
  speakerTimeoutMs = 3000,
  speakingThresholdMs = 2000,
  qualityHighThresholdMs = 500,
  qualityMediumThresholdMs = 200,
  qualityLossHigh = 10,
  qualityLossMedium = 3,
  now = () => Date.now(),
  setTimer = (handler, delay) => setTimeout(handler, delay),
  clearTimer = timerId => clearTimeout(timerId),
  qualityPollMs = 5000,
  setIntervalFn = (handler, delay) => setInterval(handler, delay),
  clearIntervalFn = timerId => clearInterval(timerId)
}) {
  const activeSpeakerId = ref(null)
  const peerQuality = reactive({})
  const trackAudioStart = {}
  let speakerTimer = null
  let qualityTimer = null

  const resetSpeakerTimer = () => {
    if (speakerTimer) clearTimer(speakerTimer)
    speakerTimer = setTimer(() => {
      activeSpeakerId.value = null
      speakerTimer = null
    }, speakerTimeoutMs)
  }

  const updateAudioActivity = (peerId, isEnabled) => {
    if (isEnabled) {
      if (trackAudioStart[peerId] === undefined) trackAudioStart[peerId] = now()
      const elapsed = now() - trackAudioStart[peerId]
      if (elapsed > speakingThresholdMs && activeSpeakerId.value !== peerId) {
        activeSpeakerId.value = peerId
        resetSpeakerTimer()
      }
    } else {
      delete trackAudioStart[peerId]
      if (activeSpeakerId.value === peerId) {
        activeSpeakerId.value = null
      }
    }
  }

  const getQualityLevel = (peerId) => {
    const q = peerQuality[peerId]
    if (!q) return 0
    if (q.lossRate > qualityLossHigh || q.rtt > qualityHighThresholdMs) return 1
    if (q.lossRate > qualityLossMedium || q.rtt > qualityMediumThresholdMs) return 2
    return 3
  }

  const pollPeerQuality = async () => {
    if (!currentRoom.value) return
    for (const [peerId, pc] of Object.entries(peerConnections)) {
      try {
        const stats = await pc.getStats()
        let rtt = 0
        let lost = 0
        let total = 0
        stats.forEach(report => {
          if (report.type === 'candidate-pair' && report.state === 'succeeded') {
            rtt = report.currentRoundTripTime ? Math.round(report.currentRoundTripTime * 1000) : 0
          }
          if (report.type === 'inbound-rtp') {
            lost = report.packetsLost || 0
            total = (report.packetsReceived || 0) + lost
          }
        })
        peerQuality[peerId] = {
          rtt,
          lossRate: total > 0 ? Math.round((lost / total) * 100) : 0
        }
      } catch (e) {}
    }
  }

  const startQualityPolling = () => {
    if (qualityTimer) return false
    qualityTimer = setIntervalFn(() => {
      pollPeerQuality()
    }, qualityPollMs)
    return true
  }

  const stopQualityPolling = () => {
    if (!qualityTimer) return false
    clearIntervalFn(qualityTimer)
    qualityTimer = null
    return true
  }

  const cleanupPresenceSignals = () => {
    if (speakerTimer) {
      clearTimer(speakerTimer)
      speakerTimer = null
    }
    stopQualityPolling()
    activeSpeakerId.value = null
    Object.keys(trackAudioStart).forEach(peerId => {
      delete trackAudioStart[peerId]
    })
    Object.keys(peerQuality).forEach(peerId => {
      delete peerQuality[peerId]
    })
  }

  return {
    activeSpeakerId,
    peerQuality,
    updateAudioActivity,
    getQualityLevel,
    pollPeerQuality,
    startQualityPolling,
    stopQualityPolling,
    cleanupPresenceSignals
  }
}
