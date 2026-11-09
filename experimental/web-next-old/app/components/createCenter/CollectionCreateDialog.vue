<template>
  <el-dialog
    v-model="dialogVisible"
    :title="mode === 'edit' ? '编辑合集' : '新建合集'"
    width="600px"
  >
    <el-form :model="form" label-width="80px">
      <el-form-item label="合集名称" required>
        <el-input v-model="form.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
      </el-form-item>
      <el-form-item label="合集描述">
        <el-input
          v-model="form.description"
          type="textarea"
          placeholder="请输入合集描述"
          :rows="3"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
      <el-form-item label="公开">
        <el-switch v-model="form.isPublic" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="emit('submit')" :loading="submitting">
        {{ mode === 'edit' ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  mode: {
    type: String,
    default: 'create'
  },
  form: {
    type: Object,
    required: true
  },
  submitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'submit'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})
</script>