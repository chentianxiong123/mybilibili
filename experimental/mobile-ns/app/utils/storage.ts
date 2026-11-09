import { ApplicationSettings } from '@nativescript/core'

const storage = {
  getItem(key: string): string | null {
    return ApplicationSettings.getString(key)
  },

  setItem(key: string, value: string): void {
    ApplicationSettings.setString(key, value)
  },

  removeItem(key: string): void {
    ApplicationSettings.remove(key)
  },

  getObject<T>(key: string): T | null {
    const val = ApplicationSettings.getString(key)
    if (!val) return null
    try {
      return JSON.parse(val) as T
    } catch {
      return null
    }
  },

  setObject(key: string, value: any): void {
    ApplicationSettings.setString(key, JSON.stringify(value))
  },

  getToken(): string | null {
    return this.getItem('token')
  },

  setToken(token: string): void {
    this.setItem('token', token)
  },

  removeToken(): void {
    this.removeItem('token')
  },

  getUser<T>(): T | null {
    return this.getObject<T>('user')
  },

  setUser(user: any): void {
    this.setObject('user', user)
  },

  setViewHistory(item: { aId: number; title: string; pic: string; viewAt: number }): void {
    const history = this.getObject<any[]>('view_history') || []
    const idx = history.findIndex(h => h.aId === item.aId)
    if (idx >= 0) history.splice(idx, 1)
    history.unshift(item)
    if (history.length > 100) history.pop()
    this.setObject('view_history', history)
  },

  getViewHistory(): any[] {
    return this.getObject<any[]>('view_history') || []
  }
}

export default storage