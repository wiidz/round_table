import type { Translator } from '@/lib/i18n/translate'
import type { SettingsFieldState } from '@/types/settings'

function translateOrFallback(t: Translator, key: string, fallback: string): string {
  const translated = t(key)
  return translated === key ? fallback : translated
}

export function settingsFieldLabel(t: Translator, field: SettingsFieldState): string {
  return translateOrFallback(t, `settings.fields.${field.key}.label`, field.label)
}

export function settingsFieldDescription(t: Translator, field: SettingsFieldState): string | undefined {
  const key = `settings.fields.${field.key}.description`
  const translated = t(key)
  if (translated === key) return field.description
  return translated
}

export function settingsSectionTitle(t: Translator, section: string): string {
  if (section === '设定上限') {
    return t('pages.settings.sections.limits')
  }
  return section
}
