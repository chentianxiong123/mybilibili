<template>
  <div class="subtitle-manager">
    <!-- 页面级链接：我的视频 / 投稿的字幕（对齐真页 new-link-top-item） -->
    <div class="top-tabs">
      <span class="top-tab" :class="{ active: mode === 'mine' }" @click="mode = 'mine'">
        我的视频
      </span>
      <span class="top-tab" :class="{ active: mode === 'video' }" @click="mode = 'video'">
        投稿的字幕
      </span>
    </div>

    <div class="subtitle-content">
      <!-- 搜索栏 -->
      <div class="search-type">
        <el-select v-model="manuscriptFilter" clearable filterable placeholder="全部视频" class="video-select">
          <el-option label="全部视频" value=""></el-option>
          <el-option
            v-for="it in manuscripts"
            :key="it.manuscriptId"
            :label="it.title"
            :value="String(it.manuscriptId)"
          />
        </el-select>
        <el-input
          v-model="keyword"
          placeholder="搜索视频关键字"
          clearable
          class="keyword-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 稿件列表：对齐真页 archive-item（封面120x69 + 标题 + 文件名 + 字幕状态） -->
      <div class="zimu-body" v-loading="loading">
        <div v-if="filtered.length" class="archive-list">
          <div v-for="it in filtered" :key="it.manuscriptId" class="archive-item-wrap">
            <div class="archive-item">
              <img v-if="it.cover" :src="it.cover" class="cover" alt="" />
              <div v-else class="cover cover-fallback">暂无封面</div>
              <div class="archive-info">
                <div class="title">{{ it.title }}</div>
                <div class="file-name">{{ it.fileName }}</div>
              </div>
              <div class="text">
                <span class="subtitle-state" :class="subtitleClass(it)">
                  {{ subtitleText(it) }}
                </span>
                <span class="arrow">▸</span>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-else-if="!loading" description="暂无字幕任务" />
      </div>

      <BccPagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { manuscriptApi } from '@/api/creator'
import BccPagination from '@/components/createCenter/BccPagination.vue'

interface SubtitleVideo {
  id: number
  videoOrder: number
  title: string
  durationSeconds: number
  hasSubtitle: number
  processStage: string
  processStatus: number
}

interface SubtitleManuscript {
  manuscriptId: number
  title: string
  cover: string
  fileName: string
  hasSubtitle: number
  videos: SubtitleVideo[]
}

const mode = ref('mine')
const manuscripts = ref<SubtitleManuscript[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const manuscriptFilter = ref('')
const keyword = ref('')

const flatten = (list: any[]): SubtitleManuscript[] => {
  const out: SubtitleManuscript[] = []
  for (const it of list || []) {
    const videos: SubtitleVideo[] = Array.isArray(it.videos) ? it.videos : []
    const fileName = videos.map(v => v.title || `P${v.videoOrder + 1}`).join(' / ')
    out.push({
      manuscriptId: it.manuscriptId,
      title: it.title || '',
      cover: it.cover || '',
      videos,
      hasSubtitle: videos.length ? Math.max(...videos.map(v => v.hasSubtitle || 0)) : 0,
      fileName: fileName || '未命名视频',
    })
  }
  return out
}

const loadData = async () => {
  loading.value = true
  try {
    const res: any = await manuscriptApi.getMySubtitles({ page: page.value, size: pageSize.value })
    if (res.code === 200) {
      manuscripts.value = flatten(res.data?.list || [])
      total.value = res.data?.total || 0
    } else {
      ElMessage.error(res.message || '获取字幕任务失败')
    }
  } catch (e) {
    ElMessage.error('获取字幕任务失败')
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  let list = manuscripts.value
  if (manuscriptFilter.value) {
    list = list.filter(it => String(it.manuscriptId) === manuscriptFilter.value)
  }
  if (keyword.value.trim()) {
    const k = keyword.value.trim().toLowerCase()
    list = list.filter(it => it.title.toLowerCase().includes(k) || it.fileName.toLowerCase().includes(k))
  }
  return list
})

const subtitleText = (it: SubtitleManuscript) => {
  if (it.hasSubtitle > 0) return '多语言'
  if (it.videos.some(v => v.processStage === 'SUB_TITLE_GENERATING' || v.processStatus === 3)) return '生成中'
  return '无字幕'
}

const subtitleClass = (it: SubtitleManuscript) => {
  if (it.hasSubtitle > 0) return 'state-has'
  if (it.videos.some(v => v.processStage === 'SUB_TITLE_GENERATING' || v.processStatus === 3)) return 'state-gen'
  return 'state-none'
}

const handlePageChange = () => {
  loadData()
}

watch(mode, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.subtitle-manager {
  padding: 0 40px;
}

/* 页面级链接：16px 选中 700 粉色 + 底边线（对齐真页 link-top-item） */
.top-tabs {
  display: flex;
  align-items: center;
  height: 64px;
  border-bottom: 1px solid #e7e7e7;
}

.top-tab {
  position: relative;
  font-size: 16px;
  color: #505050;
  cursor: pointer;
  margin-right: 40px;
  line-height: 64px;
}

.top-tab.active {
  color: var(--brand_pink, #ff6699);
  font-weight: 700;
}

.top-tab.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 2.5px;
  background: var(--brand_pink, #ff6699);
}

.subtitle-content {
  padding: 0 40px 20px;
}

/* 搜索栏：对齐真页 search-type */
.search-type {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 25px 0;
}

.video-select {
  width: 200px;
}

.keyword-input {
  width: 240px;
}

.keyword-input :deep(.el-input__wrapper) {
  border-radius: 4px;
  box-shadow: 0 0 0 1px #e3e5e7 inset;
}

/* 稿件列表：对齐真页 archive-item（封面120x69，item 间隔 16px） */
.zimu-body {
  margin-top: 10px;
}

.archive-list {
  display: flex;
  flex-direction: column;
}

.archive-item-wrap {
  padding: 16px 0;
  border-bottom: 1px solid #f0f1f4;
}

.archive-item {
  display: flex;
  align-items: center;
}

.cover {
  width: 120px;
  height: 69px;
  flex-shrink: 0;
  margin-right: 12px;
  border-radius: 6px;
  object-fit: cover;
  background: #e3e5e8;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #99a2aa;
}

.archive-info {
  width: 320px;
  min-width: 0;
}

.archive-info .title {
  font-size: 14px;
  font-weight: 500;
  color: #18191c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-info .file-name {
  margin-top: 6px;
  font-size: 12px;
  color: #9499a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  font-size: 14px;
  color: #18191c;
}

.subtitle-state {
  cursor: pointer;
}

.subtitle-state.state-has {
  color: #00b96b;
}

.subtitle-state.state-gen {
  color: #e6a23c;
}

.subtitle-state.state-none {
  color: #9499a0;
}

.arrow {
  font-size: 12px;
  color: #9499a0;
}
</style>