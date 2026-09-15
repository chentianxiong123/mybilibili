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
  devServer: {
    host: '0.0.0.0',  // 让 traefik 通过 172.18.0.1 网关能访问到 (裸跑进程)
    port: 3200,
  },
  routeRules: {
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
