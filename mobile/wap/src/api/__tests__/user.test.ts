import { describe, it, expect, vi, beforeEach } from 'vitest'
import { normalizeUser } from '../user'

describe('normalizeUser', () => {
  it('标准字段直接映射', () => {
    const raw = {
      id: 1,
      username: 'admin',
      nickname: '管理员',
      avatar: 'avatar.jpg',
      bio: '签名',
      followerCount: 100,
      followingCount: 50,
      dynamicCount: 10,
      manuscriptCount: 5,
      coinCount: 200,
      level: 5
    }
    const result = normalizeUser(raw)
    expect(result.id).toBe(1)
    expect(result.username).toBe('admin')
    expect(result.nickname).toBe('管理员')
    expect(result.avatar).toBe('avatar.jpg')
    expect(result.bio).toBe('签名')
    expect(result.followerCount).toBe(100)
    expect(result.followingCount).toBe(50)
    expect(result.dynamicCount).toBe(10)
    expect(result.manuscriptCount).toBe(5)
    expect(result.coinCount).toBe(200)
    expect(result.level).toBe(5)
  })

  it('后端字段名映射（mid/user_id/face/sign 等）', () => {
    const raw = {
      user: {
        mid: 42,
        name: 'testuser',
        face: 'face.jpg',
        sign: '后端签名',
        fans: 999,
        attentions: 10,
        dynamics: 3,
        videos: 7,
        coins: 50
      }
    }
    const result = normalizeUser(raw)
    expect(result.id).toBe(42)
    expect(result.username).toBe('testuser')
    expect(result.nickname).toBe('testuser')
    expect(result.avatar).toBe('face.jpg')
    expect(result.signature).toBe('后端签名')
    expect(result.followerCount).toBe(999)
    expect(result.followingCount).toBe(10)
    expect(result.dynamicCount).toBe(3)
    expect(result.manuscriptCount).toBe(7)
    expect(result.coinCount).toBe(50)
  })

  it('data.data.user 嵌套结构', () => {
    const raw = { data: { user: { id: 7, username: 'nested' } } }
    const result = normalizeUser(raw)
    expect(result.id).toBe(7)
    expect(result.username).toBe('nested')
  })

  it('空/null 输入 → 默认值', () => {
    expect(normalizeUser(null).id).toBeUndefined()
    expect(normalizeUser(undefined).id).toBeUndefined()
    expect(normalizeUser({}).id).toBeUndefined()
  })

  it('followerCount 兼容 follower/followers/fans', () => {
    expect(normalizeUser({ follower: 10 }).followerCount).toBe(10)
    expect(normalizeUser({ followers: 20 }).followerCount).toBe(20)
    expect(normalizeUser({ fans: 30 }).followerCount).toBe(30)
    expect(normalizeUser({ followerCount: 40 }).followerCount).toBe(40)
  })

  it('level 默认值为1', () => {
    expect(normalizeUser({}).level).toBe(1)
  })
})
