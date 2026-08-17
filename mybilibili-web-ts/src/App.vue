<template>
  <component :is="layoutComponent">
    <router-view v-slot="{ Component: RouteComponent }">
      <keep-alive :include="['HomeView', 'CategoryView']">
        <component :is="RouteComponent" />
      </keep-alive>
    </router-view>
  </component>
</template>

<script setup lang="ts">
import { computed, ref, provide } from 'vue'
import { useRoute } from 'vue-router'
import LayoutHome from './layouts/LayoutHome.vue'
import LayoutSimple from './layouts/LayoutSimple.vue'
import LayoutNone from './layouts/LayoutNone.vue'

const route = useRoute()

const showLoginDialog = ref(false)
provide('showLoginDialog', showLoginDialog)

const layoutMap: Record<string, any> = {
  home: LayoutHome,
  simple: LayoutSimple,
  none: LayoutNone,
}

const layoutComponent = computed(() => {
  return layoutMap[route.meta.layout as string] || LayoutHome
})
</script>

<style>
:root {
  --bili-pink: #fb7299;
  --bili-pink-hover: #fc8bab;
  --bili-bg: #f4f5f7;
  --bili-text: #212121;
  --bili-text-secondary: #999;
  --bili-border: #e3e5e7;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  background-color: var(--bili-bg);
  color: var(--bili-text);
}

a {
  text-decoration: none;
  color: inherit;
}
</style>