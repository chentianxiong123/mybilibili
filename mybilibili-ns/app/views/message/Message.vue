<template>
  <Page class="page">
    <ActionBar title="消息">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="chatList.length > 0">
          <StackLayout v-for="(chat, i) in chatList" :key="i" class="chat-item" @tap="onChatTap(chat)">
            <GridLayout columns="auto, *, auto">
              <Image :src="chat.avatar || chat.face || ''" class="chat-avatar" col="0" />
              <StackLayout col="1" class="chat-info">
                <Label :text="chat.name || chat.nickname || chat.userName || ''" class="chat-name" />
                <Label :text="chat.lastMessage || chat.lastMsg || chat.content || ''" class="chat-last-msg" textWrap="true" />
              </StackLayout>
              <StackLayout col="2" class="chat-meta">
                <Label v-if="chat.lastTime" :text="formatTimeLabel(chat.lastTime)" class="chat-time" />
                <Label v-if="chat.unread > 0" :text="String(chat.unread)" class="chat-unread" />
              </StackLayout>
            </GridLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无消息" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $navigateTo } from 'nativescript-vue'
import { getChatList } from '../../api/message'
import { formatTimeLabel } from '../../utils/format'

const chatList = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getChatList()
    if (res.code === '1') {
      chatList.value = res.data || []
    }
  } catch (e) {
    console.error('加载消息列表失败:', e)
  } finally {
    loading.value = false
  }
})

function onBack() {
  $navigateBack()
}

function onChatTap(chat: any) {
  const userId = chat.userId || chat.id || chat.uid
  if (userId) {
    $navigateTo(require('./Chat.vue').default, {
      props: { userId }
    })
  }
}
</script>

<style scoped>
.chat-item {
  padding: 12 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.chat-avatar {
  width: 44;
  height: 44;
  border-radius: 22;
  margin-right: 12;
}

.chat-info {
  margin-left: 4;
}

.chat-name {
  font-size: 15;
  font-weight: 500;
  color: #18191c;
}

.chat-last-msg {
  font-size: 13;
  color: #9499a0;
  margin-top: 4;
  line-height: 1.3;
}

.chat-meta {
  align-items: flex-end;
}

.chat-time {
  font-size: 11;
  color: #9499a0;
}

.chat-unread {
  font-size: 11;
  background-color: #fb7299;
  color: white;
  width: 20;
  height: 20;
  border-radius: 10;
  text-align: center;
  margin-top: 4;
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