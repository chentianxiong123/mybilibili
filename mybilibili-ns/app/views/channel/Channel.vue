<template>
  <Page class="page" actionBarHidden="true">
    <GridLayout rows="auto, *">
      <GridLayout row="0" columns="auto, *" class="channel-header">
        <Label text="◀" col="0" class="back-btn" @tap="onBack" />
        <Label text="分区" col="1" class="channel-title" />
      </GridLayout>
      <ScrollView row="1">
        <StackLayout>
          <StackLayout v-if="partitions.length > 0">
            <StackLayout v-for="(p, i) in partitions" :key="i" class="partition-item" @tap="onPartitionTap(p)">
              <GridLayout columns="auto, *, auto">
                <Image :src="p.icon || ''" class="partition-icon" col="0" />
                <StackLayout col="1" class="partition-info">
                  <Label :text="p.name || ''" class="partition-name" />
                  <Label v-if="p.description" :text="p.description" class="partition-desc" textWrap="true" />
                </StackLayout>
                <Label text="›" col="2" class="partition-arrow" />
              </GridLayout>
            </StackLayout>
          </StackLayout>
          <StackLayout v-else class="empty-state">
            <Label text="暂无分区数据" />
          </StackLayout>
        </StackLayout>
      </ScrollView>
    </GridLayout>
  </Page>
</template>

<script lang="ts" setup>
import { ref } from 'nativescript-vue'
import { $navigateBack, $navigateTo } from 'nativescript-vue'

const props = defineProps<{ partitions: any[] }>()

const partitions = ref(props.partitions || [])

function onBack() {
  $navigateBack()
}

function onPartitionTap(p: any) {
  const rId = p.id || p.rId
  if (rId) {
    $navigateTo(require('../ranking/Ranking.vue').default, {
      props: { rId }
    })
  }
}
</script>

<style scoped>
.channel-header {
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

.channel-title {
  font-size: 16;
  font-weight: bold;
  color: #18191c;
}

.partition-item {
  padding: 14 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
  background-color: white;
}

.partition-icon {
  width: 40;
  height: 40;
  border-radius: 8;
  margin-right: 12;
}

.partition-info {
  margin-left: 4;
}

.partition-name {
  font-size: 15;
  font-weight: 500;
  color: #18191c;
}

.partition-desc {
  font-size: 12;
  color: #9499a0;
  margin-top: 2;
}

.partition-arrow {
  font-size: 18;
  color: #9499a0;
}

.empty-state {
  padding: 60 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>