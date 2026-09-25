// 会话管理
//
// 阶段 3：access_token / refresh_token 全部是 HttpOnly cookie，JS 读不到也写不了。
// 本地只留两份东西：
//   - user：登录时服务端返回的展示信息，同时是前端判断登录态的唯一信号
//   - K.token / K.refreshToken：升级前的明文遗留，只在清理时被碰
import storage, { K } from './storage_layer'
import api from '../api/client'

// 登录响应的 data 里会同时带 token / refresh_token（兼容旧客户端），
// 而 Login.vue 又常常把整个 data 当 user 传进来。不洗一遍就等于把 HttpOnly
// 想保护的东西又写回 localStorage，等于没做。
const CREDENTIAL_KEYS = new Set([
  'token', 'refresh_token', 'refreshToken', 'access_token', 'accessToken',
  'password', 'pwd'
])

function scrubCredentials<T>(value: T): T {
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

export interface SessionUser {
  id?: number
  userId?: number
  name?: string
  nickname?: string
  username?: string
  avatar?: string
  [k: string]: any
}

export function getLocalUser(): SessionUser | null {
  return scrubCredentials(storage.get<SessionUser>(K.user))
}

export function getLocalUserId(): number {
  const u = getLocalUser()
  if (!u) return 0
  return u.id || u.userId || u.user?.id || 0
}

/** 登录态由本地展示信息回答；凭证本身在 HttpOnly cookie 里，读不到也不该读。 */
export function isLogin(): boolean {
  return !!getLocalUser()
}

/**
 * 登录成功后保存会话。
 * token / refreshToken 只是签名兼容——它们已由服务端写进 HttpOnly cookie，
 * 这里刻意丢弃，避免把能当身份用的东西落进 localStorage（XSS 可直接整包读走）。
 */
export function saveSession(token: string, refreshToken: string, user: SessionUser): void {
  storage.set(K.user, scrubCredentials(user))
}

/** 清本地展示信息 + 升级前的明文凭证明文。不含服务端作废，见 logout()。 */
export function clearSession(): void {
  storage.remove(K.user)
  storage.remove(K.token)
  storage.remove(K.refreshToken)
}

/**
 * 让服务端作废 HttpOnly 登录 cookie。
 * 凭证读不到，所以"看起来已登出但接口仍带凭证"的假登出必须靠这个接口避免。
 */
export async function logout(): Promise<void> {
  try {
    await api.post('/user/logout', {})
  } catch {
    // 登出幂等，失败不阻断本地清理
  }
  clearSession()
}

let refreshing: Promise<boolean> | null = null

/**
 * 尝试静默续期：refresh_token 走 HttpOnly cookie，body 不带值。
 * 成功返回 true（新令牌已落 cookie）；失败返回 false（调用方决定是否登出）。
 */
export function tryRefresh(): Promise<boolean> {
  if (!isLogin()) return Promise.resolve(false)
  if (refreshing) return refreshing

  refreshing = api
    .post('/user/token/refresh', {})
    .then((res: any) => res?.code === 200)
    .catch(() => false)
    .finally(() => {
      refreshing = null
    })
  return refreshing
}

/**
 * 启动时校验：本地有登录态 → 调 /user/me 校验 cookie 是否仍有效。
 * 401 → 尝试 refresh 续期；续期失败才清本地并返回 false。
 */
export async function bootstrapSession(): Promise<boolean> {
  if (!isLogin()) return false
  try {
    await api.get('/user/me')
    return true
  } catch (e: any) {
    const status = e?.response?.status
    if (status === 401 || status === 403) {
      const ok = await tryRefresh()
      if (ok) return true
      clearSession()
      return false
    }
    // 网络异常：保留本地展示信息，允许离线使用缓存
    return !!getLocalUser()
  }
}
