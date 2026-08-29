// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@element-plus/nuxt', '@pinia/nuxt'],
  css: ['@mybilibili/ui',
    'element-plus/es/components/message-box/style/css',
    'element-plus/es/components/notification/style/css',
    'element-plus/es/components/loading/style/css'],
  app: {
    head: {
      htmlAttrs: { lang: 'zh-CN' },
      title: '哔哩哔哩',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0' },
        { name: 'referrer', content: 'no-referrer' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/vite.svg' }
      ]
    }
  },
  // 开发期反代：dev server 把 /api/v1/* 转到具体后端服务端口
  // 与生产 IngressRoute 路由规则一一对应（traefik labels）
  routeRules: {
    '/api/v1/search/**':  { proxy: 'http://localhost:8084/api/v1/search/**' },
    '/api/v1/recommend/**': { proxy: 'http://localhost:8084/api/v1/recommend/**' },
    '/api/v1/ai/**':      { proxy: 'http://localhost:8088/api/v1/ai/**' },
    '/api/v1/subtitle/**': { proxy: 'http://localhost:8088/api/v1/subtitle/**' },
    '/api/v1/danmaku/**': { proxy: 'http://localhost:8086/api/v1/danmaku/**' },
    '/api/v1/message/**': { proxy: 'http://localhost:8086/api/v1/message/**' },
    '/api/v1/live/**':    { proxy: 'http://localhost:8087/api/v1/live/**' },
    '/api/v1/bili/**':    { proxy: 'http://localhost:8091/api/v1/bili/**' },
    '/api/v1/studio/**':  { proxy: 'http://localhost:8089/api/v1/studio/**' },
    '/api/v1/**':         { proxy: 'http://localhost:8080/api/v1/**' },
    '/uploads/**':        { proxy: 'http://localhost:8080/uploads/**' },
    '/covers/**':         { proxy: 'http://localhost:8080/covers/**' },
    '/videos/**':         { proxy: 'http://localhost:8080/videos/**' },
    '/message': { redirect: '/message/private' },
    '/dynamic/**': { ssr: false },
    '/profile/**': { ssr: false },
    '/personal-center/**': { ssr: false },
    '/message/**': { ssr: false },
    '/history': { ssr: false },
    '/avatar': { ssr: false },
    '/live/**': { ssr: false },
    '/create-center/**': { ssr: false },
    '/manuscript/**': { ssr: false },
    '/login': { ssr: false },
    '/collections': { ssr: false },
    '/collection/**': { ssr: false },
  },
  vite: {
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
          silenceDeprecations: ['legacy-js-api', 'import']
        }
      }
    }
  }
})
