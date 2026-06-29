import { useMemo } from 'react'

import { speakerId } from '@/lib/chat-display'
import type { ChatMessage } from '@/types/chat'

export interface MeetingTranscriptState {
  /** Messages with turn, sorted ascending. */
  turns: ChatMessage[]
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  nextTurn: number
}

export function useMeetingTranscript(messages: ChatMessage[]): MeetingTranscriptState {
  return useMemo(() => {
    const turns = messages
      .filter((m) => m.turn != null)
      .sort((a, b) => (a.turn ?? 0) - (b.turn ?? 0))

    const latestBySeat = new Map<string, ChatMessage>()
    let activeSpeakerId: string | null = null

    for (const message of turns) {
      const seat = speakerId(message)
      latestBySeat.set(seat, message)
      activeSpeakerId = seat
    }

    const nextTurn =
      turns.length > 0 ? Math.max(...turns.map((m) => m.turn ?? 0)) + 1 : 1

    return { turns, latestBySeat, activeSpeakerId, nextTurn }
  }, [messages])
}
