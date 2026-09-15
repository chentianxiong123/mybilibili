import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/global.scss'
import { registerSW } from 'virtual:pwa-register'
import { initWapTheme } from './utils/theme'
import { bootstrapSession } from './utils/session'

if ('serviceWorker' in navigator && (import.meta.env.PROD || import.meta.env.VITE_PWA_DEV)) {
  registerSW({ immediate: true })
}

initWapTheme()
// 启动时校验登录态：有效保留；过期自动续期；无效清理（不阻塞渲染，后台进行）
bootstrapSession()
const app = createApp(App)
app.use(router)
app.mount('#app')
