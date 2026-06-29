import { assignsTurn } from '@/lib/chat-display'
import type { ChatRole } from '@/types/chat'

export interface TurnAssignment {
  turn?: number
  nextTurn: number
}

/** Assign global turn for moderator/participant messages (ADR-0013). */
export function assignTurnForRole(role: ChatRole, nextTurn: number): TurnAssignment {
  if (!assignsTurn(role)) {
    return { nextTurn }
  }
  return { turn: nextTurn, nextTurn: nextTurn + 1 }
}
