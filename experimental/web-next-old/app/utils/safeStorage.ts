const isClient = typeof window !== 'undefined'

export const safeStorage = {
  getItem(key: string): string | null {
    return isClient ? localStorage.getItem(key) : null
  },
  setItem(key: string, value: string): void {
    if (isClient) localStorage.setItem(key, value)
  },
  removeItem(key: string): void {
    if (isClient) localStorage.removeItem(key)
  }
}