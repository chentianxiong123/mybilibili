<template>
    <div class="msg-settings">
        <div class="settings-title">消息设置</div>
        <div class="settings-loading" v-if="loading">加载中…</div>
        <div class="settings-error" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
        <div class="setting-row" v-else v-for="row in rows" :key="row.key">
            <div class="row-text">
                <div class="row-name">{{ row.name }}</div>
                <div class="row-desc">{{ row.desc }}</div>
            </div>
            <el-switch v-model="form[row.key]" :active-value="1" :inactive-value="0" @change="save" />
        </div>
    </div>
</template>

<script lang="ts">
import { messageApi } from '@/api/message.ts';

// 后端字段 -> 开关行（get/put /message/settings 原样透传）
const rows = [
    { key: 'private_message_notification', name: '私信', desc: '有人给你发私信时通知我' },
    { key: 'reply_notification', name: '回复我的', desc: '有人回复你的评论时通知我' },
    { key: 'at_notification', name: '@ 我的', desc: '有人在评论中@我时通知我' },
    { key: 'like_notification', name: '收到的赞', desc: '有人赞了你的视频或评论时通知我' },
    { key: 'system_notification', name: '系统通知', desc: '接收官方系统通知' },
];

export default {
    name: 'MessageSettings',
    data() {
        return {
            rows,
            form: {
                private_message_notification: 1,
                reply_notification: 1,
                at_notification: 1,
                like_notification: 1,
                system_notification: 1,
            } as Record<string, number>,
            loading: false,
            error: false,
        };
    },
    methods: {
        // 注意：后端 settings 接口直接返回设置对象（无 code 包裹）
        async load() {
            this.loading = true;
            this.error = false;
            try {
                const res = await messageApi.getMessageSettings();
                const d = (res && (res.data || res)) || {};
                for (const r of rows) {
                    if (d[r.key] === 0 || d[r.key] === 1) {
                        this.form[r.key] = d[r.key];
                    }
                }
            } catch (e) {
                this.error = true;
            } finally {
                this.loading = false;
            }
        },
        async save() {
            try {
                await messageApi.updateMessageSettings({ ...this.form });
            } catch (e) {
                // 保存失败下次进页恢复，后端 PUT 同样接受部分字段
            }
        },
    },
    mounted() {
        this.load();
    },
};
</script>

<style scoped>
.msg-settings {
    background: #fff;
    border-radius: 4px;
    min-height: 400px;
    padding: 0 20px 20px;
}
.settings-title {
    height: 48px;
    line-height: 48px;
    font-size: 14px;
    font-weight: 600;
    color: #333;
    border-bottom: 1px solid #f0f0f0;
}
.settings-loading, .settings-error {
    text-align: center;
    padding: 60px 0;
    color: #999;
    font-size: 13px;
}
.settings-error .btn {
    color: var(--brand_pink, #fb7299);
    cursor: pointer;
}
.setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    border-bottom: 1px solid #f4f4f4;
}
.setting-row:last-child {
    border-bottom: none;
}
.row-name {
    font-size: 14px;
    color: #333;
}
.row-desc {
    font-size: 12px;
    color: #999;
    margin-top: 4px;
}
</style>
