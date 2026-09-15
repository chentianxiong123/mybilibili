<script setup lang="ts">
import { ref } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'

defineProps<{
  description: string
  tags: string[]
}>()

const emit = defineEmits<{
  tagSearch: [tagName: string]
}>()

const isDescriptionCollapsed = ref(true)

const toggleDescription = () => {
  isDescriptionCollapsed.value = !isDescriptionCollapsed.value
}

const goToTagSearch = (tagName: string) => {
  emit('tagSearch', tagName)
}
</script>

<template>
  <div>
    <div class="video-description">
      <div class="description-content" :class="{ 'is-collapsed': isDescriptionCollapsed }">
        {{ description || '该视频暂无简介' }}
      </div>
      <div class="description-toggle" @click="toggleDescription">
        <span>{{ isDescriptionCollapsed ? '展开' : '收起' }}</span>
        <el-icon :class="{ 'is-rotated': !isDescriptionCollapsed }">
          <ArrowDown />
        </el-icon>
      </div>
    </div>

    <div class="video-tags" v-if="tags && tags.length > 0">
      <div class="tags-list">
        <span
          v-for="tag in tags"
          :key="tag"
          class="tag-item"
          @click="goToTagSearch(tag)"
        >
          {{ tag }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.video-description {
  background-color: #fff;
  padding: 20px 0;
}

.video-description .description-content {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  margin-bottom: 10px;
  transition: all 0.3s ease;
}

.video-description .description-content.is-collapsed {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.video-description .description-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #00a1d6;
  transition: all 0.3s ease;
}

.video-description .description-toggle:hover {
  color: #0091c6;
}

.video-description .description-toggle .el-icon {
  transition: transform 0.3s ease;
  font-size: 14px;
}

.video-description .description-toggle .el-icon.is-rotated {
  transform: rotate(180deg);
}

.video-tags {
  background-color: #fff;
  padding: 10px 0 20px 0;
  border-bottom: 1px solid #f0f0f0;
}

.video-tags .tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.video-tags .tag-item {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 16px;
  color: #666;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  user-select: none;
}

.video-tags .tag-item:hover {
  background-color: #e6f7ff;
  border-color: #00a1d6;
  color: #00a1d6;
}
</style>