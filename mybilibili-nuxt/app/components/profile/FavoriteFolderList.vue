<script setup>
import { MoreFilled } from '@element-plus/icons-vue'

const props = defineProps({
  collections: {
    type: Array,
    required: true
  },
  myFavorites: {
    type: Array,
    required: true
  },
  activeCategory: {
    type: String,
    required: true
  },
  myCollectionsExpanded: {
    type: Boolean,
    required: true
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'select-folder',
  'toggle-expand',
  'open-create',
  'open-edit',
  'delete-folder'
])
</script>

<template>
  <div class="favorites-sidebar">
    <div class="sidebar-section">
      <div class="section-header">
        <div class="section-title">我的创建</div>
        <div class="section-action" @click="emit('toggle-expand')">{{ myCollectionsExpanded ? '▼' : '▲' }}</div>
      </div>

      <div v-if="isOwnSpace && myCollectionsExpanded" class="new-collection-btn" @click="emit('open-create')">
        <div class="new-collection-icon">+</div>
        <div class="new-collection-text">新建收藏夹</div>
      </div>

      <div v-if="myCollectionsExpanded" class="collection-list">
        <div
          v-for="collection in collections"
          :key="collection.name"
          :class="['collection-item', { active: activeCategory === collection.name }]"
          @click="emit('select-folder', collection)"
        >
          <div class="collection-content">
            <span class="collection-icon">{{ collection.icon }}</span>
            <span class="collection-name">{{ collection.name }}</span>
          </div>
          <div class="collection-actions" @click.stop>
            <el-dropdown trigger="hover" placement="bottom-end">
              <el-button link class="more-btn">
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="emit('open-edit', collection)">
                    编辑信息
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="emit('delete-folder', collection)">
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <span class="collection-count">{{ collection.count }}</span>
        </div>

        <div
          v-for="favorite in myFavorites"
          :key="favorite.name"
          :class="['favorite-item', { active: activeCategory === favorite.name }]"
          @click="emit('select-folder', favorite)"
        >
          <div class="favorite-content">
            <span class="favorite-icon">{{ favorite.icon }}</span>
            <span class="favorite-name">{{ favorite.name }}</span>
          </div>
          <div class="favorite-actions" @click.stop>
            <el-dropdown trigger="hover" placement="bottom-end">
              <el-button link class="more-btn">
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="emit('open-edit', favorite)">
                    编辑信息
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="emit('delete-folder', favorite)">
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <span class="favorite-count">{{ favorite.count }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.favorites-sidebar {
  width: 220px;
  background-color: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.sidebar-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.section-action {
  cursor: pointer;
  color: #9499a0;
  font-size: 12px;
}

.new-collection-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-top: 10px;
  margin-bottom: 12px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  border: 1px dashed #e0e0e0;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.new-collection-btn:hover {
  color: #00aeec;
  border-color: #00aeec;
  background-color: rgba(0, 174, 236, 0.05);
}

.new-collection-icon {
  font-size: 16px;
  color: #9499a0;
}

.new-collection-text {
  font-size: 14px;
}

.collection-list {
  margin-bottom: 20px;
}

.collection-item, .favorite-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  border-radius: 4px;
  transition: all 0.3s ease;
  white-space: nowrap;
  flex-direction: row;
}

.collection-content, .favorite-content {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow: hidden;
  justify-content: flex-start;
  flex-direction: row;
}

.collection-count, .favorite-count {
  font-size: 12px;
  color: #9499a0;
  white-space: nowrap;
  flex-shrink: 0;
}

.collection-name, .favorite-name {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  text-align: left;
}

.collection-item:hover, .favorite-item:hover {
  background-color: rgba(0, 174, 236, 0.1);
  color: #00aeec;
}

.collection-item.active, .favorite-item.active {
  background-color: #00aeec;
  color: #fff;
}

.collection-icon, .favorite-icon {
  font-size: 16px;
  min-width: 16px;
  text-align: center;
}

.collection-item.active .collection-count, .favorite-item.active .favorite-count {
  color: rgba(255, 255, 255, 0.8);
}

.favorite-actions,
.collection-actions {
  opacity: 0;
  transition: opacity 0.3s ease;
  margin-right: 4px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.favorite-item:hover .favorite-actions,
.collection-item:hover .collection-actions {
  opacity: 1;
}

.favorite-actions .more-btn,
.collection-actions .more-btn {
  padding: 2px 4px;
  color: #666;
  font-size: 16px;
}

.favorite-actions .more-btn:hover,
.collection-actions .more-btn:hover {
  background-color: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.favorite-item.active .favorite-actions .more-btn,
.collection-item.active .collection-actions .more-btn {
  color: #fff;
}

.favorite-item.active .favorite-actions .more-btn:hover,
.collection-item.active .collection-actions .more-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}
</style>