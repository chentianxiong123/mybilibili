import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { logger } from './logger'

describe('logger', () => {
  beforeEach(() => {
    vi.spyOn(console, 'debug').mockImplementation(() => {})
    vi.spyOn(console, 'info').mockImplementation(() => {})
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('debug is silent when level is info', () => {
    logger.setLevel('info')
    logger.debug('should not print')
    expect(console.debug).not.toHaveBeenCalled()
  })

  it('info/warn/error print when level is info', () => {
    logger.setLevel('info')
    logger.info('a')
    logger.warn('b')
    logger.error('c')
    expect(console.info).toHaveBeenCalled()
    expect(console.warn).toHaveBeenCalled()
    expect(console.error).toHaveBeenCalled()
  })

  it('debug prints when level is debug', () => {
    logger.setLevel('debug')
    logger.debug('d')
    expect(console.debug).toHaveBeenCalled()
  })

  it('error still prints at error level, others silent', () => {
    logger.setLevel('error')
    logger.info('x')
    logger.error('y')
    expect(console.info).not.toHaveBeenCalled()
    expect(console.error).toHaveBeenCalled()
  })
})
