// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@element-plus/nuxt', '@pinia/nuxt'],
  css: ['element-plus/es/components/message/style/css',
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
    '/admin/**': { ssr: false },
    '/live/**': { ssr: false },
    '/create-center/**': { ssr: false }
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