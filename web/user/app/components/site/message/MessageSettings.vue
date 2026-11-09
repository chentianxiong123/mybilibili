<template>
    <div class="config">
        <div class="settings-loading" v-if="loading">加载中…</div>
        <div class="settings-error" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
        <template v-else>
            <!-- 每个分组一张 B站 .message-section 白卡，内含若干 config-block/config-section -->
            <div class="message-section message-box-shadow" v-for="group in groups" :key="group.title">
                <div class="message-section__body">
                    <div class="config-block" v-for="row in group.rows" :key="row.key">
                        <section class="config-section">
                            <div class="config-section__header">
                                <div class="config-section__title">{{ row.name }}</div>
                                <div class="config-section__desc" v-if="row.desc">{{ row.desc }}</div>
                            </div>
                            <div class="config-section__body">
                                <div class="m-radio-group">
                                    <label class="m-radio" :class="{ checked: form[row.key] === 1 }" @click="setVal(row.key, 1)">
                                        <span class="m-radio__inner"></span>
                                        <span class="m-radio__label">开启</span>
                                    </label>
                                    <label class="m-radio" :class="{ checked: form[row.key] === 0 }" @click="setVal(row.key, 0)">
                                        <span class="m-radio__inner"></span>
                                        <span class="m-radio__label">关闭</span>
                                    </label>
                                </div>
                            </div>
                        </section>
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<script lang="ts">
import { messageApi } from '@/api/message.ts';

// 后端字段 -> 设置项（get/put /message/settings 原样透传）
// 按 B 站消息设置页布局分组：每张白卡一个分组，组内每项一个 config-block
const groups = [
    {
        title: '消息提醒',
        rows: [
            { key: 'private_message_notification', name: '私信消息提醒', desc: '（关闭后，收到新私信将不再进行提醒）' },
            { key: 'system_notification', name: '系统通知提醒', desc: '（关闭后，将不再接收官方系统通知）' },
        ],
    },
    {
        title: '消息提醒范围',
        rows: [
            { key: 'reply_notification', name: '回复我的消息提醒', desc: '（接收谁的评论消息提醒）' },
            { key: 'at_notification', name: '@我的消息提醒', desc: '（接收谁的@消息提醒）' },
            { key: 'like_notification', name: '收到的赞消息提醒', desc: '' },
        ],
    },
];

const ALL_ROWS = groups.flatMap(g => g.rows);

export default {
    name: 'MessageSettings',
    data() {
        return {
            groups,
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
                for (const r of ALL_ROWS) {
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
        setVal(key: string, val: number) {
            this.form[key] = val;
            this.save();
        },
    },
    mounted() {
        this.load();
    },
};
</script>

<style scoped>
.settings-loading, .settings-error {
    text-align: center;
    padding: 60px 0;
    color: #999;
    font-size: 13px;
}
.settings-error .btn {
    color: var(--brand_blue, #00a1d6);
    cursor: pointer;
}
</style>
