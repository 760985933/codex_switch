import { defineStore } from 'pinia'
import { detectInitialLocale, normalizeLocale, type SupportedLocale } from '../i18n'

export const useUiStore = defineStore('ui', {
  state: () => ({
    showSettings: false,
    locale: detectInitialLocale() as SupportedLocale,
  }),
  actions: {
    setLocale(value: string) {
      const next = normalizeLocale(value)
      this.locale = next
      if (typeof window !== 'undefined') {
        window.localStorage.setItem('ui.locale', next)
      }
    },
  },
})
