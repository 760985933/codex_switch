import { defineStore } from 'pinia'
import { detectInitialLocale, normalizeLocale, type SupportedLocale } from '../i18n'

function loadBoolean(key: string, fallback: boolean): boolean {
  if (typeof window === 'undefined') return fallback
  const value = window.localStorage.getItem(key)
  return value !== null ? value === 'true' : fallback
}

function loadNumber(key: string, fallback: number): number {
  if (typeof window === 'undefined') return fallback
  const value = window.localStorage.getItem(key)
  return value !== null ? Number(value) : fallback
}

export const useUiStore = defineStore('ui', {
  state: () => ({
    showSettings: false,
    showHelp: false,
    locale: detectInitialLocale() as SupportedLocale,
    sidebarCollapsed: loadBoolean('ui.sidebarCollapsed', false),
    sidebarWidth: loadNumber('ui.sidebarWidth', 220),
  }),
  actions: {
    setLocale(value: string) {
      const next = normalizeLocale(value)
      this.locale = next
      if (typeof window !== 'undefined') {
        window.localStorage.setItem('ui.locale', next)
      }
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      if (typeof window !== 'undefined') {
        window.localStorage.setItem('ui.sidebarCollapsed', String(this.sidebarCollapsed))
      }
    },
    setSidebarWidth(width: number) {
      this.sidebarWidth = width
      if (typeof window !== 'undefined') {
        window.localStorage.setItem('ui.sidebarWidth', String(width))
      }
    },
  },
})
