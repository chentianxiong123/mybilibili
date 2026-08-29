import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue(), {
    name: 'healthz',
    configureServer(server) {
      server.middlewares.use('/healthz', (_req, res) => {
        res.setHeader('Content-Type', 'application/json')
        res.end(JSON.stringify({ status: 'ok', service: 'admin', ts: Date.now() }))
      })
    }
  }],
  base: '/admin/',
  resolve: {
    alias: {
      '~': path.resolve(__dirname, 'app'),
      '@': path.resolve(__dirname, 'app')
    }
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern-compiler',
        silenceDeprecations: ['legacy-js-api', 'import']
      }
    }
  },
  build: {
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/vue') || id.includes('node_modules/pinia') || id.includes('node_modules/vue-router')) return 'vendor-vue'
          if (id.includes('node_modules/element-plus')) return 'vendor-element'
          if (id.includes('node_modules/echarts')) return 'vendor-charts'
          if (id.includes('node_modules/axios')) return 'vendor-utils'
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3100,
    // dev server 反代到具体后端端口（与生产 IngressRoute 一一对应）
    proxy: {
      '/api/v1/search':   { target: 'http://localhost:8084', changeOrigin: true },
      '/api/v1/recommend': { target: 'http://localhost:8084', changeOrigin: true },
      '/api/v1/ai':       { target: 'http://localhost:8088', changeOrigin: true },
      '/api/v1/subtitle':  { target: 'http://localhost:8088', changeOrigin: true },
      '/api/v1/danmaku':  { target: 'http://localhost:8086', changeOrigin: true },
      '/api/v1/message':  { target: 'http://localhost:8086', changeOrigin: true },
      '/api/v1/live':     { target: 'http://localhost:8087', changeOrigin: true },
      '/api/v1/bili':     { target: 'http://localhost:8091', changeOrigin: true },
      '/api/v1/studio':   { target: 'http://localhost:8089', changeOrigin: true },
      '/api/v1':          { target: 'http://localhost:8080', changeOrigin: true },
      '/uploads':         { target: 'http://localhost:8080', changeOrigin: true },
      '/covers':          { target: 'http://localhost:8080', changeOrigin: true },
      '/videos':          { target: 'http://localhost:8080', changeOrigin: true },
    }
  }
})