/**
 * WebAV 适配器 - 把 WebAV 接到 Studio 状态
 * 来自 WebAV 真实接口(av-canvas / av-cliper)
 */
import { ref, shallowRef } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useTimelineStore } from '@/stores/timeline'
import { useEngineStore } from '@/stores/engine'

let _mods = null

async function loadMods() {
  if (_mods) return _mods
  if (typeof window === 'undefined' || !('VideoEncoder' in window)) {
    throw new Error('当前浏览器不支持 WebCodecs,WebAV 不可用')
  }
  const cliper = await import('@webav/av-cliper')
  const canvas = await import('@webav/av-canvas')
  _mods = { cliper, canvas }
  return _mods
}

export const canvasInstance = shallowRef(null)
export const webavSupported = ref(true)
export const webavError = ref(null)

/**
 * 初始化 AVCanvas,绑到一个 canvas DOM 元素。
 */
export async function initCanvas(canvasEl, width = 1920, height = 1080) {
  try {
    const { canvas } = await loadMods()
    const ac = new canvas.AVCanvas(canvasEl, { width, height, bgColor: '#000000' })
    canvasInstance.value = ac
    useEngineStore().setReady(true)
    webavError.value = null
    return ac
  } catch (e) {
    webavSupported.value = false
    webavError.value = e.message
    useEngineStore().setError(e.message)
    throw e
  }
}

/**
 * 把本地文件转成 IClip(MP4Clip / AudioClip / ImgClip),由 AVCanvas 渲染。
 */
export async function createClipFromFile(file) {
  const { cliper } = await loadMods()
  if (file.type.startsWith('video/')) {
    return new cliper.MP4Clip(await fetch(URL.createObjectURL(file)).then(r => r.body))
  }
  if (file.type.startsWith('audio/')) {
    return new cliper.AudioClip(await fetch(URL.createObjectURL(file)).then(r => r.body))
  }
  if (file.type.startsWith('image/')) {
    const bitmap = await createImageBitmap(file)
    return new cliper.ImgClip(bitmap)
  }
  throw new Error('不支持的文件类型:' + file.type)
}

/**
 * 把媒体追加到时间轴:创建 clip + 在 AVCanvas 上挂 sprite
 */
export async function appendMediaToCanvas({ mediaId, trackId, startTime = 0 }) {
  const project = useProjectStore()
  const media = project.mediaItems.find(m => m.id === mediaId)
  const ac = canvasInstance.value
  if (!media || !ac) return null

  const track = project.tracks.find(t => t.id === trackId)
  if (!track) return null
  if (track.kind === 'audio') {
    // 音频不渲染到画布,但保留时间轴
    const clip = project.addClipToTrack(trackId, { mediaId, label: media.name, startTime, duration: media.duration || 5 })
    return clip
  }

  // 从 file 还原为 IClip
  let file = media._file
  if (!file && media.url) {
    file = await fetch(media.url).then(r => r.blob()).then(b => new File([b], media.name, { type: media.kind + '/mp4' }))
  }
  const iclip = await createClipFromFile(file)

  // AVCanvas.capture + addSprite
  const clip = project.addClipToTrack(trackId, { mediaId, label: media.name, startTime, duration: iclip.meta.duration || 5 })
  const sprite = ac.capture(iclip, { duration: iclip.meta.duration || 5 })
  sprite.time = { offset: 0, duration: iclip.meta.duration || 5, start: startTime * 1e6 }
  sprite.zIndex = project.tracks.indexOf(track)
  ac.addSprite(sprite)
  // 记下 sprite 引用,后续操作同步
  clip._sprite = sprite
  return clip
}

/**
 * 同步 clip 的属性到已挂载的 sprite。
 */
export function syncClipToSprite(clipId, patch) {
  for (const t of useProjectStore().tracks) {
    const c = t.clips.find(c => c.id === clipId)
    if (c && c._sprite) {
      if (patch.startTime !== undefined) c._sprite.time.start = patch.startTime * 1e6
      if (patch.duration !== undefined) c._sprite.time.duration = patch.duration * 1e6
      if (patch.opacity !== undefined) c._sprite.opacity = patch.opacity
      if (patch.volume !== undefined && t.kind === 'audio') c._sprite.audio && (c._sprite.audio.volume = patch.volume)
    }
  }
}

/**
 * 删除 sprite
 */
export function removeSpriteOfClip(clipId) {
  const ac = canvasInstance.value
  if (!ac) return
  for (const t of useProjectStore().tracks) {
    const c = t.clips.find(c => c.id === clipId)
    if (c && c._sprite) {
      ac.removeSprite(c._sprite).catch(() => {})
      c._sprite = null
    }
  }
}

/**
 * 导出整个时间轴为 MP4 Blob
 */
export async function exportToMp4() {
  const { cliper } = await loadMods()
  const ac = canvasInstance.value
  if (!ac) throw new Error('画布未初始化')
  const engine = useEngineStore()
  const timeline = useTimelineStore()
  engine.setRendering(true)
  try {
    const out = ac.createCombinator({ fastStart: true, bitrate: 5e6 })
    const stream = out.output()
    const reader = stream.getReader()
    const chunks = []
    let received = 0
    const totalEstimate = (useProjectStore().totalDuration || 5) * 1e6
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      chunks.push(value)
      received += value.byteLength
      engine.setProgress(Math.min(99, Math.round(received / totalEstimate * 100)))
    }
    const blob = new Blob(chunks, { type: 'video/mp4' })
    const url = URL.createObjectURL(blob)
    engine.setRenderedUrl(url)
    engine.setProgress(100)
    return { blob, url, size: blob.size }
  } finally {
    engine.setRendering(false)
  }
}

export function downloadLastRender() {
  const engine = useEngineStore()
  if (!engine.lastRenderedUrl) return
  const a = document.createElement('a')
  a.href = engine.lastRenderedUrl
  a.download = (useProjectStore().project.name || 'studio') + '.mp4'
  a.click()
}
