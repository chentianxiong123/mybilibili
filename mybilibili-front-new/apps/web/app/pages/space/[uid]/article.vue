<template>
    <div class="space-article">
        <h3 class="page-head">
            <span class="t">{{ this.$store.state.user.uid === uid ? '我' : 'TA' }}的图文</span>
            <span class="count">{{ articleCount }}</span>
        </h3>
        <div class="article-empty">
            <i class="iconfont icon-meiyoupinleimu"></i>
            <p>图文专栏功能尚未开放</p>
            <p class="sub">当前版本暂不支持图文投稿，敬请期待</p>
        </div>
    </div>
</template>

<script lang="ts">
export default {
    name: "SpaceArticle",
    data() {
        return {
            articleCount: 0,
        }
    },
    computed: {
        uid() {
            return Number(this.$route.params.uid);
        }
    },
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
</style>