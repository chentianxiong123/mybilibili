<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { userApi, captchaApi, emailCodeApi } from '../api/index.js'
import { Close, VideoPlay } from '@element-plus/icons-vue'

const router = useRouter()

// 忘记密码表单
const forgotForm = reactive({
  email: '',
  captchaAnswer: '',
  emailCode: '',
  newPassword: '',
  confirmPassword: ''
})

// 验证码相关
const emailCaptchaId = ref('')
const emailCaptchaQuestion = ref('')
const captchaLoading = ref(false)
const sendEmailLoading = ref(false)
const emailSent = ref(false)
const emailCountdown = ref(0)

// 加载图形验证码
const loadEmailCaptcha = () => {
  captchaLoading.value = true
  captchaApi.newCaptcha().then(res => {
    if (res.code === 200) {
      emailCaptchaId.value = res.data.captchaId
      emailCaptchaQuestion.value = res.data.question
      forgotForm.captchaAnswer = ''
    }
  }).finally(() => { captchaLoading.value = false })
}
loadEmailCaptcha()

// 发送邮箱验证码
const handleSendEmailCode = () => {
  if (!forgotForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(forgotForm.email)) {
    ElMessage.error('请先输入有效的邮箱地址')
    return
  }
  if (!forgotForm.captchaAnswer.trim()) {
    ElMessage.error('请先输入图形验证码')
    return
  }
  captchaApi.verifyCaptcha(emailCaptchaId.value, forgotForm.captchaAnswer).then(res => {
    if (res.code !== 200 || !res.data) {
      ElMessage.error('图形验证码错误')
      loadEmailCaptcha()
      return
    }
    sendEmailLoading.value = true
    emailCodeApi.sendCode(forgotForm.email).then(res => {
      if (res.code === 200) {
        emailSent.value = true
        emailCountdown.value = 60
        ElMessage.success('验证码已发送到邮箱')
        const timer = setInterval(() => {
          emailCountdown.value--
          if (emailCountdown.value <= 0) clearInterval(timer)
        }, 1000)
      } else {
        ElMessage.error(res.message || '发送失败')
        loadEmailCaptcha()
      }
    }).finally(() => { sendEmailLoading.value = false })
  })
}

// 表单验证规则
const forgotRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  emailCode: [
    { required: true, message: '请输入邮箱验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: (rule, value, callback) => {
        if (value !== forgotForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      }, trigger: 'blur' }
  ]
}

// 表单引用
const forgotFormRef = ref()
const loading = ref(false)

// 重置密码处理
const handleForgotPassword = () => {
  forgotFormRef.value.validate((valid) => {
    if (!valid) return

    loading.value = true
    userApi.forgotPassword(forgotForm.email, forgotForm.emailCode, forgotForm.newPassword)
      .then(response => {
        if (response.code === 200) {
          ElMessage.success('密码重置成功，请登录')
          router.push('/login')
        } else {
          ElMessage.error(response.message || '重置失败')
        }
      })
      .catch(error => {
        ElMessage.error('重置失败，请检查验证码和邮箱')
        console.error('重置密码错误:', error)
      })
      .finally(() => {
        loading.value = false
      })
  })
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/login')
}

// 跳转到注册页
const goToRegister = () => {
  router.push({ path: '/login', query: { mode: 'register' } })
}

// 返回上一页
const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="forgot-container">
    <div class="forgot-form-wrapper">
      <div class="forgot-form">
        <!-- 关闭按钮 -->
        <button class="close-btn" @click="goBack">
          <el-icon><Close /></el-icon>
        </button>

        <div class="forgot-logo">
          <el-icon><VideoPlay /></el-icon>
          <span>哔哩哔哩</span>
        </div>
        <h2 class="forgot-title">找回密码</h2>

        <el-form
          ref="forgotFormRef"
          :model="forgotForm"
          :rules="forgotRules"
          label-width="0px"
          class="form"
        >
          <el-form-item prop="email">
            <el-input
              v-model="forgotForm.email"
              type="email"
              placeholder="请输入注册邮箱"
              prefix-icon="el-icon-message"
              clearable
            ></el-input>
          </el-form-item>

          <el-form-item>
            <div class="email-captcha-row">
              <el-input
                v-model="forgotForm.captchaAnswer"
                placeholder="图形验证码"
                style="flex:1"
              ></el-input>
              <span class="captcha-question" :class="{ loading: captchaLoading }">{{ emailCaptchaQuestion }}</span>
              <el-button link type="primary" @click="loadEmailCaptcha" :loading="captchaLoading">刷新</el-button>
            </div>
          </el-form-item>

          <el-form-item prop="emailCode">
            <div class="email-code-row">
              <el-input
                v-model="forgotForm.emailCode"
                placeholder="邮箱验证码"
                prefix-icon="el-icon-message"
              ></el-input>
              <el-button
                type="primary"
                :disabled="emailSent && emailCountdown > 0"
                :loading="sendEmailLoading"
                @click="handleSendEmailCode"
              >
                {{ emailCountdown > 0 ? emailCountdown + 's' : '发送验证码' }}
              </el-button>
            </div>
          </el-form-item>

          <el-form-item prop="newPassword">
            <el-input
              v-model="forgotForm.newPassword"
              type="password"
              placeholder="请输入新密码"
              prefix-icon="el-icon-lock"
              show-password
            ></el-input>
          </el-form-item>

          <el-form-item prop="confirmPassword">
            <el-input
              v-model="forgotForm.confirmPassword"
              type="password"
              placeholder="请确认新密码"
              prefix-icon="el-icon-lock"
              show-password
            ></el-input>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleForgotPassword" class="forgot-btn" block :loading="loading">
              重置密码
            </el-button>
          </el-form-item>
        </el-form>

        <div class="back-link">
          <el-button type="text" @click="goToLogin">返回登录</el-button>
          <span>|</span>
          <el-button type="text" @click="goToRegister">注册账号</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.forgot-container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.forgot-form-wrapper {
  width: 100%;
  max-width: 450px;
}

.forgot-form {
  position: relative;
  background-color: #fff;
  border-radius: 12px;
  padding: 30px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.close-btn {
  position: absolute;
  top: 15px;
  right: 15px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: transparent;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f0f0f0;
}

.close-btn .el-icon {
  font-size: 18px;
  color: #999;
}

.forgot-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
  color: #00a1d6;
  margin-bottom: 20px;
}

.forgot-logo el-icon {
  font-size: 32px;
  margin-right: 10px;
}

.forgot-title {
  font-size: 20px;
  font-weight: 500;
  color: #333;
  margin: 0 0 20px 0;
  text-align: center;
}

.form {
  width: 100%;
}

.form .el-form-item {
  margin-bottom: 20px;
}

.email-captcha-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.captcha-question {
  font-size: 14px;
  color: #333;
  font-weight: 500;
  white-space: nowrap;
  min-width: 80px;
}

.captcha-question.loading {
  color: #999;
}

.email-code-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.email-code-row .el-input {
  flex: 1;
}

.forgot-btn {
  height: 40px;
  border-radius: 20px;
  font-size: 16px;
}

.back-link {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #999;
}

.back-link span {
  color: #ddd;
}

.back-link el-button {
  color: #00a1d6;
  padding: 0;
}
</style>
