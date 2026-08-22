// 会话管理：统一 token 读写，启动校验 + 静默续期
import storage, { K } from './storage_layer'
import api from '../api/client'

export interface SessionUser {
  id?: number
  userId?: number
  name?: string
  nickname?: string
  username?: string
  avatar?: string
  [k: string]: any
}

export function getToken(): string {
  return storage.get<string>(K.token) || ''
}

export function getRefreshToken(): string {
  return storage.get<string>(K.refreshToken) || ''
}

export function getLocalUser(): SessionUser | null {
  return storage.get<SessionUser>(K.user)
}

export function getLocalUserId(): number {
  const u = getLocalUser()
  if (!u) return 0
  return u.id || u.userId || u.user?.id || 0
}

export function isLogin(): boolean {
  return !!getToken()
}

/** 登录成功后保存凭证 */
export function saveSession(token: string, refreshToken: string, user: SessionUser): void {
  storage.set(K.token, token)
  if (refreshToken) storage.set(K.refreshToken, refreshToken)
  storage.set(K.user, user)
}

/** 清除本地凭证 */
export function clearSession(): void {
  storage.remove(K.token)
  storage.remove(K.refreshToken)
  storage.remove(K.user)
}

let refreshing: Promise<boolean> | null = null

/**
 * 尝试静默续期：用 refresh_token 换新 token。
 * 成功返回 true 并已更新本地凭证；失败返回 false（调用方决定是否登出）。
 */
export function tryRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return Promise.resolve(false)
  if (refreshing) return refreshing

  refreshing = api
    .post('/user/token/refresh', { refreshToken })
    .then((res: any) => {
      const d = res?.data || res
      if (!d || !d.token) return false
      const user = getLocalUser()
      saveSession(d.token, d.refresh_token || refreshToken, user || {})
      return true
    })
    .catch(() => false)
    .finally(() => {
      refreshing = null
    })
  return refreshing
}

/**
 * 启动时校验：本地有 token → 调 /user/me 校验有效性。
 * 401 → 尝试 refresh 续期；续期失败才清本地并返回 false。
 */
export async function bootstrapSession(): Promise<boolean> {
  if (!getToken()) return false
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
    // 网络异常：保留本地凭证，允许离线使用缓存
    return getLocalUser() ? true : false
  }
}