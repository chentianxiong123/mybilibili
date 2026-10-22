<template>
    <div class="message-empty">
        <img :src="emptyImg" alt="" />
    </div>
</template>

<script lang="ts">
import emptyImg from '@/assets/site/img/message/empty.png';

export default {
    name: 'MessageEmpty',
    data() {
        return { emptyImg };
    },
};
</script>
