import { computed, reactive, ref } from 'vue'
import { createReconnectingWS } from '../utils/reconnectingWs.js'
import { sameUserId } from '../utils/userId.js'

const VIEWER_LEVELS = [
  { min: 0, label: '新人', color: '#909399', icon: '🌱' },
  { min: 11, label: '常客', color: '#67c23a', icon: '🌿' },
  { min: 51, label: '活跃', color: '#00a1d6', icon: '🌳' },
  { min: 201, label: '忠实', color: '#e6a23c', icon: '🏅' },
  { min: 501, label: '贵族', color: '#f56c6c', icon: '👑' }
]

export function useLivePushAudienceMonitor({
  streamKey,
  me,
  createWs = createReconnectingWS,
  getWsUrl = () => {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${proto}//${location.host}/ws/meeting`
  },
  now = () => Date.now(),
  random = () => Math.random(),
  setTimer = (handler, delay) => setInterval(handler, delay),
  clearTimer = timerId => clearInterval(timerId)
}) {
  const viewers = ref([])
  const recentChats = ref([])
  const danmakuRate = ref(0)
  const peakViewers = ref(0)
  const totalDanmaku = ref(0)
  const viewerDanmakuCount = reactive({})
  const leaderboard = computed(() => {
    return Object.entries(viewerDanmakuCount)
      .map(([userId, count]) => {
        const v = viewers.value.find(u => sameUserId(u.userId, userId))
        return {
          userId: Number(userId),
          userName: v?.userName || ('用户' + userId),
          count
        }
      })
      .filter(x => !sameUserId(x.userId, me.id))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  })

  const getViewerLevel = (userId) => {
    const count = viewerDanmakuCount[String(userId)] || 0
    for (let i = VIEWER_LEVELS.length - 1; i >= 0; i--) {
      if (count >= VIEWER_LEVELS[i].min) return VIEWER_LEVELS[i]
    }
    return VIEWER_LEVELS[0]
  }

  let ws = null
  let metricsTimer = null

  const updateMetrics = () => {
    const cutoff = now() - 60000
    danmakuRate.value = recentChats.value.filter(m => m.id > cutoff).length
    if (viewers.value.length > peakViewers.value) peakViewers.value = viewers.value.length
    totalDanmaku.value = recentChats.value.length
  }

  const ensureMetricsTimer = () => {
    if (metricsTimer) return
    metricsTimer = setTimer(updateMetrics, 1000)
  }

  const connectAudienceWs = () => {
    if (!streamKey.value) return false
    if (ws) {
      try { ws.close() } catch (e) {}
      ws = null
    }
    ws = createWs({
      url: getWsUrl(),
      onOpen: () => {
        ws.send({
          type: 'join',
          roomCode: streamKey.value,
          userId: me.id,
          userName: me.name + '（主播）'
        })
      },
      onMessage: (msg) => {
        switch (msg.type) {
          case 'room-users':
            viewers.value = (msg.data || []).map(u => ({
              userId: u.userId,
              userName: u.userName || ('用户' + u.userId),
              joinAt: now()
            }))
            break
          case 'user-joined':
            if (msg.userId != null && !sameUserId(msg.userId, me.id)) {
              if (!viewers.value.find(v => sameUserId(v.userId, msg.userId))) {
                viewers.value.push({
                  userId: msg.userId,
                  userName: msg.userName || ('用户' + msg.userId),
                  joinAt: now()
                })
              }
            }
            break
          case 'user-left':
            viewers.value = viewers.value.filter(v => !sameUserId(v.userId, msg.userId))
            break
          case 'chat':
            if (msg.data?.text) {
              recentChats.value.push({
                id: now() + random(),
                from: msg.userName || ('用户' + msg.userId),
                text: msg.data.text
              })
              if (recentChats.value.length > 80) recentChats.value.splice(0, 30)
              if (msg.userId) {
                const uid = String(msg.userId)
                viewerDanmakuCount[uid] = (viewerDanmakuCount[uid] || 0) + 1
              }
            }
            break
        }
      }
    })
    ensureMetricsTimer()
    return true
  }

  const cleanupAudienceMonitor = () => {
    if (ws) {
      try { ws.close() } catch (e) {}
      ws = null
    }
    if (metricsTimer) {
      clearTimer(metricsTimer)
      metricsTimer = null
    }
    viewers.value = []
    recentChats.value = []
    danmakuRate.value = 0
    peakViewers.value = 0
    totalDanmaku.value = 0
    Object.keys(viewerDanmakuCount).forEach(userId => delete viewerDanmakuCount[userId])
  }

  return {
    viewers,
    recentChats,
    danmakuRate,
    peakViewers,
    totalDanmaku,
    viewerDanmakuCount,
    leaderboard,
    viewerLevels: VIEWER_LEVELS,
    getViewerLevel,
    connectAudienceWs,
    cleanupAudienceMonitor,
    updateMetrics
  }
}
