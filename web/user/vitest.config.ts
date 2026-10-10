import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'app'),
      '~': resolve(__dirname, 'app'),
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    // 按领域分组：vitest run --project <name> 只跑该组，
    // vitest run 全量跑所有组
    projects: [
      { test: { name: 'api', include: ['app/api/**/*.test.ts'] } },
      { test: { name: 'composables', include: ['app/composables/**/*.test.ts'] } },
      { test: { name: 'utils', include: ['app/utils/**/*.test.ts'] } },
      { test: { name: 'components', include: ['app/components/**/*.test.ts', 'app/components/**/*.test.vue'] } },
      { test: { name: 'stores', include: ['app/stores/**/*.test.ts'] } },
      { test: { name: 'teriteri', include: ['app/__tests__/**/*.test.ts'] } },
    ],
    coverage: {
      provider: 'v8',
      include: ['app/**/*.ts', 'app/**/*.vue'],
      exclude: ['app/**/*.test.*', 'app/**/*.d.ts'],
    },
  },
})
