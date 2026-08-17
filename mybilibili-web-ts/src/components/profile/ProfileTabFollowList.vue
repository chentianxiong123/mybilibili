<script setup>
import { More, Check, Grid, Search } from '@element-plus/icons-vue'

const props = defineProps({
  followList: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['sidebar-click', 'follow-user'])

// 处理关注/粉丝筛选变化
const handleFollowFilterChange = (filterType) => {
  props.followList.filterType = filterType
  // TODO: 根据筛选类型重新加载数据
}

const handleFollowSearch = () => {
  console.log('搜索关键词:', props.followList.searchKeyword)
}
</script>

<template>
  <div class="follow-list-section">
    <div class="follow-list-container">
      <!-- 左侧边栏 -->
      <div class="follow-sidebar">
        <div
          :class="['sidebar-item', { active: followList.activeSidebar === 'following' }]"
          @click="emit('sidebar-click', 'following')"
        >
          <div class="sidebar-icon">👤</div>
          <div class="sidebar-info">
            <div class="sidebar-label">我的关注</div>
          </div>
          <div class="sidebar-count">{{ followList.followingList.length }}</div>
        </div>
        <div
          :class="['sidebar-item', { active: followList.activeSidebar === 'followers' }]"
          @click="emit('sidebar-click', 'followers')"
        >
          <div class="sidebar-icon">👥</div>
          <div class="sidebar-info">
            <div class="sidebar-label">我的粉丝</div>
          </div>
          <div class="sidebar-count">{{ followList.followersList.length }}</div>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="follow-content">
        <!-- 顶部标题和筛选 -->
        <div class="follow-header">
          <div class="follow-title">
            <h3>全部关注</h3>
            <span class="follow-count">{{ followList.followingList.length }}</span>
          </div>
          <div class="follow-filters">
            <div
              :class="['filter-item', { active: followList.filterType === 'frequent' }]"
              @click="handleFollowFilterChange('frequent')"
            >
              最常访问
            </div>
            <div
              :class="['filter-item', { active: followList.filterType === 'recent' }]"
              @click="handleFollowFilterChange('recent')"
            >
              最近关注
            </div>
            <div class="follow-search">
              <el-icon class="search-icon"><Search /></el-icon>
              <input
                type="text"
                placeholder="搜索"
                class="search-input"
                v-model="followList.searchKeyword"
                @keyup.enter="handleFollowSearch"
              >
            </div>
            <div class="view-toggle">
              <button class="view-btn">
                <el-icon><Grid /></el-icon>
              </button>
            </div>
          </div>
        </div>

        <!-- 用户列表 -->
        <div v-if="followList.loading" class="loading-state">
          <p>加载中...</p>
        </div>
        <div v-else-if="followList.followingList.length === 0" class="empty-state">
          <p>暂无关注用户</p>
        </div>
        <div v-else class="user-list">
          <div v-for="user in followList.followingList" :key="user.id" class="user-list-item">
            <div class="user-avatar-wrapper">
              <img loading="lazy" decoding="async" :src="user.avatar" :alt="user.nickname" class="user-list-avatar">
            </div>
            <div class="user-info">
              <div class="user-nickname">{{ user.nickname }}</div>
              <div class="user-signature">{{ user.signature }}</div>
            </div>
            <div class="user-actions">
              <button class="follow-btn following" @click="emit('follow-user', user.id, true)">
                <el-icon><Check /></el-icon>
                <span>已关注</span>
              </button>
              <button class="more-btn">
                <el-icon><More /></el-icon>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ==================== 关注/粉丝列表样式 ==================== */

/* 关注/粉丝列表页面整体布局 */
.follow-list-section {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.follow-list-container {
  display: flex;
  min-height: 600px;
}

/* 左侧边栏 */
.follow-sidebar {
  width: 200px;
  min-width: 200px;
  background-color: #f5f7fa;
  padding: 20px 0;
  border-right: 1px solid #e0e0e0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.sidebar-item:hover {
  background-color: rgba(0, 174, 236, 0.1);
}

.sidebar-item.active {
  background-color: #00aeec;
  color: #fff;
}

.sidebar-item.active .sidebar-count {
  color: rgba(255, 255, 255, 0.9);
}

.sidebar-icon {
  font-size: 20px;
  margin-right: 12px;
  width: 24px;
  text-align: center;
}

.sidebar-info {
  flex: 1;
}

.sidebar-label {
  font-size: 14px;
  font-weight: 500;
}

.sidebar-count {
  font-size: 13px;
  color: #999;
  margin-left: 8px;
}

/* 右侧内容区 */
.follow-content {
  flex: 1;
  padding: 20px;
  background-color: #fff;
}

/* 顶部标题和筛选区 */
.follow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 20px;
}

.follow-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.follow-title h3 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.follow-count {
  font-size: 14px;
  color: #999;
}

.follow-filters {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filter-item {
  font-size: 14px;
  color: #666;
  cursor: pointer;
  padding: 4px 0;
  border-bottom: 2px solid transparent;
  transition: all 0.3s ease;
}

.filter-item:hover {
  color: #00aeec;
}

.filter-item.active {
  color: #00aeec;
  border-bottom-color: #00aeec;
}

/* 搜索框 */
.follow-search {
  display: flex;
  align-items: center;
  background-color: #f5f7fa;
  border-radius: 16px;
  padding: 0 12px;
  height: 32px;
  width: 140px;
}

.follow-search .search-icon {
  color: #999;
  font-size: 14px;
  margin-right: 8px;
}

.follow-search .search-input {
  flex: 1;
  border: none;
  outline: none;
  background-color: transparent;
  font-size: 13px;
  color: #333;
}

.follow-search .search-input::placeholder {
  color: #999;
}

/* 视图切换按钮 */
.view-toggle {
  display: flex;
  gap: 8px;
}

.view-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.3s ease;
}

.view-btn:hover {
  border-color: #00aeec;
  color: #00aeec;
}

/* 用户列表 */
.user-list {
  display: flex;
  flex-direction: column;
}

.user-list-item {
  display: flex;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f5f5f5;
  transition: background-color 0.3s ease;
}

.user-list-item:hover {
  background-color: #fafafa;
}

.user-list-item:last-child {
  border-bottom: none;
}

/* 用户头像 */
.user-avatar-wrapper {
  margin-right: 16px;
}

.user-list-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #f0f0f0;
}

/* 用户信息 */
.user-info {
  flex: 1;
  min-width: 0;
}

.user-nickname {
  font-size: 15px;
  font-weight: 500;
  color: #333;
  margin-bottom: 6px;
  cursor: pointer;
  transition: color 0.3s ease;
}

.user-nickname:hover {
  color: #00aeec;
}

.user-signature {
  font-size: 13px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 用户操作按钮 */
.user-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.follow-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #e0e0e0;
}

.follow-btn.following {
  background-color: #fff;
  color: #666;
  border-color: #e0e0e0;
}

.follow-btn.following:hover {
  background-color: #f5f5f5;
  border-color: #ccc;
}

.follow-btn.not-following {
  background-color: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

.follow-btn.not-following:hover {
  background-color: #0095d9;
  border-color: #0095d9;
}

.follow-btn .el-icon {
  font-size: 14px;
}

.more-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #e0e0e0;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.3s ease;
}

.more-btn:hover {
  border-color: #00aeec;
  color: #00aeec;
}

/* 加载状态和空状态 */
.loading-state,
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
  color: #9499a0;
  font-size: 14px;
}

/* 关注/粉丝列表响应式设计 */
@media (max-width: 992px) {
  .follow-list-container {
    flex-direction: column;
  }

  .follow-sidebar {
    width: 100%;
    min-width: auto;
    display: flex;
    padding: 10px;
    border-right: none;
    border-bottom: 1px solid #e0e0e0;
  }

  .sidebar-item {
    flex: 1;
    justify-content: center;
  }

  .follow-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .follow-filters {
    width: 100%;
    flex-wrap: wrap;
  }
}

@media (max-width: 576px) {
  .user-list-item {
    padding: 12px 0;
  }

  .user-list-avatar {
    width: 48px;
    height: 48px;
  }

  .follow-search {
    width: 120px;
  }
}
</style>