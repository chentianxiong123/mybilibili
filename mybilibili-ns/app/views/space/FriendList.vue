<template>
  <Page class="page">
    <ActionBar title="好友列表">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="friends.length > 0">
          <StackLayout v-for="(friend, i) in friends" :key="i" class="friend-item" @tap="onFriendTap(friend)">
            <GridLayout columns="auto, *, auto">
              <Image :src="friend.face || friend.avatar || ''" class="friend-avatar" col="0" />
              <StackLayout col="1" class="friend-info">
                <Label :text="friend.name || friend.uname || friend.nickname || ''" class="friend-name" />
                <Label :text="friend.sign || friend.description || ''" class="friend-sign" textWrap="true" />
              </StackLayout>
              <Label col="2" :text="isFollowingFriend(friend) ? '已关注' : '+ 关注'" class="friend-follow-btn" :class="{ following: isFollowingFriend(friend) }" @tap.stop="handleFollowFriend(friend)" />
            </GridLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无好友" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { formatCount } from '../../utils/format'
import storage from '../../utils/storage'
import { followUser, checkFollow } from '../../api/interaction'

const props = defineProps<{ mId: number }>()

const friends = ref<any[]>([])
const loading = ref(true)
const followingMap = ref<Record<number, boolean>>({})

onMounted(async () => {
  try {
    const res = await require('../../api/user').getFriendList(props.mId)
    if (res.code === '1') {
      friends.value = res.data || []
    }
  } catch (e) {
    console.error('加载好友列表失败:', e)
  } finally {
    loading.value = false
  }
})

function isFollowingFriend(friend: any): boolean {
  const uid = friend.mId || friend.id || friend.uid
  return !!followingMap.value[uid]
}

async function handleFollowFriend(friend: any) {
  const token = storage.getToken()
  if (!token) { $showModal(require('../Login.vue').default, { fullscreen: true }); return }
  const uid = friend.mId || friend.id || friend.uid
  if (!uid) return
  try {
    const res = await followUser(uid, !followingMap.value[uid])
    if (res.code === '1') {
      followingMap.value[uid] = !followingMap.value[uid]
    }
  } catch (e) {}
}

function onBack() {
  $navigateBack()
}

function onFriendTap(friend: any) {
  const uid = friend.mId || friend.id || friend.uid
  if (uid) {
    $showModal(require('./UpUser.vue').default, {
      fullscreen: true,
      props: { mId: uid }
    })
  }
}
</script>

<style scoped>
.friend-item {
  padding: 12 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.friend-avatar {
  width: 44;
  height: 44;
  border-radius: 22;
  margin-right: 12;
}

.friend-info {
  margin-left: 4;
}

.friend-name {
  font-size: 15;
  font-weight: 500;
  color: #18191c;
}

.friend-sign {
  font-size: 12;
  color: #9499a0;
  margin-top: 2;
  line-height: 1.3;
}

.friend-follow-btn {
  font-size: 12;
  background-color: #fb7299;
  color: white;
  padding: 5 14;
  border-radius: 14;
  font-weight: bold;
}

.friend-follow-btn.following {
  background-color: #f1f2f3;
  color: #61666d;
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