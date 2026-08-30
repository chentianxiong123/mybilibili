import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import 'element-plus/es/components/notification/style/css'
import 'element-plus/es/components/loading/style/css'
import mitt from 'mitt'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const emitter = mitt()

app.use(createPinia())
app.use(router)

app.config.globalProperties.emitter = emitter

app.mount('#app')
