<template>
    <div class="carousel-index">
        <div class="carousel-container" :style="`background-color: ${color};`">
            <div
                class="carousel-transform"
                :style="isRoll == 2 ? `transform: translateX(-${((100 / carousels.length) * 2).toFixed(4)}%); transition: transform 400ms ease 0s;width: ${carousels.length}00%;`
                        : isRoll == 1 ? `transform: translateX(0); transition: transform 400ms ease 0s;width: ${carousels.length}00%;`
                        : `transform: translateX(-${(100 / carousels.length).toFixed(4)}%); transition: transform 0ms ease 0s;width: ${carousels.length}00%;`"
                >
                <div class="carousel-slide" :style="`width: ${(100 / carousels.length).toFixed(4)}%`" v-for="(item, index) in carousels" :key="index">
                    <a class="carousel-inner" :href="item.target" target="_blank">
                        <img :src="item.url" alt="">
                    </a>
                </div>
                <div class="shadow" :style="`background: linear-gradient(to top, rgba(0,0,0,0.95) 0%, rgba(0,0,0,0.7) 50%, transparent 100%);`"></div>
            </div>
            <div class="carousel-footer__left">
                <div class="title"><span>{{ title }}</span></div>
                <div class="page">
                    <div
                        class="point"
                        :class="index == current ? 'is-active' : ''"
                        v-for="index in carousels.length"
                        :key="index"
                        @click="changePoint(index)"
                    ></div>
                </div>
            </div>
            <div class="carousel-footer__right">
                <button type="button" @click="preSlide()">
                    <i class="iconfont icon-zuojiantou"></i>
                </button>
                <button type="button" @click="nextSlide()">
                    <i class="iconfont icon-youjiantou"></i>
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
    let timer;  // 定时器

    // 后端返回的轮播图数据
    interface BannerItem {
        id: number
        title: string
        imageUrl: string
        linkUrl: string
        sortOrder: number
        color?: string
    }

    // 前端轮播图数据
    interface CarouselItem {
        url: string
        title: string
        color: string
        target: string
    }

    // 预设底色（与原 carousel.json 风格一致）
    const PRESET_COLORS = ['#FF6699', '#5894d4', '#836e61', '#728cb4', '#564e3e', '#724b50'];

    export default {
        name: "CarouselIndex",
        data() {
            return {
                // 轮播图列表
                carousels: [] as CarouselItem[],
                // 是否滚动，0 不滚，1上一张，2 下一张
                isRoll: 0,
                // 当前位置
                current: 2,
                // 底色
                color: "",
                // 标题
                title: "",
            }
        },
        methods: {
            // 从后端 API 获取轮播图
            async getCarousels() {
                try {
                    const res = await this.$get('/banner-images/home')
                    const list: BannerItem[] = (res?.data?.data || res?.data || []) as BannerItem[]
                    if (list.length > 0) {
                        this.carousels = list
                            .sort((a, b) => a.sortOrder - b.sortOrder)
                            .map((item, index) => ({
                                url: item.imageUrl,
                                title: item.title,
                                color: item.color || PRESET_COLORS[index % PRESET_COLORS.length],
                                // 后端 /manuscript/11 → 前端 /video/11
                                target: item.linkUrl.replace('/manuscript/', '/video/'),
                            }))
                    } else {
                        this.loadFallback()
                    }
                } catch {
                    this.loadFallback()
                }
                this.$store.commit("updateCarousels", this.carousels.slice());
            },
            // API 失败时回退到静态 JSON
            loadFallback() {
                try {
                    const fallback = require('@/assets/teriteri/json/carousel.json')
                    this.carousels = fallback
                } catch {
                    this.carousels = []
                }
            },

            async refreshTimer() {
                clearTimeout(timer);
                // console.log(timer, "秒过去了");
                await this.nextSlide();
                this.startTimer();
            },
            startTimer() {
                timer = setTimeout(this.nextSlide, 3100);   // 加上主函数的延时时长总共3.5秒换一张图
            },

            nextSlide() {
                clearTimeout(timer);
                // 开滚
                this.isRoll = 2;
                if (this.current < this.carousels.length) {
                    this.current ++;
                } else {
                    this.current = 1;
                }
                // 换成第三张图的底色
                this.color = this.carousels[2].color;
                this.title = this.carousels[2].title;
                setTimeout(() => {
                    // 将第一张图移到数组末尾
                    this.carousels.push(this.carousels.shift());
                    this.isRoll = 0;
                    // 换回位置2的底色
                    this.color = this.carousels[1].color;
                    this.title = this.carousels[1].title;
                }, 500);    // 这里的延时时长要大于等于动画的过渡时长
                this.startTimer();
            },

            preSlide() {
                clearTimeout(timer);
                this.isRoll = 1;
                if (this.current > 1) {
                    this.current --;
                } else {
                    this.current = this.carousels.length;
                }
                // 换成第一张图的底色
                this.color = this.carousels[0].color;
                this.title = this.carousels[0].title;
                setTimeout(() => {
                    // 将最后一张图移到数组开头
                    this.carousels.unshift(this.carousels.pop());
                    this.isRoll = 0;
                    // 换回位置2的底色
                    this.color = this.carousels[1].color;
                    this.title = this.carousels[1].title;
                }, 400);    // 这里的延时时长要大于等于动画的过渡时长
                this.startTimer();
            },

            changePoint(index) {
                // 暂停自动滚动
                clearTimeout(timer);
                this.current = index;
                // 创建一个原数组的副本
                let list = this.$store.state.carousels.slice();
                if (index == 1) {
                    // for (let i = 0; i < this.carousels.length - 1; i ++) {
                    //     list.push(list.shift());
                    //     // console.log("次循环");
                    // }
                    list.unshift(list.pop());
                    this.carousels = list;
                    // console.log("当前轮播图: ", this.carousels);
                    this.color = this.carousels[1].color;
                    this.title = this.carousels[1].title;
                } else {
                    for (let i = 0; i < index - 2; i ++) {
                        list.push(list.shift());
                        // console.log("次循环");
                    }
                    this.carousels = list;
                    // console.log("当前轮播图: ", this.carousels);
                    this.color = this.carousels[1].color;
                    this.title = this.carousels[1].title;
                }
                // 重置自动滚动
                this.startTimer();
            },
        },
        async created() {
            await this.getCarousels();
            if (this.carousels.length > 0) {
                this.color = this.carousels[1]?.color || PRESET_COLORS[0];
                this.title = this.carousels[1]?.title || '';
            }
        },
        mounted() {
            if (this.carousels.length > 0) {
                this.startTimer();
            }
        },
        beforeUnmount() {
            clearTimeout(timer);
        }
    }
</script>

<style scoped>
.carousel-index {
    height: 100%;
}

.carousel-container {
    overflow: hidden;
    height: 100%;
    width: 100%;
}

.carousel-transform {
    overflow: hidden;
    display: flex;
    align-items: center;
    height: 100%;
}

.carousel-slide {
    height: 100%;
    overflow: hidden;
    flex-shrink: 0;
}

.carousel-inner {
    height: 100%;
    width: 100%;
    position: relative;
    display: inline-block;
    line-height: 1;
    vertical-align: middle;
    background-color: #000;
    cursor: pointer;
}

.carousel-inner img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
}

.shadow {
    position: absolute;
    width: 100%;
    height: 30%;
    bottom: 0;
    pointer-events: none;   /* 禁止蒙版元素捕获鼠标事件 */
}

.carousel-footer__left {
    position: absolute;
    bottom: 20px;
    left: 15px;
}

.title span{
    display: block;
    font-weight: 400;
    line-height: 25px;
    color: #fff;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 10px;
}

@media (max-width: 1700.9px) {
    .title span{
        font-size: 18px;
    }
}

@media (min-width: 1701px) {
    .title span{
        font-size: 20px;
    }
}

.page {
    display: flex;
    align-items: center;
    margin-left: -1.5px;
    font-size: 0;
    transform: translateZ(0);
}

.point {
    position: relative;
    display: inline-block;
    width: 8px;
    height: 8px;
    margin: 4px;
    border-radius: 50%;
    background-color: rgba(255,255,255,.4);
    overflow: hidden;
    cursor: pointer;
}

.is-active {
    width: 14px;
    height: 14px;
    margin: 1px;
    background-color: #fff;
}

.carousel-footer__right {
    position: absolute;
    bottom: 40px;
    right: 15px;
}

.carousel-footer__right button {
    display: -webkit-flex;
    display: flex;
    align-items: center;
    justify-content: center;
    display: -webkit-inline-flex;
    display: inline-flex;
    width: 28px;
    height: 28px;
    border-radius: 8px;
    margin-right: 12px;
    background-color: rgba(255,255,255,.1);
    cursor: pointer;
}

.carousel-footer__right button:hover {
    background-color: rgba(255,255,255,.2);
}

.carousel-footer__right .iconfont {
    color: #fff;
}
</style>