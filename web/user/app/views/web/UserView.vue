<script setup lang="ts">
// /user/[id] 与 /profile/[id] 指向同一个空间页（space.bilibili.com/{uid}）。
// 直接复用已像素对齐的 UserProfileView，保证两条路由 UI 完全一致，
// 避免维护两套 divergent 的用户页（VideoCard / 评论区头像等入口链接到 /user/[id]）。
// UserProfileView 通过 route.params.id 取用户 ID，对 /user/[id] 同样生效。
import UserProfileView from '~/views/web/UserProfileView.vue'
</script>

<template>
  <UserProfileView />
</template>
