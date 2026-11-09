import { defineNuxtPlugin } from '#app'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

export default defineNuxtPlugin({
  name: 'pinia-plugin-persistedstate',
  setup({ $pinia }) {
    if (import.meta.client) {
      $pinia.use(piniaPluginPersistedstate)
    }
  }
})