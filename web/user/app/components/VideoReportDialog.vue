<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { reportApi } from '@/api/client'

const props = defineProps<{
  manuscriptId: number
}>()

const visible = defineModel<boolean>('visible', { default: false })

const reportForm = ref({
  reason: '',
  description: ''
})
const reportReasons = ['色情低俗', '政治敏感', '血腥暴力', '欺诈诈骗', '侵权抄袭', '垃圾广告', '引战', '其他']

const submitReport = async () => {
  if (!reportForm.value.reason) {
    ElMessage.warning('请选择举报原因')
    return
  }
  try {
    const res = await reportApi.submitReport({
      targetType: 'MANUSCRIPT',
      targetId: props.manuscriptId,
      manuscriptId: props.manuscriptId,
      reason: reportForm.value.reason,
      description: reportForm.value.description
    })
    if (res.code === 200 || res.success) {
      ElMessage.success(res.message || '举报成功')
      visible.value = false
    } else {
      ElMessage.error(res.message || '举报失败')
    }
  } catch (e) {
    ElMessage.error('举报失败')
  }
}

const resetForm = () => {
  reportForm.value = { reason: '', description: '' }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="稿件举报"
    width="420px"
    :close-on-click-modal="false"
    @open="resetForm"
  >
    <div style="padding: 10px 0;">
      <div style="margin-bottom: 16px;">
        <div style="font-size: 14px; color: #606266; margin-bottom: 8px;">举报原因</div>
        <div style="display: flex; flex-wrap: wrap; gap: 8px;">
          <el-tag
            v-for="r in reportReasons"
            :key="r"
            :type="reportForm.reason === r ? '' : 'info'"
            :effect="reportForm.reason === r ? 'dark' : 'plain'"
            style="cursor: pointer;"
            @click="reportForm.reason = r"
          >
            {{ r }}
          </el-tag>
        </div>
      </div>
      <div>
        <div style="font-size: 14px; color: #606266; margin-bottom: 8px;">补充说明（选填）</div>
        <el-input
          v-model="reportForm.description"
          type="textarea"
          :rows="3"
          placeholder="请详细描述举报内容..."
          maxlength="200"
          show-word-limit
        />
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="submitReport">提交举报</el-button>
    </template>
  </el-dialog>
</template>