// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@element-plus/nuxt', '@pinia/nuxt'],
  css: ['@/assets/teriteri/css/base.css'],
  app: {
    head: {
      htmlAttrs: { lang: 'zh-CN' },
      title: '哔哩哔哩',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0' },
        { name: 'referrer', content: 'no-referrer' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/vite.svg' },
        { rel: 'stylesheet', href: 'https://at.alicdn.com/t/c/font_4179759_9hwhc7qk0zc.css' }
      ]
    }
  },
  devServer: {
    host: '0.0.0.0',
    port: 3200,
  },
  components: false,
  alias: {
    '~assets': '@/assets/teriteri'
  },
  routeRules: {
    '/': { ssr: false },
    '/message/**': { ssr: false },
    '/dynamic/**': { ssr: false },
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
