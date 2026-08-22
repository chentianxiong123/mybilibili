<template>
  <Page class="page">
    <ActionBar title="创作中心">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout class="stats-section">
          <Label text="数据概览" class="section-title" />
          <GridLayout columns="*, *, *" class="stats-grid">
            <StackLayout col="0" class="stat-card">
              <Label :text="formatCount(stats.playCount || 0)" class="stat-value" />
              <Label text="播放量" class="stat-label" />
            </StackLayout>
            <StackLayout col="1" class="stat-card">
              <Label :text="formatCount(stats.fans || 0)" class="stat-value" />
              <Label text="粉丝数" class="stat-label" />
            </StackLayout>
            <StackLayout col="2" class="stat-card">
              <Label :text="formatCount(stats.likes || 0)" class="stat-value" />
              <Label text="获赞数" class="stat-label" />
            </StackLayout>
          </GridLayout>
        </StackLayout>

        <StackLayout class="menu-section">
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onManuscriptTap">
            <Label text="🎬" class="menu-icon" col="0" />
            <Label text="稿件管理" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto">
            <Label text="📊" class="menu-icon" col="0" />
            <Label text="数据分析" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto" @tap="onCommentTap">
            <Label text="💬" class="menu-icon" col="0" />
            <Label text="评论管理" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
          <GridLayout class="menu-item" columns="auto, *, auto">
            <Label text="💰" class="menu-icon" col="0" />
            <Label text="收益中心" col="1" class="menu-label" />
            <Label text="›" class="menu-arrow" col="2" />
          </GridLayout>
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $navigateTo } from 'nativescript-vue'
import { formatCount } from '../../utils/format'

const stats = ref<any>({})

onMounted(async () => {
  try {
    const user = require('../../utils/storage').default.getUser()
    if (user) {
      stats.value = {
        playCount: user.playCount || user.play || 0,
        fans: user.followerCount || user.fans || 0,
        likes: user.likeCount || user.likes || 0
      }
    }
  } catch (e) {
    console.error('加载创作中心数据失败:', e)
  }
})

function onBack() {
  $navigateBack()
}

function onManuscriptTap() {
  $navigateTo(require('../space/ManuscriptManage.vue').default)
}

function onCommentTap() {
}
</script>

<style scoped>
.stats-section {
  padding: 16;
  background-color: white;
  margin-bottom: 8;
}

.section-title {
  font-size: 16;
  font-weight: bold;
  color: #18191c;
  margin-bottom: 16;
}

.stats-grid {
  gap: 8;
}

.stat-card {
  background-color: #f8f9fa;
  padding: 16 8;
  border-radius: 8;
  align-items: center;
}

.stat-value {
  font-size: 22;
  font-weight: bold;
  color: #fb7299;
}

.stat-label {
  font-size: 12;
  color: #9499a0;
  margin-top: 4;
}

.menu-section {
  background-color: white;
}

.menu-item {
  padding: 14 16;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.menu-icon {
  font-size: 18;
  margin-right: 12;
  width: 28;
}

.menu-label {
  font-size: 14;
  color: #18191c;
}

.menu-arrow {
  font-size: 18;
  color: #9499a0;
}
</style>