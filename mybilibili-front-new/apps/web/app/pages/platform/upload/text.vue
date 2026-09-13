<template>
  <div class="text-upload">
    <div class="upload-header">
      <h2 class="title">文字投稿</h2>
      <p class="subtitle">发布图文动态，内容将展示在你的主页动态中</p>
    </div>

    <div class="upload-card">
      <DynamicPublishPanel ref="publishPanelRef" @publish="handlePublish" />
    </div>

    <div class="upload-tips">
      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="发布说明"
        description="文字投稿以图文动态形式发布，可插入表情、添加图片（最多9张）或关联自己的视频稿件。"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import DynamicPublishPanel from '@/components/DynamicPublishPanel.vue'
import { dynamicApi } from '@/api/dynamic'

const publishPanelRef = ref<InstanceType<typeof DynamicPublishPanel> | null>(null)

const handlePublish = async (payload: any) => {
  try {
    const formData = new FormData()
    formData.append('content', payload.content)
    if (payload.refVideoId) {
      formData.append('refVideoId', payload.refVideoId)
    }
    ;(payload.images || []).forEach((file: File) => {
      formData.append('images', file)
    })
    const res: any = await dynamicApi.publishDynamic(formData)
    if (res.code === 200) {
      ElMessage.success('发布成功')
      publishPanelRef.value?.reset()
    } else {
      ElMessage.error(res.message || '发布失败')
    }
  } catch (error: any) {
    ElMessage.error('发布失败：' + (error?.message || ''))
  }
}
</script>

<style scoped>
.text-upload {
  padding: 24px;
  max-width: 720px;
}

.upload-header {
  margin-bottom: 20px;
}

.title {
  font-size: 22px;
  color: var(--text1, #18191c);
  font-weight: 600;
  margin: 0 0 6px 0;
}

.subtitle {
  font-size: 13px;
  color: var(--text3, #9499a0);
  margin: 0;
}

.upload-card {
  background: #fff;
  border: 1px solid var(--line_regular, #e3e5e7);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
}

.upload-tips {
  margin-top: 8px;
}
</style>