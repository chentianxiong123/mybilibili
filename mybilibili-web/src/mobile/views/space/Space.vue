<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../../components/Header.vue'

const router = useRouter()

const currentTab = ref('history')
</script>

<template>
  <div class="space-page">
    <Header />
    <div class="banner">
      <img src="/images/banner-top.png" alt="banner" />
    </div>
    <div class="space-tabs">
      <span
        :class="['space-tab', { active: currentTab === 'history' }]"
        @click="currentTab = 'history'; $router.replace('/m/space/history')"
      >
        历史记录
      </span>
      <span
        :class="['space-tab', { active: currentTab === 'submission' }]"
        @click="currentTab = 'submission'"
      >
        我的投稿
      </span>
    </div>
    <router-view v-if="currentTab === 'history'" />
    <div v-else class="placeholder">
      <p>小哔睡着了~</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.space-page { min-height: 100vh; background: $bg-color; }

.banner {
  width: 100%;
  img { width: 100%; display: block; }
}

.space-tabs {
  display: flex;
  background: $bg-white;
  border-bottom: 1px solid $border-color;
}

.space-tab {
  flex: 1;
  text-align: center;
  padding: 12px;
  font-size: 15px;
  color: $text-secondary;
  cursor: pointer;

  &.active {
    color: $theme-pink;
    font-weight: 600;
    border-bottom: 2px solid $theme-pink;
  }
}

.placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: $text-third;
  font-size: 14px;
}
</style>