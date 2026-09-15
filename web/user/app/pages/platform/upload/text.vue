<template>
  <div class="upload-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">文字投稿</h1>
      <p class="page-subtitle">发布图文动态，内容将展示在你的主页动态中</p>
    </div>

    <!-- 信息卡片 -->
    <div class="info-cards">
      <div class="info-card">
        <el-icon class="info-card-icon"><Document /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">文字内容</h3>
          <p class="info-card-desc">支持插入表情与@提及</p>
        </div>
      </div>
      <div class="info-card">
        <el-icon class="info-card-icon"><Picture /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">配图</h3>
          <p class="info-card-desc">最多9张图片，自动压缩为WebP</p>
        </div>
      </div>
      <div class="info-card">
        <el-icon class="info-card-icon"><Link /></el-icon>
        <div class="info-card-content">
          <h3 class="info-card-title">关联视频</h3>
          <p class="info-card-desc">可关联你的一个视频稿件</p>
        </div>
      </div>
    </div>

    <!-- 表单区域 -->
    <div class="form-area">
      <DynamicPublishPanel ref="publishPanelRef" @publish="handlePublish" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Picture, Link } from '@element-plus/icons-vue'
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
.upload-container {
  max-width: 720px;
}

.page-header {
  padding: 20px 0 8px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text1);
  margin: 0 0 6px 0;
}

.page-subtitle {
  font-size: 13px;
  color: var(--text3);
  margin: 0;
}

.info-cards {
  display: flex;
  gap: 12px;
  margin: 16px 0 20px;
}

.info-card {
  flex: 1;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  background: var(--graph_bg_thin, #f6f7f8);
  border-radius: 8px;
  padding: 12px 14px;
}

.info-card-icon {
  font-size: 22px;
  color: var(--brand_pink, #fb7299);
  margin-top: 2px;
}

.info-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text1);
  margin: 0 0 4px 0;
}

.info-card-desc {
  font-size: 12px;
  color: var(--text3);
  margin: 0;
}

.form-area {
  background: #fff;
  border: 1px solid var(--line_regular, #e3e5e7);
  border-radius: 8px;
  padding: 20px;
}
</style>