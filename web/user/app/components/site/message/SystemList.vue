<template>
    <div class="notify-list">
        <div class="list-loading" v-if="loading"><MessageLoading /></div>
        <div class="list-error" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
        <div class="list-empty" v-else-if="list.length === 0">暂时没有系统通知</div>
        <div class="system-item" v-for="item in list" :key="item.id">
            <div class="notify-item">
                <div class="notify-item__top">
                    <span class="notify-item__title">{{ item.title || '系统通知' }}</span>
                    <span class="notify-item__time">{{ formatTime(item.createdAt) }}</span>
                </div>
                <div class="notify-item__bottom">
                    <div class="notify-item__msg">{{ item.content }}</div>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import MessageLoading from './MessageLoading.vue';
import { messageApi } from '@/api/message.ts';
import { handleDateTime } from '@/shared/utils/utils';

export default {
    name: 'SystemList',
    components: { MessageLoading },
    data() {
        return {
            list: [] as any[],
            loading: false,
            error: false,
        };
    },
    methods: {
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
                const res = await messageApi.getSystemNotifications({});
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
            const ids = this.list.filter(i => !i.isRead).map(i => i.id);
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
        if (this.list.some(i => !i.isRead)) {
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
.list-loading, .list-empty, .list-error {
    text-align: center;
    padding: 60px 0;
    color: #999;
    font-size: 13px;
}
.list-error .btn {
    color: var(--brand_pink, #00a1d6);
    cursor: pointer;
}
.system-item {
    padding-top: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
}
.system-item:last-child {
    border-bottom: none;
}
.notify-item__top {
    display: block;
    line-height: 24px;
}
.notify-item__title {
    display: inline;
    font-size: 13px;
    font-weight: 700;
    color: #18191c;
    line-height: 24px;
}
.notify-item__time {
    display: inline;
    font-size: 12px;
    color: #9499a0;
    line-height: 24px;
    margin-left: 10px;
}
.notify-item__bottom {
    display: block;
    margin-top: 4px;
}
.notify-item__msg {
    font-size: 14px;
    color: #18191c;
    line-height: 24px;
    word-break: break-all;
}
</style>
