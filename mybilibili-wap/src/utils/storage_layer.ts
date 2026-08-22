// 本地存储抽象层
// 底层优先级：
//   1. window.NativeStorage（Android 壳 SQLite KV，app_webview 外，不受缓存清理影响）
//   2. window.localStorage（浏览器 / 未注入桥时兜底）
// 键统一加 'wap:' 前缀，避免与旧 localStorage 键冲突。

const PREFIX = 'wap:'

interface Bridge {
  get(key: string): string | null
  set(key: string, value: string): void
  remove(key: string): void
  multiGet(prefix: string): string
  bulkSet(json: string): void
}

function hasBridge(): boolean {
  return typeof window !== 'undefined' && !!(window as any).NativeStorage
}

function bridge(): Bridge | null {
  return hasBridge() ? (window as any).NativeStorage : null
}

type StorageValue = string | number | boolean | object | null | undefined

function serialize(v: StorageValue): string | null {
  if (v === undefined || v === null) return null
  if (typeof v === 'string') return v
  return JSON.stringify(v)
}

function deserialize<T = any>(raw: string | null): T | null {
  if (raw === null || raw === undefined) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return raw as unknown as T
  }
}

function lsKey(key: string): string {
  return PREFIX + key
}

// 统一键名（供上层使用）
export const K = {
  token: 'token',
  refreshToken: 'refresh_token',
  user: 'user',
  theme: 'theme',
  // 缓存类
  cacheRecommended: 'cache:recommended',
  cacheHot: 'cache:hot',
  cacheCategory: 'cache:category:',
  cacheDetail: 'cache:detail:',
  cacheUser: 'cache:user:',
  cacheDanmaku: 'cache:danmaku:',
  // 搜索历史
  searchHistory: 'search:history'
}

export default {
  hasBridge: hasBridge(),

  /** 读取任意数据，返回解析后的值 */
  get: function <T = any>(key: string): T | null {
    const b = this.hasBridge ? bridge() : null
    if (b) {
      return deserialize<T>(b.get(lsKey(key)))
    }
    const raw = window.localStorage.getItem(lsKey(key))
    return deserialize<T>(raw)
  },

  /** 写入任意数据 */
  set: function (key: string, value: StorageValue): void {
    const raw = serialize(value)
    const b = this.hasBridge ? bridge() : null
    if (b) {
      if (raw === null) b.remove(lsKey(key))
      else b.set(lsKey(key), raw)
      return
    }
    if (raw === null) window.localStorage.removeItem(lsKey(key))
    else window.localStorage.setItem(lsKey(key), raw)
  },

  /** 删除 */
  remove: function (key: string): void {
    const b = this.hasBridge ? bridge() : null
    if (b) {
      b.remove(lsKey(key))
      return
    }
    window.localStorage.removeItem(lsKey(key))
  },

  /** 批量读取某前缀下的所有 key-value（键剥去前缀） */
  multiGet: function <T = any>(prefix: string): Array<{ key: string; value: T }> {
    const b = this.hasBridge ? bridge() : null
    if (b) {
      try {
        const arr = JSON.parse(b.multiGet(lsKey(prefix)))
        return arr.map((o: any) => ({
          key: o.key.slice(PREFIX.length),
          value: deserialize<T>(o.value)
        }))
      } catch {
        return []
      }
    }
    const out: Array<{ key: string; value: T }> = []
    const target = lsKey(prefix)
    for (let i = 0; i < window.localStorage.length; i++) {
      const rawKey = window.localStorage.key(i)
      if (rawKey && rawKey.startsWith(target)) {
        const raw = window.localStorage.getItem(rawKey)
        out.push({
          key: rawKey.slice(PREFIX.length),
          value: deserialize<T>(raw)
        })
      }
    }
    return out
  },

  /** 批量写入 [{key, value}] */
  bulkSet: function (items: Array<{ key: string; value: StorageValue }>): void {
    const b = this.hasBridge ? bridge() : null
    if (b) {
      const payload = items
        .map(it => ({ key: lsKey(it.key), value: serialize(it.value) }))
        .filter(it => it.value !== null)
      if (payload.length) b.bulkSet(JSON.stringify(payload))
      return
    }
    for (const it of items) this.set(it.key, it.value)
  },

  /** 清空全部本应用数据 */
  clearAll: function (): void {
    const b = this.hasBridge ? bridge() : null
    if (b) {
      b.clearAll()
      return
    }
    for (let i = window.localStorage.length - 1; i >= 0; i--) {
      const rawKey = window.localStorage.key(i)
      if (rawKey && rawKey.startsWith(PREFIX)) window.localStorage.removeItem(rawKey)
    }
  }
}