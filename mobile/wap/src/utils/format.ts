// 万单位格式化
export function formatTenThousand(num: number): string {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return String(num)
}

// 时间格式化：秒 → mm:ss
export function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

// 日期格式化
export function formatDate(dateStr: string): string {
  const d = new Date(dateStr.replace(/-/g, '/'))
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return minutes + '分钟前'
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return hours + '小时前'
  const days = Math.floor(hours / 24)
  if (days < 30) return days + '天前'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}