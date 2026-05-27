// 视频详情 API - 复用 mybilibili-web 的接口
// /manuscript/{id} → 稿件详情
// /recommend/related/{id} → 相关推荐
// /comment/list → 评论列表
// /danmaku/video/{id} → 弹幕
import api from './client'

// 推荐视频适配
const adaptRecommend = (v: any) => ({
  aId: v.manuscriptId || v.id,
  title: v.title,
  pic: v.coverUrl,
  author: v.username || v.nickname || v.author || '',
  mid: v.userId,
  play: v.viewCount || 0,
  videoReview: v.danmakuCount || 0,
  duration: v.duration
})

// 稿件详情 - 完整字段映射
export async function getVideoInfo(aId: number) {
  try {
    const res = await api.get(`/manuscript/${aId}`)
    const data = res?.data || res
    if (!data) return { code: '0', data: null }

    // 当前分P (默认为第一P)
    const videos = data.videos || []
    const currentVideo = videos[0] || {}

    return {
      code: '1',
      data: {
        aId: data.id || data.manuscriptId,
        title: data.title,
        pic: data.coverUrl,
        author: data.uploader?.name || data.username || data.nickname || '',
        mid: data.uploader?.id || data.userId,
        play: data.viewCount || 0,
        videoReview: data.danmakuCount || 0,
        ctime: data.createTime,
        description: data.description || '',
        // 播放地址
        playUrl: currentVideo.playUrl || '',
        playUrlHd: currentVideo.playUrlHd || '',
        playUrlSd: currentVideo.playUrlSd || '',
        playUrlLd: currentVideo.playUrlLd || '',
        // 互动计数
        likeCount: data.likeCount || 0,
        coinCount: data.coinCount || 0,
        collectCount: data.collectCount || 0,
        shareCount: data.shareCount || 0,
        // UP主详情
        uploader: data.uploader ? {
          id: data.uploader.id,
          name: data.uploader.name,
          face: data.uploader.avatar || '',
          level: data.uploader.level || 1,
          sign: data.uploader.signature || data.uploader.bio || '',
          followerCount: data.uploader.followerCount || 0,
          followingCount: data.uploader.followingCount || 0
        } : null,
        // 标签
        tags: data.tags || [],
        // 多分P
        videos: videos.map((v: any, i: number) => ({
          id: v.id,
          title: v.title || `P${i + 1}`,
          playUrl: v.playUrl || '',
          playUrlHd: v.playUrlHd || '',
          playUrlSd: v.playUrlSd || '',
          playUrlLd: v.playUrlLd || '',
          duration: v.duration
        }))
      }
    }
  } catch (e) {
    console.error(e)
    return { code: '0', data: null }
  }
}

// 相关推荐
export async function getRecommendVides(aId: number) {
  try {
    const res = await api.get(`/recommend/related/${aId}?size=8`)
    return {
      code: '1',
      data: (res?.data || res || []).map(adaptRecommend)
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 评论列表
export async function getComments(aId: number, p = 1, size = 20) {
  try {
    const res = await api.get(`/comment/list?manuscriptId=${aId}&page=${p}&size=${size}&sort=time`)
    const list = res?.data || res || []
    return {
      code: '1',
      data: list.map((c: any) => ({
        rpid: c.id,
        content: c.content,
        ctime: c.createTime,
        likeCount: c.likeCount || 0,
        isTop: c.isTop || false,
        user: {
          name: c.userName || c.username || c.nickname || c.author || '',
          face: c.userAvatar || c.avatarUrl || c.avatar || '',
          level: c.userLevel || 1
        },
        replies: (c.replies || []).map((r: any) => ({
          rpid: r.id,
          name: r.userName || r.username || r.nickname || r.author || '',
          content: r.content,
          isUp: r.isUp || false,
          likeCount: r.likeCount || 0
        }))
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 弹幕
export async function getBarrages(videoId: number) {
  try {
    const res = await api.get(`/danmaku/video/${videoId}`)
    return {
      code: '1',
      data: (res?.data || res || []).map((b: any) => ({
        text: b.content,
        color: b.color || '#fff',
        time: b.time || 0
      }))
    }
  } catch (e) {
    return { code: '0', data: [] }
  }
}

// 发送弹幕
export async function sendBarrage(videoId: number, text: string, time: number, color = '#ffffff') {
  try {
    const res = await api.post(`/danmaku/send`, {
      videoId,
      content: text,
      time: Math.floor(time),
      color,
      type: 1
    })
    return { code: '1', data: res?.data || res }
  } catch (e) {
    return { code: '0', data: null }
  }
}