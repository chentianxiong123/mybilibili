<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <div class="not-found">
    <div class="content">
      <div class="error-code">404</div>
      <p class="message">页面不见了</p>
      <div class="actions">
        <button class="btn-primary" @click="router.push('/')">返回首页</button>
        <button class="btn-secondary" @click="router.back()">返回上一页</button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.not-found {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 80vh;
  text-align: center;
}

.error-code {
  font-size: 120px;
  font-weight: 700;
  color: #e3e5e7;
  line-height: 1;
}

.message {
  font-size: 18px;
  color: #9499a0;
  margin: 16px 0 32px;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: center;

  button {
    padding: 10px 24px;
    border-radius: 8px;
    font-size: 14px;
    cursor: pointer;
    border: none;
  }
}

.btn-primary {
  background: #fb7299;
  color: #fff;

  &:hover { background: #f5729a; }
}

.btn-secondary {
  background: #e3e5e7;
  color: #61666d;

  &:hover { background: #d5d7db; }
}
</style>
