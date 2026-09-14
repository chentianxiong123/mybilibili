<template>
    <div class="login-register">
        <div class="dialog-logo">
            <el-icon><VideoPlay /></el-icon>
            <span>哔哩哔哩</span>
        </div>

        <!-- 登录 -->
        <div v-if="dialogMode === 'login'">
            <div class="login-tabs">
                <span :class="{ active: loginMode === 'password' }" @click="loginMode = 'password'">密码登录</span>
                <span :class="{ active: loginMode === 'email_code' }" @click="loginMode = 'email_code'">验证码登录</span>
            </div>

            <el-input
                v-if="loginMode === 'password'"
                v-model="loginForm.username"
                class="input"
                placeholder="请输入用户名/邮箱"
                :prefix-icon="User"
                clearable
                @keydown.enter="handleLogin"
            />
            <el-input
                v-if="loginMode === 'password'"
                v-model="loginForm.password"
                class="input"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
                @keydown.enter="handleLogin"
            />

            <el-input
                v-if="loginMode === 'email_code'"
                v-model="emailLoginForm.email"
                class="input"
                type="email"
                placeholder="请输入邮箱"
                :prefix-icon="Message"
                clearable
                @keydown.enter="handleEmailLogin"
            />

            <div class="captcha-row">
                <el-input
                    v-model="loginCaptchaAnswer"
                    class="input"
                    placeholder="请输入右侧算术题答案"
                    :prefix-icon="Key"
                    clearable
                    @keydown.enter="handleLogin"
                />
                <span class="captcha-question" :class="{ loading: captchaLoading }">{{ loginCaptchaQuestion }}</span>
                <el-button link type="primary" @click="loadCaptcha()" :loading="captchaLoading">刷新</el-button>
            </div>

            <div v-if="loginMode === 'email_code'" class="email-code-row">
                <el-input
                    v-model="emailLoginForm.emailCode"
                    class="input"
                    placeholder="邮箱验证码"
                    :prefix-icon="Message"
                    clearable
                    @keydown.enter="handleEmailLogin"
                />
                <el-button
                    type="primary"
                    :disabled="emailSent && emailCountdown > 0"
                    :loading="sendEmailLoading"
                    @click="handleSendEmailCode"
                >
                    {{ emailCountdown > 0 ? emailCountdown + 's' : '发送验证码' }}
                </el-button>
            </div>

            <div class="form-actions">
                <el-checkbox v-model="loginForm.rememberMe">记住我</el-checkbox>
                <el-button link class="forget-password" @click="this.$router.push('/forgot-password')">忘记密码？</el-button>
            </div>

            <div class="submit" @click="loginMode === 'password' ? handleLogin() : handleEmailLogin()">登录</div>
            <div class="switch-mode">
                <span>还没有账号？</span>
                <span class="switch-btn" @click="switchToRegister">立即注册</span>
            </div>
        </div>

        <!-- 注册 -->
        <div v-if="dialogMode === 'register'">
            <div class="register-title">注册</div>
            <el-input
                v-model="registerForm.username"
                class="input"
                placeholder="用户名（3-20个字符）"
                :prefix-icon="User"
                clearable
                @keydown.enter="handleRegister"
            />
            <el-input
                v-model="registerForm.email"
                class="input"
                type="email"
                placeholder="请输入邮箱"
                :prefix-icon="Message"
                clearable
                @keydown.enter="handleRegister"
            />
            <div class="captcha-row">
                <el-input
                    v-model="registerCaptchaAnswer"
                    class="input"
                    placeholder="图形验证码"
                    :prefix-icon="Key"
                    clearable
                    @keydown.enter="handleRegister"
                />
                <span class="captcha-question" :class="{ loading: registerCaptchaLoading }">{{ registerCaptchaQuestion }}</span>
                <el-button link type="primary" @click="loadRegisterCaptcha" :loading="registerCaptchaLoading">刷新</el-button>
            </div>
            <div class="email-code-row">
                <el-input
                    v-model="registerForm.emailCode"
                    class="input"
                    placeholder="邮箱验证码"
                    :prefix-icon="Message"
                    clearable
                    @keydown.enter="handleRegister"
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
            <el-input
                v-model="registerForm.password"
                class="input"
                type="password"
                placeholder="密码（包含大小写字母和数字）"
                :prefix-icon="Lock"
                show-password
                @keydown.enter="handleRegister"
            />
            <el-input
                v-model="registerForm.confirmPassword"
                class="input"
                type="password"
                placeholder="请确认密码"
                :prefix-icon="Lock"
                show-password
                @keydown.enter="handleRegister"
            />
            <el-checkbox v-model="registerForm.agreeTerms" class="agree-terms">
                我已阅读并同意<span class="link">《用户协议》</span>和<span class="link">《隐私政策》</span>
            </el-checkbox>
            <div class="submit" @click="handleRegister">注册</div>
            <div class="switch-mode">
                <span>已有账号？</span>
                <span class="switch-btn" @click="switchToLogin">立即登录</span>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { ElMessage } from 'element-plus';
import { User, Lock, Message, VideoPlay } from '@element-plus/icons-vue';
import { captchaApi, emailCodeApi, userApi } from '@/api/client';

export default {
    name: "LoginRegister",
    components: { User, Lock, Message, VideoPlay },
    data() {
        return {
            dialogMode: 'login',
            loginMode: 'password',
            loginForm: {
                username: "",
                password: "",
                rememberMe: false,
                captchaAnswer: ""
            },
            emailLoginForm: {
                email: "",
                emailCode: "",
                captchaAnswer: ""
            },
            registerForm: {
                username: "",
                email: "",
                emailCode: "",
                password: "",
                confirmPassword: "",
                agreeTerms: false
            },
            // 验证码相关
            loginCaptchaId: "",
            loginCaptchaQuestion: "",
            loginCaptchaAnswer: "",
            captchaLoading: false,
            registerCaptchaId: "",
            registerCaptchaQuestion: "",
            registerCaptchaAnswer: "",
            registerCaptchaLoading: false,
            sendEmailLoading: false,
            emailSent: false,
            emailCountdown: 0,
            sendRegisterEmailLoading: false,
            registerEmailSent: false,
            registerEmailCountdown: 0,
            loading: false,
        }
    },
    mounted() {
        this.loadCaptcha();
        document.addEventListener('keydown', (e) => this.handleKeyboard(e));
    },
    beforeUnmount() {
        document.removeEventListener('keydown', (e) => this.handleKeyboard(e));
    },
    methods: {
        // 监听键盘回车触发登录
        handleKeyboard(event: any) {
            if (event.keyCode !== 13) return
            if (this.dialogMode === 'login') {
                this.loginMode === 'password' ? this.handleLogin() : this.handleEmailLogin()
            } else {
                this.handleRegister()
            }
        },

        // ==== 验证码 ====
        loadCaptcha() {
            this.captchaLoading = true
            captchaApi.newCaptcha().then((res: any) => {
                if (res.code === 200) {
                    this.loginCaptchaId = res.data.captchaId
                    this.loginCaptchaQuestion = res.data.question
                    this.loginCaptchaAnswer = ''
                }
            }).finally(() => { this.captchaLoading = false })
        },
        loadRegisterCaptcha() {
            this.registerCaptchaLoading = true
            captchaApi.newCaptcha().then((res: any) => {
                if (res.code === 200) {
                    this.registerCaptchaId = res.data.captchaId
                    this.registerCaptchaQuestion = res.data.question
                    this.registerCaptchaAnswer = ''
                }
            }).finally(() => { this.registerCaptchaLoading = false })
        },

        // ==== 邮箱验证码 ====
        handleSendEmailCode() {
            if (!this.emailLoginForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.emailLoginForm.email)) {
                ElMessage.error('请先输入有效的邮箱地址')
                return
            }
            if (!this.loginCaptchaAnswer.trim()) {
                ElMessage.error('请先输入图形验证码')
                return
            }
            captchaApi.verifyCaptcha(this.loginCaptchaId, this.loginCaptchaAnswer).then((res: any) => {
                if (res.code !== 200 || !res.data) {
                    ElMessage.error('图形验证码错误')
                    this.loadCaptcha()
                    return
                }
                this.sendEmailLoading = true
                emailCodeApi.sendCode(this.emailLoginForm.email).then((res: any) => {
                    if (res.code === 200) {
                        this.emailSent = true
                        this.emailCountdown = 60
                        ElMessage.success('验证码已发送到邮箱')
                        const timer = setInterval(() => {
                            this.emailCountdown--
                            if (this.emailCountdown <= 0) clearInterval(timer)
                        }, 1000)
                    } else {
                        ElMessage.error(res.message || '发送失败')
                        this.loadCaptcha()
                    }
                }).finally(() => { this.sendEmailLoading = false })
            })
        },

        handleSendRegisterEmailCode() {
            if (!this.registerForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.registerForm.email)) {
                ElMessage.error('请先输入有效的邮箱地址')
                return
            }
            if (!this.registerCaptchaAnswer.trim()) {
                ElMessage.error('请先输入图形验证码')
                return
            }
            captchaApi.verifyCaptcha(this.registerCaptchaId, this.registerCaptchaAnswer).then((res: any) => {
                if (res.code !== 200 || !res.data) {
                    ElMessage.error('图形验证码错误')
                    this.loadRegisterCaptcha()
                    return
                }
                this.sendRegisterEmailLoading = true
                emailCodeApi.sendCode(this.registerForm.email).then((res: any) => {
                    if (res.code === 200) {
                        this.registerEmailSent = true
                        this.registerEmailCountdown = 60
                        ElMessage.success('验证码已发送到邮箱')
                        const timer = setInterval(() => {
                            this.registerEmailCountdown--
                            if (this.registerEmailCountdown <= 0) clearInterval(timer)
                        }, 1000)
                    } else {
                        ElMessage.error(res.message || '发送失败')
                        this.loadRegisterCaptcha()
                    }
                }).finally(() => { this.sendRegisterEmailLoading = false })
            })
        },

        // ==== 登录 ====
        handleLogin() {
            if (!this.loginForm.username.trim()) { ElMessage.error('请输入用户名/邮箱'); return }
            if (!this.loginForm.password) { ElMessage.error('请输入密码'); return }
            if (!this.loginCaptchaAnswer.trim()) { ElMessage.error('请输入图形验证码'); return }
            this.loading = true
            captchaApi.verifyCaptcha(this.loginCaptchaId, this.loginCaptchaAnswer).then((res: any) => {
                if (res.code !== 200 || !res.data) {
                    ElMessage.error('验证码错误，请重新输入')
                    this.loadCaptcha()
                    return
                }
                this.doLogin(userApi.login(this.loginForm.username, this.loginForm.password))
            }).catch(() => {
                ElMessage.error('验证码校验失败')
                this.loadCaptcha()
            })
        },

        handleEmailLogin() {
            if (!this.emailLoginForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.emailLoginForm.email)) {
                ElMessage.error('请输入有效的邮箱地址')
                return
            }
            if (!this.emailLoginForm.emailCode) { ElMessage.error('请输入邮箱验证码'); return }
            if (!this.loginCaptchaAnswer.trim()) { ElMessage.error('请输入图形验证码'); return }
            this.loading = true
            captchaApi.verifyCaptcha(this.loginCaptchaId, this.loginCaptchaAnswer).then((res: any) => {
                if (res.code !== 200 || !res.data) {
                    ElMessage.error('验证码错误，请重新输入')
                    this.loadCaptcha()
                    return
                }
                this.doLogin(userApi.login(null, null, 'email_code', this.emailLoginForm.email, this.emailLoginForm.emailCode))
            }).catch(() => {
                ElMessage.error('验证码校验失败')
                this.loadCaptcha()
            })
        },

        // ==== 注册 ====
        handleRegister() {
            if (!this.registerForm.username.trim() || this.registerForm.username.length < 3) { ElMessage.error('用户名至少 3 个字符'); return }
            if (!this.registerForm.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.registerForm.email)) { ElMessage.error('请输入有效的邮箱地址'); return }
            if (!this.registerForm.emailCode) { ElMessage.error('请输入邮箱验证码'); return }
            if (!this.registerCaptchaAnswer.trim()) { ElMessage.error('请输入图形验证码'); return }
            if (!/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d_]{6,20}$/.test(this.registerForm.password || '')) { ElMessage.error('密码必须包含大小写字母和数字（6-20位）'); return }
            if (this.registerForm.password !== this.registerForm.confirmPassword) { ElMessage.error('两次输入的密码不一致'); return }
            if (!this.registerForm.agreeTerms) { ElMessage.error('请阅读并同意用户协议和隐私政策'); return }
            this.loading = true
            captchaApi.verifyCaptcha(this.registerCaptchaId, this.registerCaptchaAnswer).then((res: any) => {
                if (res.code !== 200 || !res.data) {
                    ElMessage.error('图形验证码错误')
                    this.loadRegisterCaptcha()
                    return
                }
                userApi.register({
                    username: this.registerForm.username,
                    email: this.registerForm.email,
                    emailCode: this.registerForm.emailCode,
                    password: this.registerForm.password
                }).then((response: any) => {
                    if (response.code === 200) {
                        ElMessage.success('注册成功，请登录')
                        this.switchToLogin()
                    } else {
                        ElMessage.error(response.message || '注册失败')
                    }
                }).catch(() => {
                    ElMessage.error('注册失败，请检查输入信息')
                }).finally(() => { this.loading = false })
            })
        },

        // ==== 内部 ====
        doLogin(promise: any) {
            this.loading = true
            promise.then((response: any) => {
                if (response.code === 200) {
                    const data = response.data || {}
                    localStorage.setItem("teri_token", data.token)
                    localStorage.setItem("teri_refresh_token", data.refresh_token || data.refreshToken || '')
                    this.$store.commit("updateUser", data.user || data)
                    this.$store.commit("updateIsLogin", true)
                    this.$emit("loginSuccess")
                    ElMessage.success('登录成功')
                    window.location.reload()
                } else {
                    ElMessage.error(response.message || '登录失败')
                }
            }).catch(() => {
                ElMessage.error('登录失败，请检查用户名和密码')
            }).finally(() => { this.loading = false })
        },

        switchToRegister() {
            this.dialogMode = 'register'
            if (!this.registerCaptchaId) this.loadRegisterCaptcha()
        },
        switchToLogin() {
            this.dialogMode = 'login'
            this.loginMode = 'password'
            this.registerForm.username = ''
            this.registerForm.email = ''
            this.registerForm.emailCode = ''
            this.registerForm.password = ''
            this.registerForm.confirmPassword = ''
            this.registerForm.agreeTerms = false
        },
    }
}
</script>

<style scoped>
.login-register {
    display: block;
    width: 100%;
    padding: 10px 4px 6px;
}

.dialog-logo {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin-bottom: 16px;
}

.dialog-logo .el-icon {
    font-size: 28px;
    color: #00a1d6;
}

.dialog-logo span {
    font-size: 22px;
    font-weight: bold;
    color: #00a1d6;
}

.login-tabs {
    display: flex;
    gap: 20px;
    margin-bottom: 18px;
    border-bottom: 2px solid #eee;
}

.login-tabs span {
    padding-bottom: 8px;
    color: #999;
    cursor: pointer;
    font-size: 14px;
}

.login-tabs span.active {
    color: #00a1d6;
    border-bottom: 2px solid #00a1d6;
}

.register-title {
    font-size: 18px;
    font-weight: 600;
    color: #333;
    margin-bottom: 16px;
    text-align: center;
}

.input {
    margin-bottom: 14px;
    width: 100%;
}

.input /deep/ .el-input__wrapper {
    border-radius: 4px;
    height: 40px;
    padding: 4px 12px;
}

.input /deep/ .el-input__inner {
    font-size: 14px;
    height: 28px;
    line-height: 28px;
}

.captcha-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
}

.captcha-row .input {
    margin-bottom: 0;
    flex: 1;
}

.captcha-question {
    min-width: 84px;
    white-space: nowrap;
    font-size: 13px;
    color: #333;
}

.captcha-question.loading {
    color: #999;
}

.email-code-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
}

.email-code-row .input {
    margin-bottom: 0;
    flex: 1;
}

.form-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 14px;
    font-size: 12px;
}

.forget-password {
    color: #00a1d6;
    padding: 0;
    font-size: 12px;
}

.agree-terms {
    font-size: 12px;
    color: #666;
    margin-bottom: 14px;
}

.link {
    color: #00a1d6;
    text-decoration: none;
}

.submit {
    color: #fff;
    width: 100%;
    border-radius: 4px;
    background-color: #409eff;
    text-align: center;
    padding: 11px 15px;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
}

.submit:hover {
    background-color: #66b1ff;
}

.switch-mode {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    margin-top: 14px;
    font-size: 12px;
    color: #666;
}

.switch-btn {
    color: #409eff;
    cursor: pointer;
}
</style>