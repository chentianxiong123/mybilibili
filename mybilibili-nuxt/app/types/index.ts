export interface User {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  signature: string
  gender: string
  level: number
  followerCount: number
  followingCount: number
  dynamicCount: number
  manuscriptCount: number
  coinCount: number
  totalViewCount: number
  totalLikeCount: number
  birthdate?: string
  announcement?: string
  experience?: number
  pinnedVideo?: Video | null
}

export interface Video {
  manuscriptId: number
  id: number
  title: string
  coverUrl: string
  description: string
  duration: number
  viewCount: number
  danmakuCount: number
  likeCount: number
  coinCount: number
  collectCount: number
  shareCount: number
  commentCount: number
  createTime: string
  updateTime: string
  status: string
  visibility: string
  userId: number
  username: string
  userAvatar: string
  categoryId: number
  categoryName: string
  tags: string[]
  videos: VideoPart[]
  uploader?: User
}

export interface VideoPart {
  id: number
  title: string
  playUrl: string
  playUrlHd: string
  playUrlSd: string
  playUrlLd: string
  duration: number
}

export interface Comment {
  id: number
  content: string
  createTime: string
  likeCount: number
  isTop: boolean
  userId: number
  userName: string
  userAvatar: string
  userLevel: number
  replies: Reply[]
  replyCount?: number
}

export interface Reply {
  id: number
  content: string
  createTime: string
  likeCount: number
  userId: number
  userName: string
  isUp: boolean
  replyToUserId?: number
  replyToUserName?: string
}

export interface Dynamic {
  id: number
  content: string
  images: string[]
  createTime: string
  likeCount: number
  commentCount: number
  shareCount: number
  userId: number
  userName: string
  userAvatar: string
  video?: Video
  isLiked?: boolean
}

export interface LiveRoom {
  id: number
  roomName: string
  coverUrl: string
  streamUrl: string
  status: string
  viewerCount: number
  anchorName: string
  description: string
  areaId: number
  areaName: string
  userId: number
  followerCount?: number
}

export interface Category {
  id: number
  name: string
  parentId?: number
}

export interface Banner {
  id: number
  title: string
  imageUrl: string
  linkUrl: string
}

export interface Collection {
  id: number
  title: string
  coverUrl: string
  description: string
  manuscriptCount: number
  createTime: string
}

export interface FavoriteFolder {
  id: number
  name: string
  manuscriptCount: number
  createTime: string
}

export interface SearchResult {
  aId: number
  title: string
  pic: string
  author: string
  play: number
  videoReview: number
  duration?: number
}

export interface DanmakuItem {
  text: string
  color: string
  time: number
}

export interface Message {
  id: number
  content: string
  createTime: string
  senderId: number
  senderName: string
  senderAvatar: string
  receiverId: number
  isRead: boolean
}

export interface Conversation {
  id: number
  lastMessage: string
  lastTime: string
  unreadCount: number
 对方: User
}

// API Response types
export interface ApiResponse<T = any> {
  code: number | string
  data: T
  message?: string
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  totalPages: number
}

// Admin types
export interface AdminUser {
  id: number
  username: string
  role: string
  permissions: string[]
  lastLoginTime: string
  status: string
}

export interface AuditLog {
  id: number
  action: string
  target: string
  operator: string
  time: string
  detail: string
}

export interface ProhibitedWord {
  id: number
  word: string
  type: string
  createTime: string
}

export interface SupportTicket {
  id: number
  title: string
  content: string
  status: string
  createTime: string
  userId: number
  userName: string
}

export interface AiSkill {
  id: number
  name: string
  description: string
  enabled: boolean
  config: Record<string, any>
}

export interface Channel {
  id: number
  name: string
  type: string
  enabled: boolean
  config: Record<string, any>
}

// SSE 回调类型
export interface SseCallbacks {
  onStart?: (data: any) => void
  onData?: (data: any) => void
  onDone?: (data: any) => void
  onError?: (error: any) => void
  onMeta?: (data: any) => void
  onTransfer?: (data: any) => void
  onToolCall?: (data: any) => void
}

export interface SseEvent {
  event?: string
  data?: string
}
