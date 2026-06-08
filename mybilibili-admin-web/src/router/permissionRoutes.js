export const permissionRouteOrder = [
  { path: '/dashboard', permission: 'statistics:manage' },
  { path: '/operation-tasks', permission: 'operation:manage' },
  { path: '/support-tickets', permission: 'operation:manage' },
  { path: '/index-manager', permission: 'search:manage' },
  { path: '/recommend-config', permission: 'search:manage' },
  { path: '/audit-logs', permission: 'audit:manage' },
  { path: '/manuscripts', permission: 'review:manage' },
  { path: '/content-review', permission: 'review:manage' },
  { path: '/prohibited-words', permission: 'comment:manage' },
  { path: '/ai-usage', permission: 'ai:manage' },
  { path: '/ai-skills', permission: 'ai:manage' },
  { path: '/api-management', permission: 'ai:manage' },
  { path: '/customer-chat', permission: 'ai:manage' },
  { path: '/categories', permission: 'category:manage' },
  { path: '/banner-images', permission: 'banner:manage' },
  { path: '/subtitles', permission: 'video:manage' },
  { path: '/live-rooms', permission: 'live:manage' },
  { path: '/meeting-admin', permission: 'meeting:manage' },
  { path: '/users', permission: 'user:manage' },
  { path: '/login-logs', permission: 'security:manage' }
]

export const firstAllowedPathByPermissions = (role, permissions = []) => {
  if (role === '超级管理员') return '/dashboard'
  const allowed = permissionRouteOrder.find(item => permissions.includes(item.permission))
  return allowed?.path || '/no-permission'
}
