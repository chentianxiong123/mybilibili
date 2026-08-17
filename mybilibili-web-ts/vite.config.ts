import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { VitePWA } from 'vite-plugin-pwa'
import viteCompression from 'vite-plugin-compression'


export default defineConfig(({ mode }) => {
  const isWeb = mode !== 'admin'

  return {
    plugins: [
      vue(),
      vueDevTools(),
      AutoImport({
        resolvers: [ElementPlusResolver()]
      }),
      Components({
        resolvers: [ElementPlusResolver({ importStyle: 'css' })]
      }),
      VitePWA({
        registerType: 'autoUpdate',
        includeAssets: ['vite.svg', 'pwa-192x192.png', 'pwa-512x512.png'],
        manifest: {
          name: '哔哩哔哩',
          short_name: '哔哩',
          description: '分享生活，遇见同好',
          theme_color: '#fb7299',
          background_color: '#f4f5f7',
          display: 'standalone',
          start_url: '/',
          icons: [
            {
              src: '/pwa-192x192.png',
              sizes: '192x192',
              type: 'image/png'
            },
            {
              src: '/pwa-512x512.png',
              sizes: '512x512',
              type: 'image/png'
            }
          ]
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,svg,png,jpg,ico,woff2}'],
          globIgnores: ['admin/**', '**/assets/videos/**']
        },
        devOptions: {
          enabled: false
        }
      }),
      viteCompression({
        algorithm: 'brotliCompress',
        ext: '.br',
        threshold: 1024,
        deleteOriginalAssets: false
      })
      ],
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
          silenceDeprecations: ['legacy-js-api', 'import']
        }
      }
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    test: {
      environment: 'jsdom'
    },
    build: {
      outDir: isWeb ? 'dist/web' : 'dist/admin',
      cssMinify: 'lightningcss',
      rollupOptions: {
        input: isWeb ? 'index.html' : 'index.html',
        output: {
          manualChunks: {
            'vendor-vue': ['vue', 'vue-router', 'pinia'],
            'vendor-charts': ['echarts'],
            'vendor-utils': ['axios']
          }
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
