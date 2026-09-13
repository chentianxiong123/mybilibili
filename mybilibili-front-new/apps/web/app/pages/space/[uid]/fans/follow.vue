<template>
    <div class="space-follow">
        <h3 class="page-head">
            <span class="t">TA的关注</span>
            <span class="count">{{ followCount }}</span>
        </h3>
        <div class="follow-wrap" v-loading="loading">
            <template v-if="!loading">
                <FansList
                    :users="followingList"
                    :loading="loading"
                    title="全部关注"
                    @follow="handleFollow"
                    @unfollow="handleUnfollow"
                />
                <div v-if="!loading && followingList.length === 0" class="empty-tip">
                    <i class="iconfont icon-meiyoupinleimu"></i>
                    <p>TA还没有关注任何人</p>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
import FansList from '@/components/FansList.vue';
import { userApi } from '@/api/client';
import { ElMessage } from 'element-plus';

export default {
    name: "SpaceFollow",
    components: {
        FansList,
    },
    data() {
        return {
            followingList: [],
            followCount: 0,
            loading: false,
        }
    },
    computed: {
        uid() {
            return Number(this.$route.params.uid);
        },
        currentUid() {
            return this.$store.state.user.uid;
        }
    },
    methods: {
        async getFollowingList() {
            this.loading = true;
            try {
                const res = await userApi.getFollowingList(this.uid);
                if (res.code === 200) {
                    this.followingList = (res.data || []).map(user => ({
                        id: user.id,
                        nickname: user.nickname || user.username || '用户',
                        avatar: user.avatar || 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png',
                        signature: user.signature || user.bio || '这个人很懒，什么都没写',
                        // 当前用户看别人空间：是否已关注该用户
                        isFollowing: this.currentUid === this.uid ? true : false,
                    }));
                    this.followCount = this.followingList.length;
                } else {
                    this.followingList = [];
                    this.followCount = 0;
                }
            } catch (e) {
                console.error('获取关注列表失败:', e);
                this.followingList = [];
                this.followCount = 0;
            } finally {
                this.loading = false;
            }
        },
        async handleFollow(userId) {
            try {
                const res = await userApi.follow(userId, true);
                if (res.code === 200) {
                    const user = this.followingList.find(u => u.id === userId);
                    if (user) user.isFollowing = true;
                    ElMessage.success('关注成功');
                } else {
                    ElMessage.error(res.message || '操作失败');
                }
            } catch (e) {
                ElMessage.error('操作失败');
            }
        },
        async handleUnfollow(userId) {
            try {
                const res = await userApi.follow(userId, false);
                if (res.code === 200) {
                    const user = this.followingList.find(u => u.id === userId);
                    if (user) user.isFollowing = false;
                    ElMessage.success('已取消关注');
                } else {
                    ElMessage.error(res.message || '操作失败');
                }
            } catch (e) {
                ElMessage.error('操作失败');
            }
        },
    },
    watch: {
        "$route.params.uid"() {
            this.getFollowingList();
        }
    },
    created() {
        this.getFollowingList();
    }
}
</script>

<style scoped>
.space-follow {
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

.page-head .count {
    font-size: 14px;
    color: #99a2aa;
    background: #f4f5f7;
    border-radius: 2px;
    padding: 0 8px;
    line-height: 20px;
}

.follow-wrap {
    min-height: 300px;
}

.empty-tip {
    text-align: center;
    padding: 80px 0;
    color: #99a2aa;
}

.empty-tip .iconfont {
    font-size: 60px;
    color: #d8dde3;
    display: block;
    margin-bottom: 16px;
}
</style>