import { useEffect, useMemo, useRef } from 'react'

import { ChatBubble } from '@/components/chat/chat-bubble'
import { useI18n } from '@/hooks/use-i18n'
import { buildMessageSequenceMap, messageSequenceNumber } from '@/lib/message-sequence'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface ImTranscriptViewProps {
  messages: ChatMessage[]
  className?: string
}

export function ImTranscriptView({ messages, className }: ImTranscriptViewProps) {
  const { t } = useI18n()
  const scrollRef = useRef<HTMLDivElement>(null)
  const sequenceMap = useMemo(() => buildMessageSequenceMap(messages), [messages])

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [messages])

  return (
    <div
      ref={scrollRef}
      className={cn(
        'min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain px-5 py-5',
        className,
      )}
    >
      {messages.length === 0 && (
        <p className="text-center text-[13px] text-text-tertiary">
          {t('chat.transcript.emptyHint')}
        </p>
      )}
      {messages.map((message) => (
        <ChatBubble
          key={message.id}
          message={message}
          sequence={messageSequenceNumber(message, sequenceMap)}
        />
      ))}
    </div>
  )
}
