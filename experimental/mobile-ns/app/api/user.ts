import api from './client'

export function login(username: string, password: string) {
  return api.post('/api/user/login', { username, password })
}

export function register(username: string, password: string) {
  return api.post('/api/user/register', { username, password })
}

export function getUserInfo() {
  return api.get('/api/user/info')
}

export function updateUserInfo(data: any) {
  return api.put('/api/user/profile', data)
}

export function logout() {
  return api.post('/api/user/logout')
}