import { describe, it, expect } from 'vitest'
import {
  adaptVideo,
  adaptUser,
  adaptStats,
  adaptCard,
  adaptDetail,
  adaptResponse,
  isListUrl,
  isDetailUrl,
  snakeToCamel,
} from '@/teriteri-src/network/request'

describe('snakeToCamel', () => {
  it('converts snake_case keys to camelCase', () => {
    expect(snakeToCamel({ view_count: 10 })).toEqual({ viewCount: 10 })
  })

  it('handles nested objects', () => {
    const input = { user_info: { first_name: 'A' } }
    expect(snakeToCamel(input)).toEqual({ userInfo: { firstName: 'A' } })
  })

  it('handles arrays', () => {
    expect(snakeToCamel([{ some_val: 1 }])).toEqual([{ someVal: 1 }])
  })

  it('returns primitives as-is', () => {
    expect(snakeToCamel(42)).toBe(42)
    expect(snakeToCamel(null)).toBeNull()
  })
})

describe('adaptVideo', () => {
  it('returns empty object for falsy input', () => {
    expect(adaptVideo(null)).toEqual({})
    expect(adaptVideo(undefined)).toEqual({})
  })

  it('maps Go backend fields to teriteri format', () => {
    const input = {
      id: 123,
      title: 'Test Video',
      description: 'A test',
      coverUrl: 'http://img.test/cover.jpg',
      firstVideoPlayUrl: 'http://video.test/play.mp4',
      durationSeconds: 120,
      createdAt: '2025-01-01',
      likeCount: 100,
      viewCount: 500,
      commentCount: 20,
      coinCount: 10,
      collectCount: 5,
      shareCount: 3,
      danmakuCount: 50,
      reviewStatus: 1,
    }
    const result = adaptVideo(input)
    expect(result.vid).toBe('123')
    expect(result.title).toBe('Test Video')
    expect(result.videoUrl).toBe('http://video.test/play.mp4')
    expect(result.duration).toBe(120)
    expect(result.good).toBe(100)
    expect(result.play).toBe(500)
    expect(result.status).toBe(1)
  })

  it('parses duration string "m:ss"', () => {
    const result = adaptVideo({ id: 1, durationSeconds: '3:45' })
    expect(result.duration).toBe(225)
  })
})

describe('adaptUser', () => {
  it('returns {uid:0} for falsy input', () => {
    expect(adaptUser(null)).toEqual({ uid: 0 })
  })

  it('maps Go backend fields', () => {
    const input = {
      id: 42,
      name: 'TestUser',
      avatar: 'http://avatar.test/a.jpg',
      experience: 500,
      vip: 1,
      gender: 2,
      followerCount: 100,
      followingCount: 50,
      totalLikeCount: 1000,
      totalViewCount: 5000,
      bio: 'Hello',
      status: 1,
      bgUrl: 'http://bg.test/bg.jpg',
      auth: 1,
      tags: ['tag1'],
    }
    const result = adaptUser(input)
    expect(result.uid).toBe(42)
    expect(result.nickname).toBe('TestUser')
    expect(result.avatar_url).toBe('http://avatar.test/a.jpg')
    expect(result.fansCount).toBe(100)
    expect(result.followsCount).toBe(50)
    expect(result.loveCount).toBe(1000)
  })

  it('falls back to alternate field names', () => {
    const input = { id: 1, nickname: 'Alt', face: 'img.jpg', fansCount: 99 }
    const result = adaptUser(input)
    expect(result.nickname).toBe('Alt')
    expect(result.fansCount).toBe(99)
  })
})

describe('adaptStats', () => {
  it('maps all stat fields', () => {
    const input = {
      viewCount: 100,
      danmakuCount: 200,
      likeCount: 300,
      coinCount: 400,
      collectCount: 500,
      shareCount: 600,
      commentCount: 700,
    }
    const result = adaptStats(input)
    expect(result).toEqual({
      play: 100,
      danmu: 200,
      good: 300,
      coin: 400,
      collect: 500,
      share: 600,
      comment: 700,
    })
  })

  it('defaults to 0 for missing fields', () => {
    expect(adaptStats({})).toEqual({
      play: 0, danmu: 0, good: 0, coin: 0, collect: 0, share: 0, comment: 0,
    })
  })
})

describe('adaptCard', () => {
  it('combines video, stats, user from manuscript', () => {
    const input = {
      id: 1,
      title: 'T',
      coverUrl: 'c.jpg',
      likeCount: 10,
      viewCount: 20,
      uploader: { id: 5, name: 'U' },
    }
    const result = adaptCard(input)
    expect(result.video).toBeDefined()
    expect(result.stats).toBeDefined()
    expect(result.user.uid).toBe(5)
  })
})

describe('adaptDetail', () => {
  it('includes videos array and category', () => {
    const input = {
      id: 1,
      title: 'T',
      videos: [{ id: 10, title: 'P1', playUrlHd: 'url', durationSeconds: 60, videoOrder: 0 }],
      uploader: { id: 5 },
      categoryId: 3,
      categoryName: 'Tech',
    }
    const result = adaptDetail(input)
    expect(result.videos).toHaveLength(1)
    expect(result.videos[0].vid).toBe('10')
    expect(result.category.mcId).toBe(3)
    expect(result.manuscriptId).toBe('1')
  })
})

describe('isListUrl', () => {
  it('recognizes list endpoints', () => {
    expect(isListUrl('/manuscript/recommended')).toBe(true)
    expect(isListUrl('/manuscript/hot')).toBe(true)
    expect(isListUrl('/manuscript/list')).toBe(true)
    expect(isListUrl('/manuscript/userManuscripts')).toBe(true)
    expect(isListUrl('/manuscript/userCollections')).toBe(true)
    expect(isListUrl('/manuscript/myLikes')).toBe(true)
    expect(isListUrl('/manuscript/category')).toBe(true)
    expect(isListUrl('/manuscript/userSearch')).toBe(true)
    expect(isListUrl('/video/random/visitor')).toBe(true)
    expect(isListUrl('/video/user-works')).toBe(true)
    expect(isListUrl('/video/user-love')).toBe(true)
    expect(isListUrl('/video/user-collect')).toBe(true)
    expect(isListUrl('/video/cumulative/visitor')).toBe(true)
    expect(isListUrl('/favorites/1/videos')).toBe(true)
  })

  it('rejects non-list endpoints', () => {
    expect(isListUrl('/manuscript/123')).toBe(false)
    expect(isListUrl('/user/info')).toBe(false)
    expect(isListUrl('/comment/get')).toBe(false)
  })
})

describe('isDetailUrl', () => {
  it('matches manuscript detail', () => {
    expect(isDetailUrl('/manuscript/42')).toBe(true)
    expect(isDetailUrl('/manuscript/12345')).toBe(true)
  })

  it('matches video getone', () => {
    expect(isDetailUrl('/video/getone')).toBe(true)
  })

  it('rejects other urls', () => {
    expect(isDetailUrl('/manuscript/recommended')).toBe(false)
    expect(isDetailUrl('/video/list')).toBe(false)
  })
})

describe('adaptResponse', () => {
  describe('favorites list', () => {
    it('adapts favorite list to teriteri format', () => {
      const data = [
        { id: 1, name: '默认收藏夹', video_count: 5, user_id: 42, visible: 1 },
        { id: 2, name: '自建', video_count: 3, user_id: 42, visible: 0 },
      ]
      const result = adaptResponse('/favorite/get-all/42', data)
      expect(result.data).toHaveLength(2)
      expect(result.data[0].fid).toBe(1)
      expect(result.data[0].title).toBe('默认收藏夹')
      expect(result.data[0].type).toBe(1) // default
      expect(result.data[1].type).toBe(0) // custom
    })
  })

  describe('user-works', () => {
    it('adapts manuscripts to card array', () => {
      const data = [{ id: 1, title: 'T', viewCount: 100, uploader: { id: 5 } }]
      const result = adaptResponse('/video/user-works', data)
      expect(result.data.list).toHaveLength(1)
      expect(result.data.list[0].video.vid).toBe('1')
    })
  })

  describe('comment list', () => {
    it('adapts comments with nested replies', () => {
      const data = [{
        id: 1,
        content: 'Hello',
        userId: 10,
        userName: 'User',
        likeCount: 5,
        replyCount: 1,
        replies: [{
          id: 2,
          commentId: 1,
          content: 'Reply',
          userId: 20,
          userName: 'RUser',
          likeCount: 1,
          replyToUserId: 10,
          replyToUserName: 'User',
        }],
      }]
      const result = adaptResponse('/comment/get', data)
      expect(result.data.comments).toHaveLength(1)
      expect(result.data.comments[0].replies).toHaveLength(1)
      expect(result.data.comments[0].replies[0].toUser.uid).toBe(10)
    })
  })

  describe('user info', () => {
    it('adapts user info', () => {
      const data = { id: 42, name: 'Test' }
      const result = adaptResponse('/user/info/get-one', data)
      expect(result.data.uid).toBe(42)
      expect(result.data.nickname).toBe('Test')
    })
  })

  describe('category list', () => {
    it('adapts categories to channel format', () => {
      const data = [{ id: 1, name: 'Tech' }, { id: 2, name: 'Game' }]
      const result = adaptResponse('/category/getall', data)
      expect(result.data).toHaveLength(2)
      expect(result.data[0].mcId).toBe(1)
      expect(result.data[0].mcName).toBe('Tech')
    })
  })

  describe('search results', () => {
    it('adapts video search results', () => {
      const data = { list: [{ id: 1, title: 'V', viewCount: 10 }] }
      const result = adaptResponse('/search/video', data)
      expect(result.data).toHaveLength(1)
    })

    it('adapts user search results', () => {
      const data = [{ id: 1, name: 'U', followerCount: 5 }]
      const result = adaptResponse('/search/user', data)
      expect(result.data[0].uid).toBe(1)
    })
  })

  describe('hot search', () => {
    it('maps backend {keyword, rank, score} to teriteri {content, type}', () => {
      const data = [
        { keyword: 'keyword1', rank: 1, score: 100 },
        { keyword: 'keyword2', rank: 2, score: 50 }
      ]
      const result = adaptResponse('/search/hot/get', data)
      expect(result.data).toEqual([
        { content: 'keyword1', rank: 1, score: 100, type: 0 },
        { content: 'keyword2', rank: 2, score: 50, type: 0 }
      ])
    })

    it('filters out empty keywords', () => {
      const data = [
        { keyword: '', rank: 1, score: 111 },
        { keyword: 'OpenClaw', rank: 2, score: 79 }
      ]
      const result = adaptResponse('/search/hot/get', data)
      expect(result.data).toEqual([{ content: 'OpenClaw', rank: 2, score: 79, type: 0 }])
    })
  })

  describe('user-works-count', () => {
    it('returns total number', () => {
      const result = adaptResponse('/video/user-works-count', { total: 42 })
      expect(result.data).toBe(42)
    })
  })

  describe('detail URL', () => {
    it('adapts manuscript detail', () => {
      const data = { id: 1, title: 'T', coverUrl: 'c.jpg', videos: [], uploader: { id: 5 } }
      const result = adaptResponse('/manuscript/1', data)
      expect(result.data.video).toBeDefined()
      expect(result.data.manuscriptId).toBe('1')
    })
  })

  describe('generic fallback', () => {
    it('passes through unrecognized list URLs', () => {
      const data = [{ id: 1 }]
      const result = adaptResponse('/unknown/list', data)
      expect(result.data).toBeDefined()
    })

    it('passes through unrecognized non-list data', () => {
      const result = adaptResponse('/something/else', { foo: 'bar' })
      expect(result.data.foo).toBe('bar')
    })
  })
})
