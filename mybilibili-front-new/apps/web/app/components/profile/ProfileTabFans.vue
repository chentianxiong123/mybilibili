<script setup>
import { Search } from '@element-plus/icons-vue'
import FansList from '@/components/FansList.vue'

defineProps({
  followList: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['sidebar-click', 'follow-user'])
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
        <FansList
          :users="followList.followersList"
          :loading="followList.loading"
          title="全部粉丝"
          :show-search="true"
          @follow="(id) => emit('follow-user', id, false)"
          @unfollow="(id) => emit('follow-user', id, true)"
        />
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

/* 关注/粉丝列表响应式设计 */

</style>