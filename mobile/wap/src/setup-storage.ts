// Polyfill localStorage for happy-dom environment
// Node 26+ requires --localstorage-file for native localStorage

if (typeof globalThis.localStorage === 'undefined' || globalThis.localStorage === null) {
  const store = new Map<string, string>()

  const localStorage = {
    getItem(key: string): string | null {
      return store.get(key) ?? null
    },
    setItem(key: string, value: string): void {
      store.set(key, String(value))
    },
    removeItem(key: string): void {
      store.delete(key)
    },
    clear(): void {
      store.clear()
    },
    get length(): number {
      return store.size
    },
    key(index: number): string | null {
      return [...store.keys()][index] ?? null
    }
  }

  Object.defineProperty(globalThis, 'localStorage', {
    value: localStorage,
    writable: true,
    configurable: true
  })
}
