import { beforeEach, describe, expect, it, vi } from 'vitest'
import { copyTextToClipboard } from '../utils/clipboard.js'

describe('copyTextToClipboard', () => {
  let message

  beforeEach(() => {
    message = {
      success: vi.fn(),
      error: vi.fn()
    }
    document.body.innerHTML = ''
    document.execCommand = vi.fn(() => true)
  })

  it('uses the async clipboard API when it is available', async () => {
    const writeText = vi.fn().mockResolvedValue()

    const copied = await copyTextToClipboard('ROOM01', {
      message,
      navigatorRef: { clipboard: { writeText } }
    })

    expect(copied).toBe(true)
    expect(writeText).toHaveBeenCalledWith('ROOM01')
    expect(message.success).toHaveBeenCalledWith('已复制')
    expect(message.error).not.toHaveBeenCalled()
  })

  it('falls back to a temporary textarea when async clipboard is unavailable', async () => {
    const copied = await copyTextToClipboard('rtmp://example/live', {
      successMessage: '复制成功',
      message,
      navigatorRef: {},
      documentRef: document
    })

    expect(copied).toBe(true)
    expect(document.execCommand).toHaveBeenCalledWith('copy')
    expect(document.querySelector('textarea')).toBeNull()
    expect(message.success).toHaveBeenCalledWith('复制成功')
  })

  it('falls back when async clipboard rejects', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('denied'))

    const copied = await copyTextToClipboard('invite-link', {
      message,
      navigatorRef: { clipboard: { writeText } },
      documentRef: document
    })

    expect(copied).toBe(true)
    expect(writeText).toHaveBeenCalledWith('invite-link')
    expect(document.execCommand).toHaveBeenCalledWith('copy')
    expect(message.success).toHaveBeenCalledWith('已复制')
  })

  it('reports failure when no copy strategy works', async () => {
    const copied = await copyTextToClipboard('secret', {
      failureMessage: '复制不了',
      message,
      navigatorRef: {},
      documentRef: { body: null }
    })

    expect(copied).toBe(false)
    expect(message.success).not.toHaveBeenCalled()
    expect(message.error).toHaveBeenCalledWith('复制不了')
  })
})
