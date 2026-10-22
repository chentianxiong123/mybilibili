<template>
    <div class="message-divider">
        <span class="message-divider__inner">{{ text }}</span>
    </div>
</template>

<script lang="ts">
export default {
    name: 'MessageDivider',
    props: {
        text: { type: String, default: '上次看到这里' },
    },
};
</script>
