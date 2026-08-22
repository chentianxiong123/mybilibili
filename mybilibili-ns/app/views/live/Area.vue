<template>
  <Page class="page">
    <ActionBar :title="areaName || '直播分类'">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else-if="areas.length > 0">
          <StackLayout v-for="(area, i) in areas" :key="i" class="area-item" @tap="onAreaTap(area)">
            <GridLayout columns="auto, *, auto">
              <Image :src="area.icon || ''" class="area-icon" col="0" />
              <StackLayout col="1" class="area-info">
                <Label :text="area.name || ''" class="area-name" />
                <Label v-if="area.description" :text="area.description" class="area-desc" textWrap="true" />
              </StackLayout>
              <Label text="›" col="2" class="area-arrow" />
            </GridLayout>
          </StackLayout>
        </StackLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无分类数据" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $navigateTo } from 'nativescript-vue'
import { getLiveAreas } from '../../api/live'

const props = defineProps<{ areaId?: number; areaName?: string }>()

const areas = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getLiveAreas()
    if (res.code === '1') {
      areas.value = res.data || []
    }
  } catch (e) {
    console.error('加载直播分类失败:', e)
  } finally {
    loading.value = false
  }
})

function onBack() {
  $navigateBack()
}

function onAreaTap(area: any) {
  $navigateTo(require('./List.vue').default, {
    props: { category: area.id || area.name }
  })
}
</script>

<style scoped>
.area-item {
  padding: 14 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.area-icon {
  width: 40;
  height: 40;
  border-radius: 20;
  margin-right: 12;
}

.area-info {
  margin-left: 4;
}

.area-name {
  font-size: 14;
  font-weight: 500;
  color: #18191c;
}

.area-desc {
  font-size: 12;
  color: #9499a0;
  margin-top: 2;
}

.area-arrow {
  font-size: 18;
  color: #9499a0;
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