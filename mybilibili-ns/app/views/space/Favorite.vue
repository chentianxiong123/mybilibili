<template>
  <Page class="page">
    <ActionBar title="我的收藏" />
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="list.length > 0">
          <GridLayout v-for="(item, i) in list" :key="i" class="fav-item" columns="auto, *" @tap="openVideo(item.aId)">
            <Image :src="item.pic || ''" class="fav-cover" col="0" />
            <StackLayout col="1" class="fav-info">
              <Label :text="item.title || ''" class="fav-title" textWrap="true" />
              <Label :text="item.author || ''" class="fav-author" />
            </StackLayout>
          </GridLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无收藏" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $showModal } from 'nativescript-vue'
import { getFavorites } from '../../api/favorite'

const list = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getFavorites()
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
.fav-item {
  padding: 10 16;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.fav-cover {
  width: 120;
  height: 72;
  border-radius: 6;
  margin-right: 12;
}

.fav-title {
  font-size: 13;
  color: #18191c;
  line-height: 1.3;
}

.fav-author {
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