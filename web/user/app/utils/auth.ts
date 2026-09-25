import { safeStorage } from '../utils/safeStorage'

/**
 * 阶段 3：凭证全部改为 HttpOnly cookie，JS 既读不到也写不了。
 *
 *   token / refresh_token / admin_token / admin_refresh —— HttpOnly，只由服务端 Set-Cookie
 *   user_info —— 可读，后端登录时下发的展示用 JSON（id/nickname/avatar），同时充当登录态信号
 *
 * JS 能拿到的登录态信号只有 user_info；"我是不是登录了"由它回答，
 * "我是谁"由它 + localStorage 的 user 回答。凭证本身永远不进 JS。
 * 能被 JS 读到的凭证就能被 XSS 整包偷走——这正是这一整轮要堵的洞。
 */

const USER_KEY = 'user'
const USER_INFO_COOKIE = 'user_info'

// 登录态相关的 localStorage key（全部废弃：凭证不落盘）
const LEGACY_TOKEN_KEY = 'teri_token'
const LEGACY_REFRESH_KEY = 'teri_refresh_token'
const LEGACY_STD_TOKEN_KEY = 'token'
const LEGACY_STD_REFRESH_KEY = 'refreshToken'

function readCookie(name: string): string {
  if (typeof document === 'undefined' || !document.cookie) return ''
  let found = ''
  for (const part of document.cookie.split(/;\s*/)) {
    const eq = part.indexOf('=')
    if (eq <= 0 || part.slice(0, eq) !== name) continue
    const value = part.slice(eq + 1)
    // 优先取非空值：同名 cookie 出现空值（残留的删除记录）时不能直接吞掉后面真正的值
    if (value) return value
    if (!found) found = value
  }
  return found
}

function removeClientCookie(name: string) {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=; Path=/; Max-Age=0`
}

// 登录接口的 data 里会同时带 token / refresh_token（为兼容旧客户端）。
// 调用方常常把整个 data 当 user 存——那样等于把 HttpOnly 想保护的东西
// 又写回可读存储，等于没做。所有写入/读取统一过这道清洗。
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

/** 后端下发的 user_info cookie（id / nickname / avatar）。 */
export function getSessionUser(): any {
  const raw = readCookie(USER_INFO_COOKIE)
  if (!raw) return null
  try {
    // 服务端用 url.QueryEscape 写入，+ 表示空格
    return scrubCredentials(JSON.parse(decodeURIComponent(raw.replace(/\+/g, '%20'))))
  } catch (error) {
    return null
  }
}

export function getStoredUser() {
  // 读路径也清洗：升级前写入的副本可能还带着 token
  return scrubCredentials(readJson(safeStorage.getItem(USER_KEY)))
}

export function getCurrentUserId() {
  const session = getSessionUser()
  if (session?.id) return session.id
  const user = getStoredUser()
  if (user?.id) return user.id
  if (user?.user_id) return user.user_id
  return null
}

export function hasAuthSession() {
  if (getSessionUser()) return true
  if (getStoredUser()) return true
  return false
}

/**
 * 记录会话的展示信息。**不接收凭证**：token/refreshToken 即使传进来也会被丢弃，
 * 它们只存在于 HttpOnly cookie 里，任何客户端存储都算泄露。
 */
export function setAuthSession(session: {
  token?: string
  refreshToken?: string
  user?: any
} = {}) {
  if (session.user) {
    safeStorage.setItem(USER_KEY, JSON.stringify(scrubCredentials(session.user)))
  }
}

/**
 * 清掉本地可见状态。HttpOnly cookie JS 删不掉，
 * 需要服务端作废凭证时请调用 api/session 的 clearServerSession()。
 */
export function clearAuthSession() {
  safeStorage.removeItem(USER_KEY)
  removeClientCookie(USER_INFO_COOKIE)
  // 顺手清掉升级前留下的凭证副本——它们是遗留的可窃取凭证明文
  safeStorage.removeItem(LEGACY_TOKEN_KEY)
  safeStorage.removeItem(LEGACY_REFRESH_KEY)
  safeStorage.removeItem(LEGACY_STD_TOKEN_KEY)
  safeStorage.removeItem(LEGACY_STD_REFRESH_KEY)
}

// ====== Admin ======
// admin 的 profile 是普通展示数据，可以留 localStorage；
// admin_token / admin_refresh 是凭证，不落盘。
const ADMIN_USER_KEY = 'admin_user'
const ADMIN_ROLE_KEY = 'admin_role'
const ADMIN_PERMISSIONS_KEY = 'admin_permissions'
const ADMIN_LEGACY_TOKEN_KEY = 'admin_token'

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
  safeStorage.removeItem(ADMIN_LEGACY_TOKEN_KEY)
}
