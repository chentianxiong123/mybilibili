const DEFAULT_ICE_SERVERS = [{ urls: 'stun:stun.l.google.com:19302' }]

export function usePeerConnectionMesh({
  localStream,
  remotePeers,
  ensurePeerEntry,
  sendSignal,
  canRestartIce = () => true,
  resetRemoteOnOffer = false,
  iceServers = DEFAULT_ICE_SERVERS,
  logPrefix = '[WebRTC]',
  logger = console
}) {
  const peerConnections = {}
  const pendingIce = {}

  const flushPendingIce = async (peerId, pc) => {
    if (!pendingIce[peerId]) return
    for (const candidate of pendingIce[peerId]) {
      await pc.addIceCandidate(candidate)
    }
    delete pendingIce[peerId]
  }

  const createPeerConnection = async (peerId, isInitiator) => {
    if (peerConnections[peerId]) return peerConnections[peerId]

    const pc = new RTCPeerConnection({ iceServers })
    peerConnections[peerId] = pc

    if (localStream.value) {
      localStream.value.getTracks().forEach(track => {
        pc.addTrack(track, localStream.value)
      })
    }

    pc.ontrack = (event) => {
      ensurePeerEntry(peerId)
      remotePeers[peerId].stream = event.streams[0]
    }

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        sendSignal('ice-candidate', peerId, event.candidate)
      }
    }

    pc.onconnectionstatechange = () => {
      if (['failed', 'closed', 'disconnected'].includes(pc.connectionState)) {
        logger?.warn?.(`${logPrefix} 连接状态`, peerId, pc.connectionState)
      }
    }

    pc.oniceconnectionstatechange = async () => {
      if (pc.iceConnectionState !== 'failed') return
      if (!canRestartIce(peerId, pc)) return
      logger?.warn?.(`${logPrefix} ICE failed, 尝试 restart`, peerId)
      try {
        const offer = await pc.createOffer({ iceRestart: true })
        await pc.setLocalDescription(offer)
        sendSignal('offer', peerId, offer.sdp)
      } catch (e) {
        logger?.error?.(`${logPrefix} ICE restart 失败`, e)
      }
    }

    if (isInitiator) {
      const offer = await pc.createOffer()
      await pc.setLocalDescription(offer)
      sendSignal('offer', peerId, offer.sdp)
    }

    return pc
  }

  const handleOffer = async (peerId, sdp) => {
    if (peerConnections[peerId]) {
      peerConnections[peerId].close()
      delete peerConnections[peerId]
    }
    if (resetRemoteOnOffer) {
      delete remotePeers[peerId]
    }
    const pc = await createPeerConnection(peerId, false)
    await pc.setRemoteDescription({ type: 'offer', sdp })
    await flushPendingIce(peerId, pc)
    const answer = await pc.createAnswer()
    await pc.setLocalDescription(answer)
    sendSignal('answer', peerId, answer.sdp)
  }

  const handleAnswer = async (peerId, sdp) => {
    const pc = peerConnections[peerId]
    if (!pc) return
    await pc.setRemoteDescription({ type: 'answer', sdp })
    await flushPendingIce(peerId, pc)
  }

  const handleIceCandidate = async (peerId, candidate) => {
    const pc = peerConnections[peerId]
    if (pc && pc.remoteDescription) {
      try {
        await pc.addIceCandidate(candidate)
      } catch (e) {
        logger?.warn?.('addIceCandidate', e)
      }
      return
    }
    ;(pendingIce[peerId] = pendingIce[peerId] || []).push(candidate)
  }

  const removePeerConnection = (peerId) => {
    const pc = peerConnections[peerId]
    if (pc) {
      pc.close()
      delete peerConnections[peerId]
    }
    delete remotePeers[peerId]
    delete pendingIce[peerId]
  }

  const closeAllPeerConnections = () => {
    Object.keys(peerConnections).forEach(removePeerConnection)
  }

  const replaceVideoTrack = (videoTrack) => {
    if (!videoTrack) return
    Object.values(peerConnections).forEach(pc => {
      const sender = pc.getSenders().find(s => s.track && s.track.kind === 'video')
      if (sender) sender.replaceTrack(videoTrack)
    })
  }

  return {
    peerConnections,
    createPeerConnection,
    handleOffer,
    handleAnswer,
    handleIceCandidate,
    removePeerConnection,
    closeAllPeerConnections,
    replaceVideoTrack
  }
}
