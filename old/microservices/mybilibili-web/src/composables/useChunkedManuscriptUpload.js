import { ref, reactive, computed } from 'vue'
import { manuscriptApi } from '../api/manuscript.js'

const CHUNK_SIZE = 8 * 1024 * 1024
const MAX_CONCURRENT = 1
const MAX_RETRIES = 3

const STAGE_PREPARING = 'preparing'
const STAGE_UPLOADING = 'uploading'
const STAGE_MERGING = 'merging'
const STAGE_SUBMITTING = 'submitting'
const STAGE_COMPLETED = 'completed'
const STAGE_FAILED = 'failed'
const STAGE_CANCELLED = 'cancelled'

export const UPLOAD_STAGES = {
  PREPARING: STAGE_PREPARING,
  UPLOADING: STAGE_UPLOADING,
  MERGING: STAGE_MERGING,
  SUBMITTING: STAGE_SUBMITTING,
  COMPLETED: STAGE_COMPLETED,
  FAILED: STAGE_FAILED,
  CANCELLED: STAGE_CANCELLED
}

const STAGE_LABELS = {
  [STAGE_PREPARING]: '准备上传',
  [STAGE_UPLOADING]: '上传视频分片',
  [STAGE_MERGING]: '服务端合并中',
  [STAGE_SUBMITTING]: '提交稿件',
  [STAGE_COMPLETED]: '上传完成，等待审核/转码',
  [STAGE_FAILED]: '上传失败',
  [STAGE_CANCELLED]: '已取消'
}

export function useChunkedManuscriptUpload({
  api = manuscriptApi,
  chunkSize = CHUNK_SIZE,
  maxConcurrent = MAX_CONCURRENT,
  maxRetries = MAX_RETRIES
} = {}) {
  const stage = ref(STAGE_PREPARING)
  const percentage = ref(0)
  const uploadedBytes = ref(0)
  const totalBytes = ref(0)
  const currentPartIndex = ref(-1)
  const currentPartTitle = ref('')
  const currentChunkIndex = ref(-1)
  const totalChunks = ref(0)
  const uploadedChunks = ref(0)
  const speed = ref(0)
  const etaSeconds = ref(-1)
  const retryCount = ref(0)
  const error = ref(null)
  const uploadId = ref(null)
  const aborted = ref(false)
  const partProgress = ref([])

  let uploadStartTime = 0
  let lastBytesSnapshot = 0
  let lastSnapshotTime = 0
  let speedTimer = null
  let confirmedUploadedBytes = 0
  const activeChunkBytes = new Map()

  const stageLabel = computed(() => STAGE_LABELS[stage.value] || stage.value)
  const isUploading = computed(() => stage.value === STAGE_UPLOADING)
  const isFinished = computed(() => [STAGE_COMPLETED, STAGE_FAILED, STAGE_CANCELLED].includes(stage.value))

  function computeChunkCount(fileSize) {
    return Math.max(1, Math.ceil(fileSize / chunkSize))
  }

  function buildVideoParts(videos) {
    return videos.map((video, index) => ({
      title: video.title || `P${index + 1}`,
      fileName: video.file.name,
      size: video.file.size,
      videoOrder: video.sortOrder ?? index,
      durationSeconds: video.duration ?? video.durationSeconds ?? 0,
      totalChunks: computeChunkCount(video.file.size)
    }))
  }

  function computeTotalBytes(videos) {
    return videos.reduce((sum, v) => sum + v.file.size, 0)
  }

  function computeTotalChunks(videos) {
    return videos.reduce((sum, v) => sum + computeChunkCount(v.file.size), 0)
  }

  function reset() {
    stage.value = STAGE_PREPARING
    percentage.value = 0
    uploadedBytes.value = 0
    totalBytes.value = 0
    currentPartIndex.value = -1
    currentPartTitle.value = ''
    currentChunkIndex.value = -1
    totalChunks.value = 0
    uploadedChunks.value = 0
    speed.value = 0
    etaSeconds.value = -1
    retryCount.value = 0
    error.value = null
    uploadId.value = null
    aborted.value = false
    partProgress.value = []
    uploadStartTime = 0
    lastBytesSnapshot = 0
    lastSnapshotTime = 0
    if (speedTimer) {
      clearInterval(speedTimer)
      speedTimer = null
    }
    confirmedUploadedBytes = 0
    activeChunkBytes.clear()
  }

  function updateSpeed() {
    const now = Date.now()
    const elapsed = (now - lastSnapshotTime) / 1000
    if (elapsed < 0.5) return
    const bytesDiff = uploadedBytes.value - lastBytesSnapshot
    speed.value = Math.round(bytesDiff / elapsed)
    lastBytesSnapshot = uploadedBytes.value
    lastSnapshotTime = now
    if (speed.value > 0) {
      etaSeconds.value = Math.round((totalBytes.value - uploadedBytes.value) / speed.value)
    } else {
      etaSeconds.value = -1
    }
  }

  function updatePercentage() {
    if (totalBytes.value > 0) {
      percentage.value = Math.min(99, Math.round((uploadedBytes.value / totalBytes.value) * 100))
    }
  }

  function chunkKey(partIndex, chunkIndex) {
    return `${partIndex}:${chunkIndex}`
  }

  function recomputeUploadedBytes() {
    const activeBytes = Array.from(activeChunkBytes.values())
      .reduce((sum, bytes) => sum + bytes, 0)
    uploadedBytes.value = Math.min(totalBytes.value, confirmedUploadedBytes + activeBytes)
    updatePercentage()
  }

  function markChunkProgress(partIndex, chunkIndex, loaded, chunkSizeActual) {
    const boundedLoaded = Math.min(Math.max(loaded || 0, 0), chunkSizeActual)
    currentPartIndex.value = partIndex
    currentChunkIndex.value = chunkIndex
    activeChunkBytes.set(chunkKey(partIndex, chunkIndex), boundedLoaded)
    recomputeUploadedBytes()
    updateSpeed()
  }

  function markChunkUploaded(partIndex, chunkIndex, chunkSizeActual) {
    activeChunkBytes.delete(chunkKey(partIndex, chunkIndex))
    confirmedUploadedBytes += chunkSizeActual
    uploadedChunks.value++
    currentPartIndex.value = partIndex
    currentChunkIndex.value = chunkIndex
    if (partProgress.value[partIndex]) {
      partProgress.value[partIndex].uploaded++
    }
    recomputeUploadedBytes()
    updateSpeed()
  }

  function resolveProgressLoaded(progress, chunkSizeActual) {
    if (typeof progress === 'number') {
      return Math.round((chunkSizeActual * progress) / 100)
    }
    if (progress && typeof progress.loaded === 'number') {
      return progress.loaded
    }
    if (progress && typeof progress.percent === 'number') {
      return Math.round((chunkSizeActual * progress.percent) / 100)
    }
    return 0
  }

  async function uploadChunkWithRetry(uploadIdVal, partIndex, chunkIndex, chunkTotal, blob) {
    const key = chunkKey(partIndex, chunkIndex)
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      if (aborted.value) throw new Error('cancelled')
      try {
        markChunkProgress(partIndex, chunkIndex, 0, blob.size)
        const res = await api.uploadChunk(
          {
            uploadId: uploadIdVal,
            partIndex,
            chunkIndex,
            totalChunks: chunkTotal,
            file: blob
          },
          progress => {
            markChunkProgress(partIndex, chunkIndex, resolveProgressLoaded(progress, blob.size), blob.size)
          }
        )
        if (res && res.code !== undefined && res.code !== 200) {
          throw new Error(res.message || '分片上传失败')
        }
        return
      } catch (e) {
        activeChunkBytes.delete(key)
        recomputeUploadedBytes()
        updateSpeed()
        retryCount.value++
        if (attempt >= maxRetries) throw e
        await new Promise(r => setTimeout(r, 1000 * Math.pow(2, attempt)))
      }
    }
  }

  async function uploadPart(uploadIdVal, video, partIndex) {
    const file = video.file
    const total = computeChunkCount(file.size)
    currentPartIndex.value = partIndex
    currentPartTitle.value = video.title || `P${partIndex + 1}`
    totalChunks.value = total

    for (let i = 0; i < total; i++) {
      if (aborted.value) throw new Error('cancelled')
      const start = i * chunkSize
      const end = Math.min(start + chunkSize, file.size)
      const blob = file.slice(start, end)
      await uploadChunkWithRetry(uploadIdVal, partIndex, i, total, blob)
      markChunkUploaded(partIndex, i, end - start)
    }
  }

  async function runConcurrentUploads(uploadIdVal, videos) {
    const tasks = videos.map((video, index) => ({ video, index }))
    let cursor = 0

    async function worker() {
      while (cursor < tasks.length) {
        if (aborted.value) throw new Error('cancelled')
        const task = tasks[cursor++]
        await uploadPart(uploadIdVal, task.video, task.index)
      }
    }

    const workers = []
    for (let i = 0; i < Math.min(maxConcurrent, tasks.length); i++) {
      workers.push(worker())
    }
    await Promise.all(workers)
  }

  async function start(manuscriptData) {
    reset()
    const videos = manuscriptData.videos
    if (!videos || videos.length === 0) {
      error.value = '至少需要一个视频分P'
      stage.value = STAGE_FAILED
      throw new Error(error.value)
    }

    totalBytes.value = computeTotalBytes(videos)
    partProgress.value = videos.map((v, i) => ({
      index: i,
      title: v.title || `P${i + 1}`,
      total: computeChunkCount(v.file.size),
      uploaded: 0
    }))

    try {
      stage.value = STAGE_PREPARING
      const sessionPayload = {
        title: manuscriptData.title,
        description: manuscriptData.description || '',
        categoryId: manuscriptData.categoryId,
        tags: manuscriptData.tags || [],
        videos: buildVideoParts(videos)
      }

      const sessionRes = await api.createUploadSession(sessionPayload)
      if (!sessionRes || sessionRes.code !== 200 || !sessionRes.data) {
        throw new Error(sessionRes?.message || '创建上传会话失败')
      }
      uploadId.value = sessionRes.data.uploadId

      stage.value = STAGE_UPLOADING
      uploadStartTime = Date.now()
      lastSnapshotTime = uploadStartTime
      lastBytesSnapshot = 0
      speedTimer = setInterval(updateSpeed, 1000)

      await runConcurrentUploads(uploadId.value, videos)

      if (speedTimer) { clearInterval(speedTimer); speedTimer = null }

      stage.value = STAGE_MERGING
      percentage.value = 99

      stage.value = STAGE_SUBMITTING
      const completeRes = await api.completeUploadSession(
        uploadId.value,
        manuscriptData.cover
      )
      if (!completeRes || completeRes.code !== 200) {
        throw new Error(completeRes?.message || '提交稿件失败')
      }

      percentage.value = 100
      stage.value = STAGE_COMPLETED
      return completeRes.data
    } catch (e) {
      if (speedTimer) { clearInterval(speedTimer); speedTimer = null }
      if (e.message === 'cancelled') {
        stage.value = STAGE_CANCELLED
      } else {
        error.value = e.message || '上传失败'
        stage.value = STAGE_FAILED
      }
      throw e
    }
  }

  async function cancel() {
    aborted.value = true
    if (uploadId.value) {
      try {
        await api.cancelUploadSession(uploadId.value)
      } catch (e) { /* ignore */ }
    }
    stage.value = STAGE_CANCELLED
  }

  return {
    stage,
    stageLabel,
    percentage,
    uploadedBytes,
    totalBytes,
    currentPartIndex,
    currentPartTitle,
    currentChunkIndex,
    totalChunks,
    uploadedChunks,
    speed,
    etaSeconds,
    retryCount,
    error,
    uploadId,
    partProgress,
    isUploading,
    isFinished,
    start,
    cancel,
    reset,
    UPLOAD_STAGES
  }
}
