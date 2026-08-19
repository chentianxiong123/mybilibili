import { onMounted, onUnmounted } from 'vue'

export function useTabsVideoAlign() {
  let ro: ResizeObserver | null = null
  let applied = 0

  const apply = () => {
    if (import.meta.server) return
    const tabs = document.querySelector('.tabs-main-container')
    const main = document.querySelector('.main-section')
    if (!tabs || !main) return
    const parent = main.parentElement
    if (!parent) return
    const tabsRect = tabs.getBoundingClientRect()
    const parentRect = parent.getBoundingClientRect()
    // 目标：main-section 左缘与分类栏内容容器左缘对齐。
    // margin-left 相对父容器（.layout-content，与 .layout-home 同左缘），
    // 因此目标 margin = tabs 相对父容器内容框左缘的距离。
    const targetMargin = tabsRect.left - parentRect.left
    if (Math.abs(targetMargin - applied) < 0.5) return
    if (Math.abs(targetMargin) < 0.5) {
      applied = 0
      main.style.marginLeft = ''
      main.style.marginRight = ''
      main.style.width = ''
      return
    }
    // .main-section CSS 为 width:100% + margin:0 auto,max-width:1980;
    // 覆盖为 width:auto 后 margin-left 生效、右缘由 max-width 约束不外溢。
    main.style.width = 'auto'
    main.style.marginLeft = `${targetMargin}px`
    main.style.marginRight = 'auto'
    applied = targetMargin
  }

  const onResize = () => apply()

  onMounted(() => {
    requestAnimationFrame(apply)
    window.addEventListener('resize', onResize)
    if (typeof ResizeObserver !== 'undefined') {
      ro = new ResizeObserver(onResize)
      const root = document.documentElement
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