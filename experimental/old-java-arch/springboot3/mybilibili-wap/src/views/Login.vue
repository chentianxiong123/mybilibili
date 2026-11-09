<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const login = async () => {
  if (!username.value.trim() || !password.value.trim()) {
    errorMsg.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await fetch('/api/user/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value })
    })
    const data = await res.json()
    if (data.code === 200 && data.data) {
      // 兼容后端返回的对象包含 user 还是扁平结构
      const userObj = data.data.user || data.data
      const tokenStr = data.data.token || data.data
      localStorage.setItem('token', tokenStr)
      localStorage.setItem('user', JSON.stringify(userObj))
      router.push('/m/index')
    } else {
      errorMsg.value = data.message || '登录失败'
    }
  } catch (e) {
    errorMsg.value = '网络错误'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-header">
      <div class="logo">
        <img src="../assets/logo.png" alt="bilibili" />
      </div>
      <h2 class="title">登录</h2>
    </div>
    <div class="login-form">
      <div class="form-item">
        <input v-model="username" placeholder="用户名" maxlength="20" @keydown.enter="login" />
      </div>
      <div class="form-item">
        <input v-model="password" type="password" placeholder="密码" maxlength="32" @keydown.enter="login" />
      </div>
      <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>
      <button class="login-btn" :disabled="loading" @click="login">
        {{ loading ? '登录中...' : '登录' }}
      </button>
      <div class="login-links">
        <router-link to="/m/index">先逛逛</router-link>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.login-page {
  min-height: 100vh;
  background: $bg-color;
  padding: 0 24px;
}

.login-header {
  text-align: center;
  padding: 60px 0 40px;
}

.logo img {
  width: 80px;
  height: auto;
  margin-bottom: 16px;
}

.title {
  font-size: 20px;
  font-weight: 600;
  color: $text-primary;
}

.login-form {
  background: $bg-white;
  border-radius: 12px;
  padding: 24px;
}

.form-item {
  margin-bottom: 16px;

  input {
    width: 100%;
    height: 44px;
    border: 1px solid $border-color;
    border-radius: 8px;
    padding: 0 14px;
    font-size: 15px;
    outline: none;
    transition: border-color 0.2s;

    &:focus {
      border-color: $theme-pink;
    }
  }
}

.error-msg {
  color: #ff4d4f;
  font-size: 13px;
  margin-bottom: 12px;
}

.login-btn {
  width: 100%;
  height: 44px;
  background: $theme-pink;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  margin-bottom: 16px;

  &:disabled {
    opacity: 0.6;
  }
}

.login-links {
  text-align: center;
  font-size: 13px;

  a {
    color: $text-secondary;
  }
}
</style>