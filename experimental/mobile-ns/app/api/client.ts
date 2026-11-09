import storage from '../utils/storage'

const BASE_URL = 'http://192.168.31.204:8080'

function showToast(msg: string) {
  console.log('[Toast]', msg)
}

interface ApiResponse {
  code: string
  data?: any
  message?: string
}

async function request<T = any>(url: string, options: RequestInit = {}): Promise<ApiResponse> {
  const token = storage.getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Client-Platform': 'wap',
    ...(options.headers as Record<string, string> || {})
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const userStr = storage.getItem('user')
  if (userStr) {
    try {
      const localUser = JSON.parse(userStr)
      const userId = localUser.id || localUser.userId || localUser.user?.id
      if (userId) {
        headers['X-User-Id'] = String(userId)
      }
    } catch {}
  }

  try {
    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      headers
    })
    const data = await response.json()
    return data
  } catch (error: any) {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          storage.removeToken()
          storage.removeItem('user')
          break
        case 403:
          showToast('没有权限访问该资源')
          break
        case 404:
          showToast('请求的资源不存在')
          break
        case 500:
          showToast('服务器内部错误')
          break
        default:
          showToast(error.response.data?.message || '请求失败')
      }
    } else {
      showToast('网络错误，请检查网络连接')
    }
    return { code: '0', message: error.message }
  }
}

const api = {
  get<T = any>(url: string): Promise<ApiResponse> {
    return request<T>(url, { method: 'GET' })
  },

  post<T = any>(url: string, data?: any): Promise<ApiResponse> {
    return request<T>(url, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined
    })
  },

  put<T = any>(url: string, data?: any): Promise<ApiResponse> {
    return request<T>(url, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined
    })
  },

  delete<T = any>(url: string): Promise<ApiResponse> {
    return request<T>(url, { method: 'DELETE' })
  }
}

export default api
export { BASE_URL }