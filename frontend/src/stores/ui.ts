import { defineStore } from 'pinia'
import { detectInitialLocale, normalizeLocale, type SupportedLocale } from '../i18n'
import type { SourceID } from '../types'

function loadBoolean(key: string, fallback: boolean): boolean {
  if (typeof window === 'undefined') return fallback
  const value = window.localStorage.getItem(key)
  return value !== null ? value === 'true' : fallback
}

export const useUiStore = defineStore('ui', {
  state: () => ({
    showSettings: false,
    showHelp: false,
    locale: detectInitialLocale() as SupportedLocale,
    sidebarCollapsed: loadBoolean('ui.sidebarCollapsed', false),
    settingsSource: 'codex' as SourceID,
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
    openSettings(source: SourceID = 'codex') {
      this.settingsSource = source
      this.showSettings = true
    },
  },
})
