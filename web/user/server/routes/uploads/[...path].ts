import { defineEventHandler, getRequestURL, setResponseStatus, setResponseHeaders } from 'h3'

export default defineEventHandler(async (event) => {
  const fullUrl = getRequestURL(event)
  const path = fullUrl.pathname
  const upstream = 'http://127.0.0.1:8080'
  const target = upstream + path + (fullUrl.search || '')

  try {
    const resp = await fetch(target, {
      method: 'GET'
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