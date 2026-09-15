<template>
  <Page class="page">
    <ActionBar title="历史记录" />
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="history.length > 0">
          <GridLayout v-for="(item, i) in history" :key="i" class="history-item" columns="auto, *" @tap="openVideo(item.aId)">
            <Image :src="item.pic || ''" class="history-cover" col="0" />
            <StackLayout col="1" class="history-info">
              <Label :text="item.title || ''" class="history-title" textWrap="true" />
              <Label :text="formatTimeLabel(item.viewAt)" class="history-time" />
            </StackLayout>
          </GridLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无历史记录" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $showModal } from 'nativescript-vue'
import storage from '../../utils/storage'
import { formatTimeLabel } from '../../utils/format'

const history = ref<any[]>([])
const loading = ref(true)

onMounted(() => {
  history.value = storage.getViewHistory()
  loading.value = false
})

function openVideo(aId: number) {
  $showModal(require('../video/Detail.vue').default, {
    fullscreen: true,
    props: { aId }
  })
}
</script>

<style scoped>
.history-item {
  padding: 10 16;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.history-cover {
  width: 120;
  height: 72;
  border-radius: 6;
  margin-right: 12;
}

.history-info {
  margin-left: 4;
}

.history-title {
  font-size: 13;
  color: #18191c;
  line-height: 1.3;
}

.history-time {
  font-size: 11;
  color: #9499a0;
  margin-top: 6;
}

.empty-state {
  padding: 80 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>