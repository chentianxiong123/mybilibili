<script setup>
import { Star } from '@element-plus/icons-vue'

const props = defineProps({
  createVisible: {
    type: Boolean,
    default: false
  },
  editVisible: {
    type: Boolean,
    default: false
  },
  creatingFavorite: {
    type: Boolean,
    default: false
  },
  updatingFavorite: {
    type: Boolean,
    default: false
  },
  newFavoriteName: {
    type: String,
    default: ''
  },
  newFavoriteDescription: {
    type: String,
    default: ''
  },
  newFavoriteIsPublic: {
    type: Boolean,
    default: true
  },
  editingFavoriteName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'update:createVisible',
  'update:editVisible',
  'update:newFavoriteName',
  'update:newFavoriteDescription',
  'update:newFavoriteIsPublic',
  'update:editingFavoriteName',
  'create',
  'update',
  'cover-upload'
])
</script>

<template>
  <!-- 新建收藏夹对话框 -->
  <el-dialog
    :model-value="createVisible"
    @update:model-value="emit('update:createVisible', $event)"
    title="收藏夹信息"
    width="400px"
  >
    <div class="favorite-cover-section">
      <div class="cover-label">收藏夹封面</div>
      <div class="cover-upload-area">
        <div class="cover-placeholder" @click="$refs.coverInput?.click()">
          <el-icon class="cover-icon"><Star /></el-icon>
        </div>
        <input
          ref="coverInput"
          type="file"
          accept="image/*"
          style="display: none"
          @change="emit('cover-upload', $event)"
        />
      </div>
    </div>

    <div class="favorite-name-section">
      <div class="name-label">*收藏夹名称</div>
      <el-input
        :model-value="newFavoriteName"
        @update:model-value="emit('update:newFavoriteName', $event)"
        placeholder="收藏夹名称"
        maxlength="20"
        show-word-limit
      />
    </div>

    <div class="favorite-description-section">
      <div class="description-label">简介:</div>
      <el-input
        :model-value="newFavoriteDescription"
        @update:model-value="emit('update:newFavoriteDescription', $event)"
        type="textarea"
        placeholder="可填写简介"
        maxlength="200"
        show-word-limit
        :rows="4"
      />
    </div>

    <div class="favorite-public-section">
      <el-checkbox
        :model-value="newFavoriteIsPublic"
        @update:model-value="emit('update:newFavoriteIsPublic', $event)"
      >公开收藏夹</el-checkbox>
    </div>

    <template #footer>
      <el-button @click="emit('update:createVisible', false)">取消</el-button>
      <el-button type="primary" :loading="creatingFavorite" @click="emit('create')">
        提交
      </el-button>
    </template>
  </el-dialog>

  <!-- 编辑收藏夹对话框 -->
  <el-dialog
    :model-value="editVisible"
    @update:model-value="emit('update:editVisible', $event)"
    title="编辑收藏夹"
    width="400px"
  >
    <div class="favorite-name-section">
      <div class="name-label">*收藏夹名称</div>
      <el-input
        :model-value="editingFavoriteName"
        @update:model-value="emit('update:editingFavoriteName', $event)"
        placeholder="收藏夹名称"
        maxlength="20"
        show-word-limit
      />
    </div>

    <template #footer>
      <el-button @click="emit('update:editVisible', false)">取消</el-button>
      <el-button type="primary" :loading="updatingFavorite" @click="emit('update')">
        保存
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.favorite-cover-section {
  margin-bottom: 20px;
}

.cover-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.cover-upload-area {
  position: relative;
}

.cover-placeholder {
  width: 100px;
  height: 100px;
  background-color: #f0f0f0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cover-placeholder:hover {
  background-color: #e0e0e0;
}

.cover-icon {
  font-size: 40px;
  color: #999;
}

.favorite-name-section {
  margin-bottom: 20px;
}

.name-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.favorite-description-section {
  margin-bottom: 20px;
}

.description-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 10px;
}

.favorite-public-section {
  margin-bottom: 20px;
}
</style>