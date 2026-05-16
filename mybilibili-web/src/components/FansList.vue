<script setup>
import { computed } from 'vue'
import { Check, More } from '@element-plus/icons-vue'

const props = defineProps({
  users: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  title: { type: String, default: '全部粉丝' },
  showSearch: { type: Boolean, default: false }
})

const emit = defineEmits(['follow', 'unfollow'])

const handleFollow = (user) => {
  if (user.isFollowing) {
    emit('unfollow', user.id)
  } else {
    emit('follow', user.id)
  }
}
</script>

<template>
  <div class="fans-list-wrapper">
    <div v-if="title" class="fans-list-header">
      <h3>{{ title }}</h3>
      <span class="fans-count">{{ users.length }}</span>
    </div>

    <div v-if="loading" class="fans-loading">
      <p>加载中...</p>
    </div>
    <div v-else-if="users.length === 0" class="fans-empty">
      <p>暂无粉丝</p>
    </div>
    <div v-else class="fans-user-list">
      <div v-for="user in users" :key="user.id" class="fans-user-item">
        <div class="fans-user-avatar">
          <img :src="user.avatar || 'https://via.placeholder.com/40'" :alt="user.nickname || user.name" />
        </div>
        <div class="fans-user-info">
          <div class="fans-user-name">{{ user.nickname || user.name || '未知用户' }}</div>
          <div class="fans-user-signature">{{ user.signature || user.bio || '' }}</div>
        </div>
        <div class="fans-user-actions">
          <button
            :class="['follow-btn', user.isFollowing ? 'following' : 'not-following']"
            @click="handleFollow(user)"
          >
            <el-icon v-if="user.isFollowing"><Check /></el-icon>
            <span>{{ user.isFollowing ? '已关注' : '关注' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fans-list-wrapper {
  flex: 1;
  min-width: 0;
}

.fans-list-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e3e5e7;
}

.fans-list-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin: 0;
}

.fans-count {
  font-size: 14px;
  color: #9499a0;
}

.fans-loading,
.fans-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #9499a0;
  font-size: 14px;
}

.fans-user-list {
  display: flex;
  flex-direction: column;
}

.fans-user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 0;
  border-bottom: 1px solid #f1f2f3;
}

.fans-user-item:last-child {
  border-bottom: none;
}

.fans-user-avatar {
  flex-shrink: 0;
}

.fans-user-avatar img {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
}

.fans-user-info {
  flex: 1;
  min-width: 0;
}

.fans-user-name {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fans-user-signature {
  font-size: 12px;
  color: #9499a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fans-user-actions {
  flex-shrink: 0;
}

.follow-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.follow-btn.not-following {
  background-color: #00aeec;
  color: #fff;
}

.follow-btn.not-following:hover {
  background-color: #0095d0;
}

.follow-btn.following {
  background-color: #f1f2f3;
  color: #9499a0;
}

.follow-btn.following:hover {
  background-color: #e3e5e7;
  color: #61666d;
}
</style>
