<template>
  <GridLayout rows="*, auto" class="page">
    <Frame row="0" ref="mainFrame" id="mainFrame">
      <HomePage />
    </Frame>
    <BottomTabBar row="1" :activeTab="activeTab" @select="onTabSelect" />
  </GridLayout>
</template>

<script lang="ts" setup>
import { ref } from 'nativescript-vue'
import { $navigateTo, $showModal } from 'nativescript-vue'
import { Frame } from '@nativescript/core'
import BottomTabBar from './components/BottomTabBar.vue'
import HomePage from './views/home/Index.vue'
import DynamicPage from './views/dynamic/Index.vue'
import SpacePage from './views/space/Space.vue'
import MallPage from './views/mall/Index.vue'
import LoginPage from './views/Login.vue'

const activeTab = ref(0)
const mainFrame = ref<Frame | null>(null)

const tabPages = [HomePage, DynamicPage, null, MallPage, SpacePage]

function onTabSelect(index: number) {
  if (index === 2) {
    const token = require('./utils/storage').default.getToken()
    if (!token) {
      $showModal(LoginPage, { fullscreen: true })
      return
    }
    return
  }
  if (index === 3) {
    return
  }
  activeTab.value = index
  const page = tabPages[index]
  if (page) {
    $navigateTo(page, {
      clearHistory: true,
      frame: 'mainFrame'
    })
  }
}
</script>