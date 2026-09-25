import { safeStorage } from '../utils/safeStorage'

/**
 * 阶段 3：凭证（admin_token / admin_refresh / token / refresh_token）全部是
 * HttpOnly cookie，JS 既读不到也写不了——能读到就能被 XSS 整包偷走。
 *
 * 这里保留的只有两类非凭证信息：
 *   - admin_user / admin_role / admin_permissions：登录时服务端返回的展示与授权数据
 *   - user：兼容老代码的用户展示信息
 */

const USER_KEY = 'user'
const ADMIN_USER_KEY = 'admin_user'
const ADMIN_ROLE_KEY = 'admin_role'
const ADMIN_PERMISSIONS_KEY = 'admin_permissions'

// 升级前遗留的可读凭证明文，见 clearAuthSession / clearAdminSession
const LEGACY_TOKEN_KEY = 'token'
const LEGACY_REFRESH_KEY = 'refreshToken'
const LEGACY_TERRI_TOKEN_KEY = 'teri_token'
const LEGACY_ADMIN_TOKEN_KEY = 'admin_token'

// 登录响应的 data 里会带 token（兼容旧客户端）；别让它跟着展示信息进 localStorage。
const CREDENTIAL_KEYS = new Set([
  'token', 'refresh_token', 'refreshToken', 'access_token', 'accessToken',
  'password', 'pwd'
])

export function scrubCredentials<T>(value: T): T {
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

function readJson(value) {
  if (!value) return null
  try {
    return JSON.parse(value)
  } catch (error) {
    return null
  }
}

export function getStoredUser() {
  return scrubCredentials(readJson(safeStorage.getItem(USER_KEY)))
}

export function getCurrentUserId() {
  const user = getStoredUser()
  return user?.id ?? null
}

export function hasAuthSession() {
  return Boolean(getStoredUser())
}

/** 只保留展示信息；token / refreshToken 即使传进来也丢弃。 */
export function setAuthSession(session: {
  token?: string
  refreshToken?: string
  user?: any
} = {}) {
  if (session.user) {
    safeStorage.setItem(USER_KEY, JSON.stringify(scrubCredentials(session.user)))
  }
}

export function clearAuthSession() {
  safeStorage.removeItem(USER_KEY)
  safeStorage.removeItem(LEGACY_TOKEN_KEY)
  safeStorage.removeItem(LEGACY_REFRESH_KEY)
  safeStorage.removeItem(LEGACY_TERRI_TOKEN_KEY)
}

// ====== Admin ======

export function getAdminUser(): any {
  return scrubCredentials(readJson(safeStorage.getItem(ADMIN_USER_KEY)))
}

export function getAdminRole(): string {
  return safeStorage.getItem(ADMIN_ROLE_KEY) || ''
}

export function getAdminPermissions(): string[] {
  try {
    return JSON.parse(safeStorage.getItem(ADMIN_PERMISSIONS_KEY) || '[]')
  } catch {
    return []
  }
}

/** 登录态由 admin_user 这份展示信息回答；凭证本身在 HttpOnly cookie 里。 */
export function hasAdminSession(): boolean {
  return Boolean(getAdminUser())
}

export function setAdminSession(data: {
  token?: string
  user?: any
  role?: string
  permissions?: string[]
}): void {
  if (data.user) safeStorage.setItem(ADMIN_USER_KEY, JSON.stringify(scrubCredentials(data.user)))
  if (data.role) safeStorage.setItem(ADMIN_ROLE_KEY, data.role)
  if (data.permissions) safeStorage.setItem(ADMIN_PERMISSIONS_KEY, JSON.stringify(data.permissions))
}

export function clearAdminSession(): void {
  safeStorage.removeItem(ADMIN_USER_KEY)
  safeStorage.removeItem(ADMIN_ROLE_KEY)
  safeStorage.removeItem(ADMIN_PERMISSIONS_KEY)
  safeStorage.removeItem(LEGACY_ADMIN_TOKEN_KEY)
}
