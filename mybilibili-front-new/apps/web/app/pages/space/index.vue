<template><div></div></template>

<script>
export default {
    name: 'SpaceRedirect',
    async created() {
        if (localStorage.getItem('teri_token')) {
            try {
                // 解析JWT获取当前用户uid
                const w = localStorage.getItem('teri_token').split('.')[1];
                const info = JSON.parse(atob(w));
                this.$router.push(`/space/${info.sub}`);
            } catch (e) {
                console.log('atob exception:', e);
                this.$router.push('/');
            }
        } else {
            this.$router.push('/');
        }
    }
}
</script>