<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const visible = ref(false)
let scrollHandler = null

onMounted(() => {
  scrollHandler = () => {
    visible.value = window.scrollY > 500
  }
  window.addEventListener('scroll', scrollHandler)
})

onUnmounted(() => {
  if (scrollHandler) window.removeEventListener('scroll', scrollHandler)
})

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>

<template>
  <transition name="fade">
    <div v-if="visible" class="to-top" @click="scrollToTop">
      <i class="icon-arrow-up" />
    </div>
  </transition>
</template>

<style scoped lang="scss">
.to-top {
  position: fixed;
  bottom: 60px;
  right: 16px;
  width: 40px;
  height: 40px;
  background: rgba(0,0,0,0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  z-index: 100;
  cursor: pointer;
}

.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>