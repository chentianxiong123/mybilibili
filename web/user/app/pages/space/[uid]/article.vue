<template>
    <div class="space-article">
        <h3 class="page-head">
            <span class="t">{{ isSelf ? '我' : 'TA' }}的专栏</span>
            <span class="count">{{ total }}</span>
        </h3>

        <div v-if="loading" class="article-loading">加载中…</div>
        <div v-else-if="!list.length" class="article-empty">
            <i class="iconfont icon-meiyoupinleimu"></i>
            <p>还没有发布专栏</p>
            <p class="sub">写点图文，分享你的想法</p>
        </div>
        <div v-else class="article-list">
            <div v-for="a in list" :key="a.id" class="article-item" @click="open(a.id)">
                <div class="article-item-cover">
                    <img v-if="a.cover_url" :src="a.cover_url" :alt="a.title" />
                    <div v-else class="article-item-cover-text">{{ a.title || '无标题' }}</div>
                </div>
                <div class="article-item-body">
                    <div class="article-item-title">{{ a.title || '无标题' }}</div>
                    <div class="article-item-summary">{{ a.summary || '（暂无摘要）' }}</div>
                    <div class="article-item-meta">
                        <span>{{ fmt(a.publish_time || a.created_at) }}</span>
                        <span class="dot">·</span>
                        <span>{{ fmtNum(a.view_count) }} 阅读</span>
                        <span class="dot">·</span>
                        <span>{{ fmtNum(a.like_count) }} 点赞</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { articleApi } from '@/api/article'

export default {
    name: 'SpaceArticle',
    data() {
        return {
            list: [] as any[],
            total: 0,
            loading: false,
        }
    },
    computed: {
        uid(): number {
            return Number(this.$route.params.uid)
        },
        isSelf(): boolean {
            return this.$store.state.user?.uid === this.uid
        }
    },
    watch: {
        uid() {
            this.load()
        }
    },
    mounted() {
        this.load()
    },
    methods: {
        async load() {
            if (!this.uid) return
            this.loading = true
            try {
                const res: any = await articleApi.getUserArticles(this.uid, 1, 20)
                const data = res?.data
                this.list = data?.list || []
                this.total = this.list.length
            } catch {
                this.list = []
                this.total = 0
            } finally {
                this.loading = false
            }
        },
        open(id: number) {
            this.$router.push(`/read/cv${id}`)
        },
        fmt(t: string) {
            if (!t) return ''
            const d = new Date(t)
            if (isNaN(d.getTime())) return String(t).replace('T', ' ').slice(0, 10)
            const p = (n: number) => String(n).padStart(2, '0')
            return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
        },
        fmtNum(n: number) {
            n = n || 0
            if (n >= 10000) return (n / 10000).toFixed(1) + '万'
            return String(n)
        }
    }
}
</script>

<style scoped>
.space-article {
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

.article-loading {
    text-align: center;
    color: #99a2aa;
    padding: 80px 0;
}

.article-empty {
    text-align: center;
    padding: 80px 0;
    color: #99a2aa;
}

.article-empty .iconfont {
    font-size: 60px;
    color: #d8dde3;
    display: block;
    margin-bottom: 16px;
}

.article-empty p {
    margin: 4px 0;
    color: #61666d;
}

.article-empty .sub {
    color: #99a2aa;
    font-size: 13px;
}

.article-list {
    display: flex;
    flex-direction: column;
}

.article-item {
    display: flex;
    gap: 16px;
    padding: 16px 0;
    border-bottom: 1px solid #f1f2f3;
    cursor: pointer;
}

.article-item:last-child {
    border-bottom: none;
}

.article-item:hover .article-item-title {
    color: #00a1d6;
}

.article-item-cover {
    flex: 0 0 160px;
    height: 100px;
    border-radius: 4px;
    overflow: hidden;
    background: #f4f5f7;
    display: flex;
    align-items: center;
    justify-content: center;
}

.article-item-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.article-item-cover-text {
    color: #99a2aa;
    font-size: 13px;
    padding: 10px;
    text-align: center;
    overflow: hidden;
}

.article-item-body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
}

.article-item-title {
    font-size: 16px;
    font-weight: 500;
    color: #222;
    margin-bottom: 8px;
    transition: color .2s;
}

.article-item-summary {
    font-size: 13px;
    color: #61666d;
    line-height: 1.6;
    margin-bottom: auto;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.article-item-meta {
    font-size: 12px;
    color: #99a2aa;
    margin-top: 10px;
    display: flex;
    gap: 6px;
}
</style>
