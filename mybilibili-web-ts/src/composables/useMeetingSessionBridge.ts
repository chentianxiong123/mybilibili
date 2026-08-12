import { shallowRef } from 'vue'

export function useMeetingSessionBridge() {
  const meetingSession = shallowRef(null)
  const signalingRouter = shallowRef(null)

  const attachMeetingSession = (session) => {
    meetingSession.value = session
    return session
  }

  const attachSignalingRouter = (router) => {
    signalingRouter.value = router
    return router
  }

  const sendMeetingMessage = (payload) => meetingSession.value?.sendMeetingMessage(payload) || false
  const sendSignaling = (type, targetUserId, data) => meetingSession.value?.sendSignaling(type, targetUserId, data) || false
  const handleSignaling = async (msg) => signalingRouter.value?.handleSignaling(msg) || false
  const joinRoomByCode = (code) => meetingSession.value?.joinRoomByCode(code) || false
  const cancelWaitingForHost = () => meetingSession.value?.cancelWaitingForHost() || false
  const leaveRoom = async () => {
    const left = await meetingSession.value?.leaveRoom()
    return left === true
  }
  const cleanupRoomSession = () => meetingSession.value?.cleanupRoomSession() || false
  const enterApprovedRoom = () => meetingSession.value?.enterApprovedRoom() || false

  return {
    meetingSession,
    signalingRouter,
    attachMeetingSession,
    attachSignalingRouter,
    sendMeetingMessage,
    sendSignaling,
    handleSignaling,
    joinRoomByCode,
    cancelWaitingForHost,
    leaveRoom,
    cleanupRoomSession,
    enterApprovedRoom
  }
}
