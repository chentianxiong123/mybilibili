import { safeStorage } from '../utils/safeStorage'
const TOKEN_KEY = 'token'
const REFRESH_TOKEN_KEY = 'refreshToken'
const USER_KEY = 'user'
// teriteri 老代码（space/account/platform 等页面与 getPersonalInfo）直接读写 localStorage，
// 与上面这套 key 并存。这里做双向桥接，避免"登了录但两套 key 只写了一套"导致请求无 token。
const LEGACY_TOKEN_KEY = 'teri_token'
const LEGACY_REFRESH_KEY = 'teri_refresh_token'

// ====== Cookie 同步（SSR 读取用，与 localStorage 保持一致） ======
function setClientCookie(name: string, value: string, maxAge: number) {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAge}; SameSite=Lax`
}

function removeClientCookie(name: string) {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=; Path=/; Max-Age=0`
}

function readJson(value) {
  if (!value) return null
  try {
    return JSON.parse(value)
  } catch (error) {
    return null
  }
}

function decodeBase64Url(value) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized.padEnd(
    normalized.length + ((4 - (normalized.length % 4)) % 4),
    '='
  )
  return atob(padded)
}

export function decodeJwtPayload(token = getToken()) {
  if (!token) return null
  const [, payload] = token.split('.')
  if (!payload) return null
  try {
    return JSON.parse(decodeBase64Url(payload))
  } catch (error) {
    return null
  }
}

export function getToken() {
  return safeStorage.getItem(TOKEN_KEY) || safeStorage.getItem(LEGACY_TOKEN_KEY) || ''
}

export function getRefreshToken() {
  return safeStorage.getItem(REFRESH_TOKEN_KEY) || safeStorage.getItem(LEGACY_REFRESH_KEY) || ''
}

export function getStoredUser() {
  return readJson(safeStorage.getItem(USER_KEY))
}

export function getCurrentUserId() {
  const user = getStoredUser()
  if (user?.id) return user.id

  const payload = decodeJwtPayload()
  return payload?.sub || payload?.userId || payload?.user_id || null
}

export function isAccessTokenExpired(token = getToken(), leewayMs = 0) {
  const payload = decodeJwtPayload(token)
  if (!payload?.exp) return true
  return payload.exp * 1000 <= Date.now() + leewayMs
}

export function hasValidAccessToken() {
  const token = getToken()
  return Boolean(token) && !isAccessTokenExpired(token)
}

export function hasAuthSession() {
  return Boolean(getToken() || getRefreshToken())
}

export function setAuthSession(session: {
  token?: string
  refreshToken?: string
  user?: any
} = {}) {
  // token / refresh_token 由服务端以 HttpOnly 下发，JS 不再写这两个 cookie：
  // document.cookie 的写入会把 HttpOnly 标记冲掉，等于自己拆掉 XSS 防线。
  // 这里只保留 localStorage 作为请求头与登录态展示的来源（阶段 3 再切换到纯 cookie）。
  if (session.token) {
    safeStorage.setItem(TOKEN_KEY, session.token)
    safeStorage.setItem(LEGACY_TOKEN_KEY, session.token)
  }

  if (session.refreshToken) {
    safeStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken)
    safeStorage.setItem(LEGACY_REFRESH_KEY, session.refreshToken)
  } else if (session.token) {
    // 换发了新 token 但没带 refreshToken：清掉旧的，避免拿到上一次会话的失效 refresh token
    safeStorage.removeItem(REFRESH_TOKEN_KEY)
    safeStorage.removeItem(LEGACY_REFRESH_KEY)
  }

  if (session.user) {
    const userStr = JSON.stringify(session.user)
    safeStorage.setItem(USER_KEY, userStr)
    setClientCookie('user_info', encodeURIComponent(userStr), 86400 * 7)
  }
}

// clearAuthSession 只清本地可见状态。
// HttpOnly 的 token / refresh_token JS 删不掉，需要服务端清的场合
// 请调用 api/session 的 clearServerSession()（登出与凭证确认失效时）。
export function clearAuthSession() {
  safeStorage.removeItem(TOKEN_KEY)
  safeStorage.removeItem(REFRESH_TOKEN_KEY)
  safeStorage.removeItem(USER_KEY)
  safeStorage.removeItem(LEGACY_TOKEN_KEY)
  safeStorage.removeItem(LEGACY_REFRESH_KEY)
  removeClientCookie('user_info')
}

// ====== Admin Auth ======
const ADMIN_TOKEN_KEY = 'admin_token'
const ADMIN_USER_KEY = 'admin_user'
const ADMIN_ROLE_KEY = 'admin_role'
const ADMIN_PERMISSIONS_KEY = 'admin_permissions'

export function getAdminToken(): string {
  return safeStorage.getItem(ADMIN_TOKEN_KEY) || ''
}

export function getAdminUser(): any {
  return readJson(safeStorage.getItem(ADMIN_USER_KEY))
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

export function hasAdminSession(): boolean {
  return Boolean(getAdminToken())
}

export function setAdminSession(data: {
  token: string
  user?: any
  role?: string
  permissions?: string[]
}): void {
  safeStorage.setItem(ADMIN_TOKEN_KEY, data.token)
  if (data.user) safeStorage.setItem(ADMIN_USER_KEY, JSON.stringify(data.user))
  if (data.role) safeStorage.setItem(ADMIN_ROLE_KEY, data.role)
  if (data.permissions) safeStorage.setItem(ADMIN_PERMISSIONS_KEY, JSON.stringify(data.permissions))
}

export function clearAdminSession(): void {
  safeStorage.removeItem(ADMIN_TOKEN_KEY)
  safeStorage.removeItem(ADMIN_USER_KEY)
  safeStorage.removeItem(ADMIN_ROLE_KEY)
  safeStorage.removeItem(ADMIN_PERMISSIONS_KEY)
}
