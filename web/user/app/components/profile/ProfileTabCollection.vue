<script setup lang="ts">
// 空间页「合集和列表」tab：直接复用已对齐 B站 space.bilibili.com/{uid}/lists 结构的 CollectionListView
// （.space-lists 1640 居中 + .lists-header + 每合集一个 .video-list 分组 + 6 列 .bili-video-card 网格）。
// 原 ProfileTabCollection 为自定义横向滚动布局（.collections-horizontal），与 B站 分组网格模型不同构，
// 改为薄封装渲染 CollectionListView，自动获得与 /collections 路由完全一致的像素对齐布局。
import CollectionListView from '~/views/web/CollectionListView.vue'

defineProps({
  userId: {
    type: [String, Number],
    default: null
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Object,
    default: () => ({})
  }
})
</script>

<template>
  <CollectionListView :user-id="userId" />
</template>
