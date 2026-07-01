import type { AppLocale } from '@/lib/locale'
import { en } from '@/lib/i18n/messages/en'
import { zh } from '@/lib/i18n/messages/zh'
import { createTranslator, type MessageTree, type Translator } from '@/lib/i18n/translate'

const MESSAGES: Record<AppLocale, MessageTree> = { zh, en }

export function getTranslator(locale: AppLocale): Translator {
  return createTranslator(MESSAGES[locale])
}

export { MESSAGES }
