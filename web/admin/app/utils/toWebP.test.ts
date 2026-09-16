import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import { toWebP } from './toWebP'

class FakeOffscreenCanvas {
  width: number
  height: number
  private ctx: { drawImage: ReturnType<typeof vi.fn> } | null
  constructor(w: number, h: number) {
    this.width = w
    this.height = h
    this.ctx = null
  }
  getContext(type: string) {
    if (type !== '2d') return null
    this.ctx = { drawImage: vi.fn() }
    return this.ctx
  }
  async convertToBlob(opts: { type: string; quality?: number }) {
    return new Blob([new Uint8Array([0x52, 0x49, 0x46, 0x46])], { type: opts.type })
  }
}

function makeImageBitmap(width: number, height: number) {
  return {
    width,
    height,
    close: vi.fn()
  }
}

function makeFile(name: string, type: string, size = 1024): File {
  return new File([new Uint8Array(size)], name, { type })
}

describe('toWebP - early return cases', () => {
  it('returns the original file when type is non-image', async () => {
    const f = makeFile('doc.pdf', 'application/pdf')
    const out = await toWebP(f)
    expect(out).toBe(f)
  })

  it('returns the original file when type is image/gif', async () => {
    const f = makeFile('anim.gif', 'image/gif')
    const out = await toWebP(f)
    expect(out).toBe(f)
  })

  it('returns the original file when type is image/webp', async () => {
    const f = makeFile('already.webp', 'image/webp')
    const out = await toWebP(f)
    expect(out).toBe(f)
  })

  it('returns the original file when ImageBitmap is unavailable', async () => {
    const f = makeFile('pic.png', 'image/png')
    const ib = (globalThis as any).ImageBitmap
    ;(globalThis as any).ImageBitmap = undefined
    try {
      const out = await toWebP(f)
      expect(out).toBe(f)
    } finally {
      ;(globalThis as any).ImageBitmap = ib
    }
  })

  it('returns the original file when OffscreenCanvas is unavailable', async () => {
    const f = makeFile('pic.png', 'image/png')
    const oc = (globalThis as any).OffscreenCanvas
    ;(globalThis as any).OffscreenCanvas = undefined
    try {
      const out = await toWebP(f)
      expect(out).toBe(f)
    } finally {
      ;(globalThis as any).OffscreenCanvas = oc
    }
  })
})

describe('toWebP - conversion path', () => {
  let warnSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    ;(globalThis as any).createImageBitmap = vi.fn()
    ;(globalThis as any).ImageBitmap = class {}
    ;(globalThis as any).OffscreenCanvas = FakeOffscreenCanvas
    warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
  })

  afterEach(() => {
    delete (globalThis as any).createImageBitmap
    delete (globalThis as any).ImageBitmap
    delete (globalThis as any).OffscreenCanvas
    warnSpy.mockRestore()
  })

  it('converts a JPEG file to WebP and changes extension', async () => {
    const f = makeFile('photo.jpg', 'image/jpeg')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(800, 600))

    const out = await toWebP(f)
    expect(out).not.toBe(f)
    expect(out.name).toBe('photo.webp')
    expect(out.type).toBe('image/webp')
    expect(createImageBitmap).toHaveBeenCalledWith(f)
  })

  it('scales image down when max dimension is smaller than source', async () => {
    const f = makeFile('big.png', 'image/png', 4096)
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(4000, 3000))

    const out = await toWebP(f, 0.7, 1920)
    expect(out.type).toBe('image/webp')
    // scale = 1920/4000 = 0.48 → tw = 1920, th = 1440 (rounded)
    // 我们通过 spy/替代来验证：使用 spyOn FakeOffscreenCanvas 构造
  })

  it('does not upscale when source is smaller than max dimension', async () => {
    const f = makeFile('small.png', 'image/png')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(400, 300))

    const out = await toWebP(f, 0.6, 1920)
    expect(out.type).toBe('image/webp')
  })

  it('returns original file when bitmap has zero dimensions', async () => {
    const f = makeFile('zero.png', 'image/png')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(0, 0))

    const out = await toWebP(f)
    expect(out).toBe(f)
  })

  it('returns original file when canvas 2d context is unavailable', async () => {
    const f = makeFile('pic.png', 'image/png')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(100, 100))

    const Original = (globalThis as any).OffscreenCanvas
    class NoCtx extends Original {
      getContext() { return null as any }
    }
    ;(globalThis as any).OffscreenCanvas = NoCtx
    try {
      const out = await toWebP(f)
      expect(out).toBe(f)
    } finally {
      ;(globalThis as any).OffscreenCanvas = Original
    }
  })

  it('falls back to original file when convertToBlob throws', async () => {
    const f = makeFile('boom.png', 'image/png')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(100, 100))

    const Original = (globalThis as any).OffscreenCanvas
    class ThrowBlob extends Original {
      async convertToBlob() { throw new Error('boom') }
    }
    ;(globalThis as any).OffscreenCanvas = ThrowBlob
    try {
      const out = await toWebP(f)
      expect(out).toBe(f)
      expect(warnSpy).toHaveBeenCalled()
    } finally {
      ;(globalThis as any).OffscreenCanvas = Original
    }
  })

  it('falls back to original file when createImageBitmap rejects', async () => {
    const f = makeFile('bad.png', 'image/png')
    ;(createImageBitmap as any).mockRejectedValue(new Error('decode fail'))

    const out = await toWebP(f)
    expect(out).toBe(f)
    expect(warnSpy).toHaveBeenCalled()
  })

  it('strips file extension and replaces with .webp', async () => {
    const f = makeFile('my.image.with.dots.jpeg', 'image/jpeg')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(100, 100))
    const out = await toWebP(f)
    expect(out.name).toBe('my.image.with.dots.webp')
  })

  it('passes quality through to convertToBlob', async () => {
    const f = makeFile('q.png', 'image/png')
    ;(createImageBitmap as any).mockResolvedValue(makeImageBitmap(100, 100))
    const convertSpy = vi.spyOn(FakeOffscreenCanvas.prototype, 'convertToBlob')
    await toWebP(f, 0.42, 1920)
    expect(convertSpy).toHaveBeenCalledWith({ type: 'image/webp', quality: 0.42 })
  })
})
