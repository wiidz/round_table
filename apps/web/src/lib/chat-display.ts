import type { ChatMessage, ChatRole } from '@/types/chat'

/** Stable seat key for Live projection and transcript grouping. */
export function speakerId(message: ChatMessage): string {
  if (message.role === 'moderator') return 'moderator'
  if (message.role === 'participant') {
    return message.authorId?.trim() || 'participant'
  }
  if (message.role === 'user') return 'user'
  return 'system'
}

export function messageLabel(message: ChatMessage): string {
  if (message.role === 'user') return '我'
  if (message.role === 'system') return '系统'
  if (message.role === 'participant') {
    return message.authorName?.trim() || message.authorId || '专家'
  }
  return message.authorName?.trim() || '司仪'
}

export function assignsTurn(role: ChatRole): boolean {
  return role === 'moderator' || role === 'participant'
}
