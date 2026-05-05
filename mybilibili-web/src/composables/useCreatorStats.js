import { ref, reactive } from 'vue'
import { statsApi } from '@/api/creator'

export function useCreatorStats() {
  const loading = reactive({
    overview: false,
    trend: false,
    ranking: false,
    fansTrend: false
  })

  const overview = ref({
    totalViews: 0, totalLikes: 0, totalComments: 0, totalDanmaku: 0,
    totalCoins: 0, totalCollections: 0, totalShares: 0, totalManuscripts: 0,
    totalFollowers: 0, followersIncrease: 0,
    viewsIncrease: 0, likesIncrease: 0, commentsIncrease: 0, danmakuIncrease: 0,
    sharesIncrease: 0, collectionsIncrease: 0, coinsIncrease: 0,
    updateTime: ''
  })

  const trendData = ref({ dates: [], views: [], likes: [], comments: [], followers: [] })
  const rankingList = ref([])
  const fansTrend = ref({
    dates: [], newFollowers: [], unfollows: [], totalFollowers: [],
    currentFollowers: 0, newFollowersToday: 0, unfollowsToday: 0, growthRate: 0
  })

  const cache = new Map()
  const CACHE_TTL = 5 * 60 * 1000

  function isCacheValid(key) {
    const entry = cache.get(key)
    if (!entry) return false
    return Date.now() - entry.time < CACHE_TTL
  }

  async function fetchWithCache(key, fetcher) {
    if (isCacheValid(key)) {
      return cache.get(key).data
    }
    const data = await fetcher()
    cache.set(key, { data, time: Date.now() })
    return data
  }

  async function loadOverview(force = true) {
    if (loading.overview) return
    loading.overview = true
    try {
      const res = await statsApi.getOverview()
      if (res.code === 200 && res.data) {
        overview.value = { ...overview.value, ...res.data }
      }
    } finally {
      loading.overview = false
    }
  }

  async function loadTrend(days = 7, force = true) {
    if (loading.trend) return
    loading.trend = true
    try {
      const res = await statsApi.getTrend(days)
      if (res.code === 200 && res.data) {
        trendData.value = res.data
      }
    } finally {
      loading.trend = false
    }
  }

  async function loadRanking(sortBy = 'views', limit = 10, force = true) {
    if (loading.ranking) return
    loading.ranking = true
    try {
      const res = await statsApi.getRanking(sortBy, limit)
      if (res.code === 200 && res.data) {
        rankingList.value = res.data.list || []
      }
    } finally {
      loading.ranking = false
    }
  }

  async function loadFansTrend(days = 30, force = true) {
    if (loading.fansTrend) return
    loading.fansTrend = true
    try {
      const res = await statsApi.getFansTrend(days)
      if (res.code === 200 && res.data) {
        fansTrend.value = res.data
      }
    } finally {
      loading.fansTrend = false
    }
  }

  function clearCache() {
    cache.clear()
  }

  return {
    loading,
    overview,
    trendData,
    rankingList,
    fansTrend,
    loadOverview,
    loadTrend,
    loadRanking,
    loadFansTrend,
    clearCache
  }
}
