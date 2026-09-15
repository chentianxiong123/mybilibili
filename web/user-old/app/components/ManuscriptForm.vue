<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  form: { type: Object, required: true },
  categories: { type: Array, default: () => [] },
  rules: { type: Object, default: () => ({}) }
})

const formRef = ref(null)
const tagInput = ref('')

const addTag = () => {
  const tag = tagInput.value.trim()
  if (!tag) return
  if (props.form.tags.length >= 20) {
    ElMessage.warning('最多只能添加20个标签')
    return
  }
  if (props.form.tags.includes(tag)) {
    ElMessage.warning('标签已存在')
    return
  }
  props.form.tags.push(tag)
  tagInput.value = ''
}

const removeTag = (index) => {
  props.form.tags.splice(index, 1)
}

const handleTagInputKeydown = (event) => {
  if (event.key === 'Enter') {
    event.preventDefault()
    addTag()
  }
}

defineExpose({
  validate: (callback) => formRef.value.validate(callback),
  formRef
})
</script>

<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="100px"
    class="upload-form"
  >
    <!-- 稿件标题 -->
    <el-form-item label="标题" prop="title">
      <el-input
        v-model="form.title"
        placeholder="请输入稿件标题（最多100字）"
        maxlength="100"
        show-word-limit
        class="title-input"
      ></el-input>
    </el-form-item>

    <!-- 稿件类型 -->
    <el-form-item label="类型">
      <el-radio-group v-model="form.type">
        <el-radio value="original">自制</el-radio>
        <el-radio value="repost">转载</el-radio>
      </el-radio-group>
    </el-form-item>

    <!-- 稿件分区 -->
    <el-form-item label="分区" prop="categoryId">
      <el-select
        v-model="form.categoryId"
        placeholder="请选择分区"
        class="category-select"
      >
        <el-option
          v-for="category in categories"
          :key="category.value"
          :label="category.label"
          :value="category.value"
        ></el-option>
      </el-select>
    </el-form-item>

    <!-- 稿件标签 -->
    <el-form-item label="标签">
      <div class="tag-section">
        <div class="tag-input-section">
          <el-input
            v-model="tagInput"
            placeholder="输入标签，按回车创建"
            @keydown="handleTagInputKeydown"
            class="tag-input"
            maxlength="20"
          >
            <template #append>
              <el-button @click="addTag" :disabled="!tagInput || form.tags.length >= 20">添加</el-button>
            </template>
          </el-input>
          <span class="tag-count">{{ form.tags.length }}/20</span>
        </div>
        <div class="tags-container">
          <el-tag
            v-for="(tag, index) in form.tags"
            :key="index"
            closable
            @close="removeTag(index)"
            class="tag-item"
            effect="plain"
          >
            {{ tag }}
          </el-tag>
        </div>
      </div>
    </el-form-item>

    <!-- 稿件简介 -->
    <el-form-item label="简介">
      <el-input
        v-model="form.description"
        type="textarea"
        placeholder="请输入稿件简介（最多2000字）"
        rows="4"
        maxlength="2000"
        show-word-limit
        class="description-input"
      ></el-input>
    </el-form-item>
  </el-form>
</template>

<style scoped>
.upload-form {
  max-width: 800px;
}

.title-input,
.category-select,
.description-input {
  width: 100%;
  max-width: 600px;
}

.tag-section {
  width: 100%;
  max-width: 600px;
}

.tag-input-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.tag-input {
  flex: 1;
}

.tag-count {
  font-size: 13px;
  color: #909399;
  white-space: nowrap;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  margin: 0;
}
</style>