/**
 * 让服务端清掉 HttpOnly 登录 cookie。
 *
 * 阶段 1 起 token / refresh_token 由 Go 侧以 HttpOnly 下发，JS 既读不到
 * 也删不掉，所以"退出登录"必须调服务端接口把它们作废，
 * 否则会出现看起来已登出、接口却仍然带凭证的假登出。
 *
 * 失败静默：登出是幂等的，清不掉也不该打断用户。
 */
export async function clearServerSession(path = '/api/v1/user/logout'): Promise<void> {
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
