<template>
    <div class="message-view">
        <div class="message-bgc"></div>
        <div class="message-bg" :style="{ backgroundImage: `url(${messageBg})` }"></div>
        <div class="message-layout">
            <aside class="message-aside">
                <div class="message-sidebar">
                    <div class="message-sidebar__title">
                        <div class="message-sidebar__icon">
                            <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path d="M512 128C264.6 128 64 292.6 64 496c0 118.9 66.3 224.6 169.3 292.9-6.8 39.8-22.2 84.6-50.7 124.7-8.5 12-0.9 28.9 13.6 30.5 82.3 9 148.5-17.6 192.9-45.1 38.1 9.6 78.7 14.7 120.9 14.7 247.4 0 448-164.6 448-368S759.4 128 512 128z" fill="currentColor"></path></svg>
                        </div>
                        <span>消息中心</span>
                    </div>
                    <ul class="message-sidebar__interactions">
                        <li class="message-sidebar__item" :class="{'is-active': route.path.startsWith(item.path)}"
                            v-for="item in menu" :key="item.path"
                            @click="router.push(item.path)"
                        >
                            <div class="message-sidebar__item-icon dot"></div>
                            <div class="message-sidebar__item-name">{{ item.name }}</div>
                            <div class="message-sidebar__item-notify"
                                v-if="item.unread >= 0 && store.user?.uid && store.msgUnread[item.unread] > 0">
                                <div class="message-sidebar__item-num">
                                    {{ store.msgUnread[item.unread] <= 99 ? store.msgUnread[item.unread] : '99+' }}
                                </div>
                            </div>
                        </li>
                    </ul>
                    <div class="message-sidebar__div"></div>
                    <ul class="message-sidebar__settings">
                        <li class="message-sidebar__item" :class="{'is-active': routerIndex === SETTING_INDEX}"
                            @click="router.push('/message/config')">
                            <div class="message-sidebar__item-icon">
                                <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path d="M512 336a176 176 0 1 0 0 352 176 176 0 0 0 0-352zm0 288a112 112 0 1 1 0-224 112 112 0 0 1 0 224z" fill="currentColor"></path><path d="M899.1 469.3l-46.3-11.6a356 356 0 0 0-25-60.2l24.7-40.9a32 32 0 0 0-4.6-41.2l-45.3-45.3a32 32 0 0 0-41.2-4.6l-40.9 24.7a356 356 0 0 0-60.2-25l-11.6-46.3a32 32 0 0 0-31.1-24.9h-64a32 32 0 0 0-31.1 24.9l-11.6 46.3a356 356 0 0 0-60.2 25l-40.9-24.7a32 32 0 0 0-41.2 4.6L294 276.4a32 32 0 0 0-4.6 41.2l24.7 40.9a356 356 0 0 0-25 60.2l-46.3 11.6a32 32 0 0 0-24.9 31.1v64a32 32 0 0 0 24.9 31.1l46.3 11.6a356 356 0 0 0 25 60.2l-24.7 40.9a32 32 0 0 0 4.6 41.2l45.3 45.3a32 32 0 0 0 41.2 4.6l40.9-24.7a356 356 0 0 0 60.2 25l11.6 46.3a32 32 0 0 0 31.1 24.9h64a32 32 0 0 0 31.1-24.9l11.6-46.3a356 356 0 0 0 60.2-25l40.9 24.7a32 32 0 0 0 41.2-4.6l45.3-45.3a32 32 0 0 0 4.6-41.2l-24.7-40.9a356 356 0 0 0 25-60.2l46.3-11.6a32 32 0 0 0 24.9-31.1v-64a32 32 0 0 0-24.9-31.1zM512 672a160 160 0 1 1 0-320 160 160 0 0 1 0 320z" fill="currentColor"></path></svg>
                            </div>
                            <div class="message-sidebar__item-name">消息设置</div>
                        </li>
                    </ul>
                </div>
            </aside>
            <main class="message-main">
                <div class="message-header message-box-shadow">
                    <div class="message-header__leading">
                        <span class="message-header__title">{{ pageTitle }}</span>
                    </div>
                    <div class="message-header__actions"></div>
                </div>
                <div class="message-content">
                    <div class="message-content__wrapper">
                        <div class="message-content__inner">
                            <NuxtPage />
                        </div>
                    </div>
                </div>
            </main>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import messageBg from '@/assets/site/img/message-bg.png'
definePageMeta({ layout: 'simple' })

const route = useRoute()
const router = useRouter()
const store = useAppStore()

// 原版侧栏菜单：unread 为 store.msgUnread 数组下标，-1 表示无红点
const menu = [
  { path: '/message/whisper', name: '我的消息', unread: 4 },
  { path: '/message/reply', name: '回复我的', unread: 0 },
  { path: '/message/at', name: '@ 我的', unread: 1 },
  { path: '/message/love', name: '收到的赞', unread: 2 },
  { path: '/message/system', name: '系统通知', unread: 3 },
  { path: '/message/ai', name: 'AI客服', unread: -1 },
]
const SETTING_INDEX = 6

// 根据当前路由计算侧栏下标，未命中则为消息设置
const routerIndex = computed(() => {
  const i = menu.findIndex(item => route.path.startsWith(item.path))
  return i === -1 ? SETTING_INDEX : i
})
const pageTitle = computed(() => menu[routerIndex.value]?.name ?? '消息设置')

if (import.meta.client) {
  if (route.path === '/message') {
    navigateTo('/message/whisper', { replace: true })
  }
}

onMounted(() => {
  // 原版消息页：根节点透明，露出固定背景层
  document.querySelectorAll('#app').forEach((el) => {
    ;(el as HTMLElement).style.cssText = 'background-color: transparent;'
  })
})

onBeforeUnmount(() => {
  document.querySelectorAll('#app').forEach((el) => {
    ;(el as HTMLElement).style.cssText = ''
  })
})

// 未登录踢回首页（原版行为）
watch(() => store.isLogin, (curr) => {
  if (!curr) {
    router.push('/')
  }
})

// 父级 SSE：只管全站红点校准（unread_init/unread_counts）。
// whisper 子页另有一条流处理私信增量；两条流写同一份红点，幂等。
let sse: EventSource | null = null
function applyCounts(d: any) {
  if (!d) return
  store.msgUnread[0] = d.reply || 0
  store.msgUnread[1] = d.at || 0
  store.msgUnread[2] = d.like || 0
  store.msgUnread[3] = d.system || 0
  store.msgUnread[4] = d.private || 0
  if (store.msgUnread.length > 5) store.msgUnread[5] = d.dynamic || 0
}
function connectSSE() {
  if (sse || !store.isLogin || typeof EventSource === 'undefined') return
  try {
    const es = new EventSource('/sse/notification')
    es.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data)
        if (msg && (msg.type === 'unread_init' || msg.type === 'unread_counts')) {
          applyCounts(msg.data)
        }
      } catch (_) {
        // 坏帧忽略，下次事件会再校准
      }
    }
    sse = es
  } catch (_) {
    sse = null
  }
}
watch(() => store.isLogin, (curr) => {
  if (curr) connectSSE()
})
onMounted(() => {
  connectSSE()
})
onBeforeUnmount(() => {
  if (sse) {
    sse.close()
    sse = null
  }
})
</script>
