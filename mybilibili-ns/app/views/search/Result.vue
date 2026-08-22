<template>
  <Page class="page">
    <ActionBar title="搜索结果">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="results.length > 0">
          <StackLayout v-for="(item, i) in results" :key="i" class="result-item" @tap="openVideo(item)">
            <GridLayout columns="132, *">
              <Image :src="item.pic || ''" class="result-cover" col="0" />
              <StackLayout col="1" class="result-info">
                <Label :text="item.title || ''" class="result-title" textWrap="true" />
                <Label :text="item.author || ''" class="result-author" />
                <GridLayout columns="auto, auto" class="result-stats-row">
                  <Label :text="'▶ ' + formatCount(item.play || 0)" col="0" class="result-stats" />
                  <Label :text="'💬 ' + (item.videoReview || 0)" col="1" class="result-stats" />
                </GridLayout>
              </StackLayout>
            </GridLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="未找到相关结果" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { search } from '../../api/search'
import { formatCount } from '../../utils/format'

const props = defineProps<{ keyword: string }>()

const results = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await search(props.keyword)
    if (res.code === '1') {
      results.value = res.data || []
    }
  } catch (e) {
    console.error('搜索失败:', e)
  } finally {
    loading.value = false
  }
})

function onBack() {
  $navigateBack()
}

function openVideo(item: any) {
  const aId = item.aId || item.id || item.aid
  if (aId) {
    $showModal(require('../../views/video/Detail.vue').default, {
      fullscreen: true,
      props: { aId }
    })
  }
}
</script>

<style scoped>
.result-item {
  padding: 10 12;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.result-cover {
  width: 120;
  height: 72;
  border-radius: 4;
}

.result-info {
  padding-left: 10;
}

.result-title {
  font-size: 14;
  color: #18191c;
  font-weight: 500;
  line-height: 1.3;
}

.result-author {
  font-size: 12;
  color: #9499a0;
  margin-top: 6;
}

.result-stats-row {
  margin-top: 4;
}

.result-stats {
  font-size: 11;
  color: #9499a0;
  margin-right: 12;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}

.empty-state {
  padding: 60 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>