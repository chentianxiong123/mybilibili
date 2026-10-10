<template>
    <div class="placeholder">
        <div class="placeholder-img is-im"></div>
        <div class="tip">快找小伙伴聊天吧 ( ゜- ゜)つロ</div>
    </div>
</template>

<script lang="ts">
export default {
    name: "MessageWhisperIndex",
    mounted() {
        // 父组件 whisper.vue 拉完会话列表后自动打开第一个；
        // 列表为空（新号/未登录）则保持占位
        const first = this.$store.state.chatList[0];
        if (first) {
            this.$router.replace(`/message/whisper/${first.user.uid}`);
        }
    },
    watch: {
        // 父组件异步拉取完成后补一次跳转
        "$store.state.chatList.length"(n) {
            if (n > 0 && this.$route.path === '/message/whisper') {
                const first = this.$store.state.chatList[0];
                this.$router.replace(`/message/whisper/${first.user.uid}`);
            }
        }
    }
}
</script>

<style scoped>
.placeholder {
    width: 100%;
    height: 100%;
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-pack: center;
    -ms-flex-pack: center;
    justify-content: center;
    -webkit-box-align: center;
    -ms-flex-align: center;
    align-items: center;
    -webkit-box-orient: vertical;
    -webkit-box-direction: normal;
    -ms-flex-direction: column;
    flex-direction: column;
    overflow: hidden;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
}

.placeholder .placeholder-img {
    width: 402px;
    height: 304px;
}

.placeholder .is-im {
    background-image: url('@/assets/teriteri/img/bilibili/no_message.png');
    background-size: 402px 204px;
    margin-bottom: 32px;
    height: 204px;
}

.placeholder .tip {
    color: #8896b8;
    font-size: 14px;
    line-height: 1.5em;
}
</style>
