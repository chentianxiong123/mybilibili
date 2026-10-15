<template>
    <div class="space-collections">
        <h3 class="page-head">
            <span class="t">{{ this.$store.state.user.uid === uid ? '我' : 'TA' }}的合集</span>
            <span class="count">{{ collectionCount }}</span>
            <div class="page-head__right" v-if="this.$store.state.user.uid === uid">
                <span class="create-btn" @click="openCreateDialog">
                    <i class="iconfont icon-add"></i>新建合集
                </span>
            </div>
        </h3>

        <!-- 合集详情视图 -->
        <div class="collection-detail" v-if="detailVisible && currCollection">
            <div class="detail-info">
                <div class="detail-cover">
                    <img v-if="currCollection.coverUrl" :src="currCollection.coverUrl" alt="">
                    <div class="cover-icon" v-else>
                        <i class="iconfont icon-bofangliang"></i>
                    </div>
                </div>
                <div class="detail-desc">
                    <a class="detail-name">{{ currCollection.name }}</a>
                    <div class="detail-meta">
                        <span>{{ currCollection.manuscriptCount }}个内容</span>
                        <span class="dot">·</span>
                        <span>{{ currCollection.isPublic ? '公开' : '私密' }}</span>
                        <span class="dot">·</span>
                        <span>创建于 {{ handleDate(currCollection.createTime) }}</span>
                    </div>
                    <div class="detail-actions">
                        <div class="play-all" @click="playAll">
                            <i class="iconfont icon-bofang"></i>
                            <span>播放全部</span>
                        </div>
                        <div class="more-actions" v-if="this.$store.state.user.uid === uid">
                            <span class="action-btn" @click="openEditDialog">编辑信息</span>
                            <span class="action-btn" @click="openAddVideoDialog">添加视频</span>
                            <span class="action-btn danger" @click="deleteCurrentCollection">删除合集</span>
                        </div>
                    </div>
                </div>
            </div>
            <div class="detail-sort">
                <div class="be-tab">
                    <ul class="be-tab-inner">
                        <li class="be-tab-item" :class="{'active': rule === 1}" @click="changeRule(1)">
                            <span>最新发布</span>
                        </li>
                        <li class="be-tab-item" :class="{'active': rule === 2}" @click="changeRule(2)">
                            <span>最多播放</span>
                        </li>
                        <li class="be-tab-item" :class="{'active': rule === 3}" @click="changeRule(3)">
                            <span>默认排序</span>
                        </li>
                    </ul>
                    <div class="be-tab-cursor" :style="`transform: translateX(${63 * (rule - 1)}px); width: 48px;`"></div>
                </div>
            </div>
            <div class="detail-list" v-loading="detailLoading">
                <div class="back-btn" @click="backToList"><i class="iconfont icon-youjiantou"></i> 返回合集列表</div>
                <ul class="collection-video-list clearfix" v-if="detailManuscripts.length > 0">
                    <li class="small-item" v-for="(item, index) in detailManuscripts" :key="index">
                        <a :href="`/manuscript/${item.id}`" target="_blank" class="cover" v-if="item.status === 3 || item.status === 1">
                            <img v-if="item.coverUrl" :src="item.coverUrl" alt="">
                            <span v-if="item.duration" class="length">{{ handleDuration(item.duration) }}</span>
                        </a>
                        <div class="disable-cover" v-else>
                            <p>视频已失效</p>
                        </div>
                        <a :href="`/manuscript/${item.id}`" target="_blank" class="title" :title="item.title" v-if="item.status === 3 || item.status === 1">{{ item.title }}</a>
                        <p class="title" v-else>{{ item.title }}</p>
                        <div class="meta">
                            <span class="play">
                                <i class="iconfont icon-bofangliang"></i>
                                <span class="num">{{ handleNum(item.viewCount) }}</span>
                            </span>
                            <span class="time">{{ handleDate(item.uploadTime) }}</span>
                            <span v-if="this.$store.state.user.uid === uid" class="remove" @click="removeManuscript(item.id)">移除</span>
                        </div>
                    </li>
                </ul>
                <el-empty v-else description="该合集暂无视频"></el-empty>
            </div>
        </div>

        <!-- 合集列表视图 -->
        <div class="collection-grid" v-show="!detailVisible">
            <div class="collection-list clearfix">
                <div class="collection-item" v-for="(item, index) in collections" :key="index" @click="goDetail(item.id)">
                    <div class="collection-cover">
                        <img v-if="item.cover_url || item.coverUrl" :src="item.cover_url || item.coverUrl" alt="">
                        <div class="cover-tint"></div>
                        <span class="cover-count"><i class="iconfont icon-shipin"></i>{{ item.manuscript_count || item.manuscriptCount || 0 }}个视频</span>
                    </div>
                    <a class="collection-name" :title="item.title || item.name">{{ item.title || item.name }}</a>
                    <div class="collection-meta">
                        <span>{{ item.manuscript_count || item.manuscriptCount || 0 }}个内容</span>
                        <span class="dot">·</span>
                        <span>创建于 {{ handleDate(item.created_at || item.createTime) }}</span>
                    </div>
                </div>
                <el-empty v-if="collections.length === 0" description="暂无合集"></el-empty>
            </div>
        </div>

        <!-- 新建/编辑合集 -->
        <el-dialog v-model="showEditDialog" :title="editDialogTitle" width="420px" :close-on-click-modal="false">
            <el-form :model="editForm" label-width="72px">
                <el-form-item label="合集名">
                    <el-input v-model="editForm.name" maxlength="40" placeholder="请输入合集名称"></el-input>
                </el-form-item>
                <el-form-item label="简介">
                    <el-input v-model="editForm.description" type="textarea" :rows="2" maxlength="200" placeholder="补充说明（可选）"></el-input>
                </el-form-item>
                <el-form-item label="可见性">
                    <el-radio-group v-model="editForm.isPublic">
                        <el-radio :value="true">公开</el-radio>
                        <el-radio :value="false">私密</el-radio>
                    </el-radio-group>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="showEditDialog = false">取消</el-button>
                <el-button type="primary" :loading="saving" @click="saveCollection">确定</el-button>
            </template>
        </el-dialog>

        <!-- 添加视频到合集 -->
        <el-dialog v-model="showAddVideoDialog" title="添加视频" width="640px" :close-on-click-modal="false">
            <div class="add-video-list" v-loading="videoLoading">
                <div v-for="(item, index) in myVideos" :key="index" class="add-video-item" :class="{'selected': selectedVideos.includes(item.id)}" @click="toggleSelectVideo(item.id)">
                    <img v-if="item.coverUrl" :src="item.coverUrl" alt="" class="av-cover">
                    <div class="av-info">
                        <p class="av-title">{{ item.title }}</p>
                        <p class="av-meta">{{ handleNum(item.viewCount) }}播放 · {{ item.durationSeconds ? handleDuration(item.durationSeconds) : '' }}</p>
                    </div>
                    <i class="iconfont" :class="selectedVideos.includes(item.id) ? 'icon-duigou' : 'icon-weixuanzhong'"></i>
                </div>
                <el-empty v-if="myVideos.length === 0" description="暂无已发布视频"></el-empty>
            </div>
            <template #footer>
                <el-button @click="showAddVideoDialog = false">取消</el-button>
                <el-button type="primary" :loading="adding" @click="addSelectedVideos">添加所选</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts">
import { hasAuthSession } from "@/utils/auth.ts"
import { handleTime, handleNum, handleDate } from '@/teriteri-src/utils/utils';
import { ElMessage, ElMessageBox } from 'element-plus';

export default {
    name: "SpaceCollections",
    data() {
        return {
            collections: [],          // 合集列表(raw)
            collectionCount: 0,
            detailVisible: false,
            currCollection: null,
            detailManuscripts: [],
            detailLoading: false,
            rule: 1,                  // 1 最新 2 最多播放 3 默认
            activeCollectionId: null,
            showEditDialog: false,
            editDialogTitle: '',
            editForm: { name: '', description: '', isPublic: true, cover: null },
            editingCollectionId: null,
            saving: false,
            showAddVideoDialog: false,
            myVideos: [],
            videoLoading: false,
            selectedVideos: [],
            adding: false,
        }
    },
    computed: {
        uid() {
            return Number(this.$route.params.uid);
        }
    },
    methods: {
        // 获取用户的合集列表
        async getCollections() {
            const res = await this.$get(`/collection/user/${this.uid}`);
            if (!res.data || !res.data.data) return;
            const list = res.data.data.list || [];
            this.collections = list;
            this.collectionCount = list.length;
        },

        // 获取稿件列表（为合集封面兜底：若合集无封面上传，用第一稿封面）
        async getFirstVideoCover(collection) {
            if (collection.cover_url || collection.coverUrl) return;
            const res = await this.$get(`/collection/${collection.id}/manuscripts?page=1&pageSize=1`);
            if (res.data && res.data.data) {
                const list = res.data.data.list || [];
                if (list[0] && list[0].cover_url) collection.cover_url = list[0].cover_url;
            }
        },

        // 进入合集详情
        goDetail(id) {
            this.detailVisible = true;
            this.activeCollectionId = id;
            this.rule = 1;
            this.loadDetail();
        },

        backToList() {
            this.detailVisible = false;
            this.activeCollectionId = null;
            this.currCollection = null;
            this.detailManuscripts = [];
            this.getCollections();
        },

        // 加载合集详情
        async loadDetail() {
            if (!this.activeCollectionId) return;
            this.detailLoading = true;
            // 基本信息
            const res = await this.$get(`/collection/${this.activeCollectionId}`);
            if (res.data && res.data.data) {
                const c = res.data.data;
                this.currCollection = {
                    id: c.id,
                    name: c.title || c.name,
                    coverUrl: c.cover_url || c.coverUrl,
                    manuscriptCount: c.manuscript_count || c.manuscriptCount || 0,
                    isPublic: c.status === 1,
                    createTime: c.created_at || c.updated_at || '',
                };
            }
            // 稿件列表
            const page = 1, size = 100;
            const resList = await this.$get(`/collection/${this.activeCollectionId}/manuscripts?page=${page}&pageSize=${size}`);
            if (resList.data && resList.data.data) {
                this.detailManuscripts = (resList.data.data.list || []).map((m: any) => ({
                    id: m.id,
                    title: m.title,
                    coverUrl: m.cover_url || m.coverUrl,
                    viewCount: m.view_count || m.viewCount || 0,
                    uploadTime: m.upload_time || m.created_at || '',
                    duration: m.duration || m.durationSeconds || 0,
                    status: m.status || 0,
                }));
                this.sortManuscripts(false);
            }
            this.detailLoading = false;
        },

        // 详情排序
        changeRule(r) {
            this.rule = r;
            this.sortManuscripts(true);
        },

        sortManuscripts(byRule) {
            if (!byRule || this.rule === 3) return;
            const list = [...this.detailManuscripts];
            if (this.rule === 1) {
                list.sort((a, b) => new Date(b.uploadTime as any) - new Date(a.uploadTime as any));
            } else if (this.rule === 2) {
                list.sort((a, b) => (b.viewCount as any) - (a.viewCount as any));
            }
            this.detailManuscripts = list;
        },

        // 播放全部
        playAll() {
            if (this.detailManuscripts.length > 0) {
                window.open(`/manuscript/${this.detailManuscripts[0].id}`, '_blank');
            } else {
                ElMessage.info('该合集暂无视频');
            }
        },

        // 从合集移除视频
        async removeManuscript(manuscriptId) {
            try {
                await ElMessageBox.confirm('确定将该视频移出合集？', '提示', { type: 'warning' });
            } catch (e) {
                return;
            }
            const res = await this.$delete(`/collection/${this.activeCollectionId}/manuscript/${manuscriptId}`);
            if (res.data && res.data.code === 200) {
                ElMessage.success('移除成功');
                this.loadDetail();
            } else {
                ElMessage.error('移除失败');
            }
        },

        // 新建合集
        openCreateDialog() {
            this.editDialogTitle = '新建合集';
            this.editingCollectionId = null;
            this.editForm = { name: '', description: '', isPublic: true, cover: null };
            this.showEditDialog = true;
        },

        // 编辑合集
        openEditDialog() {
            this.editDialogTitle = '编辑合集';
            this.editingCollectionId = this.activeCollectionId;
            this.editForm = {
                name: this.currCollection ? this.currCollection.name : '',
                description: this.currCollection && this.currCollection.description ? this.currCollection.description : '',
                isPublic: this.currCollection ? this.currCollection.isPublic : true,
                cover: null,
            };
            this.showEditDialog = true;
        },

        // 保存合集（新建/编辑共用）
        async saveCollection() {
            if (!this.editForm.name.trim()) {
                ElMessage.warning('请输入合集名称');
                return;
            }
            this.saving = true;
            try {
                const formData = new FormData();
                formData.append('name', this.editForm.name.trim());
                formData.append('description', this.editForm.description || '');
                formData.append('isPublic', this.editForm.isPublic === false ? 'false' : 'true');
                let res;
                if (this.editingCollectionId) {
                    res = await this.$put(`/collection/${this.editingCollectionId}`, formData);
                } else {
                    res = await this.$post('/collection', formData);
                }
                if (res.data && res.data.code === 200) {
                    ElMessage.success(this.editingCollectionId ? '编辑成功' : '创建成功');
                    this.showEditDialog = false;
                    await this.getCollections();
                    if (this.editingCollectionId) this.loadDetail();
                } else {
                    ElMessage.error((res.data && res.data.message) || '保存失败');
                }
            } catch (e) {
                console.error('保存合集失败:', e);
                ElMessage.error('保存失败');
            }
            this.saving = false;
        },

        // 删除当前合集
        async deleteCurrentCollection() {
            try {
                await ElMessageBox.confirm('删除后不可恢复，确定删除该合集？', '删除合集', { type: 'warning' });
            } catch (e) {
                return;
            }
            const res = await this.$delete(`/collection/${this.activeCollectionId}`);
            if (res.data && res.data.code === 200) {
                ElMessage.success('删除成功');
                this.backToList();
            } else {
                ElMessage.error('删除失败');
            }
        },

        // 打开添加视频对话框
        async openAddVideoDialog() {
            this.showAddVideoDialog = true;
            this.selectedVideos = [];
            await this.getMyVideos();
        },

        // 获取自己的已发布稿件
        async getMyVideos() {
            this.videoLoading = true;
            const res = await this.$get(`/video/user-works`, {
                params: { uid: this.$store.state.user.uid, rule: 1, page: 1, quantity: 50 }
            });
            if (res.data && res.data.data) {
                this.myVideos = (res.data.data.list || []).map((item: any) => {
                    const v = item.video;
                    return {
                        id: Number(v.vid),
                        title: v.title,
                        coverUrl: v.coverUrl,
                        viewCount: v.play,
                        durationSeconds: v.duration,
                        alreadyIn: this.detailManuscripts.some((m: any) => m.id === Number(v.vid)),
                    };
                }).filter((i: any) => !i.alreadyIn);
            }
            this.videoLoading = false;
        },

        toggleSelectVideo(id) {
            const idx = this.selectedVideos.indexOf(id);
            if (idx >= 0) this.selectedVideos.splice(idx, 1);
            else this.selectedVideos.push(id);
        },

        // 批量添加所选视频
        async addSelectedVideos() {
            if (this.selectedVideos.length === 0) {
                ElMessage.warning('请先选择视频');
                return;
            }
            this.adding = true;
            for (let i = 0; i < this.selectedVideos.length; i++) {
                await this.$post(`/collection/${this.activeCollectionId}/manuscript/${this.selectedVideos[i]}?order=${i}`);
            }
            ElMessage.success('添加成功');
            this.showAddVideoDialog = false;
            this.adding = false;
            this.loadDetail();
        },

        handleDuration(time) {
            return handleTime(time);
        },

        handleNum(number) {
            return handleNum(number);
        },

        handleDate(date) {
            return handleDate(date);
        },
    },
    async mounted() {
        await this.getCollections();
        // 为未上传封面的合集兜底拉第一稿封面
        for (const c of this.collections) {
            this.getFirstVideoCover(c);
        }
    },
    watch: {
        "$route.params.uid"() {
            this.getCollections();
        }
    }
}
</script>

<style scoped>
ol, ul {
    list-style: none;
}

.space-collections * {
    box-sizing: initial;
}

.page-head {
    line-height: 33px;
    padding: 0 0 15px;
    position: relative;
}
.page-head .t {
    font-size: 20px;
    color: var(--text1);
    font-weight: 400;
}
.page-head .count {
    background: #f6fafb;
    border: 1px solid #ddd;
    border-radius: 3px;
    color: #777;
    display: inline-block;
    font-size: 12px;
    line-height: 18px;
    height: 18px;
    margin-left: 11px;
    padding: 0 5px;
    position: relative;
    vertical-align: middle;
}
.page-head__right {
    position: absolute;
    right: 0;
    top: 0;
}
.create-btn {
    border: 1px solid #b8c0cc;
    border-radius: 4px;
    color: #6d757a;
    font-size: 12px;
    line-height: 22px;
    padding: 0 10px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
}
.create-btn:hover {
    border-color: var(--Pi4);
    color: var(--Pi4);
}
.create-btn .icon-add {
    font-size: 12px;
    margin-right: 3px;
}

/* 合集列表 */
.collection-list {
    display: flex;
    flex-wrap: wrap;
    margin-right: -20px;
}
.collection-item {
    width: 170px;
    margin: 0 20px 28px 0;
    cursor: pointer;
}
.collection-cover {
    width: 170px;
    height: 96px;
    border-radius: 4px;
    overflow: hidden;
    position: relative;
    background: url('@/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
    background-size: cover;
    border: 1px solid #eeeff1;
}
.collection-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
}
.cover-tint {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 28px;
    background: linear-gradient(360deg, rgba(0, 0, 0, .6), transparent);
}
.cover-count {
    position: absolute;
    right: 6px;
    bottom: 5px;
    color: #fff;
    font-size: 12px;
    line-height: 18px;
}
.cover-count .icon-shipin {
    font-size: 12px;
    margin-right: 2px;
}
.collection-name {
    display: block;
    line-height: 20px;
    height: 38px;
    margin-top: 7px;
    font-size: 14px;
    color: var(--text1);
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    overflow: hidden;
    text-overflow: ellipsis;
    word-break: break-word;
}
.collection-name:hover {
    color: var(--brand_pink);
}
.collection-meta {
    color: #99a2aa;
    font-size: 12px;
    line-height: 18px;
    margin-top: 2px;
}
.collection-meta .dot {
    margin: 0 4px;
}

/* 详情视图 */
.detail-info {
    display: flex;
    margin-bottom: 20px;
}
.detail-cover {
    width: 160px;
    height: 90px;
    border-radius: 4px;
    overflow: hidden;
    position: relative;
    background: url('@/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
    background-size: cover;
    flex-shrink: 0;
}
.detail-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.detail-desc {
    margin-left: 16px;
    flex: 1;
}
.detail-name {
    font-size: 20px;
    color: var(--text1);
    line-height: 28px;
    display: block;
    font-weight: 400;
}
.detail-meta {
    color: #9499a0;
    font-size: 12px;
    line-height: 20px;
    margin-top: 6px;
}
.detail-meta .dot {
    margin: 0 4px;
}
.detail-actions {
    margin-top: 14px;
}
.play-all {
    display: inline-flex;
    align-items: center;
    background: var(--brand_pink);
    color: #fff;
    border-radius: 4px;
    padding: 4px 14px;
    font-size: 13px;
    cursor: pointer;
}
.play-all .icon-bofang {
    font-size: 16px;
    margin-right: 4px;
}
.play-all:hover {
    opacity: .9;
}
.more-actions {
    display: inline-block;
    margin-left: 14px;
}
.action-btn {
    color: var(--text2);
    font-size: 12px;
    margin-right: 12px;
    cursor: pointer;
}
.action-btn:hover {
    color: var(--brand_pink);
}
.action-btn.danger:hover {
    color: #00a1d6;
}

.detail-sort {
    border-bottom: 1px solid #eee;
    margin-bottom: 10px;
    position: relative;
}
.be-tab {
    position: relative;
    font-family: PingFang SC, Helvetica Neue, Microsoft YaHei, sans-serif;
}
.be-tab-item {
    float: left;
    font-size: 14px;
    margin-right: 20px;
    line-height: 28px;
    cursor: pointer;
    padding-bottom: 3px;
}
.be-tab-item.active {
    color: var(--brand_pink);
}
.be-tab-cursor {
    position: absolute;
    bottom: 0;
    left: 0;
    border-bottom: 1px solid var(--brand_pink);
    transition: width .2s ease, transform .2s ease;
}

.back-btn {
    color: var(--text2);
    font-size: 13px;
    cursor: pointer;
    margin-bottom: 10px;
    display: inline-block;
}
.back-btn:hover {
    color: var(--brand_pink);
}
.back-btn .icon-youjiantou {
    transform: rotate(180deg);
    display: inline-block;
    font-size: 13px;
}

.collection-video-list {
    display: flex;
    flex-wrap: wrap;
}
.small-item {
    display: block;
    float: left;
    width: 160px;
    position: relative;
    margin: 0 0 14px;
    padding: 10px;
}
.small-item:nth-child(4n+1) {
    padding-left: 0;
}
.small-item:nth-child(4n+4) {
    padding-right: 0;
}
.small-item .cover {
    background: url('@/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
    background-size: cover;
    border-radius: 4px;
    display: block;
    width: 160px;
    height: 100px;
    overflow: hidden;
    position: relative;
}
.small-item .cover img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.small-item .length {
    background: rgba(0, 0, 0, .5);
    color: #fff;
    line-height: 18px;
    font-size: 12px;
    padding: 0 4px;
    position: absolute;
    right: 4px;
    bottom: 6px;
    border-radius: 2px;
}
.small-item .title {
    display: block;
    line-height: 20px;
    height: 38px;
    margin-top: 6px;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    overflow: hidden;
    text-overflow: ellipsis;
    word-break: break-word;
    font-size: 12px;
    color: var(--text1);
}
.small-item .title:hover {
    color: var(--brand_pink);
}
.small-item .meta {
    color: #999;
    white-space: nowrap;
    font-size: 0;
    margin-top: 3px;
    height: 16px;
    line-height: 16px;
    display: flex;
    align-items: center;
}
.small-item .meta > span {
    display: inline-block;
    white-space: nowrap;
    height: 16px;
    line-height: 18px;
    font-size: 12px;
    overflow: hidden;
}
.small-item .meta .play {
    display: inline-flex;
    align-items: center;
    width: 92px;
}
.small-item .meta .num {
    margin-left: 3px;
}
.small-item .meta .time {
    width: 60px;
}
.small-item .meta .remove {
    color: var(--brand_pink);
    cursor: pointer;
    visibility: hidden;
}
.small-item:hover .meta .remove {
    visibility: visible;
}

/* 添加视频对话框 */
.add-video-list {
    max-height: 360px;
    overflow-y: auto;
}
.add-video-item {
    display: flex;
    align-items: center;
    padding: 8px;
    border-radius: 4px;
    cursor: pointer;
}
.add-video-item:hover {
    background: #f4f5f7;
}
.add-video-item.selected {
    background: #f1fbff;
}
.add-video-item .av-cover {
    width: 80px;
    height: 50px;
    object-fit: cover;
    border-radius: 2px;
    flex-shrink: 0;
    background: url('@/assets/teriteri/img/bilibili/video-placeholder.png') 50%;
    background-size: cover;
}
.add-video-item .av-info {
    flex: 1;
    margin-left: 10px;
    overflow: hidden;
}
.add-video-item .av-title {
    font-size: 13px;
    color: var(--text1);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.add-video-item .av-meta {
    font-size: 12px;
    color: #99a2aa;
    margin-top: 2px;
}
.add-video-item .icon-duigou {
    color: var(--brand_pink);
    font-size: 18px;
}
.add-video-item .icon-weixuanzhong {
    color: #ccc;
    font-size: 18px;
}
</style>