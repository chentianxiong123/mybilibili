<template>
    <div />
</template>

<script setup lang="ts">
// 兜底：未知 /message/<type> 收敛到规范路由
// 旧内部 key 别名：private -> whisper，like -> love，settings -> config
const route = useRoute()
const alias: Record<string, string> = {
  private: '/message/whisper',
  like: '/message/love',
  settings: '/message/config',
}

if (import.meta.client) {
  const t = String(route.params.type || '')
  navigateTo(alias[t] || '/message/whisper', { replace: true })
}
</script>
