import type { LiveBubbleVariant } from '@/components/round-table/live-bubble'
import type { ChatMessage } from '@/types/chat'

/** Resolve bubble tier relative to the highlighted message turn. */
export function resolveLiveBubbleVariant(
  message: ChatMessage,
  highlightMessageId: string | null,
  referenceTurn: number | null,
): LiveBubbleVariant {
  if (highlightMessageId && message.id === highlightMessageId) return 'active'

  const turn = message.turn
  if (referenceTurn == null || turn == null) return 'before'
  if (turn < referenceTurn) return 'before'
  if (turn > referenceTurn) return 'after'
  return 'before'
}
