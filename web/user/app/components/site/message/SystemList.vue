<template>
    <div class="system">
        <div class="system-layout">
            <div class="system-list">
                <div class="system-item" v-if="loading"><MessageLoading /></div>
                <div class="system-item" v-else-if="error">
                    加载失败，<span class="btn" @click="load">点击重试</span>
                </div>
                <div class="system-item empty" v-else-if="list.length === 0"><MessageEmpty /></div>
                <div class="system-item" v-for="item in list" :key="item.id">
                    <div class="notify-item message-box-shadow">
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
        </div>
    </div>
</template>

<script lang="ts">
import MessageLoading from './MessageLoading.vue';
import MessageEmpty from './MessageEmpty.vue';
import { messageApi } from '@/api/message.ts';
import { handleDateTime } from '@/shared/utils/utils';

export default {
    name: 'SystemList',
    components: { MessageLoading, MessageEmpty },
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
.system-layout {
    height: 100%;
    overflow-y: auto;
}
.system-item.empty {
    margin-bottom: 0;
}
.system-item .btn {
    color: var(--brand_blue, #00a1d6);
    cursor: pointer;
}
</style>
