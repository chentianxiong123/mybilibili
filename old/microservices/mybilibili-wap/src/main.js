import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/global.scss'
import { initWapTheme } from './utils/theme'

initWapTheme()
const app = createApp(App)
app.use(router)
app.mount('#app')
