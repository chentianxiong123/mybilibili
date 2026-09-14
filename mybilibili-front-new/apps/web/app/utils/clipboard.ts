import { ElMessage } from 'element-plus'

const fallbackCopyText = (text, documentRef) => {
  if (
    !documentRef?.body ||
    typeof documentRef.createElement !== 'function' ||
    typeof documentRef.execCommand !== 'function'
  ) {
    return false
  }

  const textarea = documentRef.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '-9999px'
  documentRef.body.appendChild(textarea)
  textarea.focus()
  textarea.select()

  try {
    return documentRef.execCommand('copy')
  } catch (e) {
    return false
  } finally {
    textarea.remove()
  }
}

export const copyTextToClipboard = async (text, {
  successMessage = '已复制',
  failureMessage = '复制失败，请手动复制',
  message = ElMessage,
  navigatorRef = typeof navigator !== 'undefined' ? navigator : null,
  documentRef = typeof document !== 'undefined' ? document : null
} = {}) => {
  const value = String(text ?? '')
  let copied = false

  if (typeof navigatorRef?.clipboard?.writeText === 'function') {
    try {
      await navigatorRef.clipboard.writeText(value)
      copied = true
    } catch (e) {
      copied = fallbackCopyText(value, documentRef)
    }
  } else {
    copied = fallbackCopyText(value, documentRef)
  }

  if (copied) {
    message?.success?.(successMessage)
    return true
  }

  message?.error?.(failureMessage)
  return false
}
