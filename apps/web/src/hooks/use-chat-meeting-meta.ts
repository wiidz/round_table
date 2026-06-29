import { useEffect, useMemo, useState } from 'react'

import { fetchMeeting } from '@/api/meetings'
import { parseMeetingIdFromMessages } from '@/lib/chat-meeting-phase'
import type { ChatMessage } from '@/types/chat'

export interface ChatMeetingMeta {
  meetingId: string | null
  topic: string | null
  loading: boolean
}

export function useChatMeetingMeta(messages: ChatMessage[]): ChatMeetingMeta {
  const meetingId = useMemo(() => parseMeetingIdFromMessages(messages), [messages])
  const [topic, setTopic] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!meetingId) {
      setTopic(null)
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    fetchMeeting(meetingId)
      .then((detail) => {
        if (!cancelled) setTopic(detail.topic?.trim() || null)
      })
      .catch(() => {
        if (!cancelled) setTopic(null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [meetingId])

  return { meetingId, topic, loading }
}
