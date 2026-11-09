<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getMyInfo } from '../../api/user'

const router = useRouter()
const userInfo = ref<any>(null)
const isLoggedIn = ref(false)

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    // 未登录，跳转登录页
    router.replace('/m/login')
    return
  }
  isLoggedIn.value = true

  const res = await getMyInfo()
  if (res.code === '1' && res.data) {
    userInfo.value = res.data
  } else {
    // 降级使用本地存储
    const userStr = localStorage.getItem('user')
    if (userStr) {
      const localObj = JSON.parse(userStr)
      userInfo.value = localObj.user || localObj
    }
  }
})

// 退出登录
const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  router.push('/m/login')
}

// 导航处理
const navigateTo = (path: string) => {
  router.push(path)
}
</script>

<template>
  <div class="space-page">
    <!-- 顶部状态栏与操作图标 -->
    <div class="top-nav">
      <div class="left-icons"></div>
      <div class="right-icons">
        <div class="icon-btn" title="扫一扫">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <path d="M7 7h3v3H7zM14 7h3v3h-3zM7 14h3v3H7zM14 14h3v3h-3z" />
          </svg>
        </div>
        <div class="icon-btn" title="皮肤">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20.37 8.91l-8-5a1 1 0 0 0-1.07 0l-8 5A1 1 0 0 0 3 9.76v8.48a1 1 0 0 0 .53.88l8 5a1 1 0 0 0 1.07 0l8-5a1 1 0 0 0 .53-.88V9.76a1 1 0 0 0-.53-.85z" />
          </svg>
        </div>
        <div class="icon-btn" title="夜间模式">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
        </div>
      </div>
    </div>

    <!-- 用户基本信息 -->
    <div class="user-info-section" v-if="userInfo">
      <div class="avatar-row">
        <img :src="userInfo.avatar || '../../../src/assets/noface.gif'" class="avatar" />
        <div class="meta-info">
          <div class="name-row">
            <span class="nickname">{{ userInfo.nickname || userInfo.username }}</span>
            <span class="level-badge">LV{{ userInfo.level || 1 }}</span>
          </div>
          <div class="member-tag">正式会员</div>
          <div class="coins-info">
            <span class="coin-item">B币: 0.0</span>
            <span class="coin-item">硬币: {{ userInfo.coinCount || 0 }}</span>
          </div>
        </div>
        <div class="space-link" @click="navigateTo(`/m/space/${userInfo.id}`)">
          空间 <span class="arrow">&gt;</span>
        </div>
      </div>

      <!-- 动态/关注/粉丝 统计 -->
      <div class="stats-row">
        <div class="stat-item" @click="navigateTo(`/m/space/${userInfo.id}`)">
          <div class="num">{{ userInfo.dynamicCount || 0 }}</div>
          <div class="label">动态</div>
        </div>
        <div class="divider"></div>
        <div class="stat-item">
          <div class="num">{{ userInfo.followingCount || 0 }}</div>
          <div class="label">关注</div>
        </div>
        <div class="divider"></div>
        <div class="stat-item">
          <div class="num">{{ userInfo.followerCount || 0 }}</div>
          <div class="label">粉丝</div>
        </div>
      </div>
    </div>

    <!-- 常用功能入口（离线缓存、历史记录、我的收藏、稍后再看） -->
    <div class="common-tools-grid">
      <div class="tool-item">
        <div class="icon-wrap blue-light">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#23ade5" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
        </div>
        <div class="tool-label">离线缓存</div>
      </div>
      <div class="tool-item" @click="navigateTo('/m/space/history')">
        <div class="icon-wrap blue-light">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#23ade5" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <polyline points="12 6 12 12 16 14" />
          </svg>
        </div>
        <div class="tool-label">历史记录</div>
      </div>
      <div class="tool-item" @click="navigateTo('/m/space/favorite')">
        <div class="icon-wrap blue-light">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#23ade5" stroke-width="2">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
        </div>
        <div class="tool-label">我的收藏</div>
      </div>
      <div class="tool-item">
        <div class="icon-wrap blue-light">
          <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="#23ade5" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <polyline points="12 8 12 12 15 15" />
            <circle cx="12" cy="12" r="3" fill="#23ade5" />
          </svg>
        </div>
        <div class="tool-label">稍后再看</div>
      </div>
    </div>

    <!-- 创作中心 -->
    <div class="card-section">
      <div class="section-header">
        <span class="title">创作中心</span>
        <div class="publish-btn">
          <span class="upload-icon">↑</span>
          发布
        </div>
      </div>
      <div class="tools-grid">
        <div class="tool-item">
          <div class="icon-wrap orange">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#ff9e1b" stroke-width="2">
              <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
            </svg>
          </div>
          <div class="tool-label">创作中心</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap orange">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#ff9e1b" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
              <polyline points="10 9 9 9 8 9" />
            </svg>
          </div>
          <div class="tool-label">稿件管理</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap orange">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#ff9e1b" stroke-width="2">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
            </svg>
          </div>
          <div class="tool-label">创作激励</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap orange">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#ff9e1b" stroke-width="2">
              <path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z" />
              <line x1="4" y1="22" x2="4" y2="15" />
            </svg>
          </div>
          <div class="tool-label">有奖活动</div>
        </div>
      </div>
    </div>

    <!-- 我的服务 -->
    <div class="card-section">
      <div class="section-header">
        <span class="title">我的服务</span>
      </div>
      <div class="tools-grid services-grid">
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
              <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
            </svg>
          </div>
          <div class="tool-label">我的课程</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <rect x="5" y="2" width="14" height="20" rx="2" ry="2" />
              <line x1="12" y1="18" x2="12.01" y2="18" stroke-width="3" />
            </svg>
          </div>
          <div class="tool-label">免流量服务</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <path d="M20.37 8.91l-8-5a1 1 0 0 0-1.07 0l-8 5A1 1 0 0 0 3 9.76v8.48a1 1 0 0 0 .53.88l8 5a1 1 0 0 0 1.07 0l8-5a1 1 0 0 0 .53-.88V9.76a1 1 0 0 0-.53-.85z" />
            </svg>
          </div>
          <div class="tool-label">个性装扮</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <rect x="2" y="4" width="20" height="16" rx="2" />
              <line x1="12" y1="10" x2="12" y2="16" />
              <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
          </div>
          <div class="tool-label">我的钱包</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <path d="M23 7a2 2 0 0 0-2-2h-4l-3-3H10L7 5H3a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h18a2 2 0 0 0 2-2V7z" />
              <circle cx="12" cy="13" r="4" />
            </svg>
          </div>
          <div class="tool-label">我的直播</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
            </svg>
          </div>
          <div class="tool-label">社区中心</div>
        </div>
        <div class="tool-item">
          <div class="icon-wrap pink">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#fb7299" stroke-width="2">
              <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
            </svg>
          </div>
          <div class="tool-label">能量加油站</div>
        </div>
      </div>
    </div>

    <!-- 更多服务（含退出登录等） -->
    <div class="card-section last-section">
      <div class="section-header">
        <span class="title">更多服务</span>
      </div>
      <div class="logout-btn-wrap">
        <button class="logout-btn" @click="handleLogout">退出登录</button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.space-page {
  min-height: 100vh;
  background: #f6f7f9;
  padding-bottom: 60px; /* 为底部TabBar留空 */
  box-sizing: border-box;
}

/* 顶部操作栏 */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #fff;
  .right-icons {
    display: flex;
    gap: 16px;
    color: #61666d;
    .icon-btn {
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      &:active {
        opacity: 0.7;
      }
    }
  }
}

/* 用户基本信息 */
.user-info-section {
  background: #fff;
  padding: 8px 16px 16px;
  margin-bottom: 12px;

  .avatar-row {
    display: flex;
    align-items: center;
    position: relative;

    .avatar {
      width: 60px;
      height: 60px;
      border-radius: 50%;
      object-fit: cover;
      border: 1px solid #e3e5e7;
    }

    .meta-info {
      margin-left: 14px;
      flex: 1;

      .name-row {
        display: flex;
        align-items: center;
        gap: 6px;

        .nickname {
          font-size: 18px;
          font-weight: bold;
          color: #18191c;
        }

        .level-badge {
          background: #ff6699;
          color: #fff;
          font-size: 9px;
          padding: 1px 4px;
          border-radius: 3px;
          font-weight: bold;
        }
      }

      .member-tag {
        display: inline-block;
        border: 1px solid #fb7299;
        color: #fb7299;
        font-size: 10px;
        padding: 0px 4px;
        border-radius: 3px;
        margin-top: 4px;
        margin-bottom: 4px;
        font-weight: 500;
      }

      .coins-info {
        font-size: 12px;
        color: #9499a0;
        display: flex;
        gap: 12px;
      }
    }

    .space-link {
      font-size: 13px;
      color: #9499a0;
      display: flex;
      align-items: center;
      cursor: pointer;
      position: absolute;
      right: 0;
      top: 50%;
      transform: translateY(-50%);
      .arrow {
        margin-left: 4px;
        font-size: 14px;
      }
    }
  }

  /* 动态、关注、粉丝 */
  .stats-row {
    display: flex;
    align-items: center;
    margin-top: 20px;
    padding-top: 8px;

    .stat-item {
      flex: 1;
      text-align: center;
      cursor: pointer;

      .num {
        font-size: 16px;
        font-weight: bold;
        color: #18191c;
      }

      .label {
        font-size: 11px;
        color: #9499a0;
        margin-top: 4px;
      }
    }

    .divider {
      width: 1px;
      height: 20px;
      background: #e3e5e7;
    }
  }
}

/* 常用功能网格（离线缓存、历史记录、我的收藏、稍后再看） */
.common-tools-grid {
  display: flex;
  background: #fff;
  padding: 16px 0;
  margin-bottom: 12px;

  .tool-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    cursor: pointer;

    .icon-wrap {
      width: 44px;
      height: 44px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 6px;
      background: #f4f7fc;
    }

    .tool-label {
      font-size: 12px;
      color: #18191c;
    }
  }
}

/* 板块卡片卡槽 */
.card-section {
  background: #fff;
  margin-bottom: 12px;
  padding: 14px 16px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 14px;

    .title {
      font-size: 15px;
      font-weight: bold;
      color: #18191c;
    }

    .publish-btn {
      background: #fb7299;
      color: #fff;
      font-size: 12px;
      padding: 4px 12px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      gap: 4px;
      font-weight: bold;
      cursor: pointer;
      &:active {
        background: #f5729a;
      }
    }
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    row-gap: 16px;
    column-gap: 8px;

    .tool-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      cursor: pointer;

      .icon-wrap {
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 6px;
      }

      .tool-label {
        font-size: 11px;
        color: #61666d;
        text-align: center;
        white-space: nowrap;
      }
    }
  }

  .services-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.last-section {
  margin-bottom: 24px;
}

.logout-btn-wrap {
  padding: 8px 0;
  .logout-btn {
    width: 100%;
    height: 40px;
    background: #fff;
    border: 1px solid #e3e5e7;
    color: #f65c5c;
    border-radius: 8px;
    font-size: 14px;
    font-weight: bold;
    cursor: pointer;
    &:active {
      background: #fafafa;
    }
  }
}
</style>
