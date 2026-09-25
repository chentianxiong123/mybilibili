import { ElMessage } from 'element-plus'
import axios from 'axios'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { useTeriteriStore } from '@/stores/teriteri'
import { get, post, del } from '@/teriteri-src/network/request'

export default defineNuxtPlugin((nuxtApp) => {
  const router = nuxtApp.$router
  const route = useRoute()

  nuxtApp.vueApp.config.globalProperties.$message = ElMessage
  nuxtApp.vueApp.config.globalProperties.$axios = axios
  nuxtApp.vueApp.config.globalProperties.$get = get
  nuxtApp.vueApp.config.globalProperties.$post = post
  nuxtApp.vueApp.config.globalProperties.$delete = del

  // 注册全部 element-plus 图标（teriteri main.js 原样）
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    nuxtApp.vueApp.component(key, component)
  }

  // Vuex → Pinia 兼容层：teriteri 组件全用 Vuex 语法 this.$store.state.x / .commit / .dispatch
  function createVuexShim() {
    const store = useTeriteriStore()
    return {
      state: new Proxy({}, {
        get(_, key) { return store[key] },
        set(_, key, value) { store[key] = value; return true }
      }),
      commit(type, payload) {
        const fn = store[type]
        if (typeof fn === 'function') fn.call(store, payload)
        else store[type] = payload
      },
      dispatch(type, payload) {
        const fn = store[type]
        if (typeof fn === 'function') return fn.call(store, payload)
        return Promise.resolve()
      }
    }
  }

  nuxtApp.vueApp.mixin({
    computed: {
      $store() {
        return createVuexShim()
      },
      $router() {
        return router
      },
      $route() {
        return route
      }
    }
  })
})