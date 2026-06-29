import { useMemo } from 'react'

import { projectTranscriptAtTurn } from '@/lib/meeting-transcript-projection'
import type { ChatMessage } from '@/types/chat'

export interface MeetingTranscriptState {
  /** Messages with turn, sorted ascending. */
  turns: ChatMessage[]
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  /** Message for the active turn (live or scrubbed). */
  activeMessage: ChatMessage | null
  nextTurn: number
  /** True when viewing historical turn rather than live tail. */
  isScrubbing: boolean
}

export function useMeetingTranscript(
  messages: ChatMessage[],
  scrubTurn: number | null = null,
): MeetingTranscriptState {
  return useMemo(() => {
    const turns = messages
      .filter((m) => m.turn != null)
      .sort((a, b) => (a.turn ?? 0) - (b.turn ?? 0))

    const { latestBySeat, activeSpeakerId, activeMessage } = projectTranscriptAtTurn(
      turns,
      scrubTurn,
    )

    const nextTurn =
      turns.length > 0 ? Math.max(...turns.map((m) => m.turn ?? 0)) + 1 : 1

    return {
      turns,
      latestBySeat,
      activeSpeakerId,
      activeMessage,
      nextTurn,
      isScrubbing: scrubTurn != null,
    }
  }, [messages, scrubTurn])
}
