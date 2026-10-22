<template>
    <div class="love">
        <div class="love-layout">
            <div class="love-actions" v-if="unreadIds.length > 0">
                <span class="mark-all" @click="markAllRead">全部已读</span>
            </div>
            <div class="love-empty" v-if="loading"><MessageLoading /></div>
            <div class="love-empty" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
            <div class="love-empty empty" v-else-if="list.length === 0"><MessageEmpty /></div>
            <div class="love-item" v-for="item in list" :key="item.id" :class="{ unread: !item.isRead }">
                <div class="love-item__left">
                    <a class="love-item__avatar" :href="`/space/${item.senderId}`" target="_blank"
                        :style="`background-image: url('${item.userAvatar || defaultAvatar}');`"></a>
                </div>
                <div class="love-item__right">
                    <div class="love-item__info">
                        <div class="love-item__title">
                            <a class="love-item__uname" :href="`/space/${item.senderId}`" target="_blank">
                                {{ item.username || item.usernames || '用户' }}
                            </a>
                            <div class="love-item__action">{{ item.actionText || '赞了我' }}</div>
                        </div>
                        <div class="love-item__msg" v-if="item.commentId || isCommentLike(item)">
                            {{ item.commentContent || stripVideoTitle(item.content) }}
                        </div>
                        <div class="love-item__reference" v-else-if="item.articleId">
                            专栏《{{ item.articleTitle }}》
                        </div>
                        <div class="love-item__reference" v-else-if="item.videoTitle">
                            视频《{{ item.videoTitle }}》
                        </div>
                        <div class="love-item__time">{{ formatTime(item.createdAt) }}</div>
                    </div>
                    <a class="love-item__cover" v-if="item.videoCover && !isCommentLike(item)"
                        :href="`/video/${item.manuscriptId}`" target="_blank"
                        :style="`background-image: url('${item.videoCover}');`"></a>
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

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png';

export default {
    name: 'LikeList',
    components: { MessageLoading, MessageEmpty },
    data() {
        return {
            list: [] as any[],
            loading: false,
            error: false,
            defaultAvatar,
        };
    },
    computed: {
        unreadIds(): number[] {
            return this.list.filter(i => !i.isRead).map(i => i.id);
        },
    },
    methods: {
        isCommentLike(item: any) {
            return String(item.actionText || '').includes('评论');
        },
        stripVideoTitle(s: string) {
            return String(s || '').replace(/^赞了你的评论/, '');
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
                const res = await messageApi.getLikes({});
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
.love-layout {
    height: 100%;
    overflow-y: auto;
}
.love-actions {
    display: flex;
    justify-content: flex-end;
    padding: 8px 16px 0;
}
.mark-all {
    color: var(--brand_blue, #00a1d6);
    font-size: 13px;
    cursor: pointer;
}
.mark-all:hover {
    opacity: 0.8;
}
.love-empty .btn {
    color: var(--brand_blue, #00a1d6);
    cursor: pointer;
}
.love-empty.empty {
    padding: 0;
}
.love-item__avatar {
    background-size: cover;
    background-position: center;
    background-color: var(--graph_bg_thick, #f1f2f3);
}
.love-item.unread .love-item__uname {
    color: var(--brand_blue, #00a1d6);
}
.love-item__action {
    display: flex;
    align-items: center;
}
.love-item__msg {
    margin-bottom: 6px;
    color: var(--text2);
    font-size: 13px;
    line-height: 18px;
    word-break: break-word;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}
.love-item__reference {
    margin-bottom: 6px;
    color: var(--text2);
    font-size: 13px;
    line-height: 18px;
}
.love-item__cover {
    flex-shrink: 0;
    width: 112px;
    height: 63px;
    margin-left: 16px;
    border-radius: 4px;
    background-size: cover;
    background-position: center;
    background-color: var(--graph_bg_thick, #f1f2f3);
}
</style>
