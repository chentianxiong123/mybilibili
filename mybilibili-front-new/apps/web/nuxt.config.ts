// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  ssr: false,
  modules: ['@element-plus/nuxt', '@pinia/nuxt'],
  css: ['@/assets/teriteri/css/base.css', '@/assets/teriteri/css/artplayer-overrides.css', '@mybilibili/ui',
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
    '/message': { redirect: '/message/whisper' },
    '/message/**': { ssr: false },
    '/dynamic/**': { ssr: false },
    '/profile/**': { ssr: false },
    '/personal-center/**': { ssr: false },
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
        } as any
      }
    }
  }
})
