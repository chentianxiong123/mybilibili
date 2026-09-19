<template>
    <div id="app">
        <NuxtPage />
        <div class="loading-mark" :class="isMarkShow ? 'show' : 'hide'" :style="`display: ${markDisplay};`">
            <div class="loading-box">
                <img src="@/assets/teriteri/img/loading.gif" alt="">
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useTeriteriStore } from '@/stores/teriteri'

const store = useTeriteriStore()
const markDisplay = ref('none')
const isMarkShow = ref(false)

function show() {
  markDisplay.value = ''
  isMarkShow.value = true
}

function hide() {
  isMarkShow.value = false
  setTimeout(() => {
    markDisplay.value = 'none'
  }, 200)
}

async function getChannels() {
  const { get } = await import('@/teriteri-src/network/request')
  const res = await get('/category/getall')
  store.updateChannels(res.data.data)
}

async function getHotSearch() {
  const { get } = await import('@/teriteri-src/network/request')
  const res = await get('/search/hot/get')
  store.updateTrendings(res.data.data)
}

async function initIMServer() {
  await store.connectWebSocket()
  const connection = JSON.stringify({
    code: 100,
    content: 'Bearer ' + (typeof window !== 'undefined' ? localStorage.getItem('teri_token') : '')
  })
  if (store.ws) store.ws.send(connection)
}

async function closeIMWebSocket() {
  await store.closeWebSocket()
}

async function getFavorites() {
  const { get } = await import('@/teriteri-src/network/request')
  const res = await get('/favorite/get-all/user', {
    params: { uid: store.user.uid },
    headers: { Authorization: 'Bearer ' + localStorage.getItem('teri_token') }
  })
  if (!res.data) return
  const defaultFav = res.data.data.find(item => item.type === 1)
  const list = res.data.data.filter(item => item.type !== 1)
  if (defaultFav) list.unshift(defaultFav)
  store.updateFavorites(list)
}

async function getLikeAndDisLikeComment() {
  const { get } = await import('@/teriteri-src/network/request')
  const res = await get('/comment/get-like-and-dislike', {
    params: { uid: store.user.uid },
    headers: { Authorization: 'Bearer ' + localStorage.getItem('teri_token') }
  })
  if (!res.data) return
  store.updateLikeComment(res.data.data.userLike)
  store.updateDislikeComment(res.data.data.userDislike)
}

onMounted(async () => {
  if (typeof window === 'undefined') return
  if (localStorage.getItem('teri_token')) {
    await store.getPersonalInfo()
    await initIMServer()
    await getFavorites()
    try { await getLikeAndDisLikeComment() } catch { /* 404 route not found */ }
  }
  getChannels()
  getHotSearch()
  window.addEventListener('beforeunload', closeIMWebSocket)
})

onBeforeUnmount(async () => {
  await closeIMWebSocket()
  if (typeof window !== 'undefined') window.removeEventListener('beforeunload', closeIMWebSocket)
})

watch(() => store.isLoading, (current) => {
  if (current) show()
  else hide()
})
</script>

<style>
#app {
    margin: 0 auto;
    max-width: 2560px;
    background-color: var(--bg1);
}

.loading-mark {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 50000;
    user-select: none;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
}

.loading-box {
    display: flex;
    height: 100vh;
    width: 100vw;
    align-items: center;
    justify-content: center;
}

.loading-box img {
    max-height: 33vh;
    max-width: 33vw;
}

.hide {
    animation: fade-out 0.2s ease-out forwards;
}

.show {
    animation: fade-in 0.2s ease-out forwards;
}

@keyframes fade-in {
    0% { opacity: 0; }
    100% { opacity: 1; }
}

@keyframes fade-out {
    0% { opacity: 1; }
    100% { opacity: 0; }
}
</style>
