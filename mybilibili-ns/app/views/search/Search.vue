<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *">
      <GridLayout row="0" columns="auto, *, auto" class="search-header">
        <Label text="◀" col="0" class="back-btn" @tap="onBack" />
        <SearchBar col="1" v-model="keyword" hint="搜索" @submit="onSearch" @textChange="onTextChange" class="search-bar" />
        <Label col="2" text="搜索" class="search-action" @tap="onSearch" />
      </GridLayout>

      <ScrollView row="1">
        <StackLayout>
          <StackLayout v-if="!searched">
            <Label text="热门搜索" class="section-title" />
            <StackLayout v-if="hotwordsLoading" class="loading">
              <Label text="加载中..." />
            </StackLayout>
            <WrapLayout v-else class="hotwords-wrap">
              <Label v-for="(hw, i) in hotwords" :key="i" :text="hw.keyword || hw" class="hotword-item" @tap="onHotwordTap(hw.keyword || hw)" />
            </WrapLayout>
          </StackLayout>

          <StackLayout v-else>
            <StackLayout v-if="resultsLoading" class="loading">
              <Label text="加载中..." />
            </StackLayout>
            <StackLayout v-else-if="results.length > 0">
              <StackLayout v-for="(item, i) in results" :key="i" class="result-item" @tap="openVideo(item)">
                <GridLayout columns="132, *">
                  <Image :src="item.pic || ''" class="result-cover" col="0" />
                  <StackLayout col="1" class="result-info">
                    <Label :text="item.title || ''" class="result-title" textWrap="true" />
                    <Label :text="item.author || ''" class="result-author" />
                    <Label :text="'▶ ' + formatCount(item.play || 0)" class="result-stats" />
                  </StackLayout>
                </GridLayout>
              </StackLayout>
            </StackLayout>
            <StackLayout v-else class="empty-state">
              <Label text="未找到相关结果" />
            </StackLayout>
          </StackLayout>
        </StackLayout>
      </ScrollView>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { search, getHotwords } from '../../api/search'
import { formatCount } from '../../utils/format'

const keyword = ref('')
const hotwords = ref<any[]>([])
const hotwordsLoading = ref(true)
const results = ref<any[]>([])
const resultsLoading = ref(false)
const searched = ref(false)

onMounted(async () => {
  try {
    const res = await getHotwords()
    if (res.code === '1' && res.data) {
      hotwords.value = res.data
    }
  } catch (e) {
    console.error('加载热词失败:', e)
  } finally {
    hotwordsLoading.value = false
  }
})

function onBack() {
  $navigateBack()
}

function onTextChange(args: any) {
  if (!keyword.value && searched.value) {
    searched.value = false
    results.value = []
  }
}

async function onSearch() {
  const kw = keyword.value.trim()
  if (!kw) return
  searched.value = true
  resultsLoading.value = true
  try {
    const res = await search(kw)
    if (res.code === '1') {
      results.value = res.data || []
    }
  } catch (e) {
    console.error('搜索失败:', e)
  } finally {
    resultsLoading.value = false
  }
}

function onHotwordTap(hw: string) {
  keyword.value = hw
  onSearch()
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
.search-header {
  padding: 8 12;
  background-color: white;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.back-btn {
  font-size: 20;
  color: #18191c;
  padding: 8 0;
  margin-right: 8;
}

.search-bar {
  background-color: #f4f5f7;
  border-radius: 18;
  height: 36;
}

.search-action {
  font-size: 14;
  color: #fb7299;
  padding: 8 0 8 12;
  font-weight: bold;
}

.section-title {
  font-size: 16;
  font-weight: bold;
  color: #18191c;
  padding: 16 16 8;
}

.hotwords-wrap {
  padding: 8 12;
}

.hotword-item {
  font-size: 13;
  color: #61666d;
  background-color: #f4f5f7;
  padding: 6 14;
  border-radius: 16;
  margin: 4 6;
}

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

.result-stats {
  font-size: 11;
  color: #9499a0;
  margin-top: 4;
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