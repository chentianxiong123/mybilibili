<template>
    <div class="login-register">
        <div class="dialog-logo">
            <el-icon><VideoPlay /></el-icon>
            <span>哔哩哔哩</span>
        </div>
        <el-tabs v-model="activeTab" class="login-tabs" stretch @tab-click="handleClick">
            <el-tab-pane label="登录" name="login" lazy>
                <div class="login-box">
                    <el-input
                        type="text"
                        class="input"
                        v-model="usernameLogin"
                        placeholder="请输入账号"
                        :prefix-icon="User"
                        clearable
                    />
                    <el-input
                        type="password"
                        show-password
                        class="input"
                        v-model="passwordLogin"
                        placeholder="请输入密码"
                        :prefix-icon="Lock"
                        @keydown.enter="submitLogin"
                    />
                    <div class="submit" @click="submitLogin">登录</div>
                    <div class="tips">登录即代表你同意我们的<span class="agreement">用户协议</span></div>
                </div>
            </el-tab-pane>
            <el-tab-pane label="注册" name="register" lazy>
                <div class="register-box">
                    <el-input
                        type="text"
                        class="input"
                        v-model="usernameRegister"
                        placeholder="请输入账号（3-20个字符）"
                        maxlength="50"
                        :prefix-icon="User"
                        clearable
                    />
                    <el-input
                        type="password"
                        show-password
                        class="input"
                        v-model="passwordRegister"
                        placeholder="请输入密码"
                        :prefix-icon="Lock"
                        @keydown.enter="submitRegister"
                    />
                    <el-input
                        type="password"
                        show-password
                        class="input"
                        v-model="confirmedPassword"
                        placeholder="再次确认密码"
                        :prefix-icon="Lock"
                        @keydown.enter="submitRegister"
                    />
                    <div class="submit" @click="submitRegister">注册</div>
                </div>
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script lang="ts">
import { ElMessage } from 'element-plus';
import axios from 'axios';
import { User, Lock, VideoPlay } from '@element-plus/icons-vue';

export default {
    name: "LoginRegister",
    data() {
        return {
            activeTab: "login",
            usernameLogin: "",
            passwordLogin: "",
            usernameRegister: "",
            passwordRegister: "",
            confirmedPassword: "",
            type: 1,    // 1登录 2注册
        }
    },
    mounted() {
        document.addEventListener('keydown', (e) => this.handleKeyboard(e));
    },
    beforeUnmount() {
        document.removeEventListener('keydown', (e) => this.handleKeyboard(e));
    },
    methods: {
        // 点击标签页触发的事件
        handleClick(tab) {
            if (tab.props.name === 'login' || tab.props.label === '登录') {
                this.type = 1;
            } else {
                this.type = 2;
            }
        },

        // 监听键盘回车触发登录
        handleKeyboard(event) {
            if (event.keyCode === 13 && this.type === 1) {
                this.submitLogin();
            }
        },

        // 登录的回调
        async submitLogin() {
            if (this.usernameLogin.trim() == "") {
                ElMessage.error("请输入账号");
                return;
            }
            if (this.passwordLogin == "") {
                ElMessage.error("请输入密码");
                return;
            }
            this.$store.state.isLoading = true;
            const result = await axios.post("/api/user/account/login", {
                username: this.usernameLogin.toString(),
                password: this.passwordLogin.toString(),
            }).catch(() => {
                ElMessage.error("特丽丽被玩坏了");
                this.$store.state.isLoading = false;
            });
            if (!result) {
                this.$store.state.isLoading = false;
                return;
            }
            if (result.data.code !== 200) {
                ElMessage.error(result.data.message);
                this.$store.state.isLoading = false;
            }
            if (result.data.code === 200) {
                localStorage.setItem("teri_token", result.data.data.token);
                this.$store.commit("updateUser", result.data.data.user);
                await this.$store.dispatch("getMsgUnread");
                await this.initIMServer();
                await this.getFavorites();
                await this.getLikeAndDisLikeComment();
                ElMessage.success(result.data.message);
                this.$store.commit("updateIsLogin", true);
                this.$emit("loginSuccess");
                this.$store.state.isLoading = false;
            }
        },

        async submitRegister() {
            if (this.usernameRegister.trim() == "") {
                ElMessage.error("账号不能为空");
                return;
            }
            if (this.passwordRegister == "" || this.confirmedPassword == "") {
                ElMessage.error("密码不能为空");
                return;
            }
            if (this.passwordRegister != this.confirmedPassword) {
                ElMessage.error("两次输入的密码不一致");
                return;
            }

            const result = await this.$post("/user/account/register", {
                username: this.usernameRegister.toString(),
                password: this.passwordRegister.toString(),
                confirmedPassword: this.confirmedPassword.toString(),
            });
            if (!result) return;
            if (result.data.code === 200) {
                ElMessage.success(result.data.message);
                this.usernameRegister = "";
                this.passwordRegister = "";
                this.confirmedPassword = "";
            }
        },

        // 开启实时通信消息服务
        async initIMServer() {
            await this.$store.dispatch("connectWebSocket");
            const connection = JSON.stringify({
                code: 100,
                content: "Bearer " + localStorage.getItem('teri_token'),
            });
            this.$store.state.ws.send(connection);
        },

        // 获取当前用户的收藏夹列表
        async getFavorites() {
            const res = await this.$get("/favorite/get-all/user", {
                params: { uid: this.$store.state.user.uid },
                headers: { Authorization: "Bearer " + localStorage.getItem("teri_token") }
            });
            if (!res.data.data) return;
            const defaultFav = res.data.data.find(item => item.type === 1);
            const list = res.data.data.filter(item => item.type !== 1);
            list.unshift(defaultFav);
            this.$store.commit("updateFavorites", list);
        },

        // 获取用户赞踩的评论集合
        async getLikeAndDisLikeComment() {
            const res = await this.$get("/comment/get-like-and-dislike", {
                params: { uid: this.$store.state.user.uid },
                headers: { Authorization: "Bearer " + localStorage.getItem("teri_token") }
            });
            if (!res.data) return;
            this.$store.commit("updateLikeComment", res.data.data.userLike);
            this.$store.commit("updateDislikeComment", res.data.data.userDislike);
        }
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
    margin-bottom: 12px;
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
    margin: 0 auto;
}

.login-tabs /deep/ .el-tabs__active-bar {
    background-color: #00a1d6;
}

.login-tabs /deep/ .el-tabs__nav-wrap::after {
    height: 1px;
    background-color: #eee;
}

.login-tabs /deep/ .el-tabs__item {
    font-size: 16px;
    color: #999;
}

.login-tabs /deep/ .el-tabs__item.is-active {
    color: #00a1d6;
}

.login-box, .register-box {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.login-box .input, .register-box .input {
    margin-top: 18px;
    width: 100%;
}

.login-box .input /deep/ .el-input__wrapper,
.register-box .input /deep/ .el-input__wrapper {
    border-radius: 4px;
    height: 40px;
    padding: 4px 12px;
}

.login-box .input /deep/ .el-input__inner,
.register-box .input /deep/ .el-input__inner {
    font-size: 14px;
    height: 28px;
    line-height: 28px;
}

.submit {
    color: #fff;
    width: 100%;
    margin-top: 24px;
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

.tips {
    margin-top: 12px;
    color: #999;
    font-size: 12px;
    text-align: center;
}

.tips .agreement {
    color: #00a1d6;
    margin-left: 4px;
    cursor: pointer;
}
</style>