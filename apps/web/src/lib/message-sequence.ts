import { assignsTurn } from '@/lib/chat-display'
import type { ChatMessage } from '@/types/chat'

/** Chronological speech index map (moderator / participant / user). */
export function buildMessageSequenceMap(messages: ChatMessage[]): Map<string, number> {
  const sorted = [...messages].sort((a, b) => a.createdAt - b.createdAt)
  let next = 1
  const map = new Map<string, number>()

  for (const message of sorted) {
    if (!assignsTurn(message.role)) continue
    if (message.turn != null) {
      map.set(message.id, message.turn)
      next = Math.max(next, message.turn + 1)
      continue
    }
    map.set(message.id, next)
    next += 1
  }

  return map
}

export function messageSequenceNumber(
  message: ChatMessage,
  sequenceMap: Map<string, number>,
): number | null {
  if (message.turn != null) return message.turn
  return sequenceMap.get(message.id) ?? null
}

export function formatMessageSequence(
  message: ChatMessage,
  sequenceMap: Map<string, number>,
): string | null {
  const n = messageSequenceNumber(message, sequenceMap)
  return n != null ? `#${n}` : null
}
