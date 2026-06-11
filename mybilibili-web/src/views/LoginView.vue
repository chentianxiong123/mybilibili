<script setup>
import { ref, reactive, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { userApi, captchaApi, emailCodeApi } from '../api/index.js'
import api from '../api/index.js'
import { setAuthSession } from '../utils/auth.js'
import { Close, VideoPlay } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const goAfterLogin = () => {
  router.push(route.query.redirect || '/')
}

const authMode = ref(route.query.mode === 'register' ? 'register' : 'login')

watch(
  () => route.query.mode,
  (mode) => {
    authMode.value = mode === 'register' ? 'register' : 'login'
    if (authMode.value === 'register' && !registerCaptchaId.value) {
      loadRegisterCaptcha()
    }
  }
)

// 登录方式：password / email_code
const loginMode = ref('password')

// 密码登录表单
const loginForm = reactive({
  username: '',
  password: '',
  captchaAnswer: '',
  captchaId: '',
  captchaQuestion: ''
})

// 邮箱验证码登录表单
const emailLoginForm = reactive({
  email: '',
  emailCode: '',
  captchaAnswer: '',
  captchaId: '',
  captchaQuestion: ''
})

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  emailCode: '',
  agreeTerms: false
})

// 表单验证规则
const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  captchaAnswer: [
    { required: true, message: '请输入验证码', trigger: 'blur' }
  ]
}

const emailLoginRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  emailCode: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ],
  captchaAnswer: [
    { required: true, message: '请输入验证码', trigger: 'blur' }
  ]
}

const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  emailCode: [
    { required: true, message: '请输入邮箱验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' },
    { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d_]{6,20}$/, message: '密码必须包含大小写字母和数字', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: (rule, value, callback) => {
        if (value !== registerForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      }, trigger: 'blur' }
  ],
  agreeTerms: [
    { validator: (rule, value, callback) => {
        if (!value) callback(new Error('请阅读并同意用户协议和隐私政策'))
        else callback()
      }, trigger: 'change' }
  ]
}

// 表单引用
const loginFormRef = ref()
const emailLoginFormRef = ref()
const registerFormRef = ref()
const loading = ref(false)
const captchaLoading = ref(false)
const sendEmailLoading = ref(false)
const emailSent = ref(false)
const emailCountdown = ref(0)
const registerCaptchaId = ref('')
const registerCaptchaQuestion = ref('')
const registerCaptchaAnswer = ref('')
const registerCaptchaLoading = ref(false)
const sendRegisterEmailLoading = ref(false)
const registerEmailSent = ref(false)
const registerEmailCountdown = ref(0)
const showInterestDialog = ref(false)
const registeredUser = ref(null)
const selectedTags = ref([])
const allInterestTags = [
  '游戏', '科技', '音乐', '舞蹈', '动画', '番剧', '影视', '娱乐',
  '生活', '美食', '知识', '教育', '体育', '汽车', '时尚', '搞笑',
  '动物', '自然', '历史', '军事', '财经', '新闻', '鬼畜', '手工',
  '绘画', '摄影', '编程', 'AI', '数码', '家电', '旅行', '健身'
]

// 加载新验证码
const loadCaptcha = (form) => {
  captchaLoading.value = true
  captchaApi.newCaptcha().then(res => {
    if (res.code === 200) {
      if (form === 'email') {
        emailLoginForm.captchaId = res.data.captchaId
        emailLoginForm.captchaQuestion = res.data.question
        emailLoginForm.captchaAnswer = ''
      } else {
        loginForm.captchaId = res.data.captchaId
        loginForm.captchaQuestion = res.data.question
        loginForm.captchaAnswer = ''
      }
    }
  }).finally(() => {
    captchaLoading.value = false
  })
}

// 初始化加载验证码
loadCaptcha()

const loadRegisterCaptcha = () => {
  registerCaptchaLoading.value = true
  captchaApi.newCaptcha().then(res => {
    if (res.code === 200) {
      registerCaptchaId.value = res.data.captchaId
      registerCaptchaQuestion.value = res.data.question
      registerCaptchaAnswer.value = ''
    }
  }).finally(() => {
    registerCaptchaLoading.value = false
  })
}

if (authMode.value === 'register') {
  loadRegisterCaptcha()
}

// 发送邮箱验证码
const handleSendEmailCode = () => {
  if (!emailLoginForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailLoginForm.email)) {
    ElMessage.error('请先输入有效的邮箱地址')
    return
  }
  if (!emailLoginForm.captchaAnswer.trim()) {
    ElMessage.error('请先输入图形验证码')
    return
  }
  captchaApi.verifyCaptcha(emailLoginForm.captchaId, emailLoginForm.captchaAnswer).then(res => {
    if (res.code !== 200 || !res.data) {
      ElMessage.error('图形验证码错误')
      loadCaptcha('email')
      return
    }
    sendEmailLoading.value = true
    emailCodeApi.sendCode(emailLoginForm.email).then(res => {
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
        loadCaptcha('email')
      }
    }).finally(() => { sendEmailLoading.value = false })
  })
}

const handleSendRegisterEmailCode = () => {
  if (!registerForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(registerForm.email)) {
    ElMessage.error('请先输入有效的邮箱地址')
    return
  }
  if (!registerCaptchaAnswer.value.trim()) {
    ElMessage.error('请先输入图形验证码')
    return
  }
  captchaApi.verifyCaptcha(registerCaptchaId.value, registerCaptchaAnswer.value).then(res => {
    if (res.code !== 200 || !res.data) {
      ElMessage.error('图形验证码错误')
      loadRegisterCaptcha()
      return
    }
    sendRegisterEmailLoading.value = true
    emailCodeApi.sendCode(registerForm.email).then(res => {
      if (res.code === 200) {
        registerEmailSent.value = true
        registerEmailCountdown.value = 60
        ElMessage.success('验证码已发送到邮箱')
        const timer = setInterval(() => {
          registerEmailCountdown.value--
          if (registerEmailCountdown.value <= 0) clearInterval(timer)
        }, 1000)
      } else {
        ElMessage.error(res.message || '发送失败')
        loadRegisterCaptcha()
      }
    }).finally(() => {
      sendRegisterEmailLoading.value = false
    })
  })
}

// 密码登录处理
const handleLogin = () => {
  loginFormRef.value.validate((valid) => {
    if (!valid) return

    captchaApi.verifyCaptcha(loginForm.captchaId, loginForm.captchaAnswer).then(res => {
      if (res.code !== 200 || !res.data) {
        ElMessage.error('验证码错误，请重新输入')
        loadCaptcha()
        return
      }

      loading.value = true
      userApi.login(loginForm.username, loginForm.password)
        .then(response => {
          if (response.code === 200) {
            setAuthSession(response.data)
            ElMessage.success('登录成功')
            goAfterLogin()
          } else {
            ElMessage.error(response.message || '登录失败')
          }
        })
        .catch(error => {
          ElMessage.error('登录失败，请检查用户名和密码')
          console.error('登录错误:', error)
        })
        .finally(() => {
          loading.value = false
        })
    }).catch(() => {
      ElMessage.error('验证码校验失败')
      loadCaptcha()
    })
  })
}

// 邮箱验证码登录处理
const handleEmailLogin = () => {
  emailLoginFormRef.value.validate((valid) => {
    if (!valid) return

    loading.value = true
    userApi.login(null, null, 'email_code', emailLoginForm.email, emailLoginForm.emailCode)
      .then(response => {
        if (response.code === 200) {
          setAuthSession(response.data)
          ElMessage.success('登录成功')
          goAfterLogin()
        } else {
          ElMessage.error(response.message || '登录失败')
        }
      })
      .catch(error => {
        ElMessage.error('登录失败，请检查邮箱和验证码')
        console.error('登录错误:', error)
      })
      .finally(() => {
        loading.value = false
      })
  })
}

const goToRegister = () => {
  authMode.value = 'register'
  router.replace({
    path: '/login',
    query: { ...route.query, mode: 'register' }
  })
  if (!registerCaptchaId.value) {
    loadRegisterCaptcha()
  }
}

const goToLogin = () => {
  authMode.value = 'login'
  const nextQuery = { ...route.query }
  delete nextQuery.mode
  router.replace({ path: '/login', query: nextQuery })
}

const handleRegister = () => {
  registerFormRef.value.validate((valid) => {
    if (!valid) return

    loading.value = true
    userApi.register({
      username: registerForm.username,
      email: registerForm.email,
      emailCode: registerForm.emailCode,
      password: registerForm.password
    })
      .then(response => {
        if (response.code === 200) {
          registeredUser.value = response.data?.user || null
          showInterestDialog.value = true
        } else {
          ElMessage.error(response.message || '注册失败')
        }
      })
      .catch(error => {
        ElMessage.error('注册失败，请检查输入信息')
        console.error('注册错误:', error)
      })
      .finally(() => {
        loading.value = false
      })
  })
}

const toggleTag = (tag) => {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) selectedTags.value.splice(idx, 1)
  else if (selectedTags.value.length < 10) selectedTags.value.push(tag)
}

const finishInterestSelection = async () => {
  if (registeredUser.value?.id && selectedTags.value.length > 0) {
    try {
      await api.post(`/profile/init/${registeredUser.value.id}`, { tags: selectedTags.value })
    } catch (e) {
      console.warn('兴趣标签初始化失败:', e)
    }
  }
  showInterestDialog.value = false
  ElMessage.success('注册成功，请登录')
  goToLogin()
}

const skipInterest = () => {
  showInterestDialog.value = false
  ElMessage.success('注册成功，请登录')
  goToLogin()
}
</script>

<template>
  <div class="login-container">
    <div class="login-form-wrapper">
      <div class="login-form">
        <!-- 关闭按钮 -->
        <button class="close-btn" @click="router.back()">
          <el-icon><Close /></el-icon>
        </button>

        <div class="login-logo">
          <el-icon><VideoPlay /></el-icon>
          <span>哔哩哔哩</span>
        </div>
        <h2 class="login-title">{{ authMode === 'login' ? '登录' : '注册' }}</h2>

        <template v-if="authMode === 'login'">
        <!-- 登录方式切换 -->
        <div class="login-tabs">
          <span :class="{ active: loginMode === 'password' }" @click="loginMode = 'password'">密码登录</span>
          <span :class="{ active: loginMode === 'email_code' }" @click="loginMode = 'email_code'">验证码登录</span>
        </div>

        <!-- 密码登录 -->
        <el-form
          v-if="loginMode === 'password'"
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          label-width="0px"
          class="form"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名/邮箱"
              prefix-icon="el-icon-user"
              clearable
            ></el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="el-icon-lock"
              show-password
            ></el-input>
          </el-form-item>

          <el-form-item prop="captchaAnswer">
            <div class="captcha-row">
              <el-input
                v-model="loginForm.captchaAnswer"
                placeholder="请输入右侧算术题答案"
                prefix-icon="el-icon-question"
                @keydown.enter="handleLogin"
                style="flex:1"
              ></el-input>
              <span class="captcha-question" :class="{ loading: captchaLoading }">{{ loginForm.captchaQuestion }}</span>
              <el-button link type="primary" @click="loadCaptcha()" :loading="captchaLoading" title="刷新验证码">刷新</el-button>
            </div>
          </el-form-item>

          <div class="form-actions">
            <el-checkbox v-model="loginForm.rememberMe">记住我</el-checkbox>
            <el-button type="text" class="forget-password" @click="router.push('/forgot-password')">忘记密码？</el-button>
          </div>

          <el-form-item>
            <el-button type="primary" @click="handleLogin" class="login-btn" block :loading="loading">
              登录
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 邮箱验证码登录 -->
        <el-form
          v-else
          ref="emailLoginFormRef"
          :model="emailLoginForm"
          :rules="emailLoginRules"
          label-width="0px"
          class="form"
        >
          <el-form-item prop="email">
            <el-input
              v-model="emailLoginForm.email"
              type="email"
              placeholder="请输入邮箱"
              prefix-icon="el-icon-message"
              clearable
            ></el-input>
          </el-form-item>

          <el-form-item>
            <div class="captcha-row">
              <el-input
                v-model="emailLoginForm.captchaAnswer"
                placeholder="图形验证码"
                style="flex:1"
              ></el-input>
              <span class="captcha-question" :class="{ loading: captchaLoading }">{{ emailLoginForm.captchaQuestion }}</span>
              <el-button link type="primary" @click="loadCaptcha('email')" :loading="captchaLoading">刷新</el-button>
            </div>
          </el-form-item>

          <el-form-item prop="emailCode">
            <div class="email-code-row">
              <el-input
                v-model="emailLoginForm.emailCode"
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

          <el-form-item>
            <el-button type="primary" @click="handleEmailLogin" class="login-btn" block :loading="loading">
              登录
            </el-button>
          </el-form-item>
        </el-form>

        <div class="register-link">
          <span>还没有账号？</span>
          <el-button type="text" @click="goToRegister">立即注册</el-button>
        </div>
        </template>

        <template v-else>
          <el-form
            ref="registerFormRef"
            :model="registerForm"
            :rules="registerRules"
            label-width="0px"
            class="form"
          >
            <el-form-item prop="username">
              <el-input
                v-model="registerForm.username"
                placeholder="请输入用户名（3-20个字符）"
                prefix-icon="el-icon-user"
                clearable
              />
            </el-form-item>

            <el-form-item prop="email">
              <el-input
                v-model="registerForm.email"
                type="email"
                placeholder="请输入邮箱"
                prefix-icon="el-icon-message"
                clearable
              />
            </el-form-item>

            <el-form-item>
              <div class="captcha-row">
                <el-input
                  v-model="registerCaptchaAnswer"
                  placeholder="图形验证码"
                  style="flex:1"
                />
                <span class="captcha-question" :class="{ loading: registerCaptchaLoading }">{{ registerCaptchaQuestion }}</span>
                <el-button link type="primary" @click="loadRegisterCaptcha" :loading="registerCaptchaLoading">刷新</el-button>
              </div>
            </el-form-item>

            <el-form-item prop="emailCode">
              <div class="email-code-row">
                <el-input
                  v-model="registerForm.emailCode"
                  placeholder="邮箱验证码"
                  prefix-icon="el-icon-message"
                />
                <el-button
                  type="primary"
                  :disabled="registerEmailSent && registerEmailCountdown > 0"
                  :loading="sendRegisterEmailLoading"
                  @click="handleSendRegisterEmailCode"
                >
                  {{ registerEmailCountdown > 0 ? registerEmailCountdown + 's' : '发送验证码' }}
                </el-button>
              </div>
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="registerForm.password"
                type="password"
                placeholder="请输入密码（包含大小写字母和数字）"
                prefix-icon="el-icon-lock"
                show-password
              />
            </el-form-item>

            <el-form-item prop="confirmPassword">
              <el-input
                v-model="registerForm.confirmPassword"
                type="password"
                placeholder="请确认密码"
                prefix-icon="el-icon-lock"
                show-password
              />
            </el-form-item>

            <el-form-item prop="agreeTerms">
              <el-checkbox v-model="registerForm.agreeTerms">
                我已阅读并同意
                <a href="#" class="link">《用户协议》</a>
                和
                <a href="#" class="link">《隐私政策》</a>
              </el-checkbox>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleRegister" class="login-btn" block :loading="loading">
                注册
              </el-button>
            </el-form-item>
          </el-form>

          <div class="register-link">
            <span>已有账号？</span>
            <el-button type="text" @click="goToLogin">立即登录</el-button>
          </div>
        </template>
      </div>
    </div>
  </div>

  <el-dialog
    v-model="showInterestDialog"
    title="选择你感兴趣的内容"
    width="560px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
  >
    <p class="interest-tip">选择 1-10 个你感兴趣的标签，我们会为你推荐更精准的内容</p>
    <div class="interest-tags">
      <button
        v-for="tag in allInterestTags"
        :key="tag"
        type="button"
        class="interest-tag"
        :class="{ selected: selectedTags.includes(tag) }"
        @click="toggleTag(tag)"
      >
        {{ tag }}
      </button>
    </div>
    <template #footer>
      <el-button @click="skipInterest">跳过</el-button>
      <el-button type="primary" @click="finishInterestSelection" :disabled="selectedTags.length === 0">
        完成 ({{ selectedTags.length }}/10)
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.login-container {
  width: 100%;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-form-wrapper {
  width: 100%;
  max-width: 420px;
}

.login-form {
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

.login-form {
  background-color: #fff;
  border-radius: 10px;
  padding: 30px;
  box-shadow: 0 5px 20px rgba(0, 0, 0, 0.1);
}

.login-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
  color: #00a1d6;
  margin-bottom: 20px;
}

.login-logo el-icon {
  font-size: 32px;
  margin-right: 10px;
}

.login-title {
  font-size: 20px;
  font-weight: 500;
  color: #333;
  margin: 0 0 20px 0;
  text-align: center;
}

.login-tabs {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  border-bottom: 2px solid #eee;
}

.login-tabs span {
  padding-bottom: 10px;
  color: #999;
  cursor: pointer;
  font-size: 14px;
}

.login-tabs span.active {
  color: #00a1d6;
  border-bottom: 2px solid #00a1d6;
}

.form {
  width: 100%;
}

.form .el-form-item {
  margin-bottom: 20px;
}

.form-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.forget-password {
  color: #00a1d6;
}

.link {
  color: #00a1d6;
  text-decoration: none;
}

.captcha-row {
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

.login-btn {
  height: 40px;
  border-radius: 20px;
  font-size: 16px;
}

.register-link {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #999;
  margin-top: 20px;
}

.register-link el-button {
  color: #00a1d6;
  padding: 0;
}

.interest-tip {
  color: #61666d;
  margin-bottom: 16px;
  font-size: 14px;
}

.interest-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.interest-tag {
  padding: 6px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
  border: 2px solid #e3e5e7;
  background: #fff;
  color: #61666d;
  transition: all 0.2s;
}

.interest-tag.selected {
  border-color: #00a1d6;
  background: rgba(0, 161, 214, 0.1);
  color: #00a1d6;
}
</style>
