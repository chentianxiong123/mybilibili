import api from './client'

export const getUserList = (params: Record<string, any>) => api.get('/user/admin/list', { params })
export const getUserById = (id: number) => api.get(`/user/admin/${id}`)
export const updateUserStatus = (id: number, status: number) => api.put(`/user/admin/${id}/status`, { status })
export const resetPassword = (id: number, newPassword: string) => api.put(`/user/admin/${id}/password`, { newPassword })