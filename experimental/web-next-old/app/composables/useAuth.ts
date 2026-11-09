import { getStoredUser, getToken, getRefreshToken } from '~/utils/auth'

export interface AuthUser {
  id?: number
  nickname?: string
  username?: string
  avatar?: string
  [key: string]: any
}

/**
 * SSR 安全的认证状态读取。
 *
 * - 服务端：从请求头 Cookie 里读 token/user_info（由 Go 登录接口 Set-Cookie 写入），
 *   与真实登录态一致，水合时不再 mismatch。
 * - 客户端：优先读 Cookie（与 SSR 一致），回退到 localStorage（兼容旧数据）。
 */
export const useAuth = () => {
  const cookieToken = useCookie<string | null>('token', { default: () => null })
  const cookieRefresh = useCookie<string | null>('refresh_token', { default: () => null })
  const cookieUserRaw = useCookie<string | null>('user_info', { default: () => null })

  const cookieUser = computed<AuthUser | null>(() => {
    if (cookieUserRaw.value) {
      try {
        const decoded = decodeURIComponent(cookieUserRaw.value)
        return JSON.parse(decoded)
      } catch {
        return null
      }
    }
    return null
  })

  // 兜底：Cookie 缺失时用 localStorage（兼容旧登录数据），仅客户端有效。
  const user = computed<AuthUser | null>(() => {
    if (cookieUser.value) return cookieUser.value
    if (import.meta.client) return getStoredUser()
    return null
  })

  const token = computed<string | null>(() => {
    if (cookieToken.value) return cookieToken.value
    if (import.meta.client) return getToken() || null
    return null
  })

  const refreshToken = computed<string | null>(() => {
    if (cookieRefresh.value) return cookieRefresh.value
    if (import.meta.client) return getRefreshToken() || null
    return null
  })

  const isLoggedIn = computed<boolean>(() => Boolean(token.value))

  return { token, refreshToken, user, isLoggedIn }
}

export default useAuth