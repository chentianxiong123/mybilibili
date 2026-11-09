<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *, auto">
      <GridLayout row="0" columns="auto, *, auto" class="chat-header">
        <Label text="◀" col="0" class="back-btn" @tap="onBack" />
        <Label :text="chatUserName || '聊天'" col="1" class="chat-title" />
      </GridLayout>

      <ScrollView row="1" class="messages-scroll">
        <StackLayout class="messages-list">
          <StackLayout v-if="loading" class="loading">
            <Label text="加载中..." />
          </StackLayout>
          <StackLayout v-else-if="messages.length > 0">
            <StackLayout v-for="(msg, i) in messages" :key="i" class="msg-item" :class="{ 'msg-self': msg.isSelf }">
              <StackLayout>
                <Label :text="msg.content || msg.text || ''" class="msg-bubble" :class="{ 'msg-bubble-self': msg.isSelf }" textWrap="true" />
                <Label :text="formatTimeLabel(msg.time || msg.createTime || msg.createdAt)" class="msg-time" />
              </StackLayout>
            </StackLayout>
          </StackLayout>
          <StackLayout v-else class="empty-state">
            <Label text="暂无消息，开始聊天吧" />
          </StackLayout>
        </StackLayout>
      </ScrollView>

      <GridLayout row="2" columns="*, auto" class="msg-input-bar">
        <TextField col="0" v-model="inputText" hint="输入消息..." class="msg-input" />
        <Button col="1" text="发送" class="send-btn" @tap="handleSend" />
      </GridLayout>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack } from 'nativescript-vue'
import { getChatMessages, sendMessage } from '../../api/message'
import { formatTimeLabel } from '../../utils/format'
import storage from '../../utils/storage'

const props = defineProps<{ userId: number }>()

const messages = ref<any[]>([])
const inputText = ref('')
const loading = ref(true)
const chatUserName = ref('')

onMounted(async () => {
  try {
    const res = await getChatMessages(props.userId)
    if (res.code === '1') {
      const data = res.data || []
      messages.value = data.map((m: any) => ({
        ...m,
        isSelf: m.isSelf || m.isMine || m.fromUserId === storage.getUser()?.id
      }))
      if (data.length > 0) {
        chatUserName.value = data[0]?.userName || data[0]?.nickname || ''
      }
    }
  } catch (e) {
    console.error('加载聊天消息失败:', e)
  } finally {
    loading.value = false
  }
})

function onBack() {
  $navigateBack()
}

async function handleSend() {
  const text = inputText.value.trim()
  if (!text) return
  try {
    const res = await sendMessage(props.userId, text)
    if (res.code === '1') {
      messages.value.push({
        content: text,
        isSelf: true,
        time: Date.now()
      })
      inputText.value = ''
    }
  } catch (e) {
    console.error('发送消息失败:', e)
  }
}
</script>

<style scoped>
.chat-header {
  padding: 8 12;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.back-btn {
  font-size: 20;
  color: #18191c;
  padding: 8 0;
  margin-right: 8;
}

.chat-title {
  font-size: 16;
  font-weight: bold;
  color: #18191c;
  text-align: center;
}

.messages-scroll {
  background-color: #f4f5f7;
}

.messages-list {
  padding: 12 16;
}

.msg-item {
  margin-bottom: 12;
  align-items: flex-start;
}

.msg-item.msg-self {
  align-items: flex-end;
}

.msg-bubble {
  font-size: 14;
  color: #18191c;
  background-color: white;
  padding: 10 14;
  border-radius: 12;
  max-width: 280;
  line-height: 1.4;
}

.msg-bubble-self {
  background-color: #fb7299;
  color: white;
}

.msg-time {
  font-size: 10;
  color: #9499a0;
  margin-top: 4;
}

.msg-input-bar {
  padding: 8 12;
  background-color: white;
  border-top-width: 1;
  border-top-color: #e3e5e7;
}

.msg-input {
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

.empty-state {
  padding: 60 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>