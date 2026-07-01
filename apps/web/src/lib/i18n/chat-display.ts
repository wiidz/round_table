import { getTranslator } from '@/lib/i18n'
import type { AppLocale } from '@/lib/locale'
import type { ChatMessage } from '@/types/chat'

export function messageLabel(locale: AppLocale, message: ChatMessage): string {
  const t = getTranslator(locale)
  if (message.role === 'user') return t('chat.labels.me')
  if (message.role === 'system') return t('chat.labels.system')
  if (message.role === 'participant') {
    return message.authorName?.trim() || message.authorId || t('chat.labels.participant')
  }
  return message.authorName?.trim() || t('chat.labels.moderator')
}

export function messageAvatar(
  locale: AppLocale,
  message: ChatMessage,
): { id: string; name: string } {
  const t = getTranslator(locale)
  if (message.role === 'user') return { id: 'user', name: t('chat.labels.me') }
  if (message.role === 'system') return { id: 'system', name: t('chat.labels.system') }
  if (message.role === 'participant') {
    return {
      id: message.authorId?.trim() || 'participant',
      name: message.authorName?.trim() || message.authorId || t('chat.labels.participant'),
    }
  }
  return { id: 'moderator', name: message.authorName?.trim() || t('chat.labels.moderator') }
}
