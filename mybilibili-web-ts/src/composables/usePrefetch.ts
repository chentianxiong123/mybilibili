import { useRouter } from 'vue-router'

export function usePrefetch() {
  const router = useRouter()
  const loaded = new Set<string>()

  const loadedKey = (target: any) => {
    if (typeof target === 'string') return `path:${target}`
    if (typeof target === 'object') return `${target.name}:${JSON.stringify(target.params || {})}`
    return ''
  }

  function prefetch(target: any) {
    const key = loadedKey(target)
    if (!key || loaded.has(key)) return
    loaded.add(key)
    const route = router.resolve(target)
    const comps = route.matched.flatMap(m => {
      const c = m.components?.default
      return typeof c === 'function' ? [c] : []
    }) as Array<() => Promise<any>>
    comps.forEach(comp => comp().catch(() => {}))
  }

  return { prefetch }
}