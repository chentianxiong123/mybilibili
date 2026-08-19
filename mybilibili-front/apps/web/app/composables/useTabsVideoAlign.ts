import { onMounted, onUnmounted } from 'vue'

export function useTabsVideoAlign() {
  let ro: ResizeObserver | null = null
  let applied = 0

  const apply = () => {
    if (import.meta.server) return
    const tabs = document.querySelector('.tabs-main-container')
    const main = document.querySelector('.main-section')
    if (!tabs || !main) return
    // 必须以"未变换"位置为准：先清掉自身 transform 再测量
    if (main.style.transform) main.style.transform = ''
    const tabsRect = tabs.getBoundingClientRect()
    const mainRect = main.getBoundingClientRect()
    const delta = tabsRect.left - mainRect.left
    if (Math.abs(delta) < 0.5) {
      if (applied !== 0) {
        applied = 0
        main.style.transform = ''
      }
      return
    }
    main.style.transform = `translateX(${delta}px)`
    applied = delta
  }

  const onResize = () => apply()

  onMounted(() => {
    requestAnimationFrame(apply)
    window.addEventListener('resize', onResize)
    if (typeof ResizeObserver !== 'undefined') {
      ro = new ResizeObserver(onResize)
      const root = document.querySelector('.layout-home') || document.documentElement
      ro.observe(root)
    }
  })
  onUnmounted(() => {
    window.removeEventListener('resize', onResize)
    ro?.disconnect()
    ro = null
  })

  return { apply }
}