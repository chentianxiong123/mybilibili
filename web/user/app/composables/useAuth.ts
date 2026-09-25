import { getStoredUser } from '~/utils/auth'

export interface AuthUser {
  id?: number
  nickname?: string
  username?: string
  avatar?: string
  [key: string]: any
}

/**
 * SSR 安全的登录态读取。
 *
 * 只读 user_info（后端下发的可读展示信息）。凭证 token / refresh_token 是 HttpOnly，
 * 绝不能经 useCookie 或 payload 暴露给客户端——那等于把 HttpOnly 又拆了。
 *
 * - 服务端：直接解析请求头 Cookie，与真实登录态一致，水合时不再 mismatch。
 * - 客户端：优先读 Cookie（与 SSR 一致），回退到 localStorage（兼容旧数据）。
 */
export const useAuth = () => {
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

  const isLoggedIn = computed<boolean>(() => Boolean(user.value))

  return { user, isLoggedIn }
}

export default useAuth