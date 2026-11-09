import { onMounted, onUnmounted } from 'vue'

export function useTabsVideoAlign() {
  let ro: ResizeObserver | null = null
  let applied = ''

  const apply = () => {
    if (import.meta.server) return
    const tabs = document.querySelector('.tabs-main-container')
    const main = document.querySelector('.main-section')
    if (!tabs || !main) return
    const parent = main.parentElement
    if (!parent) return

    const tabsRect = tabs.getBoundingClientRect()
    const parentRect = parent.getBoundingClientRect()
    const tabsPad = parseFloat(getComputedStyle(tabs).paddingLeft) || 0
    const tabsW = tabsRect.width

    // 目标：让 .main-section 与 .tabs-main-container 同宽、同左侧、
    // 同水平 padding ⇒ 内容(视频网格)左右边界与分类栏内容完全对齐。
    const targetMargin = tabsRect.left - parentRect.left
    const key = `${targetMargin},${tabsW},${tabsPad}`
    if (applied === key) return

    const mStyle = main.style
    if (Math.abs(targetMargin) < 0.5 && Math.abs(tabsW) < 0.5) {
      if (applied !== '') {
        mStyle.width = ''
        mStyle.maxWidth = ''
        mStyle.marginLeft = ''
        mStyle.marginRight = ''
        mStyle.paddingLeft = ''
        applied = ''
      }
      return
    }

    mStyle.width = `${tabsW}px`
    mStyle.maxWidth = `${tabsW}px`
    mStyle.marginLeft = `${targetMargin}px`
    mStyle.marginRight = 'auto'
    mStyle.paddingLeft = `${tabsPad}px`
    mStyle.paddingRight = `${tabsPad}px`
    applied = key
  }

  const onResize = () => apply()

  onMounted(() => {
    requestAnimationFrame(apply)
    window.addEventListener('resize', onResize)
    if (typeof ResizeObserver !== 'undefined') {
      ro = new ResizeObserver(onResize)
      ro.observe(document.documentElement)
    }
  })
  onUnmounted(() => {
    window.removeEventListener('resize', onResize)
    ro?.disconnect()
    ro = null
  })

  return { apply }
}