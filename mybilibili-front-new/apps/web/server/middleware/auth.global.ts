import { defineEventHandler, getRequestURL } from 'h3'

function decodeJwtPayload(token: string): { user_id?: number; role?: string } | null {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return null
    const b64 = parts[1]!.replace(/-/g, '+').replace(/_/g, '/')
    const padded = b64 + '='.repeat((4 - (b64.length % 4)) % 4)
    const json = Buffer.from(padded, 'base64').toString('utf8')
    return JSON.parse(json)
  } catch {
    return null
  }
}

export default defineEventHandler((event) => {
  const url = getRequestURL(event)
  if (!url.pathname.startsWith('/api/')) return
  const auth = event.node.req.headers['authorization']
  if (!auth || typeof auth !== 'string') return
  const m = auth.match(/^Bearer\s+(.+)$/i)
  if (!m) return
  const payload = decodeJwtPayload(m[1]!.trim())
  if (!payload || typeof payload.user_id !== 'number' || payload.user_id <= 0) return
  // 后端 manuscript handler 读 X-User-Id / X-Admin-Id 决定身份
  event.node.req.headers['x-user-id'] = String(payload.user_id)
  if (payload.role === 'admin') {
    event.node.req.headers['x-admin-id'] = String(payload.user_id)
  }
})
