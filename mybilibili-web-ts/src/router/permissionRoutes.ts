export const permissionRouteOrder = [
  { path: '/admin/dashboard', permission: 'statistics:manage' },
  { path: '/admin/operation-tasks', permission: 'operation:manage' },
  { path: '/admin/support-tickets', permission: 'operation:manage' },
  { path: '/admin/index-manager', permission: 'search:manage' },
  { path: '/admin/recommend-config', permission: 'search:manage' },
  { path: '/admin/audit-logs', permission: 'audit:manage' },
  { path: '/admin/manuscripts', permission: 'review:manage' },
  { path: '/admin/content-review', permission: 'review:manage' },
  { path: '/admin/prohibited-words', permission: 'comment:manage' },
  { path: '/admin/ai-usage', permission: 'ai:manage' },
  { path: '/admin/ai-skills', permission: 'ai:manage' },
  { path: '/admin/api-management', permission: 'ai:manage' },
  { path: '/admin/customer-chat', permission: 'ai:manage' },
  { path: '/admin/categories', permission: 'category:manage' },
  { path: '/admin/banner-images', permission: 'banner:manage' },
  { path: '/admin/subtitles', permission: 'video:manage' },
  { path: '/admin/live-rooms', permission: 'live:manage' },
  { path: '/admin/meeting-admin', permission: 'meeting:manage' },
  { path: '/admin/users', permission: 'user:manage' },
  { path: '/admin/login-logs', permission: 'security:manage' }
]

export const firstAllowedPathByPermissions = (role: string, permissions: string[] = []): string => {
  if (role === '超级管理员') return '/admin/dashboard'
  const allowed = permissionRouteOrder.find(item => permissions.includes(item.permission))
  return allowed?.path || '/admin/no-permission'
}
