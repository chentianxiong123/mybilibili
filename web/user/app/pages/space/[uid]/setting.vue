<template>
    <div class="space-setting">
        <h3 class="page-head">
            <span class="t">空间设置</span>
        </h3>
        <div v-if="isOwner">
            <div class="settings-group">
                <h3 class="settings-group-title">隐私设置</h3>
                <div class="settings-list">
                    <div class="setting-item">
                        <span class="setting-label">公开我的收藏</span>
                        <div class="setting-control">
                            <el-switch
                                v-model="privacySettings.publicCollection"
                                :active-value="true"
                                :inactive-value="false"
                                active-text="公开"
                                inactive-text="隐藏"
                                @change="handlePrivacyChange('publicCollection', $event)"
                            />
                        </div>
                    </div>
                    <div class="setting-item">
                        <span class="setting-label">公开我的生日、个人标签</span>
                        <div class="setting-control">
                            <el-switch
                                v-model="privacySettings.publicBirthdayTags"
                                :active-value="true"
                                :inactive-value="false"
                                active-text="公开"
                                inactive-text="隐藏"
                                @change="handlePrivacyChange('publicBirthdayTags', $event)"
                            />
                        </div>
                    </div>
                    <div class="setting-item">
                        <span class="setting-label">公开我的关注列表</span>
                        <div class="setting-control">
                            <el-switch
                                v-model="privacySettings.publicFollowingList"
                                :active-value="true"
                                :inactive-value="false"
                                active-text="公开"
                                inactive-text="隐藏"
                                @change="handlePrivacyChange('publicFollowingList', $event)"
                            />
                        </div>
                    </div>
                    <div class="setting-item">
                        <span class="setting-label">公开我的粉丝列表</span>
                        <div class="setting-control">
                            <el-switch
                                v-model="privacySettings.publicFollowersList"
                                :active-value="true"
                                :inactive-value="false"
                                active-text="公开"
                                inactive-text="隐藏"
                                @change="handlePrivacyChange('publicFollowersList', $event)"
                            />
                        </div>
                    </div>
                </div>
            </div>

            <div class="settings-group">
                <h3 class="settings-group-title">我的个人标签</h3>
                <div class="tags-section">
                    <div class="tags-list">
                        <el-tag
                            v-for="tag in userTags"
                            :key="tag"
                            closable
                            class="user-tag"
                            @close="handleRemoveTag(tag)"
                        >
                            {{ tag }}
                        </el-tag>
                    </div>
                    <div class="tag-input-wrapper">
                        <el-input
                            v-model="newTagInput"
                            placeholder="输入标签名称"
                            maxlength="10"
                            show-word-limit
                            class="tag-input"
                            @keyup.enter="handleAddTag"
                        />
                        <el-button type="primary" @click="handleAddTag" :disabled="!newTagInput.trim()">
                            新增
                        </el-button>
                    </div>
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
import { userPrivacyApi } from '@/api/userPrivacy';
import { ElMessage } from 'element-plus';

export default {
    name: "SpaceSetting",
    data() {
        return {
            privacySettings: {
                publicCollection: true,
                publicBirthdayTags: false,
                publicFollowingList: false,
                publicFollowersList: false
            },
            userTags: [] as string[],
            newTagInput: '',
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
        async loadPrivacySettings() {
            if (!this.isOwner) return;
            try {
                const res = await userPrivacyApi.getPrivacySettings();
                if (res.code === 200) {
                    this.privacySettings = {
                        publicCollection: res.data.publicCollection ?? true,
                        publicBirthdayTags: res.data.publicBirthdayTags ?? false,
                        publicFollowingList: res.data.publicFollowingList ?? false,
                        publicFollowersList: res.data.publicFollowersList ?? false
                    };
                    this.userTags = res.data.tags || [];
                }
            } catch (error: any) {
                if (error?.response?.status !== 404) {
                    console.error('加载隐私设置失败:', error);
                }
            }
        },
        async handlePrivacyChange(key: string, value: boolean) {
            try {
                const data = { [key]: value };
                const res = await userPrivacyApi.updatePrivacySettings(data);
                if (res.code === 200) {
                    ElMessage.success('设置已保存');
                } else {
                    ElMessage.error(res.message || '保存失败');
                }
            } catch (error) {
                console.error('保存隐私设置失败:', error);
                ElMessage.error('保存失败');
            }
        },
        async handleAddTag() {
            const tagName = this.newTagInput.trim();
            if (!tagName) return;
            if (this.userTags.includes(tagName)) {
                ElMessage.warning('标签已存在');
                return;
            }
            if (this.userTags.length >= 10) {
                ElMessage.warning('最多只能添加10个标签');
                return;
            }
            try {
                const res = await userPrivacyApi.addUserTag(tagName);
                if (res.code === 200) {
                    this.userTags.push(tagName);
                    this.newTagInput = '';
                    ElMessage.success('添加成功');
                } else {
                    ElMessage.error(res.message || '添加失败');
                }
            } catch (error) {
                console.error('添加标签失败:', error);
                ElMessage.error('添加失败');
            }
        },
        async handleRemoveTag(tag: string) {
            try {
                const res = await userPrivacyApi.removeUserTag(tag);
                if (res.code === 200) {
                    this.userTags = this.userTags.filter(t => t !== tag);
                    ElMessage.success('删除成功');
                } else {
                    ElMessage.error(res.message || '删除失败');
                }
            } catch (error) {
                console.error('删除标签失败:', error);
                ElMessage.error('删除失败');
            }
        },
    },
    mounted() {
        if (this.isOwner) {
            this.loadPrivacySettings();
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

.settings-group {
    margin-bottom: 32px;
}

.settings-group-title {
    font-size: 16px;
    font-weight: 600;
    color: #333;
    margin: 0 0 20px 0;
    padding-bottom: 12px;
    border-bottom: 1px solid #f0f0f0;
}

.settings-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.setting-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
}

.setting-label {
    font-size: 14px;
    color: #333;
}

.setting-control {
    display: flex;
    align-items: center;
}

.tags-section {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.tags-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
}

.user-tag {
    font-size: 13px;
    padding: 6px 12px;
    border-radius: 4px;
}

.tag-input-wrapper {
    display: flex;
    gap: 10px;
    align-items: center;
}

.tag-input {
    width: 200px;
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
