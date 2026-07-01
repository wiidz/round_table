import {
  messageAvatar as messageAvatarI18n,
  messageLabel as messageLabelI18n,
} from '@/lib/i18n/chat-display'
import type { AppLocale } from '@/lib/locale'
import type { ChatMessage, ChatRole } from '@/types/chat'

const FALLBACK_LOCALE: AppLocale = 'zh'

/** Stable seat key for Live projection and transcript grouping. */
export function speakerId(message: ChatMessage): string {
  if (message.role === 'moderator') return 'moderator'
  if (message.role === 'participant') {
    return message.authorId?.trim() || 'participant'
  }
  if (message.role === 'user') return 'user'
  return 'system'
}

/** Non-React helper; defaults to zh. Prefer useI18n().messageLabel in components. */
export function messageLabel(message: ChatMessage, locale: AppLocale = FALLBACK_LOCALE): string {
  return messageLabelI18n(locale, message)
}

/** Non-React helper; defaults to zh. Prefer useI18n().messageAvatar in components. */
export function messageAvatar(
  message: ChatMessage,
  locale: AppLocale = FALLBACK_LOCALE,
): { id: string; name: string } {
  return messageAvatarI18n(locale, message)
}

export function assignsTurn(role: ChatRole): boolean {
  return role === 'moderator' || role === 'participant' || role === 'user'
}
