import { defineEventHandler, getRequestURL, getMethod, getRequestHeaders, readRawBody, setResponseStatus, setResponseHeaders } from 'h3'

// teriteri 前端 URL → Go 后端真实路径映射（来自 mybilibili-refactor 验证过的 PATH_MAP）
const PATH_MAP: Record<string, string> = {
  '/video/random/visitor': '/manuscript/recommended',
  '/video/cumulative/visitor': '/manuscript/hot',
  '/video/getone': '/manuscript/',
  '/video/user-works': '/manuscript/userManuscripts',
  '/video/user-love': '/manuscript/myLikes',
  '/video/user-collect': '/manuscript/userCollections',
  '/category/getall': '/category',
  '/search/hot/get': '/search/hot',
  '/search/word/add': '/search/hot/increment',
  '/search/word/get': '/search/suggest',
  '/search/video/only-pass': '/search/videos',
  '/search/user': '/search/users',
  '/video/love-or-not': '/interaction/video/love-or-not',
  '/video/collected-fids': '/favorites/check',
  '/comment/get': '/comment/list',
  '/comment/add': '/comment/add',
  '/comment/love-or-not': '/comment/',
  '/danmu-list': '/danmaku/video',
  '/favorite/get-all/user': '/favorites',
  '/favorite/get-all/visitor': '/favorites',
  '/user/info/get-one': '/user/',
  '/user/info': '/user/',
  '/user/login': '/user/token',
  '/user/register': '/user/register',
}

// 按服务端口分发的路由
function pickUpstream(realPath: string): string {
  if (realPath.startsWith('/search/')) return 'http://127.0.0.1:8084'
  if (realPath.startsWith('/danmaku/') || realPath.startsWith('/danmu')) return 'http://127.0.0.1:8086'
  return 'http://127.0.0.1:8080'
}

function adaptUrl(url: string, query: URLSearchParams): { target: string; port: string } {
  // /video/getone 特殊：vid 从 query 或路径
  if (url.startsWith('/video/getone')) {
    const vid = query.get('vid') || url.split('/').pop()
    return { target: `/manuscript/${vid}`, port: 'http://127.0.0.1:8080' }
  }
  const keys = Object.keys(PATH_MAP).sort((a, b) => b.length - a.length)
  for (const from of keys) {
    if (url === from || url.startsWith(from + '/') || url.startsWith(from + '?')) {
      const real = url.replace(from, PATH_MAP[from])
      return { target: real, port: pickUpstream(real) }
    }
  }
  return { target: url, port: pickUpstream(url) }
}

export default defineEventHandler(async (event) => {
  const fullUrl = getRequestURL(event)
  const path = fullUrl.pathname

  if (!path.startsWith('/api/')) {
    setResponseStatus(event, 404)
    return { error: true, statusCode: 404, statusMessage: 'Not an API prefix', path }
  }

  const teriteriUrl = path.slice('/api'.length) || '/'
  const query = new URLSearchParams(fullUrl.search || '')
  const method = getMethod(event)

  let mapped: { target: string; port: string }
  try {
    mapped = adaptUrl(teriteriUrl, query)
  } catch (e) {
    setResponseStatus(event, 500)
    return { code: 500, message: `adaptUrl failed: ${e}`, data: null }
  }

  const target = mapped.port + '/api/v1' + mapped.target + (fullUrl.search || '')

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
    const buf = await resp.arrayBuffer()
    return Buffer.from(buf)
  } catch (e) {
    setResponseStatus(event, 500)
    return { code: 500, message: String(e), data: null }
  }
})