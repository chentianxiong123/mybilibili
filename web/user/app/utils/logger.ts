/**
 * 统一前端日志工具。
 *
 * 级别(从低到高): debug < info < warn < error。
 *   - 默认级别: 开发(import.meta.env.DEV)= debug，生产 = warn（只留警告/错误）。
 *   - 可通过环境变量 NUXT_PUBLIC_LOG_LEVEL 覆盖（部署时可配）。
 *   - 也可在运行时用 logger.setLevel() 调整（如控制台调试）。
 *
 * 所有方法最终转发到原生 console，但按级别过滤，避免开发期高频调试日志刷屏。
 * 业务代码请统一 import { logger } from '@/utils/logger'，不要直接散落 console.log。
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const LEVEL_ORDER: Record<LogLevel, number> = { debug: 0, info: 1, warn: 2, error: 3 }

function envLevel(): LogLevel {
  const fromEnv = typeof process !== 'undefined' ? process.env?.NUXT_PUBLIC_LOG_LEVEL : undefined
  const fromVite = (import.meta as any)?.env?.NUXT_PUBLIC_LOG_LEVEL as string | undefined
  const raw = fromEnv || fromVite
  if (raw && raw in LEVEL_ORDER) return raw as LogLevel
  if (typeof window !== 'undefined' && (import.meta as any)?.env?.DEV) return 'debug'
  return 'warn'
}

let currentLevel: LogLevel = envLevel()

function write(method: 'debug' | 'info' | 'warn' | 'error', args: unknown[]) {
  const fn = console[method]
  if (typeof fn === 'function') {
    fn(`[${method}]`, ...args)
  }
}

function enabled(level: LogLevel): boolean {
  return LEVEL_ORDER[level] >= LEVEL_ORDER[currentLevel]
}

export const logger = {
  setLevel(level: LogLevel) {
    currentLevel = level
  },
  getLevel(): LogLevel {
    return currentLevel
  },
  debug(...args: unknown[]) {
    if (enabled('debug')) write('debug', args)
  },
  info(...args: unknown[]) {
    if (enabled('info')) write('info', args)
  },
  warn(...args: unknown[]) {
    if (enabled('warn')) write('warn', args)
  },
  error(...args: unknown[]) {
    if (enabled('error')) write('error', args)
  },
}

export default logger
