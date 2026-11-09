// 上传前把图片转成 WebP（容器内不再依赖 ffmpeg 压缩）。
// GIF 保留原样（避免丢失动画）；JS 环境下转换失败则原样返回，后端兜底。
export async function toWebP(file: File, quality = 0.6, maxDimension = 1920): Promise<File> {
  if (!file.type.startsWith('image/') || file.type === 'image/gif' || file.type === 'image/webp') {
    return file
  }
  try {
    if (typeof ImageBitmap === 'undefined' || typeof OffscreenCanvas === 'undefined') {
      return file
    }
    const bitmap = await createImageBitmap(file)
    const { width, height } = bitmap
    if (width <= 0 || height <= 0) {
      bitmap.close()
      return file
    }
    const scale = Math.min(1, maxDimension / Math.max(width, height))
    const tw = Math.max(1, Math.round(width * scale))
    const th = Math.max(1, Math.round(height * scale))
    const canvas = new OffscreenCanvas(tw, th)
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      bitmap.close()
      return file
    }
    ctx.drawImage(bitmap, 0, 0, tw, th)
    bitmap.close()
    const blob = await canvas.convertToBlob({ type: 'image/webp', quality })
    const name = file.name.replace(/\.[^.]+$/i, '') + '.webp'
    return new File([blob], name, { type: 'image/webp' })
  } catch (e) {
    console.warn('toWebP failed, upload original:', e)
    return file
  }
}