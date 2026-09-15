<template>
  <div class="not-found">
    <div class="content">
      <div class="error-code">404</div>
      <p class="message">页面不存在</p>
      <button class="back-btn" @click="goHome">返回首页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
const goHome = () => router.push('/m/index')
</script>

<style scoped>
.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f6f7f8;
  text-align: center;
  padding: 0 24px;
}
.content {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.error-code {
  font-size: 96px;
  font-weight: 700;
  color: #e3e5e7;
  line-height: 1;
  letter-spacing: 4px;
}
.message {
  font-size: 16px;
  color: #9499a0;
  margin: 16px 0 32px;
}
.back-btn {
  padding: 10px 24px;
  border: none;
  border-radius: 20px;
  background: #00a1d6;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}
.back-btn:active {
  opacity: 0.8;
}
</style>
