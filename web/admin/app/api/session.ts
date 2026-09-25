/**
 * 让服务端作废 HttpOnly 登录 cookie。
 *
 * 阶段 3 起 admin_token / admin_refresh 只存在于 HttpOnly cookie，JS 读不到也删不掉，
 * 所以"退出登录"必须调服务端接口，否则会出现看起来已登出、接口仍带凭证的假登出。
 * 失败静默：登出是幂等的。
 */
export async function clearServerSession(path = '/api/v1/admin/logout'): Promise<void> {
  if (typeof fetch !== 'function') return
  try {
    await fetch(path, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' }
    })
  } catch {
    // 忽略：下次登录会覆盖同名 cookie
  }
}

export default clearServerSession
