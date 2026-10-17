<template>
    <div class="videoUpload">
        <!-- 投稿类型 tab（视频/图文） -->
        <UploadTypeTabs active="video" @change="switchUploadType" />
        <!-- 选择文件界面 -->
        <div class="upload-body-content" :style="selectedVideo ? 'display: none;' : ''">
            <div class="upload-wrp">
                <div class="video-input">
                    <div
                        class="video-input-wrapper"
                        @dragover.prevent
                        @dragenter.prevent
                        @dragleave.prevent
                        @drop.prevent
                        @drop="handleDrop"
                        @click="selectVideo"
                    >
                        <!-- <i class="iconfont icon-shangchuan"></i> -->
                        <el-icon size="40"><UploadFilled /></el-icon>
                        <div class="upload-tips-text">点击上传，或将视频拖拽到此处</div>
                        <div class="upload-btn">上传视频</div>
                        <input
                            type="file"
                            accept=".mp4"
                            multiple
                            ref="videoInput"
                            @change="handleVideoChange"
                            style="display: none;"
                        >
                    </div>
                </div>
            </div>
            <div class="footer-item info-grid">
                <span class="i-list i-1" @click="showUploadInfo('size')">
                    <span class="title">视频大小</span>
                    <span class="info-icon">&#63;</span>
                    <span class="info-pop-wrap" :class="activeInfo === 'size' ? 'info-pop-show' : ''">
                        <span class="info-pop-title">视频大小</span>
                        <span class="info-pop-text">网页端上传的文件大小上限为300M</span>
                        <span class="info-pop-text">视频内容时长最大10小时</span>
                    </span>
                </span>
                <span class="i-list i-1" @click="showUploadInfo('format')">
                    <span class="title">视频格式</span>
                    <span class="info-icon">&#63;</span>
                    <span class="info-pop-wrap" :class="activeInfo === 'format' ? 'info-pop-show' : ''">
                        <span class="info-pop-title">视频格式</span>
                        <span class="info-pop-text">网页端、桌面客户端推荐上传的格式为：mp4</span>
                        <span class="info-pop-text">暂不允许上传其他格式</span>
                    </span>
                </span>
                <span class="i-list i-1" @click="showUploadInfo('resolution')">
                    <span class="title">视频分辨率</span>
                    <span class="info-icon">&#63;</span>
                    <span class="info-pop-wrap" :class="activeInfo === 'resolution' ? 'info-pop-show' : ''">
                        <span class="info-pop-title">视频分辨率</span>
                        <span class="info-pop-text">推荐视频分辨率：1280*720 或者 1920*1080</span>
                        <span class="info-pop-text">网站不提供转码、压制服务</span>
                    </span>
                </span>
            </div>
            <div class="footer-item">
                <span>上传视频，即表示您已同意 <a target="_blank" class="agree-link">特丽特丽使用协议</a> 与 <a target="_blank" class="agree-link">特丽特丽社区公约</a> ，请勿上传色情，反动等违法视频。</span>
            </div>
        </div>
        <!-- 填写投稿信息界面 -->
        <div class="info-content" :style="selectedVideo ? '' : 'display: none;'">
            <div class="info-header">
                <span>发布视频</span>
                <div class="cancel-btn" @click="beforeCancel">取消发布</div>
            </div>
            <div class="info-body">
                <!-- 上传文件的进度显示 -->
                <div class="video-parts">
                    <!-- 上传总进度统计条 -->
                    <div class="upload-summary" v-if="videoParts.length > 0">
                        <div class="summary-left">
                            <i class="iconfont icon-shangchuanzhong"></i>
                            <span class="summary-stage">{{ uploadStage || '待上传' }}</span>
                            <span class="summary-percent">{{ summaryPercent }}%</span>
                        </div>
                        <div class="summary-right">
                            <span v-if="uploadSpeed > 0" class="summary-speed">{{ formatSpeed(uploadSpeed) }}/s</span>
                            <span v-if="uploadEta >= 0" class="summary-eta">剩余 {{ formatEta(uploadEta) }}</span>
                            <span class="summary-bytes">{{ formatBytes(uploadDoneBytes) }} / {{ formatBytes(uploadTotalBytes) }}</span>
                        </div>
                        <div class="summary-bar">
                            <div class="summary-bar-inner" :style="`width: ${summaryPercent}%;`"></div>
                        </div>
                    </div>
                    <div
                        v-for="(part, index) in videoParts"
                        :key="part.key"
                        class="video-part-row"
                        :class="part.isFailed ? 'part-failed' : ''"
                    >
                        <div class="video-icon-wrp"><i class="iconfont icon-video"></i></div>
                        <div class="file-detail">
                            <div class="part-title-row">
                                <span class="part-index">P{{ index + 1 }}</span>
                                <span class="part-name">{{ part.name }}</span>
                                <div class="part-title-actions">
                                    <span
                                        class="file-status-manage"
                                        :class="{ 'sort-disabled': index === 0 }"
                                        @click="movePart(index, -1)"
                                        title="上移"
                                    >
                                        <i class="iconfont icon-shangyishou"></i>
                                        <span>上移</span>
                                    </span>
                                    <span
                                        class="file-status-manage"
                                        :class="{ 'sort-disabled': index === videoParts.length - 1 }"
                                        @click="movePart(index, 1)"
                                        title="下移"
                                    >
                                        <i class="iconfont icon-xiayishou"></i>
                                        <span>下移</span>
                                    </span>
                                    <span
                                        class="file-status-manage"
                                        v-if="!part.isPause && part.progress < 100"
                                        @click="changeState(index)"
                                    >
                                        <i class="iconfont icon-zantingshangchuan"></i>
                                        <span>暂停</span>
                                    </span>
                                    <span
                                        class="file-status-manage"
                                        v-if="part.isPause"
                                        @click="changeState(index)"
                                    >
                                        <i class="iconfont icon-jixu"></i>
                                        <span>继续</span>
                                    </span>
                                    <span
                                        class="file-status-manage"
                                        v-if="part.progress >= 100 || part.isPause"
                                        @click="selectVideo(index)"
                                    >
                                        <i class="iconfont icon-genghuan"></i>
                                        <span>更换</span>
                                    </span>
                                    <span class="file-status-manage" @click="removePart(index)">
                                        <i class="iconfont icon-close"></i>
                                        <span>删除</span>
                                    </span>
                                </div>
                            </div>
                            <div class="file-status">
                                <div class="file-status-text" v-if="part.progress >= 100">
                                    <i class="iconfont icon-wancheng success"></i>
                                    <span class="success">分片上传完成</span>
                                </div>
                                <div class="file-status-text" v-else-if="part.isFailed">
                                    <i class="iconfont icon-shibai failed"></i>
                                    <span class="failed">分片上传失败</span>
                                </div>
                                <div class="file-status-text" v-else-if="!part.isPause">
                                    <i class="iconfont icon-shangchuanzhong"></i>
                                    <span style="margin-right: 8px;">上传中</span>
                                    <span>{{ part.progress }}%</span>
                                </div>
                                <div class="file-status-text" v-else>
                                    <i class="iconfont icon-zantingshangchuan"></i>
                                    <span style="margin-right: 8px;">暂停中</span>
                                    <span>{{ part.progress }}%</span>
                                </div>
                            </div>
                            <div class="file-progress">
                                <div class="file-progress-outer">
                                    <div
                                        class="file-progress-inner"
                                        :class="part.progress >= 100 ? 'success-bg' : part.isFailed ? 'failed-bg' : 'doing-bg'"
                                        :style="`width: ${part.progress}%;`"
                                    ></div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="add-part-row" v-if="selectablePartCount < 10">
                        <div class="add-part-btn" @click="addMoreVideos">
                            <i class="iconfont icon-jia"></i>
                            <span>添加分P（当前 {{ videoParts.length }}/10）</span>
                        </div>
                        <div class="add-part-tip">支持一次选择多个视频添加分P，投稿前可调整顺序</div>
                    </div>
                </div>
                <!-- 投稿信息 -->
                <div class="upload-form">
                    <div class="form-title">基本信息</div>
                    <!-- 封面 -->
                    <div class="form-item cover">
                        <div class="section-title">
                            <span class="section-title-deg">*</span>
                            <span class="section-title-main">封面</span>
                        </div>
                        <div class="cover-container">
                            <img :src="coverURL" v-if="coverURL">
                        </div>
                        <div class="cover-right">
                            <el-button class="change-cover-btn" @click="openDialog">更改封面</el-button>
                            <span class="cover-tips">系统默认截取视频第一秒作为封面</span>
                            <span class="cover-tips">*部分情况下您的封面将以4:3比例展示</span>
                        </div>                        
                    </div>
                    <!-- 标题 -->
                    <div class="form-item title">
                        <div class="section-title">
                            <span class="section-title-deg">*</span>
                            <span class="section-title-main">标题</span>
                        </div>
                        <el-input
                            v-model="form.title"
                            maxlength="80"
                            placeholder="起个吸引人的标题吧"
                            show-word-limit
                            type="text"
                        />
                    </div>
                    <!-- 类型 -->
                    <div class="form-item type">
                        <div class="section-title">
                            <span class="section-title-deg">*</span>
                            <span class="section-title-main">类型</span>
                        </div>
                        <div class="type-radio">
                            <div class="type-radio-item" @click="form.type = 1">
                                <div class="radio-box">
                                    <div class="radio-box-selected" :style="form.type == 1 ? '' : 'display: none;'"></div>
                                </div>
                                <span class="radio-name" :class="form.type == 1 ? 'radio-selected' : ''">自制</span>
                            </div>
                            <div class="type-radio-item" @click="form.type = 2; form.auth = 0;">
                                <div class="radio-box">
                                    <div class="radio-box-selected" :style="form.type == 2 ? '' : 'display: none;'"></div>
                                </div>
                                <span class="radio-name" :class="form.type == 2 ? 'radio-selected' : ''">转载</span>
                            </div>
                        </div>
                        <div class="auth" v-if="form.type == 1" @click="changeAuth">
                            <span class="auth-box" :class="form.auth == 1 ? 'auth-selected' : ''">
                                <i class="iconfont icon-gou"></i>
                            </span>
                            <span class="auth-main">未经作者授权 禁止转载</span>
                        </div>
                    </div>
                    <!-- 分区 -->
                    <div class="form-item category">
                        <div class="section-title">
                            <span class="section-title-deg">*</span>
                            <span class="section-title-main">分区</span>
                        </div>
                        <CategorySelect v-model:category="form.category"></CategorySelect>
                    </div>
                    <!-- 标签 -->
                    <div class="form-item tag">
                        <div class="section-title">
                            <span class="section-title-deg">*</span>
                            <span class="section-title-main">标签</span>
                        </div>
                        <!-- 用v-if重新渲染不留缓存 -->
                        <TagInput v-if="this.selectedVideo" style="width: 80%;" v-model:sendTag="sendTag" @updateTags="val => {form.tags = val;}"></TagInput>
                    </div>
                    <!-- 推荐标签 -->
                    <div class="tag-wrp" v-if="rcmTags.length != 0">
                        <p class="tag-label">推荐标签:</p>
                        <div class="tag-list">
                            <div
                                class="hot-tag-container"
                                :class="form.tags.findIndex(curr=>curr==item) != -1 ? 'isTagSelected' : ''"
                                @click="sendTag = item;"
                                v-for="(item, index) in rcmTags"
                                :key="index"
                            >
                                <div class="hot-tag-item">
                                    <span>{{ item }}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                    <!-- 简介 -->
                    <div class="form-item descr" style="align-items: flex-start;">
                        <div class="section-title" style="padding-top: 35px;">
                            <span class="section-title-deg"></span>
                            <span class="section-title-main">简介</span>
                        </div>
                        <el-input
                            class="desc-text"
                            type="textarea"
                            maxlength="2000"
                            show-word-limit
                            v-model="form.descr"
                            placeholder="填写更全面的相关信息，让更多的人能找到你的视频吧:)"
                        ></el-input>
                    </div>
                    <div class="form-item submit">
                        <div class="submit-container">
                            <span class="submit-draft" @click="draft">存草稿</span>
                            <span class="submit-drafts-box" @click="goDraftsBox">{{ draftsBoxLabel }}</span>
                            <span class="submit-add" @click="submit">立即投稿</span>
                        </div>
                    </div>
                </div>
            </div>
            <!-- 用于裁剪视频封面的元素 -->
            <video :src="videoURL" ref="uploadVideoElement" style="display: none"></video>
            <!-- 用于生成原画的画布 -->
            <canvas ref="cutCanvas" style="display: none"></canvas>
        </div>
    </div>
    <!-- 封面裁剪框 -->
    <div class="cover-dialog">
    <el-dialog v-model="dialogVisible" :close-on-click-modal="false" destroy-on-close align-center>
        <div class="cover-dialog-header">
            <div class="header-tab">
                <div class="header-tab-item" :class="coverType == 1 ? 'currentTab' : ''" @click="coverType = 1">
                    <div class="tab-text">截取封面</div>
                    <div class="tab-line"></div>
                </div>
                <div class="header-tab-item" :class="coverType == 2 ? 'currentTab' : ''" @click="coverType = 2">
                    <div class="tab-text">上传封面</div>
                    <div class="tab-line"></div>
                </div>
            </div>
        </div>
        <div class="cover-dialog-body">
            <div class="cover-cut-content" :class="coverType == 1 ? 'pick-version' : 'upload-version'">
                <div class="cover-cut-content-pick" :style="coverType == 1 ? '' : 'display: none;'">
                    <!-- 裁剪组件 -->
                    <CropperComp :imgURL="coverImageURL" v-model:cutOrder="cutOrder1" :title="form.title" @cut="changeCover"></CropperComp>
                    <!-- 进度条组件 -->
                    <div class="cover-cut-content-pick-bar">
                        <div class="cover-slider" ref="coverSlider">
                            <PlayerProgress style="top: 20px; visibility: hidden;" :currentPer="currentPer" @changeCurrent="changeCurrentPer"></PlayerProgress>
                            <div class="slider-handle" ref="sliderHandle" :style="`left: ${currentPer * 100}%;`"></div>
                            <div class="cover-slider-image-list" v-if="sliderImages.length == 7">
                                <div
                                    class="cover-slider-image"
                                    v-for="index in 7"
                                    :key="index"
                                    :style="`background: url(${sliderImages[index-1]}) center center / cover no-repeat;`"
                                ></div>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="cover-cut-content-upload" :style="coverType == 2 ? '' : 'display: none;'">
                    <!-- 选择文件界面 -->
                    <div class="upload-body-content" style="top: 73px; width: 600px;" :style="picURL ? 'display: none;' : ''">
                        <div class="upload-wrp" style="margin-top: 0;">
                            <div class="pic-input">
                                <div
                                    class="pic-input-wrapper"
                                    @dragover.prevent
                                    @dragenter.prevent
                                    @dragleave.prevent
                                    @drop.prevent
                                    @drop="handlePicDrop"
                                    @click="selectPic"
                                >
                                    <!-- <i class="iconfont icon-shangchuan"></i> -->
                                    <el-icon size="40"><UploadFilled /></el-icon>
                                    <div class="upload-tips-text">拖拽图片到此处或者点击上传</div>
                                    <div class="upload-btn">选择图片</div>
                                    <input
                                        type="file"
                                        accept="image/png, image/jpg, image/jpeg"
                                        ref="picInput"
                                        @change="handlePicChange"
                                        style="display: none;"
                                    >
                                </div>
                            </div>
                        </div>
                    </div>
                    <!-- 裁剪组件 -->
                    <CropperComp v-if="picURL" :imgURL="picURL" v-model:cutOrder="cutOrder2" :title="form.title" @cut="changeCover"></CropperComp>
                    <!-- 更换图片按钮 -->
                    <div class="reselect">
                        <div class="file-status-manage refresh" v-if="picURL" @click="selectPic">
                            <i class="iconfont icon-genghuan"></i>
                            <span>重新选择</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="cover-dialog-footer">
            <div class="cover-dialog-submit" @click="sendOrder" v-if="coverType == 1 || picURL"><span>完成</span></div>
        </div>
    </el-dialog>
    </div>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import SparkMD5 from 'spark-md5';
import CropperComp from '@/components/teriteri/cropper/CropperComp.vue';
import CategorySelect from '@/components/teriteri/categorySelect/CategorySelect.vue';
import TagInput from '@/components/teriteri/tagInput/TagInput.vue';
import PlayerProgress from '@/components/teriteri/player/PlayerProgress.vue';
import UploadTypeTabs from '@/components/teriteri/uploadTypeTabs/UploadTypeTabs.vue';
import { saveDraft, getDraft, listDrafts } from '@/utils/drafts';

export default {
    name: "VideoUpload",
    components: {
        CropperComp,
        CategorySelect,
        TagInput,
        PlayerProgress,
        UploadTypeTabs,
    },
    // 提前声明要使用的自定义事件，防止提示一堆警告
    emits: ['changeNavBarShow'],
    data() {
        return {
            dialogVisible: false,   // 封面裁剪框的显隐
            coverType: 1,   // 当前显示封面裁剪类型，1 视频帧截取 2 上传裁剪
            activeInfo: "", // 当前展开的信息说明（size/format/resolution）
            selectedVideo: null,    // 选择视频文件
            videoURL: null, // 上传的视频的内存地址
            videoName: "",  // 视频原文件名
            chunkSize: 10*1024*1024,  // 分片大小小于等于后端，分片越小越多断点续传效果越好，但上传速度相对也会慢
            // 多 P 上传：每个分P 独立的状态
            videoParts: [], // 分P列表 [{file,name,key,partIndex,hash,current,totalChunks,progress,isPause,isFailed,isCancel}]
            sessionHash: null,  // 整个投稿会话的稳定 id（用第一个分P 文件的 hash）
            uploadId: null, // 后端 upload-session id（用 sessionHash 作稳定 id）
            replacePartIndex: null, // selectVideo 时带的替换目标分P下标，null 表示新增
            uploadingIndex: -1, // 当前正在上传的分P下标，-1 表示无
            isCancel: false,    // 是否取消上传
            uploadTotalBytes: 0,   // 全部待上传字节
            uploadDoneBytes: 0,    // 已上传字节
            uploadSpeed: 0,        // 上传速度 B/s
            uploadEta: -1,         // 剩余秒数，-1 不可估算
            uploadStage: "",       // 当前上传阶段文案
            uploadSpeedTimer: null,
            uploadSpeedSnapshot: 0,    // 速度计算上一采样
            uploadSpeedSnapshotTime: 0,
            duration: 0,    // 视频总时长
            coverImageURL: null,   // 封面原图url
            currentPer: 0.00001,   // 当前封面的视频时间占总时长的百分比 进度条百分比 小数0~1
            sliderImages: [],   // 滑块预览的7张图
            picURL: null,   // 上传的原图url
            coverURL: null, // 裁剪后的封面url
            cutOrder1: false,   // 用于传递裁剪信息的参数，true就触发裁剪 v-model双向绑定，裁剪完自动变回false;
            cutOrder2: false,
            sendTag: "",    // 点击推荐标签时发送的想要添加的标签
            draftCount: 0,   // 本地草稿数量
            form: {
                title: "",  // 投稿标题
                type: 1,    // 投稿类型 1自制 2转载
                auth: 0,    // 0不设置权限 1未经作者授权禁止转载
                category: [{id: "anime", name: "番剧"}, {id: "finish", name: "完结动画"}],  // 投稿分区
                tags: [],   // 投稿标签
                descr: "",  // 投稿简介
            },
        }
    },
    watch: {
        // 频道加载完成后，若当前仍是旧版占位分区，则换成本项目真实分区（数字 id）
        categoryList() {
            this.ensureRealCategory();
        },
    },
    created() {
        // 本地草稿数量
        this.draftCount = listDrafts().length;
        this.ensureRealCategory();
        // 从草稿箱跳转：?draftId= 恢复已填写的表单信息（视频文件需重新选择）
        const draftId = this.$route.query.draftId;
        if (draftId) {
            const draft = getDraft(draftId);
            if (draft) {
                const channels = this.$store.state.channels || [];
                const cat = { id: null, name: "" };
                if (draft.categoryId) {
                    for (let i = 0; i < channels.length; i++) {
                        for (let j = 0; j < (channels[i].scList || []).length; j++) {
                            if (String(channels[i].scList[j].scId) === String(draft.categoryId)) {
                                cat.id = channels[i].scList[j].scId;
                                cat.name = channels[i].scList[j].scName;
                                break;
                            }
                        }
                    }
                }
                this.form.title = draft.title || "";
                this.form.type = draft.type === 'repost' ? 2 : 1;
                this.form.tags = Array.isArray(draft.tags) ? draft.tags.slice() : [];
                this.form.descr = draft.description || "";
                if (typeof this.form.category[1] === 'object') {
                    this.form.category[1] = { id: cat.id || "anime", name: cat.name || "番剧" };
                }
                this.$nextTick(() => {
                    ElMessage.success("草稿已载入，请重新选择视频文件后继续编辑");
                });
            }
        }
    },
    computed: {
        // 分区列表
        categoryList() {
            return this.$store.state.channels;
        },
        // 还有多少分P可添加（上限10）
        selectablePartCount() {
            return 10 - this.videoParts.length;
        },
        // 上传总进度百分比
        summaryPercent() {
            if (this.uploadTotalBytes <= 0) return 0;
            return Math.min(100, Math.round((this.uploadDoneBytes / this.uploadTotalBytes) * 100));
        },
        // 草稿箱按钮文案
        draftsBoxLabel() {
            const count = this.draftCount || 0;
            return count > 0 ? `草稿箱 (${count})` : '草稿箱';
        },
        // 当前选择分区的推荐可选标签
        rcmTags() {
            for (let i = 0; i < this.categoryList.length; i ++) {
                if (this.categoryList[i].mcId == this.form.category[0].id) {
                    for (let j = 0; j < this.categoryList[i].scList.length; j ++) {
                        if (this.categoryList[i].scList[j].scId == this.form.category[1].id) {
                            return this.categoryList[i].scList[j].rcmTag;
                        }
                    }
                }
            }
            return [];
        },
    },
    methods: {
        // 初始化内容
        init() {
            this.selectedVideo = null;
            this.sessionHash = null;
            this.videoURL = null;
            this.videoName = "";
            this.videoParts = [];
            this.uploadId = null;
            this.isCancel = false;
            this.activeInfo = "";
            this.uploadTotalBytes = 0;
            this.uploadDoneBytes = 0;
            this.uploadSpeed = 0;
            this.uploadEta = -1;
            this.uploadStage = "";
            this.uploadSpeedSnapshot = 0;
            this.uploadSpeedSnapshotTime = 0;
            this.stopUploadSpeedTimer();
            this.coverImageURL = null;
            this.coverURL = null;
            this.duration = 0;
            this.currentPer = 0.00001;
            this.sliderImages = [],
            this.picURL = null,
            this.form = {
                title: "",
                type: 1,
                auth: 0,
                category: [{id: "anime", name: "番剧"}, {id: "finish", name: "完结动画"}],
                tags: [],
                descr: "",
            }
        },

        // 用项目真实分区（数字 id）替换 teriteri 遗留的字符串占位分区，并把用户已选项
        // 按数字 id 归一化，保证提交时 category_id 一定是有效数字。
        ensureRealCategory() {
            const channels = this.categoryList || [];
            if (!channels.length) return;
            // 已选的真实数字分区（>0）视为合法，不覆盖
            const cur = this.form.category && this.form.category[1];
            if (cur && typeof cur.id === 'number' && cur.id > 0) return;
            // 取第一个主分区下的第一个子分区作为默认
            const main = channels.find(c => (c.scList || []).length > 0) || channels[0];
            if (!main || !main.mcId) return;
            const sub = (main.scList || [])[0];
            if (!sub || !sub.scId) return;
            this.form.category = [
                { id: main.mcId, name: main.mcName },
                { id: sub.scId, name: sub.scName },
            ];
        },

        // 分区是否已合法选择（数字 id 且存在）
        categoryValid() {
            const cur = this.form.category && this.form.category[1];
            if (!cur || typeof cur.id !== 'number' || cur.id <= 0) return false;
            for (let i = 0; i < this.categoryList.length; i ++) {
                for (let j = 0; j < (this.categoryList[i].scList || []).length; j ++) {
                    if (this.categoryList[i].scList[j].scId === cur.id) return true;
                }
            }
            return false;
        },

        // 切换投稿类型
        switchUploadType(type) {
            if (this.videoParts.length > 0) {
                ElMessage.warning('请先取消当前投稿再切换类型');
                return;
            }
            if (type === 'text') {
                this.$router.push('/platform/upload/text');
            }
        },

        // 打开封面裁剪对话框
        openDialog() {
            this.dialogVisible = true;
            setTimeout(() => {
                this.initDrag();
            }, 100);
        },

        // 触发文件选择对话框（index 传则更换指定分P，不传则添加分P/首次选择）
        selectVideo(index = null) {
            this.videoURL = null; // 清除先前的视频URL，防止新视频还没加载出来就画黑屏
            this.replacePartIndex = index;
            this.$refs.videoInput.click();
        },

        // 选择界面上"添加分P"按钮
        addMoreVideos() {
            this.videoURL = null;
            this.replacePartIndex = null;
            this.$refs.videoInput.click();
        },

        // 切换信息说明弹层
        showUploadInfo(key) {
            this.activeInfo = this.activeInfo === key ? "" : key;
        },

        // 触发图片选择对话框
        selectPic() {
            this.$refs.picInput.click();
        },
        
        // 处理文件选择事件
        async handleVideoChange(event) {
            if (!this.$store.state.isLogin) {
                ElMessage.error('请登录后查看');
                this.$store.dispatch("logout");
                this.$router.push("/");
                return;
            }
            const files = Array.from(event.target.files || []);
            event.target.value = "";
            this.addParts(files);
        },

        // 添加或替换分P（files 为选中的一批文件）
        async addParts(files) {
            if (files.length === 0) return;
            if (!this.$store.state.isLogin) return;

            // 替换模式：replacePartIndex 非 null，仅取第一个文件替换指定分P
            if (this.replacePartIndex !== null) {
                const idx = this.replacePartIndex;
                const part = this.videoParts[idx];
                this.replacePartIndex = null;
                if (!part) return;
                const file = files[0];
                const check = this.checkVideoFile(file);
                if (!check.ok) return;
                const replacement = {
                    file: file,
                    name: file.name,
                    partIndex: idx,
                    hash: null,
                    current: 0,
                    totalChunks: 0,
                    progress: 0,
                    isPause: false,
                    isFailed: false,
                    isCancel: false,
                };
                const uploadedBefore = part.progress >= 100;
                this.videoParts.splice(idx, 1, replacement);
                // 替换第一个分P后重生成封面（B站：封面默认取第一个分P）
                if (idx === 0) {
                    this.initCover(replacement.file);
                }
                // 原分P若已传完，替换后需要重新上传
                if (uploadedBefore) {
                    this.uploadPart(replacement);
                }
                return;
            }

            // 新增模式：逐个文件加入分P列表
            for (let i = 0; i < files.length; i++) {
                const file = files[i];
                const check = this.checkVideoFile(file);
                if (!check.ok) {
                    continue;
                }
                const isFirst = this.videoParts.length === 0;
                const part = {
                    file: file,
                    name: file.name,
                    partIndex: 0,
                    hash: null,
                    current: 0,
                    totalChunks: 0,
                    progress: 0,
                    isPause: false,
                    isFailed: false,
                    isCancel: false,
                };
                this.videoParts.push(part);
                if (isFirst && !this.sessionHash) {
                    this.sessionHash = await this.fhash(file);
                }
            }
            // 重排 partIndex
            this.videoParts.forEach((p, i) => { p.partIndex = i; });

            if (this.videoParts.length > 0) {
                this.selectedVideo = true;
                // 第一个分P初始化封面并设置默认标题
                if (!this.videoName && !this.form.title) {
                    this.videoName = this.videoParts[0].name.split(".mp4")[0];
                    this.form.title = this.videoName.slice();
                    this.initCover(this.videoParts[0].file);
                }
                this.$emit("changeNavBarShow", false);
                // 新加入的分P逐个上传
                for (let i = 0; i < this.videoParts.length; i++) {
                    const p = this.videoParts[i];
                    if (p.progress < 100 && !p.isPause && !p.isFailed) {
                        await this.uploadPart(p);
                    }
                }
            }
        },

        // 校验视频文件（大小/格式）
        checkVideoFile(file) {
            const maxSizeInBytes = 300 * 1024 * 1024; // 300MB
            if (!file) {
                return { ok: false };
            }
            if (file.type && file.type != "video/mp4" && file.name && !file.name.toLowerCase().endsWith(".mp4")) {
                ElMessage.error("视频只接收mp4格式哦");
                return { ok: false };
            }
            if (file.size > maxSizeInBytes) {
                ElMessage.error("视频文件过大，请压缩后重试");
                return { ok: false };
            }
            return { ok: true };
        },

        // 处理图片选择事件
        handlePicChange(event) {
            const file = event.target.files[0];
            const maxSizeInBytes = 5 * 1024 * 1024; // 5MB
            if (!file) {
                return;
            }
            if (file.size <= maxSizeInBytes) {
                // 文件大小符合要求，可以继续处理上传逻辑
                const reader = new FileReader();
                reader.onload = (e) => {
                    this.picURL = e.target.result;
                }
                reader.readAsDataURL(file);
            } else {
                // 文件大小超出限制
                ElMessage.error("只能上传5M的图片哦")
            }
        },

        // 处理拖拽事件
        async handleDrop(event) {
            event.preventDefault();
            if (!this.$store.state.isLogin) {
                ElMessage.error('请登录后查看');
                this.$store.dispatch("logout");
                this.$router.push("/");
                return;
            }
            const files = Array.from(event.dataTransfer.files || []);
            this.addParts(files);
        },

        handlePicDrop(event) {
            event.preventDefault();
            const file = event.dataTransfer.files[0];
            const maxSizeInBytes = 5 * 1024 * 1024; // 5MB
            if (!file) {
                return;
            }
            if (file.type != "image/png" && file.type !="image/jpeg") {
                ElMessage.error("图片只接收jpg或png格式哦");
                return;
            }
            if (file.size <= maxSizeInBytes) {
                // 文件大小符合要求，可以继续处理上传逻辑
                const reader = new FileReader();
                reader.onload = (e) => {
                    this.picURL = e.target.result;
                }
                reader.readAsDataURL(file);
            } else {
                // 文件大小超出限制
                ElMessage.error("只能上传5M的图片哦")
            }
        },

        // 触发暂停/继续事件
        changeState(index) {
            const part = this.videoParts[index];
            if (!part) return;
            part.isPause = !part.isPause;
            if (!part.isPause && part.progress < 100) {
                this.uploadPart(part);
            }
        },

        // 格式化字节数
        formatBytes(bytes) {
            if (!bytes || bytes <= 0) return "0 B";
            const units = ["B", "KB", "MB", "GB"];
            let idx = 0;
            let v = bytes;
            while (v >= 1024 && idx < units.length - 1) {
                v /= 1024;
                idx++;
            }
            return `${v >= 10 || idx === 0 ? Math.round(v) : v.toFixed(1)} ${units[idx]}`;
        },

        // 格式化速度
        formatSpeed(speed) {
            return this.formatBytes(speed);
        },

        // 格式化剩余时间
        formatEta(eta) {
            if (eta <= 0) return "0秒";
            if (eta < 60) return `${eta}秒`;
            const m = Math.floor(eta / 60);
            const s = eta % 60;
            return `${m}分${s}秒`;
        },

        // 删除指定分P
        removePart(index) {
            const part = this.videoParts[index];
            if (!part) return;
            if (this.uploadingIndex === index) {
                ElMessage.error('当前分P正在上传，请先暂停再删除');
                return;
            }
            const videoURL = this.videoURL;
            this.videoParts.splice(index, 1);
            // 重新编号 partIndex
            this.videoParts.forEach((p, i) => { p.partIndex = i; });
            if (this.videoParts.length === 0) {
                this.init();
                this.$emit("changeNavBarShow", true);
            } else {
                // 删除第一个分P后重新生成封面
                if (index === 0 && this.videoParts.length > 0) {
                    this.initCover(this.videoParts[0].file);
                }
            }
        },

        // 上下移动分P（dir: -1 上移 / 1 下移），移动后重编 partIndex
        movePart(index, dir) {
            const target = index + dir;
            if (target < 0 || target >= this.videoParts.length) return;
            const current = this.videoParts[index];
            if (this.uploadingIndex === index || this.uploadingIndex === target) {
                ElMessage.error('正在上传的分P暂不能排序，请先暂停');
                return;
            }
            const uploadedAny = current.progress >= 100 || this.videoParts[target].progress >= 100;
            if (uploadedAny) {
                ElMessage.error('已上传完成的分P暂不能排序（需重新上传该分P）');
                return;
            }
            this.videoParts.splice(index, 1, this.videoParts.splice(target, 1, current)[0]);
            this.videoParts.forEach((p, i) => { p.partIndex = i; });
        },

        // 根据整个文件的文件名和大小组合的字符串生成hash值，大概率确定文件的唯一性
        fhash(file) {
            // console.log("哈希字段: ", file.name+file.size.toString());
            return new Promise(resolve => {
                const spark = new SparkMD5();
                spark.append(file.name+file.size.toString());
                resolve(spark.end());
            })
        },

        // 生成切片
        createChunks(file) {
            const result = [];
            for (let i = 0; i < file.size; i += this.chunkSize) {
                result.push(file.slice(i, i + this.chunkSize));
            }
            return result;
        },

        // 触发上传事件（单个分P上传，partIndex 隔离目录支持多P）
        async uploadPart(part) {
            if (!part) {
                ElMessage.error('未选择文件');
                return;
            }
            const chunks = this.createChunks(part.file);
            part.totalChunks = chunks.length;
            part.isPause = false;
            part.isFailed = false;
            this.uploadingIndex = part.partIndex;
            // 上传统计初始化
            if (this.uploadTotalBytes === 0) {
                this.uploadTotalBytes = this.videoParts.reduce((s, p) => s + p.file.size, 0);
                this.uploadDoneBytes = this.videoParts
                    .filter((p) => p.progress >= 100)
                    .reduce((s, p) => s + p.file.size, 0);
            }
            // 创建/复用上传会话：id 用 sessionHash（第一个分P的 hash），多P时 videos 数组汇总各P分片数
            const partsMeta = this.videoParts.map((p) => ({
                title: p.name,
                total_chunks: this.createChunks(p.file).length,
            }));
            const totalChunks = partsMeta.reduce((s, v) => s + (v.total_chunks || 0), 0);
            const initRes = await this.createSession({
                client_id: this.sessionHash,
                title: this.form.title || this.sessionHash,
                category_id: 0,
                total_chunks: totalChunks,
                videos: partsMeta,
            });
            const initBody = initRes && (initRes.data || initRes);
            if (!initBody || !initBody.upload_id) {
                part.isFailed = true;
                part.isPause = true;
                ElMessage.error('创建上传会话失败');
                this.uploadingIndex = -1;
                return;
            }
            this.uploadId = initBody.upload_id;
            const targetIndex = part.partIndex;
            this.startUploadSpeedTimer();
            // 逐个上传分片（失败自动重试3次）
            let uploadError = false;
            for (let i = 0; i < chunks.length; i++) {
                const chunk = chunks[i];
                let chunkOk = false;
                for (let attempt = 0; attempt <= 3; attempt++) {
                    if (part.isPause) break;
                    const formData = new FormData();
                    formData.append('file', chunk); // 后端 multipart 字段名
                    formData.append('uploadId', this.uploadId);
                    formData.append('partIndex', targetIndex); // 传递分P序号（后端按 partIndex 分目录）
                    formData.append('chunkIndex', i); // 传递分片索引

                    try {
                        const res = await this.uploadChunk(formData);
                        const resBody = res && (res.data || res);
                        if (resBody && resBody.code === 200) {
                            chunkOk = true;
                            break;
                        }
                    } catch {
                        // 重试
                    }
                    if (attempt < 3) {
                        await new Promise(r => setTimeout(r, 500 * Math.pow(2, attempt)));
                    }
                }
                if (!chunkOk) {
                    uploadError = true;
                    break;
                }
                // 字节与进度统计
                this.uploadDoneBytes += chunk.size;
                part.progress = Math.round(((i + 1) / chunks.length) * 100); // 实时改进度条
                // 暂停上传
                if (part.isPause) {
                    break;
                }
            }
            part.progress = 100;    // 上传完成再次确认为100%
            if (uploadError) {
                part.isFailed = true;
                part.isPause = true;
            }
            this.uploadingIndex = -1;
            // 全部传完时停止统计
            if (this.videoParts.every((p) => p.progress >= 100)) {
                this.uploadStage = "分片上传完成";
                this.uploadSpeed = 0;
                this.stopUploadSpeedTimer();
            }
        },

        // 启动速度统计定时器
        startUploadSpeedTimer() {
            if (this.uploadSpeedTimer) return;
            this.uploadSpeedTimer = setInterval(() => {
                this.computeUploadSpeed();
            }, 1000);
        },

        // 停止速度统计定时器
        stopUploadSpeedTimer() {
            if (this.uploadSpeedTimer) {
                clearInterval(this.uploadSpeedTimer);
                this.uploadSpeedTimer = null;
            }
            this.uploadSpeed = 0;
        },

        // 每秒计算速度与剩余时间
        computeUploadSpeed() {
            if (!this.uploadSpeedSnapshot || Date.now() - this.uploadSpeedSnapshotTime < 500) {
                this.uploadSpeedSnapshot = this.uploadDoneBytes;
                this.uploadSpeedSnapshotTime = Date.now();
                return;
            }
            const bytesDelta = this.uploadDoneBytes - (this.uploadSpeedSnapshot || 0);
            const timeDelta = (Date.now() - this.uploadSpeedSnapshotTime) / 1000;
            this.uploadSpeedSnapshot = this.uploadDoneBytes;
            this.uploadSpeedSnapshotTime = Date.now();
            this.uploadSpeed = Math.max(0, Math.round(bytesDelta / Math.max(timeDelta, 0.1)));
            const remaining = this.uploadTotalBytes - this.uploadDoneBytes;
            this.uploadEta = this.uploadSpeed > 0 ? Math.round(remaining / this.uploadSpeed) : -1;
            this.uploadStage = this.uploadDoneBytes >= this.uploadTotalBytes ? "分片上传完成" : "正在上传视频分片";
        },

        // 取消上传前的最后通牒
        beforeCancel() {
            ElMessageBox.confirm(
                '取消发布将不保存你当前填写好的信息哦，确定不发布了吗？',
                {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning',
                }
            )
            .then(() => {
                this.cancel();
            })
            .catch(() => {})
        },

        // 确认取消上传
        async cancel() {
            this.videoParts.forEach((part) => {
                part.isPause = true;
            });
            this.isCancel = true;
            await this.cancelUpload(this.uploadId);
            this.init();
            setTimeout(() => {
                this.isCancel = false;
            }, 3000);   // 延迟恢复取消状态
            this.$emit("changeNavBarShow", true);
        },

        // 初始化封面
        initCover(file) {
            // 获取视频元素
            const video = this.$refs.uploadVideoElement;
            // 为视频元素设置选中的视频文件
            this.videoURL = URL.createObjectURL(file);
            video.crossOrigin = 'anonymous' // 解决跨域问题，也就是提示污染资源无法转换视频
            video.currentTime = 1 // 第一秒
            // 在视频可以播放时获取这一秒图像
            video.oncanplay = async () => {
                // 获取视频总时长
                this.duration = video.duration;
                // 获取画布元素
                const canvas = this.$refs.cutCanvas;
                canvas.width = video.videoWidth;
                canvas.height = video.videoHeight;
                // 将视频的这一帧绘制到画布上
                const context = canvas.getContext("2d");
                context.drawImage(video, 0, 0, canvas.width, canvas.height);
                // 将画布内容转换为图像数据URL
                this.coverImageURL = canvas.toDataURL("image/jpeg");
                this.cutCover(this.coverImageURL);  // 从原图裁剪中间部分作为封面展示
                await this.initSliderImages();
            };
        },

        // 截取滑块预览的7张图
        async initSliderImages() {
            // 获取视频元素
            const video = this.$refs.uploadVideoElement;
            video.crossOrigin = 'anonymous' // 解决跨域问题，也就是提示污染资源无法转换视频
            for (let i = 1; i < 14; i += 2) {   // 截中间的7张
                video.currentTime = this.duration * (i * 0.07143);
                await new Promise((resolve) => {
                    video.oncanplay = () => {
                        const canvas = this.$refs.cutCanvas;
                        canvas.width = video.videoWidth;
                        canvas.height = video.videoHeight;
                        
                        const context = canvas.getContext("2d");
                        context.drawImage(video, 0, 0, canvas.width, canvas.height);
                        this.sliderImages.push(canvas.toDataURL("image/jpeg"));
                        
                        resolve();
                    };
                });
            }
            // console.log(this.sliderImages);
        },

        // 封面截取， imgURL 原图的url
        cutCover(imgURL) {
            // 创建一个 Image 元素来加载原始图像
            const img = new Image();
            img.src = imgURL;
            // 获取用于裁剪的 canvas 元素
            const canvas = this.$refs.cutCanvas;
            const ctx = canvas.getContext('2d');
            // 定义所需的宽高比例
            const targetAspectRatio = 16 / 9;
            // 当图像加载完成后执行裁剪操作
            img.onload = () => {
                // 计算原始图像和目标图像的宽高
                const srcWidth = img.width;
                const srcHeight = img.height;
                const srcAspectRatio = srcWidth / srcHeight;
                let targetWidth = srcWidth;
                let targetHeight = srcHeight;
                // 如果原始图像的宽高比例较大（宽长高短），根据目标宽高比例来计算宽度
                if (srcAspectRatio > targetAspectRatio) {
                    targetWidth = srcHeight * targetAspectRatio;
                }
                // 否则（宽短高长），根据目标宽高比例来计算高度
                else {
                    targetHeight = srcWidth / targetAspectRatio;
                }
                // 计算裁剪后的图像在画布上的位置
                const offsetX = (srcWidth - targetWidth) / 2;
                const offsetY = (srcHeight - targetHeight) / 2;
                // 设置 Canvas 的宽高为目标宽高
                canvas.width = targetWidth;
                canvas.height = targetHeight;
                // 在 Canvas 上绘制裁剪后的图像
                ctx.drawImage(img, offsetX, offsetY, targetWidth, targetHeight, 0, 0, targetWidth, targetHeight);
                // 将 Canvas 转换为DataURL并返回
                this.coverURL = canvas.toDataURL('image/jpeg');
            };
        },

        // 初始化拖动
        initDrag() {
            const coverSlider = this.$refs.coverSlider;
            const sliderHandle = this.$refs.sliderHandle;

            let isDragging = false; // 是否正在拖动
            let offsetX, currPer;

            // 鼠标按下事件处理程序
            sliderHandle.addEventListener("mousedown", () => {
                isDragging = true;
            });

            // 移动端 按下
            sliderHandle.addEventListener("touchstart", () => {
                isDragging = true;
            });

            // 鼠标移动事件处理程序
            document.addEventListener("mousemove", (e) => {
                if (!isDragging) return;
                offsetX = e.clientX - coverSlider.getBoundingClientRect().left; // 鼠标相对进度条左侧的X偏移
                currPer = offsetX / coverSlider.getBoundingClientRect().width; // 计算进度比例
                // 边界值判定
                currPer = Math.max(0.00001, currPer);
                currPer = Math.min(0.99999, currPer);
                this.currentPer = currPer;    // 更新
                const currentTime = currPer * this.duration;
                this.selectCoverSlice(currentTime);
            });

            // 移动端 移动
            document.addEventListener("touchmove", (e) => {
                if (!isDragging) return;
                e.preventDefault();
                offsetX = e.touches[0].clientX - coverSlider.getBoundingClientRect().left;
                currPer = offsetX / coverSlider.getBoundingClientRect().width; // 计算进度比例
                // 边界值判定
                currPer = Math.max(0.00001, currPer);
                currPer = Math.min(0.99999, currPer);
                this.currentPer = currPer;    // 更新
                const currentTime = currPer * this.duration;
                this.selectCoverSlice(currentTime);
            }, { passive: false });

            // 鼠标释放事件处理程序
            document.addEventListener("mouseup", () => {
                isDragging = false;
            });

            // 移动端 松开
            document.addEventListener("touchend", () => {
                isDragging = false;
            });
        },

        changeCurrentPer(curr) {
            this.currentPer = curr;
            const currentTime = curr * this.duration;
            this.selectCoverSlice(currentTime);
        },

        // 选择视频帧
        selectCoverSlice(time) {
            // 获取视频元素
            const video = this.$refs.uploadVideoElement;
            video.crossOrigin = 'anonymous' // 解决跨域问题，也就是提示污染资源无法转换视频
            video.currentTime = time;
            // 在视频可以播放时获取这一秒图像
            video.oncanplay = () => {
                // 获取画布元素
                const canvas = this.$refs.cutCanvas;
                canvas.width = video.videoWidth;
                canvas.height = video.videoHeight;
                // 将视频的这一帧绘制到画布上
                const context = canvas.getContext("2d");
                context.drawImage(video, 0, 0, canvas.width, canvas.height);
                // 将画布内容转换为图像数据URL
                this.coverImageURL = canvas.toDataURL("image/jpeg");
            };
        },

        // 更改作者声明
        changeAuth() {
            if (this.form.auth === 1) {
                this.form.auth = 0;
            } else {
                this.form.auth = 1;
            }
        },

        changeCover(url) {
            this.coverURL = url;
            // 换完源后关闭对话框
            this.dialogVisible = false;
            // 初始化上传的图片
            this.picURL = null;
        },

        // 发送裁剪指令
        sendOrder() {
            if (this.coverType === 1) {
                this.cutOrder1 = true;
            } else {
                this.cutOrder2 = true;
            }
        },

        // 存草稿（保存到本地 localStorage，可稍后继续编辑）
        draft() {
            if (this.videoParts.length === 0) {
                ElMessage.error('先选择视频再存草稿吧');
                return;
            }
            const draft = saveDraft({
                title: this.form.title || (this.videoParts[0] ? this.videoParts[0].name.split(".mp4")[0] : ""),
                categoryId: this.form.category && this.form.category[1] ? this.form.category[1].id : null,
                tags: this.form.tags.slice(),
                description: this.form.descr,
                type: this.form.type === 1 ? 'original' : 'repost',
                coverPreview: this.coverURL || "",
                videoParts: this.videoParts.map((part) => ({
                    title: part.name.split(".mp4")[0],
                    sortOrder: part.partIndex,
                })),
                hasLocalVideoFiles: this.videoParts.some((p) => p.progress >= 100),
            });
            if (draft) {
                this.draftCount = listDrafts().length;
                ElMessage.success('草稿已保存到本地，可稍后在"草稿箱"中继续编辑');
            } else {
                ElMessage.error('保存草稿失败');
            }
        },

        // 打开草稿箱
        goDraftsBox() {
            this.$router.push('/platform/drafts');
        },

        async submit() {
            // 判断分片上传情况、表单填写情况
            const notDone = this.videoParts.find((p) => p.progress < 100);
            if (notDone) {
                ElMessage.error('先等视频上传完啦');
                return;
            }
            if (this.videoParts.length === 0) {
                ElMessage.error('还没有选择视频哦');
                return;
            }
            if (this.form.title.trim().length === 0) {
                ElMessage.error('写个标题再上传吧');
                return;
            }
            if (this.form.tags.length === 0) {
                ElMessage.error('至少选一个标签哦');
                return;
            }
            // 强制选择真实分区（数字 id），避免沿用 teriteri 遗留的字符串占位分区
            this.ensureRealCategory();
            if (!this.categoryValid()) {
                ElMessage.error('请选择投稿分区');
                return;
            }
            // 提交使用归一后的真实分区 id
            const subCat = this.form.category[1];
            this.$store.state.isLoading = true;
            const submitComplete = async (coverBlobOrNull) => {
                // 1) 用真实元数据更新 upload-session（同一 client_id 命中 ON CONFLICT UPDATE）
                const videosMeta = this.videoParts.map((p) => ({
                    title: p.name.split(".mp4")[0],
                    total_chunks: this.createChunks(p.file).length,
                }));
                const totalChunks = videosMeta.reduce((s, v) => s + (v.total_chunks || 0), 0);
                const updateRes = await this.createSession({
                    client_id: this.sessionHash,
                    title: this.form.title,
                    description: this.form.descr,
                    category_id: this.form.category[1].id,
                    tags: this.form.tags.slice(),
                    videos: videosMeta,
                    total_chunks: totalChunks,
                });
                const updateBody = updateRes && (updateRes.data || updateRes);
                if (!updateBody || !updateBody.upload_id) {
                    ElMessage.error('更新稿件信息失败，请稍后再试');
                    this.$store.state.isLoading = false;
                    return;
                }
                // 2) 上传封面（如有）+ 合并分片创建稿件（后端按 partIndex 目录合并生成多个 video）
                const formData = new FormData();
                if (coverBlobOrNull) {
                    const cover = new File([coverBlobOrNull], this.sessionHash + Date.now() + '.jpg', { type: 'image/jpeg' });
                    formData.append('cover', cover);
                }
                formData.append('uploadId', this.uploadId);
                const res = await this.$post("/manuscript/upload-complete", formData, {
                    headers: {
                        'Content-Type': 'multipart/form-data',
                    }
                });
                const resBody = res && (res.data || res);
                if (resBody && resBody.code === 200) {
                    ElMessage.success('投稿成功，视频马上就能和大家见面啦');
                    this.init();
                    this.$emit("changeNavBarShow", true);
                    this.$store.state.isLoading = false;
                    setTimeout(() => {
                        this.$router.push('/platform/upload-manager/manuscript?status=processing');
                    }, 600);
                } else {
                    ElMessage.error('投稿失败，请稍后再试');
                    this.$store.state.isLoading = false;
                }
            };
            if (this.coverURL) {
                fetch(this.coverURL)
                .then(response => response.blob())
                .then(blob => submitComplete(blob))
                .catch(() => submitComplete(null))
            } else {
                submitComplete(null);
            }
        },


        // 创建上传会话（同一 client_id 重复调用会 ON CONFLICT 更新 title/desc/category/tags/videos/total_chunks）
        async createSession(payload) {
            return await this.$post("/manuscript/upload-session", payload);
        },

        // 查询上传进度（后端 GET 返回 uploaded_chunks = 已上传分片数 = 下一个待传序号）
        async askCurrentChunk(uploadId) {
            return await this.$get(`/manuscript/upload-session/${uploadId}`);
        },

        // 上传分片
        async uploadChunk(formData) {
            return await this.$post("/manuscript/upload-chunk", formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                }
            })
        },

        // 取消上传（DELETE 会话）
        async cancelUpload(uploadId) {
            return await this.$axios.delete(`/manuscript/upload-session/${uploadId}`);
        },
    }
}
</script>

<style scoped>
.videoUpload {
    padding-top: 8px;
    overflow: auto;
}

.cover-cut-content-pick-bar {
    margin-top: 20px;
    margin-left: 12px;
    width: 704px;
    height: 60px;
    user-select: none;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
}

.cover-slider {
    position: relative;
    background: #e7e7e7;
    box-shadow: 0 0 2px rgba(0,0,0,.04), 0 4px 10px rgba(0,0,0,.08);
    border-radius: 8px;
    height: 60px;
}

.slider-handle {
    position: absolute;
    z-index: 5;
    width: 12px;
    height: 66px;
    cursor: pointer;
    background: url('@/assets/teriteri/img/slider-handle.png');
    box-sizing: border-box;
    background-size: cover;
    border-radius: 4px;
    top: -3px;
    transform: translateX(-50%);
}

.cover-slider-image-list {
    display: flex;
    position: absolute;
    top: 0;
    width: 100%;
    justify-content: center;
}

.cover-slider-image {
    height: 60px;
    width: 100px;
}

.upload-body-content {
    max-width: 830px;
    margin: 0 auto;
    position: relative;
    width: 80%;
}

.upload-wrp {
    margin-top: 20px;
    position: relative;
    border: 2px dashed #ccc;
}

.video-input {
    margin-top: 62px;
    margin-bottom: 32px;
    display: flex;
    justify-content: center;
    position: relative;
    color: #999;
    font-size: 14px;
    text-align: center;
}

.pic-input {
    display: flex;
    justify-content: center;
    position: relative;
    color: #999;
    font-size: 14px;
    text-align: center;
}

.video-input-wrapper {
    flex: 1;
    cursor: pointer;
    padding-bottom: 20px;
    display: inline-block;
}

.pic-input-wrapper {
    flex: 1;
    cursor: pointer;
    padding: 70px 0;
    display: inline-block;
}

.icon-shangchuan {
    font-size: 40px;
}

.upload-tips-text {
    font-size: 13px;
    color: #999;
    margin-top: 6px;
}

.upload-btn {
    color: #fff;
    margin: 20px auto;
    width: 200px;
    height: 44px;
    cursor: pointer;
    background: var(--brand_pink);
    border-radius: 4px;
    text-align: center;
    line-height: 44px;
    white-space: nowrap;
}

.footer-item {
    margin-top: 18px;
    text-align: center;
    color: #99a2aa;
    fill: #99a2aa;
    font-size: 12px;
    line-height: 20px;
}

.agree-link {
    color: var(--brand_pink);
    cursor: pointer;
}

.agree-link:hover {
    color: var(--Pi4);
}

.info-grid {
    display: flex;
    justify-content: center;
    align-items: center;
    flex-wrap: wrap;
}

.footer-item .i-list {
    margin: 0 10px;
}

.footer-item .i-1 {
    cursor: pointer;
    position: relative;
    display: inline-flex;
    align-items: center;
}

.footer-item .i-1 .title {
    color: #99a2aa;
}

.footer-item .i-1 .title:hover {
    color: var(--brand_pink);
}

.info-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1px solid #99a2aa;
    color: #99a2aa;
    font-size: 10px;
    line-height: 1;
    margin-left: 4px;
    user-select: none;
    -webkit-user-select: none;
}

.info-icon:hover {
    border-color: var(--brand_pink);
    color: var(--brand_pink);
}

.footer-item .i-1 .info-pop-wrap {
    position: absolute;
    font-size: 12px;
    border: 1px solid #e7e7e7;
    color: #61666d;
    text-align: left;
    padding: 14px 16px;
    background-color: #fff;
    white-space: nowrap;
    z-index: 100;
    bottom: 26px;
    left: 50%;
    transform: translateX(-50%);
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0,0,0,.12);
    display: none;
}

.footer-item .i-1 .info-pop-show {
    display: inline-block;
}

.info-pop-wrap .info-pop-title {
    display: block;
    font-weight: 500;
    color: #18191c;
    margin-bottom: 6px;
}

.info-pop-wrap .info-pop-text {
    display: block;
    line-height: 20px;
    color: #61666d;
}

.footer-item .i-1 .info-pop-wrap:after {
    content: "";
    position: absolute;
    width: 12px;
    height: 12px;
    transform: rotate(45deg);
    box-shadow: 2px 2px 3px rgba(0,0,0,.08);
    border: 1px solid;
    border-color: transparent #e7e7e7 #e7e7e7 transparent;
    background-color: #fff;
    left: 50%;
    margin-left: -6px;
    bottom: -7px;
}

.info-content {
    padding: 2px 30px;
}

.info-header {
    position: relative;
    font-size: 16px;
    font-weight: 600;
    color: #212121;
    line-height: 28px;
    height: 70px;
    box-shadow: 0 1px 0 #e7e7e7;
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.cancel-btn {
    font-size: 14px;
    line-height: 1;
    padding: 12px 16px;
    color: var(--text2);
    cursor: pointer;
    border-radius: 4px;
    box-sizing: border-box;
    display: inline-block;
    text-align: center;
    border: 1px solid #e7e7e7;
    background-color: #fff;
    margin-right: 30px;
}

.cancel-btn:hover {
    background-color: #f4f4f4;
}

.video-name {
    padding-top: 25px;
    font-size: 14px;
    color: #212121;
}

.video-parts {
    padding-top: 15px;
}

.upload-summary {
    position: relative;
    padding: 12px 16px;
    border: 1px solid #e7e7e7;
    border-radius: 6px;
    margin-bottom: 12px;
    background: #fafafa;
    font-size: 13px;
    color: #61666d;
}

.summary-left {
    display: flex;
    align-items: center;
}

.summary-left .iconfont {
    color: var(--brand_pink);
    margin-right: 8px;
    font-size: 16px;
}

.summary-stage {
    color: #212121;
}

.summary-percent {
    margin-left: 8px;
    font-weight: 600;
    color: var(--brand_pink);
}

.summary-right {
    position: absolute;
    right: 16px;
    top: 12px;
    display: flex;
    align-items: center;
    gap: 14px;
    font-size: 12px;
    color: #99a2aa;
}

.summary-bar {
    margin-top: 10px;
    height: 4px;
    border-radius: 4px;
    background: #e7e7e7;
    overflow: hidden;
}

.summary-bar-inner {
    height: 100%;
    border-radius: 4px;
    background: var(--brand_pink);
    transition: width 0.5s ease;
}

.video-part-row {
    display: flex;
    align-items: flex-start;
    padding: 14px 16px;
    border: 1px solid #e7e7e7;
    border-radius: 6px;
    margin-bottom: 12px;
    background: #fafafa;
}

.video-part-row.part-failed {
    border-color: var(--stress_red);
}

.part-title-row {
    display: flex;
    align-items: center;
    margin-bottom: 6px;
}

.part-index {
    flex-shrink: 0;
    width: 34px;
    height: 22px;
    line-height: 22px;
    text-align: center;
    border-radius: 4px;
    background: var(--Pi0);
    color: var(--brand_pink);
    font-size: 13px;
    font-weight: 500;
    margin-right: 10px;
}

.part-name {
    font-size: 14px;
    color: #212121;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
}

.part-title-actions {
    display: flex;
    align-items: center;
    margin-left: auto;
    padding-left: 16px;
    flex-shrink: 0;
}

.part-title-actions .file-status-manage {
    margin-left: 14px;
    font-size: 13px;
}

.part-title-actions .file-status-manage.sort-disabled {
    color: #ccc;
    cursor: not-allowed;
    pointer-events: none;
}

.add-part-row {
    display: flex;
    align-items: center;
    margin-top: 4px;
}

.add-part-btn {
    display: inline-flex;
    align-items: center;
    height: 36px;
    padding: 0 18px;
    border: 1px dashed var(--brand_pink);
    border-radius: 4px;
    color: var(--brand_pink);
    font-size: 13px;
    cursor: pointer;
    background: #fff;
    user-select: none;
    transition: all .2s;
}

.add-part-btn .iconfont {
    margin-right: 6px;
    font-size: 14px;
}

.add-part-btn:hover {
    background: var(--Pi0);
}

.add-part-tip {
    margin-left: 14px;
    font-size: 12px;
    color: #99a2aa;
}

.video-file {
    padding-top: 15px;
    display: flex;
    width: 80%;
}

.video-icon-wrp {
    padding-right: 10px;
    width: 50px;
    height: 50px;
    display: flex;
    align-items: center;
}

.icon-video {
    font-size: 40px;
    color: var(--brand_pink);
}

.file-detail {
    flex: 1;
}

.success {
    color: var(--success_green);
}

.failed {
    color: var(--stress_red);
}

.file-status {
    display: flex;
    padding-top: 5px;
}

.file-status .iconfont {
    font-size: 20px;
    margin-right: 8px;
}

.file-status-text {
    flex: 1;
    height: 20px;
    font-size: 14px;
    color: #999;
    line-height: 28px;
    display: flex;
    align-items: center;
}

.file-status-manage {
    color: var(--brand_pink);
    cursor: pointer;
    user-select: none;
    font-size: 14px;
    display: flex;
    align-items: center;
    margin-left: 10px;
}

.file-progress {
    padding-top: 12px;
}

.file-progress-outer {
    height: 3px;
    width: 100%;
    border-radius: 4px;
    background-color: #e7e7e7;
}

.file-progress-inner {
    height: 3px;
    border-radius: 4px;
    transition: width 0.7s ease;
}

.doing-bg {
    background-color: var(--brand_pink);
}

.success-bg {
    background-color: var(--success_green);
}

.failed-bg {
    background-color: var(--stress_red);
}

.form-title {
    font-size: 16px;
    font-weight: 600;
    color: #222;
    margin-top: 30px;
}

.form-item {
    margin-top: 24px;
    color: #212121;
    display: flex;
    align-items: center;
}

.section-title {
    display: inline-flex;
    align-items: center;
    position: relative;
    flex-wrap: wrap;
    width: 134px;
}

.section-title-deg {
    font-size: 16px;
    color: #ff3b30;
    line-height: 16px;
    width: 12px;
}

.section-title-main {
    font-size: 14px;
    color: #212121;
    line-height: 21px;
    font-weight: 400;
}

.tag-wrp {
    display: flex;
    margin-top: 15px;
    width: 80%;
}

.tag-label {
    font-size: 14px;
    color: #222;
    padding: 14px 0 0 134px;
}

.tag-list {
    display: flex;
    flex-wrap: wrap;
    flex: 1;
}

.hot-tag-container {
    text-align: center;
    padding: 0 15px 0 16px;
    margin: 6px 0 6px 12px;
    height: 30px;
    color: #6d757a;
    cursor: pointer;
    border-radius: 4px;
    background: #f6f6f6;
    font-size: 12px;
    color: #212121;
    line-height: 14px;
    border: none;
}

.isTagSelected {
    background: var(--brand_pink);
    color: #fff;
    cursor: not-allowed;
    border: 0;
}

.hot-tag-item {
    font-size: 12px;
    line-height: 30px;
}

.type-radio {    
    display: flex;
    align-items: center;
}

.type-radio-item {
    cursor: pointer;
    margin-right: 20px;
    display: flex;
    align-items: center;
    position: relative;
}

.radio-box {
    border: 1px solid #bec3cc;
    height: 18px;
    width: 18px;
    border-radius: 50%;
    margin-right: 6px;
    padding: 3px;
}

.radio-box-selected {
    height: 100%;
    width: 100%;
    border-radius: 50%;
    background-color: var(--brand_pink);
}

.radio-name {
    font-size: 14px;
    color: var(--text2);
    line-height: 14px;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
}

.radio-selected {
    color: var(--brand_pink);
}

.auth {
    margin-left: 40px;
    color: #505050;
    font-size: 14px;
    display: flex;
    align-items: center;
    cursor: pointer;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
}

.auth-box {
    position: relative;
    display: inline-flex;
    width: 14px;
    height: 14.5px;
    border: 1px solid silver;
    border-radius: 2px;
    align-items: center;
    justify-content: center;
}

.auth-box:hover {
    border: 1px solid var(--brand_pink);
}

.icon-gou {
    color: #fff;
    font-size: 16px;
}

.auth-selected {
    background-color: var(--brand_pink);
}

.auth-main {
    padding-left: 8px;
}

.cover-container {
    width: 192px;
    height: 108px;
    border-radius: 4px;
    overflow: hidden; /* 隐藏超出容器的部分 */
}

.cover-container img {
    height: 100%;
    width: 100%;
    object-fit: cover; /* 居中填充容器，不会变形 */
}

.cover-right {
    display: flex;
    flex-direction: column;
    margin-left: 20px;
}

.change-cover-btn {
    margin-bottom: 4px;
}

.cover-tips {
    margin-top: 6px;
    font-weight: 400;
    font-size: 12px;
    line-height: 17px;
    color: #9499a0;
}

.reselect {
    margin-top: 8px;
    width: 94px;
}

.reselect .iconfont {
    font-size: 20px;
    margin-right: 8px;
}

.cover-dialog-header,
.cover-dialog-body,
.cover-dialog-footer {
    width: 800px;
}

.cover-dialog-header {
    border-bottom: 1px solid #e7e7e7;
}

.header-tab {
    padding: 0 32px;
    height: 61px;
    display: flex;
}

.header-tab-item {
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding-right: 29px;
}

.tab-text {
    color: #18191c;
    font-weight: 500;
    font-size: 16px;
    padding: 20px 0 18px 0;
}

.tab-line {
    height: 3px;
    width: 66px;
}

.currentTab .tab-text {
    color: var(--brand_pink);
}

.currentTab .tab-line {
    background-color: var(--brand_pink);
}

.desc-text {
    width: 80%;
    margin-top: 12px;
}

.submit-container {
    padding: 32px 0 32px 134px;
    position: relative;
}

.submit-draft {
    line-height: 38px;
    border: 1px solid #ccc;
    color: #505050;
    background: #fff;
}

.submit-add {
    margin-left: 16px;
    line-height: 40px;
    color: #fff;
    background-color: var(--brand_pink);
}

.submit-add:hover {
    background-color: var(--Pi4);
}

.submit-draft, .submit-add, .submit-drafts-box {
    display: inline-block;
    height: 40px;
    font-size: 14px;
    border-radius: 4px;
    text-align: center;
    vertical-align: middle;
    width: 120px;
    cursor: pointer;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
    transition: all .3s ease-in-out;
}

.submit-drafts-box {
    margin-left: 16px;
    line-height: 40px;
    border: 1px solid var(--brand_pink);
    color: var(--brand_pink);
    background: #fff;
}

.submit-drafts-box:hover {
    background: var(--Pi0);
}

.pick-version {
    padding: 0 30px;
    height: 450px;
    transition: height .5s;
}

.upload-version {
    padding: 0 30px;
    height: 372px;
    transition: height .5s;
}

.cover-dialog-footer {
    height: 86px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.cover-dialog-submit {
    width: 128px;
    font-size: 16px;
    padding: 12px 16px;
    color: #fff;
    background-color: var(--brand_pink);
    display: inline-block;
    text-align: center;
    cursor: pointer;
    border-radius: 4px;
    position: relative;
    line-height: 1;
}

/* element 元素 */
.cover-dialog /deep/ .el-dialog__body {
    margin: -30px 0 0 0;
    padding: 0;
}

.el-input {
    width: 80%;
    --el-input-hover-border-color: var(--brand_pink);
}

.el-input /deep/ .el-input__wrapper {
    padding: 4px 11px;
}

.el-textarea {
    --el-input-focus-border-color: #dcdfe6;
    --el-input-hover-border-color: #dcdfe6;
}

.el-textarea /deep/ .el-textarea__inner {
    height: 160px !important;
    resize: none;
    font-size: 13px;
    padding: 12px 15px;
    padding-right: 60px;
    line-height: 1.42;
    border-radius: 8px;
}

.el-textarea /deep/ .el-textarea__inner::-webkit-scrollbar {
    width: 8px;
    height: 8px;
}

.el-textarea /deep/ .el-textarea__inner::-webkit-scrollbar-thumb {
    border-radius: 10px;
    background-color: #ccc;
}

</style>