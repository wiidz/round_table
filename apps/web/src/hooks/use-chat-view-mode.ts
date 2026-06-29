import { useMemo, useState } from 'react'

import {
  inferChatMeetingPhase,
  suggestedViewMode,
  type ChatMeetingPhase,
} from '@/lib/chat-meeting-phase'
import type { ChatMessage } from '@/types/chat'

export type ChatViewMode = 'list' | 'roundtable'

export function useChatViewMode(messages: ChatMessage[]) {
  const phase = useMemo(() => inferChatMeetingPhase(messages), [messages])
  const suggested = useMemo(() => suggestedViewMode(phase), [phase])
  const [override, setOverride] = useState<ChatViewMode | null>(null)

  const mode = override ?? suggested

  return {
    mode,
    phase,
    suggested,
    setMode: setOverride,
    clearOverride: () => setOverride(null),
  }
}

export type { ChatMeetingPhase }
