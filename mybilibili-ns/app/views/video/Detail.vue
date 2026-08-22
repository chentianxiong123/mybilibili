<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *">
      <GridLayout row="0" class="player-wrapper" height="240">
        <WebView ref="playerWebView" src="~/assets/player.html" @loadStarted="onWebViewLoad" />
        <GridLayout class="player-overlay" @tap="onOverlayTap">
          <Label text="◀" class="back-btn" @tap="onBack" />
        </GridLayout>
      </GridLayout>
      <ScrollView row="1">
        <StackLayout class="detail-content">
          <StackLayout v-if="loading" class="loading">
            <Label text="加载中..." />
          </StackLayout>
          <StackLayout v-else-if="video">
            <GridLayout class="up-info-row" columns="auto, *, auto">
              <Image :src="video.uploader?.face || ''" class="up-avatar" col="0" />
              <StackLayout col="1" class="up-meta">
                <Label :text="video.author || ''" class="up-name" />
                <Label :text="followerCount + '粉丝'" class="up-fans" />
              </StackLayout>
              <Label col="2" :text="isFollowing ? '已关注' : '+ 关注'" class="follow-btn" :class="{ following: isFollowing }" @tap="handleFollow" />
            </GridLayout>

            <StackLayout class="video-main-details">
              <Label :text="video.title || ''" class="video-main-title" textWrap="true" />
              <GridLayout columns="auto, auto, auto" class="video-plays-meta">
                <Label :text="'▶ ' + formatCount(video.play)" col="0" />
                <Label :text="'💬 ' + (video.videoReview || 0)" col="1" />
                <Label :text="formatTimeLabel(video.ctime)" col="2" />
              </GridLayout>
              <Label v-if="video.description" :text="video.description" class="video-full-desc" textWrap="true" />
            </StackLayout>

            <GridLayout columns="*, *, *, *, *" class="interactions-actions-row">
              <StackLayout col="0" class="action-btn" :class="{ active: isLiked }" @tap="handleLike">
                <Label text="👍" class="action-icon" />
                <Label :text="String(likeCount)" class="action-num" />
              </StackLayout>
              <StackLayout col="1" class="action-btn" :class="{ active: isDisliked }" @tap="handleDislike">
                <Label text="👎" class="action-icon" />
                <Label text="不喜欢" class="action-num" />
              </StackLayout>
              <StackLayout col="2" class="action-btn" :class="{ active: isCoined }" @tap="handleCoin">
                <Label text="🪙" class="action-icon" />
                <Label :text="String(coinCount)" class="action-num" />
              </StackLayout>
              <StackLayout col="3" class="action-btn" :class="{ active: isStarred }" @tap="handleStar">
                <Label text="⭐" class="action-icon" />
                <Label :text="String(starCount)" class="action-num" />
              </StackLayout>
              <StackLayout col="4" class="action-btn" @tap="handleShare">
                <Label text="↪️" class="action-icon" />
                <Label :text="String(shareCount)" class="action-num" />
              </StackLayout>
            </GridLayout>

            <StackLayout class="comments-section">
              <Label text="评论" class="section-title" />
              <StackLayout v-for="c in comments" :key="c.rpid" class="comment-card">
                <GridLayout columns="auto, *" class="comment-header">
                  <Image :src="c.user?.face || ''" class="comment-avatar" col="0" />
                  <StackLayout col="1">
                    <Label :text="c.user?.name || ''" class="comment-name" />
                    <Label :text="c.content || ''" class="comment-text" textWrap="true" />
                  </StackLayout>
                </GridLayout>
              </StackLayout>
              <Label v-if="hasMoreComments" text="点击加载更多评论" class="load-more" @tap="loadMoreComments" />
              <Label v-else-if="comments.length > 0" text="— 已经到底啦 —" class="no-more" />
            </StackLayout>
          </StackLayout>
          <StackLayout v-else class="loading">
            <Label text="视频不存在" />
          </StackLayout>
        </StackLayout>
      </ScrollView>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'nativescript-vue'
import { WebView } from '@nativescript/core'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { getVideoInfo, getRecommendVides, getComments, getBarrages } from '../../api/video'
import { followUser, checkFollow, likeManuscript, coinManuscript, collectManuscript, shareManuscript, getInteractionStatus } from '../../api/interaction'
import { postComment } from '../../api/comment'
import storage from '../../utils/storage'
import { formatCount, formatTimeLabel } from '../../utils/format'

const props = defineProps<{ aId: number }>()

const playerWebView = ref<any>(null)
const video = ref<any>(null)
const comments = ref<any[]>([])
const loading = ref(true)
const commentPage = ref(1)
const hasMoreComments = ref(true)
const isLiked = ref(false)
const isDisliked = ref(false)
const isCoined = ref(false)
const isStarred = ref(false)
const isFollowing = ref(false)
const likeCount = ref(0)
const coinCount = ref(0)
const starCount = ref(0)
const shareCount = ref(0)
const followerCount = ref(0)
let playerReady = false

onMounted(async () => {
  await loadData()
  await loadInteractionState()
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
  if (url.startsWith('native://')) {
    const parts = url.replace('native://', '').split('?')
    const action = parts[0]
    const params = new URLSearchParams(parts[1] || '')
    handleNativeEvent(action, params)
    return
  }
  if (url.includes('player.html') && video.value) {
    setTimeout(() => initPlayer(), 500)
  }
}

function handleNativeEvent(action: string, params: URLSearchParams) {
  switch (action) {
    case 'ready':
      playerReady = true
      break
    case 'play':
      break
    case 'pause':
      break
    case 'ended':
      break
    case 'timeupdate':
      break
  }
}

function initPlayer() {
  const wv = playerWebView.value?.nativeView as WebView
  if (!wv || !video.value) return
  const v = video.value
  const videoData = {
    aId: v.aId,
    playUrl: v.playUrl || '',
    playUrlHd: v.playUrlHd || '',
    playUrlSd: v.playUrlSd || '',
    playUrlLd: v.playUrlLd || '',
    pic: v.pic || ''
  }
  wv.evaluateJavaScript(`initVideo(${JSON.stringify(videoData)})`)
  loadBarrages()
}

async function loadBarrages() {
  if (!props.aId) return
  try {
    const res = await getBarrages(props.aId)
    if (res.code === '1' && res.data) {
      const barrages = res.data.map((b: any) => ({
        text: b.text || b.content,
        time: parseFloat(b.time) || 0,
        color: b.color || '#ffffff'
      }))
      const wv = playerWebView.value?.nativeView as WebView
      if (wv) {
        wv.evaluateJavaScript(`setBarrages(${JSON.stringify(barrages)})`)
      }
    }
  } catch (e) {}
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getVideoInfo(props.aId)
    if (res.code === '1' && res.data) {
      video.value = res.data
      const d = res.data
      likeCount.value = d.likeCount || 0
      coinCount.value = d.coinCount || 0
      starCount.value = d.collectCount || 0
      shareCount.value = d.shareCount || 0
      followerCount.value = d.uploader?.followerCount || 0
      storage.setViewHistory({ aId: d.aId, title: d.title, pic: d.pic, viewAt: Date.now() })
    }
    await loadComments()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadComments = async () => {
  try {
    const res = await getComments(props.aId, commentPage.value)
    if (res.code === '1' && res.data?.length > 0) {
      if (commentPage.value === 1) {
        comments.value = res.data
      } else {
        comments.value.push(...res.data)
      }
      hasMoreComments.value = res.data.length >= 20
    } else if (commentPage.value === 1) {
      comments.value = []
      hasMoreComments.value = false
    }
  } catch (e) {}
}

const loadMoreComments = () => {
  commentPage.value++
  loadComments()
}

const loadInteractionState = async () => {
  try {
    const token = storage.getToken()
    if (!token) return
    const res = await getInteractionStatus(props.aId)
    if (res.code === '1' && res.data) {
      const d = res.data
      isLiked.value = !!d.liked || !!d.isLiked
      isCoined.value = (d.coinCount || 0) > 0 || !!d.coined
      isStarred.value = !!d.collected || !!d.isCollected
    }
  } catch (e) {}
}

const loadFollowState = async () => {
  try {
    const token = storage.getToken()
    if (!token || !video.value?.mid) return
    const res = await checkFollow(video.value.mid)
    if (res.code === '1') {
      isFollowing.value = typeof res.data === 'object' ? !!res.data.following : !!res.data
    }
  } catch (e) {}
}

const handleLike = async () => {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  try {
    const res = await likeManuscript(props.aId, !isLiked.value)
    if (res.code === '1') {
      isLiked.value = !isLiked.value
      likeCount.value += isLiked.value ? 1 : -1
      if (isDisliked.value && isLiked.value) isDisliked.value = false
    }
  } catch (e) {}
}

const handleDislike = () => {
  isDisliked.value = !isDisliked.value
  if (isDisliked.value && isLiked.value) {
    isLiked.value = false
    likeCount.value = Math.max(0, likeCount.value - 1)
  }
}

const handleCoin = async () => {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  if (isCoined.value) return
  try {
    const res = await coinManuscript(props.aId, 1)
    if (res.code === '1') {
      isCoined.value = true
      coinCount.value++
    }
  } catch (e) {}
}

const handleStar = async () => {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  try {
    const res = await collectManuscript(props.aId, !isStarred.value)
    if (res.code === '1') {
      isStarred.value = !isStarred.value
      starCount.value += isStarred.value ? 1 : -1
    }
  } catch (e) {}
}

const handleShare = async () => {
  try {
    await shareManuscript(props.aId)
    shareCount.value++
  } catch (e) {}
}

const handleFollow = async () => {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  if (!video.value?.mid) return
  try {
    const res = await followUser(video.value.mid, !isFollowing.value)
    if (res.code === '1') {
      isFollowing.value = !isFollowing.value
      followerCount.value += isFollowing.value ? 1 : -1
    }
  } catch (e) {}
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
}

.back-btn {
  font-size: 20;
  color: white;
  padding: 8;
}

.detail-content {
  background-color: #f4f5f7;
}

.up-info-row {
  padding: 12 16;
  background-color: white;
}

.up-avatar {
  width: 36;
  height: 36;
  border-radius: 18;
  margin-right: 12;
}

.up-meta {
  margin-left: 8;
}

.up-name {
  font-size: 13;
  font-weight: bold;
  color: #18191c;
}

.up-fans {
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

.video-main-details {
  padding: 0 16;
  background-color: white;
}

.video-main-title {
  font-size: 15;
  font-weight: bold;
  color: #18191c;
  line-height: 1.4;
  margin-bottom: 8;
}

.video-plays-meta {
  font-size: 11;
  color: #9499a0;
  margin-bottom: 12;
}

.video-full-desc {
  font-size: 12;
  color: #61666d;
  line-height: 1.5;
  padding-bottom: 12;
}

.interactions-actions-row {
  padding: 16 10;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.action-btn {
  align-items: center;
  gap: 6;
}

.action-icon {
  font-size: 20;
  color: #61666d;
}

.action-num {
  font-size: 11;
  color: #9499a0;
}

.action-btn.active .action-icon,
.action-btn.active .action-num {
  color: #fb7299;
  font-weight: bold;
}

.comments-section {
  padding: 14 16;
  background-color: white;
  margin-top: 8;
}

.section-title {
  font-size: 14;
  font-weight: bold;
  color: #18191c;
  margin-bottom: 14;
}

.comment-card {
  margin-bottom: 12;
  padding-bottom: 12;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.comment-header {
  gap: 12;
}

.comment-avatar {
  width: 34;
  height: 34;
  border-radius: 17;
}

.comment-name {
  font-size: 12;
  font-weight: bold;
  color: #61666d;
}

.comment-text {
  font-size: 13;
  color: #18191c;
  line-height: 1.5;
  margin-top: 4;
}

.load-more {
  text-align: center;
  padding: 16;
  font-size: 13;
  color: #fb7299;
}

.no-more {
  text-align: center;
  padding: 24;
  font-size: 11;
  color: #9499a0;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}
</style>