<template>
  <Page class="page" actionBarHidden="true">
    <ScrollView>
      <StackLayout class="login-content">
        <Label text="登录" class="login-title" />
        <StackLayout class="input-group">
          <TextField v-model="username" hint="用户名" class="input-field" />
          <TextField v-model="password" hint="密码" secure="true" class="input-field" />
        </StackLayout>
        <Button text="登录" class="login-btn" @tap="handleLogin" />
        <Label text="还没有账号？注册" class="register-link" @tap="handleRegister" />
        <Label v-if="error" :text="error" class="error-text" />
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref } from 'nativescript-vue'
import { $closeModal, $navigateTo } from 'nativescript-vue'
import { login } from '../api/user'
import storage from '../utils/storage'

const username = ref('')
const password = ref('')
const error = ref('')

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  try {
    const res = await login(username.value, password.value)
    if (res.code === '1') {
      storage.setToken(res.data?.token || '')
      storage.setUser(res.data?.user || res.data)
      $closeModal()
    } else {
      error.value = res.message || '登录失败'
    }
  } catch (e: any) {
    error.value = e.message || '网络错误'
  }
}

function handleRegister() {
  $closeModal()
}
</script>

<style scoped>
.login-content {
  padding: 60 30;
  align-items: center;
}

.login-title {
  font-size: 28;
  font-weight: bold;
  color: #18191c;
  margin-bottom: 40;
}

.input-group {
  width: 100%;
  margin-bottom: 20;
}

.input-field {
  width: 100%;
  height: 44;
  padding: 0 12;
  font-size: 15;
  border-width: 1;
  border-color: #e3e5e7;
  border-radius: 8;
  margin-bottom: 12;
  background-color: #f8f9fa;
}

.login-btn {
  width: 100%;
  height: 44;
  background-color: #fb7299;
  color: white;
  font-size: 16;
  font-weight: bold;
  border-radius: 22;
}

.register-link {
  font-size: 13;
  color: #fb7299;
  margin-top: 20;
}

.error-text {
  font-size: 12;
  color: #f44336;
  margin-top: 12;
}
</style>