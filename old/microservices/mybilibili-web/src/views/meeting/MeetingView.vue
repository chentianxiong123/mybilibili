<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Microphone, Mute, VideoCamera, VideoPause, UserFilled, CopyDocument, Plus, Monitor, ChatDotRound, Share, Calendar } from '@element-plus/icons-vue'
import { getCurrentUser } from '../../utils/currentUser.js'
import { getMeetingStatusText, getMeetingStatusType } from '../../utils/liveMeetingStatus.js'
import { useMeetingSessionBridge } from '../../composables/useMeetingSessionBridge.js'
import { usePeerConnectionMesh } from '../../composables/usePeerConnectionMesh.js'
import { useMeetingChat } from '../../composables/useMeetingChat.js'
import { useMeetingHostControls } from '../../composables/useMeetingHostControls.js'
import { useMeetingMediaControls } from '../../composables/useMeetingMediaControls.js'
import { useMeetingRecording } from '../../composables/useMeetingRecording.js'
import { useMeetingPresenceSignals } from '../../composables/useMeetingPresenceSignals.js'
import { useMeetingRoomCatalog } from '../../composables/useMeetingRoomCatalog.js'
import { useMeetingRoomSession } from '../../composables/useMeetingRoomSession.js'
import { useMeetingSessionUi } from '../../composables/useMeetingSessionUi.js'
import { useMeetingSessionLifecycle } from '../../composables/useMeetingSessionLifecycle.js'
import { useMeetingSignalingRouter } from '../../composables/useMeetingSignalingRouter.js'

const joinCode = ref('')

const me = ref(getCurrentUser({
  temporary: true,
  temporaryKey: 'meetingTmpId',
  temporaryPrefix: '访客'
}))

const localStream = ref(null)
const localVideoRef = ref(null)
const currentRoom = ref(null)
const hostUserId = ref(null)

// 远端：peerId(=对端 userId) -> { stream, name }
const remotePeers = reactive({})    // 响应式，驱动 video 网格

const audioEnabled = ref(true)
const videoEnabled = ref(true)
const screenShareEnabled = ref(false)
const virtualBgEnabled = ref(false)   // 虚拟背景/背景模糊
const virtualBgStream = ref(null)    // 处理后的流（含背景模糊）
const roomLocked = ref(false)       // 会议是否被锁定

const {
  sendMeetingMessage,
  sendSignaling,
  handleSignaling,
  joinRoomByCode,
  cancelWaitingForHost,
  leaveRoom: leaveMeetingSession,
  cleanupRoomSession,
  enterApprovedRoom,
  attachMeetingSession,
  attachSignalingRouter
} = useMeetingSessionBridge()

const broadcastSelfState = () => {
  sendMeetingMessage({
    type: 'peer-state',
    data: {
      audioEnabled: audioEnabled.value,
      videoEnabled: videoEnabled.value,
      screenShareEnabled: screenShareEnabled.value
    }
  })
}

const {
  isRecording,
  recordingTime,
  startRecording,
  stopRecording,
  cleanupRecording,
  formatRecordingTime
} = useMeetingRecording({
  localStream,
  remotePeers
})

const {
  loading,
  myRooms,
  newRoomName,
  reserveLoading,
  reserveForm,
  loadMyRooms,
  createRoom,
  reserveRoom
} = useMeetingRoomCatalog({
  onRoomCreated: async room => {
    if (room?.roomCode) await joinRoomByCode(room.roomCode)
  }
})

// 占位（避免 ontrack 来时还没有条目导致 UI 不显示昵称）
const ensurePeerEntry = (peerId, name) => {
  if (!remotePeers[peerId]) {
    remotePeers[peerId] = {
      name: name || ('用户' + peerId),
      audioEnabled: true,
      videoEnabled: true,
      screenShareEnabled: false,
      handRaised: false
    }
  } else if (name && !remotePeers[peerId].name) {
    remotePeers[peerId].name = name
  }
}

const {
  peerConnections,
  createPeerConnection,
  handleOffer,
  handleAnswer,
  handleIceCandidate,
  removePeerConnection,
  closeAllPeerConnections,
  replaceVideoTrack
} = usePeerConnectionMesh({
  localStream,
  remotePeers,
  ensurePeerEntry,
  sendSignal: sendSignaling,
  logPrefix: '[Meeting]'
})

const {
  activeSpeakerId,
  peerQuality,
  updateAudioActivity,
  getQualityLevel,
  startQualityPolling,
  cleanupPresenceSignals
} = useMeetingPresenceSignals({
  currentRoom,
  peerConnections
})

const {
  startLocalStream,
  toggleAudio,
  toggleVideo,
  startScreenShare,
  toggleVirtualBg,
  cleanupMedia
} = useMeetingMediaControls({
  localStream,
  localVideoRef,
  audioEnabled,
  videoEnabled,
  screenShareEnabled,
  virtualBgEnabled,
  virtualBgStream,
  replaceVideoTrack,
  broadcastSelfState,
  updateAudioActivity: (...args) => updateAudioActivity(...args)
})

const {
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
  formatDuration,
  showControls,
  restartMeetingTimer,
  registerKeyboardShortcuts,
  unregisterKeyboardShortcuts,
  setRemoteVideo,
  togglePiP,
  handleVideoDoubleClick,
  startEditName,
  saveName,
  toggleHandRaise,
  cleanupMeetingSessionUi
} = useMeetingSessionUi({
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
  startScreenShare
})

const {
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
} = useMeetingChat({
  me,
  participantList,
  sendMeetingMessage
})

const {
  approveHandRaise,
  lowerHand,
  admitUser,
  rejectUser,
  spotlightUser,
  toggleLockRoom,
  transferHost,
  kickUser,
  muteUser,
  muteAll,
  handleJoinRequestMessage,
  handleHandRaisedMessage,
  handleLockRoomMessage,
  handleTransferHostMessage
} = useMeetingHostControls({
  me,
  isHost,
  remotePeers,
  waitingUsers,
  roomLocked,
  spotlightId,
  hostUserId,
  sendMeetingMessage,
  removePeerConnection
})

const {
  cleanupRoomArtifacts,
  leaveRoom,
  cleanup,
  shareRoom,
  copyRoomCode
} = useMeetingSessionLifecycle({
  currentRoom,
  hostUserId,
  cleanupRecording,
  cleanupMeetingSessionUi,
  closeAllPeerConnections,
  cleanupMedia,
  cleanupPresenceSignals,
  cleanupRoomSession,
  leaveMeetingSession
})

attachSignalingRouter(useMeetingSignalingRouter({
  me,
  currentRoom,
  isHost,
  waitingForHost,
  remotePeers,
  audioEnabled,
  spotlightId,
  ensurePeerEntry,
  createPeerConnection,
  broadcastSelfState,
  removePeerConnection,
  handleOffer,
  handleAnswer,
  handleIceCandidate,
  updateAudioActivity,
  toggleAudio,
  enterApprovedRoom,
  cancelWaitingForHost,
  leaveRoom,
  cleanupRoomSession,
  cleanupRoomArtifacts,
  loadMyRooms,
  receiveChatMessage,
  handleJoinRequestMessage,
  handleHandRaisedMessage,
  handleLockRoomMessage,
  handleTransferHostMessage
}))

attachMeetingSession(useMeetingRoomSession({
  me,
  joinCode,
  currentRoom,
  hostUserId,
  loading,
  waitingForHost,
  startLocalStream,
  restartMeetingTimer,
  startQualityPolling,
  loadMyRooms,
  onMessage: handleSignaling
}))

onMounted(async () => {
  await loadMyRooms()
  registerKeyboardShortcuts()
})

onUnmounted(() => {
  cleanup()
  unregisterKeyboardShortcuts()
})
</script>

<template>
  <div class="meeting-page">
    <!-- 非主持人等待审批 -->
    <div v-if="waitingForHost" class="waiting-room">
      <div class="waiting-card">
        <div class="waiting-icon">⏳</div>
        <h3>等待主持人批准</h3>
        <p>请耐心等待，主持人允许后将自动加入会议</p>
        <el-button type="danger" @click="cancelWaitingForHost">取消等待</el-button>
      </div>
    </div>

    <!-- 主持人审批面板 -->
    <div v-if="isHost && waitingUsers.length > 0" class="waiting-approve-bar">
      <div class="approve-title">加入申请 ({{ waitingUsers.length }})</div>
      <div v-for="w in waitingUsers" :key="w.userId" class="waiting-item">
        <span>{{ w.userName || ('用户' + w.userId) }}</span>
        <el-button size="small" type="success" @click="admitUser(w.userId)">允许</el-button>
        <el-button size="small" type="danger" @click="rejectUser(w.userId)">拒绝</el-button>
      </div>
    </div>

    <!-- 会议中视图 -->
    <div v-if="currentRoom && !waitingForHost" class="meeting-active">
      <div class="meeting-header">
        <div class="header-left">
          <h2>{{ currentRoom.roomName }}</h2>
          <span class="room-code" @click="copyRoomCode(currentRoom.roomCode)">
            <CopyDocument style="width: 12px; height: 12px;" />
            {{ currentRoom.roomCode }}
          </span>
          <el-button size="small" @click="shareRoom" style="font-size:12px;">
            <el-icon><Share /></el-icon>
            邀请加入
          </el-button>
          <el-tag v-if="isHost" type="warning" size="small" effect="dark">主持人</el-tag>
          <el-tag type="info" size="small">{{ formatDuration(meetingDuration) }}</el-tag>
        </div>
        <div class="header-right">
          <el-button v-if="isHost" type="warning" plain @click="muteAll">全员静音</el-button>
          <el-button @click="chatOpen = !chatOpen">
            <el-icon><ChatDotRound /></el-icon>
            <span style="margin-left:4px">聊天</span>
          </el-button>
          <el-button @click="sidebarOpen = !sidebarOpen">
            <el-icon><UserFilled /></el-icon>
            <span style="margin-left:4px">{{ participantList.length }}</span>
          </el-button>
          <el-button v-if="!isRecording" type="info" plain @click="startRecording">
            <el-icon><VideoCamera /></el-icon>
            录制
          </el-button>
          <el-button v-else type="danger" plain @click="stopRecording">
            <span class="rec-dot" />{{ formatRecordingTime(recordingTime) }}
          </el-button>
          <el-button type="danger" @click="leaveRoom">离开会议</el-button>
        </div>
      </div>

      <div class="meeting-body" @mousemove="showControls" :class="{ 'speaker-mode': activeSpeakerId && !screenSharerId }">
        <div class="video-grid" :class="{ 'has-sharer': screenSharerId, 'has-speaker': activeSpeakerId && !screenSharerId, 'has-spotlight': spotlightId && !screenSharerId }">
          <!-- 本地视频 -->
          <div class="video-card local" :class="{ 'main-sharer': screenSharerId === 'local', 'active-speaker': activeSpeakerId === 'local', 'spotlighted': spotlightId === 'local' }" @dblclick="handleVideoDoubleClick('local')">
            <video ref="localVideoRef" autoplay muted playsinline class="video-element" />
            <div class="video-label" :class="{ 'controls-hidden': !controlsVisible }">我</div>
            <div class="video-controls" :class="{ 'controls-hidden': !controlsVisible }">
              <el-button :type="audioEnabled ? '' : 'danger'" circle @click="toggleAudio">
                <el-icon><Microphone v-if="audioEnabled" /><Mute v-else /></el-icon>
              </el-button>
              <el-button :type="videoEnabled ? '' : 'danger'" circle @click="toggleVideo">
                <el-icon><VideoCamera v-if="videoEnabled" /><VideoPause v-else /></el-icon>
              </el-button>
              <el-button :type="screenShareEnabled ? 'primary' : ''" circle @click="startScreenShare">
                <el-icon><Monitor /></el-icon>
              </el-button>
              <el-button :type="virtualBgEnabled ? 'warning' : ''" circle @click="toggleVirtualBg" title="虚拟背景/背景模糊">
                🎨
              </el-button>
              <el-button circle @click="togglePiP('local')" title="画中画">
                <el-icon><VideoCamera /></el-icon>
              </el-button>
              <el-button :type="handRaised ? 'warning' : ''" circle @click="toggleHandRaise" title="举手">
                ✋
              </el-button>
              <el-button v-if="isHost" :type="roomLocked ? 'danger' : ''" circle @click="toggleLockRoom" :title="roomLocked ? '解锁会议' : '锁定会议'">
                🔒
              </el-button>
            </div>
          </div>
          <!-- 举手计数 -->
          <div v-if="totalHandRaised > 0" class="hand-count-badge">
            ✋ {{ totalHandRaised }} 人举手
          </div>

          <!-- 远端视频 -->
          <div v-for="peerId in remotePeerIds" :key="peerId" class="video-card" :class="{ 'main-sharer': screenSharerId === peerId, 'active-speaker': activeSpeakerId === peerId, 'spotlighted': String(spotlightId) === String(peerId) }" @dblclick="handleVideoDoubleClick(peerId)">
            <video :ref="el => setRemoteVideo(peerId, el)" autoplay playsinline class="video-element" />
            <div class="video-label" :class="{ 'controls-hidden': !controlsVisible }">{{ remotePeers[peerId]?.name || ('参与者 ' + (remotePeerIds.indexOf(peerId) + 1)) }}</div>
            <div v-if="remotePeers[peerId]" class="video-status" :class="{ 'controls-hidden': !controlsVisible }">
              <el-icon v-if="!remotePeers[peerId].audioEnabled" class="status-off"><Mute /></el-icon>
              <el-icon v-if="!remotePeers[peerId].videoEnabled" class="status-off"><VideoPause /></el-icon>
              <el-icon v-if="remotePeers[peerId].screenShareEnabled" class="status-on"><Monitor /></el-icon>
            </div>
          </div>
        </div>

        <!-- 参与者侧栏 -->
        <transition name="slide-x">
          <aside v-if="sidebarOpen" class="participant-sidebar">
            <div class="sidebar-header">
              参与者 ({{ participantList.length }})
              <span v-if="!editingName" class="my-name" @click="startEditName" title="点击修改昵称">
                我: {{ me.name }} ✏️
              </span>
              <input v-else ref="editNameInput" v-model="editNameDraft" class="name-input" @blur="saveName" @keydown.enter="saveName" />
            </div>
            <div class="participant-list">
              <div v-for="p in participantList" :key="p.userId" class="participant-item">
                <div class="p-info">
                  <span class="p-name">{{ p.name }}</span>
                  <el-tag v-if="p.isHost" type="warning" size="small">主持</el-tag>
                  <el-tag v-if="p.handRaised" type="info" size="small">✋ 举手</el-tag>
                </div>
                <div class="p-status">
                  <div :class="['quality-bars', `q${getQualityLevel(p.userId)}`]" :title="`延迟${peerQuality[p.userId]?.rtt || '?'}ms 丢包${peerQuality[p.userId]?.lossRate || 0}%`">
                    <span /><span /><span />
                  </div>
                  <el-icon :class="p.audioEnabled ? 'on' : 'off'" :title="p.audioEnabled ? '麦克风开' : '已静音'">
                    <Microphone v-if="p.audioEnabled" /><Mute v-else />
                  </el-icon>
                  <el-icon :class="p.videoEnabled ? 'on' : 'off'" :title="p.videoEnabled ? '摄像头开' : '摄像头关'">
                    <VideoCamera v-if="p.videoEnabled" /><VideoPause v-else />
                  </el-icon>
                  <el-icon v-if="p.screenShareEnabled" class="on" title="屏幕共享中"><Monitor /></el-icon>
                </div>
                <div v-if="isHost && !p.self" class="p-actions">
                  <el-button v-if="p.handRaised" size="small" type="success" @click="approveHandRaise(p.userId)">批准发言</el-button>
                  <el-button v-if="p.handRaised" size="small" link type="warning" @click="lowerHand(p.userId)">放下</el-button>
                  <el-button size="small" :type="String(spotlightId) === String(p.userId) ? 'warning' : ''" link @click="spotlightUser(p.userId)">聚焦</el-button>
                  <el-button size="small" link @click="muteUser(p.userId)" :disabled="!p.audioEnabled">静音</el-button>
                  <el-button size="small" link type="warning" @click="transferHost(p.userId)">转交主持</el-button>
                  <el-button size="small" link type="danger" @click="kickUser(p.userId)">移出</el-button>
                </div>
              </div>
            </div>
          </aside>
        </transition>

        <!-- 聊天侧栏 -->
        <transition name="slide-x">
          <aside v-if="chatOpen" class="chat-sidebar">
            <div class="sidebar-header">会议聊天</div>
            <div class="chat-messages">
              <div v-if="chatMessages.length === 0" class="chat-empty">暂无消息</div>
              <div v-for="m in chatMessages" :key="m.id" :class="['chat-msg', { self: m.self, mentioned: m.mentionId }]">
                <div class="chat-from">
                  {{ m.from }}
                  <el-tag v-if="m.mentionId" size="small" type="warning">@我</el-tag>
                </div>
                <div class="chat-text">{{ m.text }}</div>
              </div>
            </div>
            <!-- @提及下拉 -->
            <div v-if="showMentionDropdown && mentionSuggestions.length > 0" class="mention-dropdown">
              <div v-for="p in mentionSuggestions" :key="p.userId" class="mention-item" @click="selectMention(p)">
                {{ p.name }}<el-tag v-if="p.isHost" size="small" type="warning" style="margin-left:4px">主持</el-tag>
              </div>
            </div>
            <div class="chat-input">
              <div class="emoji-wrap">
                <el-button circle @click="toggleEmojiPicker" title="表情">😀</el-button>
                <transition name="fade">
                  <div v-if="showEmojiPicker" class="emoji-picker">
                    <span v-for="e in emojis" :key="e" class="emoji-item" @click="insertEmoji(e)">{{ e }}</span>
                  </div>
                </transition>
              </div>
              <el-input v-model="chatText" placeholder="说点什么..." @keydown.enter="sendChat" @input="handleChatInput" />
              <el-button type="primary" @click="sendChat">发送</el-button>
            </div>
          </aside>
        </transition>
      </div>

      <!-- 空槽位提示 -->
      <div v-if="remotePeerIds.length === 0" class="waiting-hint">
        等待其他人加入...
      </div>
    </div>

    <!-- 会议室列表视图 -->
    <div v-else class="meeting-list">
      <div class="page-banner">
        <h2>视频会议</h2>
      </div>

      <div class="meeting-grid">
        <!-- 创建会议室 -->
        <el-card class="create-card" shadow="never">
          <template #header>
            <div class="card-header">
              <Plus style="width: 18px; height: 18px;" />
              <span>创建会议室</span>
            </div>
          </template>
          <div class="create-form">
            <el-input v-model="newRoomName" placeholder="会议室名称" />
            <el-button type="primary" style="width: 100%; margin-top: 12px;" @click="createRoom" :loading="loading">
              创建并加入
            </el-button>
          </div>
        </el-card>

        <!-- 加入会议室 -->
        <el-card class="join-card" shadow="never">
          <template #header>
            <div class="card-header">
              <UserFilled style="width: 18px; height: 18px;" />
              <span>加入会议</span>
            </div>
          </template>
          <div class="join-form">
            <el-input v-model="joinCode" placeholder="输入邀请码" maxlength="6" />
            <el-button type="success" style="width: 100%; margin-top: 12px;" @click="joinRoomByCode()" :loading="loading">
              加入
            </el-button>
          </div>
        </el-card>

        <!-- 预约会议室 -->
        <el-card class="reserve-card" shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon><Calendar /></el-icon>
              <span>预约会议</span>
            </div>
          </template>
          <div class="reserve-form">
            <el-input v-model="reserveForm.roomName" placeholder="会议室名称" />
            <el-date-picker
              v-model="reserveForm.timeRange"
              type="datetimerange"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              value-format="YYYY-MM-DDTHH:mm:ss"
              format="MM-DD HH:mm"
              style="width: 100%; margin-top: 12px;"
            />
            <el-input
              v-model="reserveForm.reason"
              type="textarea"
              :rows="2"
              maxlength="200"
              show-word-limit
              placeholder="预约事由"
              style="margin-top: 12px;"
            />
            <el-button type="primary" plain style="width: 100%; margin-top: 12px;" @click="reserveRoom" :loading="reserveLoading">
              提交预约
            </el-button>
          </div>
        </el-card>

        <!-- 我的会议室 -->
        <el-card class="rooms-card" shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon><UserFilled /></el-icon>
              <span>我的会议室</span>
            </div>
          </template>
          <div class="rooms-list" v-if="myRooms.length > 0">
            <div v-for="room in myRooms" :key="room.id" class="room-item" @click="joinRoomByCode(room.roomCode)">
              <div class="room-info">
                <span class="room-name">{{ room.roomName }}</span>
                <span class="room-code" @click.stop="copyRoomCode(room.roomCode)">
                  <CopyDocument style="width: 12px; height: 12px;" />
                  {{ room.roomCode }}
                </span>
              </div>
              <el-tag :type="getMeetingStatusType(room.status)" size="small">
                {{ getMeetingStatusText(room.status) }}
              </el-tag>
            </div>
          </div>
          <el-empty v-else description="暂无会议室" />
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.meeting-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px;
  min-height: calc(100vh - 64px);
  background: #f6f7f8;
}

/* 会议中 */
.meeting-active {
  height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
}
.meeting-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 20px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.header-left h2 {
  margin: 0;
  font-size: 20px;
}
.room-code {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  background: #e3e5e7;
  border-radius: 6px;
  font-size: 13px;
  color: #61666d;
  cursor: pointer;
}
.room-code:hover {
  background: #dcdfe6;
}

.video-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  align-content: start;
}
.meeting-body {
  flex: 1;
  display: flex;
  gap: 16px;
  overflow: hidden;
}
.meeting-body .video-grid {
  flex: 1;
  overflow-y: auto;
}
.video-status {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  gap: 6px;
  padding: 4px 8px;
  background: rgba(0,0,0,0.5);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}
.video-status .status-off { color: #ff5252; }
.video-status .status-on { color: #4fc3f7; }
.video-card.spotlighted {
  border: 3px solid #e6a23c;
  box-shadow: 0 0 20px rgba(230, 162, 60, 0.5);
}
.video-grid.has-spotlight {
  display: grid;
  grid-template-columns: 1fr 200px;
  gap: 16px;
}
.video-grid.has-spotlight .video-card.spotlighted {
  grid-column: 1;
  aspect-ratio: 16/9;
}
.video-grid.has-spotlight .video-card:not(.spotlighted) {
  grid-column: 2;
  aspect-ratio: 16/9;
  max-height: 160px;
  align-self: start;
}

/* 参与者侧栏 */
.participant-sidebar {
  width: 280px;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
}
.sidebar-header {
  padding: 14px 16px;
  font-weight: 600;
  border-bottom: 1px solid #e3e5e7;
  font-size: 14px;
  color: #18191c;
}
.participant-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}
.participant-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  gap: 8px;
  border-bottom: 1px solid #f6f7f8;
  flex-wrap: wrap;
}
.p-info {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}
.p-name {
  font-size: 13px;
  color: #18191c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.p-status {
  display: flex;
  gap: 6px;
  font-size: 14px;
}
.p-status .on { color: #67c23a; }
.p-status .off { color: #c0c4cc; }
.p-actions {
  width: 100%;
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}
.slide-x-enter-active, .slide-x-leave-active { transition: transform .25s ease, opacity .25s ease; }
.slide-x-enter-from, .slide-x-leave-to { transform: translateX(20px); opacity: 0; }

/* 聊天侧栏 */
.chat-sidebar {
  width: 300px;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.chat-empty {
  text-align: center;
  color: #9499a0;
  font-size: 13px;
  padding-top: 40px;
}
.chat-msg {
  background: #f6f7f8;
  padding: 8px 12px;
  border-radius: 8px;
  max-width: 80%;
  word-break: break-word;
}
.chat-msg.self {
  align-self: flex-end;
  background: #e6f4ff;
}
.chat-from {
  font-size: 11px;
  color: #9499a0;
  margin-bottom: 2px;
}
.chat-text {
  font-size: 13px;
  color: #18191c;
}
.chat-input {
  display: flex;
  gap: 6px;
  padding: 10px;
  border-top: 1px solid #e3e5e7;
}
.chat-input { display: flex; gap: 6px; padding: 10px; border-top: 1px solid #e3e5e7; align-items: center; }
.chat-input .el-input { flex: 1; }
.emoji-wrap { position: relative; }
.emoji-picker {
  position: absolute;
  bottom: 40px;
  left: 0;
  background: #fff;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 8px;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  width: 240px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 10;
}
.emoji-item { font-size: 20px; cursor: pointer; padding: 4px; border-radius: 4px; }
.emoji-item:hover { background: #f6f7f8; }
.fade-enter-active, .fade-leave-active { transition: opacity .2s; }
/* @提及 */
.mention-dropdown {
  border-top: 1px solid #e3e5e7;
  max-height: 120px;
  overflow-y: auto;
}
.mention-item {
  padding: 6px 12px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.mention-item:hover { background: #f6f7f8; }
.chat-msg.mentioned .chat-text { color: #e6a23c; font-weight: 500; }
.hand-count-badge {
  position: absolute;
  bottom: 16px;
  left: 16px;
  padding: 4px 12px;
  background: rgba(240, 64, 64, 0.9);
  color: #fff;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  z-index: 10;
}
/* 昵称编辑 */
.sidebar-header .my-name {
  font-size: 12px;
  color: #00a1d6;
  cursor: pointer;
  margin-left: 8px;
  font-weight: 400;
}
.sidebar-header .name-input {
  border: 1px solid #00a1d6;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 12px;
  width: 100px;
  margin-left: 8px;
}
/* 连接质量指示 */
.quality-bars {
  display: inline-flex;
  align-items: flex-end;
  gap: 1px;
  height: 12px;
}
.quality-bars span {
  width: 3px;
  border-radius: 1px;
  background: #c0c4cc;
}
.quality-bars span:nth-child(1) { height: 4px; }
.quality-bars span:nth-child(2) { height: 7px; }
.quality-bars span:nth-child(3) { height: 10px; }
.quality-bars.q1 span { background: #f04040; }
.quality-bars.q2 span { background: #e6a23c; }
.quality-bars.q3 span { background: #67c23a; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.header-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
.rec-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  background: #f04040;
  border-radius: 50%;
  margin-right: 4px;
  animation: rec-blink 1s infinite;
}
@keyframes rec-blink { 0%,100%{opacity:1} 50%{opacity:0.3} }

/* 等待室 */
.waiting-room {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
}
.waiting-card {
  text-align: center;
  padding: 48px;
  background: #fff;
  border-radius: 16px;
  max-width: 400px;
}
.waiting-icon { font-size: 64px; margin-bottom: 16px; }
.waiting-card h3 { margin: 0 0 12px; font-size: 20px; color: #18191c; }
.waiting-card p { margin: 0 0 24px; color: #9499a0; font-size: 14px; }

/* 待审批加入栏 */
.waiting-approve-bar {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}
.approve-title { font-size: 14px; font-weight: 600; margin-bottom: 12px; color: #18191c; }
.waiting-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid #f6f7f8;
}
.waiting-item:last-child { border-bottom: none; }
.waiting-item span { flex: 1; font-size: 13px; }

.video-card {
  position: relative;
  background: #1a1a1a;
  border-radius: 12px;
  overflow: hidden;
  aspect-ratio: 16/9;
}
.video-card.local {
  border: 2px solid #00a1d6;
}
.video-element {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.video-label {
  position: absolute;
  bottom: 12px;
  left: 12px;
  padding: 4px 12px;
  background: rgba(0,0,0,0.6);
  color: #fff;
  border-radius: 4px;
  font-size: 13px;
}
.video-controls {
  position: absolute;
  bottom: 12px;
  right: 12px;
  display: flex;
  gap: 8px;
}

.waiting-hint {
  text-align: center;
  color: #9499a0;
  padding: 40px;
  font-size: 14px;
}

/* 会议室列表 */
.meeting-list .page-banner {
  margin-bottom: 28px;
}
.meeting-list .page-banner h2 {
  font-size: 24px;
  font-weight: 700;
  color: #18191c;
  margin: 0;
}
.meeting-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}
@media (max-width: 768px) {
  .meeting-grid {
    grid-template-columns: 1fr;
  }
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #18191c;
}

.rooms-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f6f7f8;
  border-radius: 8px;
  cursor: pointer;
}
.room-item:hover {
  background: #e3e5e7;
}
.room-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.room-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
}
.controls-hidden {
  opacity: 0 !important;
  transition: opacity 0.4s ease;
}
.video-card:hover .video-label,
.video-card:hover .video-controls,
.video-card:hover .video-status {
  opacity: 1;
}
.video-label.controls-hidden,
.video-controls.controls-hidden,
.video-status.controls-hidden {
  opacity: 0;
}
.video-status { transition: opacity 0.4s ease; }
.video-label { transition: opacity 0.4s ease; }
.video-controls { transition: opacity 0.4s ease; }

/* 屏幕共享主画面布局 */
.video-grid.has-sharer {
  display: grid;
  grid-template-columns: 1fr 220px;
  grid-template-rows: 1fr;
  gap: 16px;
}
.video-grid.has-sharer .video-card.main-sharer {
  grid-column: 1;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: none;
}
.video-grid.has-sharer .video-card:not(.main-sharer) {
  grid-column: 2;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: 200px;
  align-self: start;
}

/* 发言者视图布局（与屏幕共享互斥） */
.video-grid.has-speaker {
  display: grid;
  grid-template-columns: 1fr 200px;
  grid-template-rows: 1fr;
  gap: 16px;
}
.video-grid.has-speaker .video-card.active-speaker {
  grid-column: 1;
  grid-row: 1;
  aspect-ratio: 16/9;
}
.video-grid.has-speaker .video-card:not(.active-speaker) {
  grid-column: 2;
  grid-row: 1;
  aspect-ratio: 16/9;
  max-height: 180px;
  align-self: start;
}
</style>
