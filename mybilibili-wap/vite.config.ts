import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'path'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      strategies: 'generateSW',
      
      manifest: {
        name: '哔哩哔哩移动端',
        short_name: 'B站WAP',
        description: '哔哩哔哩移动端离线版',
        lang: 'zh-CN',
        theme_color: '#fb7299',
        background_color: '#ffffff',
        display: 'standalone',
        scope: '/wap/',
        start_url: '/wap/',
        icons: [
          { src: 'wap-icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any maskable' },
          { src: 'wap-icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any maskable' },
          { src: 'wap-icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,png,jpg,jpeg,gif,webp,woff,woff2}'],
        navigateFallback: '/wap/index.html',
        navigateFallbackDenylist: [/^\/api\//, /\/healthz$/],
        runtimeCaching: [
          {
            urlPattern: /\.(?:png|jpg|jpeg|gif|webp|svg|avif)(?:\?.*)?$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'wap-images',
              expiration: { maxEntries: 200, maxAgeSeconds: 60 * 60 * 24 * 30 },
              cacheableResponse: { statuses: [0, 200] }
            }
          },
          {
            urlPattern: /^\/api\/v1\//,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'wap-api',
              networkTimeoutSeconds: 5,
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 },
              cacheableResponse: { statuses: [0, 200] }
            }
          }
        ]
      },
      devOptions: { enabled: false }
    }),
    {
    name: 'healthz',
    configureServer(server) {
      server.middlewares.use('/healthz', (_req, res) => {
        res.setHeader('Content-Type', 'application/json')
        res.end(JSON.stringify({ status: 'ok', service: 'wap', ts: Date.now() }))
      })
    }
  }],
  base: '/wap/',
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
      '@': path.resolve(__dirname, 'src')
    }
  },
  build: {
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-vue': ['vue', 'vue-router'],
          'vendor-utils': ['axios'],
          'vendor-player': ['artplayer', 'artplayer-plugin-danmuku'],
          'vendor-hls': ['hls.js']
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5174,
    proxy: {
      '/api/v1/search/history': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/api/v1/search/': {
        target: 'http://localhost:8084',
        changeOrigin: true
      },
      '/api/v1/recommend/': {
        target: 'http://localhost:8084',
        changeOrigin: true
      },
      '/api/v1/live/': {
        target: 'http://localhost:8087',
        changeOrigin: true
      },
      '/api/v1/danmaku/': {
        target: 'http://localhost:8086',
        changeOrigin: true
      },
      '/api/v1/message/': {
        target: 'http://localhost:8086',
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
  },
  esbuild: {
    target: 'esnext'
  },
  optimizeDeps: {
    esbuildOptions: {
      target: 'esnext'
    }
  }
})
