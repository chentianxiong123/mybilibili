<template>
    <div class="notify-list">
        <div class="list-header">
            <span class="header-title">@ 我的</span>
            <span class="mark-all" v-if="unreadIds.length > 0" @click="markAllRead">全部已读</span>
        </div>
        <div class="list-loading" v-if="loading"><MessageLoading /></div>
        <div class="list-error" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
        <div class="list-empty" v-else-if="list.length === 0">暂时没有人 @ 你</div>
        <div class="notify-item" v-for="item in list" :key="item.id" :class="{ unread: !item.isRead }">
            <a class="avatar" :href="`/space/${item.senderId}`" target="_blank"
                :style="`background-image: url('${item.userAvatar || defaultAvatar}');`"></a>
            <div class="item-main">
                <div class="item-top">
                    <a class="username" :href="`/space/${item.senderId}`" target="_blank">{{ item.username || '用户' }}</a>
                    <span class="action">{{ item.actionText || '在评论中@了我' }}</span>
                    <span class="time">{{ formatTime(item.createdAt) }}</span>
                </div>
                <div class="item-content" v-html="escapeHtml(stripActionText(item.content))"></div>
                <div class="item-link" v-if="item.manuscriptId">
                    <a :href="`/video/${item.manuscriptId}`" target="_blank">查看原视频 &gt;</a>
                </div>
            </div>
            <span class="unread-dot" v-if="!item.isRead"></span>
        </div>
    </div>
</template>

<script lang="ts">
import MessageLoading from './MessageLoading.vue';
import { messageApi } from '@/api/message.ts';
import { handleDateTime } from '@/teriteri-src/utils/utils';

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png';

export default {
    name: 'AtList',
    components: { MessageLoading },
    data() {
        return {
            list: [] as any[],
            loading: false,
            error: false,
        };
    },
    computed: {
        unreadIds(): number[] {
            return this.list.filter(i => !i.isRead).map(i => i.id);
        },
    },
    methods: {
        escapeHtml(s: string) {
            return String(s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        },
        // 后端 content 自带前缀（"回复了你的评论："/"在评论中@了你："），
        // 顶部已用 actionText 显示过一次，这里剥掉避免重复
        stripActionText(s: string) {
            return String(s || '').replace(/^(回复了你的评论|在评论中@了你)[：:]?/, '');
        },
        formatTime(t: string) {
            try {
                return handleDateTime(t);
            } catch (e) {
                return t || '';
            }
        },
        async load() {
            this.loading = true;
            this.error = false;
            try {
                const res = await messageApi.getAtList({});
                if (res && res.code === 200 && Array.isArray(res.data)) {
                    this.list = res.data;
                } else {
                    throw new Error('bad response');
                }
            } catch (e) {
                this.error = true;
            } finally {
                this.loading = false;
            }
        },
        async markAllRead() {
            const ids = this.unreadIds;
            if (ids.length === 0) return;
            try {
                await messageApi.batchMarkAsRead(ids);
            } catch (e) {
                // 本地先行，失败下次进页再清
            }
            this.list.forEach(i => { i.isRead = true; });
            this.refreshUnread();
        },
        async refreshUnread() {
            try {
                const res = await messageApi.getUnreadCounts();
                const d = res && (res.data || res);
                const unread = (this as any).$store.state.msgUnread;
                if (unread && d) {
                    unread[0] = d.reply || 0;
                    unread[1] = d.at || 0;
                    unread[2] = d.like || 0;
                    unread[3] = d.system || 0;
                    unread[4] = d.private || 0;
                }
            } catch (e) {
                // 红点刷新失败不影响列表展示
            }
        },
    },
    async mounted() {
        await this.load();
        if (this.unreadIds.length > 0) {
            this.markAllRead();
        } else {
            this.refreshUnread();
        }
    },
};
</script>

<style scoped>
.notify-list {
    background: #fff;
    border-radius: 4px;
    min-height: 400px;
    padding: 0 16px 16px;
}
.list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 48px;
    border-bottom: 1px solid #f0f0f0;
}
.header-title {
    font-size: 14px;
    font-weight: 600;
    color: #333;
}
.mark-all {
    font-size: 12px;
    color: var(--brand_pink, #fb7299);
    cursor: pointer;
}
.mark-all:hover {
    opacity: 0.8;
}
.list-loading, .list-empty, .list-error {
    text-align: center;
    padding: 60px 0;
    color: #999;
    font-size: 13px;
}
.list-error .btn {
    color: var(--brand_pink, #fb7299);
    cursor: pointer;
}
.notify-item {
    display: flex;
    padding: 14px 0;
    border-bottom: 1px solid #f4f4f4;
    position: relative;
}
.notify-item:last-child {
    border-bottom: none;
}
.avatar {
    width: 44px;
    height: 44px;
    min-width: 44px;
    border-radius: 50%;
    background-size: cover;
    background-position: center;
    background-color: #f0f0f0;
}
.item-main {
    flex: 1;
    margin-left: 12px;
    min-width: 0;
}
.item-top {
    font-size: 13px;
    line-height: 20px;
}
.username {
    color: #333;
    font-weight: 600;
    text-decoration: none;
}
.username:hover {
    color: var(--brand_pink, #fb7299);
}
.action {
    color: #666;
    margin-left: 6px;
}
.time {
    color: #999;
    font-size: 12px;
    margin-left: 8px;
}
.item-content {
    margin-top: 6px;
    font-size: 13px;
    color: #333;
    line-height: 20px;
    word-break: break-all;
    background: #f7f7f7;
    border-radius: 4px;
    padding: 8px 10px;
}
.item-link {
    margin-top: 6px;
    font-size: 12px;
}
.item-link a {
    color: var(--brand_pink, #fb7299);
    text-decoration: none;
}
.unread-dot {
    position: absolute;
    right: 2px;
    top: 20px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--stress_red, #fb7299);
}
</style>
