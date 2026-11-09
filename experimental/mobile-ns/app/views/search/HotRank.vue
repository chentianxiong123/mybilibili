<template>
  <Page class="page">
    <ActionBar title="热门榜单">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="rankList.length > 0">
          <StackLayout v-for="(item, i) in rankList" :key="i" class="rank-item" @tap="openVideo(item)">
            <GridLayout columns="auto, *, auto">
              <Label :text="String(i + 1)" class="rank-number" :class="{ 'rank-top': i < 3 }" col="0" />
              <StackLayout col="1" class="rank-info">
                <Label :text="item.title || ''" class="rank-title" textWrap="true" />
                <Label :text="'▶ ' + formatCount(item.play || 0)" class="rank-stats" />
              </StackLayout>
              <Image :src="item.pic || ''" class="rank-cover" col="2" />
            </GridLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无榜单数据" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { getHotRank } from '../../api/search'
import { formatCount } from '../../utils/format'

const rankList = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getHotRank()
    if (res.code === '1') {
      rankList.value = res.data || []
    }
  } catch (e) {
    console.error('加载榜单失败:', e)
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
.rank-item {
  padding: 12 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.rank-number {
  font-size: 18;
  font-weight: bold;
  color: #9499a0;
  width: 32;
  text-align: center;
  margin-right: 12;
}

.rank-number.rank-top {
  color: #fb7299;
}

.rank-info {
  margin-right: 12;
}

.rank-title {
  font-size: 14;
  color: #18191c;
  font-weight: 500;
  line-height: 1.3;
}

.rank-stats {
  font-size: 12;
  color: #9499a0;
  margin-top: 6;
}

.rank-cover {
  width: 80;
  height: 56;
  border-radius: 4;
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