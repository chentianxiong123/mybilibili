// Web 端（创作中心 / 投稿管理 / 草稿箱等）页面地址。
// WAP 与 Web 是两套应用：开发期 WAP(vite :5174) 与 Web(nuxt :3200) 端口不同，
// 跳转 Web 端页面时需要显式指向 Web dev server；生产/同域部署时直接使用当前域名。
const DEV_WEB_ORIGIN = 'http://localhost:3200'

export function getWebOrigin(): string {
  const origin = window.location.origin
  // 开发环境（vite 默认端口等）→ 指向 Web dev server；可用 VITE_WEB_ORIGIN 覆盖
  if (import.meta.env.DEV || origin.includes(':5174') || origin.includes(':5173')) {
    return (import.meta.env.VITE_WEB_ORIGIN as string | undefined) || DEV_WEB_ORIGIN
  }
  return origin
}