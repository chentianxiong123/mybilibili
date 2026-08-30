<template>
  <div class="article-analysis">
    <div class="ranking-header">
      <h3 class="section-title">稿件排行</h3>
      <div class="ranking-actions">
        <el-select v-model="rankingSortBy" size="small" style="width: 120px;" @change="loadRanking">
          <el-option label="播放量排行" value="views" />
          <el-option label="点赞数排行" value="likes" />
          <el-option label="评论数排行" value="comments" />
        </el-select>
      </div>
    </div>

    <div class="manuscript-table" v-if="rankingList.length > 0">
      <el-table :data="rankingList" style="width: 100%">
        <el-table-column prop="title" label="稿件名称" min-width="200">
          <template #default="scope">
            <div class="manuscript-title-cell">
              <img :src="scope.row.coverUrl || '/default-cover.jpg'" class="manuscript-cover">
              <span class="manuscript-title">{{ scope.row.title }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="viewCount" label="播放量" width="100" align="center" />
        <el-table-column prop="likeCount" label="点赞数" width="100" align="center" />
        <el-table-column prop="commentCount" label="评论数" width="100" align="center" />
        <el-table-column prop="coinCount" label="硬币数" width="100" align="center" />
        <el-table-column prop="collectCount" label="收藏数" width="100" align="center" />
      </el-table>
    </div>

    <el-empty v-else description="暂无稿件数据" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useCreatorStats } from '@/composables/useCreatorStats'

const { rankingList, loadRanking } = useCreatorStats()

const rankingSortBy = ref('views')

onMounted(() => {
  loadRanking('views', 10)
})
</script>

<style scoped>
.article-analysis {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.ranking-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 16px 20px;
}
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
}
.manuscript-table {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 16px;
}
.manuscript-title-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.manuscript-cover {
  width: 80px;
  height: 45px;
  border-radius: 4px;
  object-fit: cover;
}
.manuscript-title {
  font-size: 14px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}
</style>
