<template>
    <div class="ai-chat">
        <div class="chat-header">
            <span class="header-title">AI客服</span>
            <span class="transfer" @click="transfer">转人工</span>
        </div>
        <div class="chat-body" ref="body">
            <div class="chat-loading" v-if="loading">加载历史记录…</div>
            <div class="chat-empty" v-else-if="messages.length === 0">你好，我是 AI 客服，有什么可以帮你？</div>
            <div class="msg" v-for="(m, i) in messages" :key="m.id || i" :class="m.role === 'user' ? 'is-me' : 'is-ai'">
                <div class="bubble" v-html="escapeHtml(m.content)"></div>
            </div>
            <div class="msg is-ai" v-if="streaming">
                <div class="bubble"><span class="typing">正在输入…</span></div>
            </div>
        </div>
        <div class="chat-input">
            <el-input v-model="input" placeholder="输入你的问题，回车发送" :disabled="sending"
                @keyup.enter="send" maxlength="500" show-word-limit />
            <el-button type="primary" :loading="sending" :disabled="!input.trim()" @click="send">发送</el-button>
        </div>
    </div>
</template>

<script lang="ts">
import { aiChatApi } from '@/api/aiChat.ts';

export default {
    name: 'AiChatWindow',
    data() {
        return {
            messages: [] as any[],
            input: '',
            loading: false,
            sending: false,
            streaming: false,
            streamHandle: null as any,
        };
    },
    methods: {
        escapeHtml(s: string) {
            return String(s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/\n/g, '<br/>');
        },
        scrollBottom() {
            this.$nextTick(() => {
                const el = this.$refs.body as HTMLElement;
                if (el) el.scrollTop = el.scrollHeight;
            });
        },
        async loadHistory() {
            this.loading = true;
            try {
                const convs = await aiChatApi.getConversations();
                if (convs && convs.code === 200 && Array.isArray(convs.data) && convs.data.length > 0) {
                    const res = await aiChatApi.getMessages(convs.data[0].id);
                    if (res && res.code === 200 && Array.isArray(res.data)) {
                        this.messages = res.data;
                    }
                }
            } catch (e) {
                // 历史拉失败不阻塞新会话，直接显示欢迎语
            } finally {
                this.loading = false;
                this.scrollBottom();
            }
        },
        send() {
            const content = this.input.trim();
            if (!content || this.sending) return;
            this.messages.push({ id: `u-${Date.now()}`, role: 'user', content });
            this.input = '';
            this.sending = true;
            this.streaming = true;
            this.scrollBottom();
            const aiMsg = { id: `a-${Date.now()}`, role: 'assistant', content: '' };
            this.messages.push(aiMsg);
            try {
                this.streamHandle = aiChatApi.sendMessage('customer-service', content, {
                    onData: (d: string) => {
                        this.streaming = false;
                        aiMsg.content += d || '';
                        this.scrollBottom();
                    },
                    onDone: () => {
                        this.sending = false;
                        this.streaming = false;
                        this.scrollBottom();
                    },
                    onTransfer: () => {
                        this.streaming = false;
                        aiMsg.content += '\n[已为你转接人工客服]';
                        this.sending = false;
                        this.scrollBottom();
                    },
                    onError: (err: string) => {
                        this.streaming = false;
                        aiMsg.content = aiMsg.content || `发送失败：${err || '网络错误'}`;
                        this.sending = false;
                        this.scrollBottom();
                    },
                });
            } catch (e) {
                aiMsg.content = '发送失败，请稍后重试';
                this.sending = false;
                this.streaming = false;
            }
        },
        async transfer() {
            try {
                await aiChatApi.transferToHuman();
                this.messages.push({ id: `s-${Date.now()}`, role: 'assistant', content: '已为你转接人工客服，请稍候…' });
                this.scrollBottom();
            } catch (e) {
                // 转人工失败静默，下方仍可继续 AI 对话
            }
        },
    },
    mounted() {
        this.loadHistory();
    },
    beforeUnmount() {
        if (this.streamHandle && this.streamHandle.abort) {
            this.streamHandle.abort();
        }
    },
};
</script>

<style scoped>
.ai-chat {
    background: #fff;
    border-radius: 4px;
    min-height: 500px;
    max-height: calc(100vh - 260px);
    display: flex;
    flex-direction: column;
}
.chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 48px;
    padding: 0 16px;
    border-bottom: 1px solid #f0f0f0;
}
.header-title {
    font-size: 14px;
    font-weight: 600;
    color: #333;
}
.transfer {
    font-size: 12px;
    color: var(--brand_pink, #fb7299);
    cursor: pointer;
}
.chat-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    min-height: 380px;
}
.chat-loading, .chat-empty {
    text-align: center;
    padding: 60px 0;
    color: #999;
    font-size: 13px;
}
.msg {
    display: flex;
    margin-bottom: 12px;
}
.msg.is-me {
    justify-content: flex-end;
}
.msg.is-ai {
    justify-content: flex-start;
}
.bubble {
    max-width: 75%;
    padding: 8px 12px;
    border-radius: 8px;
    font-size: 13px;
    line-height: 20px;
    word-break: break-all;
    white-space: pre-wrap;
}
.is-me .bubble {
    background: var(--brand_pink, #fb7299);
    color: #fff;
}
.is-ai .bubble {
    background: #f4f4f4;
    color: #333;
}
.typing {
    color: #999;
}
.chat-input {
    display: flex;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid #f0f0f0;
}
</style>
