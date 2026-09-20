<template>
    <div class="videoDetail">
        <HeaderBar :isFixHeaderBar="true"></HeaderBar>
        <div class="video-container">
            <div class="left-container" :style="`width: ${playerSize.width}px;`">
                <!-- 标题 -->
                <div class="video-info-container">
                    <h1 :title="video.title" class="video-title">{{ video.title }}</h1>
                    <div class="video-info-detail">
                        <div class="video-info-detail-list">
                            <!-- 排名暂时没做，到时可能还得在数据库加一个属性 -->
                            <a href="/popular/rank/all" target="_blank" class="honor item honor-rank" v-if="video.top">
                                <i class="iconfont icon-paihang"></i>
                                <span class="honor-text">全站排行榜最高第1名</span>
                                <i class="iconfont icon-youjiantou"></i>
                            </a>
                            <span class="view item">
                                <i class="iconfont icon-bofangshu"></i>
                                &nbsp;{{ handleNum(view) }}&nbsp;
                            </span>
                            <span class="danmu item">
                                <i class="iconfont icon-danmushu"></i>
                                &nbsp;{{ handleNum(store.danmuList.length) }}&nbsp;
                            </span>
                            <span class="date item">
                                {{ video.uploadDate }}
                            </span>
                            <span class="copyright item" v-if="video.type === 1 && video.auth === 1">
                                <i class="iconfont icon-jinzhi"></i>
                                未经作者授权，禁止转载
                            </span>
                        </div>
                    </div>
                </div>
                <!-- 播放器组件 -->
                <VideoPlayer
                    ref="videoPlayerRef"
                    :current-manuscript-id="Number(manuscriptId)"
                    :manuscript-info="manuscriptInfoForPlayer"
                    :video-info="videoInfoForPlayer"
                    :current-p="currentPartIndex + 1"
                    :current-video-index="currentPartIndex"
                    :resume-time="0"
                    :danmu-list="danmuList"
                    :loading-danmus="false"
                    @update:video-info="handleVideoInfoUpdate"
                    @update:danmu-list="handleDanmuListUpdate"
                    @update:loading-danmus="() => {}"
                    @time-update="() => {}"
                />
                <!-- teriteri 风格状态栏 -->
                <div class="player-sending-area">
                    <div class="player-video-info">
                        <div class="player-video-info-text">
                            {{ population }} 人正在观看，已装填 {{ danmuList.length }} 条弹幕
                        </div>
                    </div>
                    <!-- ArtPlayer 内部 .art-controls-center DOM 元素搬到这里 -->
                    <div ref="artControlsCenterSlot" class="art-controls-center-slot"></div>
                </div>
                <!-- 三连转发 -->
                <div class="video-toolbar-container">
                    <div class="video-toolbar-left">
                        <div class="toolbar-left-item-wrap">
                            <div class="video-toolbar-left-item"
                                :class="{ 'on': store.attitudeToVideo.love }"
                                @click="loveOrNot(true, !store.attitudeToVideo.love)">
                                <i class="iconfont icon-dianzan"></i>
                                <span class="video-toolbar-item-text">{{ handleNum(good) }}</span>
                                <div class="dianzan-gif" :class="isGifShow ? 'gif-show' : 'gif-hide'">
                                    <img src="@/assets/teriteri/img/dianzan.gif" alt="" v-if="gifDisplay">
                                </div>
                            </div>
                        </div>
                        <div class="toolbar-left-item-wrap">
                            <div class="video-toolbar-left-item"
                                :class="{ 'on': store.attitudeToVideo.unlove }"
                                @click="loveOrNot(false, !store.attitudeToVideo.unlove)">
                                <i class="iconfont icon-diancai"></i>
                                <span class="video-toolbar-item-text">不喜欢</span>
                            </div>
                        </div>
                        <div class="toolbar-left-item-wrap">
                            <div class="video-toolbar-left-item"
                                :class="{ 'on': store.attitudeToVideo.coin > 0 }" @click="noPage">
                                <i class="iconfont icon-toubi"></i>
                                <span class="video-toolbar-item-text">{{ handleNum(coin) }}</span>
                            </div>
                        </div>
                        <div class="toolbar-left-item-wrap">
                            <div class="video-toolbar-left-item"
                                :class="{ 'on': store.attitudeToVideo.collect }" @click="openCollectDialog">
                                <i class="iconfont icon-shoucang1"></i>
                                <span class="video-toolbar-item-text">{{ handleNum(collect) }}</span>
                            </div>
                        </div>
                        <div class="toolbar-left-item-wrap">
                            <div class="video-toolbar-left-item" @click="noPage">
                                <i class="iconfont icon-zhuanfa"></i>
                                <span class="video-toolbar-item-text">{{ handleNum(share) }}</span>
                            </div>
                        </div>
                    </div>
                    <div class="video-toolbar-right">
                        <div class="video-toolbar-right-item">
                            <VPopover placement="top" popStyle="padding-bottom: 10px; z-index: 1000;">
                                <template #reference>
                                    <div class="video-tool-more">
                                        <i class="iconfont icon-gengduo"></i>
                                    </div>
                                </template>

                                <template #content>
                                    <div class="video-tool-more-dropdown">
                                        <div class="dropdown-item">
                                            <i class="iconfont icon-jubao1"></i>
                                            <span>举报稿件</span>
                                        </div>
                                    </div>
                                </template>
                            </VPopover>
                        </div>
                    </div>
                </div>
                <!-- 简介评论区 -->
                <div class="left-container-under-player">
                    <!-- 简介 -->
                    <div class="video-desc-container"
                        :style="(!video.descr || video.descr === '') ? 'display: none;' : ''">
                        <div class="basic-desc-info" :style="showAllDesc ? 'height: auto;' : 'height: 84px;'">
                            <span class="desc-info-text" v-html="handleLinkify(video.descr)"></span>
                        </div>
                        <div class="toggle-btn" v-if="descTooLong">
                            <span class="toggle-btn-text" @click="showAllDesc = !showAllDesc">
                                {{ showAllDesc ? '收起' : '展开更多' }}
                            </span>
                        </div>
                    </div>
                    <!-- 标签 -->
                    <div class="video-tag-container">
                        <div class="tag-container">
                            <a :href="`/v/${category.mcId}`" target="_blank" class="tag-link">{{ category.mcName }}</a>
                        </div>
                        <div class="tag-container">
                            <a :href="`/v/${category.mcId}/${category.scId}`" target="_blank" class="tag-link">
                                {{ category.scName }}</a>
                        </div>
                        <div class="tag-container" v-for="(item, index) in tags" :key="index">
                            <a :href="`/search/video?keyword=${item}`" target="_blank" class="tag-link">{{ item }}</a>
                        </div>
                    </div>
                    <!-- 评论 -->
                    <CommentVue :uid="user.uid" :count="comment"></CommentVue>

                </div>
            </div>
            <div class="right-container">
                <div class="right-container-inner">
                    <!-- UP主信息 -->
                    <div class="up-panel-container">
                        <div class="up-info-container">
                            <div class="up-info--left">
                                <div class="up-avatar-wrap">
                                    <VPopover popStyle="z-index: 2000; cursor: default; padding-top: 30px;">

                                        <template #reference>
                                            <a :href="`/space/${user.uid}`" target="_blank" class="up-avatar">
                                                <VAvatar :img="user.avatar_url" :size="48" :auth="user.auth"></VAvatar>
                                            </a>
                                        </template>

                                        <template #content>
                                            <UserCard :user="user"></UserCard>
                                        </template>
                                    </VPopover>
                                </div>
                            </div>
                            <div class="up-info--right">
                                <div class="up-info__detail">
                                    <div class="up-detail-top">
                                        <a :href="`/space/${user.uid}`" target="_blank" class="up-name"
                                            :class="user.vip !== 0 ? 'vip-name' : ''">{{ user.nickname }}</a>
                                        <a class="send-msg" @click="createChat">
                                            <i class="iconfont icon-xinfeng1"></i>
                                            发消息
                                        </a>
                                    </div>
                                    <div class="up-description" :title="user.description">{{ user.description }}</div>
                                </div>
                                <div class="up-info__btn-panel">
                                    <div class="default-btn follow-btn not-follow" v-if="true" @click="noPage">
                                        <i class="iconfont icon-jia"></i>
                                        关注 {{ handleNum(user.fansCount) }}
                                    </div>
                                    <VPopover popStyle="padding-top: 10px;">

                                        <template #reference>
                                            <div class="default-btn follow-btn following" v-if="false">
                                                <i class="iconfont icon-caidan"></i>
                                                已关注 {{ handleNum(user.fansCount) }}
                                            </div>
                                        </template>

                                        <template #content>
                                            <div class="following-dropdown">
                                                <div class="dropdown-item">
                                                    <span>设置分组</span>
                                                </div>
                                                <div class="dropdown-item">
                                                    <span>取消关注</span>
                                                </div>
                                            </div>
                                        </template>
                                    </VPopover>
                                    <VPopover popStyle="padding-top: 10px;">

                                        <template #reference>
                                            <div class="default-btn follow-btn following" v-if="false">
                                                <i class="iconfont icon-caidan"></i>
                                                已互粉 {{ handleNum(user.fansCount) }}
                                            </div>
                                        </template>

                                        <template #content>
                                            <div class="following-dropdown">
                                                <div class="dropdown-item">
                                                    <span>设置分组</span>
                                                </div>
                                                <div class="dropdown-item">
                                                    <span>取消关注</span>
                                                </div>
                                            </div>
                                        </template>
                                    </VPopover>
                                </div>
                            </div>
                        </div>
                    </div>
                    <!-- 弹幕组件 -->
                    <DanmuBox :boxHeight="playerSize.height" :authorId="user.uid"
                        @jump="(time) => jumpTimePoint = time">
                    </DanmuBox>
                    <!-- 分P列表 -->
                    <div class="video-parts-list" v-if="manuscriptParts.length > 1">
                        <p class="parts-title">视频分P ({{ manuscriptParts.length }}P)</p>
                        <div class="parts-items">
                            <div
                                v-for="(part, index) in manuscriptParts"
                                :key="part.vid"
                                class="part-item"
                                :class="{ 'active': index === currentPartIndex }"
                                @click="switchPart(index)"
                            >
                                <span class="part-index">P{{ index + 1 }}</span>
                                <span class="part-title" :title="part.title">{{ part.title }}</span>
                                <span class="part-duration">{{ handleDuration(part.duration) }}</span>
                            </div>
                        </div>
                    </div>
                    <!-- 相关视频列表 -->
                    <div class="recommend-list">
                        <div class="next-play">
                            <p class="rec-title">
                                接下来播放
                                <span class="next-button" @click="autonext = !autonext">
                                    <span class="txt">自动连播</span>
                                    <span class="switch-button" :class="{ 'on': autonext }"></span>
                                </span>
                            </p>
                            <!-- 视频卡片 -->
                            <div class="video-page-card-small" v-if="recommendVideos.length > 0">
                                <div class="card-box">
                                    <div class="pic-box">
                                        <div class="pic" @click="changeVideo(recommendVideos[0].video.vid)">
                                            <img :src="recommendVideos[0].video.coverUrl" alt="">
                                            <span class="duration">
                                                {{ handleDuration(recommendVideos[0].video.duration) }}
                                            </span>
                                        </div>
                                    </div>
                                    <div class="info">
                                        <p class="title" @click="changeVideo(recommendVideos[0].video.vid)">
                                            {{ recommendVideos[0].video.title }}
                                        </p>
                                        <a :href="`/space/${recommendVideos[0].user.uid}`" target="_blank"
                                            class="upname">
                                            <svg t="1703614018039" class="icon" viewBox="0 0 1024 1024" version="1.1"
                                                xmlns="http://www.w3.org/2000/svg" p-id="4221" width="18" height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="4222"></path>
                                                <path
                                                    d="M419.2 544c0 51.2-3.2 108.8-83.2 108.8S252.8 595.2 252.8 544v-217.6H192v243.2c0 96 51.2 140.8 140.8 140.8 89.6 0 147.2-48 147.2-144v-240h-60.8V544zM710.4 326.4h-156.8V704h60.8v-147.2h96c102.4 0 121.6-67.2 121.6-115.2 0-44.8-19.2-115.2-121.6-115.2z m-3.2 179.2h-92.8V384h92.8c32 0 60.8 12.8 60.8 60.8 0 44.8-32 60.8-60.8 60.8z"
                                                    p-id="4223"></path>
                                            </svg>
                                            <span class="name">{{ recommendVideos[0].user.nickname }}</span>
                                        </a>
                                        <div class="playinfo">
                                            <svg t="1703614610134" class="playinfo-play" viewBox="0 0 1024 1024"
                                                version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4864" width="18"
                                                height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="4865" fill="#aaaaaa"></path>
                                                <path
                                                    d="M684.8 483.2l-256-112c-22.4-9.6-44.8 6.4-44.8 28.8v224c0 22.4 22.4 38.4 44.8 28.8l256-112c25.6-9.6 25.6-48 0-57.6z"
                                                    p-id="4866"></path>
                                            </svg>
                                            {{ handleNum(recommendVideos[0].stats.play) }}
                                            <svg t="1703614725224" class="playinfo-dm" viewBox="0 0 1024 1024"
                                                version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="5601"
                                                id="mx_n_1703614725225" width="18" height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="5602"></path>
                                                <path
                                                    d="M240 384h64v64h-64zM368 384h384v64h-384zM432 576h352v64h-352zM304 576h64v64h-64z"
                                                    p-id="5603"></path>
                                            </svg>
                                            {{ handleNum(recommendVideos[0].stats.danmu) }}
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <!-- 分隔线 -->
                            <div class="split-line"></div>
                        </div>
                        <div class="rec-list" v-if="recommendVideos.length > 1">
                            <!-- 视频卡片 -->
                            <div class="video-page-card-small" v-for="(item, index) in recommendVideos.slice(1,)"
                                :key="index">
                                <div class="card-box">
                                    <div class="pic-box">
                                        <div class="pic" @click="changeVideo(item.video.vid)">
                                            <img :src="item.video.coverUrl" alt="">
                                            <span class="duration">{{ handleDuration(item.video.duration) }}</span>
                                        </div>
                                    </div>
                                    <div class="info">
                                        <p class="title" @click="changeVideo(item.video.vid)">{{ item.video.title }}</p>
                                        <a :href="`/space/${item.user.uid}`" target="_blank" class="upname">
                                            <svg t="1703614018039" class="icon" viewBox="0 0 1024 1024" version="1.1"
                                                xmlns="http://www.w3.org/2000/svg" p-id="4221" width="18" height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="4222"></path>
                                                <path
                                                    d="M419.2 544c0 51.2-3.2 108.8-83.2 108.8S252.8 595.2 252.8 544v-217.6H192v243.2c0 96 51.2 140.8 140.8 140.8 89.6 0 147.2-48 147.2-144v-240h-60.8V544zM710.4 326.4h-156.8V704h60.8v-147.2h96c102.4 0 121.6-67.2 121.6-115.2 0-44.8-19.2-115.2-121.6-115.2z m-3.2 179.2h-92.8V384h92.8c32 0 60.8 12.8 60.8 60.8 0 44.8-32 60.8-60.8 60.8z"
                                                    p-id="4223"></path>
                                            </svg>
                                            <span class="name">{{ item.user.nickname }}</span>
                                        </a>
                                        <div class="playinfo">
                                            <svg t="1703614610134" class="playinfo-play" viewBox="0 0 1024 1024"
                                                version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4864" width="18"
                                                height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="4865" fill="#aaaaaa"></path>
                                                <path
                                                    d="M684.8 483.2l-256-112c-22.4-9.6-44.8 6.4-44.8 28.8v224c0 22.4 22.4 38.4 44.8 28.8l256-112c25.6-9.6 25.6-48 0-57.6z"
                                                    p-id="4866"></path>
                                            </svg>
                                            {{ handleNum(item.stats.play) }}
                                            <svg t="1703614725224" class="playinfo-dm" viewBox="0 0 1024 1024"
                                                version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="5601"
                                                id="mx_n_1703614725225" width="18" height="18">
                                                <path
                                                    d="M800 128H224C134.4 128 64 198.4 64 288v448c0 89.6 70.4 160 160 160h576c89.6 0 160-70.4 160-160V288c0-89.6-70.4-160-160-160z m96 608c0 54.4-41.6 96-96 96H224c-54.4 0-96-41.6-96-96V288c0-54.4 41.6-96 96-96h576c54.4 0 96 41.6 96 96v448z"
                                                    p-id="5602"></path>
                                                <path
                                                    d="M240 384h64v64h-64zM368 384h384v64h-384zM432 576h352v64h-352zM304 576h64v64h-64z"
                                                    p-id="5603"></path>
                                            </svg>
                                            {{ handleNum(item.stats.danmu) }}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <!-- 收藏框 -->
        <el-dialog v-model="collectVisible" :close-on-click-modal="false" destroy-on-close align-center>
            <AddToFavorite :lastSelected="collectedFids" :vid="video.vid ? video.vid : 0" @collected="updateCollect">
            </AddToFavorite>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTeriteriStore } from '@/stores/teriteri'
import { handleTime, handleNum as _handleNum, handleDate, linkify } from '@/teriteri-src/utils/utils'
import { ElMessage } from 'element-plus'

import CommentVue from '@/components/teriteri/comment/CommentVue.vue'
import HeaderBar from '@/components/teriteri/headerBar/HeaderBar.vue'
import VideoPlayer from '@/components/VideoPlayer.vue'
import VPopover from '@/components/teriteri/popover/VPopover.vue'
import VAvatar from '@/components/teriteri/avatar/VAvatar.vue'
import UserCard from '@/components/teriteri/UserCard/UserCard.vue'
import DanmuBox from '@/components/teriteri/danmu/DanmuBox.vue'
import AddToFavorite from '@/components/teriteri/favorite/AddToFavorite.vue'

const store = useTeriteriStore()
const route = useRoute()
const router = useRouter()

// ===== State =====
const socket = ref<WebSocket | null>(null)
const playerSize = reactive({ width: 704, height: 442 })
const video = ref<Record<string, any>>({})
const view = ref(0)
const danmu = ref(0)
const good = ref(0)
const coin = ref(0)
const collect = ref(0)
const share = ref(0)
const comment = ref(0)
const population = ref(0)
const user = ref<Record<string, any>>({ uid: 0 })
const category = ref<Record<string, any>>({})
const tags = ref<string[]>([])
const showAllDesc = ref(true)
const descTooLong = ref(false)
const jumpTimePoint = ref(-1)
const autonext = ref(false)
const recommendVideos = ref<any[]>([])
const vids = ref<number[]>([])
const isGifShow = ref(false)
const gifDisplay = ref(false)
const collectVisible = ref(false)
const collectedFids = ref<Set<number>>(new Set())
const isMounted = ref(false)
const loveLoading = ref(false)
const manuscriptParts = ref<any[]>([])
const currentPartIndex = ref(0)
const manuscriptId = ref('')
const danmuList = ref<any[]>([])
const videoPlayerRef = ref<any>(null)
const artControlsCenterSlot = ref<HTMLDivElement | null>(null)

// ===== Computed =====
const manuscriptInfoForPlayer = computed(() => ({
    id: Number(manuscriptId.value) || 0,
    title: video.value.title || '',
    description: video.value.descr || '',
    coverUrl: video.value.coverUrl || '',
    tags: tags.value,
    videos: manuscriptParts.value.map((p: any, i: number) => ({
        id: Number(p.vid) || 0,
        title: p.title || '',
        playUrl: p.playUrl || '',
        playUrlHd: p.playUrl || '',
        playUrlSd: '',
        playUrlLd: '',
        duration: p.duration || 0,
        videoOrder: i,
    })),
}))

const videoInfoForPlayer = computed(() => {
    const current = manuscriptParts.value[currentPartIndex.value] || {}
    return {
        title: current.title || video.value.title || '',
        coverUrl: video.value.coverUrl || '',
        playUrl: current.playUrl || video.value.videoUrl || '',
        playUrlHd: current.playUrl || '',
        playUrlSd: '',
        playUrlLd: '',
        duration: video.value.duration || 0,
        watchingCount: population.value,
        danmuLoadedCount: danmu.value,
    }
})

// ===== Utility wrappers (keep template binding names) =====
function handleNum(number: any) {
    return _handleNum(number)
}

function handleDuration(time: number) {
    return handleTime(time)
}

function handleLinkify(text: string) {
    return linkify(text)
}

function isDescTooLong() {
    nextTick(() => {
        const desc = document.querySelector('.basic-desc-info') as HTMLElement
        if (desc && desc.clientHeight > 84) {
            descTooLong.value = true
            showAllDesc.value = false
        }
    })
}

function openNewPage(routePath: string) {
    window.open(router.resolve(routePath).href, '_blank')
}

// ===== Request functions =====
async function getVideoDetail() {
    const { get } = await import('@/teriteri-src/network/request')
    const res = await get('/video/getone', {
        params: { vid: route.params.vid },
    })
    if (res.data.code === 404) {
        router.push('/404')
        return false
    }
    if (res.data.data) {
        video.value = res.data.data.video
        user.value = res.data.data.user
        category.value = res.data.data.category
        tags.value = res.data.data.video.tags.split('\r\n').filter((tag: string) => tag.trim() !== '')
        view.value = res.data.data.stats.play
        danmu.value = res.data.data.stats.danmu
        good.value = res.data.data.stats.good
        coin.value = res.data.data.stats.coin
        collect.value = res.data.data.stats.collect
        share.value = res.data.data.stats.share
        comment.value = res.data.data.stats.comment
        manuscriptId.value = res.data.data.manuscriptId || ''
        manuscriptParts.value = res.data.data.videos || []
        const currentVid = String(route.params.vid)
        const idx = manuscriptParts.value.findIndex((p: any) => p.vid === currentVid)
        currentPartIndex.value = idx >= 0 ? idx : 0
    }
    isDescTooLong()
    if (localStorage.getItem('teri_token')) {
        getCollectedFids()
    }
    return true
}

async function getRecommendVideos() {
    const { get } = await import('@/teriteri-src/network/request')
    recommendVideos.value = []
    vids.value = []
    vids.value.push(Number(route.params.vid))
    let ids = vids.value.join(',')
    const res = await get('/video/cumulative/visitor', {
        params: { vids: ids },
    })
    if (res.data.data) {
        recommendVideos.value.push(...res.data.data.videos)
        vids.value.push(...res.data.data.vids)
        ids = vids.value.join(',')
        const res2 = await get('/video/cumulative/visitor', {
            params: { vids: ids },
        })
        if (res2.data.data) {
            recommendVideos.value.push(...res2.data.data.videos)
            vids.value.push(...res2.data.data.vids)
        }
    }
}

async function getDanmuList() {
    const { get } = await import('@/teriteri-src/network/request')
    const res = await get(`/danmu-list/${route.params.vid}`)
    if (res.data.data == null || res.data.data.length === 0) {
        store.updateDanmuList([])
    } else if (res.data.data.length > 0) {
        store.updateDanmuList(res.data.data)
    }
}

async function initWebsocket() {
    const wsBaseUrl = process.env.VUE_APP_WS_DANMU_URL
    if (!wsBaseUrl) return
    const socketUrl = `${wsBaseUrl}/ws/danmu/${route.params.vid}`
    if (socket.value != null) {
        await socket.value.close()
        socket.value = null
    }
    socket.value = new WebSocket(socketUrl)
    socket.value.addEventListener('close', handleWsClose)
    socket.value.addEventListener('message', handleWsMessage)
    socket.value.addEventListener('error', handleWsError)
}

async function closeWebSocket() {
    if (socket.value != null) {
        await socket.value.close()
        socket.value = null
    }
}

async function loveOrNot(isLove: boolean, isSet: boolean) {
    if (loveLoading.value) return
    if (!store.user.uid) {
        store.openLogin = true
        nextTick(() => {
            store.openLogin = false
        })
        return
    }
    if (!video.value.vid) {
        ElMessage.error('视频不存在')
        return
    }
    loveLoading.value = true
    const originalLove = store.attitudeToVideo.love
    const { post } = await import('@/teriteri-src/network/request')
    const formData = new FormData()
    formData.append('vid', String(Number(video.value.vid)))
    formData.append('isLove', String(isLove))
    formData.append('isSet', String(isSet))
    const res = await post('/video/love-or-not', formData, {
        headers: { Authorization: 'Bearer ' + localStorage.getItem('teri_token') },
    })
    if (!res.data.data) {
        loveLoading.value = false
        return
    }
    const data = res.data.data
    const atv = {
        love: data.love === 1,
        unlove: data.unlove === 1,
        coin: data.coin,
        collect: data.collect === 1,
    }
    store.updateAttitudeToVideo(atv)
    if (isLove && isSet) {
        good.value++
        gifShow()
        setTimeout(() => {
            gifHide()
        }, 3000)
    } else if (isLove || (!isLove && isSet && originalLove)) {
        good.value = good.value - 1 < 0 ? 0 : good.value - 1
    }
    loveLoading.value = false
}

async function getCollectedFids() {
    const { get } = await import('@/teriteri-src/network/request')
    const res = await get('/video/collected-fids', {
        params: { vid: Number(video.value.vid) },
        headers: { Authorization: 'Bearer ' + localStorage.getItem('teri_token') },
    })
    if (!res.data) return
    collectedFids.value = new Set(res.data.data)
}

// ===== Event handlers =====
function createChat() {
    if (!store.user.uid) {
        store.openLogin = true
        nextTick(() => {
            store.openLogin = false
        })
        return
    }
    openNewPage(`/message/whisper/${user.value.uid}`)
}

function handleScroll() {
    const windowHeight = window.innerHeight
    const leftPart = document.querySelector('.left-container') as HTMLElement
    const rightPart = document.querySelector('.right-container-inner') as HTMLElement
    if (leftPart) {
        if (leftPart.clientHeight <= windowHeight - 64) {
            leftPart.style.top = '64px'
        } else {
            leftPart.style.top = `-${leftPart.clientHeight - windowHeight}px`
        }
    }
    if (rightPart) {
        if (rightPart.clientHeight <= windowHeight - 64) {
            rightPart.style.top = '64px'
        } else {
            rightPart.style.top = `-${rightPart.clientHeight - windowHeight}px`
        }
    }
}

function handleVideoInfoUpdate(_info: any) {
    // 播放器更新 videoInfo 时同步回来
}

function handleDanmuListUpdate(list: any[]) {
    danmuList.value = list
}

function changeWindowSize() {
    const viewportWidth = document.documentElement.clientWidth || window.innerWidth
    const rightContainer = document.querySelector('.right-container')
    const rightWidth = rightContainer ? rightContainer.getBoundingClientRect().width : 350
    const gap = 20
    const maxLeftWidth = Math.max(320, viewportWidth - rightWidth - gap)
    const windowHeight = window.innerHeight
    let height = (windowHeight - 64) * 0.7
    let width = height * (16 / 9)
    if (width > maxLeftWidth) {
        width = maxLeftWidth
        height = width * (9 / 16)
    }
    height = Math.max(360, Math.min(720, height))
    playerSize.width = width
    playerSize.height = height
    const playerEl = document.querySelector('.video-player')
    if (playerEl) {
        const minRequired = 632
        const actualWidth = width
        const scale = Math.max(0.5, Math.min(1, actualWidth / minRequired))
        const root = playerEl.querySelector('.art-video-player') as HTMLElement
        if (root) {
            root.style.setProperty('--art-control-height', `${46 * scale}px`)
            root.style.setProperty('--art-control-icon-size', `${30 * scale}px`)
            root.style.setProperty('--art-padding', `${10 * scale}px`)
            root.style.setProperty('--art-bottom-gap', `${5 * scale}px`)
            root.style.setProperty('--art-control-opacity', '0.75')
        }
    }
}

function moveArtControlsCenter() {
    const slot = artControlsCenterSlot.value
    if (!slot) return
    const center = document.querySelector('.art-video-player .art-controls-center')
    if (center) {
        slot.appendChild(center)
        ;(center as HTMLElement).style.display = 'flex'
        ;(center as HTMLElement).style.flex = '0 0 auto'
        ;(center as HTMLElement).style.padding = '0'
        ;(center as HTMLElement).style.height = 'auto'
        ;(center as HTMLElement).style.color = '#61666D'
        ;(center as HTMLElement).style.fill = '#61666D'
        const svgs = center.querySelectorAll('svg')
        svgs.forEach((svg: SVGElement) => {
            svg.style.fill = '#61666D'
            const paths = svg.querySelectorAll('path')
            paths.forEach((p: SVGPathElement) => {
                p.style.fill = '#61666D'
            })
        })
        const observer = new MutationObserver(() => {
            const visible = center.getAttribute('data-danmuku-visible')
            if (visible !== null) {
                slot.setAttribute('data-danmuku-visible', visible)
            }
        })
        observer.observe(center, { attributes: true, attributeFilter: ['data-danmuku-visible'] })
        const initialVisible = center.getAttribute('data-danmuku-visible')
        if (initialVisible !== null) {
            slot.setAttribute('data-danmuku-visible', initialVisible)
        }
    }
}

function handleWsClose() {
    setTimeout(() => {
        if (!socket.value) {
            initWebsocket()
        }
    }, 2000)
}

function handleWsMessage(e: MessageEvent) {
    if (e.data === '登录已过期') {
        ElMessage.error(e.data)
    } else if (e.data.startsWith('当前观看人数')) {
        const numberPart = e.data.substring(6).trim()
        population.value = parseInt(numberPart, 10)
    } else {
        const dm = JSON.parse(e.data)
        store.danmuList.push(dm)
    }
}

function handleWsError(e: Event) {
    console.log('弹幕websocket信道报错: ', e)
}

function sendDanmu(dm: any) {
    if (!localStorage.getItem('teri_token')) {
        store.openLogin = true
        nextTick(() => {
            store.openLogin = false
        })
        return
    }
    const dmJson = JSON.stringify({
        token: 'Bearer ' + localStorage.getItem('teri_token'),
        data: dm,
    })
    socket.value?.send(dmJson)
}

async function changeVideo(vid: string | number) {
    await router.push(`/video/${vid}`)
    await initWebsocket()
    if (await getVideoDetail()) {
        await getDanmuList()
        await getRecommendVideos()
    }
}

function switchPart(index: number) {
    if (index === currentPartIndex.value) return
    const part = manuscriptParts.value[index]
    if (!part) return
    changeVideo(part.vid)
}

function next() {
    if (recommendVideos.value[0]) {
        changeVideo(recommendVideos.value[0].video.vid)
    }
}

function gifShow() {
    gifDisplay.value = true
    isGifShow.value = true
}

function gifHide() {
    isGifShow.value = false
    setTimeout(() => {
        gifDisplay.value = false
    }, 300)
}

function openCollectDialog() {
    if (!store.user.uid) {
        store.openLogin = true
        nextTick(() => {
            store.openLogin = false
        })
        return
    }
    if (!video.value.vid) {
        ElMessage.error('视频不存在')
        return
    }
    collectVisible.value = true
}

function updateCollect(info: { fids: Set<number>; num: number }) {
    collectedFids.value = info.fids
    collect.value += info.num
    collectVisible.value = false
}

function noPage() {
    ElMessage.warning('该功能暂未开放')
}

// ===== created() — top-level await =====
changeWindowSize()
if (localStorage.getItem('playerSetting')) {
    const setting = JSON.parse(localStorage.getItem('playerSetting')!)
    autonext.value = setting.autonext
}
await initWebsocket()
if (await getVideoDetail()) {
    await getDanmuList()
    await getRecommendVideos()
}

// ===== mounted() =====
onMounted(() => {
    window.addEventListener('scroll', handleScroll)
    handleScroll()
    window.addEventListener('resize', changeWindowSize)
    if (window.visualViewport) {
        window.visualViewport.addEventListener('resize', changeWindowSize)
    }
    window.addEventListener('beforeunload', closeWebSocket)
    nextTick(() => {
        let attempts = 0
        const tryMove = () => {
            const center = document.querySelector('.art-video-player .art-controls-center')
            if (center) {
                moveArtControlsCenter()
            } else if (attempts++ < 20) {
                setTimeout(tryMove, 250)
            }
        }
        tryMove()
    })
    setTimeout(() => {
        isMounted.value = true
    }, 3000)
})

// ===== beforeUnmount() =====
onBeforeUnmount(async () => {
    await closeWebSocket()
    window.removeEventListener('beforeunload', closeWebSocket)
    window.removeEventListener('scroll', handleScroll)
    window.removeEventListener('resize', changeWindowSize)
    if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', changeWindowSize)
    }
})

// ===== watch =====
watch(
    () => route.path,
    () => {
        collectVisible.value = false
    },
)

watch(
    () => store.isLogin,
    (curr) => {
        if (isMounted.value && curr) {
            getCollectedFids()
        } else if (!curr) {
            collectedFids.value = new Set()
        }
    },
)
</script>

<style scoped>
.video-container {
    width: auto;
    padding: 64px 10px 0px;
    max-width: 2540px;
    margin: 0 auto;
    display: flex;
    justify-content: center;
    box-sizing: content-box;
    position: relative;
    overflow: hidden;
}

.left-container {
    position: sticky;
    height: fit-content;
    max-width: 100%;
    min-width: 0;
    overflow: hidden;
    flex-shrink: 1;
}

.video-info-container {
    height: 104px;
    box-sizing: border-box;
    padding-top: 22px;
}

.video-title {
    font-size: 20px;
    font-weight: 500;
    -webkit-font-smoothing: antialiased;
    color: var(--text1);
    line-height: 28px;
    margin-bottom: 6px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.video-info-detail {
    font-size: 13px;
    color: var(--text3);
    display: flex;
    align-items: center;
    height: 24px;
    line-height: 18px;
    position: relative;
    overflow: hidden;
}

.video-info-detail-list {
    display: flex;
    align-items: center;
    overflow: hidden;
    box-sizing: border-box;
}

.video-info-detail-list .item {
    flex-shrink: 0;
    margin-right: 12px;
    overflow: hidden;
}

.video-info-detail-list .item:last-child {
    margin-right: 0;
}

.view,
.danmu,
.copyright {
    display: inline-flex;
    align-items: center;
}

.icon-bofangshu .icon-danmushu {
    font-size: 18px;
}

.honor {
    display: inline-flex;
    align-items: center;
    font-size: 13px;
    height: 24px;
    border-radius: 2px;
    padding: 0px 6px;
}

.honor.honor-rank {
    color: var(--brand_pink);
    background-color: rgba(255, 102, 153, 0.1);
}

.honor .icon-paihang {
    font-size: 14px;
    margin: 0 5px 0 3px;
}

.honor .icon-youjiantou {
    font-size: 14px;
}

.date {
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
    line-height: 24px;
    font-size: 13px;
    height: 100%;
    display: inline-block;
    vertical-align: middle;
}

.icon-jinzhi {
    font-size: 12px;
    margin-right: 4px;
    color: var(--stress_red);
}

.video-toolbar-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 16px;
    padding-bottom: 12px;
    line-height: 28px;
    border-bottom: 1px solid var(--line_regular);
}

.video-toolbar-left {
    position: relative;
    display: flex;
    align-items: center;
    -webkit-user-select: none;
    user-select: none;
}

.toolbar-left-item-wrap {
    position: relative;
    margin-right: 8px;
}

.video-toolbar-left-item {
    position: relative;
    display: -ms-flexbox;
    display: flex;
    -ms-flex-align: center;
    align-items: center;
    width: 92px;
    white-space: nowrap;
    transition: all .3s;
    font-size: 13px;
    color: var(--text2);
    font-weight: 500;
    cursor: pointer;
}

.video-toolbar-left-item.on,
.video-toolbar-left-item:hover {
    color: var(--brand_pink);
}

.video-toolbar-left-item .iconfont {
    margin-right: 8px;
    font-size: 26px;
}

.video-toolbar-left-item .icon-diancai {
    transform: translateY(2px);
}

.video-toolbar-item-text {
    overflow: hidden;
    text-overflow: ellipsis;
    word-break: break-word;
    white-space: nowrap;
}

.dianzan-gif {
    position: absolute;
    top: -50px;
    left: -10px;
    height: 40px;

}

.dianzan-gif img {
    height: 100%;
}

.gif-hide {
    animation: disappear 0.2s ease-out forwards;
    transform-origin: bottom;
}

.gif-show {
    animation: appear 0.2s ease-out forwards;
    transform-origin: bottom;
}

@keyframes appear {
    0% {
        opacity: 0;
        transform: translateY(5px) scale(0);
    }

    100% {
        opacity: 1;
        transform: translateY(0) scale(1);
    }
}

@keyframes disappear {
    0% {
        opacity: 1;
        transform: translateY(0) scale(1);
    }

    100% {
        opacity: 0;
        transform: translateY(5px) scale(0);
    }
}

.video-toolbar-right {
    display: flex;
    align-items: center;
    -webkit-user-select: none;
    user-select: none;
}

.video-toolbar-right-item {
    display: -ms-inline-flexbox;
    display: inline-flex;
    -ms-flex-align: center;
    align-items: center;
    font-size: 13px;
    color: var(--text2);
    transition: all .3s;
    cursor: pointer;
}

.video-toolbar-right-item:hover {
    color: var(--brand_pink);
}

.video-tool-more {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 30px;
}

.icon-gengduo {
    font-size: 18px;
}

.video-tool-more-dropdown {
    padding: 12px 0px;
    cursor: auto;
}

.dropdown-item {
    position: relative;
    display: flex;
    align-items: center;
    height: 40px;
    width: 120px;
    padding: 0 20px;
    color: var(--text1);
    cursor: pointer;
}

.dropdown-item:hover {
    background-color: var(--Ga1);
}

.icon-jubao1 {
    margin-right: 10px;
}

.video-desc-container {
    margin: 16px 0;
}

.basic-desc-info {
    white-space: pre-line;
    letter-spacing: 0;
    color: var(--text1);
    font-size: 15px;
    line-height: 24px;
    overflow: hidden;
}

.toggle-btn {
    margin-top: 10px;
    font-size: 13px;
    line-height: 18px;
}

.toggle-btn-text {
    cursor: pointer;
    color: var(--text2);
}

.toggle-btn-text:hover {
    color: var(--brand_pink);
}

.video-tag-container {
    padding-bottom: 6px;
    margin: 16px 0 20px 0;
    border-bottom: 1px solid var(--line_regular);
    display: flex;
    flex-wrap: wrap;
}

.tag-container {
    margin: 0px 12px 8px 0;
}

.tag-link {
    color: var(--text2);
    background: var(--graph_bg_regular);
    height: 28px;
    line-height: 28px;
    border-radius: 14px;
    font-size: 13px;
    padding: 0 12px;
    box-sizing: border-box;
    transition: all .3s;
    display: -ms-inline-flexbox;
    display: inline-flex;
    -ms-flex-align: center;
    align-items: center;
    cursor: pointer;
}

.right-container {
    width: 350px;
    flex: none;
    margin-left: 30px;
    position: relative;
    pointer-events: none;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
}

.right-container-inner {
    padding-bottom: 250px;
    position: sticky;
    overflow: hidden;
    max-width: 100%;
}

.right-container-inner * {
    pointer-events: all;
    max-width: 100%;
}

.up-info-container {
    box-sizing: border-box;
    height: 104px;
    display: flex;
    align-items: center;
}

.up-avatar-wrap {
    width: 48px;
    height: 48px;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
}

.up-avatar {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    background-color: var(--graph_weak);
}

.up-info--right {
    margin-left: 12px;
    flex: 1;
}

.up-info__detail {
    margin-bottom: 5px;
}

.up-detail-top {
    display: flex;
    align-items: center;
}

.up-name {
    font-size: 15px;
    color: var(--text1);
    font-weight: 500;
    position: relative;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
    margin-right: 12px;
    max-width: calc(100% - 12px - 56px);
}

.send-msg {
    color: var(--text2);
    font-size: 13px;
    transition: color 0.3s;
    flex-shrink: 0;
    cursor: pointer;
}

.send-msg:hover {
    color: var(--brand_pink);
}

.up-description {
    margin-top: 2px;
    font-size: 13px;
    line-height: 16px;
    height: 16px;
    color: var(--text3);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.up-info__btn-panel {
    clear: both;
    display: flex;
    margin-top: 5px;
    white-space: nowrap;
}

.up-info__btn-panel .default-btn {
    box-sizing: border-box;
    padding: 0;
    line-height: 30px;
    height: 30px;
    border-radius: 6px;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    background: var(--graph_weak);
    position: relative;
    transition: 0.3s all;
}

.follow-btn {
    width: 200px;
}

.follow-btn.following {
    color: var(--text3);
    background-color: var(--graph_bg_thick);
}

.follow-btn.following:hover {
    background-color: var(--graph_bg_regular);
}

.follow-btn.not-follow {
    background: var(--brand_pink);
    color: var(--text_white);
}

.follow-btn.not-follow:hover {
    background: var(--Pi4);
}

.follow-btn .iconfont {
    font-size: 14px;
    margin-right: 2px;
}

.following-dropdown {
    padding: 8px 0px;
}

.following-dropdown .dropdown-item:hover {
    color: var(--brand_pink);
}

/* 隐藏旧播放器自带的状态栏 */
:deep(.video-status-bar-simple) {
    display: none !important;
}

.player-sending-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #fff;
    flex: none;
    font-size: 13px;
    height: 46px;
    padding: 0 12px;
}

.player-video-info {
    flex-shrink: 1;
    align-items: center;
    color: var(--text2);
    display: flex;
    height: 16px;
    line-height: 18px;
    margin-right: 24px;
    overflow: hidden;
    position: relative;
    user-select: none;
    white-space: nowrap;
}

.player-video-info-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

/* ArtPlayer .art-controls-center 搬运到状态栏右侧 */
.art-controls-center-slot {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: 4px;
}

/* 覆盖 ArtPlayer 内置的 .art-controls-center 默认样式（display:none） */
.art-controls-center-slot > .art-controls-center {
    display: flex !important;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto !important;
    height: auto !important;
    padding: 0 !important;
    gap: 4px;
    color: #61666D;
    fill: #61666D !important;
}

.art-controls-center-slot > .art-controls-center:hover {
    color: var(--brand_pink);
    fill: var(--brand_pink) !important;
}

.art-controls-center-slot > .art-controls-center:hover .art-icon {
    fill: var(--brand_pink) !important;
}

.art-controls-center-slot > .art-controls-center .art-control {
    min-width: 32px;
    min-height: 32px;
    opacity: 0.85;
}

.art-controls-center-slot > .art-controls-center .art-control:hover {
    opacity: 1;
}

.art-controls-center-slot > .art-controls-center .art-icon {
    width: 22px;
    height: 22px;
    fill: #61666D !important;
}

.art-controls-center-slot > .art-controls-center .art-icon:hover,
.art-controls-center-slot > .art-controls-center .art-control:hover .art-icon {
    fill: var(--brand_pink) !important;
}

.art-controls-center-slot > .art-controls-center svg {
    fill: #61666D !important;
}

.art-controls-center-slot > .art-controls-center svg:hover,
.art-controls-center-slot > .art-controls-center .art-control:hover svg {
    fill: var(--brand_pink) !important;
}

/* artplayer-plugin-danmuku 样式适配（用 :deep 穿透 scoped） */
/* 输入框背景改灰色 */
.art-controls-center-slot :deep(.apd-emitter) {
    background-color: #f1f2f3 !important;
}

.art-controls-center-slot :deep(.apd-input) {
    color: #18191c !important;
    background-color: transparent !important;
}

.art-controls-center-slot :deep(.apd-input::placeholder) {
    color: #9499a0 !important;
}

/* 开关弹幕按钮只显示一个：开状态显示 on，关状态显示 off */
.art-controls-center-slot :deep(.apd-toggle-on) {
    display: block !important;
}

.art-controls-center-slot :deep(.apd-toggle-off) {
    display: none !important;
}

/* 监听 data-danmuku-visible 属性变化（通过 [data-danmuku-visible="false"] 控制） */
.art-controls-center-slot[data-danmuku-visible="false"] :deep(.apd-toggle-on) {
    display: none !important;
}

.art-controls-center-slot[data-danmuku-visible="false"] :deep(.apd-toggle-off) {
    display: block !important;
}

/* danmuku 插件图标颜色 */
.art-controls-center-slot :deep(.apd-icon) {
    fill: #61666D !important;
}

.art-controls-center-slot :deep(.apd-icon:hover) {
    fill: var(--brand_pink) !important;
}

.art-controls-center-slot :deep(.apd-send) {
    background-color: #00AEEC !important;
    color: #fff !important;
}

/* 让整个 danmuku 插件 UI bar 的高度合适 */
.art-controls-center-slot :deep(.artplayer-plugin-danmuku) {
    color: #18191c;
    gap: 8px;
}

.video-parts-list {
    margin-bottom: 12px;
    border-radius: 6px;
    background: var(--bg2);
    overflow: hidden;
}

.parts-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--text1);
    padding: 10px 12px 6px;
}

.parts-items {
    max-height: 240px;
    overflow-y: auto;
}

.part-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    cursor: pointer;
    transition: background 0.15s;
}

.part-item:hover {
    background: var(--bg3);
}

.part-item.active {
    background: var(--brand_pink);
}

.part-item.active .part-index,
.part-item.active .part-title,
.part-item.active .part-duration {
    color: #fff;
}

.part-index {
    font-size: 12px;
    font-weight: 600;
    color: var(--text2);
    min-width: 28px;
}

.part-title {
    flex: 1;
    font-size: 13px;
    color: var(--text1);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.part-duration {
    font-size: 12px;
    color: var(--text2);
    min-width: 40px;
    text-align: right;
}

.recommend-list {
    margin-top: 18px;
}

.rec-title {
    font-size: 15px;
    -webkit-font-smoothing: antialiased;
    color: var(--text1);
    display: flex;
    justify-content: space-between;
    margin-bottom: 12px;
    line-height: 20px;
}

.next-button {
    color: var(--text3);
    font-size: 13px;
    line-height: 16px;
    cursor: pointer;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
}

.next-button .txt {
    margin-right: 4px;
    vertical-align: middle;
}

.next-button .switch-button {
    margin: 0;
    display: inline-block;
    position: relative;
    width: 30px;
    height: 20px;
    outline: none;
    border-radius: 10px;
    box-sizing: border-box;
    cursor: pointer;
    transition: border-color .2s, background-color .2s;
    vertical-align: middle;
    background: var(--text3);
    border: 1px solid var(--text3);
}

.next-button .switch-button.on {
    background: var(--brand_pink);
    border: 1px solid var(--brand_pink);
}

.next-button .switch-button:after {
    content: "";
    position: absolute;
    top: 1px;
    left: 1px;
    border-radius: 100%;
    width: 16px;
    height: 16px;
    background-color: #fff;
    transition: all .2s;
}

.next-button .switch-button.on:after {
    left: 11px;
}

.split-line {
    width: 100%;
    height: 1px;
    background: var(--line_regular);
}

.rec-list {
    margin-top: 18px;
}

.video-page-card-small {
    margin-bottom: 12px;
}

.video-page-card-small a {
    color: #222;
    background-color: transparent;
    text-decoration: none;
    outline: none;
    cursor: pointer;
    transition: color .3s;
    -webkit-text-decoration-skip: objects;
}

.video-page-card-small .card-box {
    display: flex;
}

.video-page-card-small .card-box .pic-box {
    position: relative;
    width: 141px;
    height: 80px;
    border-radius: 6px;
    background: var(--graph_weak);
    flex: 0 0 auto;
}

.video-page-card-small .card-box .pic-box .pic {
    position: relative;
    overflow: hidden;
    border-radius: 6px;
    width: 100%;
    height: 100%;
    cursor: pointer;
}

.video-page-card-small .card-box .pic-box .pic img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    image-rendering: crisp-edges;
}

.video-page-card-small .card-box .pic-box .pic .duration {
    position: absolute;
    bottom: 6px;
    right: 6px;
    color: #fff;
    height: 20px;
    line-height: 20px;
    transition: opacity 0.3s;
    z-index: 5;
    font-size: 13px;
    background-color: rgba(0, 0, 0, 0.4);
    border-radius: 2px;
    padding: 0 4px;
}

.video-page-card-small .card-box .info {
    margin-left: 10px;
    flex: 1;
    font-size: 13px;
    line-height: 15px;
}

.video-page-card-small .card-box .info .title {
    cursor: pointer;
    color: var(--text1);
    display: block;
    font-size: 15px;
    line-height: 19px;
    transition: color 0.3s;
    display: -webkit-box;
    overflow: hidden;
    -webkit-box-orient: vertical;
    text-overflow: -o-ellipsis-lastline;
    text-overflow: ellipsis;
    word-break: break-word;
    -webkit-line-clamp: 2;
    -webkit-font-smoothing: antialiased;
}

.video-page-card-small .card-box .info .title:hover {
    color: var(--brand_pink);
}

.video-page-card-small .card-box .info .upname {
    cursor: pointer;
    margin: 2px 0;
    height: 20px;
    color: var(--text3);
    transition: color 0.3s;
    display: flex;
    align-items: center;
}

.video-page-card-small .card-box .info .upname:hover {
    color: var(--brand_pink);
}

.video-page-card-small .card-box .info .upname svg {
    margin-right: 4px;
    fill: var(--text3);
    transition: fill 0.3s;
}

.video-page-card-small .card-box .info .upname:hover svg {
    fill: var(--brand_pink);
}

.video-page-card-small .card-box .info .upname .name {
    display: -webkit-box;
    overflow: hidden;
    -webkit-box-orient: vertical;
    text-overflow: -o-ellipsis-lastline;
    text-overflow: ellipsis;
    word-break: break-word;
    -webkit-line-clamp: 1;
}

.video-page-card-small .card-box .info .playinfo {
    color: var(--text3);
    fill: var(--text3);
    display: inline-flex;
    align-items: center;
}

.video-page-card-small .card-box .info .playinfo svg {
    margin-right: 4px;
}

.playinfo-dm {
    margin-left: 8px;
}

@media (min-width: 1681px) {
    .video-info-container {
        height: 108px;
    }

    .up-info-container {
        height: 108px;
    }

    .video-info-container .video-title {
        font-size: 22px;
        line-height: 34px;
    }

    .right-container {
        width: 411px;
    }

    .up-name {
        font-size: 16px;
        max-width: calc(100% - 12px - 60px);
    }

    .send-msg {
        font-size: 14px;
    }

    .up-description {
        font-size: 14px;
    }

    .follow-btn {
        width: 230px;
    }
}
</style>