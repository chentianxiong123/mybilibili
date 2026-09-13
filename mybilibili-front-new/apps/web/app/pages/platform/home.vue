<template>
    <div class="platform-home-overview">
        <h2 class="page-title">创作中心首页</h2>

        <!-- 概览卡片 -->
        <div class="overview-cards" v-loading="loading.overview">
            <div class="stat-card primary">
                <div class="stat-label">总粉丝数</div>
                <div class="stat-value">{{ overview.totalFollowers?.toLocaleString() || 0 }}</div>
                <div class="stat-trend" :class="overview.followersIncrease >= 0 ? 'up' : 'down'">
                    {{ overview.followersIncrease >= 0 ? '+' : '' }}{{ overview.followersIncrease || 0 }} 今日
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-label">总播放量</div>
                <div class="stat-value">{{ overview.totalViews?.toLocaleString() || 0 }}</div>
                <div class="stat-trend" :class="overview.viewsIncrease >= 0 ? 'up' : 'down'">
                    {{ overview.viewsIncrease >= 0 ? '+' : '' }}{{ overview.viewsIncrease || 0 }} 今日
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-label">总点赞数</div>
                <div class="stat-value">{{ overview.totalLikes?.toLocaleString() || 0 }}</div>
                <div class="stat-trend" :class="overview.likesIncrease >= 0 ? 'up' : 'down'">
                    {{ overview.likesIncrease >= 0 ? '+' : '' }}{{ overview.likesIncrease || 0 }} 今日
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-label">总评论数</div>
                <div class="stat-value">{{ overview.totalComments?.toLocaleString() || 0 }}</div>
                <div class="stat-trend" :class="overview.commentsIncrease >= 0 ? 'up' : 'down'">
                    {{ overview.commentsIncrease >= 0 ? '+' : '' }}{{ overview.commentsIncrease || 0 }} 今日
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-label">总弹幕数</div>
                <div class="stat-value">{{ overview.totalDanmaku?.toLocaleString() || 0 }}</div>
                <div class="stat-trend" :class="overview.danmakuIncrease >= 0 ? 'up' : 'down'">
                    {{ overview.danmakuIncrease >= 0 ? '+' : '' }}{{ overview.danmakuIncrease || 0 }} 今日
                </div>
            </div>
            <div class="stat-card">
                <div class="stat-label">稿件数</div>
                <div class="stat-value">{{ overview.totalManuscripts?.toLocaleString() || 0 }}</div>
            </div>
        </div>

        <!-- 快捷入口 -->
        <div class="quick-actions">
            <h3 class="section-title">快捷操作</h3>
            <div class="actions-grid">
                <div class="action-item" @click="$router.push('/platform/upload/video')">
                    <i class="iconfont icon-shangchuan"></i>
                    <span>上传视频</span>
                </div>
                <div class="action-item" @click="$router.push('/platform/upload-manager/manuscript')">
                    <i class="iconfont icon-gaojian"></i>
                    <span>管理稿件</span>
                </div>
                <div class="action-item" @click="$router.push('/platform/data-up')">
                    <i class="iconfont icon-shuju"></i>
                    <span>查看数据</span>
                </div>
                <div class="action-item" @click="$router.push('/platform/comment')">
                    <i class="iconfont icon-hudong"></i>
                    <span>回复评论</span>
                </div>
            </div>
        </div>

        <div class="welcome-tip">
            <p>欢迎回到创作中心！从左侧菜单可以管理你的稿件、查看数据统计、处理评论与弹幕。</p>
            <p v-if="overview.updateTime" class="update-time">数据更新于 {{ overview.updateTime }}</p>
        </div>
    </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useCreatorStats } from '@/composables/useCreatorStats'

const { loading, overview, loadOverview } = useCreatorStats()

onMounted(() => {
    loadOverview()
})
</script>

<style scoped>
.platform-home-overview {
    padding: 24px;
}

.page-title {
    font-size: 22px;
    font-weight: 600;
    color: var(--text1);
    margin: 0 0 24px 0;
}

.overview-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
    min-height: 120px;
}

.stat-card {
    background: #fff;
    border: 1px solid var(--line_regular);
    border-radius: 8px;
    padding: 20px;
    transition: transform .2s, box-shadow .2s;
}

.stat-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.stat-card.primary {
    background: linear-gradient(135deg, #FF7DA1 0%, #fa799e 100%);
    color: #fff;
    border-color: transparent;
}

.stat-card.primary .stat-label,
.stat-card.primary .stat-trend {
    color: rgba(255, 255, 255, 0.85);
}

.stat-label {
    font-size: 13px;
    color: var(--text3);
    margin-bottom: 8px;
}

.stat-value {
    font-size: 26px;
    font-weight: 600;
    color: inherit;
    line-height: 1.2;
    margin-bottom: 4px;
}

.stat-trend {
    font-size: 12px;
    color: var(--text3);
}

.stat-trend.up {
    color: #67c23a;
}

.stat-trend.down {
    color: #f56c6c;
}

.stat-card.primary .stat-trend.up,
.stat-card.primary .stat-trend.down {
    color: rgba(255, 255, 255, 0.95);
}

.section-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--text1);
    margin: 0 0 16px 0;
}

.quick-actions {
    background: #fff;
    border: 1px solid var(--line_regular);
    border-radius: 8px;
    padding: 24px;
    margin-bottom: 24px;
}

.actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 16px;
}

.action-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    border: 1px solid var(--line_regular);
    border-radius: 6px;
    cursor: pointer;
    transition: all .2s;
    font-size: 14px;
    color: var(--text1);
}

.action-item:hover {
    border-color: var(--brand_pink);
    color: var(--brand_pink);
    background: rgba(255, 125, 161, 0.04);
}

.action-item .iconfont {
    font-size: 22px;
    color: var(--brand_pink);
}

.welcome-tip {
    background: #fff;
    border: 1px solid var(--line_regular);
    border-radius: 8px;
    padding: 20px 24px;
    color: var(--text2);
    font-size: 14px;
    line-height: 1.8;
}

.welcome-tip p {
    margin: 0;
}

.welcome-tip .update-time {
    margin-top: 8px;
    font-size: 12px;
    color: var(--text3);
}
</style>
