<template>
  <Page class="page">
    <ActionBar title="我的" />
    <ScrollView>
      <StackLayout>
        <GridLayout class="user-profile" columns="auto, *" @tap="onLoginTap">
          <Image :src="user?.face || ''" class="profile-avatar" col="0" />
          <StackLayout col="1" class="profile-info">
            <Label :text="isLoggedIn ? (user?.name || '用户') : '点击登录'" class="profile-name" />
            <Label v-if="isLoggedIn" :text="'UID: ' + (user?.id || '')" class="profile-uid" />
          </StackLayout>
        </GridLayout>

        <StackLayout class="menu-section">
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onHistoryTap">
            <Label text="📋" class="menu-icon" col="0" />
            <Label text="历史记录" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onFavoriteTap">
            <Label text="⭐" class="menu-icon" col="0" />
            <Label text="我的收藏" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onManuscriptTap">
            <Label text="🎬" class="menu-icon" col="0" />
            <Label text="稿件管理" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onMessageTap">
            <Label text="💬" class="menu-icon" col="0" />
            <Label text="消息" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onCreatorTap">
            <Label text="🎨" class="menu-icon" col="0" />
            <Label text="创作中心" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
        </StackLayout>

        <Button v-if="isLoggedIn" text="退出登录" class="logout-btn" @tap="handleLogout" />
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'nativescript-vue'
import { $navigateTo, $showModal } from 'nativescript-vue'
import storage from '../../utils/storage'

const user = ref<any>(null)
const isLoggedIn = computed(() => !!user.value)

onMounted(() => {
  loadUser()
})

function loadUser() {
  const u = storage.getUser()
  if (u) user.value = u
}

function onLoginTap() {
  if (!isLoggedIn.value) {
    $showModal(require('../Login.vue').default, { fullscreen: true })
  } else {
    $showModal(require('./ProfileEdit.vue').default, { fullscreen: true })
  }
}

function onHistoryTap() {
  $navigateTo(require('./History.vue').default)
}

function onFavoriteTap() {
  $navigateTo(require('./Favorite.vue').default)
}

function onManuscriptTap() {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  $navigateTo(require('./ManuscriptManage.vue').default)
}

function onMessageTap() {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  $navigateTo(require('../message/Message.vue').default)
}

function onCreatorTap() {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  $navigateTo(require('../creator/Center.vue').default)
}

function handleLogout() {
  storage.removeToken()
  storage.removeItem('user')
  user.value = null
}
</script>

<style scoped>
.user-profile {
  padding: 20 16;
  background-color: white;
  margin-bottom: 8;
}

.profile-avatar {
  width: 56;
  height: 56;
  border-radius: 28;
  margin-right: 14;
}

.profile-info {
  margin-left: 4;
}

.profile-name {
  font-size: 18;
  font-weight: bold;
  color: #18191c;
}

.profile-uid {
  font-size: 12;
  color: #9499a0;
  margin-top: 4;
}

.menu-section {
  background-color: white;
}

.menu-item {
  padding: 14 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.menu-icon {
  font-size: 18;
  margin-right: 12;
  width: 28;
}

.menu-label {
  font-size: 14;
  color: #18191c;
}

.menu-arrow {
  font-size: 18;
  color: #9499a0;
}

.logout-btn {
  margin: 30 16;
  height: 44;
  background-color: #f1f2f3;
  color: #f44336;
  font-size: 15;
  border-radius: 22;
}
</style>