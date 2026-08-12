/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'artplayer'
declare module 'artplayer-plugin-danmuku'
declare module 'flv.js'
declare module 'hls.js'
declare module 'mitt'
