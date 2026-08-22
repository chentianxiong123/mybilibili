import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
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
