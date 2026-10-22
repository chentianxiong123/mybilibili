<template>
    <div class="message-loading">
        <img :src="loadingImg" alt="" />
    </div>
</template>

<script lang="ts">
import loadingImg from '@/assets/site/img/message/loading_tv.gif';

export default {
    name: 'MessageLoading',
    data() {
        return { loadingImg };
    },
};
</script>
