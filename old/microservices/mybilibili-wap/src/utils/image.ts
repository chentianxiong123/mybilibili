// 获取浏览器支持的图片后缀
export function getPicSuffix(): string {
  const canvas = document.createElement('canvas')
  if (canvas.getContext && canvas.toDataURL('image/webp').indexOf('image/webp') === 0) {
    return '.webp'
  }
  return '.jpg'
}