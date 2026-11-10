<template>
    <div class="login-register">
        <!-- 主区（对齐 B站 .login__main 900×304 flex） -->
        <div class="login-main">
        <!-- 左栏：扫码登录区（对齐 B站 .login-scan.main__left，仅二维码+文案，无TV图标） -->
        <div class="login-brand">
            <div class="login-scan__qrcode">
                <img class="login-scan__qr-img" :src="qrImg" alt="扫码登录">
            </div>
            <div class="login-scan__txt">
                <p>使用手机</p>
                <p>哔哩哔哩客户端</p>
                <p>扫码登录</p>
            </div>
        </div>

        <!-- 中分隔线（对齐 B站 .main__middle-line） -->
        <div class="login-divider"></div>

        <!-- 右栏：表单区（对齐 B站 .main__right） -->
        <div class="login-panel">
        <!-- 登录 -->
        <div v-if="dialogMode === 'login'" class="login-form">
            <!-- 登录方式 tab 栏（对齐 B站 .tabs_wp：18px #9499a0 500，分隔线 1×20 #ddd） -->
            <div class="tabs_wp">
                <div :class="{ tab_active: loginMode === 'password' }" @click="loginMode = 'password'">密码登录</div>
                <div></div>
                <div :class="{ tab_active: loginMode === 'sms' }" @click="loginMode = 'sms'">短信登录</div>
            </div>

            <!-- 验证码登录 tab 暂时注释 ↓
            <div class="login-tabs">
                <span :class="{ active: loginMode === 'password' }" @click="loginMode = 'password'">密码登录</span>
                <span :class="{ active: loginMode === 'email_code' }" @click="loginMode = 'email_code'">验证码登录</span>
            </div>
            -->

            <el-input
                v-model="loginForm.username"
                class="input"
                placeholder="请输入用户名/邮箱"
                :prefix-icon="User"
                clearable
                @keydown.enter="handleLogin"
            />
            <el-input
                v-model="loginForm.password"
                class="input"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
                @keydown.enter="handleLogin"
            />

            <!-- 邮箱验证码登录区 暂时注释 ↓
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
            注释 ↑ -->

            <div class="submit" @click="handleLogin()">登录</div>

            <!-- 其他登录方式（对齐 B站：按钮下方紧跟第三方登录区，标题13px #999 mb12 + 三图标28×28） -->
            <div class="third-party-login-wrapper">
                <div class="title">其他方式登录</div>
                <div class="sns">
                    <div class="sns__item" v-for="sns in snsIcons" :key="sns.name">
                        <img class="sns__icon" :src="sns.img" :alt="sns.name">
                        <span class="sns__name">{{ sns.name }}</span>
                    </div>
                </div>
            </div>

            <div class="switch-mode">
                <span>还没有账号？</span>
                <span class="switch-btn" @click="switchToRegister">立即注册</span>
            </div>
            <div class="form-actions">
                <el-checkbox v-model="loginForm.rememberMe">记住我</el-checkbox>
                <el-button link class="forget-password" @click="this.$router.push('/forgot-password')">忘记密码？</el-button>
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
            <!-- 注册邮箱验证码区 暂时注释 ↓
            <el-input
                v-model="registerForm.email"
                class="input"
                type="email"
                placeholder="请输入邮箱"
                :prefix-icon="Message"
                clearable
                @keydown.enter="handleRegister"
            />
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
            注释 ↑ -->
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
        </div><!-- /.login-panel -->
        </div><!-- /.login-main -->

        <!-- 通栏协议区（对齐 B站 .login-protocol 900×34） -->
        <div class="login-protocol">
            <p>未注册过哔哩哔哩的手机号，我们将自动帮你注册账号</p>
            <p>登录即代表你同意<label><a class="clickable" href="https://www.bilibili.com/protocal/licence.html" target="_blank" rel="noopener noreferrer">用户协议</a></label>与<label><a class="clickable" href="https://www.bilibili.com/blackboard/privacy-pc.html" target="_blank" rel="noopener noreferrer">隐私政策</a></label></p>
        </div>
    </div>
</template>

<script lang="ts">
import { ElMessage } from 'element-plus';
import { User, Lock, Message, VideoPlay } from '@element-plus/icons-vue';
// captchaApi / emailCodeApi 暂时注释：验证码、邮箱验证码功能已关闭
import { userApi } from '@/api/client';
import { setAuthSession } from '@/utils/auth';
// 二维码静态资源（从 B站 passport 快照抠出的真实 140×140 二维码图，像素级对齐用）
import qrImg from '@/assets/site/img/login-qr.png';
// 第三方登录图标（从 B站 passport 快照抠出的真实图标）
import snsWechat from '@/assets/site/img/sns-wechat.png';
import snsWeibo from '@/assets/site/img/sns-weibo.png';
import snsQq from '@/assets/site/img/sns-qq.png';

export default {
    name: "LoginRegister",
    components: { User, Lock, Message, VideoPlay },
    data() {
        return {
            qrImg,
            snsIcons: [
                { name: '微信登录', img: snsWechat },
                { name: '微博登录', img: snsWeibo },
                { name: 'QQ登录', img: snsQq },
            ],
            dialogMode: 'login',
            loginMode: 'password',
            loginForm: {
                username: "",
                password: "",
                rememberMe: false
            },
            emailLoginForm: {
                email: "",
                emailCode: ""
            },
            registerForm: {
                username: "",
                email: "",
                emailCode: "",
                password: "",
                confirmPassword: "",
                agreeTerms: false
            },
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
                this.handleLogin()
            } else {
                this.handleRegister()
            }
        },

        // ================= 验证码 / 邮箱验证码功能暂时注释 ↓ =================
        // loadCaptcha() {
        //     this.captchaLoading = true
        //     captchaApi.newCaptcha().then((res: any) => {
        //         if (res.code === 200) {
        //             this.loginCaptchaId = res.data.captchaId
        //             this.loginCaptchaQuestion = res.data.question
        //             this.loginCaptchaAnswer = ''
        //         }
        //     }).finally(() => { this.captchaLoading = false })
        // },
        // loadRegisterCaptcha() {
        //     this.registerCaptchaLoading = true
        //     captchaApi.newCaptcha().then((res: any) => {
        //         if (res.code === 200) {
        //             this.registerCaptchaId = res.data.captchaId
        //             this.registerCaptchaQuestion = res.data.question
        //             this.registerCaptchaAnswer = ''
        //         }
        //     }).finally(() => { this.registerCaptchaLoading = false })
        // },

        // handleSendEmailCode() { ... 邮箱登录验证码发送 }
        // handleSendRegisterEmailCode() { ... 注册邮箱验证码发送 }

        // handleEmailLogin() { ... 邮箱验证码登录 }
        // ================= 注释 ↑ =================

        // ==== 登录 (密码直连) ====
        handleLogin() {
            if (!this.loginForm.username.trim()) { ElMessage.error('请输入用户名/邮箱'); return }
            if (!this.loginForm.password) { ElMessage.error('请输入密码'); return }
            this.doLogin(userApi.login(this.loginForm.username, this.loginForm.password))
        },

        // ==== 注册 (直连) ====
        handleRegister() {
            if (!this.registerForm.username.trim() || this.registerForm.username.length < 3) { ElMessage.error('用户名至少 3 个字符'); return }
            if (!/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d_]{6,20}$/.test(this.registerForm.password || '')) { ElMessage.error('密码必须包含大小写字母和数字（6-20位）'); return }
            if (this.registerForm.password !== this.registerForm.confirmPassword) { ElMessage.error('两次输入的密码不一致'); return }
            if (!this.registerForm.agreeTerms) { ElMessage.error('请阅读并同意用户协议和隐私政策'); return }
            this.loading = true
            userApi.register({
                username: this.registerForm.username,
                email: this.registerForm.email || '',
                emailCode: this.registerForm.emailCode || '',
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
        },

        // ==== 内部 ====
        doLogin(promise: any) {
            this.loading = true
            promise.then((response: any) => {
                if (response.code === 200) {
                    const data = response.data || {}
                    // 凭证已由服务端写进 HttpOnly cookie，客户端只保留展示信息
                    setAuthSession({
                        user: data.user || data
                    })
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
/* 外层（对齐 B站 .login_wp 900×588：pad-top102[含bread-cut20] + 主区304 + gap40 + 协议34 + pad-bottom108） */
.login-register {
    display: block;
    width: 900px;
    box-sizing: border-box;
    padding: 102px 0 108px;
    background: #fff;
}

/* 主区（对齐 B站 .login__main 900×304 flex） */
.login-main {
    display: flex;
    align-items: flex-start;
    width: 900px;
    height: 304px;
    box-sizing: border-box;
}

/* 左栏扫码登录区（对齐 .login-scan.main__left 386 宽，仅二维码+文案，内容顶从 y146） */
.login-brand {
    flex: 0 0 386px;
    display: flex;
    flex-direction: column;
    align-items: center;
    box-sizing: border-box;
    padding-top: 44px; /* 主区顶 y102 → 二维码 y146 */
}

/* 二维码框（对齐 .login-scan__qrcode 158×158 border #e3e5e7 radius6.84 pad8，相对面板 y146） */
.login-scan__qrcode {
    box-sizing: border-box;
    border: 1px solid #e3e5e7;
    border-radius: 6.84393px;
    width: 158px;
    height: 158px;
    padding: 8px;
    position: relative;
}
.login-scan__qr-img {
    width: 140px;
    height: 140px;
    display: block;
}
/* 扫码文案区（对齐 .login-scan__txt 386×90，二维码底 y304 → 文案 y312，2行×19行高居中） */
.login-scan__txt {
    width: 386px;
    margin-top: 8px;
    height: 90px;
    box-sizing: border-box;
}
.login-scan__txt p {
    text-align: center;
    color: #61666d;
    font-size: 14px;
    font-weight: 400;
    line-height: 19px;
    margin: 0;
    height: 19px;
}

/* 中分隔线（对齐 .main__middle-line 1×260 #ddd margin 44px 48px 0） */
.login-divider {
    flex: 0 0 1px;
    width: 1px;
    background: #ddd;
    margin: 44px 48px 0;
    align-self: flex-start;
    height: 260px;
}

/* 右栏表单区（对齐 .main__right 402 宽，内容左对齐 x483，顶从 y106=102+4，tab栏占44后输入框落y150） */
.login-panel {
    flex: 0 0 402px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 4px 0 0;
    box-sizing: border-box;
}
/* 输入框微偏：B站输入框有真实1px border(占布局)故 x484 y151，本地面板用box-shadow画边框(不占布局)故 x483 y150。
   二者机制不同，强行偏移会推动按钮下移1px。取舍：保证按钮(大色块视觉锚点)精确对齐，输入框容忍1px。 */
.login-form, .login-panel > div { width: 400px; box-sizing: border-box; }

/* 登录方式 tab 栏（对齐 B站 .tabs_wp：20高 mb24 内容居中，文字18px #9499a0 500，分隔线1×20 #ddd） */
.tabs_wp {
    align-items: center;
    justify-content: center; /* tab内容居中，对齐 B站 tabs_wp 中心≈表单中心 */
    height: 20px;
    width: 100%; /* 撑满表单，否则父容器 align-items:flex-start 收缩宽度 */
    margin: 0 auto 24px;
    display: flex;
}
.tabs_wp div {
    color: #9499a0;
    font-size: 18px;
    font-style: normal;
    font-weight: 500;
    line-height: 100%;
}
.tabs_wp div:nth-child(odd) { cursor: pointer; min-width: 72px; text-align: center; }
.tabs_wp div:nth-child(2n) { background: #ddd; width: 1px; height: 20px; margin: auto 20px; }
.tabs_wp div.tab_active { color: #00a1d6; }

/* 其他登录方式（对齐 B站 .third-party-login-wrapper：wrapper收缩到内容宽216居中 + 标题13px #999 mb12 + 图标28×28 + 名称12px） */
.third-party-login-wrapper {
    margin: 26px auto 0; /* mt26：按钮下方到第三方区对齐 B站 y328 */
    width: fit-content; /* 收缩到内容宽(3×72=216)并居中，对齐 B站 wrapper x576 */
}
.third-party-login-wrapper .title {
    text-align: center;
    color: #999;
    margin-bottom: 12px;
    font-size: 13px;
    line-height: 16px;
}
.third-party-login-wrapper .sns {
    justify-content: center;
    align-items: flex-start;
    display: flex;
}
.third-party-login-wrapper .sns__item {
    box-sizing: border-box;
    cursor: pointer;
    flex-direction: column;
    align-items: center;
    width: 72px;
    padding: 0 2px;
    display: flex;
}
.third-party-login-wrapper .sns__icon {
    width: 28px;
    height: 28px;
    margin-bottom: 2px;
    display: block;
}
.third-party-login-wrapper .sns__name {
    text-align: center;
    font-size: 12px;
    line-height: 16px;
    color: #61666d;
}

.dialog-logo { display: none; }

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
    color: #18191c;
    margin-top: 44px; /* 补偿 panel padding-top 48→4：注册无 tab 栏，保持标题落 y150 */
    margin-bottom: 16px;
    text-align: center;
}

/* 输入框（对齐 B站 .form__item 44 高 + .tab__form 边框 #e3e5e7） */
.input {
    margin-bottom: 0;
    width: 100%;
}

.input /deep/ .el-input__wrapper {
    border-radius: 0;
    height: 44px;
    padding: 12px 20px;
    box-shadow: 0 0 0 1px #e3e5e7 inset;
    background: transparent;
}

.input /deep/ .el-input__wrapper:hover { box-shadow: 0 0 0 1px #c9ccd0 inset; }

.input /deep/ .el-input__inner {
    font-size: 14px;
    font-weight: 400;
    height: 20px;
    line-height: 20px;
    color: #18191c;
}

.input /deep/ .el-input__inner::placeholder { color: #c9ccd0; }

/* 表单框整体圆角（对齐 .tab__form radius 8 + 输入项分隔线） */
.login-form .input:first-of-type /deep/ .el-input__wrapper { border-radius: 8px 8px 0 0; }
.login-form .input:nth-of-type(2) /deep/ .el-input__wrapper { border-radius: 0 0 8px 8px; }

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
    color: #61666d;
}

.captcha-question.loading {
    color: #9499a0;
}

.email-code-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
}

.email-code-row .input {
    margin-bottom: 0;
    flex: 1;
}

/* 记住我/忘记密码（对齐 B站 .login-protocol 区 margin-top） */
.form-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
    margin-bottom: 0;
    font-size: 12px;
}

.forget-password {
    color: #00a1d6;
    padding: 0;
    font-size: 12px;
}

.agree-terms {
    font-size: 12px;
    color: #9499a0;
    margin-top: 12px;
    margin-bottom: 0;
}

.link {
    color: #00a1d6;
    text-decoration: none;
}

/* 登录/注册按钮（对齐 B站 .btn_primary 194×40 #00a1d6 radius8；大按钮 400 宽） */
.submit {
    color: #fff;
    width: 400px;
    height: 40px;
    line-height: 40px;
    border-radius: 8px;
    background-color: #00a1d6;
    text-align: center;
    padding: 0;
    font-size: 14px;
    font-weight: 400;
    cursor: pointer;
    transition: all 0.2s;
    margin-top: 24px;
    box-sizing: border-box;
}

.submit:hover {
    background-color: #00aeec;
}

.switch-mode {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    margin-top: 16px;
    font-size: 12px;
    color: #9499a0;
}

.switch-btn {
    color: #00a1d6;
    cursor: pointer;
}

/* 通栏协议区（对齐 B站 .login-protocol 900×34 mt40 12px #9499a0 居中，2行×17行高） */
.login-protocol {
    color: #9499a0;
    margin-top: 40px;
    font-size: 12px;
    font-weight: 400;
    line-height: 17px;
}
.login-protocol p {
    display: flex;
    justify-content: center;
    margin: 0;
    line-height: 17px;
    height: 17px;
}
.login-protocol label { margin: 0 6px; }
.login-protocol .clickable { color: #00a1d6; cursor: pointer; text-decoration: none; }
</style>