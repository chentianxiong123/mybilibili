import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import api from '@/api/client'
import { clearAuthSession } from '@/utils/auth'
import { clearServerSession } from '@/api/session'

export const useTeriteriStore = defineStore('teriteri', {
  state: () => ({
    isLoading: false,
    isLogin: false,
    openLogin: false,
    user: {},
    channels: [],
    carousels: [],
    danmuList: [],
    msgUnread: [0, 0, 0, 0, 0, 0],
    chatList: [],
    chatId: -1,
    isChatPage: false,
    ws: null,
    attitudeToVideo: {},
    likeComment: [],
    dislikeComment: [],
    favorites: [],
    userFavList: [],
    trendings: [],
    matchingCount: [0, 0]
  }),
  actions: {
    initData() {
      this.isLogin = false
      this.user = {}
      this.msgUnread = [0, 0, 0, 0, 0, 0]
      this.attitudeToVideo = {}
      this.favorites = []
      this.likeComment = []
      this.dislikeComment = []
    },
    updateIsLogin(v) { this.isLogin = v },
    updateUser(u) { this.user = u },
    updateChannels(c) { this.channels = c },
    updateCarousels(c) { this.carousels = c },
    updateDanmuList(d) { this.danmuList = d },
    updateChatList(c) { this.chatList.push(...c) },
    updateAttitudeToVideo(a) { this.attitudeToVideo = a },
    updateLikeComment(l) { this.likeComment = l },
    updateDislikeComment(d) { this.dislikeComment = d },
    updateFavorites(f) { this.favorites = f },
    updateUserFavList(f) { this.userFavList = f || [] },
    updateTrendings(t) { this.trendings = t },
    updateMatchingCount(m) { this.matchingCount = m },
    setWebSocket(ws) { this.ws = ws },

    handleWsOpen() {
      // ws open
    },

    handleWsClose() {
      console.log('实时通信websocket关闭,请登录并刷新页面重试')
      this.isLogin = false
      this.user = {}
      this.msgUnread = [0, 0, 0, 0, 0, 0]
      this.attitudeToVideo = {}
      this.favorites = []
      this.likeComment = []
      this.dislikeComment = []
    },

    handleWsMessage(e) {
      const data = JSON.parse(e.data)
      switch (data.type) {
        case 'error': {
          if (data.data === '登录已过期') {
            this.initData()
            clearAuthSession()
          }
          ElMessage.error(data.data)
          break
        }
        case 'unread_init':
        case 'unread_counts': {
          const d = data.data
          if (d) {
            this.msgUnread[0] = d.reply || 0
            this.msgUnread[1] = d.at || 0
            this.msgUnread[2] = d.like || 0
            this.msgUnread[3] = d.system || 0
            this.msgUnread[4] = d.private || 0
            this.msgUnread[5] = d.dynamic || 0
          }
          break
        }
        case 'reply': {
          const content = data.data
          if (content.type === '全部已读') this.msgUnread[0] = 0
          else if (content.type === '接收') this.msgUnread[0]++
          break
        }
        case 'at': {
          const content = data.data
          if (content.type === '全部已读') this.msgUnread[1] = 0
          else if (content.type === '接收') this.msgUnread[1]++
          break
        }
        case 'love': {
          const content = data.data
          if (content.type === '全部已读') this.msgUnread[2] = 0
          else if (content.type === '接收') this.msgUnread[2]++
          break
        }
        case 'system': {
          const content = data.data
          if (content.type === '全部已读') this.msgUnread[3] = 0
          else if (content.type === '接收') this.msgUnread[3]++
          break
        }
        case 'whisper': {
          const content = data.data
          switch (content.type) {
            case '全部已读':
              this.msgUnread[4] = 0
              this.chatList.forEach(item => { item.chat.unread = 0 })
              break
            case '已读': {
              const chatid = content.id
              const count = content.count
              this.msgUnread[4] = Math.max(0, this.msgUnread[4] - count)
              const chat = this.chatList.find(item => item.chat.id === chatid)
              if (chat) chat.chat.unread = 0
              break
            }
            case '移除': {
              const chatid = content.id
              const count = content.count
              this.msgUnread[4] = Math.max(0, this.msgUnread[4] - count)
              const i = this.chatList.findIndex(item => item.chat.id === chatid)
              if (i !== -1) {
                if (this.chatList[i].user.uid === this.chatId) this.chatId = -1
                this.chatList.splice(i, 1)
              }
              break
            }
            case '接收': {
              const chat = content.chat
              const detail = content.detail
              const user = content.user
              const sortByLatestTime = list => {
                list.sort((a, b) => {
                  const timeA = new Date(a.chat.latestTime).getTime()
                  const timeB = new Date(b.chat.latestTime).getTime()
                  return timeB - timeA
                })
              }
              if (detail.userId === this.user.uid) {
                const chatItem = this.chatList.find(item => item.chat.userId === detail.anotherId)
                if (chatItem && this.isChatPage) {
                  chatItem.detail.list.push(detail)
                  chatItem.chat.latestTime = chat.latestTime
                  sortByLatestTime(this.chatList)
                }
              } else {
                if (!content.online) this.msgUnread[4]++
                let chatItem = this.chatList.find(item => item.chat.userId === detail.userId)
                if (chatItem) {
                  chatItem.detail.list.push(detail)
                  chatItem.chat = chat
                  sortByLatestTime(this.chatList)
                } else {
                  chatItem = {
                    chat,
                    user,
                    detail: { more: true, list: [] }
                  }
                  chatItem.detail.list.push(detail)
                  this.chatList.unshift(chatItem)
                }
              }
              break
            }
            case '撤回': {
              const msgId = content.id
              const sendId = content.sendId
              const acceptId = content.acceptId
              let chat
              if (sendId === this.user.uid) chat = this.chatList.find(item => item.chat.userId === acceptId)
              else chat = this.chatList.find(item => item.chat.userId === sendId)
              if (chat) {
                const msg = chat.detail.list.find(item => item.id === msgId)
                if (msg) msg.withdraw = 1
              }
              break
            }
          }
          break
        }
        case 'dynamic': {
          const content = data.content
          if (content.type === '全部已读') this.msgUnread[5] = 0
          else if (content.type === '接收') this.msgUnread[5]++
          break
        }
      }
    },

    handleWsError(_, e) {
      console.log('实时通信websocket报错: ', e)
    },

    async getPersonalInfo() {
      // 走共享 axios 实例：自动带 Authorization，401 时自动续签并重试，
      // 否则 access token 过期会把人打回未登录（下拉面板显示空）。
      const result = await api.get('/user/me').catch(() => null)
      if (!result) return
      if (result.code === 401) {
        // client 层已清理失效会话，同步重置 teriteri 登录态
        this.initData()
        if (this.ws) {
          this.ws.close()
          this.setWebSocket(null)
        }
        return
      }
      if (result.code === 200) {
        const d = result.data
        this.updateUser({
          ...d,
          uid: d.id,
          avatar_url: d.avatar_url || d.avatar || '',
          followsCount: d.following_count || 0,
          fansCount: d.follower_count || 0,
          exp: d.experience || 0,
          coin: d.coin_count || 0,
          dynamicCount: d.dynamic_count || 0,
          gender: d.gender || 0,
          vip: d.vip || 0,
          auth: d.auth || 0,
        })
        this.isLogin = true
      }
    },

    logout() {
      this.initData()
      if (this.ws) {
        this.ws.close()
        this.setWebSocket(null)
      }
      void clearServerSession()
      clearAuthSession()
    },

    async connectWebSocket() {
      return new Promise((resolve) => {
        if (this.ws) {
          this.ws.close()
          this.setWebSocket(null)
        }
        const wsBaseUrl = (typeof window !== 'undefined' && window.__TERI_WS_URL__) || ''
        if (!wsBaseUrl) {
          // 未配置 IM WS 端点（环境变量 VUE_APP_WS_IM_URL 未设置），跳过实时连接
          resolve()
          return
        }
        const ws = new WebSocket(`${wsBaseUrl}/im`)
        this.setWebSocket(ws)
        ws.addEventListener('open', () => {
          this.handleWsOpen()
          resolve()
        })
        ws.addEventListener('close', () => this.handleWsClose())
        ws.addEventListener('message', e => this.handleWsMessage(e))
        ws.addEventListener('error', e => this.handleWsError({}, e))
      })
    },

    async closeWebSocket() {
      if (this.ws) {
        await this.ws.close()
        this.setWebSocket(null)
      }
    }
  }
})
