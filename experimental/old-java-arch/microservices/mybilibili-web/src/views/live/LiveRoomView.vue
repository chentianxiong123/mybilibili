<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Phone, PhoneFilled, Microphone, Mute, VideoCamera, VideoPause, Share } from '@element-plus/icons-vue'
import { isOfflineRoomStatus } from '../../utils/liveMeetingStatus.js'
import { useLiveRoomSession } from '../../composables/useLiveRoomSession.js'
import { useLiveRoomFollow } from '../../composables/useLiveRoomFollow.js'
import { useLiveRoomPlayer } from '../../composables/useLiveRoomPlayer.js'
import { useLiveRoomRealtime } from '../../composables/useLiveRoomRealtime.js'
import { useLiveRoomInteractions } from '../../composables/useLiveRoomInteractions.js'
import { useLiveLinkmic } from '../../composables/useLiveLinkmic.js'

const route = useRoute()
const roomId = route.params.roomId
const {
  room,
  loading,
  me,
  currentUserId,
  isStreamer,
  initializeLiveRoom,
  shareRoom
} = useLiveRoomSession({ roomId })

const {
  isFollowing,
  followerCount,
  syncFollowState,
  toggleFollow
} = useLiveRoomFollow({ room })

const {
  playerRef,
  replayPlayerRef,
  isBuffering,
  currentQuality,
  showReplayTab,
  replayList,
  currentReplay,
  initPlayer,
  emitDanmaku,
  loadReplays,
  playReplay,
  closeReplay,
  cleanupLiveRoomPlayer
} = useLiveRoomPlayer({ room })

const {
  realtimeViewerCount,
  sendRoomMessage,
  connectLiveRoomSocket,
  cleanupLiveRoomRealtime,
  attachRoomMessageHandlers
} = useLiveRoomRealtime({
  room,
  me,
  currentUserId
})

const {
  myLinkmicId,
  myQueuePosition,
  activeLinkmics,
  pendingApplications,
  linkmicLoading,
  linkmicStreams,
  localLinkmicVideoRef,
  localLinkmicAudioEnabled,
  localLinkmicVideoEnabled,
  viewerHandRaised,
  loadLinkmicStatus,
  cleanupLinkmic,
  setRemoteVideo,
  applyLinkmic,
  acceptLinkmic,
  rejectLinkmic,
  disconnectLinkmic,
  viewerHangup,
  toggleViewerHandRaise,
  toggleLocalAudio,
  toggleLocalVideo,
  muteLinkmicAudio,
  muteLinkmicVideo,
  handleLinkmicApplyMessage,
  handleLinkmicAcceptedMessage,
  handleLinkmicRejectedMessage,
  handleLinkmicDisconnectedMessage,
  handleHandRaisedMessage,
  handleMuteAudioMessage,
  handleMuteVideoMessage,
  handlePeerStateMessage,
  handleOffer,
  handleAnswer,
  handleIceCandidate
} = useLiveLinkmic({
  roomId,
  room,
  isStreamer,
  currentUserId,
  sendRoomMessage
})

const {
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
} = useLiveRoomInteractions({
  room,
  me,
  currentUserId,
  sendRoomMessage,
  emitDanmaku
})

attachRoomMessageHandlers({
  handleLinkmicApplyMessage,
  handleLinkmicAcceptedMessage,
  handleLinkmicRejectedMessage,
  handleLinkmicDisconnectedMessage,
  handleHandRaisedMessage,
  receiveGift,
  handleOffer,
  handleAnswer,
  handleIceCandidate,
  receiveReaction,
  receivePinMessage,
  clearPinnedMessage,
  receiveDanmaku,
  handleMuteAudioMessage,
  handleMuteVideoMessage,
  handlePeerStateMessage
})

onMounted(() => {
  initializeLiveRoom({
    syncFollowState,
    initPlayer,
    loadLinkmicStatus,
    connectLiveRoomSocket
  })
})

onUnmounted(() => {
  cleanupLiveRoomPlayer()
  cleanupLiveRoomInteractions()
  cleanupLinkmic()
  cleanupLiveRoomRealtime()
})

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
              <div v-if="isOfflineRoomStatus(room.status)" class="stream-ended-overlay">
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
