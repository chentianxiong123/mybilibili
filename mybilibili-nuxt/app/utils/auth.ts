import { safeStorage } from '../utils/safeStorage'
const TOKEN_KEY = 'token'
const REFRESH_TOKEN_KEY = 'refreshToken'
const USER_KEY = 'user'

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
  return safeStorage.getItem(TOKEN_KEY) || ''
}

export function getRefreshToken() {
  return safeStorage.getItem(REFRESH_TOKEN_KEY) || ''
}

export function getStoredUser() {
  return readJson(safeStorage.getItem(USER_KEY))
}

export function getCurrentUserId() {
  const user = getStoredUser()
  if (user?.id) return user.id

  const payload = decodeJwtPayload()
  return payload?.sub || payload?.userId || null
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
  if (session.token) {
    safeStorage.setItem(TOKEN_KEY, session.token)
  }

  if (session.refreshToken) {
    safeStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken)
  }

  if (session.user) {
    safeStorage.setItem(USER_KEY, JSON.stringify(session.user))
  }
}

export function clearAuthSession() {
  safeStorage.removeItem(TOKEN_KEY)
  safeStorage.removeItem(REFRESH_TOKEN_KEY)
  safeStorage.removeItem(USER_KEY)
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
