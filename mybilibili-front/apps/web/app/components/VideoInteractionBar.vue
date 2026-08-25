<script setup lang="ts">
import { safeStorage } from '@/utils/safeStorage'
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Star, Share, Comment, Edit, MoreFilled, CircleCheck, CirclePlus, ChatDotRound } from '@element-plus/icons-vue'
import { interactionApi, videoApi } from '@/api/client'

const props = defineProps<{
  manuscriptId: number
  videoInfo: any
  interactionStatus: any
}>()

const emit = defineEmits<{
  (e: 'update:interactionStatus', value: any): void
  (e: 'update:videoInfo', value: any): void
  (e: 'ai-assistant'): void
  (e: 'report'): void
}>()

const showFavoriteDialog = ref(false)
const favoriteFolders = ref<any[]>([])
const newFolderName = ref('')
const showNewFolderInput = ref(false)

const handleLike = async () => {
  const token = safeStorage.getItem("token")
  if (!token) { ElMessage.warning('请先登录'); return }

  const status = { ...props.interactionStatus }
  const info = { ...props.videoInfo }

  if (status.liked) {
    try {
      const response = await interactionApi.likeManuscript(props.manuscriptId, false)
      if (response.code === 200) {
        info.likeCount = Math.max(0, info.likeCount - 1)
        status.liked = false
        emit('update:videoInfo', info)
        emit('update:interactionStatus', status)
        ElMessage.success('取消点赞成功')
      }
    } catch (error) {
      console.error('取消点赞失败:', error)
      ElMessage.error('操作失败，请稍后重试')
    }
  } else {
    try {
      const response = await interactionApi.likeManuscript(props.manuscriptId, true)
      if (response.code === 200) {
        info.likeCount++
        status.liked = true
        emit('update:videoInfo', info)
        emit('update:interactionStatus', status)
        const likeBtn = document.querySelector('.like-btn')
        if (likeBtn) {
          likeBtn.classList.add('is-animating')
          setTimeout(() => likeBtn.classList.remove('is-animating'), 300)
        }
        ElMessage.success('点赞成功')
      }
    } catch (error) {
      console.error('点赞失败:', error)
      ElMessage.error('操作失败，请稍后重试')
    }
  }
}

const handleCoin = async () => {
  const token = safeStorage.getItem("token")
  if (!token) { ElMessage.warning('请先登录'); return }

  try {
    const { value: coinCount } = await ElMessageBox.prompt('请选择投币数量', '投币', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^[12]$/,
      inputErrorMessage: '请输入1或2',
      inputPlaceholder: '请输入投币数量（1或2）',
      inputValue: '1'
    })
    if (coinCount) {
      const response = await interactionApi.coinManuscript(props.manuscriptId, parseInt(coinCount))
      if (response.code === 200) {
        const info = { ...props.videoInfo }
        info.coinCount += parseInt(coinCount)
        emit('update:videoInfo', info)
        const status = { ...props.interactionStatus, coined: true, coinCount: (props.interactionStatus.coinCount || 0) + parseInt(coinCount) }
        emit('update:interactionStatus', status)
        const coinBtn = document.querySelector('.coin-btn')
        if (coinBtn) {
          coinBtn.classList.add('is-animating')
          setTimeout(() => coinBtn.classList.remove('is-animating'), 300)
        }
        ElMessage.success(`投币成功，投了${coinCount}个币`)
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('投币失败:', error)
      ElMessage.error('操作失败，请稍后重试')
    }
  }
}

const loadFavoriteFolders = async () => {
  try {
    const foldersResponse = await interactionApi.getFavoriteFolders()
    if (foldersResponse.code === 200) {
      favoriteFolders.value = foldersResponse.data
      const videoFoldersResponse = await interactionApi.getVideoFavoriteFolders(props.manuscriptId)
      if (videoFoldersResponse.code === 200) {
        favoriteFolders.value = favoriteFolders.value.map((folder: any) => ({ ...folder, selected: false }))
        const selectedFolderIds = videoFoldersResponse.data.map((f: any) => Number(f.id))
        favoriteFolders.value = favoriteFolders.value.map((folder: any) => ({
          ...folder, selected: selectedFolderIds.includes(Number(folder.id))
        }))
      }
    } else if (foldersResponse.code === 401) {
      ElMessage.warning('请先登录')
      showFavoriteDialog.value = false
    }
  } catch (error) {
    console.error('获取收藏夹失败:', error)
    ElMessage.error('获取收藏夹失败，请稍后重试')
  }
}

const handleFavorite = async () => {
  const token = safeStorage.getItem("token")
  if (!token) { ElMessage.warning('请先登录'); return }
  await loadFavoriteFolders()
  showFavoriteDialog.value = true
}

const toggleFolderSelection = (folderId: number) => {
  favoriteFolders.value = favoriteFolders.value.map((folder: any) => {
    if (Number(folder.id) === folderId) return { ...folder, selected: !folder.selected }
    return folder
  })
}

const showNewFolderForm = () => { showNewFolderInput.value = true }

const createNewFolder = async () => {
  if (!newFolderName.value.trim()) { ElMessage.warning('请输入收藏夹名称'); return }
  try {
    const response = await interactionApi.createFavoriteFolder({ name: newFolderName.value })
    if (response.code === 200) {
      await loadFavoriteFolders()
      newFolderName.value = ''
      showNewFolderInput.value = false
      ElMessage.success('新建收藏夹成功')
    }
  } catch (error) {
    console.error('新建收藏夹失败:', error)
    ElMessage.error('新建收藏夹失败，请稍后重试')
  }
}

const confirmFavorite = async () => {
  try {
    const selectedFolderIds = favoriteFolders.value.filter((f: any) => f.selected).map((f: any) => Number(f.id))
    const response = await interactionApi.updateVideoFavoriteFolders(props.manuscriptId, selectedFolderIds)
    if (response.code === 200) {
      const statusResponse = await interactionApi.getInteractionStatus(props.manuscriptId)
      if (statusResponse.code === 200) {
        const status = { ...props.interactionStatus, favorited: statusResponse.data.isCollected || statusResponse.data.collected || false }
        emit('update:interactionStatus', status)
      }
      const videoResponse = await videoApi.getVideoById(props.manuscriptId)
      if (videoResponse.code === 200) {
        const info = { ...props.videoInfo, collectCount: videoResponse.data.collectCount || 0 }
        emit('update:videoInfo', info)
      }
      showFavoriteDialog.value = false
      ElMessage.success('收藏更新成功')
    }
  } catch (error) {
    console.error('收藏失败:', error)
    ElMessage.error('操作失败，请稍后重试')
  }
}

const handleShare = async () => {
  const shareUrl = `${window.location.origin}/manuscript/${props.manuscriptId}`
  navigator.clipboard.writeText(shareUrl)
  try {
    const response = await interactionApi.shareManuscript(props.manuscriptId)
    if (response.code === 200) {
      const info = { ...props.videoInfo, shareCount: props.videoInfo.shareCount + 1 }
      emit('update:videoInfo', info)
      const status = { ...props.interactionStatus, shared: true }
      emit('update:interactionStatus', status)
    }
  } catch (error) {
    console.error('分享失败:', error)
  }
  const shareBtn = document.querySelector('.share-btn')
  if (shareBtn) {
    shareBtn.classList.add('is-animating')
    setTimeout(() => shareBtn.classList.remove('is-animating'), 300)
  }
  ElMessage.success('分享链接已复制到剪贴板')
}

const handleAIAssistant = () => { emit('ai-assistant') }
const handleTakeNotes = () => { console.log('打开记笔记') }
const handleReport = () => { emit('report') }
</script>

<template>
  <div>
    <div class="interaction-bar">
      <div class="left-actions">
        <el-button class="action-btn like-btn" :class="{ 'is-active': interactionStatus.liked }" @click="handleLike">
          <el-icon><CircleCheck /></el-icon>
          <span>{{ (videoInfo.likeCount || 0).toLocaleString() }}</span>
        </el-button>
        <el-button class="action-btn coin-btn" :class="{ 'is-active': interactionStatus.coined }" @click="handleCoin">
          <el-icon><CirclePlus /></el-icon>
          <span>{{ (videoInfo.coinCount || 0).toLocaleString() }}</span>
        </el-button>
        <el-button class="action-btn" :class="{ 'is-active': interactionStatus.favorited }" @click="handleFavorite">
          <el-icon><Star /></el-icon>
          <span>{{ (videoInfo.collectCount || 0).toLocaleString() }}</span>
        </el-button>
        <el-button class="action-btn share-btn" :class="{ 'is-active': interactionStatus.shared }" @click="handleShare">
          <el-icon><Share /></el-icon>
          <span>{{ (videoInfo.shareCount || 0).toLocaleString() }}</span>
        </el-button>
      </div>
      <div class="right-actions">
        <el-button class="action-btn ai-assistant-btn" @click="handleAIAssistant">
          <el-icon><Comment /></el-icon>
          <span>AI小助手</span>
        </el-button>
        <el-button class="action-btn" @click="handleTakeNotes">
          <el-icon><Edit /></el-icon>
          <span>记笔记</span>
        </el-button>
        <el-dropdown trigger="click">
          <el-button class="action-btn more-btn">
            <el-icon><MoreFilled /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleReport">
                <el-icon><ChatDotRound /></el-icon>
                稿件举报
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <el-dialog v-model="showFavoriteDialog" title="添加到收藏夹" width="400px" :close-on-click-modal="false">
      <div class="favorite-folders">
        <div v-for="folder in favoriteFolders" :key="folder.id" class="folder-item">
          <el-checkbox :checked="folder.selected" @change="toggleFolderSelection(Number(folder.id))">
            {{ folder.name }}
          </el-checkbox>
          <span class="folder-count">{{ folder.video_count || 0 }}/1000</span>
        </div>
        <div class="new-folder-section">
          <div v-if="!showNewFolderInput" class="new-folder-btn" @click="showNewFolderForm">
            <el-icon><CirclePlus /></el-icon>
            新建收藏夹
          </div>
          <div v-else class="new-folder-input">
            <el-input v-model="newFolderName" placeholder="最多可输入20个字" maxlength="20" size="small" />
            <el-button type="primary" size="small" @click="createNewFolder" style="margin-left: 8px;">新建</el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showFavoriteDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmFavorite">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.interaction-bar {
  background-color: #fff;
  padding: 8px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}
.interaction-bar .left-actions {
  display: flex;
  gap: 15px;
}
.interaction-bar .right-actions {
  display: flex;
  gap: 10px;
}
.interaction-bar .action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 6px 16px;
  border: none;
  background-color: #fff;
  color: #666;
  font-size: 14px;
  border-radius: 6px;
  transition: all 0.3s ease;
  min-width: 80px;
  min-height: 32px;
}
.interaction-bar .action-btn:hover {
  background-color: #f5f5f5;
  color: #00aeec;
}
.interaction-bar .action-btn.is-active {
  color: #00aeec;
  font-weight: 500;
}
.interaction-bar .action-btn.is-active:hover {
  background-color: #e6f7ff;
}
.interaction-bar .action-btn.is-animating {
  animation: likeAnimation 0.3s ease;
}
@keyframes likeAnimation {
  0% { transform: scale(1); }
  50% { transform: scale(1.3); }
  100% { transform: scale(1); }
}
.interaction-bar .action-btn .el-icon { font-size: 18px; }
.interaction-bar .action-btn span { font-size: 14px; }
.interaction-bar .ai-assistant-btn { gap: 8px; }
.interaction-bar .more-btn { padding: 6px 12px; min-width: 40px; }
.favorite-folders { max-height: 300px; overflow-y: auto; padding: 10px 0; }
.folder-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
.folder-count { color: #999; font-size: 12px; }
.new-folder-section { margin-top: 20px; padding-top: 10px; border-top: 1px solid #f0f0f0; }
.new-folder-btn { display: flex; align-items: center; gap: 8px; padding: 8px 0; cursor: pointer; color: #23ade5; font-size: 14px; }
.new-folder-btn:hover { color: #1a91d0; }
.new-folder-input { display: flex; align-items: center; margin-top: 10px; }
</style>