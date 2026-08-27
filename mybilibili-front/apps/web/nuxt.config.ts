// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@element-plus/nuxt', '@pinia/nuxt'],
  css: ['@mybilibili/ui',
    'element-plus/es/components/message/style/css',
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
  routeRules: {
    // /message 重定向到默认私信页，避免嵌套路由无 <NuxtPage /> 的 E4016 报错
    '/message': { redirect: '/message/private' },
    // === 私有/个性化页面关闭 SSR（避免 hydration 不一致；业界标准做法，见 docs/decisions） ===
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
    '/api/v1/search/**': { proxy: 'http://localhost:8084/api/v1/search/**' },
    '/api/v1/recommend/**': { proxy: 'http://localhost:8084/api/v1/recommend/**' },
    '/api/v1/creator/stats/**': { proxy: 'http://localhost:8084/api/v1/creator/stats/**' },
    '/api/v1/profile/**': { proxy: 'http://localhost:8084/api/v1/profile/**' },
    '/api/v1/message/**': { proxy: 'http://localhost:8086/api/v1/message/**' },
    '/api/v1/danmaku/**': { proxy: 'http://localhost:8086/api/v1/danmaku/**' },
    '/api/v1/creator/danmaku/**': { proxy: 'http://localhost:8086/api/v1/creator/danmaku/**' },
    '/api/v1/subtitle/**': { proxy: 'http://localhost:8088/api/v1/subtitle/**' },
    '/api/v1/ai/**': { proxy: 'http://localhost:8088/api/v1/ai/**' },
    '/api/v1/live/**': { proxy: 'http://localhost:8087/api/v1/live/**' },
    '/api/**': { proxy: 'http://localhost:8080/api/**' },
    '/uploads/**': { proxy: 'http://localhost:8080/uploads/**' },
    '/covers/**': { proxy: 'http://localhost:8080/covers/**' },
    '/videos/**': { proxy: 'http://localhost:8080/videos/**' },
    '/ws/**': { proxy: 'http://localhost:8086/ws/**' },
    '/sse/**': { proxy: 'http://localhost:8086/sse/**' }
  },
  vite: {
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
          silenceDeprecations: ['legacy-js-api', 'import']
        }
      }
    },
    server: {
      proxy: {
        '/api/v1/search': {
          target: 'http://localhost:8084',
          changeOrigin: true
        },
        '/api/v1/recommend': {
          target: 'http://localhost:8084',
          changeOrigin: true
        },
        '/api/v1/creator/stats': {
          target: 'http://localhost:8084',
          changeOrigin: true
        },
        '/api/v1/profile': {
          target: 'http://localhost:8084',
          changeOrigin: true
        },
        '/api/v1/message': {
          target: 'http://localhost:8086',
          changeOrigin: true
        },
        '/api/v1/danmaku': {
          target: 'http://localhost:8086',
          changeOrigin: true
        },
        '/api/v1/creator/danmaku': {
          target: 'http://localhost:8086',
          changeOrigin: true
        },
        '/api/v1/subtitle': {
          target: 'http://localhost:8088',
          changeOrigin: true
        },
        '/api/v1/ai': {
          target: 'http://localhost:8088',
          changeOrigin: true
        },
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        },
        '/uploads': {
          target: 'http://localhost:8080',
          changeOrigin: true
        },
        '/covers': {
          target: 'http://localhost:8080',
          changeOrigin: true
        },
        '/videos': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      }
    }
  }
})