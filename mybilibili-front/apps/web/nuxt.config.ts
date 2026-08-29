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
  // 开发期静态资源：nuxt dev server 把 /uploads /covers /videos 反代到 traefik (80)
  // 与生产 IngressRoute 同源效果一致
  routeRules: {
    '/uploads/**': { proxy: 'http://localhost:80/uploads/**' },
    '/covers/**':  { proxy: 'http://localhost:80/covers/**' },
    '/videos/**':  { proxy: 'http://localhost:80/videos/**' },
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
