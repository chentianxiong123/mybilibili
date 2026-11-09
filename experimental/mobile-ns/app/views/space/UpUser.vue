<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *">
      <GridLayout row="0" columns="auto, *" class="up-header">
        <Label text="◀" col="0" class="back-btn" @tap="onBack" />
      </GridLayout>
      <ScrollView row="1">
        <StackLayout>
          <StackLayout v-if="loading" class="loading">
            <Label text="加载中..." />
          </StackLayout>
          <StackLayout v-else-if="userInfo">
            <StackLayout class="user-profile">
              <Image :src="userInfo.face || ''" class="user-avatar" />
              <Label :text="userInfo.name || userInfo.uname || ''" class="user-name" />
              <Label :text="'UID: ' + (userInfo.mid || userInfo.id || '')" class="user-uid" />
              <GridLayout columns="*, *" class="user-stats">
                <StackLayout col="0" class="stat-item">
                  <Label :text="formatCount(userInfo.followerCount || 0)" class="stat-value" />
                  <Label text="粉丝" class="stat-label" />
                </StackLayout>
                <StackLayout col="1" class="stat-item">
                  <Label :text="formatCount(userInfo.videoCount || 0)" class="stat-value" />
                  <Label text="视频" class="stat-label" />
                </StackLayout>
              </GridLayout>
              <Label v-if="userInfo.sign || userInfo.description" :text="userInfo.sign || userInfo.description" class="user-sign" textWrap="true" />
              <Label :text="isFollowing ? '已关注' : '+ 关注'" class="follow-btn" :class="{ following: isFollowing }" @tap="handleFollow" />
            </StackLayout>

            <StackLayout class="video-section">
              <Label text="投稿视频" class="section-title" />
              <StackLayout v-if="videos.length > 0">
                <StackLayout v-for="(v, i) in videos" :key="i" class="video-item" @tap="openVideo(v)">
                  <GridLayout columns="132, *">
                    <Image :src="v.pic || ''" class="video-cover" col="0" />
                    <StackLayout col="1" class="video-info">
                      <Label :text="v.title || ''" class="video-title" textWrap="true" />
                      <Label :text="'▶ ' + formatCount(v.play || 0)" class="video-stats" />
                      <Label :text="formatTimeLabel(v.ctime || v.createdAt)" class="video-time" />
                    </StackLayout>
                  </GridLayout>
                </StackLayout>
              </StackLayout>
              <Label v-else class="no-video" text="暂无投稿视频" />
            </StackLayout>
          </StackLayout>
          <StackLayout v-else class="empty-state">
            <Label text="用户不存在" />
          </StackLayout>
        </StackLayout>
      </ScrollView>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { getUpUserInfo, getUpUserVideos } from '../../api/up-user'
import { followUser, checkFollow } from '../../api/interaction'
import { formatCount, formatTimeLabel } from '../../utils/format'
import storage from '../../utils/storage'

const props = defineProps<{ mId: number }>()

const userInfo = ref<any>(null)
const videos = ref<any[]>([])
const loading = ref(true)
const isFollowing = ref(false)

onMounted(async () => {
  try {
    const [infoRes, videoRes] = await Promise.all([
      getUpUserInfo(props.mId),
      getUpUserVideos(props.mId)
    ])
    if (infoRes.code === '1') {
      userInfo.value = infoRes.data
    }
    if (videoRes.code === '1') {
      videos.value = videoRes.data || []
    }
    await loadFollowState()
  } catch (e) {
    console.error('加载UP主信息失败:', e)
  } finally {
    loading.value = false
  }
})

async function loadFollowState() {
  try {
    const token = storage.getToken()
    if (!token) return
    const res = await checkFollow(props.mId)
    if (res.code === '1') {
      isFollowing.value = typeof res.data === 'object' ? !!res.data.following : !!res.data
    }
  } catch (e) {}
}

async function handleFollow() {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  try {
    const res = await followUser(props.mId, !isFollowing.value)
    if (res.code === '1') {
      isFollowing.value = !isFollowing.value
    }
  } catch (e) {}
}

function onBack() {
  $navigateBack()
}

function openVideo(v: any) {
  const aId = v.aId || v.id || v.aid
  if (aId) {
    $showModal(require('../../views/video/Detail.vue').default, {
      fullscreen: true,
      props: { aId }
    })
  }
}
</script>

<style scoped>
.up-header {
  padding: 8 12;
  background-color: white;
}

.back-btn {
  font-size: 20;
  color: #18191c;
  padding: 8 0;
}

.user-profile {
  align-items: center;
  padding: 24 16;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.user-avatar {
  width: 72;
  height: 72;
  border-radius: 36;
  margin-bottom: 12;
}

.user-name {
  font-size: 20;
  font-weight: bold;
  color: #18191c;
}

.user-uid {
  font-size: 12;
  color: #9499a0;
  margin-top: 4;
}

.user-stats {
  margin-top: 16;
  width: 100%;
}

.stat-item {
  align-items: center;
}

.stat-value {
  font-size: 18;
  font-weight: bold;
  color: #18191c;
}

.stat-label {
  font-size: 12;
  color: #9499a0;
  margin-top: 2;
}

.user-sign {
  font-size: 13;
  color: #61666d;
  margin-top: 12;
  text-align: center;
  line-height: 1.4;
}

.follow-btn {
  font-size: 14;
  background-color: #fb7299;
  color: white;
  padding: 8 40;
  border-radius: 20;
  font-weight: bold;
  margin-top: 16;
}

.follow-btn.following {
  background-color: #f1f2f3;
  color: #61666d;
}

.video-section {
  padding: 12 8;
  background-color: white;
  margin-top: 8;
}

.section-title {
  font-size: 15;
  font-weight: bold;
  color: #18191c;
  padding: 0 8 10;
}

.video-item {
  padding: 8;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.video-cover {
  width: 120;
  height: 72;
  border-radius: 4;
}

.video-info {
  padding-left: 10;
}

.video-title {
  font-size: 14;
  color: #18191c;
  font-weight: 500;
  line-height: 1.3;
}

.video-stats {
  font-size: 12;
  color: #9499a0;
  margin-top: 6;
}

.video-time {
  font-size: 11;
  color: #9499a0;
  margin-top: 2;
}

.no-video {
  padding: 20;
  text-align: center;
  color: #9499a0;
  font-size: 13;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}

.empty-state {
  padding: 60 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>