import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'

const PROFILE_FILE_KEYS: Record<string, string> = {
  'USER.md': 'profile.files.USER.md',
  'SOUL.md': 'profile.files.SOUL.md',
  'AGENTS.md': 'profile.files.AGENTS.md',
  'TOOLS.md': 'profile.files.TOOLS.md',
}

export function profileFileLabel(locale: AppLocale, filename: string): string {
  const key = PROFILE_FILE_KEYS[filename]
  if (!key) return filename
  return getTranslator(locale)(key)
}

export function profileFileCaption(locale: AppLocale, filename: string): string {
  const title = profileFileLabel(locale, filename)
  if (title === filename) return filename
  return `${title} · ${filename}`
}

export function profileFileHasTitle(filename: string): boolean {
  return filename in PROFILE_FILE_KEYS
}

export const PARTICIPANT_STANDARD_FILES = ['SOUL.md', 'AGENTS.md', 'TOOLS.md'] as const
export const PRINCIPAL_STANDARD_FILES = ['USER.md'] as const

export function participantFileHint(locale: AppLocale, filename: string): string {
  const key = `profile.fileHints.${filename}`
  const translated = getTranslator(locale)(key)
  return translated === key ? '' : translated
}

export function principalFileHint(locale: AppLocale, filename: string): string {
  const key = `profile.fileHints.${filename}`
  const translated = getTranslator(locale)(key)
  return translated === key ? '' : translated
}
