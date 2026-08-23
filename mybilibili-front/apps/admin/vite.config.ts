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
        manualChunks: {
          'vendor-vue': ['vue', 'vue-router', 'pinia'],
          'vendor-element': ['element-plus', '@element-plus/icons-vue'],
          'vendor-charts': ['echarts'],
          'vendor-utils': ['axios']
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3100,
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
      '/api/v1/ai': {
        target: 'http://localhost:8088',
        changeOrigin: true
      },
      '/api/v1/subtitle': {
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
})