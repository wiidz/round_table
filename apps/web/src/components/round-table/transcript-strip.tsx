import { useEffect, useMemo, useRef, useState } from 'react'

import { condenseMessage } from '@/lib/condense-message'
import { messageLabel } from '@/lib/chat-display'
import { formatChatTime } from '@/lib/format-date'
import {
  buildMessageSequenceMap,
  messageSequenceNumber,
} from '@/lib/message-sequence'
import {
  filterTranscriptBySpeaker,
  listTranscriptSpeakers,
} from '@/lib/transcript-speakers'
import { cn } from '@/lib/utils'
import type { ChatMessage } from '@/types/chat'

interface TranscriptStripProps {
  messages: ChatMessage[]
  /** Latest turn message id (current speaker). */
  activeMessageId?: string | null
  selectedId?: string | null
  onSelect: (message: ChatMessage) => void
  /** Hide outer title row when nested in TranscriptHistoryPanel. */
  embedded?: boolean
  className?: string
}

export function TranscriptStrip({
  messages,
  activeMessageId,
  selectedId,
  onSelect,
  embedded = false,
  className,
}: TranscriptStripProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [pinnedBottom, setPinnedBottom] = useState(true)
  const [newBelow, setNewBelow] = useState(0)
  const [filterSpeakerId, setFilterSpeakerId] = useState<string | null>(null)

  const speakers = useMemo(() => listTranscriptSpeakers(messages), [messages])
  const sequenceMap = useMemo(() => buildMessageSequenceMap(messages), [messages])
  const visibleMessages = useMemo(
    () => filterTranscriptBySpeaker(messages, filterSpeakerId),
    [messages, filterSpeakerId],
  )

  useEffect(() => {
    if (filterSpeakerId && !speakers.some((s) => s.id === filterSpeakerId)) {
      setFilterSpeakerId(null)
    }
  }, [speakers, filterSpeakerId])

  useEffect(() => {
    const el = scrollRef.current
    if (!el || !pinnedBottom) return
    el.scrollTop = el.scrollHeight
    setNewBelow(0)
  }, [visibleMessages, pinnedBottom])

  const handleScroll = () => {
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24
    setPinnedBottom(atBottom)
    if (atBottom) setNewBelow(0)
  }

  useEffect(() => {
    if (pinnedBottom) return
    setNewBelow((n) => n + 1)
  }, [visibleMessages.length, pinnedBottom])

  const scrollToBottom = () => {
    const el = scrollRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
    setPinnedBottom(true)
    setNewBelow(0)
  }

  const showFilters = speakers.length > 1

  return (
    <div
      className={cn(
        'relative flex shrink-0 flex-col border-t border-black/[0.06] bg-black/[0.015]',
        className ?? 'h-36',
      )}
    >
      {( !embedded || showFilters) && (
      <div className={cn('flex shrink-0 flex-col gap-2 px-5 py-2', embedded && 'px-4 sm:px-5 pt-3')}>
        {!embedded && (
          <p className="text-[11px] font-medium uppercase tracking-[0.12em] text-text-tertiary">
            发言记录 · {filterSpeakerId ? `${visibleMessages.length}/${messages.length}` : messages.length} 条
          </p>
        )}
        {embedded && showFilters && (
          <p className="text-[11px] font-medium uppercase tracking-[0.12em] text-text-tertiary">
            筛选 · {filterSpeakerId ? `${visibleMessages.length}/${messages.length}` : messages.length} 条
          </p>
        )}
        {showFilters && (
          <div className="flex gap-1.5 overflow-x-auto pb-0.5 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            <button
              type="button"
              onClick={() => setFilterSpeakerId(null)}
              className={cn(
                'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
                filterSpeakerId == null
                  ? 'bg-brand-soft text-brand ring-1 ring-brand/25'
                  : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
              )}
            >
              全部
            </button>
            {speakers.map((speaker) => (
              <button
                key={speaker.id}
                type="button"
                onClick={() =>
                  setFilterSpeakerId((current) =>
                    current === speaker.id ? null : speaker.id,
                  )
                }
                className={cn(
                  'shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors',
                  filterSpeakerId === speaker.id
                    ? 'bg-brand-soft text-brand ring-1 ring-brand/25'
                    : 'bg-black/[0.04] text-text-tertiary hover:text-text-secondary',
                )}
              >
                {speaker.label}
              </button>
            ))}
          </div>
        )}
      </div>
      )}

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 space-y-1 overflow-y-auto overscroll-contain px-3 pb-3"
      >
        {visibleMessages.length === 0 && (
          <p className="px-2 py-3 text-center text-[12px] text-text-tertiary">
            {filterSpeakerId ? '该专家暂无消息' : '暂无消息'}
          </p>
        )}
        {visibleMessages.map((message) => {
          const label = messageLabel(message)
          const { summary, truncated } = condenseMessage(message.content)
          const isCurrentTurn = activeMessageId != null && message.id === activeMessageId
          const sequence = messageSequenceNumber(message, sequenceMap)

          return (
            <button
              key={message.id}
              type="button"
              onClick={() => onSelect(message)}
              className={cn(
                'flex w-full items-start gap-2 rounded-lg px-2 py-2 text-left transition-colors',
                'hover:bg-black/[0.04]',
                selectedId === message.id && 'bg-brand-soft/80 ring-1 ring-brand/20',
                isCurrentTurn && 'ring-1 ring-ai/25 bg-ai-soft/50',
              )}
            >
              <span
                className={cn(
                  'shrink-0 pt-0.5 font-mono text-[11px] font-semibold tabular-nums',
                  sequence != null ? 'text-brand' : 'text-text-tertiary',
                )}
              >
                {sequence != null ? `#${sequence}` : '·'}
              </span>
              <span className="min-w-0 flex-1">
                <span className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                  <span className="text-[12px] font-medium text-text-secondary">{label}</span>
                  <span className="text-[11px] tabular-nums text-text-tertiary">
                    {formatChatTime(message.createdAt)}
                  </span>
                </span>
                <span
                  className={cn(
                    'mt-0.5 block text-[13px] leading-snug text-text-primary',
                    truncated ? 'line-clamp-1' : 'line-clamp-2',
                  )}
                >
                  {summary || '（空）'}
                </span>
              </span>
              {truncated && (
                <span className="shrink-0 pt-0.5 text-[11px] text-brand">详情</span>
              )}
            </button>
          )
        })}
      </div>

      {!pinnedBottom && newBelow > 0 && (
        <button
          type="button"
          onClick={scrollToBottom}
          className="absolute bottom-3 left-1/2 z-10 -translate-x-1/2 rounded-full bg-surface px-3 py-1 text-[12px] text-brand shadow-md ring-1 ring-black/[0.08]"
        >
          ↓ 新消息
        </button>
      )}
    </div>
  )
}
