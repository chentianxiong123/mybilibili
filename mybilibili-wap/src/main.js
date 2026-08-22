import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/global.scss'
import { initWapTheme } from './utils/theme'
import { bootstrapSession } from './utils/session'

initWapTheme()
// 启动时校验登录态：有效保留；过期自动续期；无效清理（不阻塞渲染，后台进行）
bootstrapSession()
const app = createApp(App)
app.use(router)
app.mount('#app')
