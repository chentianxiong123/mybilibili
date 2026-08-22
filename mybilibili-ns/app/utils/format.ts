export function formatCount(num: number | string): string {
  const n = parseInt(String(num))
  if (isNaN(n)) return '0'
  if (n >= 10000) {
    return (Math.floor(n / 1000) / 10) + '万'
  }
  return String(n)
}

export function formatTimeLabel(timeStr: string | number): string {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return String(timeStr)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const r = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${y}年${m}月${r}日 ${h}:${min}`
}

export function formatStandardTime(timeStr: string | number): string {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return String(timeStr)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const r = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${r} ${h}:${min}:${s}`
}

export function formatTenThousand(num: number | string): string {
  const n = parseInt(String(num))
  if (isNaN(n)) return '0'
  if (n >= 10000) {
    return (n / 10000).toFixed(1) + '万'
  }
  return String(n)
}