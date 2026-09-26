import { defineStore } from 'pinia'
import { userApi } from '@/api/client'
import { clearServerSession } from '@/api/session'
import {
  clearAuthSession,
  getCurrentUserId,
  getStoredUser,
  hasAuthSession,
  scrubCredentials,
  setAuthSession
} from '@/utils/auth'

// 用户store
export const useUserStore = (defineStore as any)('user', {
  // 状态
  state: () => ({
    // 用户信息
    userInfo: {
      id: '',
      username: '',
      nickname: '',
      avatar: '',
      email: '',
      phone: '',
      gender: 0, // 0: 未知, 1: 男, 2: 女
      birthdate: '',
      signature: '',
      level: 1,
      followingCount: 0,
      followerCount: 0,
      videoCount: 0,
      likedCount: 0,
      coinCount: 0,
      pointCount: 0
    },
    // 登录状态（凭证在 HttpOnly cookie 里，store 里不留任何可读副本）
    isLoggedIn: false,
    // 登录加载状态
    loginLoading: false,
    // 注册加载状态
    registerLoading: false
  }),

  // Getters
  getters: {
    // 获取用户ID
    getUserId: (state) => state.userInfo.id,
    // 获取用户名
    getUsername: (state) => state.userInfo.username,
    // 获取用户昵称
    getNickname: (state) => state.userInfo.nickname,
    // 获取用户头像
    getAvatar: (state) => state.userInfo.avatar,
    // 获取登录状态
    getIsLoggedIn: (state) => state.isLoggedIn,
    // 获取用户等级
    getUserLevel: (state) => state.userInfo.level,
    // 获取关注数
    getFollowingCount: (state) => state.userInfo.followingCount,
    // 获取粉丝数
    getFollowerCount: (state) => state.userInfo.followerCount,
    // 获取视频数
    getVideoCount: (state) => state.userInfo.videoCount
  },

  // Actions
  actions: {
    // 设置登录状态
    setLoginStatus(status) {
      this.isLoggedIn = status
    },

    // 设置用户信息
    // userInfo 会被 persist 到 localStorage，所以登录响应里的 token 必须在这里洗掉
    setUserInfo(userInfo) {
      this.userInfo = { ...this.userInfo, ...scrubCredentials(userInfo) }
    },

    // 设置令牌
    // 保留旧签名以兼容调用方；凭证不再落盘，这里只维护展示用的登录态
    setToken() {
      this.isLoggedIn = hasAuthSession()
    },

    // 清除令牌
    clearToken() {
      clearAuthSession()
    },

    // 从本地可见的登录态信号恢复（凭证本身在 HttpOnly cookie 里）
    loadTokenFromStorage() {
      if (!hasAuthSession()) return
      this.isLoggedIn = true
      const user = getStoredUser()
      if (user) {
        if (!user.id && user.user_id) {
          user.id = user.user_id
        }
        this.setUserInfo(user)
      }
    },

    // 登录操作
    async login(loginForm) {
      try {
        this.loginLoading = true
        const response = await userApi.login(loginForm.username, loginForm.password, null, null, null)
        // 成功与否只看 code：凭证在 HttpOnly cookie 里，body 默认不再回 token
        if (response.code !== 200 || !response.data) {
          return { success: false, message: response.message || '登录失败，请检查用户名和密码' }
        }

        // token/refresh_token 已由服务端写进 HttpOnly cookie，这里只留展示信息
        setAuthSession({ user: response.data })
        const userData = response.data.user || response.data
        if (!userData.id && userData.user_id) {
          userData.id = userData.user_id
        }
        this.setUserInfo(userData)
        this.setLoginStatus(true)
        return { success: true, message: '登录成功' }
      } catch (error) {
        console.error('登录失败:', error)
        return { success: false, message: '登录失败，请检查用户名和密码' }
      } finally {
        this.loginLoading = false
      }
    },

    // 注册操作
    async register(registerForm) {
      try {
        this.registerLoading = true
        const response = await userApi.register(registerForm)
        if (response.code !== 200) {
          return { success: false, message: response.message || '注册失败，请稍后重试' }
        }
        return { success: true, message: response.message || '注册成功，请登录' }
      } catch (error) {
        console.error('注册失败:', error)
        return { success: false, message: '注册失败，请稍后重试' }
      } finally {
        this.registerLoading = false
      }
    },

    // 退出登录
    logout() {
      // 清除用户信息
      this.setUserInfo({
        id: '',
        username: '',
        nickname: '',
        avatar: '',
        email: '',
        phone: '',
        gender: 0,
        birthdate: '',
        signature: '',
        level: 1,
        followingCount: 0,
        followerCount: 0,
        videoCount: 0,
        likedCount: 0,
        coinCount: 0,
        pointCount: 0
      })
      
      void clearServerSession()
      clearAuthSession()
      
      // 设置登录状态
      this.setLoginStatus(false)
    },

    // 获取用户信息
    async getUserInfo() {
      try {
        const userId = getCurrentUserId()
        if (!userId) return { success: false }
        const response = await userApi.getUserById(userId)
        if (response.code !== 200) {
          return { success: false }
        }
        this.setUserInfo(response.data)
        setAuthSession({ user: response.data })
        this.setLoginStatus(true)
        return { success: true }
      } catch (error) {
        console.error('获取用户信息失败:', error)
        return { success: false }
      }
    },

    // 更新用户信息
    async updateUserInfo(updateInfo) {
      try {
        const userId = getCurrentUserId()
        if (!userId) return { success: false, message: '请先登录' }
        const response = await userApi.updateUser(userId, updateInfo)
        if (response.code !== 200) {
          return { success: false, message: response.message || '更新失败，请稍后重试' }
        }
        this.setUserInfo(response.data)
        setAuthSession({ user: response.data })
        return { success: true, message: '用户信息更新成功' }
      } catch (error) {
        console.error('更新用户信息失败:', error)
        return { success: false, message: '更新失败，请稍后重试' }
      }
    },

    // 更新用户头像
    async updateAvatar(avatarUrl) {
      try {
        this.setUserInfo({ avatar: avatarUrl })
        const user = getStoredUser()
        if (user) {
          setAuthSession({ user: { ...user, avatar: avatarUrl } })
        }
        return { success: true, message: '头像更新成功' }
      } catch (error) {
        console.error('更新头像失败:', error)
        return { success: false, message: '头像更新失败，请稍后重试' }
      }
    }
  },

  // 持久化配置
  persist: {
    // 启用持久化
    enabled: true,
    // 持久化策略
    strategies: [
      {
        // 存储名称
        key: 'userStore',
        // 存储方式
        storage: localStorage,
        // 存储字段
        // 不持久化凭证：token / refreshToken 只存在于 HttpOnly cookie
        paths: ['userInfo', 'isLoggedIn']
      }
    ]
  }
})
