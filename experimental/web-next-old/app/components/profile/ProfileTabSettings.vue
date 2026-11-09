<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { userPrivacyApi } from '@/api/userPrivacy.ts'

const props = defineProps({
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

// 隐私设置数据
const privacySettings = ref({
  publicCollection: true,
  publicBirthdayTags: false,
  publicFollowingList: false,
  publicFollowersList: false
})

// 用户标签
const userTags = ref([])
const newTagInput = ref('')

// 加载隐私设置
const loadPrivacySettings = async () => {
  if (!props.isOwnSpace) return
  try {
    const res = await userPrivacyApi.getPrivacySettings()
    if (res.code === 200) {
      privacySettings.value = {
        publicCollection: res.data.publicCollection ?? true,
        publicBirthdayTags: res.data.publicBirthdayTags ?? false,
        publicFollowingList: res.data.publicFollowingList ?? false,
        publicFollowersList: res.data.publicFollowersList ?? false
      }
      userTags.value = res.data.tags || []
    }
  } catch (error) {
    // 接口不存在时不输出错误日志，避免控制台报错
    if (error.response?.status !== 404) {
      console.error('加载隐私设置失败:', error)
    }
  }
}

// 处理隐私设置变更
const handlePrivacyChange = async (key, value) => {
  try {
    const data = { [key]: value }
    const res = await userPrivacyApi.updatePrivacySettings(data)
    if (res.code === 200) {
      ElMessage.success('设置已保存')
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    console.error('保存隐私设置失败:', error)
    ElMessage.error('保存失败')
  }
}

// 添加标签
const handleAddTag = async () => {
  const tagName = newTagInput.value.trim()
  if (!tagName) return
  if (userTags.value.includes(tagName)) {
    ElMessage.warning('标签已存在')
    return
  }
  if (userTags.value.length >= 10) {
    ElMessage.warning('最多只能添加10个标签')
    return
  }
  try {
    const res = await userPrivacyApi.addUserTag(tagName)
    if (res.code === 200) {
      userTags.value.push(tagName)
      newTagInput.value = ''
      ElMessage.success('添加成功')
    } else {
      ElMessage.error(res.message || '添加失败')
    }
  } catch (error) {
    console.error('添加标签失败:', error)
    ElMessage.error('添加失败')
  }
}

// 删除标签
const handleRemoveTag = async (tag) => {
  try {
    const res = await userPrivacyApi.removeUserTag(tag)
    if (res.code === 200) {
      userTags.value = userTags.value.filter(t => t !== tag)
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    console.error('删除标签失败:', error)
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadPrivacySettings()
})
</script>

<template>
  <div class="settings-section">
    <div class="settings-container">
      <!-- 隐私设置 -->
      <div class="settings-group">
        <h3 class="settings-group-title">隐私设置</h3>
        <div class="settings-list">
          <div class="setting-item">
            <span class="setting-label">公开我的收藏</span>
            <div class="setting-control">
              <el-switch
                v-model="privacySettings.publicCollection"
                :active-value="true"
                :inactive-value="false"
                active-text="公开"
                inactive-text="隐藏"
                @change="handlePrivacyChange('publicCollection', $event)"
              />
            </div>
          </div>
          <div class="setting-item">
            <span class="setting-label">公开我的生日、个人标签</span>
            <div class="setting-control">
              <el-switch
                v-model="privacySettings.publicBirthdayTags"
                :active-value="true"
                :inactive-value="false"
                active-text="公开"
                inactive-text="隐藏"
                @change="handlePrivacyChange('publicBirthdayTags', $event)"
              />
            </div>
          </div>
          <div class="setting-item">
            <span class="setting-label">公开我的关注列表</span>
            <div class="setting-control">
              <el-switch
                v-model="privacySettings.publicFollowingList"
                :active-value="true"
                :inactive-value="false"
                active-text="公开"
                inactive-text="隐藏"
                @change="handlePrivacyChange('publicFollowingList', $event)"
              />
            </div>
          </div>
          <div class="setting-item">
            <span class="setting-label">公开我的粉丝列表</span>
            <div class="setting-control">
              <el-switch
                v-model="privacySettings.publicFollowersList"
                :active-value="true"
                :inactive-value="false"
                active-text="公开"
                inactive-text="隐藏"
                @change="handlePrivacyChange('publicFollowersList', $event)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 我的个人标签 -->
      <div class="settings-group">
        <h3 class="settings-group-title">我的个人标签</h3>
        <div class="tags-section">
          <div class="tags-list">
            <el-tag
              v-for="tag in userTags"
              :key="tag"
              closable
              class="user-tag"
              @close="handleRemoveTag(tag)"
            >
              {{ tag }}
            </el-tag>
          </div>
          <div class="tag-input-wrapper">
            <el-input
              v-model="newTagInput"
              placeholder="输入标签名称"
              maxlength="10"
              show-word-limit
              class="tag-input"
              @keyup.enter="handleAddTag"
            />
            <el-button type="primary" @click="handleAddTag" :disabled="!newTagInput.trim()">
              新增
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ==================== 设置页面样式 ==================== */
.settings-section {
  background-color: #fff;
  border-radius: 8px;
  padding: 24px;
  min-height: 500px;
}

.settings-container {
  max-width: 800px;
}

.settings-group {
  margin-bottom: 32px;
}

.settings-group-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0 0 20px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.settings-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
}

.setting-label {
  font-size: 14px;
  color: #333;
}

.setting-control {
  display: flex;
  align-items: center;
}

.setting-control .el-switch {
  margin-right: 12px;
}

.setting-status {
  font-size: 14px;
  color: #999;
  min-width: 40px;
  text-align: right;
}

/* 标签管理样式 */
.tags-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.user-tag {
  font-size: 13px;
  padding: 6px 12px;
  border-radius: 4px;
}

.tag-input-wrapper {
  display: flex;
  gap: 10px;
  align-items: center;
}

.tag-input {
  width: 200px;
}

.tag-input .el-input__inner {
  border-radius: 4px;
}
</style>