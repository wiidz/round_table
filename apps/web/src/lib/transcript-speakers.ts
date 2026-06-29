import { messageLabel, speakerId } from '@/lib/chat-display'
import type { ChatMessage } from '@/types/chat'

export interface TranscriptSpeaker {
  id: string
  label: string
}

/** Unique speakers in message order; excludes system. */
export function listTranscriptSpeakers(messages: ChatMessage[]): TranscriptSpeaker[] {
  const seen = new Map<string, string>()
  for (const message of messages) {
    const id = speakerId(message)
    if (id === 'system') continue
    if (!seen.has(id)) {
      seen.set(id, messageLabel(message))
    }
  }
  return [...seen.entries()].map(([id, label]) => ({ id, label }))
}

export function filterTranscriptBySpeaker(
  messages: ChatMessage[],
  speakerFilterId: string | null,
): ChatMessage[] {
  if (!speakerFilterId) return messages
  return messages.filter((message) => speakerId(message) === speakerFilterId)
}
