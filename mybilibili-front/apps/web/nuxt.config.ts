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
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/vite.svg' }
      ]
    }
  },
  routeRules: {
    '/live/**': { ssr: false },
    '/create-center/**': { ssr: false },
    '/api/v1/search/**': { proxy: 'http://localhost:8084/api/v1/search/**' },
    '/api/v1/recommend/**': { proxy: 'http://localhost:8084/api/v1/recommend/**' },
    '/api/v1/creator/stats/**': { proxy: 'http://localhost:8084/api/v1/creator/stats/**' },
    '/api/v1/profile/**': { proxy: 'http://localhost:8084/api/v1/profile/**' },
    '/api/**': { proxy: 'http://localhost:8080/api/**' },
    '/uploads/**': { proxy: 'http://localhost:8080/uploads/**' },
    '/covers/**': { proxy: 'http://localhost:8080/covers/**' },
    '/videos/**': { proxy: 'http://localhost:8080/videos/**' },
    '/ws/**': { proxy: 'http://localhost:8080/ws/**' }
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
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        },
        '/ws/notification': {
          target: 'ws://localhost:8080',
          ws: true,
          changeOrigin: true
        },
        '/ws/danmaku': {
          target: 'ws://localhost:8080',
          ws: true,
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