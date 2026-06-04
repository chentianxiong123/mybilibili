import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { fileURLToPath, URL } from 'node:url'

// 端口策略:已有 5173(mybilibili-web) / 5174(mybilibili-wap) / 3002(mybilibili-admin-web)
// Studio 排在 5180,避开默认 5173-5179 的范围
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({ resolvers: [ElementPlusResolver()] }),
    Components({ resolvers: [ElementPlusResolver({ importStyle: 'css' })] })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5180,
    strictPort: true,
    host: '127.0.0.1'
  }
})
