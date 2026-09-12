import { defineStore } from 'pinia'
import axios from 'axios'
import { ElMessage } from 'element-plus'

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
            if (typeof window !== 'undefined') localStorage.removeItem('teri_token')
          }
          ElMessage.error(data.data)
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
      const result = await axios.get('/api/user/personal/info', {
        headers: {
          Authorization: 'Bearer ' + (typeof window !== 'undefined' ? localStorage.getItem('teri_token') : '')
        }
      }).catch(() => {
        this.initData()
        if (this.ws) {
          this.ws.close()
          this.setWebSocket(null)
        }
        if (typeof window !== 'undefined') localStorage.removeItem('teri_token')
        ElMessage.error('请登录后查看')
      })
      if (!result) return
      if (result.data.code === 200) {
        this.updateUser(result.data.data)
        this.isLogin = true
      }
    },

    logout() {
      this.initData()
      if (this.ws) {
        this.ws.close()
        this.setWebSocket(null)
      }
      axios.get('/api/user/account/logout', {
        headers: {
          Authorization: 'Bearer ' + (typeof window !== 'undefined' ? localStorage.getItem('teri_token') : '')
        }
      }).catch(() => {})
      if (typeof window !== 'undefined') localStorage.removeItem('teri_token')
    },

    async getMsgUnread() {
      const { get } = await import('@/teriteri-src/network/request')
      const res = await get('/msg-unread/all', {
        headers: { Authorization: 'Bearer ' + localStorage.getItem('teri_token') }
      })
      const data = res.data.data
      this.msgUnread[0] = data.reply
      this.msgUnread[1] = data.at
      this.msgUnread[2] = data.love
      this.msgUnread[3] = data.system
      this.msgUnread[4] = data.whisper
      this.msgUnread[5] = data.dynamic
    },

    async connectWebSocket() {
      return new Promise((resolve) => {
        if (this.ws) {
          this.ws.close()
          this.setWebSocket(null)
        }
        const wsBaseUrl = (typeof window !== 'undefined' && window.__TERI_WS_URL__) || ''
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
