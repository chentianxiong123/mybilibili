import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { AxiosRequestConfig, AxiosResponse } from 'axios'
import { useTeriteriStore } from '@/stores/teriteri'

function getStore() {
  try {
    return useTeriteriStore()
  } catch (e) {
    return null
  }
}

// ===== Go 后端 ManuscriptInfo(扁平) → teriteri 组件(嵌套) 转换 =====
function snakeToCamel(obj: any): any {
  if (Array.isArray(obj)) return obj.map(snakeToCamel)
  if (obj && typeof obj === 'object') {
    const out: any = {}
    for (const [k, v] of Object.entries(obj)) {
      out[k.replace(/_([a-z])/g, (_, c) => c.toUpperCase())] = snakeToCamel(v)
    }
    return out
  }
  return obj
}

function adaptVideo(v: any) {
  if (!v) return {}
  const video = (v.videos && v.videos[0]) || {}
  let duration: number | string = v.durationSeconds || v.duration || 0
  if (typeof duration === 'string' && duration.includes(':')) {
    const [m, s] = duration.split(':').map(Number)
    duration = m * 60 + (s || 0)
  } else {
    duration = Number(duration) || 0
  }
  return {
    vid: String(v.id),
    videoUrl: v.firstVideoPlayUrl || video.playUrlHd || video.playUrl || '',
    coverUrl: v.coverUrl || '',
    title: v.title || '',
    duration,
    uploadDate: v.createdAt || v.uploadTime || '',
    descr: v.description || '',
    good: v.likeCount || 0,
    play: v.viewCount || 0,
    comment: v.commentCount || 0,
    coin: v.coinCount || 0,
    collect: v.collectCount || 0,
    share: v.shareCount || 0,
    danmaku: v.danmakuCount || 0,
    type: v.sourceType || 0,
    top: v.status === 1,
    auth: v.reviewStatus || 0,
    tags: Array.isArray(v.tags) ? v.tags.join('\r\n') : (v.tags || ''),
  }
}

function adaptUser(u: any) {
  if (!u) return { uid: 0 }
  return {
    uid: String(u.id),
    nickname: u.name || u.nickname || '',
    avatar_url: u.avatar || '',
    exp: u.level || 0,
    vip: u.vip || 0,
    gender: u.gender || 0,
    fansCount: u.followerCount || 0,
    followsCount: u.followingCount || 0,
    loveCount: u.likedCount || 0,
    playCount: u.viewCount || 0,
    description: u.bio || u.signature || '',
    state: u.status || 0,
    bg_url: u.bgUrl || '',
    auth: u.auth || 0,
  }
}

function adaptStats(m: any) {
  return {
    play: m.viewCount || 0,
    danmu: m.danmakuCount || 0,
    good: m.likeCount || 0,
    coin: m.coinCount || 0,
    collect: m.collectCount || 0,
    share: m.shareCount || 0,
    comment: m.commentCount || 0,
  }
}

function adaptCard(m: any) {
  return { video: adaptVideo(m), stats: adaptStats(m), user: adaptUser(m.uploader) }
}

function adaptDetail(m: any) {
  const videos = Array.isArray(m.videos)
    ? m.videos.map((v: any) => ({
        vid: String(v.id || ''),
        title: v.title || '',
        playUrl: v.playUrlHd || v.playUrl || '',
        duration: Number(v.durationSeconds) || 0,
        videoOrder: v.videoOrder || 0,
      }))
    : []
  return {
    video: adaptVideo(m),
    user: adaptUser(m.uploader),
    category: { mcId: m.categoryId || 0, mcName: m.categoryName || '', scId: 0, scName: '' },
    stats: adaptStats(m),
    videos,
    manuscriptId: String(m.id || ''),
  }
}

function isListUrl(url: string) {
  return url.includes('/manuscript/recommended')
    || url.includes('/manuscript/hot')
    || url.includes('/manuscript/list')
    || url.includes('/manuscript/userManuscripts')
    || url.includes('/manuscript/userCollections')
    || url.includes('/manuscript/myLikes')
    || url.includes('/manuscript/category')
    || url.includes('/manuscript/userSearch')
    || url.includes('/video/random/visitor')
    || url.includes('/video/user-works')
    || url.includes('/video/user-love')
    || url.includes('/video/user-collect')
    || url.includes('/video/cumulative/visitor')
}

function isDetailUrl(url: string) {
  return /^\/manuscript\/\d+/.test(url) || /^\/video\/getone/.test(url)
}

function adaptResponse(originalUrl: string, data: any) {
  // 空间投稿列表(/video/user-works) → {list: cards, count: total}
  if (String(originalUrl || '').includes('/video/user-works')) {
    const list = Array.isArray(data) ? data : (data && data.list ? data.list : [])
    const cards = list.map((m: any) => adaptCard(snakeToCamel(m)))
    return { code: 200, data: { list: cards, count: (data && data.total) || cards.length }, message: 'ok' }
  }
  // 频道列表(/category/getall) → {mcId, mcName, scList}
  if (String(originalUrl || '').includes('/category/getall')) {
    const channels = (Array.isArray(data) ? data : []).map((c: any, idx: number) => ({
      mcId: c.id,
      mcName: c.name,
      scList: idx < 2 ? [{ mcId: c.id, scId: 0, scName: '全部' }] : [],
    }))
    return { code: 200, data: channels, message: 'ok' }
  }

  // 热搜(/search/hot/get) → 保持数组
  if (String(originalUrl || '').includes('/search/hot')) {
    return { code: 200, data: Array.isArray(data) ? data : (data || []), message: 'ok' }
  }

  // 搜索视频 → {video, stats, user}
  if (String(originalUrl || '').includes('/search/video') || String(originalUrl || '').includes('/search/videos')) {
    const list = data && data.list ? data.list : []
    const cards = list.map((item: any) => adaptCard(snakeToCamel(item)))
    return { code: 200, data: cards, message: 'ok' }
  }

  // 搜索用户 → {uid, nickname, avatar_url}
  if (String(originalUrl || '').includes('/search/user')) {
    const u = Array.isArray(data) ? data : (data && data.list ? data.list : [])
    const users = u.map((item: any) => {
      const c = snakeToCamel(item)
      return {
        uid: String(c.id || c.uid || c.mid || ''),
        nickname: c.name || c.nickname || '',
        avatar_url: c.avatar || c.face || '',
        exp: c.level || 0,
        vip: c.vip || 0,
        fansCount: c.followerCount || c.fans || 0,
        auth: c.auth || 0,
        sign: c.signature || c.bio || c.sign || '',
        level: c.level || 0,
        videos: c.videos || 0,
      }
    })
    return { code: 200, data: users, message: 'ok' }
  }

  // 评论列表(/comment/get) → {comments, more}
  if (String(originalUrl || '').includes('/comment/get')) {
    const raw = Array.isArray(data) ? data : []
    const comments = raw.map((c: any) => ({
      id: c.id || 0,
      content: c.content || '',
      createTime: c.createTime || c.createdAt || '',
      love: c.likeCount || 0,
      count: c.replyCount || 0,
      bad: c.dislikeCount || 0,
      liked: !!c.liked,
      user: {
        uid: String(c.userId ?? c.userUid ?? ''),
        nickname: c.userName || '',
        avatar_url: c.userAvatar || '',
        exp: (c.userLevel || 0) * 50,
        vip: 0,
        auth: 0,
        ...(c.user || {}),
      },
      replies: (c.replies || []).map((r: any) => ({
        id: r.id || 0,
        parentId: r.commentId || 0,
        content: r.content || '',
        createTime: r.createTime || r.createdAt || '',
        love: r.likeCount || 0,
        bad: r.dislikeCount || 0,
        liked: !!r.liked,
        user: {
          uid: String(r.userId ?? r.userUid ?? ''),
          nickname: r.userName || '',
          avatar_url: r.userAvatar || '',
          exp: (r.userLevel || 0) * 50,
          vip: 0,
          auth: 0,
          ...(r.user || {}),
        },
        toUser: r.replyToUserId || r.replyToUserName ? {
          uid: String(r.replyToUserId || ''),
          nickname: r.replyToUserName || '',
        } : null,
      })),
    }))
    return { code: 200, data: { comments, more: comments.length >= 20 }, message: 'ok' }
  }

  // 累加推荐(/manuscript/hot cumulative) → {videos, vids, more}
  if (String(originalUrl || '').includes('cumulative') && String(originalUrl || '').includes('visitor')) {
    const cards = Array.isArray(data) ? data.map(adaptCard) : []
    return {
      code: 200,
      data: {
        videos: cards,
        vids: cards.map(c => Number(c.video.vid)),
        more: cards.length > 0,
      },
      message: 'ok',
    }
  }

  const isObj = data && typeof data === 'object' && typeof data.id !== 'undefined' && typeof data.coverUrl !== 'undefined'

  if (isDetailUrl(originalUrl) && isObj) {
    return { code: 200, data: adaptDetail(data), message: 'ok' }
  }
  if (isListUrl(originalUrl) && Array.isArray(data)) {
    return { code: 200, data: data.map(adaptCard), message: 'ok' }
  }
  return { code: 200, data, message: 'ok' }
}

function handleAuthFailure(err: any) {
  const store = getStore()
  if (err.response && err.response.headers && err.response.headers.message === 'not login') {
    if (store) {
      store.initData()
      if (store.ws) {
        store.ws.close()
        store.setWebSocket(null)
      }
    }
    if (typeof window !== 'undefined') localStorage.removeItem('teri_token')
    ElMessage.error('请登录后查看')
  } else {
    ElMessage.error('特丽丽被玩坏了(¯﹃¯)')
  }
  if (store) store.isLoading = false
}

export interface RequestConfig extends AxiosRequestConfig {
  headers?: any
}

export function get<T = any>(url: string, config?: RequestConfig): Promise<AxiosResponse<T>> {
  const instance = axios.create({
    baseURL: '/api',
    timeout: 30000,
    withCredentials: true
  })

  instance.interceptors.response.use(
    (origResponse) => {
      if (origResponse.data && origResponse.data.code !== undefined && origResponse.data.code !== 200) {
        ElMessage.error(origResponse.data.message || '未知错误, 请打开控制台查看')
        return origResponse
      }
      const adapted = adaptResponse(url, origResponse.data && origResponse.data.data)
      return { ...origResponse, data: adapted }
    },
    (err) => {
      console.log(err)
      handleAuthFailure(err)
      return Promise.reject(err)
    }
  )

  instance.defaults.withCredentials = true

  if (config) {
    if (config.params) {
      if (config.headers) {
        return instance.get(url, { params: config.params, headers: config.headers })
      }
      return instance.get(url, { params: config.params })
    }
    if (config.headers) {
      return instance.get(url, { headers: config.headers })
    }
  }
  return instance.get(url)
}

export function post<T = any>(url: string, data?: any, headers?: any): Promise<AxiosResponse<T>> {
  const instance = axios.create({
    baseURL: '/api',
    timeout: 30000,
    withCredentials: true
  })

  instance.interceptors.response.use(
    (origResponse) => {
      if (origResponse.data && origResponse.data.code !== undefined && origResponse.data.code !== 200) {
        ElMessage.error(origResponse.data.message || '未知错误, 请打开控制台查看')
      }
      return origResponse
    },
    (err) => {
      console.log(err)
      handleAuthFailure(err)
      return Promise.reject(err)
    }
  )

  instance.defaults.withCredentials = true

  if (headers) return instance.post(url, data, headers)
  return instance.post(url, data)
}