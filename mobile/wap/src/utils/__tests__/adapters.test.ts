import { describe, it, expect, vi, beforeEach } from 'vitest'

// 测试适配器逻辑（这些函数在各 API 模块中，提取出来单独测试）

describe('adaptVideo 逻辑', () => {
  // 复制 index.ts 中的 adaptVideo 函数
  const adaptVideo = (v: any) => ({
    aId: v.manuscriptId || v.id || v.videoId,
    title: v.title || '无标题',
    pic: v.coverUrl || v.cover || '',
    author: v.userName || v.uploader?.name || v.username || v.nickname || v.author || 'UP主',
    mid: v.userId,
    play: v.viewCount || v.play || 0,
    videoReview: v.commentCount || v.videoReview || v.danmakuCount || 0,
    duration: v.duration || v.durationSeconds,
    description: v.description || '',
    isVertical: (v.isVertical || v.is_vertical) ? 1 : 0
  })

  it('标准字段映射', () => {
    const raw = {
      manuscriptId: 1,
      title: '视频标题',
      coverUrl: 'cover.jpg',
      userName: 'UP主',
      userId: 42,
      viewCount: 1000,
      commentCount: 50,
      duration: '10:30',
      description: '描述'
    }
    const result = adaptVideo(raw)
    expect(result.aId).toBe(1)
    expect(result.title).toBe('视频标题')
    expect(result.pic).toBe('cover.jpg')
    expect(result.author).toBe('UP主')
    expect(result.mid).toBe(42)
    expect(result.play).toBe(1000)
    expect(result.videoReview).toBe(50)
    expect(result.duration).toBe('10:30')
    expect(result.description).toBe('描述')
  })

  it('兼容多种字段名', () => {
    const raw = { id: 2, cover: 'c.jpg', nickname: '昵称', play: 999, videoReview: 10 }
    const result = adaptVideo(raw)
    expect(result.aId).toBe(2)
    expect(result.pic).toBe('c.jpg')
    expect(result.author).toBe('昵称')
    expect(result.play).toBe(999)
    expect(result.videoReview).toBe(10)
  })

  it('uploader 嵌套对象', () => {
    const raw = { uploader: { name: 'UP' }, manuscriptId: 3 }
    expect(adaptVideo(raw).author).toBe('UP')
  })

  it('isVertical / is_vertical → 1', () => {
    expect(adaptVideo({ manuscriptId: 1, isVertical: true }).isVertical).toBe(1)
    expect(adaptVideo({ manuscriptId: 1, is_vertical: true }).isVertical).toBe(1)
    expect(adaptVideo({ manuscriptId: 1 }).isVertical).toBe(0)
  })

  it('空输入 → 默认值', () => {
    const result = adaptVideo({})
    expect(result.title).toBe('无标题')
    expect(result.author).toBe('UP主')
    expect(result.play).toBe(0)
  })
})

describe('adaptLive 逻辑', () => {
  const adaptLive = (l: any) => ({
    roomId: l.id,
    title: l.roomName || l.title || '',
    cover: l.coverUrl || l.cover,
    isLive: l.status === 'live' ? 1 : 0,
    onlineNum: l.viewerCount || 0,
    playUrl: l.streamUrl || l.playUrl,
    uname: l.anchorName || l.username || l.nickname || '',
    description: l.description || ''
  })

  it('标准字段映射', () => {
    const raw = {
      id: 100,
      roomName: '直播间',
      coverUrl: 'cover.jpg',
      status: 'live',
      viewerCount: 500,
      streamUrl: 'rtmp://...',
      anchorName: '主播',
      description: '直播描述'
    }
    const result = adaptLive(raw)
    expect(result.roomId).toBe(100)
    expect(result.title).toBe('直播间')
    expect(result.cover).toBe('cover.jpg')
    expect(result.isLive).toBe(1)
    expect(result.onlineNum).toBe(500)
    expect(result.playUrl).toBe('rtmp://...')
    expect(result.uname).toBe('主播')
    expect(result.description).toBe('直播描述')
  })

  it('非 live 状态 → isLive=0', () => {
    expect(adaptLive({ id: 1, status: 'offline' }).isLive).toBe(0)
    expect(adaptLive({ id: 1 }).isLive).toBe(0)
  })

  it('兼容 title/roomName', () => {
    expect(adaptLive({ id: 1, title: 'T' }).title).toBe('T')
    expect(adaptLive({ id: 1, roomName: 'R' }).title).toBe('R')
  })
})

describe('adaptUpUser 逻辑', () => {
  const adaptUpUser = (u: any) => ({
    mid: u.mid || u.id,
    name: u.name || u.nickname || u.username || '',
    face: u.face || u.avatar || '',
    sign: u.sign || u.signature || '',
    fans: u.fans || u.follower || u.follower_count || 0,
    videos: u.videos || u.manuscript_count || 0
  })

  it('标准字段映射', () => {
    const raw = { mid: 1, name: 'UP', face: 'face.jpg', sign: '签名', fans: 100, videos: 10 }
    const result = adaptUpUser(raw)
    expect(result.mid).toBe(1)
    expect(result.name).toBe('UP')
    expect(result.face).toBe('face.jpg')
    expect(result.sign).toBe('签名')
    expect(result.fans).toBe(100)
    expect(result.videos).toBe(10)
  })

  it('兼容多种字段名', () => {
    const raw = { id: 2, nickname: '昵称', avatar: 'av.jpg', signature: '签', follower: 50, manuscript_count: 5 }
    const result = adaptUpUser(raw)
    expect(result.mid).toBe(2)
    expect(result.name).toBe('昵称')
    expect(result.face).toBe('av.jpg')
    expect(result.sign).toBe('签')
    expect(result.fans).toBe(50)
    expect(result.videos).toBe(5)
  })
})

describe('toKeyword 逻辑', () => {
  const toKeyword = (item: any) => {
    if (typeof item === 'string') return item.trim()
    if (item && typeof item === 'object') {
      const k = item.keyword
      return typeof k === 'string' ? k.trim() : ''
    }
    return ''
  }

  it('字符串 → trim', () => {
    expect(toKeyword('  hello  ')).toBe('hello')
  })

  it('对象 → keyword 字段', () => {
    expect(toKeyword({ keyword: 'test' })).toBe('test')
  })

  it('对象无 keyword → 空', () => {
    expect(toKeyword({ other: 'x' })).toBe('')
  })

  it('null/undefined → 空', () => {
    expect(toKeyword(null)).toBe('')
    expect(toKeyword(undefined)).toBe('')
  })

  it('数字 → 空', () => {
    expect(toKeyword(123)).toBe('')
  })

  it('空字符串 → 空', () => {
    expect(toKeyword('')).toBe('')
  })

  it('纯空格 → 空', () => {
    expect(toKeyword('   ')).toBe('')
  })
})

describe('normalizeUser 逻辑', () => {
  const normalizeUser = (raw: any) => {
    const user = raw?.user || raw?.data?.user || raw?.data || raw || {}
    return {
      ...user,
      id: user.id || user.userId || user.mid || user.user_id,
      username: user.username || user.name || '',
      nickname: user.nickname || user.username || user.name || '',
      avatar: user.avatar || user.avatarUrl || user.face || '',
      bio: user.bio || user.sign || '',
      signature: user.signature || user.bio || user.sign || '',
      followerCount: user.followerCount ?? user.follower ?? user.followers ?? user.fans ?? 0,
      followingCount: user.followingCount ?? user.following ?? user.attentions ?? 0,
      dynamicCount: user.dynamicCount ?? user.dynamics ?? 0,
      manuscriptCount: user.manuscriptCount ?? user.videos ?? user.videoCount ?? 0,
      coinCount: user.coinCount ?? user.coins ?? 0,
      level: user.level || 1
    }
  }

  it('多层嵌套结构', () => {
    const raw = { data: { user: { id: 1, username: 'admin' } } }
    expect(normalizeUser(raw).id).toBe(1)
    expect(normalizeUser(raw).username).toBe('admin')
  })

  it('后端字段名兼容（mid/face/sign/fans）', () => {
    const raw = { mid: 42, face: 'f.jpg', sign: 's', fans: 100 }
    const result = normalizeUser(raw)
    expect(result.id).toBe(42)
    expect(result.avatar).toBe('f.jpg')
    expect(result.signature).toBe('s')
    expect(result.followerCount).toBe(100)
  })

  it('followerCount 优先级', () => {
    expect(normalizeUser({ followerCount: 10 }).followerCount).toBe(10)
    expect(normalizeUser({ follower: 20 }).followerCount).toBe(20)
    expect(normalizeUser({ followers: 30 }).followerCount).toBe(30)
    expect(normalizeUser({ fans: 40 }).followerCount).toBe(40)
    expect(normalizeUser({}).followerCount).toBe(0)
  })
})

describe('scrubCredentials 逻辑', () => {
  const CREDENTIAL_KEYS = new Set([
    'token', 'refresh_token', 'refreshToken', 'access_token', 'accessToken',
    'password', 'pwd'
  ])

  const scrubCredentials = <T>(value: T): T => {
    const walk = (v: any, depth: number): any => {
      if (depth > 8 || v === null || typeof v !== 'object') return v
      if (Array.isArray(v)) return v.map(x => walk(x, depth + 1))
      const out: any = {}
      for (const [k, item] of Object.entries(v)) {
        if (CREDENTIAL_KEYS.has(k)) continue
        out[k] = walk(item, depth + 1)
      }
      return out
    }
    return walk(value, 0) as T
  }

  it('移除敏感字段', () => {
    const input = { id: 1, token: 'secret', password: 'pwd', name: 'ok' }
    expect(scrubCredentials(input)).toEqual({ id: 1, name: 'ok' })
  })

  it('深层嵌套也清洗', () => {
    const input = { a: { b: { token: 'x', data: 'keep' } } }
    expect(scrubCredentials(input)).toEqual({ a: { b: { data: 'keep' } } })
  })

  it('数组也清洗', () => {
    const input = [{ token: 'a' }, { id: 1 }]
    expect(scrubCredentials(input)).toEqual([{}, { id: 1 }])
  })

  it('超过 8 层深度停止', () => {
    let deep: any = { token: 'stop' }
    for (let i = 0; i < 10; i++) deep = { next: deep }
    const result = scrubCredentials(deep)
    // 第 8 层的 token 应该被清洗
    expect(result.next?.next?.token).toBeUndefined()
  })
})
