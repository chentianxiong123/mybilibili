import 'vue'
import type { ElMessage } from 'element-plus'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import type { ApiResponse } from './index'

export {}

declare global {
  interface Window {
    __NUXT_DISABLE_ZOOM__?: boolean
  }
}

declare module 'vue' {
  export interface ComponentCustomProperties {
    // teriteri 全局方法注入（见 plugins/teriteri.client.ts）
    $message: typeof ElMessage
    $axios: AxiosInstance
    $get: <T = any>(url: string, config?: AxiosRequestConfig) => Promise<AxiosResponse<ApiResponse<T> & any>>
    $post: <T = any>(url: string, data?: any, headers?: any) => Promise<AxiosResponse<ApiResponse<T> & any>>
    // Vuex → Pinia 兼容层（Proxy）
    $store: VuexShim
  }
}

export interface VuexShim<State = Record<string, any>> {
  state: State & Record<string, any>
  commit: (type: string, payload?: any) => void
  dispatch: (type: string, payload?: any) => Promise<any>
}