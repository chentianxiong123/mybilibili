<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getHomeBanners,
  addHomeBanner,
  updateHomeBanner,
  deleteHomeBanner,
  getCategoryBanners,
  addCategoryBanner,
  updateCategoryBanner,
  deleteCategoryBanner,
  getBackgroundImage,
  saveBackgroundImage,
  deleteBackgroundImage,
  getUserProfileBackground,
  saveUserProfileBackground,
  deleteUserProfileBackground,
  uploadBannerImage
} from '../api/bannerImage'
import { getCategoryList } from '../api/category'

// 类型选项
const typeOptions = [
  { label: '首页轮播图', value: 'home' },
  { label: '分类轮播图', value: 'category' },
  { label: '顶部背景图', value: 'background' },
  { label: '用户主页背景', value: 'userProfile' }
]

// 分类选项
const categoryOptions = ref([])

// 当前选中的类型
const currentType = ref('home')
const currentCategoryId = ref(0)

// 数据列表
const bannerList = ref([])
const backgroundData = ref(null)
const userProfileBackgroundData = ref(null)
const loading = ref(false)

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加轮播图')
const isEdit = ref(false)
const editId = ref(null)

// 表单数据
const formData = ref({
  title: '',
  imageUrl: '',
  linkUrl: '',
  sortOrder: 1,
  status: 1,
  startTime: null,
  endTime: null
})

// 图片上传加载状态
const uploadLoading = ref(false)

// 根据类型获取数据
const loadData = async () => {
  loading.value = true
  try {
    if (currentType.value === 'home') {
      const res = await getHomeBanners()
      if (res.code === 200) {
        bannerList.value = res.data || []
      }
    } else if (currentType.value === 'category') {
      const res = await getCategoryBanners(currentCategoryId.value)
      if (res.code === 200) {
        bannerList.value = res.data || []
      }
    } else if (currentType.value === 'background') {
      const res = await getBackgroundImage()
      if (res.code === 200) {
        backgroundData.value = res.data
      }
    } else if (currentType.value === 'userProfile') {
      const res = await getUserProfileBackground()
      if (res.code === 200) {
        userProfileBackgroundData.value = res.data
      }
    }
  } catch (error) {
    ElMessage.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 类型改变
const handleTypeChange = () => {
  bannerList.value = []
  backgroundData.value = null
  userProfileBackgroundData.value = null
  loadData()
}

// 分类改变
const handleCategoryChange = () => {
  loadData()
}

// 打开添加对话框
const handleAdd = () => {
  isEdit.value = false
  editId.value = null
  if (currentType.value === 'background' || currentType.value === 'userProfile') {
    dialogTitle.value = '设置背景图'
  } else {
    dialogTitle.value = '添加轮播图'
  }
  formData.value = {
    title: '',
    imageUrl: '',
    linkUrl: '',
    sortOrder: 1,
    status: 1,
    startTime: null,
    endTime: null
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row) => {
  isEdit.value = true
  editId.value = row.id
  dialogTitle.value = '编辑轮播图'
  formData.value = {
    title: row.title,
    imageUrl: row.imageUrl,
    linkUrl: row.linkUrl,
    sortOrder: row.sortOrder,
    status: row.status,
    startTime: row.startTime,
    endTime: row.endTime
  }
  dialogVisible.value = true
}

// 保存
const handleSave = async () => {
  if (!formData.value.title) {
    ElMessage.warning('请输入标题')
    return
  }
  if (!formData.value.imageUrl) {
    ElMessage.warning('请上传图片')
    return
  }

  try {
    let res
    if (currentType.value === 'home') {
      if (isEdit.value) {
        res = await updateHomeBanner(editId.value, formData.value)
      } else {
        res = await addHomeBanner(formData.value)
      }
    } else if (currentType.value === 'category') {
      if (isEdit.value) {
        res = await updateCategoryBanner(currentCategoryId.value, editId.value, formData.value)
      } else {
        res = await addCategoryBanner(currentCategoryId.value, formData.value)
      }
    } else if (currentType.value === 'background') {
      res = await saveBackgroundImage(formData.value)
    } else if (currentType.value === 'userProfile') {
      res = await saveUserProfileBackground(formData.value)
    }

    if (res.code === 200) {
      ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除吗？', '提示', { type: 'warning' })
    let res
    if (currentType.value === 'home') {
      res = await deleteHomeBanner(row.id)
    } else if (currentType.value === 'category') {
      res = await deleteCategoryBanner(currentCategoryId.value, row.id)
    }
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadData()
    }
  } catch {}
}

// 删除背景图
const handleDeleteBackground = async () => {
  try {
    await ElMessageBox.confirm('确定要删除背景图吗？', '提示', { type: 'warning' })
    const res = await deleteBackgroundImage()
    if (res.code === 200) {
      ElMessage.success('删除成功')
      backgroundData.value = null
    }
  } catch {}
}

// 删除用户主页背景图
const handleDeleteUserProfileBackground = async () => {
  try {
    await ElMessageBox.confirm('确定要删除用户主页背景图吗？', '提示', { type: 'warning' })
    const res = await deleteUserProfileBackground()
    if (res.code === 200) {
      ElMessage.success('删除成功')
      userProfileBackgroundData.value = null
    }
  } catch {}
}

// 图片上传
const handleUpload = async (file) => {
  uploadLoading.value = true
  try {
    const res = await uploadBannerImage(file.raw)
    if (res.code === 200) {
      formData.value.imageUrl = res.data.url
      ElMessage.success('上传成功')
    } else {
      ElMessage.error(res.message || '上传失败')
    }
  } catch (error) {
    ElMessage.error('上传失败')
  } finally {
    uploadLoading.value = false
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '永久'
  return new Date(time).toLocaleString()
}

// 状态文本
const getStatusText = (status) => {
  return status === 1 ? '启用' : '禁用'
}

onMounted(async () => {
  await loadCategories()
  loadData()
})

const loadCategories = async () => {
  try {
    const res = await getCategoryList()
    if (res.code === 200) {
      categoryOptions.value = res.data.map(c => ({
        label: c.name,
        value: c.id
      }))
    }
  } catch (error) {
    console.error('获取分类列表失败:', error)
  }
}
</script>

<template>
  <div class="banner-images-page">
    <h2 class="page-title">图片管理</h2>

    <!-- 筛选区域 -->
    <div class="filter-bar">
      <el-select v-model="currentType" @change="handleTypeChange" style="width: 150px">
        <el-option
          v-for="item in typeOptions"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>

      <el-select
        v-if="currentType === 'category'"
        v-model="currentCategoryId"
        @change="handleCategoryChange"
        style="width: 150px; margin-left: 12px"
      >
        <el-option
          v-for="item in categoryOptions"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>

      <el-button type="primary" @click="handleAdd" style="margin-left: auto">
        <el-icon><Plus /></el-icon>
        {{ currentType === 'background' || currentType === 'userProfile' ? '设置背景图' : '添加轮播图' }}
      </el-button>
    </div>

    <!-- 首页/分类轮播图列表 -->
    <el-table
      v-if="currentType !== 'background' && currentType !== 'userProfile'"
      v-loading="loading"
      :data="bannerList"
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="图片" width="150">
        <template #default="{ row }">
          <el-image
            :src="row.imageUrl"
            style="width: 120px; height: 60px; object-fit: cover"
            fit="cover"
          />
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="150" />
      <el-table-column prop="linkUrl" label="链接" min-width="150" show-overflow-tooltip />
      <el-table-column prop="sortOrder" label="排序" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="有效期" min-width="180">
        <template #default="{ row }">
          <div v-if="row.startTime || row.endTime">
            <div v-if="row.startTime">开始: {{ formatTime(row.startTime) }}</div>
            <div v-if="row.endTime">结束: {{ formatTime(row.endTime) }}</div>
          </div>
          <span v-else>永久有效</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="150">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 顶部背景图展示 -->
    <div v-else-if="currentType === 'background'" class="background-section">
      <el-empty v-if="!backgroundData" description="暂无背景图" />
      <div v-else class="background-card">
        <el-image
          :src="backgroundData.imageUrl"
          style="width: 100%; height: 200px; object-fit: cover"
          fit="cover"
        />
        <div class="background-info">
          <h4>{{ backgroundData.title }}</h4>
          <p>链接: {{ backgroundData.linkUrl || '无' }}</p>
          <p>状态: <el-tag :type="backgroundData.status === 1 ? 'success' : 'danger'">{{ getStatusText(backgroundData.status) }}</el-tag></p>
          <p>有效期: {{ formatTime(backgroundData.startTime) }} ~ {{ formatTime(backgroundData.endTime) }}</p>
          <div class="background-actions">
            <el-button type="primary" size="small" @click="handleEdit(backgroundData)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDeleteBackground">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户主页背景图展示 -->
    <div v-else-if="currentType === 'userProfile'" class="background-section">
      <el-empty v-if="!userProfileBackgroundData" description="暂无用户主页背景图" />
      <div v-else class="background-card">
        <el-image
          :src="userProfileBackgroundData.imageUrl"
          style="width: 100%; height: 200px; object-fit: cover"
          fit="cover"
        />
        <div class="background-info">
          <h4>{{ userProfileBackgroundData.title }}</h4>
          <p>链接: {{ userProfileBackgroundData.linkUrl || '无' }}</p>
          <p>状态: <el-tag :type="userProfileBackgroundData.status === 1 ? 'success' : 'danger'">{{ getStatusText(userProfileBackgroundData.status) }}</el-tag></p>
          <p>有效期: {{ formatTime(userProfileBackgroundData.startTime) }} ~ {{ formatTime(userProfileBackgroundData.endTime) }}</p>
          <div class="background-actions">
            <el-button type="primary" size="small" @click="handleEdit(userProfileBackgroundData)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDeleteUserProfileBackground">删除</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="formData" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="formData.title" placeholder="请输入标题" />
        </el-form-item>

        <el-form-item label="图片" required>
          <el-upload
            class="avatar-uploader"
            action="#"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleUpload"
            accept="image/*"
          >
            <img v-if="formData.imageUrl" :src="formData.imageUrl" class="avatar" />
            <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">建议尺寸：首页轮播图 1200x400，分类轮播图 800x300，背景图 1920x150</div>
        </el-form-item>

        <el-form-item label="链接">
          <el-input v-model="formData.linkUrl" placeholder="请输入跳转链接，如 /video/1" />
        </el-form-item>

        <el-form-item label="排序" v-if="currentType !== 'background'">
          <el-input-number v-model="formData.sortOrder" :min="1" :max="100" />
        </el-form-item>

        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="开始时间">
          <el-date-picker
            v-model="formData.startTime"
            type="datetime"
            placeholder="选择开始时间"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="结束时间">
          <el-date-picker
            v-model="formData.endTime"
            type="datetime"
            placeholder="选择结束时间"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="uploadLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.banner-images-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.background-section {
  padding: 20px;
}

.background-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  max-width: 600px;
}

.background-info {
  padding: 20px;
}

.background-info h4 {
  margin: 0 0 12px;
  font-size: 18px;
}

.background-info p {
  margin: 8px 0;
  color: #666;
}

.background-actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
}

.avatar-uploader {
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: var(--el-transition-duration-fast);
  width: 300px;
  height: 150px;
}

.avatar-uploader:hover {
  border-color: var(--el-color-primary);
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 300px;
  height: 150px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar {
  width: 300px;
  height: 150px;
  object-fit: cover;
  display: block;
}

.upload-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}
</style>
