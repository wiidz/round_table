import type { SettingsFieldState } from '@/types/settings'

export type AppLocale = 'zh' | 'en'

export const LOCALE_SETTINGS_KEY = 'ROUND_TABLE_LOCALE'

const DEFAULT_LOCALE: AppLocale = 'zh'

/** Normalize server / env locale to Web UI locale. */
export function normalizeLocale(value: string | undefined | null): AppLocale {
  const v = (value ?? '').trim().toLowerCase()
  if (v === 'en' || v.startsWith('en-')) return 'en'
  return 'zh'
}

export function localeFromSettingsFields(fields: SettingsFieldState[]): AppLocale {
  const field = fields.find((f) => f.key === LOCALE_SETTINGS_KEY)
  return normalizeLocale(field?.value ?? field?.placeholder)
}

export function defaultLocale(): AppLocale {
  return DEFAULT_LOCALE
}

/** USER.md Language field value from app locale (settings ROUND_TABLE_LOCALE). */
export function localeToUserLanguage(locale: AppLocale): string {
  return locale === 'en' ? 'en' : 'zh-CN'
}
