// SWR 风格缓存工具：本地命中先行渲染 + 后台拉到新数据再覆盖
// 配合 storage 抽象层使用。
import storage from './storage_layer'

interface CacheEntry<T> {
  data: T
  fetchedAt: number
}

export interface SwrResult<T> {
  local: T | null
  fresh: T | null
}

const DEFAULT_TTL = {
  list: 5 * 60 * 1000,      // 列表 5 分钟
  detail: 30 * 60 * 1000,   // 详情 30 分钟
  user: 30 * 60 * 1000,     // 用户资料 30 分钟
  danmaku: 24 * 60 * 60 * 1000 // 弹幕 24 小时
} as const

export type CacheKind = keyof typeof DEFAULT_TTL

function entryKey(kind: CacheKind, id = ''): string {
  return `cache:${kind}:${id}`
}

/** 读取本地缓存，若过期返回 null */
export function readCache<T>(kind: CacheKind, id = ''): T | null {
  const raw = storage.get<CacheEntry<T>>(entryKey(kind, id))
  if (!raw) return null
  const ttl = DEFAULT_TTL[kind]
  if (Date.now() - (raw.fetchedAt || 0) > ttl) return null
  return raw.data
}

/** 写入缓存（覆盖 fetchedAt） */
export function writeCache<T>(kind: CacheKind, id: string, data: T): void {
  const entry: CacheEntry<T> = { data, fetchedAt: Date.now() }
  storage.set(entryKey(kind, id), entry)
}

/**
 * 通用 SWR 请求：先返回本地缓存，同时发起网络请求，
 * 成功后写缓存并返回 fresh。网络失败时返回本地缓存。
 */
export async function swr<T>(
  kind: CacheKind,
  id: string,
  fetcher: () => Promise<T>,
  opts: { ttlOverride?: number } = {}
): Promise<SwrResult<T>> {
  const local = readCache<T>(kind, id)
  try {
    const freshData = await fetcher()
    writeCache(kind, id, freshData)
    return { local, fresh: freshData }
  } catch (e) {
    return { local, fresh: null }
  }
}

/** 即时失效某缓存键（如下发失败、用户更新资料后） */
export function invalidateCache(kind: CacheKind, id = ''): void {
  storage.remove(entryKey(kind, id))
}