<template>
    <div class="space-dynamic">
        <h3 class="page-head">
            <span class="t">{{ this.$store.state.user.uid === uid ? '我' : 'TA' }}的动态</span>
            <span class="count">{{ dynamicCount }}</span>
        </h3>
        <div class="dynamic-list-wrap" v-loading="loading">
            <template v-if="!loading">
                <div class="dynamic-list" v-if="list.length > 0">
                    <div class="dynamic-item" v-for="item in list" :key="item.id"
                        @click="$router.push(`/dynamic/${item.id}`)">
                        <div class="dynamic-header">
                            <img class="avatar" :src="item.user?.avatar || defaultAvatar" alt="">
                            <div class="info">
                                <div class="name">{{ item.user?.username || '用户' }}</div>
                                <div class="time">{{ formatTime(item.createdAt) }}</div>
                            </div>
                        </div>
                        <div class="dynamic-content">
                            <p class="text">{{ item.content }}</p>
                            <div class="images" v-if="item.imageUrls && item.imageUrls.length > 0">
                                <img v-for="(url, i) in item.imageUrls" :key="i" :src="url" alt="" loading="lazy">
                            </div>
                            <div class="ref-video" v-if="item.refVideo">
                                <i class="iconfont icon-bofangshu"></i>
                                <span class="ref-title">{{ item.refVideo.title || '引用了视频' }}</span>
                            </div>
                        </div>
                        <div class="dynamic-footer">
                            <span><i class="iconfont icon-dianzan"></i> {{ item.likeCount || 0 }}</span>
                            <span><i class="iconfont icon-pinglun"></i> {{ item.commentCount || 0 }}</span>
                            <span><i class="iconfont icon-zhuanfa"></i> {{ item.shareCount || 0 }}</span>
                        </div>
                    </div>
                </div>
                <div v-else class="empty-tip">
                    <i class="iconfont icon-meiyoupinleimu"></i>
                    <p>TA还没有发布动态</p>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
import { dynamicApi } from '@/api/dynamic';

export default {
    name: "SpaceDynamic",
    data() {
        return {
            list: [],
            dynamicCount: 0,
            loading: false,
            defaultAvatar: 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
        }
    },
    computed: {
        uid() {
            return Number(this.$route.params.uid);
        }
    },
    methods: {
        async getDynamicList() {
            this.loading = true;
            try {
                const res = await dynamicApi.getUserDynamics(this.uid, 1, 20);
                if (res.code === 200) {
                    const data = res.data && res.data.list ? res.data.list : (Array.isArray(res.data) ? res.data : []);
                    this.list = data;
                    this.dynamicCount = data.length;
                } else {
                    this.list = [];
                    this.dynamicCount = 0;
                }
            } catch (e) {
                console.error('获取动态列表失败:', e);
                this.list = [];
                this.dynamicCount = 0;
            } finally {
                this.loading = false;
            }
        },
        formatTime(time) {
            if (!time) return '';
            const d = new Date(time);
            const now = new Date();
            const diff = now - d;
            if (diff < 60000) return '刚刚';
            if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
            if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
            if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`;
            return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' });
        },
    },
    watch: {
        "$route.params.uid"() {
            this.getDynamicList();
        }
    },
    created() {
        this.getDynamicList();
    }
}
</script>

<style scoped>
.space-dynamic {
    background: #fff;
    padding: 20px;
    min-height: 400px;
    box-shadow: 0 0 0 1px #eee;
    border-radius: 4px;
}

.page-head {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 16px;
    border-bottom: 1px solid #eee;
    margin-bottom: 20px;
}

.page-head .t {
    font-size: 17px;
    color: #222;
    font-weight: 500;
}

.page-head .count {
    font-size: 14px;
    color: #99a2aa;
    background: #f4f5f7;
    border-radius: 2px;
    padding: 0 8px;
    line-height: 20px;
}

.dynamic-list-wrap {
    min-height: 300px;
}

.dynamic-item {
    padding: 16px 0;
    border-bottom: 1px solid #f1f2f3;
    cursor: pointer;
    transition: background-color .2s;
    border-radius: 4px;
}

.dynamic-item:hover {
    background-color: #fafafa;
}

.dynamic-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;
}

.dynamic-header .avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
}

.dynamic-header .info .name {
    font-size: 14px;
    color: #61666d;
    line-height: 18px;
}

.dynamic-header .info .time {
    font-size: 12px;
    color: #99a2aa;
}

.dynamic-content .text {
    font-size: 14px;
    color: #222;
    line-height: 22px;
    margin-bottom: 8px;
    word-break: break-word;
}

.dynamic-content .images {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 8px;
}

.dynamic-content .images img {
    width: 80px;
    height: 80px;
    object-fit: cover;
    border-radius: 4px;
}

.ref-video {
    display: flex;
    align-items: center;
    gap: 8px;
    background: #f4f5f7;
    border-radius: 4px;
    padding: 10px 12px;
    margin-bottom: 8px;
    color: #61666d;
    font-size: 13px;
}

.ref-video .icon-bofangshu {
    color: var(--brand_pink);
}

.dynamic-footer {
    display: flex;
    gap: 24px;
    color: #9499a0;
    font-size: 13px;
}

.dynamic-footer span {
    display: flex;
    align-items: center;
    gap: 4px;
}

.empty-tip {
    text-align: center;
    padding: 80px 0;
    color: #99a2aa;
}

.empty-tip .iconfont {
    font-size: 60px;
    color: #d8dde3;
    display: block;
    margin-bottom: 16px;
}
</style>