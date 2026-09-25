<template>
  <div class="dashboard-content">
    <p>视频数据</p>

    <!-- Loading状态 -->
    <div v-if="homeLoading" class="loading-container" v-loading="homeLoading" element-loading-text="加载中..."></div>

    <!-- 错误提示 -->
    <el-alert v-else-if="homeError" :title="homeError" type="error" show-icon :closable="false" style="margin-bottom: 20px;" />

    <template v-else>
    <!-- 第一行统计数据：粉丝总数、播放量、评论、弹幕 -->
    <div class="dashboard-stats">
      <div class="stat-card">
        <div class="stat-number">{{ statsData.followerCount }}</div>
        <div class="stat-label">粉丝总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalViewCount }}</div>
        <div class="stat-label">总播放量</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalCommentCount }}</div>
        <div class="stat-label">总评论数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalDanmuCount }}</div>
        <div class="stat-label">总弹幕数</div>
      </div>
    </div>

    <!-- 第二行统计数据：点赞、分享、收藏、投币 -->
    <div class="dashboard-stats">
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalLikeCount }}</div>
        <div class="stat-label">总点赞数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalShareCount }}</div>
        <div class="stat-label">总分享数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalFavoriteCount }}</div>
        <div class="stat-label">总收藏数</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ statsData.totalCoinCount }}</div>
        <div class="stat-label">总投币数</div>
      </div>
    </div>
    </template>

    <!-- 评论/弹幕选择栏 -->
    <div class="comment-danmu-section">
      <div class="section-header">
        <div class="tab-buttons">
          <el-button
            type="primary"
            :plain="activeCommentTab !== 'comment'"
            @click="activeCommentTab = 'comment'"
            size="small"
          >
            评论
          </el-button>
          <el-button
            type="primary"
            :plain="activeCommentTab !== 'danmu'"
            @click="activeCommentTab = 'danmu'"
            size="small"
          >
            弹幕
          </el-button>
        </div>
      </div>

      <!-- 评论列表 -->
      <div v-if="activeCommentTab === 'comment'" class="interaction-list">
        <div v-for="comment in paginatedComments" :key="comment.id" class="interaction-item">
          <el-avatar :size="28" :src="comment.avatar" class="item-avatar"></el-avatar>
          <span class="item-username">{{ comment.username }}</span>
          <span class="item-content">{{ comment.content }}</span>
          <el-link
            v-if="comment.manuscriptId"
            type="primary"
            underline="never"
            @click="router.push(`/manuscript/${comment.manuscriptId}`)"
            class="item-link"
          >
            {{ comment.manuscriptTitle || '查看视频' }}
          </el-link>
          <span class="item-time">{{ comment.createTime ? formatDate(comment.createTime) : comment.time }}</span>
          <el-button type="danger" size="small" link @click="handleDeleteComment(comment)" class="item-delete">删除</el-button>
        </div>
        <el-pagination
          v-if="latestComments.length > homeCommentPageSize"
          v-model:current-page="homeCommentPage"
          :page-size="homeCommentPageSize"
          :total="latestComments.length"
          layout="prev, pager, next"
          small
        />
      </div>

      <!-- 弹幕列表 -->
      <div v-else-if="activeCommentTab === 'danmu'" class="interaction-list" v-loading="danmakuLoading">
        <div v-if="danmakuList.length === 0" class="developing-tip">
          <el-empty description="暂无弹幕数据" :image-size="100" />
        </div>
        <div v-else class="danmaku-list">
          <div v-for="item in danmakuList" :key="item.id" class="danmaku-card">
            <span class="danmaku-text">{{ item.content }}</span>
            <el-link
              type="primary"
              underline="never"
              @click="goToVideo(item.manuscriptId, item.time, item.videoOrder)"
              class="item-link"
            >
              {{ item.videoName || '未知视频' }}
            </el-link>
            <el-tag size="small" :type="getDanmakuModeType(item.mode)" effect="plain">
              {{ formatTime(item.time) }}
            </el-tag>
            <el-tag size="small" type="info" effect="plain">
              {{ getDanmakuModeText(item.mode) }}
            </el-tag>
            <span class="item-time">{{ formatDate(item.createTime) }}</span>
            <el-button type="danger" size="small" link @click="deleteDanmaku(item.id)" class="item-delete">删除</el-button>
          </div>
        </div>
        <el-pagination
          v-if="danmakuTotal > danmakuPageSize"
          v-model:current-page="danmakuCurrentPage"
          :page-size="danmakuPageSize"
          :total="danmakuTotal"
          layout="prev, pager, next"
          small
          @current-change="fetchDanmakuList"
        />
      </div>
    </div>

    <!-- 观看排行和互动排行选择栏 -->
    <div class="ranking-section">
      <div class="section-header">
        <div class="tab-buttons">
          <el-button
            type="primary"
            :plain="activeRankingTab !== 'view'"
            @click="activeRankingTab = 'view'"
            size="small"
          >
            观看排行
          </el-button>
          <el-button
            type="primary"
            :plain="activeRankingTab !== 'interaction'"
            @click="activeRankingTab = 'interaction'"
            size="small"
          >
            互动排行
          </el-button>
        </div>
      </div>

      <!-- 观看排行列表 -->
      <div v-if="activeRankingTab === 'view'" class="ranking-list-horizontal">
        <div v-for="(user, index) in viewRanking" :key="user.id" class="ranking-item-horizontal">
          <div class="user-info">
            <el-avatar :size="32" :src="user.avatar" :class="getRankingClass(index)"></el-avatar>
            <span class="username" :class="getRankingClass(index)">{{ user.username }}</span>
          </div>
          <div class="ranking-value">{{ user.interactionCount || 0 }} 次评论</div>
        </div>
      </div>

      <!-- 互动排行列表 -->
      <div v-else-if="activeRankingTab === 'interaction'" class="ranking-list-horizontal">
        <div v-for="(user, index) in interactionRanking" :key="user.id" class="ranking-item-horizontal">
          <div class="user-info">
            <el-avatar :size="32" :src="user.avatar" :class="getRankingClass(index)"></el-avatar>
            <span class="username" :class="getRankingClass(index)">{{ user.username }}</span>
          </div>
          <div class="ranking-value">{{ user.interactionCount || 0 }} 次互动</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi, statsApi, manuscriptApi } from '@/api/creator'
import { commentApi } from '@/api/client'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()

// 首页统计数据
const homeLoading = ref(false)
const homeError = ref(null)
const statsData = ref({
  followerCount: 0,
  totalViewCount: 0,
  totalCommentCount: 0,
  totalDanmuCount: 0,
  totalLikeCount: 0,
  totalShareCount: 0,
  totalFavoriteCount: 0,
  totalCoinCount: 0
})

// 最新评论数据
const latestComments = ref([])
const homeCommentPage = ref(1)
const homeCommentPageSize = 5
const paginatedComments = computed(() => {
  const start = (homeCommentPage.value - 1) * homeCommentPageSize
  return latestComments.value.slice(start, start + homeCommentPageSize)
})

// 最新弹幕数据（开发中）
const latestDanmus = ref([])
const danmuDeveloping = ref(true)

// 观看排行数据
const viewRanking = ref([])

const interactionRanking = ref([])

// 评论/弹幕切换标签
const activeCommentTab = ref('comment')

// 排行切换标签
const activeRankingTab = ref('view')

// 评论管理相关数据（首页使用）
const commentRawList = ref([])
const comments = ref([])
const videoList = ref([])

const loadHomeData = async () => {
  homeLoading.value = true
  homeError.value = null
  try {
    const [overviewRes, commentsRes, viewRankingRes, interactionRankingRes] = await Promise.all([
      statsApi.getOverview(),
      statsApi.getLatestComments(5),
      statsApi.getFansRanking('view', 5),
      statsApi.getFansRanking('interaction', 5)
    ])

    if (overviewRes?.code === 200 && overviewRes.data) {
      statsData.value = {
        followerCount: overviewRes.data.totalFollowers || 0,
        totalViewCount: overviewRes.data.totalViews || 0,
        totalCommentCount: overviewRes.data.totalComments || 0,
        totalDanmuCount: overviewRes.data.totalDanmaku || 0,
        totalLikeCount: overviewRes.data.totalLikes || 0,
        totalShareCount: overviewRes.data.totalShares || 0,
        totalFavoriteCount: overviewRes.data.totalCollections || 0,
        totalCoinCount: overviewRes.data.totalCoins || 0
      }
    }

    if (commentsRes?.code === 200 && commentsRes.data) {
      latestComments.value = commentsRes.data
    }

    if (viewRankingRes?.code === 200 && viewRankingRes.data) {
      viewRanking.value = viewRankingRes.data
    }

    if (interactionRankingRes?.code === 200 && interactionRankingRes.data) {
      interactionRanking.value = interactionRankingRes.data
    }
  } catch (error) {
    console.error('加载主页数据失败:', error)
    homeError.value = '加载数据失败，请稍后重试'
  } finally {
    homeLoading.value = false
  }
}

// 获取排名样式类
const getRankingClass = (index) => {
  if (index === 0) {
    return 'ranking-gold'
  } else if (index === 1) {
    return 'ranking-silver'
  } else if (index === 2) {
    return 'ranking-bronze'
  }
  return ''
}

const danmakuList = ref([])
const danmakuLoading = ref(false)
const danmakuTotal = ref(0)
const danmakuCurrentPage = ref(1)
const danmakuPageSize = ref(10)

const fetchDanmakuList = async () => {
  danmakuLoading.value = true
  try {
    const res = await creatorApi.getDanmakuList({
      page: danmakuCurrentPage.value,
      size: danmakuPageSize.value
    })
    console.log('弹幕API返回:', res)
    if (res.code === 200 && res.data) {
      danmakuList.value = res.data.list || []
      danmakuTotal.value = res.data.total || 0
      console.log('弹幕列表:', danmakuList.value, '总数:', danmakuTotal.value)
    }
  } catch (error) {
    console.error('获取弹幕列表失败:', error)
  } finally {
    danmakuLoading.value = false
  }
}

const deleteDanmaku = async (danmakuId) => {
  try {
    await ElMessageBox.confirm('确定要删除这条弹幕吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await creatorApi.deleteDanmaku(danmakuId)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchDanmakuList()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除弹幕失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleDeleteComment = async (comment) => {
  const isReply = comment.commentType === 'reply'
  const label = isReply ? '回复' : '评论'
  try {
    await ElMessageBox.confirm(`确定要删除这条${label}吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = isReply
      ? await creatorApi.deleteReply(comment.id)
      : await creatorApi.deleteComment(comment.id)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchComments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error(`删除${label}失败:`, error)
      ElMessage.error('删除失败')
    }
  }
}

// 点赞/取消点赞评论
const handleLikeComment = async (comment) => {
  const isReply = comment.commentType === 'reply'
  try {
    if (comment.liked) {
      const res = isReply
        ? await commentApi.unlikeReply(comment.id)
        : await commentApi.unlikeComment(comment.id)
      if (res.code === 200) {
        comment.likeCount = Math.max(0, (comment.likeCount || 0) - 1)
        comment.liked = false
      } else {
        ElMessage.error(res.message || '取消点赞失败')
      }
    } else {
      const res = isReply
        ? await commentApi.likeReply(comment.id)
        : await commentApi.likeComment(comment.id)
      if (res.code === 200) {
        comment.likeCount = (comment.likeCount || 0) + 1
        comment.liked = true
      } else {
        ElMessage.error(res.message || '点赞失败')
      }
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    ElMessage.error('操作失败')
  }
}

const getDanmakuModeText = (mode) => {
  const modeMap = {
    1: '滚动',
    4: '底部',
    5: '顶部',
    6: '逆向',
    7: '定位'
  }
  return modeMap[mode] || '滚动'
}

const getDanmakuModeType = (mode) => {
  const typeMap = {
    1: 'primary',
    4: 'warning',
    5: 'success',
    6: 'danger',
    7: 'info'
  }
  return typeMap[mode] || 'primary'
}

const formatTime = (seconds) => {
  if (!seconds && seconds !== 0) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const goToVideo = (manuscriptId, time, videoOrder) => {
  if (!manuscriptId) return
  const pParam = videoOrder ? `&p=${videoOrder}` : '&p=1'
  const timeParam = time ? `&t=${Math.floor(time)}` : ''
  window.open(`/manuscript/${manuscriptId}?${pParam.substring(1)}${timeParam}`, '_blank')
}

const fetchComments = async () => {
  try {
    const params = {
      page: 1,
      size: 1000,
      manuscriptId: undefined,
      sort: 'latest',
      commentType: 'all'
    }
    const res = await creatorApi.getComments(params)
    if (res.code === 200 && res.data) {
      commentRawList.value = res.data.list.map(item => ({
        id: item.id,
        selected: false,
        username: item.userName || '未知用户',
        avatar: item.userAvatar || '',
        content: item.content,
        time: item.createTime,
        videoThumbnail: item.manuscriptCover || '',
        videoTitle: item.manuscriptTitle || '',
        likeCount: item.likeCount || 0,
        replyCount: item.replyCount || 0,
        liked: item.liked || false,
        userId: item.userId,
        manuscriptId: item.manuscriptId,
        commentType: item.commentType || 'comment',
        parentCommentId: item.parentCommentId,
        replyToUserName: item.replyToUserName
      }))
      comments.value = commentRawList.value
    }
  } catch (error) {
    console.error('获取评论列表失败:', error)
  }
}

const fetchVideoList = async () => {
  try {
    const res = await manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    if (res.code === 200 && res.data) {
      videoList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取视频列表失败:', error)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 监听评论/弹幕切换标签
watch(activeCommentTab, (newVal) => {
  if (newVal === 'comment') {
    fetchComments()
    fetchVideoList()
  } else if (newVal === 'danmu') {
    fetchDanmakuList()
  }
})

// 监听当前激活菜单变化，加载首页数据
watch(
  () => props.active,
  (newVal) => {
    if (newVal) {
      loadHomeData()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.dashboard-content {
  width: 100%;
}

.loading-container {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.developing-tip {
  padding: 40px 20px;
  text-align: center;
  color: #909399;
}

/* 仪表盘统计卡片样式 */
.dashboard-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.stat-card {
  background-color: #fafafa;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: #1890ff;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

/* 评论/弹幕选择栏样式 */
.comment-danmu-section {
  margin-top: 30px;
  padding: 0;
  background-color: transparent;
  box-shadow: none;
}

/* 排行选择栏样式 */
.ranking-section {
  margin-top: 30px;
  background-color: transparent;
  padding: 0;
  border-radius: 0;
  box-shadow: none;
}

.section-header {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.section-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.tab-buttons {
  display: flex;
  gap: 10px;
}

/* 互动列表样式 - 垂直布局 */
.interaction-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.interaction-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s ease;
}

.interaction-item:last-child {
  border-bottom: none;
}

.interaction-item:hover {
  background-color: #f5f7fa;
}

.interaction-item .item-avatar {
  flex-shrink: 0;
}

.interaction-item .item-username {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.interaction-item .item-content {
  flex: 1;
  font-size: 13px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.interaction-item .item-link {
  flex-shrink: 0;
  font-size: 12px;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.interaction-item .item-time {
  flex-shrink: 0;
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.interaction-item .item-delete {
  flex-shrink: 0;
}

/* 互动列表样式 - 水平布局 */
.interaction-list-horizontal {
  display: flex;
  gap: 20px;
  margin-top: 10px;
}

.interaction-item-horizontal {
  flex: 1;
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
  min-height: 120px;
  display: flex;
  flex-direction: column;
}

.interaction-item-horizontal:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 用户信息样式 */
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.username {
  font-weight: 500;
  color: #303133;
}

/* 评论弹幕按钮样式 */
.tab-buttons .el-button {
  border: none;
  box-shadow: none;
}

.tab-buttons .el-button--primary:not(.is-plain) {
  background-color: #1890ff;
  border-color: transparent;
}

.tab-buttons .el-button--primary.is-plain {
  background-color: transparent;
  border-color: transparent;
  color: #1890ff;
}

/* 排行列表样式 */
.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.ranking-item:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.ranking-index {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: #1890ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
}

.ranking-count {
  margin-left: auto;
  font-weight: 600;
  color: #f77825;
}

/* 排行列表样式 - 水平布局 */
.ranking-list-horizontal {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 15px;
  margin-top: 10px;
}

.ranking-item-horizontal {
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  min-height: 120px;
}

.ranking-item-horizontal:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 水平排行项中的排名索引 */
.ranking-item-horizontal .ranking-index {
  margin-bottom: 10px;
}

/* 水平排行项中的用户信息 */
.ranking-item-horizontal .user-info {
  flex-direction: column;
  gap: 5px;
  margin-bottom: 6px;
}

/* 水平排行项中的用户名称 */
.ranking-item-horizontal .username {
  font-size: 14px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 水平排行项中的数值 */
.ranking-value {
  font-size: 12px;
  color: #909399;
  margin-top: auto;
}

/* 排名样式 - 金色（第一名） */
.ranking-gold {
  border-color: #f7c13b;
  color: #f7c13b;
}

.ranking-gold + .username {
  color: #f7c13b;
  font-weight: bold;
}

/* 排名样式 - 银色（第二名） */
.ranking-silver {
  border-color: #c0c4cc;
  color: #c0c4cc;
}

.ranking-silver + .username {
  color: #c0c4cc;
  font-weight: bold;
}

/* 排名样式 - 铜色（第三名） */
.ranking-bronze {
  border-color: #e8a055;
  color: #e8a055;
}

.ranking-bronze + .username {
  color: #e8a055;
  font-weight: bold;
}

/* 头像边框样式 */
.el-avatar.ranking-gold {
  border: 2px solid #f7c13b;
  box-shadow: 0 0 8px rgba(247, 193, 59, 0.4);
}

.el-avatar.ranking-silver {
  border: 2px solid #c0c4cc;
  box-shadow: 0 0 8px rgba(192, 196, 204, 0.4);
}

.el-avatar.ranking-bronze {
  border: 2px solid #e8a055;
  box-shadow: 0 0 8px rgba(232, 160, 85, 0.4);
}

/* 水平排行项中的计数 */
.ranking-item-horizontal .ranking-count {
  margin: auto 0 0 0;
}

/* 弹幕卡片列表样式 */
.danmaku-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 0;
}

/* 弹幕列表 */
.danmaku-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* 弹幕卡片样式 */
.danmaku-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s ease;
}

.danmaku-card:last-child {
  border-bottom: none;
}

.danmaku-card:hover {
  background-color: #f5f7fa;
}

.danmaku-text {
  flex: 1;
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

/* 弹幕空状态 */
.developing-tip {
  padding: 60px 0;
  text-align: center;
}

/* 评论/弹幕分页样式 */
.interaction-list .el-pagination {
  margin-top: 12px;
  justify-content: center;
}
</style>