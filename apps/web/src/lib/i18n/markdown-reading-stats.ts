import { getTranslator } from '@/lib/i18n'
import { localeIntlTag } from '@/lib/i18n/translate'
import type { AppLocale } from '@/lib/locale'

const CHARS_PER_MINUTE = 400

export function markdownCharCount(content: string): number {
  return [...content].length
}

export function markdownReadingMinutes(content: string): number {
  const chars = markdownCharCount(content)
  if (chars === 0) return 0
  return Math.max(1, Math.ceil(chars / CHARS_PER_MINUTE))
}

export function formatMarkdownReadingStats(locale: AppLocale, content: string): string {
  const t = getTranslator(locale)
  const chars = markdownCharCount(content)
  if (chars === 0) return t('common.markdown.readingStatsEmpty')
  const minutes = markdownReadingMinutes(content)
  return t('common.markdown.readingStats', {
    chars: chars.toLocaleString(localeIntlTag(locale)),
    minutes,
  })
}
