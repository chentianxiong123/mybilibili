import { describe, it, expect } from 'vitest'
import { getPicSuffix } from '../image'

describe('getPicSuffix', () => {
  it('返回 .webp 或 .jpg（取决于浏览器支持）', () => {
    const result = getPicSuffix()
    expect(['.webp', '.jpg']).toContain(result)
  })
})
