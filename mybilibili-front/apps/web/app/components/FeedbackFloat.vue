<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import { feedbackApi } from '@/api/client'

const visible = ref(false)
const sending = ref(false)
const feedbackType = ref('功能建议')
const feedbackContent = ref('')
const feedbackContact = ref('')
const feedbackTypes = ['功能建议', '问题反馈', '其他']

const canSubmit = computed(() => !!feedbackContent.value.trim() && !sending.value)

const open = () => {
  feedbackType.value = '功能建议'
  feedbackContent.value = ''
  feedbackContact.value = ''
  visible.value = true
}

const submit = async () => {
  const content = feedbackContent.value.trim()
  if (!content || sending.value) return
  sending.value = true
  try {
    await feedbackApi.submit({
      type: feedbackType.value,
      content,
      contact: feedbackContact.value.trim()
    })
    ElMessage.success('感谢反馈，我们会尽快处理')
    visible.value = false
  } catch (e) {
    ElMessage.error('提交失败，请稍后再试')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="feedback-float">
    <!-- 右下角悬浮反馈按钮 -->
    <button class="feedback-float-btn" @click="open" aria-label="意见反馈">
      <el-icon><ChatDotRound /></el-icon>
      <span>反馈</span>
    </button>

    <!-- 意见反馈弹窗 -->
    <el-dialog v-model="visible" title="意见反馈" width="480px" append-to-body>
      <div class="feedback-form">
        <p class="feedback-label">反馈类型</p>
        <el-radio-group v-model="feedbackType">
          <el-radio-button v-for="t in feedbackTypes" :key="t" :value="t">{{ t }}</el-radio-button>
        </el-radio-group>

        <p class="feedback-label">反馈内容 <i class="required">必填</i></p>
        <el-input
          v-model="feedbackContent"
          type="textarea"
          :rows="4"
          maxlength="500"
          show-word-limit
          placeholder="请描述你遇到的问题或建议…"
        />

        <p class="feedback-label">联系方式 <span class="optional">选填</span></p>
        <el-input
          v-model="feedbackContact"
          maxlength="100"
          placeholder="手机号 / QQ / 邮箱，便于我们联系你"
        />
      </div>

      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :disabled="!canSubmit" :loading="sending" @click="submit">
          提交反馈
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.feedback-float-btn {
  position: fixed;
  right: 20px;
  bottom: 120px;
  z-index: 900;
  display: flex;
  align-items: center;
  gap: 6px;
  height: 42px;
  padding: 0 16px;
  border: 1px solid rgba(251, 114, 153, 0.4);
  border-radius: 21px;
  background: #fff;
  color: #fb7299;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, box-shadow 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(251, 114, 153, 0.3);
  }
}

.feedback-form {
  .feedback-label {
    margin: 14px 0 8px;
    font-size: 13px;
    color: #61666d;
    font-weight: 500;

    &:first-child {
      margin-top: 0;
    }

    .required {
      color: #f56c6c;
      font-style: normal;
      font-size: 12px;
    }

    .optional {
      color: #909399;
      font-weight: 400;
      font-size: 12px;
    }
  }
}
</style>