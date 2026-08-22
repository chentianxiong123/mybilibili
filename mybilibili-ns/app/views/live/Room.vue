<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *, auto">
      <GridLayout row="0" class="player-wrapper" height="240">
        <WebView ref="playerWebView" src="~/assets/player.html" @loadStarted="onWebViewLoad" />
        <GridLayout class="player-overlay" @tap="onOverlayTap">
          <Label text="◀" class="back-btn" @tap="onBack" />
          <Label :text="'🔴 LIVE'" class="live-badge" />
        </GridLayout>
      </GridLayout>

      <ScrollView row="1">
        <StackLayout class="room-content">
          <StackLayout v-if="loading" class="loading">
            <Label text="加载中..." />
          </StackLayout>
          <StackLayout v-else-if="roomInfo">
            <Label :text="roomInfo.title || ''" class="room-title" textWrap="true" />
            <GridLayout columns="auto, *, auto" class="host-row">
              <Image :src="roomInfo.face || roomInfo.avatar || ''" class="host-avatar" col="0" />
              <StackLayout col="1" class="host-meta">
                <Label :text="roomInfo.uname || roomInfo.nickname || ''" class="host-name" />
                <Label :text="'粉丝: ' + formatTenThousand(roomInfo.followerCount || 0)" class="host-fans" />
              </StackLayout>
              <Label col="2" :text="isFollowing ? '已关注' : '+ 关注'" class="follow-btn" :class="{ following: isFollowing }" @tap="handleFollow" />
            </GridLayout>

            <Label text="聊天" class="section-title" />
            <StackLayout class="chat-area">
              <StackLayout v-for="(msg, i) in chatMessages" :key="i" class="chat-msg">
                <GridLayout columns="auto, *">
                  <Label :text="msg.nickname || msg.userName || ''" col="0" class="chat-name" />
                  <Label :text="msg.text || msg.content || ''" col="1" class="chat-text" textWrap="true" />
                </GridLayout>
              </StackLayout>
              <Label v-if="chatMessages.length === 0" text="暂无聊天消息" class="chat-empty" />
            </StackLayout>
          </StackLayout>
          <StackLayout v-else class="loading">
            <Label text="直播间不存在" />
          </StackLayout>
        </StackLayout>
      </ScrollView>

      <GridLayout row="2" columns="*, auto" class="chat-input-bar">
        <TextField col="0" v-model="chatInput" hint="发送消息..." class="chat-input" />
        <Button col="1" text="发送" class="send-btn" @tap="handleSendChat" />
      </GridLayout>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { WebView } from '@nativescript/core'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { getRoomInfo } from '../../api/live'
import { formatTenThousand } from '../../utils/format'
import storage from '../../utils/storage'
import { followUser, checkFollow } from '../../api/interaction'

const props = defineProps<{ roomId: string | number }>()

const playerWebView = ref<any>(null)
const roomInfo = ref<any>(null)
const loading = ref(true)
const isFollowing = ref(false)
const chatMessages = ref<any[]>([])
const chatInput = ref('')
let playerReady = false

onMounted(async () => {
  await loadRoomData()
  await loadFollowState()
})

function onBack() {
  $navigateBack()
}

function onOverlayTap() {
}

function onWebViewLoad(args: any) {
  if (playerReady) return
  const url = args.url || ''
  if (url.includes('player.html') && roomInfo.value) {
    setTimeout(() => initPlayer(), 500)
  }
}

function initPlayer() {
  const wv = playerWebView.value?.nativeView as WebView
  if (!wv || !roomInfo.value) return
  const r = roomInfo.value
  const liveData = {
    roomId: r.roomId || r.room_id || props.roomId,
    isLive: true,
    playUrl: r.playUrl || '',
    title: r.title || ''
  }
  wv.evaluateJavaScript(`initVideo(${JSON.stringify(liveData)})`)
  playerReady = true
}

async function loadRoomData() {
  loading.value = true
  try {
    const res = await getRoomInfo(props.roomId)
    if (res.code === '1' && res.data) {
      roomInfo.value = res.data
    }
  } catch (e) {
    console.error('加载直播间失败:', e)
  } finally {
    loading.value = false
  }
}

async function loadFollowState() {
  try {
    const token = storage.getToken()
    if (!token || !roomInfo.value?.mid) return
    const res = await checkFollow(roomInfo.value.mid)
    if (res.code === '1') {
      isFollowing.value = typeof res.data === 'object' ? !!res.data.following : !!res.data
    }
  } catch (e) {}
}

async function handleFollow() {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  if (!roomInfo.value?.mid) return
  try {
    const res = await followUser(roomInfo.value.mid, !isFollowing.value)
    if (res.code === '1') {
      isFollowing.value = !isFollowing.value
    }
  } catch (e) {}
}

function handleSendChat() {
  const text = chatInput.value.trim()
  if (!text) return
  chatMessages.value.push({
    nickname: '我',
    text: text
  })
  chatInput.value = ''
}
</script>

<style scoped>
.player-wrapper {
  position: relative;
  background-color: #000;
}

.player-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 44;
  padding: 0 12;
  flex-direction: row;
  align-items: center;
}

.back-btn {
  font-size: 20;
  color: white;
  padding: 8;
}

.live-badge {
  font-size: 12;
  color: white;
  background-color: #fb7299;
  padding: 2 8;
  border-radius: 4;
  margin-left: 8;
}

.room-content {
  background-color: #f4f5f7;
}

.room-title {
  font-size: 15;
  font-weight: bold;
  color: #18191c;
  padding: 12 16;
  background-color: white;
  line-height: 1.4;
}

.host-row {
  padding: 12 16;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.host-avatar {
  width: 36;
  height: 36;
  border-radius: 18;
  margin-right: 12;
}

.host-meta {
  margin-left: 4;
}

.host-name {
  font-size: 14;
  font-weight: bold;
  color: #18191c;
}

.host-fans {
  font-size: 11;
  color: #9499a0;
  margin-top: 2;
}

.follow-btn {
  font-size: 12;
  background-color: #fb7299;
  color: white;
  padding: 5 16;
  border-radius: 14;
  font-weight: bold;
}

.follow-btn.following {
  background-color: #f1f2f3;
  color: #61666d;
}

.section-title {
  font-size: 14;
  font-weight: bold;
  color: #18191c;
  padding: 12 16 8;
  background-color: white;
  margin-top: 8;
}

.chat-area {
  padding: 8 16;
  background-color: white;
  min-height: 200;
}

.chat-msg {
  padding: 6 0;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.chat-name {
  font-size: 12;
  font-weight: bold;
  color: #fb7299;
  margin-right: 8;
  width: 60;
}

.chat-text {
  font-size: 13;
  color: #18191c;
}

.chat-empty {
  font-size: 12;
  color: #9499a0;
  text-align: center;
  padding: 20;
}

.chat-input-bar {
  padding: 8 12;
  background-color: white;
  border-top-width: 1;
  border-top-color: #e3e5e7;
}

.chat-input {
  height: 40;
  padding: 0 12;
  font-size: 14;
  border-width: 1;
  border-color: #e3e5e7;
  border-radius: 20;
  background-color: #f8f9fa;
  margin-right: 8;
}

.send-btn {
  height: 40;
  background-color: #fb7299;
  color: white;
  font-size: 14;
  font-weight: bold;
  border-radius: 20;
  padding: 0 20;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}
</style>