import { speakerId } from '@/lib/chat-display'
import type { ChatMessage } from '@/types/chat'

export interface TranscriptProjection {
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  activeMessage: ChatMessage | null
}

/** Project Live state at a given turn (null = latest). */
export function projectTranscriptAtTurn(
  turns: ChatMessage[],
  scrubTurn: number | null,
): TranscriptProjection {
  const visible =
    scrubTurn != null ? turns.filter((m) => (m.turn ?? 0) <= scrubTurn) : turns

  const latestBySeat = new Map<string, ChatMessage>()
  let activeSpeakerId: string | null = null
  let activeMessage: ChatMessage | null = null

  for (const message of visible) {
    const seat = speakerId(message)
    latestBySeat.set(seat, message)
    activeSpeakerId = seat
    activeMessage = message
  }

  return { latestBySeat, activeSpeakerId, activeMessage }
}

export function maxTurnNumber(turns: ChatMessage[]): number {
  if (turns.length === 0) return 0
  return Math.max(...turns.map((m) => m.turn ?? 0))
}
