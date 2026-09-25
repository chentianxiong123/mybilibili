<template><div></div></template>

<script lang="ts">
import { hasAuthSession, getCurrentUserId } from "@/utils/auth.ts"
export default {
    name: 'SpaceRedirect',
    async created() {
        if (hasAuthSession()) {
            try {
                // uid 来自可读的 user_info cookie（token 是 HttpOnly，JS 拿不到）
                const uid = getCurrentUserId();
                if (!uid) throw new Error('no uid');
                this.$router.push(`/space/${uid}`);
            } catch (e) {
                console.log('resolve uid exception:', e);
                this.$router.push('/');
            }
        } else {
            this.$router.push('/');
        }
    }
}
</script>