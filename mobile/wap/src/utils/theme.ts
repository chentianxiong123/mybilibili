// 键前缀必须与 utils/storage_layer.ts 的 PREFIX('wap:') 一致：
// clearAll() 只清 `wap:*`，用 `wap-theme` 会清不掉。
const THEME_KEY = 'wap:theme'
// 升级前的旧键（连字符），读到就迁移一次再删。
const LEGACY_THEME_KEY = 'wap-theme'

export type WapTheme = 'light' | 'dark'

export function getWapTheme(): WapTheme {
  const current = localStorage.getItem(THEME_KEY)
  if (current === 'dark' || current === 'light') return current

  const legacy = localStorage.getItem(LEGACY_THEME_KEY)
  if (legacy === 'dark' || legacy === 'light') {
    localStorage.setItem(THEME_KEY, legacy)
    localStorage.removeItem(LEGACY_THEME_KEY)
    return legacy
  }
  return 'light'
}

export function applyWapTheme(theme: WapTheme) {
  document.documentElement.dataset.wapTheme = theme
  localStorage.setItem(THEME_KEY, theme)
}

export function initWapTheme() {
  applyWapTheme(getWapTheme())
}

export function toggleWapTheme(): WapTheme {
  const next = getWapTheme() === 'dark' ? 'light' : 'dark'
  applyWapTheme(next)
  return next
}
