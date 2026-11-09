<template>
  <GridLayout :col="col" class="video-card" @tap="onTap">
    <StackLayout>
      <Image :src="video?.pic || ''" class="video-cover" />
      <Label :text="video?.title || ''" class="video-title" textWrap="true" />
      <GridLayout columns="*, auto" class="video-meta">
        <Label :text="formatCount(video?.play)" class="plays" col="0" />
        <Label :text="formatCount(video?.videoReview)" class="reviews" col="1" />
      </GridLayout>
    </StackLayout>
  </GridLayout>
</template>

<script lang="ts" setup>
import { $navigateTo, $showModal } from 'nativescript-vue'
import { formatCount } from '../utils/format'

const props = defineProps<{
  video: any
  col: number
}>()

function onTap() {
  $showModal(require('../views/video/Detail.vue').default, {
    fullscreen: true,
    props: { aId: props.video?.aId }
  })
}
</script>

<style scoped>
.video-card {
  padding: 2;
  margin-bottom: 4;
}

.video-cover {
  width: 100%;
  height: 100;
  border-radius: 4;
}

.video-title {
  font-size: 12;
  color: #18191c;
  padding: 4 4 0;
  line-height: 1.3;
}

.video-meta {
  padding: 2 4;
  font-size: 10;
  color: #9499a0;
}

.plays {
  font-size: 10;
  color: #9499a0;
}

.reviews {
  font-size: 10;
  color: #9499a0;
  text-align: right;
}
</style>