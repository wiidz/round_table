import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'

export interface PrincipalUserPreset {
  label: string
  value: string
}

export function getPrincipalUserPresets(locale: AppLocale) {
  const t = getTranslator(locale)
  return {
    confirmation: [
      {
        label: t('profile.principal.presets.confirmation.numbered.label'),
        value: t('profile.principal.presets.confirmation.numbered.value'),
      },
      {
        label: t('profile.principal.presets.confirmation.brief.label'),
        value: t('profile.principal.presets.confirmation.brief.value'),
      },
      {
        label: t('profile.principal.presets.confirmation.owners.label'),
        value: t('profile.principal.presets.confirmation.owners.value'),
      },
    ],
    context: [
      {
        label: t('profile.principal.presets.context.mobileGame.label'),
        value: t('profile.principal.presets.context.mobileGame.value'),
      },
      {
        label: t('profile.principal.presets.context.indie.label'),
        value: t('profile.principal.presets.context.indie.value'),
      },
      {
        label: t('profile.principal.presets.context.b2b.label'),
        value: t('profile.principal.presets.context.b2b.value'),
      },
    ],
  } satisfies Record<string, PrincipalUserPreset[]>
}

export function applyPrincipalFieldPreset(current: string, snippet: string): string {
  const next = snippet.trim()
  if (!next) return current
  const trimmed = current.trim()
  if (!trimmed) return next
  if (trimmed.includes(next)) return trimmed
  if (!trimmed.includes('\n')) {
    return `${trimmed}; ${next}`
  }
  return `${trimmed}\n${next}`
}

export function principalPresetApplied(current: string, snippet: string): boolean {
  const next = snippet.trim()
  if (!next) return false
  return current.trim().includes(next)
}
