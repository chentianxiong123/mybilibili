<template>
  <GridLayout columns="*, *, *, *, *" class="bottom-tab-bar" height="56">
    <GridLayout col="0" @tap="onTap(0)">
      <Label :text="activeTab === 0 ? '🏠' : '🏠'" class="tab-icon" :class="{ active: activeTab === 0 }" row="0" />
      <Label text="首页" class="tab-label" :class="{ active: activeTab === 0 }" row="1" />
    </GridLayout>
    <GridLayout col="1" @tap="onTap(1)">
      <Label text="🔔" class="tab-icon" :class="{ active: activeTab === 1 }" />
      <Label text="关注" class="tab-label" :class="{ active: activeTab === 1 }" />
    </GridLayout>
    <GridLayout col="2" @tap="onTap(2)">
      <Label text="➕" class="publish-icon" />
    </GridLayout>
    <GridLayout col="3" @tap="onTap(3)">
      <Label text="🛒" class="tab-icon" :class="{ active: activeTab === 3 }" />
      <Label text="会员购" class="tab-label" :class="{ active: activeTab === 3 }" />
    </GridLayout>
    <GridLayout col="4" @tap="onTap(4)">
      <Label text="👤" class="tab-icon" :class="{ active: activeTab === 4 }" />
      <Label text="我的" class="tab-label" :class="{ active: activeTab === 4 }" />
    </GridLayout>
  </GridLayout>
</template>

<script lang="ts" setup>
const props = defineProps<{ activeTab: number }>()
const emit = defineEmits<{ select: [index: number] }>()

function onTap(index: number) {
  emit('select', index)
}
</script>

<style scoped>
.bottom-tab-bar {
  background-color: white;
  border-top-width: 1;
  border-top-color: #e3e5e7;
  padding-bottom: 0;
}

.tab-icon {
  font-size: 22;
  text-align: center;
  color: #61666d;
}

.tab-icon.active {
  color: #fb7299;
}

.tab-label {
  font-size: 10;
  text-align: center;
  color: #61666d;
  margin-top: 2;
}

.tab-label.active {
  color: #fb7299;
  font-weight: bold;
}

.publish-icon {
  font-size: 26;
  text-align: center;
  color: white;
  background-color: #fb7299;
  width: 44;
  height: 44;
  border-radius: 22;
  margin-top: -8;
}
</style>