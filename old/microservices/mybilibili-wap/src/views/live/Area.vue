<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Header from '../../components/Header.vue'
import { getAreas } from '../../api/live'

const router = useRouter()
const areas = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getAreas()
    if (res.code === '1') {
      areas.value = (res.data || []).filter(a => a.id !== 99)
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="live-area-page">
    <Header />
    <div class="area-title">全部分类</div>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else class="area-grid">
      <router-link
        v-for="area in areas"
        :key="area.id"
        :to="`/m/live/list?parent_area_id=${area.parentId || area.id}&areaId=${area.id}`"
        class="area-item"
      >
        <img v-if="area.entranceIcon?.src" :src="area.entranceIcon.src" class="area-icon" />
        <div v-else class="area-icon-placeholder">{{ area.name?.charAt(0) }}</div>
        <span class="area-name">{{ area.name }}</span>
      </router-link>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../../styles/variables';

.live-area-page { padding-bottom: 20px; }

.area-title {
  padding: 12px;
  font-size: 16px;
  font-weight: 600;
  background: $bg-white;
}

.area-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1px;
  background: $border-color;
  padding: 1px;
}

.area-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 8px;
  background: $bg-white;
  cursor: pointer;

  &:active { background: #f0f0f0; }
}

.area-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.area-icon-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: $theme-pink;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
}

.area-name { font-size: 12px; color: $text-primary; }

.loading {
  text-align: center;
  padding: 40px;
  color: $text-secondary;
}
</style>
