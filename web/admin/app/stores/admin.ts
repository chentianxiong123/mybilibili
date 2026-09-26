import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminLogin } from '@/api/admin'
import { clearServerSession } from '@/api/session'
import { scrubCredentials } from '@/utils/auth'

export const useAdminStore = defineStore('admin', () => {
  // 跳转责任收归 store：之前散落在 app.vue / NoPermissionView 的 router.push('/login')
  // 任何一处遗漏都会出现「logout 200 但页面停在 dashboard」的 UX bug。
  // store 里读不到 router 时（比如单测）降级为 no-op，不影响功能。
  let router: ReturnType<typeof useRouter> | null = null
  try { router = useRouter() } catch { /* 不在 router 上下文 */ }

  // 凭证 admin_token 是 HttpOnly cookie，store 里只留展示与授权信息
  const userInfo = ref(scrubCredentials(JSON.parse(localStorage.getItem('admin_user')) || null))
  const role = ref(localStorage.getItem('admin_role') || '')
  const permissions = ref(JSON.parse(localStorage.getItem('admin_permissions') || '[]'))

  const login = async (loginData) => {
    try {
      const res = await adminLogin(loginData)
      if (res.code === 200 || res.success) {
        // 后台登录响应给的是 admin_id/username/role，没有 adminUser 对象，
        // 这里拼出一份展示信息——它同时是前端判断登录态的唯一信号。
        const adminId = res.data.adminUser?.id || res.data.user?.id || res.data.admin_id || res.data.adminId || null
        userInfo.value = res.data.adminUser || res.data.user || res.user
          || { id: adminId, username: res.data.username || loginData.username }
        role.value = res.data.role || '管理员'
        permissions.value = res.data.permissions || []
        // admin_user 会被 persist，登录响应里的 token 不能跟着进去
        localStorage.setItem('admin_user', JSON.stringify(scrubCredentials(userInfo.value)))
        localStorage.setItem('admin_role', role.value)
        localStorage.setItem('admin_permissions', JSON.stringify(permissions.value))
        if (adminId) localStorage.setItem('admin_id', adminId)
        return { success: true }
      }
      // 业务级失败：服务端返回的 message 是中文（"账号或密码错误" 等），
      // 不要再回退到英文 axios 报错
      return { success: false, message: res.message || '登录失败' }
    } catch (error: any) {
      // axios 抛错：优先拿服务端 body 里的 message，否则按状态码给友好提示，
      // 最后才回退到 axios 的英文 message（避免给用户看 "Request failed with status code 401"）
      const serverMsg = error?.response?.data?.message as string | undefined
      const status = error?.response?.status as number | undefined
      const friendly = status === 401 ? '账号或密码错误'
        : status === 400 ? '请求参数错误'
        : status === 429 ? '请求过于频繁，请稍后再试'
        : status && status >= 500 ? '服务器开小差，请稍后重试'
        : !error?.response ? '网络错误，请检查网络连接'
        : ''
      // 优先用服务端 message（最准确，多是中文）；然后按状态码推断的友好提示（避免
      // 把 axios 的 "Request failed with status code 401" / "Network Error" 给用户看）；
      // 最后回退到 error.message（caller 自定义 message），兜底 "登录失败"。
      // 注意：no-response 情况（!error.response）一律视作网络层失败，caller 自定义
      // message 在这种场景会被覆盖——这是为修复 #6 UX bug 故意做的取舍。
      return { success: false, message: serverMsg || friendly || error?.message || '登录失败' }
    }
  }

  const logout = async () => {
    // 先让服务端作废 HttpOnly 的 admin_token / admin_refresh，否则 cookie 会留到过期
    await clearServerSession()
    userInfo.value = null
    role.value = ''
    permissions.value = []
    localStorage.removeItem('admin_token') // 清掉升级前的遗留副本
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_role')
    localStorage.removeItem('admin_permissions')
    localStorage.removeItem('admin_id')
    // 跳转由 store 兜底，调用方不用关心。已在 /login 上时 push 会触发导航守卫 false，吞掉
    if (router) {
      try { await router.push('/login') } catch { /* 已在 /login */ }
    }
  }

  const hasPermission = (permission) => {
    if (!permission) return true
    if (role.value === '超级管理员') return true
    return permissions.value.includes(permission)
  }

  return {
    userInfo,
    role,
    permissions,
    hasPermission,
    login,
    logout
  }
})
