const THEME_KEY = 'wap-theme'

export type WapTheme = 'light' | 'dark'

export function getWapTheme(): WapTheme {
  return localStorage.getItem(THEME_KEY) === 'dark' ? 'dark' : 'light'
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
