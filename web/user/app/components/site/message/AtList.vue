<template>
    <div class="at">
        <div class="at-layout">
            <div class="at-actions" v-if="unreadIds.length > 0">
                <span class="mark-all" @click="markAllRead">全部已读</span>
            </div>
            <div class="at-empty" v-if="loading"><MessageLoading /></div>
            <div class="at-empty" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
            <div class="at-empty empty" v-else-if="list.length === 0"><MessageEmpty /></div>
            <template v-for="(item, idx) in list" :key="item.id">
                <MessageDivider v-if="idx === dividerIndex" />
                <div class="interaction-item"
                    :class="{ unread: !item.isRead }" @click="openTarget(item)">
                    <div class="interaction-item__left">
                        <div class="interaction-item__avatar">
                            <div class="message-user-avatar"
                                :style="`background-image: url('${item.userAvatar || defaultAvatar}');`"></div>
                        </div>
                    </div>
                    <div class="interaction-item__right">
                        <div class="interaction-item__content">
                            <a class="interaction-item__cover" v-if="item.articleCover || item.videoCover"
                                :href="item.articleId ? `/read/cv${item.articleId}` : `/video/${item.manuscriptId}`"
                                target="_blank" @click.stop>
                                <div class="b-img"
                                    :style="`background-image: url('${item.articleCover || item.videoCover}');`"></div>
                            </a>
                            <div class="interaction-item__title">
                                <a class="interaction-item__uname" :href="`/space/${item.senderId}`" target="_blank"
                                    @click.stop>{{ item.username || '用户' }}</a>
                                <span class="interaction-item__action">{{ item.actionText || '在评论中@了我' }}</span>
                            </div>
                            <div class="interaction-item__msg">{{ stripActionText(item.content) }}</div>
                            <div class="interaction-item__reference" v-if="item.articleId">
                                查看专栏{{ item.articleTitle ? `《${item.articleTitle}》` : '' }} &gt;
                            </div>
                            <div class="interaction-item__reference" v-else-if="item.manuscriptId">
                                查看原视频 &gt;
                            </div>
                            <div class="interaction-item__metadata">
                                <span class="interaction-item__time">{{ formatTime(item.createdAt) }}</span>
                                <div class="interaction-item__btn invisible" @click.stop="removeItem(item)">
                                    <span>删除该通知</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </div>
    </div>
</template>

<script lang="ts">
import MessageLoading from './MessageLoading.vue';
import MessageEmpty from './MessageEmpty.vue';
import MessageDivider from './MessageDivider.vue';
import { messageApi } from '@/api/message.ts';
import { handleDateTime } from '@/shared/utils/utils';

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png';

export default {
    name: 'AtList',
    components: { MessageLoading, MessageEmpty, MessageDivider },
    data() {
        return {
            list: [] as any[],
            loading: false,
            error: false,
            dividerIndex: -1,
            defaultAvatar,
        };
    },
    computed: {
        unreadIds(): number[] {
            return this.list.filter(i => !i.isRead).map(i => i.id);
        },
    },
    methods: {
        updateDivider() {
            const idx = this.list.findIndex((i: any) => i.isRead);
            this.dividerIndex = idx > 0 ? idx : -1;
        },
        escapeHtml(s: string) {
            return String(s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        },
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
        openTarget(item: any) {
            if (item.articleId) {
                window.open(`/read/cv${item.articleId}`, '_blank');
            } else if (item.manuscriptId) {
                window.open(`/video/${item.manuscriptId}`, '_blank');
            }
        },
        async removeItem(item: any) {
            try {
                await messageApi.deleteMessage(item.id);
            } catch (e) {
                // 删除失败保留本地列表，下次刷新会重新出现
                return;
            }
            this.list = this.list.filter(i => i.id !== item.id);
        },
        async load() {
            this.loading = true;
            this.error = false;
            try {
                const res = await messageApi.getAtList({});
                if (res && res.code === 200 && Array.isArray(res.data)) {
                    this.list = res.data;
                    this.updateDivider();
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
            try {
                await messageApi.markCategoryRead([3]);
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
.at-layout {
    overflow-y: auto;
}
.at-actions {
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
.at-empty .btn {
    color: var(--brand_blue, #00a1d6);
    cursor: pointer;
}
.at-empty.empty {
    padding: 0;
}
.message-user-avatar {
    width: 100%;
    height: 100%;
    background-size: cover;
    background-position: center;
    background-color: var(--graph_bg_thick, #f1f2f3);
    border-radius: 50%;
}
.interaction-item.unread .interaction-item__msg {
    font-weight: 600;
}
</style>
