import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/admin/',
  server: {
    port: 3002,
    proxy: {
      '/admin-api': {
        target: 'http://localhost:7070/admin',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/admin-api/, '')
      },
      '/uploads': {
        target: 'http://localhost:7070/admin',
        changeOrigin: true
      }
    }
  }
})