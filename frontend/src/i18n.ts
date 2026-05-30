import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'
import jaJP from './locales/ja-JP'
import koKR from './locales/ko-KR'
import frFR from './locales/fr-FR'
import deDE from './locales/de-DE'
import esAR from './locales/es-AR'

export const SUPPORTED_LOCALES = [
  'zh-CN',
  'en-US',
  'ja-JP',
  'ko-KR',
  'fr-FR',
  'de-DE',
  'es-AR',
] as const

export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]

const STORAGE_KEY = 'ui.locale'

export function normalizeLocale(value: string): SupportedLocale {
  if (!value) return 'zh-CN'
  const lowered = value.toLowerCase()

  if (lowered.startsWith('zh')) return 'zh-CN'
  if (lowered.startsWith('en')) return 'en-US'
  if (lowered.startsWith('ja')) return 'ja-JP'
  if (lowered.startsWith('ko')) return 'ko-KR'
  if (lowered.startsWith('fr')) return 'fr-FR'
  if (lowered.startsWith('de')) return 'de-DE'
  if (lowered.startsWith('es')) return 'es-AR'

  return 'en-US'
}

export function detectInitialLocale(): SupportedLocale {
  if (typeof window === 'undefined') return 'zh-CN'
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored) return normalizeLocale(stored)
  return normalizeLocale(window.navigator.language)
}

export const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
  'ja-JP': jaJP,
  'ko-KR': koKR,
  'fr-FR': frFR,
  'de-DE': deDE,
  'es-AR': esAR,
} as const

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectInitialLocale(),
  fallbackLocale: 'en-US',
  messages,
})
