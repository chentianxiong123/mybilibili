import { describe, it, expect, vi, beforeEach } from 'vitest'
import { copyTextToClipboard } from './clipboard'

describe('copyTextToClipboard', () => {
  const mockMessage = {
    success: vi.fn(),
    error: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls message.success and returns true on successful copy', async () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined)
    const navigatorRef = { clipboard: { writeText: mockWriteText } }

    const result = await copyTextToClipboard('hello', {
      message: mockMessage,
      navigatorRef,
      documentRef: null,
    })

    expect(result).toBe(true)
    expect(mockWriteText).toHaveBeenCalledWith('hello')
    expect(mockMessage.success).toHaveBeenCalled()
  })

  it('calls message.error and returns false on copy failure', async () => {
    const mockWriteText = vi.fn().mockRejectedValue(new Error('fail'))
    const navigatorRef = { clipboard: { writeText: mockWriteText } }
    const textarea = { value: '', setAttribute: vi.fn(), style: {}, focus: vi.fn(), select: vi.fn(), remove: vi.fn() }
    const body = { appendChild: vi.fn(), removeChild: vi.fn() }
    const docRef = { body, createElement: vi.fn().mockReturnValue(textarea), execCommand: vi.fn().mockReturnValue(false) }

    const result = await copyTextToClipboard('hello', {
      message: mockMessage,
      navigatorRef,
      documentRef: docRef,
    })

    expect(result).toBe(false)
    expect(mockMessage.error).toHaveBeenCalled()
  })

  it('falls back when navigator.clipboard.writeText is not a function', async () => {
    const textarea = { value: '', setAttribute: vi.fn(), style: {}, focus: vi.fn(), select: vi.fn(), remove: vi.fn() }
    const body = { appendChild: vi.fn(), removeChild: vi.fn() }
    const docRef = { body, createElement: vi.fn().mockReturnValue(textarea), execCommand: vi.fn().mockReturnValue(false) }

    const result = await copyTextToClipboard('hello', {
      message: mockMessage,
      navigatorRef: {},
      documentRef: docRef,
    })

    expect(result).toBe(false)
    expect(mockMessage.error).toHaveBeenCalled()
  })

  it('converts null text to empty string', async () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined)
    const navigatorRef = { clipboard: { writeText: mockWriteText } }

    const result = await copyTextToClipboard(null, {
      message: mockMessage,
      navigatorRef,
      documentRef: null,
    })

    expect(mockWriteText).toHaveBeenCalledWith('')
    expect(result).toBe(true)
  })

  it('converts undefined text to empty string', async () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined)
    const navigatorRef = { clipboard: { writeText: mockWriteText } }

    const result = await copyTextToClipboard(undefined, {
      message: mockMessage,
      navigatorRef,
      documentRef: null,
    })

    expect(mockWriteText).toHaveBeenCalledWith('')
    expect(result).toBe(true)
  })
})
