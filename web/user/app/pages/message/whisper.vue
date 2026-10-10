<template>
    <div class="message-whisper">
        <div class="im">
            <div class="left">
                <div class="title">
                    <span>近期消息</span>
                </div>
                <div class="list-container" id="chat-list">
                    <div class="list">
                        <div class="list-item" :class="{'active': this.$store.state.chatId === item.user.uid}"
                            @click="changeChat(item.user.uid)"
                            v-for="(item, index) in this.$store.state.chatList" :key="index"
                        >
                            <VAvatar class="avatar" :size="40" :img="item.user.avatar_url" :auth="item.user.auth"></VAvatar>
                            <div class="name-box">
                                <div class="name">{{ item.user.nickname }}</div>
                                <div class="last-word" v-if="lastWord(item)">
                                    {{ lastWord(item) }}
                                </div>
                            </div>
                            <div class="close">
                                <svg @click.stop="closeChat(item)" t="1702107603641" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="7402" width="20" height="20"><path d="M791.67016602 269.61962891l-37.28979493-37.28320313L512 474.71020508l-242.3737793-242.3803711-37.28979492 37.28979493L474.71020508 512l-242.3737793 242.38037109 37.28320313 37.28979493L512 549.28979492l242.38037109 242.3803711 37.28979493-37.28979493L549.28979492 512z" p-id="7403"></path></svg>
                            </div>
                            <div class="notify notify-number" v-if="item.chat.unread > 0">{{ item.chat.unread }}</div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="right">
                <NuxtPage />
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import VAvatar from '@/components/teriteri/avatar/VAvatar.vue';
import { messageApi } from '@/api/message.ts';

export default {
    name: "MessageWhisper",
    components: {
        VAvatar,
    },
    data() {
        return {
            loading: false,     // 是否正在加载会话列表
            sse: null,          // 私信实时推送（EventSource，连 /sse/notification）
        }
    },
    methods: {
        // 左栏预览：优先用后端 last_message_content（preview 字段），
        // 发出去的新消息直接读 detail 末尾
        lastWord(item) {
            const list = (item.detail && item.detail.list) || [];
            if (list.length > 0) return list[list.length - 1].content;
            return item.preview || '';
        },

        ///////// 请求 /////////
        // 获取会话列表：走适配层（request.ts 把后端 Conversation 转成
        // {user, chat, detail, preview} 旧形状）。
        // 后端无分页，一次全量；未登录/失败时静默空列表（message.vue 负责踢回登录页）
        async getChatList() {
            if (this.loading) return;
            this.loading = true;
            try {
                const res = await this.$get("/msg/chat/recent-list");
                const payload = res && res.data && res.data.data;
                const list = payload && Array.isArray(payload.list) ? payload.list : [];
                if (list.length > 0) {
                    // 去重后追加（路由切换回来时不重复拉取）
                    const exists = new Set(this.$store.state.chatList.map(i => i.user.uid));
                    const fresh = list.filter(i => i && i.user && !exists.has(i.user.uid));
                    this.$store.commit("updateChatList", fresh);
                }
            } catch (e) {
                // 静默失败，保持空列表
            }
            this.$nextTick(() => {
                this.loading = false;
            });
        },

        // 移除聊天
        async closeChat(item) {
            const uid = item.user.uid;
            const i = this.$store.state.chatList.findIndex(c => c.user.uid === uid);
            if (i !== -1) this.$store.state.chatList.splice(i, 1);
            // 占位会话（id=0，尚未在后端落库）无需调删除接口
            if (item.chat && item.chat.id) {
                try {
                    await messageApi.deleteConversation(item.chat.id);
                } catch (e) {
                    // 本地已移除，后端失败不回滚
                }
            }
            if (Number(this.$route.params.mid) === uid) {
                // 如果移除的是当前聊天 就打开第一个聊天
                if (this.$store.state.chatList.length > 0) {
                    this.changeChat(this.$store.state.chatList[0].user.uid);
                } else {
                    this.$store.state.chatId = -1;
                    this.$router.push('/message/whisper');
                }
            }
        },

        ///////// 实时推送 /////////
        // 连后端 /sse/notification（网关已指到 msg-danmaku:8086）。
        // EventSource 设不了请求头，靠同源 HttpOnly cookie 鉴权（与 videoProcess.ts 同理）。
        // 事件：connected（忽略）、unread_init/unread_counts（校准红点）、
        // message（新私信，增量进会话）。断线浏览器自动重连，无需自建逻辑
        connectSSE() {
            if (this.sse || !this.$store.state.isLogin) return;
            try {
                const es = new EventSource('/sse/notification');
                es.onmessage = (e) => this.handleSSEMessage(e);
                es.onerror = () => {
                    // 掉登录后服务端拒连，无限重试无意义，关掉等下次进页重连
                    if (!this.$store.state.isLogin && this.sse) {
                        this.sse.close();
                        this.sse = null;
                    }
                };
                this.sse = es;
            } catch (e) {
                // 不支持 EventSource 的环境降级为纯刷新模式
                this.sse = null;
            }
        },

        handleSSEMessage(e) {
            let msg = null;
            try {
                msg = JSON.parse(e.data);
            } catch (err) {
                return;
            }
            if (!msg || !msg.type) return;
            if (msg.type === 'unread_init' || msg.type === 'unread_counts') {
                // 校准全站红点（下标见 message.vue 菜单注释）
                const d = msg.data || {};
                const unread = this.$store.state.msgUnread;
                if (unread) {
                    unread[0] = d.reply || 0;
                    unread[1] = d.at || 0;
                    unread[2] = d.like || 0;
                    unread[3] = d.system || 0;
                    unread[4] = d.private || 0;
                    if (unread.length > 5) unread[5] = d.dynamic || 0;
                }
                return;
            }
            if (msg.type === 'message' && msg.from_uid) {
                this.handleIncoming(msg);
            }
            // connected / system 等与私信无关，忽略
        },

        // 新私信到达：会话存在则取最新一条落本地（带真实 id，去重安全），
        // 不存在则先建会话项；正看着的直接标已读，否则红点+1
        async handleIncoming(msg) {
            const fromUid = Number(msg.from_uid);
            const myUid = this.$store.state.user && this.$store.state.user.uid;
            if (!fromUid || fromUid === myUid) return;
            try {
                let item = this.$store.state.chatList.find(i => i.user.uid === fromUid);
                if (!item) {
                    const res = await this.$get(`/msg/chat/create/${fromUid}`);
                    const got = res && res.data && res.data.data;
                    if (got && got.user) {
                        this.$store.commit("updateChatList", [got]);
                        item = this.$store.state.chatList.find(i => i.user.uid === fromUid);
                    }
                }
                if (!item || !item.chat || !item.chat.id) return;
                // 取该会话最新一条（后端已落库，拿真实 id，避免与本地临时消息重复）
                const res = await messageApi.getMessages(item.chat.id, 1, 1);
                const raw = (res && res.code === 200 && Array.isArray(res.data)) ? res.data : [];
                if (raw.length === 0) return;
                const m = raw[0];
                const exists = item.detail.list.some(x => x.id === m.id);
                if (!exists) {
                    item.detail.list.push({
                        id: m.id,
                        userId: m.sender_id ?? m.senderId ?? fromUid,
                        content: m.content || msg.content || '',
                        withdraw: 0,
                        time: m.created_at || m.createdAt || msg.created_at || '',
                    });
                }
                item.preview = m.content || msg.content || '';
                // 提到最前（最新会话置顶，与后端排序一致）
                const i = this.$store.state.chatList.findIndex(x => x.user.uid === fromUid);
                if (i > 0) {
                    const [it] = this.$store.state.chatList.splice(i, 1);
                    this.$store.state.chatList.unshift(it);
                }
                if (Number(this.$route.params.mid) === fromUid) {
                    // 正看着这个对话框：直接标已读，不涨红点
                    item.chat.unread = 0;
                    try {
                        await messageApi.markConversationRead(item.chat.id);
                    } catch (err) {
                        // 本地已清零，后端失败不回滚
                    }
                } else {
                    item.chat.unread = (item.chat.unread || 0) + 1;
                    const unread = this.$store.state.msgUnread;
                    if (unread) unread[4] = (unread[4] || 0) + 1;
                }
            } catch (err) {
                // 推送处理失败不影响主流程，下次刷新/进页会全量校准
            }
        },

        ///////// 事件 /////////
        // 切换聊天
        changeChat(mid) {
            this.$router.push(`/message/whisper/${mid}`);
        },
    },
    async mounted() {
        this.$store.state.isChatPage = true;
        await this.getChatList();
        if (this.$route.path.startsWith('/message/whisper/')) {
            this.$store.state.chatId = Number(this.$route.params.mid);
        } else {
            this.$store.state.chatId = -1;
        }
        this.connectSSE();
    },
    beforeUnmount() {
        this.$store.state.isChatPage = false;
        if (this.sse) {
            this.sse.close();
            this.sse = null;
        }
    },
    watch: {
        // 登录态是异步回来的（app.vue getPersonalInfo），挂载时可能还没就绪；
        // 就绪后补连 SSE，退出后断开
        "$store.state.isLogin"(v) {
            if (v) {
                this.connectSSE();
            } else if (this.sse) {
                this.sse.close();
                this.sse = null;
            }
        },
        // 监听路由变化打开对应聊天
        "$route.path"() {
            if (this.$route.path.startsWith('/message/whisper/')) {
                this.$store.state.chatId = Number(this.$route.params.mid);
            } else {
                this.$store.state.chatId = -1;
            }
        }
    }
}
</script>

<style scoped>
.message-whisper {
    height: calc(100vh - 138px);
}

.im {
    font-size: 12px;
    color: #666;
    background-color: #fff;
    -webkit-box-shadow: 0 2px 4px 0 rgba(121,146,185,0.54);
    box-shadow: 0 2px 4px 0 rgba(121,146,185,0.54);
    border-radius: 4px;
    height: 100%;
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    position: relative;
}

.left {
    width: 240px;
    border-right: 1px solid #e9eaec;
    position: relative;
}

.left .title {
    padding-left: 24px;
    line-height: 35px;
    height: 36px;
    border-bottom: 1px solid #e9eaec;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
    overflow: hidden;
}

.left .list-container {
    height: calc(100% - 36px);
    position: relative;
    overflow-x: hidden;
    overflow-y: auto;
}

.left .list-container::-webkit-scrollbar {
    width: 4px;
}

.left .list-container::-webkit-scrollbar-thumb {
    border-radius: 4px;
    background-color: #ddd;
}

.list-item {
    width: 100%;
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-orient: horizontal;
    -webkit-box-direction: normal;
    -ms-flex-direction: row;
    flex-direction: row;
    padding: 19px 24px;
    position: relative;
    overflow: hidden;
    cursor: pointer;
}

.active {
    background-color: #e4e5e6;
}

.list-item:hover {
    background-color: #e4e5e6;
}

.list-item .avatar {
    width: 40px;
    height: 40px;
    margin-right: 8px;
    -ms-flex-negative: 0;
    flex-shrink: 0;
    position: relative;
}

.list-item .name-box {
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-orient: vertical;
    -webkit-box-direction: normal;
    -ms-flex-direction: column;
    flex-direction: column;
    -webkit-box-pack: center;
    -ms-flex-pack: center;
    justify-content: center;
}

.list-item .last-word, .list-item .name {
    overflow: hidden;
    width: 155px;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.list-item .name {
    color: #333;
    font-size: 14px;
    min-height: 16px;
    line-height: 1;
}

.list-item .last-word {
    color: #999;
    padding: 8px 0;
    margin-bottom: -6px;
}

.list-item .close {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-pack: center;
    -ms-flex-pack: center;
    justify-content: center;
    -webkit-box-align: center;
    -ms-flex-align: center;
    align-items: center;
    width: 24px;
    -webkit-transform: translateX(-100%);
    transform: translateX(-100%);
    opacity: 0;
    -webkit-transition: 300ms;
    transition: 300ms;
    fill: #999;
}

.list-item:hover .close {
    -webkit-transform: translateX(0);
    transform: translateX(0);
    opacity: 1;
}

.notify {
    text-align: center;
    position: absolute;
    background-color: var(--stress_red);
}

.notify-number {
    color: #fff;
    line-height: 16px;
    font-size: 12px;
    border-radius: 8px;
    height: 16px;
    min-width: 16px;
    top: 20px;
    right: 8px;
    padding: 0 4px;
}

.right {
    width: calc(100% - 240px);
    -webkit-box-flex: 1;
    -ms-flex: 1;
    flex: 1;
}
</style>
