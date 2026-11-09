<template>
  <Page class="page">
    <ActionBar title="关注动态" />
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="feed.length > 0">
          <StackLayout v-for="(item, i) in feed" :key="i" class="feed-card">
            <GridLayout columns="auto, *" class="feed-header">
              <Image :src="item.user?.face || ''" class="feed-avatar" col="0" />
              <StackLayout col="1" class="feed-user">
                <Label :text="item.user?.name || ''" class="feed-username" />
                <Label :text="item.type === 'video' ? '发布了新视频' : '发布了新动态'" class="feed-action" />
              </StackLayout>
            </GridLayout>
            <StackLayout v-if="item.video" class="feed-video" @tap="openVideo(item.video.aId)">
              <Image :src="item.video.pic || ''" class="feed-video-cover" />
              <Label :text="item.video.title || ''" class="feed-video-title" textWrap="true" />
            </StackLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无关注动态" class="empty-text" />
          <Label text="去首页发现更多精彩内容" class="empty-hint" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $showModal } from 'nativescript-vue'
import { getDynamicFeed } from '../../api/dynamic'

const feed = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getDynamicFeed()
    if (res.code === '1') {
      feed.value = res.data || []
    }
  } catch (e) {
    console.error('加载动态失败:', e)
  } finally {
    loading.value = false
  }
})

function openVideo(aId: number) {
  $showModal(require('../../views/video/Detail.vue').default, {
    fullscreen: true,
    props: { aId }
  })
}
</script>

<style scoped>
.feed-card {
  background-color: white;
  margin-bottom: 8;
  padding: 12;
}

.feed-header {
  margin-bottom: 8;
}

.feed-avatar {
  width: 36;
  height: 36;
  border-radius: 18;
  margin-right: 10;
}

.feed-username {
  font-size: 14;
  font-weight: bold;
  color: #18191c;
}

.feed-action {
  font-size: 12;
  color: #9499a0;
  margin-top: 2;
}

.feed-video-cover {
  width: 100%;
  height: 160;
  border-radius: 6;
}

.feed-video-title {
  font-size: 13;
  color: #18191c;
  margin-top: 8;
  line-height: 1.3;
}

.empty-state {
  padding: 80 20;
  align-items: center;
}

.empty-text {
  font-size: 16;
  color: #61666d;
  margin-bottom: 8;
}

.empty-hint {
  font-size: 13;
  color: #9499a0;
}
</style>