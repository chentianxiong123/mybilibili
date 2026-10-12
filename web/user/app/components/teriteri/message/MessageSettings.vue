<template>
    <div class="msg-settings">
        <div class="settings-loading" v-if="loading">加载中…</div>
        <div class="settings-error" v-else-if="error">加载失败，<span class="btn" @click="load">点击重试</span></div>
        <template v-else>
            <!-- 分组：消息提醒 -->
            <div class="settings-group">
                <div class="group-name">消息提醒</div>
                <div class="setting-row" v-for="row in groupMain" :key="row.key">
                    <div class="row-text">
                        <div class="row-name">{{ row.name }}</div>
                        <div class="row-desc">{{ row.desc }}</div>
                    </div>
                    <div class="row-options">
                        <div class="opt" :class="{ active: form[row.key] === 1 }" @click="setVal(row.key, 1)">
                            <span class="opt-radio"></span>开启
                        </div>
                        <div class="opt" :class="{ active: form[row.key] === 0 }" @click="setVal(row.key, 0)">
                            <span class="opt-radio"></span>关闭
                        </div>
                    </div>
                </div>
            </div>
            <!-- 分组：提醒范围 -->
            <div class="settings-group">
                <div class="group-name">提醒范围</div>
                <div class="setting-row" v-for="row in groupRange" :key="row.key">
                    <div class="row-text">
                        <div class="row-name">{{ row.name }}</div>
                        <div class="row-desc">{{ row.desc }}</div>
                    </div>
                    <div class="row-options">
                        <div class="opt" :class="{ active: form[row.key] === 1 }" @click="setVal(row.key, 1)">
                            <span class="opt-radio"></span>开启
                        </div>
                        <div class="opt" :class="{ active: form[row.key] === 0 }" @click="setVal(row.key, 0)">
                            <span class="opt-radio"></span>关闭
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<script lang="ts">
import { messageApi } from '@/api/message.ts';

// 后端字段 -> 设置项（get/put /message/settings 原样透传）
// 按 B 站消息设置页布局分两组：消息提醒 / 提醒范围
const groupMain = [
    { key: 'private_message_notification', name: '私信消息提醒', desc: '收到新私信时提醒我' },
    { key: 'system_notification', name: '系统通知提醒', desc: '接收官方系统通知' },
];
const groupRange = [
    { key: 'reply_notification', name: '回复我的消息提醒', desc: '有人回复你的评论时提醒我' },
    { key: 'at_notification', name: '@我的消息提醒', desc: '有人在评论中@我时提醒我' },
    { key: 'like_notification', name: '收到的赞消息提醒', desc: '有人赞了你的视频或评论时提醒我' },
];

export default {
    name: 'MessageSettings',
    data() {
        return {
            groupMain,
            groupRange,
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
                for (const r of [...groupMain, ...groupRange]) {
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
.msg-settings {
    background: #fff;
    border-radius: 4px;
    min-height: 400px;
    padding: 0 24px 24px;
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
.settings-group {
    padding-top: 8px;
}
.settings-group + .settings-group {
    margin-top: 16px;
}
.group-name {
    font-size: 15px;
    font-weight: 600;
    color: #333;
    padding: 20px 0 8px;
    border-bottom: 1px solid #f0f0f0;
    margin-bottom: 4px;
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
.row-options {
    display: flex;
    align-items: center;
    gap: 24px;
    flex-shrink: 0;
}
.opt {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #999;
    cursor: pointer;
    user-select: none;
}
.opt-radio {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 1px solid #ddd;
    position: relative;
    box-sizing: border-box;
}
.opt.active {
    color: #333;
}
.opt.active .opt-radio {
    border-color: var(--brand_pink, #fb7299);
}
.opt.active .opt-radio::after {
    content: '';
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--brand_pink, #fb7299);
}
</style>
