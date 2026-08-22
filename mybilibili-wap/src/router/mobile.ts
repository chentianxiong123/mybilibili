import type { RouteRecordRaw } from 'vue-router'

const mobileRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/m/index'
  },
  {
    path: '/m/index',
    name: 'mobile-home',
    component: () => import('../views/home/Index.vue'),
    meta: { title: '哔哩哔哩', mobile: true }
  },
  {
    path: '/m/login',
    name: 'mobile-login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录 - 哔哩哔哩', mobile: true }
  },
  {
    path: '/m/channel/:rId',
    name: 'mobile-channel',
    component: () => import('../views/Channel.vue'),
    meta: { title: '分区', mobile: true }
  },
  {
    path: '/m/ranking/:rId',
    name: 'mobile-ranking',
    component: () => import('../views/ranking/Ranking.vue'),
    meta: { title: '排行榜', mobile: true }
  },
  {
    path: '/m/video/:aId',
    name: 'mobile-video',
    component: () => import('../views/video/Detail.vue'),
    meta: { title: '视频详情', mobile: true }
  },
  {
    path: '/m/search',
    name: 'mobile-search',
    component: () => import('../views/search/Search.vue'),
    meta: { title: '搜索', mobile: true }
  },
  {
    path: '/m/search/hot',
    name: 'mobile-search-hot',
    component: () => import('../views/search/HotRank.vue'),
    meta: { title: 'bilibili热搜', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/search/result',
    name: 'mobile-search-result',
    component: () => import('../views/search/Result.vue'),
    meta: { title: '搜索结果', mobile: true }
  },
  {
    path: '/m/message',
    name: 'mobile-message',
    component: () => import('../views/message/Message.vue'),
    meta: { title: '消息', mobile: true }
  },
  {
    path: '/m/message/chat/:id',
    name: 'mobile-message-chat',
    component: () => import('../views/message/Chat.vue'),
    meta: { title: '私信', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/creator',
    name: 'mobile-creator-center',
    component: () => import('../views/creator/Center.vue'),
    meta: { title: '创作中心', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/space',
    name: 'mobile-space',
    component: () => import('../views/space/Space.vue'),
    meta: { title: '我的', mobile: true }
  },
  {
    path: '/m/space/history',
    name: 'mobile-space-history',
    component: () => import('../views/space/History.vue'),
    meta: { title: '历史记录', mobile: true }
  },
  {
    path: '/m/space/favorite',
    name: 'mobile-space-favorite',
    component: () => import('../views/space/Favorite.vue'),
    meta: { title: '我的收藏', mobile: true }
  },
  {
    path: '/m/space/manuscripts',
    name: 'mobile-space-manuscripts',
    component: () => import('../views/space/ManuscriptManage.vue'),
    meta: { title: '稿件管理', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/space/profile/edit',
    name: 'mobile-space-profile-edit',
    component: () => import('../views/space/ProfileEdit.vue'),
    meta: { title: '账号资料', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/space/:mId/friends',
    name: 'mobile-space-friends',
    component: () => import('../views/space/FriendList.vue'),
    meta: { title: '我的好友', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/space/:mId',
    name: 'mobile-space-user',
    component: () => import('../views/space/UpUser.vue'),
    meta: { title: 'UP主主页', mobile: true, hideTabBar: true }
  },
  {
    path: '/m/dynamic',
    name: 'mobile-dynamic',
    component: () => import('../views/dynamic/Index.vue'),
    meta: { title: '关注动态 - 哔哩哔哩', mobile: true }
  },
  {
    path: '/m/live',
    name: 'mobile-live',
    component: () => import('../views/live/Index.vue'),
    meta: { title: '直播', mobile: true }
  },
  {
    path: '/m/live/list',
    name: 'mobile-live-list',
    component: () => import('../views/live/List.vue'),
    meta: { title: '直播列表', mobile: true }
  },
  {
    path: '/m/live/areas',
    name: 'mobile-live-areas',
    component: () => import('../views/live/Area.vue'),
    meta: { title: '直播分类', mobile: true }
  },
  {
    path: '/m/live/:roomId',
    name: 'mobile-live-room',
    component: () => import('../views/live/Room.vue'),
    meta: { title: '直播间', mobile: true }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'mobile-not-found',
    component: () => import('../views/NotFound.vue'),
    meta: { title: '页面未找到 - 哔哩哔哩', mobile: true }
  }
]

export default mobileRoutes
