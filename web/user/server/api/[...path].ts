import { defineEventHandler, getRequestURL, getMethod, getRequestHeaders, readRawBody, setResponseStatus, setResponseHeaders } from 'h3'

// teriteri 前端 URL → Go 后端真实路径映射（来自 mybilibili-refactor 验证过的 PATH_MAP）
const PATH_MAP: Record<string, string> = {
  '/video/random/visitor': '/manuscript/recommended',
  '/video/cumulative/visitor': '/manuscript/hot',
  '/video/getone': '/manuscript/',
  '/video/user-works': '/manuscript/user',
  '/video/user-love': '/manuscript/user/likes',
  '/video/user-collect': '/manuscript/user/collections',
  '/category/getall': '/category',
  '/search/hot/get': '/search/hot',
  '/search/word/add': '/search/hot/increment',
  '/search/word/get': '/search/suggest',
  '/search/video/only-pass': '/search/videos',
  '/search/user': '/search/users',
  '/search/count': '/search/count',
  '/video/love-or-not': '/interaction/video/love-or-not',
  '/comment/get': '/comment/list',
  '/comment/add': '/comment/add',
  '/comment/get-up-like': '/comment/get-up-like',
  '/comment/love-or-not': '/comment/',
  '/danmu-list': '/danmaku/video',
  '/favorite/get-all/user': '/favorites',
  '/favorite/get-all/visitor': '/favorites',
  '/user/info/get-one': '/user/',
  '/user/info': '/user/',
  '/video/add': '/manuscript/upload-complete',
  '/video/upload-chunk': '/manuscript/upload-chunk',
  '/video/ask-chunk': '/manuscript/upload-session',
  '/video/cancel-upload': '/manuscript/upload-session',
  '/user/account/login': '/user/login',
  '/user/account/register': '/user/register',
  '/user/account/logout': '/user/logout',
}

// 按服务端口分发的路由（Docker 环境用容器名，宿主机直连用 127.0.0.1）
const CORE_HOST = process.env.CORE_HOST || '127.0.0.1'
const SEARCH_HOST = process.env.SEARCH_HOST || '127.0.0.1'
const MSG_HOST = process.env.MSG_HOST || '127.0.0.1'
const LIVE_HOST = process.env.LIVE_HOST || '127.0.0.1'

function pickUpstream(realPath: string): string {
  if (realPath.startsWith('/search/')) return `http://${SEARCH_HOST}:8084`
  if (realPath.includes('/danmaku/') || realPath.startsWith('/danmu') || realPath.startsWith('/creator/danmaku')) return `http://${MSG_HOST}:8086`
  return `http://${CORE_HOST}:8080`
}

function adaptUrl(url: string, query: URLSearchParams): { target: string; port: string; qs: string } {
  // /user/info/get-one?uid=X → /user/X（Go 后端路径参数）
  if (url.startsWith('/user/info/get-one')) {
    const uid = query.get('uid') || ''
    return { target: `/user/${uid}`, port: `http://${CORE_HOST}:8080`, qs: '' }
  }
  // /video/user-works-count?uid=X → /manuscript/user/X（取 total 作为投稿数）
  if (url.startsWith('/video/user-works-count')) {
    const uid = query.get('uid') || ''
    const qs = new URLSearchParams({ page: '1', pageSize: '1' }).toString()
    return { target: `/manuscript/user/${uid}`, port: `http://${CORE_HOST}:8080`, qs }
  }
  // /video/getone 特殊：vid 从 query 或路径
  if (url.startsWith('/video/getone')) {
    const vid = query.get('vid') || url.split('/').pop() || ''
    return { target: `/manuscript/${vid}`, port: `http://${CORE_HOST}:8080`, qs: '' }
  }
  // /video/user-works → /manuscript/user/{uid}，参数 page/quantity → page/pageSize
  if (url.startsWith('/video/user-works')) {
    const uid = query.get('uid') || ''
    const page = query.get('page') || '1'
    const size = query.get('quantity') || query.get('pageSize') || '20'
    const qs = new URLSearchParams({ page, pageSize: size }).toString()
    return { target: `/manuscript/user/${uid}`, port: `http://${CORE_HOST}:8080`, qs }
  }
  // /video/user-love → /manuscript/user/likes（需鉴权）
  if (url.startsWith('/video/user-love')) {
    const offset = Number(query.get('offset')) || 0
    const page = String(Math.floor(offset / 20) + 1)
    const size = query.get('quantity') || query.get('pageSize') || '20'
    const qs = new URLSearchParams({ page, pageSize: size }).toString()
    return { target: '/manuscript/user/likes', port: `http://${CORE_HOST}:8080`, qs }
  }
  // /video/user-collect?fid=X → /favorites/X/videos（收藏夹内视频）
  if (url.startsWith('/video/user-collect')) {
    const fid = query.get('fid') || ''
    if (fid) {
      const page = query.get('page') || '1'
      const size = query.get('quantity') || query.get('pageSize') || '20'
      const rule = query.get('rule') || '1'
      const qs = new URLSearchParams({ page, pageSize: size, rule }).toString()
      return { target: `/favorites/${fid}/videos`, port: `http://${CORE_HOST}:8080`, qs }
    }
    // 无 fid 时 fallback 到收藏的稿件列表
    const qs = new URLSearchParams({ page: '1', pageSize: '20' }).toString()
    return { target: '/manuscript/user/collections', port: `http://${CORE_HOST}:8080`, qs }
  }
  // /video/cumulative/visitor → /manuscript/hot，透传 seed/offset
  if (url.startsWith('/video/cumulative/visitor')) {
    const params = new URLSearchParams()
    const seed = query.get('seed')
    const offset = query.get('offset')
    if (seed) params.set('seed', seed)
    if (offset) params.set('offset', offset)
    return { target: '/manuscript/hot', port: `http://${CORE_HOST}:8080`, qs: params.toString() }
  }
  // /video/collected-fids?vid=X → /favorites/manuscript/X（稿件所在收藏夹列表）
  if (url.startsWith('/video/collected-fids')) {
    const vid = query.get('vid') || query.get('manuscript_id') || ''
    if (vid) {
      return { target: `/favorites/manuscript/${vid}`, port: `http://${CORE_HOST}:8080`, qs: '' }
    }
  }
  const keys = Object.keys(PATH_MAP).sort((a, b) => b.length - a.length)
  for (const from of keys) {
    if (url === from || url.startsWith(from + '/') || url.startsWith(from + '?')) {
      const real = url.replace(from, PATH_MAP[from]!)
      const qs = query.toString()
      return { target: real, port: pickUpstream(real), qs }
    }
  }
  // 搜索历史：接口在 core（8080），不走 search 服务的 /search/ 前缀
  if (url.startsWith('/search/history')) {
    return { target: '/search/history', port: `http://${CORE_HOST}:8080`, qs: '' }
  }
  return { target: url, port: pickUpstream(url), qs: '' }
}

export default defineEventHandler(async (event) => {
  const fullUrl = getRequestURL(event)
  const path = fullUrl.pathname

  if (!path.startsWith('/api/')) {
    setResponseStatus(event, 404)
    return { error: true, statusCode: 404, statusMessage: 'Not an API prefix', path }
  }

  // 兼容 /api/v1/xxx（client.ts 用 /api/v1）与 /api/xxx（teriteri 旧调用）
  const apiPrefix = path.startsWith('/api/v1/') ? '/api/v1' : '/api'
  const teriteriUrl = path.slice(apiPrefix.length) || '/'
  const query = new URLSearchParams(fullUrl.search || '')
  const method = getMethod(event)

  // /video/cancel-collect → DELETE /favorites/{fid}/manuscripts/{vid}
  if (teriteriUrl.startsWith('/video/cancel-collect') && method === 'POST') {
    const body = await readRawBody(event)
    const params = new URLSearchParams(body || '')
    const fid = params.get('fid') || ''
    const vid = params.get('vid') || ''
    if (fid && vid) {
      const targetUrl = `http://${CORE_HOST}:8080/api/v1/favorites/${fid}/manuscripts/${vid}`
      const headers: Record<string, string> = {}
      const reqHeaders = getRequestHeaders(event)
      for (const k of Object.keys(reqHeaders)) {
        const v = reqHeaders[k]
        if (v !== undefined && k.toLowerCase() !== 'host' && k.toLowerCase() !== 'content-length') {
          headers[k] = String(v)
        }
      }
      headers['host'] = new URL(`http://${CORE_HOST}:8080`).host
      const resp = await fetch(targetUrl, { method: 'DELETE', headers, redirect: 'follow' })
      setResponseStatus(event, resp.status)
      return Buffer.from(await resp.arrayBuffer())
    }
  }

  // /favorite/create → POST /favorites（创建收藏夹）
  if (teriteriUrl === '/favorite/create' && method === 'POST') {
    const body = await readRawBody(event)
    const targetUrl = `http://${CORE_HOST}:8080/api/v1/favorites`
    const headers: Record<string, string> = { 'content-type': 'application/json' }
    const reqHeaders = getRequestHeaders(event)
    for (const k of Object.keys(reqHeaders)) {
      const v = reqHeaders[k]
      if (v !== undefined && k.toLowerCase() !== 'host' && k.toLowerCase() !== 'content-length') {
        headers[k] = String(v)
      }
    }
    headers['host'] = new URL(`http://${CORE_HOST}:8080`).host
    const resp = await fetch(targetUrl, { method: 'POST', headers, body, redirect: 'follow' })
    setResponseStatus(event, resp.status)
    return Buffer.from(await resp.arrayBuffer())
  }

  // /favorite/update/{id} → PUT /favorites/{id}（编辑收藏夹）
  if (teriteriUrl.startsWith('/favorite/update/') && method === 'POST') {
    const id = teriteriUrl.split('/').pop()
    const body = await readRawBody(event)
    const targetUrl = `http://${CORE_HOST}:8080/api/v1/favorites/${id}`
    const headers: Record<string, string> = { 'content-type': 'application/json' }
    const reqHeaders = getRequestHeaders(event)
    for (const k of Object.keys(reqHeaders)) {
      const v = reqHeaders[k]
      if (v !== undefined && k.toLowerCase() !== 'host' && k.toLowerCase() !== 'content-length') {
        headers[k] = String(v)
      }
    }
    headers['host'] = new URL(`http://${CORE_HOST}:8080`).host
    const resp = await fetch(targetUrl, { method: 'PUT', headers, body, redirect: 'follow' })
    setResponseStatus(event, resp.status)
    return Buffer.from(await resp.arrayBuffer())
  }

  // /favorite/delete/{id} → DELETE /favorites/{id}（删除收藏夹）
  if (teriteriUrl.startsWith('/favorite/delete/') && method === 'POST') {
    const id = teriteriUrl.split('/').pop()
    const targetUrl = `http://${CORE_HOST}:8080/api/v1/favorites/${id}`
    const headers: Record<string, string> = {}
    const reqHeaders = getRequestHeaders(event)
    for (const k of Object.keys(reqHeaders)) {
      const v = reqHeaders[k]
      if (v !== undefined && k.toLowerCase() !== 'host' && k.toLowerCase() !== 'content-length') {
        headers[k] = String(v)
      }
    }
    headers['host'] = new URL(`http://${CORE_HOST}:8080`).host
    const resp = await fetch(targetUrl, { method: 'DELETE', headers, redirect: 'follow' })
    setResponseStatus(event, resp.status)
    return Buffer.from(await resp.arrayBuffer())
  }

  let mapped: { target: string; port: string; qs: string }
  try {
    mapped = adaptUrl(teriteriUrl, query)
  } catch (e) {
    setResponseStatus(event, 500)
    return { code: 500, message: `adaptUrl failed: ${e}`, data: null }
  }

  const target = mapped.port + '/api/v1' + mapped.target + (mapped.qs ? '?' + mapped.qs : '')

  const headers: Record<string, string> = {}
  const reqHeaders = getRequestHeaders(event)
  for (const k of Object.keys(reqHeaders)) {
    const v = reqHeaders[k]
    if (v !== undefined && k.toLowerCase() !== 'host' && k.toLowerCase() !== 'content-length') {
      headers[k] = String(v)
    }
  }
  headers['host'] = new URL(mapped.port).host

  try {
    const body = ['GET', 'HEAD'].includes(method) ? undefined : await readRawBody(event)
    const resp = await fetch(target, {
      method,
      headers,
      body,
      redirect: 'follow'
    })
    setResponseStatus(event, resp.status)
    const respHeaders: Record<string, string> = {}
    resp.headers.forEach((v, k) => { respHeaders[k] = v })
    setResponseHeaders(event, respHeaders)
    return Buffer.from(await resp.arrayBuffer())
  } catch (e) {
    setResponseStatus(event, 500)
    return { code: 500, message: String(e), data: null }
  }
})