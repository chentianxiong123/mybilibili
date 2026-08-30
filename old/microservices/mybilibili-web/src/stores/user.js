import { defineStore } from 'pinia'
import { userApi } from '../api/index.js'
import {
  clearAuthSession,
  getCurrentUserId,
  getRefreshToken,
  getStoredUser,
  getToken,
  setAuthSession
} from '../utils/auth.js'

// 用户store
export const useUserStore = defineStore('user', {
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
    // 登录状态
    isLoggedIn: false,
    // 登录令牌
    token: '',
    // 刷新令牌
    refreshToken: '',
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
    setUserInfo(userInfo) {
      this.userInfo = { ...this.userInfo, ...userInfo }
    },

    // 设置令牌
    setToken(token, refreshToken) {
      this.token = token
      this.refreshToken = refreshToken
      setAuthSession({ token, refreshToken })
    },

    // 清除令牌
    clearToken() {
      this.token = ''
      this.refreshToken = ''
      clearAuthSession()
    },

    // 加载本地存储的令牌
    loadTokenFromStorage() {
      const token = getToken()
      const refreshToken = getRefreshToken()
      if (token && refreshToken) {
        this.token = token
        this.refreshToken = refreshToken
        this.isLoggedIn = true
        const user = getStoredUser()
        if (user) {
          this.setUserInfo(user)
        }
      }
    },

    // 登录操作
    async login(loginForm) {
      try {
        this.loginLoading = true
        const response = await userApi.login(loginForm.username, loginForm.password)
        if (response.code !== 200 || !response.data?.token) {
          return { success: false, message: response.message || '登录失败，请检查用户名和密码' }
        }

        setAuthSession(response.data)
        this.token = response.data.token
        this.refreshToken = response.data.refreshToken || ''
        if (response.data.user) {
          this.setUserInfo(response.data.user)
        }
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
      
      clearAuthSession()
      this.token = ''
      this.refreshToken = ''
      
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
        paths: ['userInfo', 'isLoggedIn', 'token', 'refreshToken']
      }
    ]
  }
})
