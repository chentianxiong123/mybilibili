<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import noface from '../../assets/noface.gif'

const route = useRoute()
const router = useRouter()
const draft = ref('')
const localMessages = ref<any[]>([])

const title = computed(() => String(route.query.name || route.params.id || '私信'))
const avatar = computed(() => String(route.query.avatar || ''))
const firstTime = computed(() => String(route.query.time || '2026年5月27日 22:08'))
const lastMessage = computed(() => String(route.query.last || '欢迎关注，我会在这里接收你的私信。'))

const send = () => {
  const text = draft.value.trim()
  if (!text) return
  localMessages.value.push({
    id: Date.now(),
    content: text,
    createdAt: new Date()
  })
  draft.value = ''
}

const formatTime = (value: any) => {
  const d = value instanceof Date ? value : new Date(String(value).replace(/-/g, '/'))
  if (Number.isNaN(d.getTime())) return String(value || '')
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<template>
  <div class="chat-page">
    <header class="chat-topbar">
      <button @click="router.back()" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1>{{ title }}</h1>
      <button aria-label="设置">
        <svg viewBox="0 0 24 24"><path d="M12 2l8 4v12l-8 4-8-4V6z" /><circle cx="12" cy="12" r="3" /></svg>
      </button>
    </header>

    <main class="chat-body">
      <time>{{ firstTime }}</time>
      <section class="message-row other">
        <img :src="avatar || noface" alt="" />
        <div class="bubble">
          <p>{{ lastMessage }}</p>
          <i>此条消息为自动回复</i>
        </div>
      </section>

      <time>2026年5月27日 22:32</time>
      <p class="notice">对方主动回复或关注你前，最多发送1条消息</p>

      <template v-for="msg in localMessages" :key="msg.id">
        <time>{{ formatTime(msg.createdAt) }}</time>
        <section class="message-row mine">
          <div class="bubble">{{ msg.content }}</div>
        </section>
      </template>
    </main>

    <footer class="chat-inputbar">
      <button aria-label="图片">
        <svg viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="16" rx="2" /><circle cx="8" cy="9" r="2" /><path d="M21 16l-5-5-5 5-2-2-6 6" /></svg>
      </button>
      <input v-model="draft" placeholder="发个消息聊聊呗" @keydown.enter="send" />
      <button class="send-btn" @click="send" aria-label="发送">
        <svg v-if="!draft.trim()" viewBox="0 0 24 24"><path d="M9 14s1.5 2 3 2 3-2 3-2" /><path d="M9 9h.01M15 9h.01" /><circle cx="12" cy="12" r="10" /></svg>
        <span v-else>发</span>
      </button>
    </footer>
  </div>
</template>

<style scoped lang="scss">
.chat-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f6f7f9;
  color: #18191c;
}

.chat-topbar {
  height: 74px;
  padding: 26px 18px 0;
  display: grid;
  grid-template-columns: 42px 1fr 42px;
  align-items: center;
  background: #f6f7f9;
  box-sizing: border-box;

  h1 {
    margin: 0;
    text-align: center;
    font-size: 23px;
    font-weight: 700;
  }

  button {
    border: 0;
    padding: 0;
    background: transparent;
    color: #61666d;
  }

  svg {
    width: 34px;
    height: 34px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.1;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.chat-body {
  flex: 1;
  padding: 18px 28px 92px;
  overflow-y: auto;

  > time {
    display: block;
    margin: 12px 0 26px;
    text-align: center;
    color: #9499a0;
    font-size: 18px;
  }
}

.message-row {
  display: flex;
  gap: 14px;

  img {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    object-fit: cover;
    flex: 0 0 auto;
  }

  .bubble {
    max-width: min(72vw, 620px);
    border-radius: 4px 18px 18px 18px;
    background: #fff;
    color: #18191c;
    font-size: 23px;
    line-height: 1.45;
    box-shadow: 0 2px 10px rgba(0,0,0,0.03);
  }

  p {
    margin: 0;
    padding: 18px 20px;
  }

  i {
    display: block;
    margin: 0 20px;
    padding: 14px 0;
    border-top: 1px solid #f1f2f3;
    color: #9499a0;
    font-size: 16px;
    font-style: normal;
  }

  &.mine {
    justify-content: flex-end;

    .bubble {
      padding: 12px 16px;
      border-radius: 18px 4px 18px 18px;
      background: #fb7299;
      color: #fff;
      font-size: 18px;
    }
  }
}

.notice {
  margin: -8px 0 36px;
  text-align: center;
  color: #a5a9b1;
  font-size: 17px;
}

.chat-inputbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 78px;
  padding: 10px 16px calc(10px + env(safe-area-inset-bottom));
  display: grid;
  grid-template-columns: 42px 1fr 42px;
  align-items: center;
  gap: 10px;
  background: #fff;
  box-sizing: content-box;

  button {
    border: 0;
    padding: 0;
    background: transparent;
    color: #61666d;
  }

  svg {
    width: 31px;
    height: 31px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  input {
    min-width: 0;
    height: 48px;
    border: 0;
    border-radius: 25px;
    padding: 0 18px;
    outline: none;
    background: #f1f2f3;
    color: #18191c;
    font-size: 18px;
  }

  .send-btn span {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: #fff;
    background: #fb7299;
    font-size: 15px;
  }
}

@media (max-width: 390px) {
  .chat-body {
    padding-inline: 20px;
  }

  .message-row .bubble {
    font-size: 20px;
  }
}
</style>
