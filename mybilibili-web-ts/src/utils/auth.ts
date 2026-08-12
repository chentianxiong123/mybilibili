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
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY) || ''
}

export function getStoredUser() {
  return readJson(localStorage.getItem(USER_KEY))
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
    localStorage.setItem(TOKEN_KEY, session.token)
  }

  if (session.refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken)
  }

  if (session.user) {
    localStorage.setItem(USER_KEY, JSON.stringify(session.user))
  }
}

export function clearAuthSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

// ====== Admin Auth ======
const ADMIN_TOKEN_KEY = 'admin_token'
const ADMIN_USER_KEY = 'admin_user'
const ADMIN_ROLE_KEY = 'admin_role'
const ADMIN_PERMISSIONS_KEY = 'admin_permissions'

export function getAdminToken(): string {
  return localStorage.getItem(ADMIN_TOKEN_KEY) || ''
}

export function getAdminUser(): any {
  return readJson(localStorage.getItem(ADMIN_USER_KEY))
}

export function getAdminRole(): string {
  return localStorage.getItem(ADMIN_ROLE_KEY) || ''
}

export function getAdminPermissions(): string[] {
  try {
    return JSON.parse(localStorage.getItem(ADMIN_PERMISSIONS_KEY) || '[]')
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
  localStorage.setItem(ADMIN_TOKEN_KEY, data.token)
  if (data.user) localStorage.setItem(ADMIN_USER_KEY, JSON.stringify(data.user))
  if (data.role) localStorage.setItem(ADMIN_ROLE_KEY, data.role)
  if (data.permissions) localStorage.setItem(ADMIN_PERMISSIONS_KEY, JSON.stringify(data.permissions))
}

export function clearAdminSession(): void {
  localStorage.removeItem(ADMIN_TOKEN_KEY)
  localStorage.removeItem(ADMIN_USER_KEY)
  localStorage.removeItem(ADMIN_ROLE_KEY)
  localStorage.removeItem(ADMIN_PERMISSIONS_KEY)
}
