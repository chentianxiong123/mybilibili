<template>
    <div class="space-setting">
        <h3 class="page-head">
            <span class="t">空间设置</span>
        </h3>
        <div class="setting-form" v-if="isOwner">
            <div class="form-item">
                <label class="form-label">头像</label>
                <div class="form-content">
                    <div class="avatar-row">
                        <img class="avatar-preview" :src="user.avatar_url || defaultAvatar" alt="">
                        <a class="change-avatar" href="/account/avatar" target="_blank">更换头像</a>
                    </div>
                </div>
            </div>
            <div class="form-item">
                <label class="form-label">昵称</label>
                <div class="form-content">
                    <el-input v-model="nickname" maxlength="24" placeholder="请输入昵称" style="max-width: 320px;"/>
                </div>
            </div>
            <div class="form-item">
                <label class="form-label">简介</label>
                <div class="form-content">
                    <el-input v-model="description" type="textarea" :rows="3" maxlength="120" placeholder="介绍一下你自己" style="max-width: 480px;"/>
                    <p class="tip">当前版本暂未开放简介修改接口，仅作展示</p>
                </div>
            </div>
            <div class="form-item">
                <label class="form-label"></label>
                <div class="form-content">
                    <el-button type="primary" :loading="saving" @click="saveSetting">保存</el-button>
                </div>
            </div>
        </div>
        <div class="not-owner" v-else>
            <i class="iconfont icon-meiyoupinleimu"></i>
            <p>只能修改自己的空间设置</p>
        </div>
    </div>
</template>

<script lang="ts">
import { userApi } from '@/api/client';
import { ElMessage } from 'element-plus';

export default {
    name: "SpaceSetting",
    data() {
        return {
            nickname: '',
            description: '',
            saving: false,
            defaultAvatar: 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
        }
    },
    computed: {
        uid() {
            return Number(this.$route.params.uid);
        },
        user() {
            return this.$store.state.user;
        },
        isOwner() {
            return String(this.user.uid) === String(this.uid);
        }
    },
    methods: {
        async getUserInfo() {
            try {
                const res = await userApi.getUserById(this.uid);
                if (res.code === 200) {
                    this.nickname = res.data.nickname || res.data.username || '';
                    this.description = res.data.signature || res.data.bio || res.data.introduction || '';
                }
            } catch (e) {
                console.error('获取用户信息失败:', e);
            }
        },
        async saveSetting() {
            if (!this.nickname.trim()) {
                ElMessage.warning('昵称不能为空');
                return;
            }
            this.saving = true;
            try {
                const res = await userApi.updateUser(this.uid, { nickname: this.nickname.trim() });
                if (res.code === 200) {
                    ElMessage.success('保存成功');
                } else {
                    ElMessage.error(res.message || '保存失败');
                }
            } catch (e) {
                ElMessage.error('保存失败');
            } finally {
                this.saving = false;
            }
        },
    },
    created() {
        if (this.isOwner) {
            this.getUserInfo();
        }
    }
}
</script>

<style scoped>
.space-setting {
    background: #fff;
    padding: 20px;
    min-height: 400px;
    box-shadow: 0 0 0 1px #eee;
    border-radius: 4px;
}

.page-head {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 16px;
    border-bottom: 1px solid #eee;
    margin-bottom: 20px;
}

.page-head .t {
    font-size: 17px;
    color: #222;
    font-weight: 500;
}

.form-item {
    display: flex;
    align-items: flex-start;
    margin-bottom: 24px;
}

.form-label {
    width: 80px;
    line-height: 32px;
    color: #61666d;
    font-size: 14px;
    flex-shrink: 0;
}

.avatar-row {
    display: flex;
    align-items: center;
    gap: 16px;
}

.avatar-preview {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    object-fit: cover;
}

.change-avatar {
    color: var(--brand_pink);
    font-size: 14px;
    text-decoration: none;
}

.change-avatar:hover {
    text-decoration: underline;
}

.tip {
    margin: 8px 0 0;
    font-size: 12px;
    color: #99a2aa;
}

.not-owner {
    text-align: center;
    padding: 80px 0;
    color: #99a2aa;
}

.not-owner .iconfont {
    font-size: 60px;
    color: #d8dde3;
    display: block;
    margin-bottom: 16px;
}
</style>