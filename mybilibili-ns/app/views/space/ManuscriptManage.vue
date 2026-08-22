<template>
  <Page class="page">
    <ActionBar title="稿件管理" />
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="list.length > 0">
          <GridLayout v-for="(item, i) in list" :key="i" class="ms-item" columns="auto, *" @tap="openVideo(item.aId)">
            <Image :src="item.pic || ''" class="ms-cover" col="0" />
            <StackLayout col="1" class="ms-info">
              <Label :text="item.title || ''" class="ms-title" textWrap="true" />
              <Label :text="'播放: ' + formatCount(item.play)" class="ms-stats" />
            </StackLayout>
          </GridLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无稿件" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $showModal } from 'nativescript-vue'
import { getManuscripts } from '../../api/manuscript'
import { formatCount } from '../../utils/format'

const list = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getManuscripts()
    if (res.code === '1') {
      list.value = res.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

function openVideo(aId: number) {
  $showModal(require('../video/Detail.vue').default, {
    fullscreen: true,
    props: { aId }
  })
}
</script>

<style scoped>
.ms-item {
  padding: 10 16;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.ms-cover {
  width: 120;
  height: 72;
  border-radius: 6;
  margin-right: 12;
}

.ms-title {
  font-size: 13;
  color: #18191c;
  line-height: 1.3;
}

.ms-stats {
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