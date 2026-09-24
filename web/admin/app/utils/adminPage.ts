import { ElMessage, ElMessageBox } from 'element-plus'

export function normalizePagedResult(data) {
  if (Array.isArray(data)) {
    return { records: data, total: data.length }
  }
  const records = Array.isArray(data?.records) ? data.records : []
  return {
    records,
    total: Number(data?.total ?? records.length)
  }
}

export function formatDateTime(value) {
  if (!value) return '-'
  if (typeof value === 'string' && !value.includes('T') && !value.endsWith('Z')) {
    return value.slice(0, 19).replace('T', ' ')
  }
  const date = new Date(value)
  if (isNaN(date.getTime())) return String(value)
  const p = (n) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${p(date.getMonth() + 1)}-${p(date.getDate())} ${p(date.getHours())}:${p(date.getMinutes())}:${p(date.getSeconds())}`
}

export async function runConfirmedAction({
  message,
  title = '提示',
  action,
  successMessage,
  onSuccess,
  errorMessage = '操作失败'
}) {
  try {
    await ElMessageBox.confirm(message, title)
    await action()
    if (successMessage) ElMessage.success(successMessage)
    if (onSuccess) await onSuccess()
    return true
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(errorMessage)
    return false
  }
}
